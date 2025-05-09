package command

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/vars7899/iots/config"
	"github.com/vars7899/iots/internal/domain/model"
	"github.com/vars7899/iots/internal/repository"
	"github.com/vars7899/iots/pkg/apperror"
	"github.com/vars7899/iots/pkg/logger"
	"github.com/vars7899/iots/pkg/pubsub"
	"go.uber.org/zap"
)

type CommandProcessorService struct {
	commandRegistry *config.DeviceCommandSchemaRegistry
	pubsub          pubsub.PubSubPublisher
	deviceRepo      repository.DeviceRepository
	inboundSub      *nats.Subscription
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
	logger          *zap.Logger
}

func NewCommandProcessorService(reg *config.DeviceCommandSchemaRegistry, pubsub pubsub.PubSubPublisher, deviceRepo repository.DeviceRepository, baseLogger *zap.Logger) (*CommandProcessorService, error) {
	logger := logger.Named(baseLogger, "CommandProcessorService")

	if reg == nil {
		logger.Error("missing device command schema registry")
		return nil, apperror.ErrMissingDependency.WithMessage("device command schema registry is nil").AsInternal()
	}
	if pubsub == nil {
		logger.Error("missing pubsub publisher")
		return nil, apperror.ErrMissingDependency.WithMessage("pubsub publisher is nil").AsInternal()
	}
	if deviceRepo == nil {
		logger.Error("missing device repository")
		return nil, apperror.ErrMissingDependency.WithMessage("device repository is nil").AsInternal()
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &CommandProcessorService{
		commandRegistry: reg,
		pubsub:          pubsub,
		deviceRepo:      deviceRepo,
		ctx:             ctx,
		cancel:          cancel,
		logger:          logger,
	}, nil
}

func (s *CommandProcessorService) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	topic := pubsub.NatsTopicCommandsInbound
	s.logger.Info("Subscribing to PubSub", zap.String("topic", topic))

	msgChan, err := s.pubsub.Subscribe(s.ctx, topic)
	if err != nil {
		s.logger.Error("Failed to start service", zap.Error(err))
		return apperror.ErrInit.WithMessagef("failed to subscribe to pubsub topic '%s'", topic).Wrap(err).AsInternal()
	}

	go s.processMessages(msgChan)

	s.logger.Info("Started message processing...")

	return nil
}

func (s *CommandProcessorService) processMessages(msg <-chan []byte) {
	for {
		select {
		case messageData := <-msg:
			// TODO: use a separate context for the database, for ongoing db
			s.handleMessage(s.ctx, messageData)
		case <-s.ctx.Done():
			s.logger.Info("Shutting message processing routine...")
			return
		}
	}
}

func (s *CommandProcessorService) handleMessage(ctx context.Context, messageData []byte) {
	var cmd model.DeviceCommand

	// response event
	res := model.DeviceEvent{
		Type:      model.EventTypeCommandReceived,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"raw_size": len(messageData),
		},
	}

	err := json.Unmarshal(messageData, &cmd)
	if err != nil {
		s.logger.Error("Failed to unmarshal incoming command", zap.Error(err))

		res.Type = model.EventTypeError
		res.Payload["error"] = "failed to parse command message: invalid/corrupt command message"
		res.Payload["error_code"] = "unmarshal_error"

		s.publishEvent(ctx, pubsub.NatsTopicSystemEvents, res)
		return
	}

	// now populate event with data from the command
	res.DeviceID = cmd.DeviceID
	res.CommandCode = cmd.CommandCode
	res.Payload["command"] = cmd.Command

	s.publishEvent(ctx, pubsub.NatsTopicSystemEvents, res) // monitor whether command received

	s.logger.Info("Received command", zap.String("command", cmd.CommandCode))

	if cmd.DeviceID.String() == "" {
		s.publishErrorEvent(ctx, cmd, "missing_device_id", "device id is required for command")
		return
	}

	device, err := s.deviceRepo.GetByID(ctx, cmd.DeviceID)
	if err != nil {
		var errCode string
		var errMessage string

		if appErr, ok := err.(*apperror.AppError); !ok {
			errCode = string(appErr.Code)
			errMessage = appErr.Message
		} else {
			errCode = "internal_error"
			errMessage = "internal error while retrieving device"
		}

		s.publishErrorEvent(ctx, cmd, errCode, errMessage)
		return
	}
	if device == nil {
		s.publishErrorEvent(ctx, cmd, "device_not_found", fmt.Sprintf("device with id '%s' not found", cmd.DeviceID))
		return
	}

	commandSchema := s.commandRegistry.GetCommandByCode(cmd.CommandCode)
	if commandSchema == nil {
		s.logger.Error("Received command with unknown code", zap.String("command", cmd.Command))
		s.publishErrorEvent(ctx, cmd, "unknown_command", fmt.Sprintf("unknown command code: %s", cmd.CommandCode))
		return
	}

	originalStatus := device.Status // to track change

	processResult, err := s.processCommand(ctx, cmd, device)

	if err != nil {
		// If there was an error processing the command
		if appErr, ok := err.(*apperror.AppError); ok {
			s.publishErrorEvent(ctx, cmd, string(appErr.Code), appErr.Message)
		} else {
			s.publishErrorEvent(ctx, cmd, "PROCESSING_ERROR", err.Error())
		}
		return
	}

	// Check if state changed
	stateChanged := originalStatus != device.Status
	// Add other state change checks as needed
	// || originalFirmwareVersion != device.FirmwareVersion

	// If state changed, update database
	if stateChanged {
		if updateErr := s.deviceRepo.UpdateStatus(ctx, device.ID, device.Status); updateErr != nil {
			appErr := apperror.MapDBError(updateErr, "device state update")
			s.logger.Error("Failed to update device state",
				zap.String("device_id", cmd.DeviceID.String()),
				zap.Error(appErr))

			s.publishErrorEvent(ctx, cmd, string(appErr.Code), fmt.Sprintf("Failed to update device state: %s", appErr.Message))
			return
		}

		// Publish state change event
		stateEvent := model.DeviceEvent{
			Type:        model.EventTypeStateChanged,
			DeviceID:    cmd.DeviceID,
			CommandCode: cmd.CommandCode,
			Timestamp:   time.Now(),
			Payload: map[string]interface{}{
				"previous_status": originalStatus,
				"current_status":  device.Status,
				"device":          device,
			},
		}

		deviceEventTopic := fmt.Sprintf("%s.%s", pubsub.NatsTopicDeviceEventsPrefix, cmd.DeviceID)
		s.publishEvent(ctx, deviceEventTopic, stateEvent)
	}

	// Publish success event
	successEvent := model.DeviceEvent{
		ID:            uuid.New(),
		Type:          model.EventTypeCommandProcessed,
		DeviceID:      cmd.DeviceID,
		CommandCode:   cmd.CommandCode,
		Timestamp:     time.Now(),
		Payload:       processResult,
		CorrelationID: cmd.ID,
	}

	topic := pubsub.NatsTopicCommandsOutboundPrefixf(cmd.DeviceID)
	s.publishEvent(ctx, topic, successEvent)

	s.logger.Info("Command processed successfully",
		zap.String("command", cmd.CommandCode),
		zap.String("device_id", cmd.DeviceID.String()))
}

// processCommand handles the actual command logic
func (s *CommandProcessorService) processCommand(ctx context.Context, cmd model.DeviceCommand, device *model.Device) (map[string]interface{}, error) {
	// This is where command-specific logic lives
	switch cmd.CommandCode {
	case "device@init_ack":
		s.logger.Info("Processing device initialization acknowledgment",
			zap.String("device_id", cmd.DeviceID.String()))

		// Update device status to online
		device.Status = model.DeviceStatusOnline

		return map[string]interface{}{
			"message": "Device initialized successfully",
			"status":  device.Status,
		}, nil

	// Add more command handlers here
	case "device@status_update":
		// Handle status update command
		return map[string]interface{}{
			"message": "Status updated",
		}, nil

	case "device@reboot":
		// Handle reboot command
		return map[string]interface{}{
			"message": "Reboot initiated",
		}, nil

	default:
		return nil, apperror.Errorf(apperror.ErrCodeBadRequest, "Unsupported command code: %s", cmd.CommandCode)
	}
}

func (s *CommandProcessorService) publishEvent(ctx context.Context, topic string, event model.DeviceEvent) {
	if err := s.pubsub.Publish(ctx, topic, event); err != nil {
		s.logger.Warn("Failed to publish event",
			zap.String("topic", topic),
			zap.String("event_type", event.Type),
			zap.String("device", event.DeviceID.String()),
			zap.Error(err),
		)
		return
	}
	s.logger.Debug("Event published",
		zap.String("topic", topic),
		zap.String("event_type", event.Type),
		zap.String("device", event.DeviceID.String()),
	)
}

// publishErrorEvent publishes an error event
func (s *CommandProcessorService) publishErrorEvent(ctx context.Context, cmd model.DeviceCommand, errCode, errMessage string) {
	event := model.DeviceEvent{
		Type:        model.EventTypeError,
		DeviceID:    cmd.DeviceID,
		CommandCode: cmd.CommandCode,
		Timestamp:   time.Now(),
		Payload: map[string]interface{}{
			"error":      errMessage,
			"error_code": errCode,
			"command":    cmd.Command,
		},
		CorrelationID: cmd.ID,
	}

	// nats: publish to both the device-specific channel and system events
	if cmd.DeviceID.String() != "" {
		topic := pubsub.NatsTopicCommandsOutboundPrefixf(cmd.DeviceID)
		s.publishEvent(ctx, topic, event)
	}

	// nats: always publish to system events for monitoring
	s.publishEvent(ctx, pubsub.NatsTopicSystemEvents, event)

	s.logger.Warn("Command processing error",
		zap.String("error_code", errCode),
		zap.String("error", errMessage),
		zap.String("command", cmd.CommandCode),
		zap.String("device_id", cmd.DeviceID.String()))
}

// Stop gracefully shuts down the service
func (s *CommandProcessorService) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Info("Stopping command processor service")

	// Cancel context to signal processing goroutine to stop
	s.cancel()

	// Unsubscribe from NATS topics
	cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := s.pubsub.Unsubscribe(cleanupCtx, pubsub.NatsTopicCommandsInbound)
	if err != nil {
		s.logger.Error("Error unsubscribing from topic",
			zap.String("topic", pubsub.NatsTopicCommandsInbound),
			zap.Error(err))
		// Continue with closing even if unsubscribe fails
	}

	s.logger.Info("Command processor service stopped")
	return nil
}

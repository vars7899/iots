package worker

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/vars7899/iots/internal/api/v1/dto"
	"github.com/vars7899/iots/internal/service"
	"github.com/vars7899/iots/internal/ws"
	"github.com/vars7899/iots/pkg/logger"
	"go.uber.org/zap"
)

func TelemetryWorker(ctx context.Context, wg *sync.WaitGroup, message <-chan ws.ClientMessage, telemetryService *service.TelemetryService, baseLogger *zap.Logger) {
	defer wg.Done()

	l := logger.Named(baseLogger, "TelemetryWorker")
	l.Info("telemetry worker started")

	for {
		select {
		case msg, ok := <-message:
			if !ok {
				l.Info("telemetry channel closed, closing telemetry worker")
				return
			}
			l.Debug("received message from telemetry client", zap.String("client_id", msg.Client.ID), zap.ByteString("raw", msg.Message))

			var payloadDTO dto.TelemetryPayloadDTO
			if err := json.Unmarshal(msg.Message, &payloadDTO); err != nil {
				l.Error("failed to pares telemetry message", zap.String("client_id", msg.Client.ID), zap.Error(err), zap.ByteString("raw", msg.Message))
				// TODO: Send error feedback to client msg.Client.SendMessage([]byte("Error: Invalid JSON"))
				continue // Skip to next message
			}

			if err := payloadDTO.ValidateBasicStructure(); err != nil {
				l.Error("telemetry payload basic validation failed", zap.String("client_id", msg.Client.ID), zap.Error(err), zap.ByteString("raw", msg.Message))
				// TODO: Send error feedback to client msg.Client.SendMessage([]byte(fmt.Sprintf("Error: Basic validation failed - %v", err.Error())))
				continue // Skip to next message
			}

			if err := payloadDTO.ValidateAgainstSchema(); err != nil {
				l.Error("telemetry payload schema validation failed", zap.String("client_id", msg.Client.ID), zap.Error(err), zap.Any("payload_dto", payloadDTO))
				// TODO: Send error feedback to client msg.Client.SendMessage([]byte(fmt.Sprintf("Error: Schema validation failed - %v", err.Error())))
				continue // Skip to next message
			}

			telemetryModel, err := payloadDTO.AsModel() // <-- Use the updated AsModel

			if err != nil {
				// This error indicates a problem during conversion (e.g., JSON marshalling of Data)
				l.Error("failed to convert telemetry DTO to model", zap.String("client_id", msg.Client.ID), zap.Error(err), zap.Any("payload_dto", payloadDTO))
				// TODO: Send error feedback
				continue // Skip to next message
			}

			// 5. Call the Telemetry Service to ingest the model
			// Use the client's context (derived from app context) for cancellation signals
			ingestCtx, cancel := context.WithTimeout(msg.Client.Ctx, 10*time.Second) // Use a timeout for the service call
			// ingestCtx := msg.Client.ctx // Alternative: use client's context directly
			err = telemetryService.IngestSensorTelemetry(ingestCtx, telemetryModel)
			cancel() // Release context resources

			if err != nil {
				l.Error("Telemetry service failed to ingest data",
					zap.String("client_id", msg.Client.ID),
					zap.Error(err),
					zap.Any("telemetry_model", telemetryModel), // Log the model
				)
				// TODO: Handle service errors - send error back to client? Retry?
				// msg.Client.SendMessage([]byte(fmt.Sprintf("Error processing data: %v", err.Error())))
			} else {
				l.Debug("Successfully ingested telemetry data", zap.String("client_id", msg.Client.ID), zap.String("sensor_id", telemetryModel.SensorID.String()))
				// TODO: Optionally send confirmation back to client
				// msg.Client.SendMessage([]byte("OK"))
			}

		case <-ctx.Done():
			l.Info("Application context cancelled telemetry worker existing")
			return
		}
	}
}

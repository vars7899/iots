package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	gorillaWs "github.com/gorilla/websocket"
	"github.com/vars7899/iots/internal/domain/model"
	"github.com/vars7899/iots/pkg/logger"
	"go.uber.org/zap"
)

type DeviceSession struct {
	ID       uuid.UUID       // unique session id
	DeviceID uuid.UUID       // session bind to device
	Conn     *gorillaWs.Conn // ws connection reference
	Channel  chan []byte     // channel to send message to client
	closed   atomic.Bool     // session work status
	logger   *zap.Logger
	ctx      context.Context
	cancel   context.CancelFunc
	mu       sync.RWMutex
}

func NewDeviceSession(deviceID uuid.UUID, conn *gorillaWs.Conn, baseLogger *zap.Logger) *DeviceSession {
	ctx, cancel := context.WithCancel(context.Background())
	return &DeviceSession{
		ID:       uuid.New(),
		DeviceID: deviceID,
		Conn:     conn,
		Channel:  make(chan []byte, 100),
		logger:   logger.Named(baseLogger, fmt.Sprintf("DeviceSession-%s", deviceID)),
		ctx:      ctx,
		cancel:   cancel,
	}
}

func (s *DeviceSession) SendMessage(messageData []byte) error {
	select {
	case <-s.ctx.Done():
		return s.ctx.Err()
	default:
		writeDeadline := time.Now().Add(15 * time.Second)
		if err := s.Conn.SetWriteDeadline(writeDeadline); err != nil {
			s.logger.Warn("Failed to set write deadline", zap.Error(err))
			// continue without write deadline
		}

		var message model.DeviceEvent
		if err := json.Unmarshal(messageData, &message); err != nil {
			s.logger.Warn("Failed to unmarshal message", zap.Error(err))
		}

		fmt.Println("mmmm", message.CommandCode)

		err := s.Conn.WriteJSON(message)
		// err := s.Conn.WriteMessage(gorillaWs.TextMessage, messageData)
		// reset write deadline
		if err := s.Conn.SetWriteDeadline(time.Time{}); err != nil {
			s.logger.Warn("Failed to clear write deadline", zap.Error(err))
		}
		if err != nil {
			s.logger.Warn("Websocket write error", zap.Error(err))
		}
		return err
	}
}

func (s *DeviceSession) Close() error {
	// Use the context cancellation to signal closure
	s.cancel() // Signal goroutines to stop

	// Close the WebSocket connection
	// Send a close message to the client
	err := s.Conn.WriteMessage(gorillaWs.CloseMessage, gorillaWs.FormatCloseMessage(gorillaWs.CloseNormalClosure, ""))
	if err != nil {
		s.logger.Warn("Failed to send close message to websocket", zap.Error(err))
	}

	// Close the underlying connection
	if connErr := s.Conn.Close(); connErr != nil {
		s.logger.Error("Failed to close websocket connection", zap.Error(connErr))
		// TODO: Map this error using apperror
		// return apperror.ErrShutdown.WithMessagef("failed to close websocket connection for session %s for device ID: %s", s.ID, s.DeviceID)
	}

	s.logger.Info("Device session closed", zap.String("session_id", s.ID.String()), zap.String("device_id", s.DeviceID.String()))

	// Closing the session.Channel will be handled by the writePump goroutine
	return nil
}

// func (s *DeviceSession) Close() error {
// 	if !s.closed.CompareAndSwap(false, true) {
// 		return nil // Already closed
// 	}

// 	s.logger.Info("Closing device session", zap.String("session_id", s.ID.String()), zap.String("device_id", s.DeviceID.String()))

// 	close(s.Channel)
// 	if err := s.Conn.Close(); err != nil {
// 		return apperror.ErrShutdown.WithMessagef("failed to close session for session %s for device ID: %s", s.ID, s.DeviceID)
// 	}

// 	return nil
// }

package websocket

import (
	"fmt"
	"sync/atomic"

	"github.com/google/uuid"
	gorillaWs "github.com/gorilla/websocket"
	"github.com/vars7899/iots/pkg/apperror"
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
}

func NewDeviceSession(deviceID uuid.UUID, conn *gorillaWs.Conn, baseLogger *zap.Logger) *DeviceSession {
	return &DeviceSession{
		ID:       uuid.New(),
		DeviceID: deviceID,
		Conn:     conn,
		Channel:  make(chan []byte),
		logger:   logger.Named(baseLogger, fmt.Sprintf("UserSession-%s", deviceID)),
	}
}

func (s *DeviceSession) SendMessage(message []byte) error {
	return s.Conn.WriteMessage(gorillaWs.TextMessage, message)
}

func (s *DeviceSession) Close() error {
	if !s.closed.CompareAndSwap(false, true) {
		return nil // Already closed
	}

	s.logger.Info("Closing device session", zap.String("session_id", s.ID.String()), zap.String("device_id", s.DeviceID.String()))

	close(s.Channel)
	if err := s.Conn.Close(); err != nil {
		return apperror.ErrShutdown.WithMessagef("failed to close session for session %s for device ID: %s", s.ID, s.DeviceID)
	}

	return nil
}

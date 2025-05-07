package websocket

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/vars7899/iots/pkg/logger"
	"go.uber.org/zap"
)

type SessionManager struct {
	sessions             map[uuid.UUID]*DeviceSession
	deviceSessionMapping map[uuid.UUID]uuid.UUID
	mu                   sync.RWMutex
	logger               *zap.Logger
}

func NewSessionManager(baseLogger *zap.Logger) *SessionManager {
	return &SessionManager{
		sessions:             make(map[uuid.UUID]*DeviceSession),
		deviceSessionMapping: make(map[uuid.UUID]uuid.UUID),
		logger:               logger.Named(baseLogger, "SessionManager"),
	}
}

func (m *SessionManager) AddSession(session *DeviceSession) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if sessionExist, ok := m.deviceSessionMapping[session.DeviceID]; ok {
		m.logger.Info("Active session exist for device", zap.String("session_id", sessionExist.String()), zap.String("device_id", session.DeviceID.String()))
		return
	}
	m.sessions[session.ID] = session
	m.deviceSessionMapping[session.DeviceID] = session.ID
	m.logger.Info(fmt.Sprintf("Session added for device %s. Total sessions: %d", session.DeviceID.String(), len(m.sessions)))
}
func (m *SessionManager) RemoveSession(sessionID uuid.UUID) {
	m.mu.Lock()
	defer m.mu.Unlock()
	exist, ok := m.sessions[sessionID]
	if ok {
		delete(m.sessions, sessionID)
		delete(m.deviceSessionMapping, exist.DeviceID)
		m.logger.Info(fmt.Sprintf("Session removed for device %s. Total sessions: %d", exist.DeviceID.String(), len(m.sessions)))
	} else {
		m.logger.Warn(fmt.Sprintf("Attempted to remove non-existent session: %s", sessionID.String()))
	}
}
func (m *SessionManager) GetSession(sessionID uuid.UUID) (*DeviceSession, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	session, ok := m.sessions[sessionID]
	return session, ok
}

func (m *SessionManager) GetSessionByDeviceID(deviceID uuid.UUID) (*DeviceSession, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	sessionID, ok := m.deviceSessionMapping[deviceID]
	if !ok || sessionID == uuid.Nil {
		m.logger.Warn("No active session exist", zap.String("device_id", deviceID.String()))
		return nil, false
	}
	session, ok := m.sessions[sessionID]
	return session, ok
}
func (m *SessionManager) Broadcast(message []byte) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, session := range m.sessions {
		select {
		case session.Channel <- message:
		default:
			m.logger.Info("Session channel full for device message", zap.String("session", session.ID.String()))
		}
	}
}

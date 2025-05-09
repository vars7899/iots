package websocket

import (
	"sync"

	"github.com/google/uuid"
	"github.com/vars7899/iots/pkg/logger"
	"go.uber.org/zap"
)

type SessionManager struct {
	sessions       map[uuid.UUID]*DeviceSession               // key <session_id>
	deviceSessions map[uuid.UUID]map[uuid.UUID]*DeviceSession // key <device_id>, value <session_id>
	mu             sync.RWMutex
	logger         *zap.Logger
}

func NewSessionManager(baseLogger *zap.Logger) *SessionManager {
	return &SessionManager{
		sessions:       make(map[uuid.UUID]*DeviceSession),
		deviceSessions: make(map[uuid.UUID]map[uuid.UUID]*DeviceSession),
		logger:         logger.Named(baseLogger, "SessionManager"),
	}
}

func (m *SessionManager) AddSession(session *DeviceSession) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sessions[session.ID] = session // add session to main session list

	activeDeviceSessions, ok := m.deviceSessions[session.DeviceID]
	if !ok {
		// if no active device session exist, add new entry for device
		activeDeviceSessions = make(map[uuid.UUID]*DeviceSession)
		m.deviceSessions[session.DeviceID] = activeDeviceSessions
	}

	activeDeviceSessions[session.ID] = session

	m.logger.Info("Device session added",
		zap.String("session_id", session.ID.String()),
		zap.String("device_id", session.DeviceID.String()),
		zap.Int("total_session", len(m.sessions)),
		zap.Int("device_session", len(activeDeviceSessions)),
	)
}

func (m *SessionManager) RemoveSession(sessionID uuid.UUID) {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, ok := m.sessions[sessionID]
	if !ok {
		m.logger.Warn("Attempted to remove non-existent session")
		return
	}

	delete(m.sessions, sessionID) // remove session from the main session list

	if activeDeviceSessions, ok := m.deviceSessions[session.DeviceID]; ok {
		delete(activeDeviceSessions, sessionID)

		if len(activeDeviceSessions) == 0 {
			// if no active session, delete the device sessions
			delete(m.deviceSessions, session.DeviceID)
			m.logger.Info("Last active session removed",
				zap.String("session_id", sessionID.String()),
				zap.String("device_id", session.DeviceID.String()),
			)
		} else {
			m.logger.Info("Session removed",
				zap.String("session_id", sessionID.String()),
				zap.String("device_id", session.DeviceID.String()),
				zap.Int("total_session", len(m.sessions)),
				zap.Int("device_session", len(activeDeviceSessions)),
			)
		}
	} else {
		m.logger.Error("Inconsistent state: session found but not available in active device session",
			zap.String("session_id", sessionID.String()),
			zap.String("device_id", session.DeviceID.String()),
		)
	}
}
func (m *SessionManager) GetSession(sessionID uuid.UUID) (*DeviceSession, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	session, ok := m.sessions[sessionID]
	if ok {
		select {
		case <-session.ctx.Done():
			// if session is marked for closure, it is not active
			return nil, false
		default:
			return session, true
		}
	}
	// not found
	return nil, false
}

func (m *SessionManager) GetSessionByDeviceID(deviceID uuid.UUID) []*DeviceSession {
	m.mu.RLock()
	defer m.mu.RUnlock()

	activeDeviceSessions, ok := m.deviceSessions[deviceID]
	if !ok || len(activeDeviceSessions) == 0 {
		m.logger.Debug("No active sessions found",
			zap.String("device_id", deviceID.String()),
		)
		return nil
	}
	sessionList := make([]*DeviceSession, 0, len(activeDeviceSessions))
	for _, s := range activeDeviceSessions {
		select {
		case <-s.ctx.Done():
			m.logger.Debug("Skipping session marked as closure")
		default:
			sessionList = append(sessionList, s)
		}
	}

	// after filtering, no active session are found
	if len(sessionList) == 0 {
		return nil
	}
	return sessionList
}
func (m *SessionManager) Broadcast(message []byte) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, session := range m.sessions {
		select {
		case session.Channel <- message:
		case <-session.ctx.Done():
			m.logger.Debug("Skipping broadcast to session marked as closure")
		default:
			m.logger.Info("Session channel full for device message", zap.String("session", session.ID.String()))
		}
	}
}

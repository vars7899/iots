package websocket

import (
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	gorillaWs "github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/vars7899/iots/config"
	"github.com/vars7899/iots/pkg/logger"
	"github.com/vars7899/iots/pkg/utils"

	"go.uber.org/zap"
)

type WebsocketHandler struct {
	manager         *SessionManager
	config          *config.WebsocketConfig
	credentialStore *CredentialStore
	logger          *zap.Logger
}

func NewWebsocketHandler(sm *SessionManager, config *config.WebsocketConfig, baseLogger *zap.Logger) *WebsocketHandler {
	return &WebsocketHandler{
		manager:         sm,
		config:          config,
		credentialStore: NewCredentialStore(),
		logger:          logger.Named(baseLogger, "WebsocketHandler"),
	}
}

func (h *WebsocketHandler) HandleConnection(c echo.Context) error {
	var upgrader = gorillaWs.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	deviceID, _, err := h.authenticateOrRegisterDevice(c)
	if err != nil {
		h.logger.Warn("Authentication failed", zap.Error(err))
		return err
	}

	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		h.logger.Warn("Failed to upgrade websocket connectio", zap.String("device_id", deviceID.String()), zap.Error(err))
		return nil
	}

	// new session
	session := NewDeviceSession(uuid.New(), conn, h.logger)
	h.manager.AddSession(session)

	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		h.readPump(session)
	}()
	go func() {
		defer wg.Done()
		h.writePump(session)
	}()

	wg.Wait()
	return nil
}

func (h *WebsocketHandler) readPump(session *DeviceSession) {
	defer func() {
		session.Close()
		h.manager.RemoveSession(session.ID)
	}()

	session.Conn.SetReadLimit(h.config.ReadDeadline)
	session.Conn.SetReadDeadline(time.Now().Add(h.config.PongTimeout))
	session.Conn.SetPongHandler(func(string) error {
		session.Conn.SetReadDeadline(time.Now().Add(h.config.PongTimeout))
		return nil
	})

	for {
		_, msg, err := session.Conn.ReadMessage()
		if err != nil {
			if gorillaWs.IsUnexpectedCloseError(err, gorillaWs.CloseGoingAway, gorillaWs.CloseAbnormalClosure) {
				h.logger.Warn("Unexpected websocket connection close", zap.Error(err))
			}
			break
		}
		h.manager.Broadcast(msg)
	}
}

func (h *WebsocketHandler) writePump(session *DeviceSession) {
	tick := time.NewTicker(h.config.PongTimeout)
	defer func() {
		tick.Stop()
		session.Close()
		h.manager.RemoveSession(session.ID)
	}()

	for {
		select {
		case message, ok := <-session.Channel:
			if !ok {
				session.Conn.WriteMessage(gorillaWs.CloseMessage, []byte{})
				return
			}
			if err := session.Conn.WriteMessage(gorillaWs.TextMessage, message); err != nil {
				h.logger.Warn("Websocket write error", zap.Error(err))
				return
			}
		case <-tick.C:
			if err := session.Conn.WriteMessage(gorillaWs.PingMessage, nil); err != nil {
				h.logger.Warn("Websocket ping error", zap.Error(err))
				return
			}
		}
	}

}

func (h *WebsocketHandler) authenticateOrRegisterDevice(c echo.Context) (uuid.UUID, string, error) {
	initToken := c.Request().Header.Get("X-Initial-Token")
	deviceIDStr := c.Request().Header.Get("X-Device-ID")

	if initToken == "" || deviceIDStr == "" {
		return uuid.Nil, "", echo.NewHTTPError(http.StatusBadRequest, "Missing initial token or device ID")
	}

	deviceID, err := uuid.Parse(deviceIDStr)
	if err != nil {
		return uuid.Nil, "", echo.NewHTTPError(http.StatusBadRequest, "Invalid UUID format for device ID")
	}

	// // --- If already registered, return existing token ---
	// if token, ok := h.credentialStore.GetPermanentToken(deviceID); ok {
	// 	return deviceID, *token, nil
	// }

	// --- Validate initial token ---
	if !h.credentialStore.ValidateInitialToken(deviceID, initToken) {
		return uuid.Nil, "", echo.NewHTTPError(http.StatusUnauthorized, "Invalid initial token or device ID")
	}

	// --- Generate permanent token ---
	permanentToken, err := utils.GenerateSecureToken(32)
	if err != nil {
		return uuid.Nil, "", echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate token")
	}

	// --- Store and delete initial ---
	h.credentialStore.StoreAsPermanent(deviceID, permanentToken)
	h.credentialStore.DeleteInitialToken(deviceID)

	return deviceID, permanentToken, nil
}

package websocket

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	gorillaWs "github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/vars7899/iots/config"
	"github.com/vars7899/iots/pkg/auth/deviceauth"
	"github.com/vars7899/iots/pkg/contextkey"
	"github.com/vars7899/iots/pkg/logger"

	"go.uber.org/zap"
)

type WebsocketHandler struct {
	manager         *SessionManager
	config          *config.WebsocketConfig
	credentialStore *CredentialStore
	authService     deviceauth.DeviceAuthService
	logger          *zap.Logger
}

func NewWebsocketHandler(sm *SessionManager, auth deviceauth.DeviceAuthService, config *config.WebsocketConfig, baseLogger *zap.Logger) *WebsocketHandler {
	return &WebsocketHandler{
		manager:         sm,
		authService:     auth,
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

	req := c.Request()
	ctx := req.Context()

	// Step 1: Extract token from headers
	connectionToken := req.Header.Get(contextkey.HeaderDeviceConnectionToken)
	refreshToken := req.Header.Get(contextkey.HeaderDeviceRefreshToken)

	if connectionToken == "" || refreshToken == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
	}

	// Step 2: Validate the token
	claims, newToken, err := h.authService.Authenticate(ctx, connectionToken, refreshToken)
	if err != nil || claims == nil {
		h.logger.Warn("Token validation failed", zap.Error(err))
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
	}

	fmt.Println("[[[[[[[[[[[[[[[]]]]]]]]]]]]]]]", claims)

	deviceID, err := claims.DeviceID()
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "malformed token")
	}

	// Step 3: Upgrade to WebSocket
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		h.logger.Error("WebSocket upgrade failed", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "websocket upgrade failed")
	}

	// Step 4: Create and register session
	session := NewDeviceSession(deviceID, conn, h.logger)
	h.manager.AddSession(session)

	// bind header if toke rotated
	if newToken != nil {
		c.Response().Header().Set(contextkey.HeaderDeviceConnectionToken, newToken.ConnectionToken)
		c.Response().Header().Set(contextkey.HeaderDeviceRefreshToken, newToken.RefreshToken)
	}

	// Step 5: Start read/write pumps
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

// func (h *WebsocketHandler) authenticateOrRegisterDevice(c echo.Context) (uuid.UUID, *deviceauth.DeviceConnectionTokens, error) {
// 	req := c.Request()
// 	ctx := req.Context()

// 	connectionToken := req.Header.Get(contextkey.HeaderDeviceConnectionToken)
// 	refreshToken := req.Header.Get(contextkey.HeaderDeviceRefreshToken)

// 	claims, genTokens, err := deviceauth.DeviceAuthService.Authenticate(ctx, connectionToken, refreshToken)
// 	if err != nil {
// 		return uuid.Nil, nil, err
// 	}

// 	initToken := c.Request().Header.Get("X-Initial-Token")
// 	deviceIDStr := c.Request().Header.Get("X-Device-ID")

// 	if initToken == "" || deviceIDStr == "" {
// 		return uuid.Nil, "", echo.NewHTTPError(http.StatusBadRequest, "Missing initial token or device ID")
// 	}

// 	deviceID, err := uuid.Parse(deviceIDStr)
// 	if err != nil {
// 		return uuid.Nil, "", echo.NewHTTPError(http.StatusBadRequest, "Invalid UUID format for device ID")
// 	}

// 	// // --- If already registered, return existing token ---
// 	// if token, ok := h.credentialStore.GetPermanentToken(deviceID); ok {
// 	// 	return deviceID, *token, nil
// 	// }

// 	// --- Validate initial token ---
// 	if !h.credentialStore.ValidateInitialToken(deviceID, initToken) {
// 		return uuid.Nil, "", echo.NewHTTPError(http.StatusUnauthorized, "Invalid initial token or device ID")
// 	}

// 	// --- Generate permanent token ---
// 	permanentToken, err := utils.GenerateSecureToken(32)
// 	if err != nil {
// 		return uuid.Nil, "", echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate token")
// 	}

// 	// --- Store and delete initial ---
// 	h.credentialStore.StoreAsPermanent(deviceID, permanentToken)
// 	h.credentialStore.DeleteInitialToken(deviceID)

// 	return deviceID, permanentToken, nil
// }

// func (h *WebsocketHandler) authenticateOrRegisterDevice(c echo.Context) (uuid.UUID, *deviceauth.DeviceConnectionTokens, error) {
// 	initToken := c.Request().Header.Get("X-Initial-Token")
// 	deviceIDStr := c.Request().Header.Get("X-Device-ID")

// 	if initToken == "" || deviceIDStr == "" {
// 		return uuid.Nil, "", echo.NewHTTPError(http.StatusBadRequest, "Missing initial token or device ID")
// 	}

// 	deviceID, err := uuid.Parse(deviceIDStr)
// 	if err != nil {
// 		return uuid.Nil, "", echo.NewHTTPError(http.StatusBadRequest, "Invalid UUID format for device ID")
// 	}

// 	// // --- If already registered, return existing token ---
// 	// if token, ok := h.credentialStore.GetPermanentToken(deviceID); ok {
// 	// 	return deviceID, *token, nil
// 	// }

// 	// --- Validate initial token ---
// 	if !h.credentialStore.ValidateInitialToken(deviceID, initToken) {
// 		return uuid.Nil, "", echo.NewHTTPError(http.StatusUnauthorized, "Invalid initial token or device ID")
// 	}

// 	// --- Generate permanent token ---
// 	permanentToken, err := utils.GenerateSecureToken(32)
// 	if err != nil {
// 		return uuid.Nil, "", echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate token")
// 	}

// 	// --- Store and delete initial ---
// 	h.credentialStore.StoreAsPermanent(deviceID, permanentToken)
// 	h.credentialStore.DeleteInitialToken(deviceID)

// 	return deviceID, permanentToken, nil
// }

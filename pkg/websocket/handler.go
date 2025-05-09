package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	gorillaWs "github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/vars7899/iots/config"
	"github.com/vars7899/iots/internal/domain/model"
	"github.com/vars7899/iots/pkg/auth/deviceauth"
	"github.com/vars7899/iots/pkg/contextkey"
	"github.com/vars7899/iots/pkg/logger"
	"github.com/vars7899/iots/pkg/pubsub"

	"go.uber.org/zap"
)

type WebsocketHandler struct {
	manager     *SessionManager
	config      *config.WebsocketConfig
	authService deviceauth.DeviceAuthService
	nats        pubsub.PubSubPublisher
	logger      *zap.Logger
}

func NewWebsocketHandler(sm *SessionManager, publisher pubsub.PubSubPublisher, auth deviceauth.DeviceAuthService, config *config.WebsocketConfig, baseLogger *zap.Logger) *WebsocketHandler {
	return &WebsocketHandler{
		manager:     sm,
		authService: auth,
		config:      config,
		nats:        publisher,
		logger:      logger.Named(baseLogger, "WebsocketHandler"),
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
	var natsWg sync.WaitGroup
	natsWg.Add(2)

	// Subscribe to command responses for this device
	go func() {
		defer natsWg.Done()
		responseTopic := pubsub.NatsTopicCommandsOutboundPrefixf(deviceID)
		h.subscribeAndPush(session, responseTopic)
	}()

	// Subscribe to device state change events for this device
	go func() {
		defer natsWg.Done()
		eventTopic := pubsub.NatsTopicDeviceEventsPrefixf(deviceID)
		h.subscribeAndPush(session, eventTopic)
	}()

	var pumpWg sync.WaitGroup
	pumpWg.Add(2)

	go func() {
		defer pumpWg.Done()
		h.readPump(session)
	}()
	go func() {
		defer pumpWg.Done()
		h.writePump(session)
	}()

	pumpWg.Wait()
	natsWg.Wait()

	h.manager.RemoveSession(session.ID)
	h.logger.Info("Websocket connection closed for device", zap.String("device_id", deviceID.String()))

	return nil
}

// subscribeAndPush handles subscribing to a NATS topic and pushing messages to the session channel
func (h *WebsocketHandler) subscribeAndPush(session *DeviceSession, topic string) {
	h.logger.Debug("Subscribing session to NATS topic",
		zap.String("session_id", session.ID.String()),
		zap.String("device_id", session.DeviceID.String()),
		zap.String("topic", topic))

	// Subscribe using the session's context
	// Your pubsub.Subscribe returns <-chan []byte and error
	msgChan, err := h.nats.Subscribe(session.ctx, topic)
	if err != nil {
		h.logger.Error("Failed to subscribe session to NATS topic",
			zap.String("session_id", session.ID.String()),
			zap.String("device_id", session.DeviceID.String()),
			zap.String("topic", topic),
			zap.Error(err))
		// TODO: Handle this failure - maybe close the session?
		return // Exit this goroutine if subscription fails
	}

	h.logger.Info("Session subscribed to NATS topic",
		zap.String("session_id", session.ID.String()),
		zap.String("device_id", session.DeviceID.String()),
		zap.String("topic", topic))

	// Goroutine to push messages from NATS channel to session channel
	go func() {
		h.logger.Debug("Starting NATS message push goroutine for topic",
			zap.String("session_id", session.ID.String()),
			zap.String("device_id", session.DeviceID.String()),
			zap.String("topic", topic))

		for {
			select {
			case msgData, ok := <-msgChan:
				if !ok {
					// Channel was closed, likely due to NATS unsubscribe or session context cancellation
					h.logger.Info("NATS message channel closed for topic",
						zap.String("session_id", session.ID.String()),
						zap.String("device_id", session.DeviceID.String()),
						zap.String("topic", topic))
					return // Stop the goroutine
				}

				// We received a message from NATS, push it to the session's channel
				select {
				case session.Channel <- msgData:
					h.logger.Debug("Pushed NATS message to session channel",
						zap.String("session_id", session.ID.String()),
						zap.String("device_id", session.DeviceID.String()),
						zap.String("topic", topic),
						zap.Int("msg_size", len(msgData)))
				case <-session.ctx.Done():
					// Session is closing, stop pushing
					h.logger.Info("Session context done while pushing NATS message for topic",
						zap.String("session_id", session.ID.String()),
						zap.String("device_id", session.DeviceID.String()),
						zap.String("topic", topic))
					return // Stop the goroutine
				default:
					// Session channel is full, this indicates the client is not reading fast enough
					h.logger.Warn("Session channel full, dropping NATS message for topic",
						zap.String("session_id", session.ID.String()),
						zap.String("device_id", session.DeviceID.String()),
						zap.String("topic", topic),
						zap.Int("msg_size", len(msgData)))
					// TODO: Implement a strategy for full channels (e.g., close session, log and drop)
					// For now, we just log and drop. Closing the session is a common approach
					// to prevent unbounded memory growth if a client stops reading.
					session.Close() // Uncomment to close session on full channel
				}

			case <-session.ctx.Done():
				// Session context cancelled, stop the goroutine
				h.logger.Info("Session context done, stopping NATS message push goroutine for topic",
					zap.String("session_id", session.ID.String()),
					zap.String("device_id", session.DeviceID.String()),
					zap.String("topic", topic))
				return
			}
		}
	}()

	// This goroutine blocks until the session context is done, then unsubscribes
	<-session.ctx.Done()
	h.logger.Debug("Session context done, unsubscribing from NATS topic",
		zap.String("session_id", session.ID.String()),
		zap.String("device_id", session.DeviceID.String()),
		zap.String("topic", topic))

	// Use a background context for cleanup as the session context is already done
	cleanupCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second) // Configurable timeout
	defer cancel()
	// Call Unsubscribe on the natsSubscriber (your PubSubPublisher implementation)
	if err := h.nats.Unsubscribe(cleanupCtx, topic); err != nil {
		h.logger.Error("Failed to unsubscribe session from NATS topic during cleanup",
			zap.String("session_id", session.ID.String()),
			zap.String("device_id", session.DeviceID.String()),
			zap.String("topic", topic),
			zap.Error(err))
	} else {
		h.logger.Info("Session unsubscribed from NATS topic",
			zap.String("session_id", session.ID.String()),
			zap.String("device_id", session.DeviceID.String()),
			zap.String("topic", topic))
	}
}

func (h *WebsocketHandler) readPump(session *DeviceSession) {
	// Ensure session is closed when readPump exits
	defer func() {
		session.Close() // This will trigger context cancellation and NATS unsubscribe
		// The HandleConnection function waits for pumps and removes from manager
		// h.manager.RemoveSession(session.ID) // Removed, HandleConnection does this
	}()

	// Configure read settings
	session.Conn.SetReadLimit(h.config.ReadDeadline) // Assuming ReadDeadline is max message size
	pongWait := h.config.PongTimeout                 // Assuming PongTimeout is the interval for expected pongs
	session.Conn.SetReadDeadline(time.Now().Add(pongWait))
	session.Conn.SetPongHandler(func(string) error {
		session.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	h.logger.Debug("Starting read pump", zap.String("session_id", session.ID.String()))

	for {
		// Check if session context is done before reading
		select {
		case <-session.ctx.Done():
			h.logger.Info("Session context done, stopping read pump", zap.String("session_id", session.ID.String()))
			return // Stop the goroutine
		default:
			// Continue reading from WebSocket
		}

		// Read message from WebSocket
		_, msg, err := session.Conn.ReadMessage()
		if err != nil {
			if gorillaWs.IsUnexpectedCloseError(err, gorillaWs.CloseGoingAway, gorillaWs.CloseAbnormalClosure) {
				h.logger.Warn("Unexpected websocket connection close",
					zap.String("session_id", session.ID.String()),
					zap.String("device_id", session.DeviceID.String()),
					zap.Error(err))
			} else {
				h.logger.Info("Websocket read error, closing connection",
					zap.String("session_id", session.ID.String()),
					zap.String("device_id", session.DeviceID.String()),
					zap.Error(err))
			}
			break // Exit the loop on read error
		}

		// --- Process incoming message from device ---
		// This message is a command or data from the device TO the server.
		// It needs to be published to the NATS inbound topic for processing by CommandProcessorService.

		// TODO: Unmarshal the incoming message 'msg' into your DeviceCommand structure
		// The device needs to send messages in the expected format.
		var incomingCmd model.DeviceCommand
		unmarshalErr := json.Unmarshal(msg, &incomingCmd)
		if unmarshalErr != nil {
			h.logger.Error("Failed to unmarshal incoming device message as DeviceCommand",
				zap.String("session_id", session.ID.String()),
				zap.String("device_id", session.DeviceID.String()),
				zap.ByteString("raw_msg", msg), // Log raw message for debugging
				zap.Error(unmarshalErr))
			// TODO: Send an error response back to the device? Publish an error event?
			continue // Skip processing this message but keep connection open
		}

		// Add DeviceID from the session to the command (important for routing in CommandProcessorService)
		incomingCmd.ID = uuid.New()
		incomingCmd.DeviceID = session.DeviceID
		incomingCmd.Timestamp = time.Now() // Set server-side timestamp

		// Publish the incoming command to the NATS inbound topic
		topic := pubsub.NatsTopicCommandsInbound
		publishErr := h.nats.Publish(context.Background(), topic, incomingCmd) // Use background context for publishing
		if publishErr != nil {
			h.logger.Error("Failed to publish incoming device command to NATS",
				zap.String("session_id", session.ID.String()),
				zap.String("device_id", session.DeviceID.String()),
				zap.String("command_code", incomingCmd.CommandCode),
				zap.Error(publishErr))
			// TODO: Handle NATS publish error - retry? Log a critical error?
		} else {
			h.logger.Debug("Published incoming device command to NATS",
				zap.String("session_id", session.ID.String()),
				zap.String("device_id", session.DeviceID.String()),
				zap.String("command_code", incomingCmd.CommandCode),
				zap.String("nats_topic", topic))
		}
	}
}

func (h *WebsocketHandler) writePump(session *DeviceSession) {
	// Ticker for sending ping messages to keep the connection alive
	pingInterval := h.config.PingTimeout // Assuming PingInterval is defined in your config
	tick := time.NewTicker(pingInterval)

	// Ensure session is closed when writePump exits
	defer func() {
		tick.Stop()
		session.Close() // This will trigger context cancellation and NATS unsubscribe
		// The HandleConnection function waits for pumps and removes from manager
		// h.manager.RemoveSession(session.ID) // Removed, HandleConnection does this
	}()

	h.logger.Debug("Starting write pump", zap.String("session_id", session.ID.String()))

	// The write pump needs to close the session.Channel when it exits
	// so the pushNatsMessagesToSession goroutines know to stop pushing messages.
	// This should be done after the defer functions are set up.
	defer close(session.Channel)

	for {
		select {
		case message, ok := <-session.Channel:
			// Message received from the session's channel (fed by NATS subscriptions)
			if !ok {
				// Channel was closed, means the session is closing
				h.logger.Info("Session channel closed, stopping write pump", zap.String("session_id", session.ID.String()))
				// Send a close message to the client and return
				if err := session.Conn.WriteMessage(gorillaWs.CloseMessage, []byte{}); err != nil {
					h.logger.Warn("Failed to send close message to websocket", zap.Error(err))
				}
				return // Stop the goroutine
			}

			// Send the message over the WebSocket connection
			if err := session.SendMessage(message); err != nil {
				h.logger.Warn("Websocket write error via SendMessage", zap.Error(err))
				return // Exit the loop on write error
			}
			h.logger.Debug("Message sent to websocket",
				zap.String("session_id", session.ID.String()),
				zap.String("device_id", session.DeviceID.String()),
				zap.Int("msg_size", len(message)))

		case <-tick.C:
			// Time to send a ping message
			if err := session.Conn.WriteMessage(gorillaWs.PingMessage, nil); err != nil {
				h.logger.Warn("Websocket ping error",
					zap.String("session_id", session.ID.String()),
					zap.String("device_id", session.DeviceID.String()),
					zap.Error(err))
				return // Exit the loop on ping error
			}
			h.logger.Debug("Sent ping message", zap.String("session_id", session.ID.String()))

		case <-session.ctx.Done():
			// Session context cancelled, stop the goroutine
			h.logger.Info("Session context done, stopping write pump", zap.String("session_id", session.ID.String()))
			return
		}
	}
}

// func (h *WebsocketHandler) readPump(session *DeviceSession) {
// 	defer func() {
// 		session.Close()
// 		h.manager.RemoveSession(session.ID)
// 	}()

// 	session.Conn.SetReadLimit(h.config.ReadDeadline)
// 	session.Conn.SetReadDeadline(time.Now().Add(h.config.PongTimeout))
// 	session.Conn.SetPongHandler(func(string) error {
// 		session.Conn.SetReadDeadline(time.Now().Add(h.config.PongTimeout))
// 		return nil
// 	})

// 	for {
// 		_, msg, err := session.Conn.ReadMessage()
// 		if err != nil {
// 			if gorillaWs.IsUnexpectedCloseError(err, gorillaWs.CloseGoingAway, gorillaWs.CloseAbnormalClosure) {
// 				h.logger.Warn("Unexpected websocket connection close", zap.Error(err))
// 			}
// 			break
// 		}
// 		// h.manager.Broadcast(msg)

// 		// nats: publish incoming request from the websocket to nats service
// 		topic := pubsub.NatsTopicCommandsInbound
// 		err = h.natsPublisher.Publish(context.TODO(), topic, model.DeviceCommand{
// 			DeviceID:    session.DeviceID,
// 			Type:        "ack",
// 			Command:     "init_ack",
// 			CommandCode: "device@init_ack",
// 			Payload: map[string]interface{}{
// 				"message": msg,
// 			},
// 			Timestamp: time.Now(),
// 		})
// 		if err != nil {
// 			h.logger.Error("Failed to publish command to NATS", zap.Error(err))
// 		}
// 	}
// }

// func (h *WebsocketHandler) writePump(session *DeviceSession) {
// 	tick := time.NewTicker(h.config.PongTimeout)
// 	defer func() {
// 		tick.Stop()
// 		session.Close()
// 		h.manager.RemoveSession(session.ID)
// 	}()

// 	for {
// 		select {
// 		case message, ok := <-session.Channel:
// 			if !ok {
// 				session.Conn.WriteMessage(gorillaWs.CloseMessage, []byte{})
// 				return
// 			}
// 			if err := session.Conn.WriteMessage(gorillaWs.TextMessage, message); err != nil {
// 				h.logger.Warn("Websocket write error", zap.Error(err))
// 				return
// 			}
// 		case <-tick.C:
// 			if err := session.Conn.WriteMessage(gorillaWs.PingMessage, nil); err != nil {
// 				h.logger.Warn("Websocket ping error", zap.Error(err))
// 				return
// 			}
// 		}
// 	}

// }

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

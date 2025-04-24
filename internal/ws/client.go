package ws

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/vars7899/iots/pkg/logger"
	"go.uber.org/zap"
)

type WebsocketClientType string

var (
	SensorTelemetryClient WebsocketClientType = "sensor_telemetry"
	SensorCommand         WebsocketClientType = "sensor_command"
)

// TODO: move this to /config
type WebSocketClientConfig struct {
	MaxReadLimit int64         // max size of message.
	PongTimeout  time.Duration // time allowed to read next pong message.
	PingPeriod   time.Duration // send pings to the peer with this period. Must be less than pong Timeout.
	WriteTimeout time.Duration // time allowed to write message to the peer.
}

type Client struct {
	ID         string
	conn       *websocket.Conn
	sendCh     chan []byte
	hub        *Hub
	l          *zap.Logger
	clientType WebsocketClientType
	config     *WebSocketClientConfig
	Ctx        context.Context
	Cancel     context.CancelFunc
	wg         sync.WaitGroup
}

type ClientMessage struct {
	Client  *Client
	Message []byte
}

func NewClient(h *Hub, clientType WebsocketClientType, conn *websocket.Conn, baseLogger *zap.Logger, cfg *WebSocketClientConfig) *Client {
	ctx, cancel := context.WithCancel(context.Background())

	client := &Client{
		ID:         uuid.NewString(),
		conn:       conn,
		sendCh:     make(chan []byte, 256),
		hub:        h,
		l:          logger.Named(baseLogger, fmt.Sprintf("Client-%s", clientType)),
		config:     cfg,
		clientType: clientType,
		Ctx:        ctx,
		Cancel:     cancel,
	}

	client.wg.Add(1)
	go client.ReadPump()

	return client
}

func (c *Client) ReadPump() {
	c.conn.SetReadLimit(c.config.MaxReadLimit)
	c.conn.SetReadDeadline(time.Now().Add(c.config.PongTimeout))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(c.config.PongTimeout))
		return nil
	})

	for {
		_, messageContent, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.l.Warn("unexpected websocket connection close", zap.Error(err))
			}
			break
		}
		c.hub.incomingMsgCh <- ClientMessage{
			Client:  c,
			Message: messageContent,
		}
	}
}

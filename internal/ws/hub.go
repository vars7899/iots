package ws

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/vars7899/iots/pkg/logger"
	"go.uber.org/zap"
)

var (
	Upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true }, // TODO: Implement proper origin check
	}
)

type Hub struct {
	clients       map[string]*Client
	register      chan *Client
	unregister    chan *Client
	broadcast     chan []byte
	incomingMsgCh chan ClientMessage
	dist          *HubDist
	l             *zap.Logger
	mu            sync.RWMutex
}

type HubDist struct {
	telemetryCh chan ClientMessage
}

func NewHub(baseLogger *zap.Logger) *Hub {
	return &Hub{
		clients:       make(map[string]*Client, 0),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		broadcast:     make(chan []byte),
		incomingMsgCh: make(chan ClientMessage, 5000),
		dist: &HubDist{
			telemetryCh: make(chan ClientMessage, 5000),
		},
		l: logger.Named(baseLogger, "WebsocketHub"),
	}
}

func (h *Hub) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	h.l.Info("Hub Started")

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client
			h.mu.Unlock()
			h.l.Info("client connected", zap.String("client_id", client.ID))
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.sendCh)
				h.l.Info("client disconnected", zap.String("client_id", client.ID))
			}
			h.mu.Unlock()
		case msg := <-h.incomingMsgCh:
			fmt.Printf("{%s}", msg.Message)
			h.handleIncomingMessage(msg)
		case <-ctx.Done():
			h.l.Info("shutting down hub")
			h.mu.Lock()
			if len(h.clients) > 0 {
				for _, client := range h.clients {
					client.conn.Close()
				}
			}
			h.mu.Unlock()
			// close channels
			close(h.incomingMsgCh)
			close(h.dist.telemetryCh)
			return
		}
	}
}

func (h *Hub) RegisterClient(client *Client) {
	h.register <- client // Send client to the hub's register channel
}

func (h *Hub) handleIncomingMessage(msg ClientMessage) {
	switch msg.Client.clientType {
	case SensorTelemetryClient:
		select {
		case h.dist.telemetryCh <- msg:
		default:
			h.l.Warn("Telemetry channel full, dropping message", zap.String("client_id", msg.Client.ID))
		}
		// do for other
	default:
		h.l.Warn("Received message from unknown client type, dropping", zap.String("client_id", msg.Client.ID), zap.String("type", string(msg.Client.clientType)), zap.ByteString("message", msg.Message))
	}
}

func (h *Hub) GetSensorTelemetryMessageChannel() <-chan ClientMessage {
	return h.dist.telemetryCh
}

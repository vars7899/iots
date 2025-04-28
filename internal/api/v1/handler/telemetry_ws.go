package handler

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/vars7899/iots/internal/ws"
	"github.com/vars7899/iots/pkg/apperror"
	"github.com/vars7899/iots/pkg/di"
	"go.uber.org/zap"
)

type TelemetryWebSocketHandler struct {
	deps *di.AppContainer
	hub  *ws.Hub
	l    *zap.Logger
}

func NewTelemetryWebSocketHandler(deps *di.AppContainer, baseLogger *zap.Logger) *TelemetryWebSocketHandler {
	return &TelemetryWebSocketHandler{
		deps: deps,
		hub:  deps.WsHub,
		l:    baseLogger.Named("TelemetryWebSocketHandler"),
	}
}

func (h *TelemetryWebSocketHandler) HandleConnection(c echo.Context) error {
	h.l.Debug("HandleConnection called for /api/v1/sensor/telemetry")
	conn, err := ws.Upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		h.l.Error("websocket upgrade failed", zap.Error(err))
		return apperror.ErrInternal.WithMessage("failed to upgrade to websocket connection").Wrap(err)
	}

	wsClient := ws.NewClient(h.hub, ws.SensorTelemetryClient, conn, h.l.Named("Client"), &ws.WebSocketClientConfig{
		MaxReadLimit: 512,
		PongTimeout:  60 * time.Second,
	})

	h.hub.RegisterClient(wsClient)

	go wsClient.ReadPump()

	return nil
}

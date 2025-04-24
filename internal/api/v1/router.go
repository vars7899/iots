package v1

import (
	"github.com/labstack/echo/v4"
	"github.com/vars7899/iots/internal/api"
	"github.com/vars7899/iots/internal/api/v1/handler"
	"github.com/vars7899/iots/internal/middleware"
	"github.com/vars7899/iots/pkg/di"
	"github.com/vars7899/iots/pkg/logger"
	"go.uber.org/zap"
)

const API_V1_PREFIX = "/api/v1"

func RegisterRoutes(e *echo.Echo, deps *di.Provider, baseLogger *zap.Logger) {
	l := logger.Named(baseLogger, "v1Router")
	r := api.NewAPIRouter(e, API_V1_PREFIX, baseLogger)

	// Middleware
	r.AddMiddleware(middleware.ErrorHandler(l))

	// Routes
	r.AddRoute(api.RouteConfig{
		Prefix:  "/device",
		Handler: handler.NewDeviceHandler(deps, l),
	})
	r.AddRoute(api.RouteConfig{
		Prefix:  "/sensor",
		Handler: handler.NewSensorHandler(deps, l),
	})
	r.AddRoute(api.RouteConfig{
		Prefix:  "/auth",
		Handler: handler.NewAuthHandler(deps, l),
	})

	if deps.WsHub == nil {
		l.Fatal("Websocket Hub is nil in DI provider")
		return
	}

	telemetryWsHandler := handler.NewTelemetryWebSocketHandler(deps, l)

	r.AddWebsocketRoute(api.WsRouteConfig{
		Path:    "/sensor/telemetry",
		Handler: telemetryWsHandler.HandleConnection,
	})
	r.Mount()
}

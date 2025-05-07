package v1

import (
	"github.com/labstack/echo/v4"
	"github.com/vars7899/iots/internal/api"
	"github.com/vars7899/iots/internal/api/v1/handler"
	"github.com/vars7899/iots/internal/middleware"
	"github.com/vars7899/iots/pkg/di"
	"github.com/vars7899/iots/pkg/logger"
	"github.com/vars7899/iots/pkg/websocket"
)

func RegisterRoutes(e *echo.Echo, container *di.AppContainer) {
	if e == nil {
		panic("RegisterRoutes: missing dependency 'echo'")
	}
	if container == nil {
		panic("RegisterRoutes: missing dependency 'container'")
	}
	if container.WsHub == nil {
		panic("RegisterRoutes: missing dependency 'WsHub' for websocket routes")
	}

	logger := logger.Named(container.Logger, "v1Router")
	manager := api.NewAPIRouterManager(e, string(api.ApiV1), logger)

	// Global Middleware
	manager.AddMiddleware(container.Api.Middleware.ErrorHandler)

	// V1 Routes
	manager.AddRoute(api.RouteConfig{
		Prefix:  "/auth",
		Handler: handler.NewAuthHandler(container, logger),
	})
	manager.AddRoute(api.RouteConfig{
		Prefix:  "/device",
		Handler: handler.NewDeviceHandler(container, logger),
		Middleware: []echo.MiddlewareFunc{
			container.Api.Middleware.JWT, container.Api.Middleware.JTI,
		},
	})
	manager.AddRoute(api.RouteConfig{
		Prefix:  "/sensor",
		Handler: handler.NewSensorHandler(container, logger),
		Middleware: []echo.MiddlewareFunc{
			container.Api.Middleware.JWT, container.Api.Middleware.JTI, container.Api.Middleware.AccessControl,
		},
	})
	// V1 Websocket upgraded routes
	manager.AddWebsocketRoute(api.WsRouteConfig{
		Path:    "/sensor/telemetry",
		Handler: handler.NewTelemetryWebSocketHandler(container, logger).HandleConnection,
	})

	// Create WebSocket session manager
	wsManager := websocket.NewSessionManager(logger)

	manager.AddWebsocketRoute(api.WsRouteConfig{
		Path:    "/device/provision",
		Handler: websocket.NewWebsocketHandler(wsManager, container.Config.Websocket, logger).HandleConnection,
		Middleware: []echo.MiddlewareFunc{
			middleware.NewJWTMiddleware(container.CoreServices.AuthTokenService, logger),
			middleware.NewJTIMiddleware(container.CoreServices.AuthTokenService, logger),
		},
	})

	manager.Mount()
}

package api

import (
	"github.com/labstack/echo/v4"
	"github.com/vars7899/iots/pkg/logger"
	"go.uber.org/zap"
)

type RouteConfig struct {
	Prefix     string
	Handler    RouteHandler
	Middleware []echo.MiddlewareFunc
}

type WsRouteConfig struct {
	Path    string
	Handler echo.HandlerFunc
}

type RouteHandler interface {
	SetupRoutes(e *echo.Group)
}

type APIRouterManager struct {
	base        *echo.Group
	logger      *zap.Logger
	middlewares []echo.MiddlewareFunc
	routes      []RouteConfig
	wsRoutes    []WsRouteConfig
}

func NewAPIRouterManager(e *echo.Echo, prefix string, baseLogger *zap.Logger) *APIRouterManager {
	return &APIRouterManager{
		base:        e.Group(prefix),
		logger:      logger.Named(baseLogger, prefix),
		routes:      make([]RouteConfig, 0),
		middlewares: make([]echo.MiddlewareFunc, 0),
	}
}

func (r *APIRouterManager) Mount() {
	r.MountMiddleware()
	r.MountRoutes()
	r.MountWebsockets()
}

func (r *APIRouterManager) MountRoutes() {
	for _, route := range r.routes {
		g := r.base.Group(route.Prefix)
		if len(route.Middleware) > 0 {
			g.Use(route.Middleware...)
		}
		route.Handler.SetupRoutes(g)
	}
	r.logger.Info("http routes mounted", zap.Int("count", len(r.routes)))
}

func (r *APIRouterManager) AddRoute(route RouteConfig) {
	r.routes = append(r.routes, route)
}

func (r *APIRouterManager) MountMiddleware() {
	if len(r.middlewares) > 0 {
		for _, middleware := range r.middlewares {
			r.base.Use(middleware)
		}
	}
	r.logger.Info("middleware mounted", zap.Int("count", len(r.middlewares)))
}

func (r *APIRouterManager) AddMiddleware(middleware echo.MiddlewareFunc) {
	r.middlewares = append(r.middlewares, middleware)
}

func (r *APIRouterManager) AddWebsocketRoute(wsRoute WsRouteConfig) {
	r.wsRoutes = append(r.wsRoutes, wsRoute)
}

func (r *APIRouterManager) MountWebsockets() {
	for _, route := range r.wsRoutes {
		r.base.GET(route.Path, route.Handler)
	}
	r.logger.Info("websocket routes mounted", zap.Int("count", len(r.wsRoutes)))
}

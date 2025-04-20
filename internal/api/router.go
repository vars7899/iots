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

type RouteHandler interface {
	RegisterRoutes(e *echo.Group)
}

type APIRouter struct {
	base        *echo.Group
	logger      *zap.Logger
	middlewares []echo.MiddlewareFunc
	routes      []RouteConfig
}

func NewAPIRouter(e *echo.Echo, prefix string, baseLogger *zap.Logger) *APIRouter {
	return &APIRouter{
		base:        e.Group(prefix),
		logger:      logger.Named(baseLogger, prefix),
		routes:      make([]RouteConfig, 0),
		middlewares: make([]echo.MiddlewareFunc, 0),
	}
}

func (r *APIRouter) Mount() {
	r.MountMiddleware()
	r.MountRoutes()
}

func (r *APIRouter) MountRoutes() {
	for _, route := range r.routes {
		g := r.base.Group(route.Prefix)
		if len(route.Middleware) > 0 {
			g.Use(route.Middleware...)
		}
		route.Handler.RegisterRoutes(g)
	}
	r.logger.Info("routes mounted successfully", zap.Int("count", len(r.routes)))
}

func (r *APIRouter) AddRoute(route RouteConfig) {
	r.routes = append(r.routes, route)
}

func (r *APIRouter) MountMiddleware() {
	if len(r.middlewares) > 0 {
		for _, middleware := range r.middlewares {
			r.base.Use(middleware)
		}
	}
	r.logger.Info("middleware mounted successfully", zap.Int("count", len(r.middlewares)))
}

func (r *APIRouter) AddMiddleware(middleware echo.MiddlewareFunc) {
	r.middlewares = append(r.middlewares, middleware)
}

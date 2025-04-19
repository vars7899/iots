package server

import (
	"github.com/labstack/echo/v4"
	"github.com/vars7899/iots/pkg/logger"
	"go.uber.org/zap"
)

type RouteRegistrar interface {
	RegisterRoutes(e *echo.Echo)
}

type Router struct {
	Registrars []RouteRegistrar
	Logger     *zap.Logger
}

func NewRouter(registrars []RouteRegistrar, baseLogger *zap.Logger) *Router {
	return &Router{Registrars: registrars, Logger: logger.Named(baseLogger, "Router")}
}

func (r *Router) InitRoutes(e *echo.Echo) {
	for _, registrar := range r.Registrars {
		registrar.RegisterRoutes(e)
	}
	r.Logger.Info("All routes module register successfully", zap.Int("count", len(r.Registrars)))
}

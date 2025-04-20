package v1

import (
	"github.com/labstack/echo/v4"
	"github.com/vars7899/iots/internal/api"
	"github.com/vars7899/iots/internal/api/v1/handler"
	"github.com/vars7899/iots/pkg/di"
	"github.com/vars7899/iots/pkg/logger"
	"go.uber.org/zap"
)

const API_V1_PREFIX = "/api/v1"

func RegisterRoutes(e *echo.Echo, deps *di.Provider, baseLogger *zap.Logger) {
	l := logger.Named(baseLogger, "v1Router")
	r := api.NewAPIRouter(e, API_V1_PREFIX, baseLogger)

	r.AddRoute(api.RouteConfig{
		Prefix:  "/device",
		Handler: handler.NewDeviceHandler(deps, l),
	})
	r.AddRoute(api.RouteConfig{
		Prefix:  "/sensor",
		Handler: handler.NewSensorHandler(deps, l),
	})
	r.Mount()
}

// type V1Router struct {
// 	Deps   *di.Provider
// 	Logger *zap.Logger
// 	routes []RouteConfig
// }

// func NewV1Router(deps *di.Provider, baseLogger *zap.Logger) *V1Router {
// 	return &V1Router{Deps: deps, Logger: logger.Named(baseLogger, "V1Router")}
// }

// type RouteConfig struct {
// 	Prefix     string
// 	Handler    handler.RouteHandler
// 	Middleware []echo.MiddlewareFunc
// }

// func (r *V1Router) RegisterRoutes(e *echo.Echo) {
// 	apiV1 := e.Group(API_PREFIX)
// 	apiV1.Use(middleware.ErrorHandler(logger.L()))

// 	routeConfigs := []RouteConfig{

// 		{
// 			Prefix:  "/sensor",
// 			Handler: handler.NewSensorHandler(r.Deps, r.Logger),
// 		},
// 	}

// 	for _, rc := range routeConfigs {
// 		g := apiV1.Group(rc.Prefix)
// 		if len(rc.Middleware) > 0 {
// 			g.Use(rc.Middleware...)
// 		}
// 		rc.Handler.RegisterRoutes(g)
// 	}
// }

// func (r *V1Router) AddRoute(route RouteConfig) {
// 	r.routes = append(r.routes, route)
// }

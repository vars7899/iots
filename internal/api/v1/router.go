package v1

import (
	"github.com/labstack/echo/v4"
	"github.com/vars7899/iots/internal/middleware"
	"github.com/vars7899/iots/pkg/logger"
	"go.uber.org/zap"
)

var (
	ApiPrefix         string = "/api/v1"
	AuthRoutePrefix   string = "/auth"
	UserRoutePrefix   string = "/user"
	DeviceRoutePrefix string = "/device"
	SensorRoutePrefix string = "/sensor"
)

type RouteHandler interface {
	RegisterRoutes(e *echo.Group)
}

type RouteConfig struct {
	Prefix     string
	Handler    RouteHandler
	Middleware []echo.MiddlewareFunc
}

func RegisterRoutes(e *echo.Echo, dep APIDependencies, baseLogger *zap.Logger) {
	apiV1 := e.Group(ApiPrefix)

	// Global middleware
	// api_v1.Use(middleware.Recovery(logger.Lgr))
	apiV1.Use(middleware.ErrorHandler(logger.Lgr))

	RouteConfigs := []RouteConfig{
		{
			Prefix:  DeviceRoutePrefix,
			Handler: NewDeviceHandler(dep, baseLogger),
		},
		{
			Prefix:  SensorRoutePrefix,
			Handler: NewSensorHandler(dep, baseLogger),
		},
	}

	for _, routeConfig := range RouteConfigs {
		g := apiV1.Group(routeConfig.Prefix)
		if len(routeConfig.Middleware) > 0 {
			g.Use(routeConfig.Middleware...)
		}
		routeConfig.Handler.RegisterRoutes(g)
	}
}

// func RegisterAuthRoutes(e *echo.Group, dep APIDependencies) {
// 	authHandler := NewAuthHandler(dep)
// 	authGroup := e.Group(AuthRoutePrefix)
// 	authHandler.RegisterRoutes(authGroup)
// }

// func RegisterUserRoutes(e *echo.Group, dep APIDependencies) {
// 	userHandler := NewUserHandler(dep)
// 	userGroup := e.Group(UserRoutePrefix)
// 	userHandler.RegisterRoutes(userGroup)
// }

// func RegisterSensorRoutes(opts *RouteOpts) {
// 	h := NewSensorHandler(opts.dependencies, opts.baseLogger)
// 	g := opts.e.Group(opts.prefix)
// 	h.RegisterRoutes(g)
// }
// func RegisterDeviceRoutes(opts *RouteOpts) {
// 	h := NewDeviceHandler(opts.dependencies, opts.baseLogger)
// 	g := opts.e.Group(DeviceRoutePrefix)
// 	h.RegisterRoutes(g)
// }

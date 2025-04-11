package v1

import (
	"github.com/labstack/echo/v4"
	"github.com/vars7899/iots/internal/middleware"
	"github.com/vars7899/iots/pkg/logger"
)

var (
	ApiVersionPrefix  string = "/api/v1"
	AuthRoutePrefix   string = "/auth"
	UserRoutePrefix   string = "/user"
	DeviceRoutePrefix string = "/device"
	SensorRoutePrefix string = "/sensor"
)

func RegisterRoutes(e *echo.Echo, dep APIDependencies) {
	api_v1 := e.Group(ApiVersionPrefix)

	api_v1.Use(middleware.Recovery(logger.Lgr))
	api_v1.Use(middleware.ErrorHandler(logger.Lgr))

	RegisterAuthRoutes(api_v1, dep)
	RegisterUserRoutes(api_v1, dep)
	RegisterSensorRoutes(api_v1, dep)
}

func RegisterAuthRoutes(e *echo.Group, dep APIDependencies) {
	authHandler := NewAuthHandler(dep)
	authGroup := e.Group(AuthRoutePrefix)
	authHandler.RegisterRoutes(authGroup)
}

func RegisterSensorRoutes(e *echo.Group, dep APIDependencies) {
	sensorHandler := NewSensorHandler(dep)
	sensorGroup := e.Group(SensorRoutePrefix)
	sensorHandler.RegisterRoutes(sensorGroup)
}

func RegisterUserRoutes(e *echo.Group, dep APIDependencies) {
	userHandler := NewUserHandler(dep)
	userGroup := e.Group(UserRoutePrefix)
	userHandler.RegisterRoutes(userGroup)
}

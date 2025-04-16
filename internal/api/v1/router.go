package v1

import (
	"github.com/labstack/echo/v4"
	"github.com/vars7899/iots/internal/middleware"
	"github.com/vars7899/iots/pkg/logger"
	"go.uber.org/zap"
)

var (
	ApiVersionPrefix  string = "/api/v1"
	AuthRoutePrefix   string = "/auth"
	UserRoutePrefix   string = "/user"
	DeviceRoutePrefix string = "/device"
	SensorRoutePrefix string = "/sensor"
)

func RegisterRoutes(e *echo.Echo, dep APIDependencies, baseLogger *zap.Logger) {
	api_v1 := e.Group(ApiVersionPrefix)

	// api_v1.Use(middleware.Recovery(logger.Lgr))
	api_v1.Use(middleware.ErrorHandler(logger.Lgr))

	RegisterDeviceRoutes(api_v1, dep, baseLogger)
	// RegisterAuthRoutes(api_v1, dep)
	// RegisterUserRoutes(api_v1, dep)
	// RegisterSensorRoutes(api_v1, dep)
}

// func RegisterAuthRoutes(e *echo.Group, dep APIDependencies) {
// 	authHandler := NewAuthHandler(dep)
// 	authGroup := e.Group(AuthRoutePrefix)
// 	authHandler.RegisterRoutes(authGroup)
// }

// func RegisterSensorRoutes(e *echo.Group, dep APIDependencies) {
// 	sensorHandler := NewSensorHandler(dep)
// 	sensorGroup := e.Group(SensorRoutePrefix)
// 	sensorHandler.RegisterRoutes(sensorGroup)
// }

// func RegisterUserRoutes(e *echo.Group, dep APIDependencies) {
// 	userHandler := NewUserHandler(dep)
// 	userGroup := e.Group(UserRoutePrefix)
// 	userHandler.RegisterRoutes(userGroup)
// }

func RegisterDeviceRoutes(e *echo.Group, dep APIDependencies, log *zap.Logger) {
	deviceHandler := NewDeviceHandler(dep, log)
	deviceGroup := e.Group(DeviceRoutePrefix)
	deviceHandler.RegisterRoutes(deviceGroup)
}

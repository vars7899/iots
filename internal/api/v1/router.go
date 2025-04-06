package v1

import (
	"github.com/labstack/echo/v4"
)

func RegisterRoutes(e *echo.Echo, dep APIDependencies) {
	api_v1 := e.Group("/api/v1")

	RegisterSensorRoutes(api_v1, dep)
}

func RegisterSensorRoutes(e *echo.Group, dep APIDependencies) {
	sensorHandler := NewSensorHandler(dep)
	sensorGroup := e.Group("/sensor")
	sensorHandler.RegisterRoutes(sensorGroup)
}

package v1

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/vars7899/iots/internal/domain/sensor"
	"github.com/vars7899/iots/internal/service"
	"github.com/vars7899/iots/pkg/response"
)

type SensorHandler struct {
	SensorService *service.SensorService
}

func NewSensorHandler(dep APIDependencies) *SensorHandler {
	return &SensorHandler{SensorService: dep.SensorService}
}

func (h SensorHandler) RegisterRoutes(e *echo.Group) {
	e.GET("", h.ListSensor)
}

func (h SensorHandler) ListSensor(c echo.Context) error {
	ctx := c.Request().Context()

	// todo: update the sensor filter
	sensorList, err := h.SensorService.ListSensor(ctx, sensor.SensorFilter{})
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, fmt.Sprintf("could not list sensors: %w", err))
	}
	return response.JSON(c, http.StatusOK, sensorList)
}

func (h SensorHandler) GetSensor(c echo.Context) error {
	ctx := c.Request().Context()
	sensorID := c.Param("id")

	sensorData, err := h.SensorService.GetSensor(ctx, sensorID)

	if err != nil {
		if errors.Is(err, sensor.ErrInvalidSensorID) {
			return response.Error(c, http.StatusBadRequest, err.Error())
		}
		if errors.Is(err, sensor.ErrSensorNotFound) {
			return response.Error(c, http.StatusNotFound, err.Error())
		}
		return response.Error(c, http.StatusInternalServerError, "internal server error")
	}

	return response.JSON(c, int(http.StatusOK), sensorData)
}

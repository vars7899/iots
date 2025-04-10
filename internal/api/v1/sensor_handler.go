package v1

import (
	"github.com/labstack/echo/v4"
	"github.com/vars7899/iots/internal/service"
)

type SensorHandler struct {
	SensorService *service.SensorService
}

func NewSensorHandler(dep APIDependencies) *SensorHandler {
	return &SensorHandler{SensorService: dep.SensorService}
}

func (h SensorHandler) RegisterRoutes(e *echo.Group) {
	// e.GET("", h.ListSensor)
	// e.POST("", h.CreateSensor)
	// e.GET("/:id", h.GetSensor)
	// e.DELETE("/:id", h.DeleteSensor)
	// e.PATCH("/:id", h.UpdateSensor)
}

// func (h SensorHandler) ListSensor(c echo.Context) error {
// 	var queryParams sensor.SensorQueryParamsDTO
// 	if err := c.Bind(&queryParams); err != nil {
// 		return response.Error(c, http.StatusBadRequest, "invalid query parameters")
// 	}

// 	if err := queryParams.Validate(); err != nil {
// 		return response.Error(c, http.StatusBadRequest, err.Error())
// 	}

// 	filter, err := queryParams.ToFilter()
// 	if err != nil {
// 		return response.Error(c, http.StatusBadRequest, "invalid filter parameters")
// 	}

// 	sensorList, err := h.SensorService.ListSensor(c.Request().Context(), filter)
// 	if err != nil {
// 		return h.handlerError(c, err)
// 	}

// 	return response.JSON(c, http.StatusOK, sensorList)
// }

// func (h SensorHandler) GetSensor(c echo.Context) error {
// 	ctx := c.Request().Context()
// 	sensorID := c.Param("id")

// 	sensorData, err := h.SensorService.GetSensor(ctx, sensorID)

// 	if err != nil {
// 		return h.handlerError(c, err)
// 	}
// 	return response.JSON(c, int(http.StatusOK), sensorData)
// }

// func (h SensorHandler) DeleteSensor(c echo.Context) error {
// 	ctx := c.Request().Context()
// 	sensorID := c.Param("id")

// 	err := h.SensorService.DeleteSensor(ctx, sensorID)
// 	if err != nil {
// 		return h.handlerError(c, err)
// 	}
// 	return response.JSON(c, http.StatusOK, map[string]interface{}{
// 		"sensor_id": sensorID,
// 		"message":   "sensor deleted successfully",
// 	})
// }

// func (h SensorHandler) CreateSensor(c echo.Context) error {
// 	var dto sensor.CreateSensorDTO
// 	if err := c.Bind(&dto); err != nil {
// 		return response.Error(c, http.StatusBadRequest, "invalid request body")
// 	}

// 	if err := dto.Validate(); err != nil {
// 		return response.Error(c, http.StatusBadRequest, err.Error())
// 	}

// 	_sensorData := dto.ToSensorModel()
// 	if err := h.SensorService.CreateSensor(c.Request().Context(), &_sensorData); err != nil {
// 		h.handlerError(c, err)
// 	}

// 	return response.JSON(c, http.StatusCreated, map[string]interface{}{
// 		"message": "sensor created successfully",
// 	})
// }

// func (h SensorHandler) UpdateSensor(c echo.Context) error {
// 	var dto sensor.UpdateSensorDTO
// 	if err := c.Bind(&dto); err != nil {
// 		return response.Error(c, http.StatusBadRequest, "invalid request body")

// 	}
// 	if err := dto.Validate(); err != nil {
// 		return response.Error(c, http.StatusBadRequest, err.Error())
// 	}

// 	sensorID := c.Param("id")

// 	// Fetch the existing sensor
// 	existing, err := h.SensorService.GetSensor(c.Request().Context(), sensorID)
// 	if err != nil {
// 		return h.handlerError(c, err)
// 	}

// 	dto.ApplyUpdates(existing)

// 	if err := h.SensorService.UpdateSensor(c.Request().Context(), existing); err != nil {
// 		return h.handlerError(c, err)
// 	}

// 	return response.JSON(c, http.StatusOK, map[string]interface{}{
// 		"sensor_id": sensorID,
// 		"message":   "sensor updated successfully",
// 	})
// }

// func (h SensorHandler) handlerError(c echo.Context, err error) error {
// 	if errors.Is(err, sensor.ErrInvalidSensorID) {
// 		return response.Error(c, http.StatusBadRequest, err.Error())
// 	}
// 	if errors.Is(err, sensor.ErrSensorNotFound) {
// 		return response.Error(c, http.StatusNotFound, err.Error())
// 	}
// 	return response.Error(c, http.StatusInternalServerError, "internal server error")
// }

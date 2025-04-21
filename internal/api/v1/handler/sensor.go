package handler

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/vars7899/iots/internal/api/v1/dto"
	"github.com/vars7899/iots/internal/domain"
	"github.com/vars7899/iots/internal/service"
	"github.com/vars7899/iots/pkg/apperror"
	"github.com/vars7899/iots/pkg/di"
	"github.com/vars7899/iots/pkg/logger"
	"github.com/vars7899/iots/pkg/response"
	"github.com/vars7899/iots/pkg/utils"
	"go.uber.org/zap"
)

type SensorHandler struct {
	SensorService *service.SensorService
	logger        *zap.Logger
}

func NewSensorHandler(dep *di.Provider, baseLogger *zap.Logger) *SensorHandler {
	return &SensorHandler{SensorService: dep.Services.SensorService, logger: logger.Named(baseLogger, "SensorHandler")}
}

func (h SensorHandler) RegisterRoutes(e *echo.Group) {
	e.POST("", h.CreateSensor)
	e.GET("", h.ListSensor)
	e.GET("/:id", h.GetSensor)
	e.DELETE("/:id", h.DeleteSensor)
	e.PATCH("/:id", h.UpdateSensor)
}

func (h SensorHandler) CreateSensor(c echo.Context) error {
	var dto dto.CreateSensorDTO
	reqPath := utils.GetRequestUrlPath(c)

	// Bind body request and validate fields
	if err := utils.BindAndValidate(c, &dto); err != nil {
		return err
	}

	createdSensor, err := h.SensorService.CreateSensor(c.Request().Context(), dto.AsModel())
	if err != nil {
		return apperror.ErrorHandler(err, apperror.ErrCodeDBInsert, fmt.Sprintf("failed to create %s", domain.EntitySensor)).WithPath(reqPath).Wrap(err)
	}

	return response.JSON(c, http.StatusCreated, echo.Map{
		"message": "sensor created successfully",
		"sensor":  createdSensor,
	})
}

func (h SensorHandler) GetSensor(c echo.Context) error {
	reqID := c.Param("id")
	reqPath := utils.GetRequestUrlPath(c)

	sensorID, err := uuid.Parse(reqID)
	if err != nil {
		return apperror.ErrBadRequest.WithMessagef("invalid %s ID format", domain.EntitySensor).WithDetails(echo.Map{
			"sensor_id": reqID,
			"error":     err.Error(),
		}).WithPath(reqPath).Wrap(err)
	}

	sensorData, err := h.SensorService.GetSensor(c.Request().Context(), sensorID)
	if err != nil {
		fmt.Println(err)
		return apperror.ErrorHandler(err, apperror.ErrCodeDBQuery, fmt.Sprintf("failed to retrieve %s with ID %s", domain.EntitySensor, reqID)).WithPath(reqPath)
	}
	return response.JSON(c, int(http.StatusOK), echo.Map{
		"sensor": sensorData,
	})
}

func (h SensorHandler) DeleteSensor(c echo.Context) error {
	reqID := c.Param("id")
	reqPath := utils.GetRequestUrlPath(c)

	sensorID, err := uuid.Parse(reqID)
	if err != nil {
		return apperror.ErrBadRequest.WithMessagef("invalid %s ID format", domain.EntitySensor).WithDetails(echo.Map{
			"sensor_id": reqID,
			"error":     err.Error(),
		}).WithPath(reqPath).Wrap(err)
	}

	if err := h.SensorService.DeleteSensor(c.Request().Context(), sensorID); err != nil {
		return apperror.ErrorHandler(err, apperror.ErrCodeDBDelete, fmt.Sprintf("failed to delete %s with ID %s", domain.EntitySensor, reqID)).WithPath(reqPath)
	}
	return response.JSON(c, http.StatusOK, map[string]interface{}{
		"sensor_id": sensorID,
		"message":   "sensor deleted successfully",
	})
}

func (h SensorHandler) UpdateSensor(c echo.Context) error {
	var dto dto.UpdateSensorDTO

	// Bind body request and validate fields
	if err := utils.BindAndValidate(c, &dto); err != nil {
		return err
	}

	reqID := c.Param("id")
	reqPath := utils.GetRequestUrlPath(c)

	sensorID, err := uuid.Parse(reqID)
	if err != nil {
		return apperror.ErrBadRequest.WithMessagef("invalid %s ID format", domain.EntitySensor).WithDetails(echo.Map{
			"sensor_id": reqID,
			"error":     err.Error(),
		}).WithPath(reqPath).Wrap(err)
	}

	sensorUpdates := dto.AsModel()
	sensorUpdates.ID = sensorID // bind id

	sensorData, err := h.SensorService.UpdateSensor(c.Request().Context(), sensorUpdates)
	if err != nil {
		return apperror.ErrorHandler(err, apperror.ErrCodeDBUpdate, fmt.Sprintf("failed to update %s with ID %s", domain.EntitySensor, reqID)).WithPath(reqPath)
	}

	return response.JSON(c, http.StatusOK, map[string]interface{}{
		"message": "sensor updated successfully",
		"sensor":  sensorData,
	})
}

func (h SensorHandler) ListSensor(c echo.Context) error {
	var dto dto.SensorQueryParamsDTO
	reqPath := utils.GetRequestUrlPath(c)

	// Bind body request and validate fields
	if err := utils.BindAndValidate(c, &dto); err != nil {
		return err
	}

	paginationConfig, filterParams, err := dto.AsModel()
	if err != nil {
		return apperror.ErrorHandler(err, apperror.ErrCodeDBQuery, fmt.Sprintf("failed to list %s", domain.EntitySensor)).WithPath(reqPath)
	}

	sensorList, err := h.SensorService.ListSensor(c.Request().Context(), filterParams, paginationConfig)
	if err != nil {
		return apperror.ErrorHandler(err, apperror.ErrCodeDBQuery, fmt.Sprintf("failed to list %s", domain.EntitySensor)).WithPath(reqPath)
	}

	return response.JSON(c, http.StatusOK, sensorList)
}

// func (h SensorHandler) handlerError(c echo.Context, err error) error {
// 	if errors.Is(err, sensor.ErrInvalidSensorID) {
// 		return response.Error(c, http.StatusBadRequest, err.Error())
// 	}
// 	if errors.Is(err, sensor.ErrSensorNotFound) {
// 		return response.Error(c, http.StatusNotFound, err.Error())
// 	}
// 	return response.Error(c, http.StatusInternalServerError, "internal server error")
// }

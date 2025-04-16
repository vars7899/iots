package v1

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/vars7899/iots/internal/api/v1/dto"
	"github.com/vars7899/iots/internal/service"
	"github.com/vars7899/iots/pkg/apperror"
	"github.com/vars7899/iots/pkg/logger"
	"github.com/vars7899/iots/pkg/response"
	"go.uber.org/zap"
)

type DeviceHandler struct {
	DeviceService *service.DeviceService
	log           *zap.Logger
}

func NewDeviceHandler(dep APIDependencies, baseLogger *zap.Logger) *DeviceHandler {
	if dep.DeviceService == nil {
		panic("missing dependency")
	}
	return &DeviceHandler{
		DeviceService: dep.DeviceService,
		log:           logger.NewNamedZapLogger(baseLogger, "handler.deviceHandler"),
	}
}

func (h *DeviceHandler) RegisterRoutes(e *echo.Group) {
	e.POST("", h.CreateNewDevice)
}

func (h *DeviceHandler) CreateNewDevice(c echo.Context) error {
	var dto dto.CreateNewDeviceDTO
	if err := c.Bind(&dto); err != nil {
		return apperror.ErrBadRequest.WithDetails(echo.Map{
			"error": "invalid request body",
		}).Wrap(err)
	}
	if err := dto.Validate(); err != nil {
		return apperror.ErrValidation.WithDetails(echo.Map{
			"error":   "validation error",
			"details": err.Error(),
		}).Wrap(err)
	}

	d := dto.ToDevice()

	deviceData, err := h.DeviceService.CreateDevice(c.Request().Context(), d)
	if err != nil {
		return apperror.ErrDBInsert.WithDetails(echo.Map{
			"details": err.Error(),
		}).Wrap(err)
	}

	return response.JSON(c, int(http.StatusCreated), echo.Map{
		"message": "device created successfully",
		"device":  deviceData,
	})
}

func (h *DeviceHandler) GetDeviceByID(c echo.Context) error {
	reqID := c.Param("id")

	deviceID, err := uuid.Parse(reqID)
	if err != nil {
		return apperror.ErrBadRequest.WithMessage("validation failed, invalid device id").WithDetails(echo.Map{
			"device_id": reqID,
		}).Wrap(err)
	}

	deviceExist, err := h.DeviceService.GetDeviceByID(c.Request().Context(), deviceID)
	if err != nil {
		return apperror.ErrDBQuery.WithDetails(echo.Map{
			"error": err.Error(),
		}).Wrap(err)
	}

	return response.JSON(c, int(http.StatusOK), echo.Map{
		"device": deviceExist,
	})

}

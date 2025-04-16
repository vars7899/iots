package v1

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/vars7899/iots/internal/api/v1/dto"
	"github.com/vars7899/iots/internal/service"
	"github.com/vars7899/iots/pkg/apperror"
	"github.com/vars7899/iots/pkg/logger"
	"github.com/vars7899/iots/pkg/response"
	"github.com/vars7899/iots/pkg/utils"
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
	e.GET("/:id", h.GetDeviceByID)
	e.DELETE("/:id", h.DeleteDeviceByID)
	e.PATCH("/:id", h.UpdateDevice)
}

func (h *DeviceHandler) CreateNewDevice(c echo.Context) error {
	var dto dto.CreateNewDeviceDTO
	reqPath := utils.GetRequestUrlPath(c)

	// Bind body request and validate fields
	if err := utils.BindAndValidate(c, &dto); err != nil {
		return err
	}

	createdDevice, err := h.DeviceService.CreateDevice(c.Request().Context(), dto.ToDevice())
	if err != nil {
		return apperror.ErrorHandler(err, apperror.ErrCodeDBInsert, "failed to insert device").WithPath(reqPath).Wrap(err)
	}

	h.log.Debug("device created", zap.String("device_id", createdDevice.ID.String()))
	return response.JSON(c, int(http.StatusCreated), echo.Map{
		"message": "device created successfully",
		"device":  createdDevice,
	})
}

func (h *DeviceHandler) GetDeviceByID(c echo.Context) error {
	reqID := c.Param("id")
	reqPath := utils.GetRequestUrlPath(c)

	deviceID, err := uuid.Parse(reqID)
	if err != nil {
		return apperror.ErrBadRequest.WithMessage("invalid device ID format").WithDetails(echo.Map{
			"error": err.Error(),
		}).WithPath(reqPath).Wrap(err)
	}

	deviceExist, err := h.DeviceService.GetDeviceByID(c.Request().Context(), deviceID)
	if err != nil {
		return apperror.ErrorHandler(err, apperror.ErrCodeDBQuery, fmt.Sprintf("failed to retrieve device with ID %s", reqID)).WithPath(reqPath)
	}

	h.log.Debug("device query executed", zap.String("device_id", deviceExist.ID.String()))
	return response.JSON(c, int(http.StatusOK), echo.Map{
		"device": deviceExist,
	})
}

func (h *DeviceHandler) DeleteDeviceByID(c echo.Context) error {
	reqID := c.Param("id")
	reqPath := utils.GetRequestUrlPath(c)

	deviceID, err := uuid.Parse(reqID)
	if err != nil {
		return apperror.ErrBadRequest.WithMessage("invalid device ID format").WithDetails(echo.Map{
			"error": err.Error(),
		}).WithPath(reqPath).Wrap(err)
	}

	if err := h.DeviceService.DeleteDevice(c.Request().Context(), deviceID); err != nil {
		return apperror.ErrorHandler(err, apperror.ErrCodeInternal, fmt.Sprintf("failed to delete device with ID %s", reqID)).WithPath(reqPath)
	}
	h.log.Debug("device deleted", zap.String("device_id", deviceID.String()))
	return response.JSON(c, int(http.StatusOK), echo.Map{
		"device_id": deviceID,
	})
}

func (h *DeviceHandler) UpdateDevice(c echo.Context) error {
	var dto dto.UpdateDeviceDTO
	reqID := c.Param("id")
	reqPath := utils.GetRequestUrlPath(c)

	deviceID, err := uuid.Parse(reqID)
	if err != nil {
		return apperror.ErrBadRequest.WithMessage("invalid device ID format").WithDetails(echo.Map{
			"error": err.Error(),
		}).WithPath(reqPath).Wrap(err)
	}

	// Bind body request and validate fields
	if err := utils.BindAndValidate(c, &dto); err != nil {
		return err
	}

	updatedDevice, err := h.DeviceService.UpdateDevice(c.Request().Context(), deviceID, dto.ToDevice())
	if err != nil {
		return apperror.ErrorHandler(err, apperror.ErrCodeDBUpdate, fmt.Sprintf("failed to update device with ID %s", reqID)).WithPath(reqPath)
	}

	h.log.Debug("device updated", zap.String("device_id", deviceID.String()))
	return response.JSON(c, int(http.StatusOK), echo.Map{
		"device": updatedDevice,
	})
}

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

type DeviceHandler struct {
	DeviceService *service.DeviceService
	log           *zap.Logger
}

func NewDeviceHandler(dep *di.Provider, baseLogger *zap.Logger) *DeviceHandler {
	return &DeviceHandler{
		DeviceService: dep.Services.DeviceService,
		log:           logger.Named(baseLogger, "handler.deviceHandler"),
	}
}

func (h *DeviceHandler) RegisterRoutes(e *echo.Group) {
	e.POST("", h.CreateNewDevice)
	e.GET("/:id", h.GetDeviceByID)
	e.DELETE("/:id", h.DeleteDeviceByID)
	e.PATCH("/:id", h.UpdateDevice)
	// bulk operation endpoints
	// TODO: add middleware to protect only for admin
	e.POST("/bulk", h.CreateNewDeviceInBulk)
	e.DELETE("/bulk", h.DeleteDeviceInBulk)
	e.PATCH("/bulk", h.UpdateDeviceInBulk)
	// status endpoints
	e.PATCH("/:id/status", h.UpdateDeviceStatus)
	e.PATCH("/:id/online", h.MarkDeviceOnline)
	e.PATCH("/:id/offline", h.MarkDeviceOffline)

}

func (h *DeviceHandler) CreateNewDevice(c echo.Context) error {
	var dto dto.CreateNewDeviceDTO
	reqPath := utils.GetRequestUrlPath(c)

	// Bind body request and validate fields
	if err := utils.BindAndValidate(c, &dto); err != nil {
		return err
	}

	createdDevice, err := h.DeviceService.CreateDevice(c.Request().Context(), dto.AsModel())
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

	deviceUpdates := dto.ToDevice()
	deviceUpdates.ID = deviceID

	updatedDevice, err := h.DeviceService.UpdateDevice(c.Request().Context(), deviceUpdates)
	if err != nil {
		return apperror.ErrorHandler(err, apperror.ErrCodeDBUpdate, fmt.Sprintf("failed to update device with ID %s", reqID)).WithPath(reqPath)
	}

	h.log.Debug("device updated", zap.String("device_id", deviceID.String()))
	return response.JSON(c, int(http.StatusOK), echo.Map{
		"device": updatedDevice,
	})
}

func (h *DeviceHandler) CreateNewDeviceInBulk(c echo.Context) error {
	var dto dto.BulkCreateDevicesDTO
	reqPath := utils.GetRequestUrlPath(c)

	if err := utils.BindAndValidate(c, &dto); err != nil {
		return err
	}

	createdDevices, err := h.DeviceService.BulkCreateDevices(c.Request().Context(), dto.ToDevices())
	if err != nil {
		return apperror.ErrorHandler(err, apperror.ErrCodeDBInsert, "failed to bulk insert devices").WithPath(reqPath)
	}

	h.log.Debug("bulk device creation successful", zap.Int("count", len(createdDevices)))
	return response.JSON(c, http.StatusCreated, echo.Map{
		"message": "devices created successfully",
		"devices": createdDevices,
	})
}

func (h *DeviceHandler) DeleteDeviceInBulk(c echo.Context) error {
	var dto dto.BulkDeleteDeviceDTO
	reqPath := utils.GetRequestUrlPath(c)

	if err := utils.BindAndValidate(c, &dto); err != nil {
		return err
	}

	uuids, err := dto.ToUUIDs()
	if err != nil {
		return apperror.ErrBadRequest.WithMessage("invalid UUID format").WithPath(reqPath).Wrap(err)
	}

	if err := h.DeviceService.BulkDeleteDevices(c.Request().Context(), uuids); err != nil {
		return apperror.ErrorHandler(err, apperror.ErrCodeDBDelete, "failed to bulk delete devices").
			WithPath(reqPath).Wrap(err)
	}

	h.log.Debug("bulk device deletion successful", zap.Int("count", len(uuids)))
	return response.JSON(c, http.StatusOK, echo.Map{
		"message":    "devices deleted successfully",
		"device_ids": dto.DeviceIDs,
	})
}

func (h *DeviceHandler) UpdateDeviceInBulk(c echo.Context) error {
	var dto dto.BulkUpdateDeviceDTO
	reqPath := utils.GetRequestUrlPath(c)

	if err := utils.BindAndValidate(c, &dto); err != nil {
		return err
	}

	updatedDevices, err := h.DeviceService.BulkUpdateDevices(c.Request().Context(), dto.ToDevices())
	if err != nil {
		return apperror.ErrorHandler(err, apperror.ErrCodeDBUpdate, "").WithPath(reqPath).Wrap(err)
	}

	return response.JSON(c, int(http.StatusOK), echo.Map{
		"message": "devices updated successfully",
		"devices": updatedDevices,
	})
}

func (h *DeviceHandler) UpdateDeviceStatus(c echo.Context) error {
	paramID := c.Param("id")
	reqPath := utils.GetRequestUrlPath(c)

	deviceID, err := uuid.Parse(paramID)
	if err != nil {
		return apperror.ErrInvalidUUID.WithMessage("invalid uuid format").WithPath(reqPath).Wrap(err)
	}

	var dto dto.UpdateDeviceStatusDTO
	if err := utils.BindAndValidate(c, &dto); err != nil {
		return err
	}

	err = h.DeviceService.UpdateDeviceStatus(c.Request().Context(), deviceID, domain.Status(dto.Status))
	if err != nil {
		return apperror.ErrorHandler(err, apperror.ErrCodeDBUpdate)
	}

	return response.JSON(c, int(http.StatusOK), echo.Map{
		"message": fmt.Sprintf("device with ID %s status updated to %s", deviceID.String(), dto.Status),
	})
}
func (h *DeviceHandler) MarkDeviceOnline(c echo.Context) error {
	paramID := c.Param("id")
	reqPath := utils.GetRequestUrlPath(c)

	deviceID, err := uuid.Parse(paramID)
	if err != nil {
		return apperror.ErrInvalidUUID.WithMessage("invalid uuid format").WithPath(reqPath).Wrap(err)
	}

	updatedDevice, err := h.DeviceService.MarkDeviceAsOnline(c.Request().Context(), deviceID)
	if err != nil {
		return apperror.ErrorHandler(err, apperror.ErrCodeDBUpdate)
	}

	return response.JSON(c, int(http.StatusOK), echo.Map{
		"message": fmt.Sprintf("device with ID %s marked as online", deviceID.String()),
		"device":  updatedDevice,
	})
}
func (h *DeviceHandler) MarkDeviceOffline(c echo.Context) error {
	paramID := c.Param("id")
	reqPath := utils.GetRequestUrlPath(c)

	deviceID, err := uuid.Parse(paramID)
	if err != nil {
		return apperror.ErrInvalidUUID.WithMessage("invalid uuid format").WithPath(reqPath).Wrap(err)
	}

	updatedDevice, err := h.DeviceService.MarkDeviceAsOffline(c.Request().Context(), deviceID)
	if err != nil {
		return apperror.ErrorHandler(err, apperror.ErrCodeDBUpdate)
	}

	return response.JSON(c, int(http.StatusOK), echo.Map{
		"message": fmt.Sprintf("device with ID %s marked as offline", deviceID.String()),
		"device":  updatedDevice,
	})
}

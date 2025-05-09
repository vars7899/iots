package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/vars7899/iots/config"
	"github.com/vars7899/iots/internal/api/v1/dto"
	"github.com/vars7899/iots/internal/middleware"
	"github.com/vars7899/iots/internal/service"
	"github.com/vars7899/iots/pkg/apperror"
	"github.com/vars7899/iots/pkg/auth/deviceauth"
	"github.com/vars7899/iots/pkg/contextkey"
	"github.com/vars7899/iots/pkg/di"
	"github.com/vars7899/iots/pkg/logger"
	"github.com/vars7899/iots/pkg/response"
	"github.com/vars7899/iots/pkg/utils"
	"go.uber.org/zap"
)

type DeviceHandler struct {
	DeviceService service.DeviceService
	middleware    *middleware.MiddlewareRegistry
	logger        *zap.Logger
}

func NewDeviceHandler(container *di.AppContainer, baseLogger *zap.Logger) *DeviceHandler {
	return &DeviceHandler{
		DeviceService: container.Services.DeviceService,
		middleware:    container.Api.Middleware,
		logger:        logger.Named(baseLogger, "DeviceHandler"),
	}
}

func (h *DeviceHandler) SetupRoutes(e *echo.Group) {
	// e.POST("", h.CreateNewDevice)
	// e.GET("/:id", h.GetDeviceByID)
	// e.DELETE("/:id", h.DeleteDeviceByID)
	// e.PATCH("/:id", h.UpdateDevice)

	e.POST("/register", h.RegisterDevice, h.middleware.PermissionRequired("device", "register"))

	// bulk operation endpoints
	// TODO: add middleware to protect only for admin
	// e.POST("/bulk", h.CreateNewDeviceInBulk)
	// e.DELETE("/bulk", h.DeleteDeviceInBulk)
	// e.PATCH("/bulk", h.UpdateDeviceInBulk)
	// status endpoints
	// e.PATCH("/:id/status", h.UpdateDeviceStatus)
	// e.PATCH("/:id/online", h.MarkDeviceOnline)
	// e.PATCH("/:id/offline", h.MarkDeviceOffline)

	// Provision flow
	e.POST("/provision", h.ProvisionDevice, h.middleware.PermissionRequired("device", "provision"))
	// e.GET("/connect", h.UpgradeToDeviceSession, h.middleware.PermissionRequired("device", "session"))
	e.POST("/session/refresh", h.RefreshSessionToken, h.middleware.PermissionRequired("device", "session_refresh"))

}

func (h *DeviceHandler) ProvisionDevice(c echo.Context) error {
	var dto dto.ProvisionDeviceRequest
	ctx := c.Request().Context()

	if err := utils.BindAndValidate(c, &dto); err != nil {
		return err
	}

	connTokens, err := h.DeviceService.ProvisionDevice(ctx, dto.DeviceID, dto.ProvisionCode)
	if err != nil {
		return err
	}
	if connTokens == nil {
		return apperror.ErrInternal.WithMessage("something went wrong while generating connection token")
	}

	bindDeviceConnectionHeaders(c, connTokens)

	responsePayload := echo.Map{
		"message": "device provisioned successfully",
	}
	if !config.InProd() {
		responsePayload["device_session"] = connTokens
	}
	return response.JSON(c, http.StatusCreated, responsePayload)
}

// Device                        Server (Echo)
//   |                                	|
//   |--- HTTP GET /connect ----------->|   ← WebSocket upgrade endpoint
//   |   Headers:                     	|
//   |     Connection-Token          	|
//   |     Refresh-Token (optional)  	|
//   |                                	|
//   |<--- Validate Tokens -------------|
//   |     - if valid:                	|
//   |         upgrade to WebSocket   	|
//   |     - if expired:              	|
//   |         try refresh           	|
//   |         - if refresh valid:   	|
//   |             upgrade to WS     	|
//   |         - else: reject        	|
//   |                                	|
//   |--- WebSocket established ------->|
//   |                                	|
//   |    (start two-way comm)       	|

// func (h *DeviceHandler) UpgradeToDeviceSession(c echo.Context) error {
// 	req := c.Request()
// 	ctx := req.Context()

// 	connectionToken := req.Header.Get(contextkey.HeaderDeviceConnectionToken)
// 	refreshToken := req.Header.Get(contextkey.HeaderDeviceRefreshToken)

// 	if connectionToken == "" {
// 		return apperror.ErrInvalidToken.WithMessage("missing session connection tokens")
// 	}

// 	return nil
// }

func (h *DeviceHandler) RefreshSessionToken(c echo.Context) error {
	req := c.Request()
	ctx := req.Context()

	connectionTokenStr := req.Header.Get(contextkey.HeaderDeviceConnectionToken)
	refreshTokenStr := req.Header.Get(contextkey.HeaderDeviceRefreshToken)

	if connectionTokenStr == "" || refreshTokenStr == "" {
		return apperror.ErrInvalidToken.WithMessage("invalid or missing device session token")
	}

	connTokens, err := h.DeviceService.RefreshDeviceTokens(ctx, connectionTokenStr, refreshTokenStr)
	if err != nil {
		return apperror.ErrorHandler(err, apperror.ErrCodeInternal)
	}

	bindDeviceConnectionHeaders(c, connTokens)

	responsePayload := echo.Map{
		"message": "device session token refreshed",
	}
	if !config.InProd() {
		responsePayload["device_session"] = connTokens
	}

	return response.JSON(c, http.StatusOK, responsePayload)
}

func bindDeviceConnectionHeaders(c echo.Context, connTokens *deviceauth.DeviceConnectionTokens) {
	c.Response().Header().Set(contextkey.HeaderDeviceConnectionToken, connTokens.ConnectionToken)
	c.Response().Header().Set(contextkey.HeaderDeviceRefreshToken, connTokens.RefreshToken)
}

// func (h *DeviceHandler) CreateNewDevice(c echo.Context) error {
// 	var dto dto.CreateNewDeviceDTO
// 	reqPath := utils.GetRequestUrlPath(c)

// 	// Bind body request and validate fields
// 	if err := utils.BindAndValidate(c, &dto); err != nil {
// 		return err
// 	}

// 	createdDevice, err := h.DeviceService.CreateDevice(c.Request().Context(), dto.AsModel())
// 	if err != nil {
// 		return apperror.ErrorHandler(err, apperror.ErrCodeDBInsert, "failed to insert device").WithPath(reqPath).Wrap(err)
// 	}

// 	h.log.Debug("device created", zap.String("device_id", createdDevice.ID.String()))
// 	return response.JSON(c, int(http.StatusCreated), echo.Map{
// 		"message": "device created successfully",
// 		"device":  createdDevice,
// 	})
// }

// func (h *DeviceHandler) GetDeviceByID(c echo.Context) error {
// 	reqID := c.Param("id")
// 	reqPath := utils.GetRequestUrlPath(c)

// 	deviceID, err := uuid.Parse(reqID)
// 	if err != nil {
// 		return apperror.ErrBadRequest.WithMessage("invalid device ID format").WithDetails(echo.Map{
// 			"error": err.Error(),
// 		}).WithPath(reqPath).Wrap(err)
// 	}

// 	deviceExist, err := h.DeviceService.GetDeviceByID(c.Request().Context(), deviceID)
// 	if err != nil {
// 		return apperror.ErrorHandler(err, apperror.ErrCodeDBQuery, fmt.Sprintf("failed to retrieve device with ID %s", reqID)).WithPath(reqPath)
// 	}

// 	h.log.Debug("device query executed", zap.String("device_id", deviceExist.ID.String()))
// 	return response.JSON(c, int(http.StatusOK), echo.Map{
// 		"device": deviceExist,
// 	})
// }

// func (h *DeviceHandler) DeleteDeviceByID(c echo.Context) error {
// 	reqID := c.Param("id")
// 	reqPath := utils.GetRequestUrlPath(c)

// 	deviceID, err := uuid.Parse(reqID)
// 	if err != nil {
// 		return apperror.ErrBadRequest.WithMessage("invalid device ID format").WithDetails(echo.Map{
// 			"error": err.Error(),
// 		}).WithPath(reqPath).Wrap(err)
// 	}

// 	if err := h.DeviceService.DeleteDevice(c.Request().Context(), deviceID); err != nil {
// 		return apperror.ErrorHandler(err, apperror.ErrCodeInternal, fmt.Sprintf("failed to delete device with ID %s", reqID)).WithPath(reqPath)
// 	}
// 	h.log.Debug("device deleted", zap.String("device_id", deviceID.String()))
// 	return response.JSON(c, int(http.StatusOK), echo.Map{
// 		"device_id": deviceID,
// 	})
// }

// func (h *DeviceHandler) UpdateDevice(c echo.Context) error {
// 	var dto dto.UpdateDeviceDTO
// 	reqID := c.Param("id")
// 	reqPath := utils.GetRequestUrlPath(c)

// 	deviceID, err := uuid.Parse(reqID)
// 	if err != nil {
// 		return apperror.ErrBadRequest.WithMessage("invalid device ID format").WithDetails(echo.Map{
// 			"error": err.Error(),
// 		}).WithPath(reqPath).Wrap(err)
// 	}

// 	// Bind body request and validate fields
// 	if err := utils.BindAndValidate(c, &dto); err !=  -->nil {
// 		return err
// 	}

// 	deviceUpdates := dto.ToDevice()
// 	deviceUpdates.ID = deviceID

// 	updatedDevice, err := h.DeviceService.UpdateDevice(c.Request().Context(), deviceUpdates)
// 	if err != nil {
// 		return apperror.ErrorHandler(err, apperror.ErrCodeDBUpdate, fmt.Sprintf("failed to update device with ID %s", reqID)).WithPath(reqPath)
// 	}

// 	h.log.Debug("device updated", zap.String("device_id", deviceID.String()))
// 	return response.JSON(c, int(http.StatusOK), echo.Map{
// 		"device": updatedDevice,
// 	})
// }

func (h *DeviceHandler) RegisterDevice(c echo.Context) error {
	var dto dto.RegisterDeviceRequest

	ctx := c.Request().Context()
	path := utils.GetRequestUrlPath(c)

	if err := utils.BindAndValidate(c, &dto); err != nil {
		return err
	}

	device, err := h.DeviceService.CreateDevice(ctx, dto.AsModel())
	if err != nil {
		return apperror.ErrorHandler(err, apperror.ErrCodeDBUpdate, "failed to register new device").WithPath(path)
	}

	responsePayload := echo.Map{
		"message": "device registered successfully",
	}
	if !config.InProd() {
		responsePayload["device"] = device
	}
	return response.JSON(c, http.StatusCreated, responsePayload)

}

// func (h *DeviceHandler) CreateNewDeviceInBulk(c echo.Context) error {
// 	var dto dto.BulkCreateDevicesDTO
// 	reqPath := utils.GetRequestUrlPath(c)

// 	if err := utils.BindAndValidate(c, &dto); err != nil {
// 		return err
// 	}

// 	createdDevices, err := h.DeviceService.BulkCreateDevices(c.Request().Context(), dto.ToDevices())
// 	if err != nil {
// 		return apperror.ErrorHandler(err, apperror.ErrCodeDBInsert, "failed to bulk insert devices").WithPath(reqPath)
// 	}

// 	h.log.Debug("bulk device creation successful", zap.Int("count", len(createdDevices)))
// 	return response.JSON(c, http.StatusCreated, echo.Map{
// 		"message": "devices created successfully",
// 		"devices": createdDevices,
// 	})
// }

// func (h *DeviceHandler) DeleteDeviceInBulk(c echo.Context) error {
// 	var dto dto.BulkDeleteDeviceDTO
// 	reqPath := utils.GetRequestUrlPath(c)

// 	if err := utils.BindAndValidate(c, &dto); err != nil {
// 		return err
// 	}

// 	uuids, err := dto.ToUUIDs()
// 	if err != nil {
// 		return apperror.ErrBadRequest.WithMessage("invalid UUID format").WithPath(reqPath).Wrap(err)
// 	}

// 	if err := h.DeviceService.BulkDeleteDevices(c.Request().Context(), uuids); err != nil {
// 		return apperror.ErrorHandler(err, apperror.ErrCodeDBDelete, "failed to bulk delete devices").
// 			WithPath(reqPath).Wrap(err)
// 	}

// 	h.log.Debug("bulk device deletion successful", zap.Int("count", len(uuids)))
// 	return response.JSON(c, http.StatusOK, echo.Map{
// 		"message":    "devices deleted successfully",
// 		"device_ids": dto.DeviceIDs,
// 	})
// }

// func (h *DeviceHandler) UpdateDeviceInBulk(c echo.Context) error {
// 	var dto dto.BulkUpdateDeviceDTO
// 	reqPath := utils.GetRequestUrlPath(c)

// 	if err := utils.BindAndValidate(c, &dto); err != nil {
// 		return err
// 	}

// 	updatedDevices, err := h.DeviceService.BulkUpdateDevices(c.Request().Context(), dto.ToDevices())
// 	if err != nil {
// 		return apperror.ErrorHandler(err, apperror.ErrCodeDBUpdate, "").WithPath(reqPath).Wrap(err)
// 	}

// 	return response.JSON(c, int(http.StatusOK), echo.Map{
// 		"message": "devices updated successfully",
// 		"devices": updatedDevices,
// 	})
// }

// func (h *DeviceHandler) UpdateDeviceStatus(c echo.Context) error {
// 	paramID := c.Param("id")
// 	reqPath := utils.GetRequestUrlPath(c)

// 	deviceID, err := uuid.Parse(paramID)
// 	if err != nil {
// 		return apperror.ErrInvalidUUID.WithMessage("invalid uuid format").WithPath(reqPath).Wrap(err)
// 	}

// 	var dto dto.UpdateDeviceStatusDTO
// 	if err := utils.BindAndValidate(c, &dto); err != nil {
// 		return err
// 	}

// 	err = h.DeviceService.UpdateDeviceStatus(c.Request().Context(), deviceID, domain.Status(dto.Status))
// 	if err != nil {
// 		return apperror.ErrorHandler(err, apperror.ErrCodeDBUpdate)
// 	}

// 	return response.JSON(c, int(http.StatusOK), echo.Map{
// 		"message": fmt.Sprintf("device with ID %s status updated to %s", deviceID.String(), dto.Status),
// 	})
// }
// func (h *DeviceHandler) MarkDeviceOnline(c echo.Context) error {
// 	paramID := c.Param("id")
// 	reqPath := utils.GetRequestUrlPath(c)

// 	deviceID, err := uuid.Parse(paramID)
// 	if err != nil {
// 		return apperror.ErrInvalidUUID.WithMessage("invalid uuid format").WithPath(reqPath).Wrap(err)
// 	}

// 	updatedDevice, err := h.DeviceService.MarkDeviceAsOnline(c.Request().Context(), deviceID)
// 	if err != nil {
// 		return apperror.ErrorHandler(err, apperror.ErrCodeDBUpdate)
// 	}

// 	return response.JSON(c, int(http.StatusOK), echo.Map{
// 		"message": fmt.Sprintf("device with ID %s marked as online", deviceID.String()),
// 		"device":  updatedDevice,
// 	})
// }
// func (h *DeviceHandler) MarkDeviceOffline(c echo.Context) error {
// 	paramID := c.Param("id")
// 	reqPath := utils.GetRequestUrlPath(c)

// 	deviceID, err := uuid.Parse(paramID)
// 	if err != nil {
// 		return apperror.ErrInvalidUUID.WithMessage("invalid uuid format").WithPath(reqPath).Wrap(err)
// 	}

// 	updatedDevice, err := h.DeviceService.MarkDeviceAsOffline(c.Request().Context(), deviceID)
// 	if err != nil {
// 		return apperror.ErrorHandler(err, apperror.ErrCodeDBUpdate)
// 	}

// 	return response.JSON(c, int(http.StatusOK), echo.Map{
// 		"message": fmt.Sprintf("device with ID %s marked as offline", deviceID.String()),
// 		"device":  updatedDevice,
// 	})
// }

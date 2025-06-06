package handlers

import (
	"v01_system_backend/apps/models"
	"v01_system_backend/apps/services"
	"v01_system_backend/apps/utils"

	"github.com/labstack/echo/v4"
)

type SystemHandler struct {
	systemService *services.SystemService
}

func NewSystemHandler(systemService *services.SystemService) *SystemHandler {
	return &SystemHandler{systemService: systemService}
}

func (h *SystemHandler) GetSettings(c echo.Context) error {
	settings, err := h.systemService.GetAllSettings()
	if err != nil {
		return utils.InternalServerErrorResponse(c, "Failed to get system settings", err.Error())
	}

	return utils.SuccessResponse(c, "System settings retrieved successfully", settings)
}

func (h *SystemHandler) UpdateSetting(c echo.Context) error {
	var req struct {
		SettingKey   string `json:"setting_key" validate:"required"`
		SettingValue string `json:"setting_value" validate:"required"`
	}
	if err := c.Bind(&req); err != nil {
		return utils.BadRequestResponse(c, "Invalid request data", err.Error())
	}

	// Get updater ID from context
	updatedBy := c.Get("user_id").(int)

	setting, err := h.systemService.UpdateSetting(req.SettingKey, req.SettingValue, updatedBy)
	if err != nil {
		return utils.BadRequestResponse(c, err.Error(), nil)
	}

	return utils.SuccessResponse(c, "Setting updated successfully", setting)
}

func (h *SystemHandler) GetUserStatuses(c echo.Context) error {
	statuses, err := h.systemService.GetUserStatuses()
	if err != nil {
		return utils.InternalServerErrorResponse(c, "Failed to get user statuses", err.Error())
	}

	return utils.SuccessResponse(c, "User statuses retrieved successfully", statuses)
}

func (h *SystemHandler) GetDepartments(c echo.Context) error {
	departments, err := h.systemService.GetDepartments()
	if err != nil {
		return utils.InternalServerErrorResponse(c, "Failed to get departments", err.Error())
	}

	return utils.SuccessResponse(c, "Departments retrieved successfully", departments)
}

func (h *SystemHandler) CreateDepartment(c echo.Context) error {
	var req models.DepartmentCreateRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequestResponse(c, "Invalid request data", err.Error())
	}

	// Get creator ID from context
	createdBy := c.Get("user_id").(int)

	department, err := h.systemService.CreateDepartment(&req, createdBy)
	if err != nil {
		return utils.BadRequestResponse(c, err.Error(), nil)
	}

	return utils.CreatedResponse(c, "Department created successfully", department)
}

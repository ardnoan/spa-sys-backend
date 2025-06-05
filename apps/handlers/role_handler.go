package handlers

import (
	"strconv"
	"v01_system_backend/apps/models"
	"v01_system_backend/apps/services"
	"v01_system_backend/apps/utils"

	"github.com/labstack/echo/v4"
)

type RoleHandler struct {
	roleService *services.RoleService
}

func NewRoleHandler(roleService *services.RoleService) *RoleHandler {
	return &RoleHandler{roleService: roleService}
}

func (h *RoleHandler) GetAll(c echo.Context) error {
	var pagination models.PaginationRequest
	if err := c.Bind(&pagination); err != nil {
		return utils.BadRequestResponse(c, "Invalid pagination parameters", err.Error())
	}

	response, err := h.roleService.GetAll(&pagination)
	if err != nil {
		return utils.InternalServerErrorResponse(c, "Failed to get roles", err.Error())
	}

	return utils.SuccessResponse(c, "Roles retrieved successfully", response)
}

func (h *RoleHandler) GetByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequestResponse(c, "Invalid role ID", err.Error())
	}

	role, err := h.roleService.GetByID(id)
	if err != nil {
		return utils.NotFoundResponse(c, err.Error())
	}

	return utils.SuccessResponse(c, "Role retrieved successfully", role)
}

func (h *RoleHandler) Create(c echo.Context) error {
	var req models.RoleCreateRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequestResponse(c, "Invalid request data", err.Error())
	}

	// Get creator ID from context
	createdBy := c.Get("user_id").(int)

	role, err := h.roleService.Create(&req, createdBy)
	if err != nil {
		return utils.BadRequestResponse(c, err.Error(), nil)
	}

	return utils.CreatedResponse(c, "Role created successfully", role)
}

func (h *RoleHandler) Update(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequestResponse(c, "Invalid role ID", err.Error())
	}

	var req models.RoleUpdateRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequestResponse(c, "Invalid request data", err.Error())
	}

	// Get updater ID from context
	updatedBy := c.Get("user_id").(int)

	role, err := h.roleService.Update(id, &req, updatedBy)
	if err != nil {
		return utils.BadRequestResponse(c, err.Error(), nil)
	}

	return utils.SuccessResponse(c, "Role updated successfully", role)
}

func (h *RoleHandler) Delete(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequestResponse(c, "Invalid role ID", err.Error())
	}

	// Get deleter ID from context
	deletedBy := c.Get("user_id").(int)

	if err := h.roleService.Delete(id, deletedBy); err != nil {
		return utils.BadRequestResponse(c, err.Error(), nil)
	}

	return utils.SuccessResponse(c, "Role deleted successfully", nil)
}

func (h *RoleHandler) GetPermissions(c echo.Context) error {
	permissions, err := h.roleService.GetAllPermissions()
	if err != nil {
		return utils.InternalServerErrorResponse(c, "Failed to get permissions", err.Error())
	}

	return utils.SuccessResponse(c, "Permissions retrieved successfully", permissions)
}

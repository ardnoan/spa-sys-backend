// controllers/user_controller.go
package controllers

import (
	"strconv"
	"v01_system_backend/apps/models"
	"v01_system_backend/apps/services"
	"v01_system_backend/apps/utils"

	"github.com/labstack/echo/v4"
)

type UserController struct {
	userService *services.UserService
}

func NewUserController(userService *services.UserService) *UserController {
	return &UserController{userService: userService}
}

func (h *UserController) GetAll(c echo.Context) error {
	var pagination models.PaginationRequest
	if err := c.Bind(&pagination); err != nil {
		return utils.BadRequestResponse(c, "Invalid pagination parameters", err.Error())
	}

	// Parse filters
	var filters models.UserFilters
	filters.Status = c.QueryParam("status")
	filters.Search = c.QueryParam("search")
	filters.Role = c.QueryParam("role")

	response, err := h.userService.GetAll(&pagination, &filters)
	if err != nil {
		return utils.InternalServerErrorResponse(c, "Failed to get users", err.Error())
	}

	return utils.SuccessResponse(c, "Users retrieved successfully", response)
}

func (h *UserController) GetByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequestResponse(c, "Invalid user ID", err.Error())
	}

	user, err := h.userService.GetByID(id)
	if err != nil {
		if err.Error() == "user not found" {
			return utils.NotFoundResponse(c, err.Error())
		}
		return utils.InternalServerErrorResponse(c, "Failed to get user", err.Error())
	}

	return utils.SuccessResponse(c, "User retrieved successfully", user)
}

func (h *UserController) Create(c echo.Context) error {
	var req models.UserCreateRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequestResponse(c, "Invalid request data", err.Error())
	}

	// Validate request
	if err := utils.ValidateStruct(&req); err != nil {
		return utils.BadRequestResponse(c, "Validation failed", err.Error())
	}

	// Get creator ID from context
	createdBy := c.Get("user_id").(int)

	user, err := h.userService.Create(&req, createdBy)
	if err != nil {
		return utils.BadRequestResponse(c, err.Error(), nil)
	}

	return utils.CreatedResponse(c, "User created successfully", user)
}

func (h *UserController) Update(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequestResponse(c, "Invalid user ID", err.Error())
	}

	var req models.UserUpdateRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequestResponse(c, "Invalid request data", err.Error())
	}

	// Validate request
	if err := utils.ValidateStruct(&req); err != nil {
		return utils.BadRequestResponse(c, "Validation failed", err.Error())
	}

	// Get updater ID from context
	updatedBy := c.Get("user_id").(int)

	user, err := h.userService.Update(id, &req, updatedBy)
	if err != nil {
		if err.Error() == "user not found" {
			return utils.NotFoundResponse(c, err.Error())
		}
		return utils.BadRequestResponse(c, err.Error(), nil)
	}

	return utils.SuccessResponse(c, "User updated successfully", user)
}

func (h *UserController) Delete(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequestResponse(c, "Invalid user ID", err.Error())
	}

	// Get deleter ID from context
	deletedBy := c.Get("user_id").(int)

	if err := h.userService.Delete(id, deletedBy); err != nil {
		if err.Error() == "user not found" {
			return utils.NotFoundResponse(c, err.Error())
		}
		return utils.BadRequestResponse(c, err.Error(), nil)
	}

	return utils.SuccessResponse(c, "User deleted successfully", nil)
}

func (h *UserController) UpdateStatus(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequestResponse(c, "Invalid user ID", err.Error())
	}

	var req models.StatusUpdateRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequestResponse(c, "Invalid request data", err.Error())
	}

	// Get updater ID from context
	updatedBy := c.Get("user_id").(int)

	if err := h.userService.UpdateStatus(id, req.Status, updatedBy); err != nil {
		if err.Error() == "user not found" {
			return utils.NotFoundResponse(c, err.Error())
		}
		return utils.BadRequestResponse(c, err.Error(), nil)
	}

	return utils.SuccessResponse(c, "User status updated successfully", nil)
}

func (h *UserController) ResetPassword(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequestResponse(c, "Invalid user ID", err.Error())
	}

	var req models.PasswordResetRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequestResponse(c, "Invalid request data", err.Error())
	}

	// Validate request
	if err := utils.ValidateStruct(&req); err != nil {
		return utils.BadRequestResponse(c, "Validation failed", err.Error())
	}

	// Get updater ID from context
	updatedBy := c.Get("user_id").(int)

	if err := h.userService.ResetPassword(id, req.NewPassword, updatedBy); err != nil {
		if err.Error() == "user not found" {
			return utils.NotFoundResponse(c, err.Error())
		}
		return utils.BadRequestResponse(c, err.Error(), nil)
	}

	return utils.SuccessResponse(c, "Password reset successfully", nil)
}

// User Roles Management
func (h *UserController) GetUserRoles(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequestResponse(c, "Invalid user ID", err.Error())
	}

	roles, err := h.userService.GetUserRoles(id)
	if err != nil {
		return utils.InternalServerErrorResponse(c, "Failed to get user roles", err.Error())
	}

	return utils.SuccessResponse(c, "User roles retrieved successfully", roles)
}

func (h *UserController) AssignRoles(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequestResponse(c, "Invalid user ID", err.Error())
	}

	var req models.RoleAssignmentRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequestResponse(c, "Invalid request data", err.Error())
	}

	// Get assigner ID from context
	assignedBy := c.Get("user_id").(int)

	if err := h.userService.AssignRoles(id, req.RoleIDs, assignedBy); err != nil {
		if err.Error() == "user not found" {
			return utils.NotFoundResponse(c, err.Error())
		}
		return utils.BadRequestResponse(c, err.Error(), nil)
	}

	return utils.SuccessResponse(c, "Roles assigned successfully", nil)
}

func (h *UserController) RemoveRoles(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequestResponse(c, "Invalid user ID", err.Error())
	}

	var req models.RoleRemovalRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequestResponse(c, "Invalid request data", err.Error())
	}

	// Get remover ID from context
	removedBy := c.Get("user_id").(int)

	if err := h.userService.RemoveRoles(id, req.RoleIDs, removedBy); err != nil {
		return utils.BadRequestResponse(c, err.Error(), nil)
	}

	return utils.SuccessResponse(c, "Roles removed successfully", nil)
}

// User Permissions
func (h *UserController) GetUserPermissions(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequestResponse(c, "Invalid user ID", err.Error())
	}

	permissions, err := h.userService.GetUserPermissions(id)
	if err != nil {
		return utils.InternalServerErrorResponse(c, "Failed to get user permissions", err.Error())
	}

	return utils.SuccessResponse(c, "User permissions retrieved successfully", permissions)
}

// User Activities
func (h *UserController) GetUserActivities(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequestResponse(c, "Invalid user ID", err.Error())
	}

	var pagination models.PaginationRequest
	if err := c.Bind(&pagination); err != nil {
		return utils.BadRequestResponse(c, "Invalid pagination parameters", err.Error())
	}

	response, err := h.userService.GetUserActivities(id, &pagination)
	if err != nil {
		return utils.InternalServerErrorResponse(c, "Failed to get user activities", err.Error())
	}

	return utils.SuccessResponse(c, "User activities retrieved successfully", response)
}

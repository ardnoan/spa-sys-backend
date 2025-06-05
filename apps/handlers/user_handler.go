package handlers

import (
	"strconv"
	"v01_system_backend/apps/models"
	"v01_system_backend/apps/services"
	"v01_system_backend/apps/utils"

	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) GetAll(c echo.Context) error {
	var pagination models.PaginationRequest
	if err := c.Bind(&pagination); err != nil {
		return utils.BadRequestResponse(c, "Invalid pagination parameters", err.Error())
	}

	response, err := h.userService.GetAll(&pagination)
	if err != nil {
		return utils.InternalServerErrorResponse(c, "Failed to get users", err.Error())
	}

	return utils.SuccessResponse(c, "Users retrieved successfully", response)
}

func (h *UserHandler) GetByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequestResponse(c, "Invalid user ID", err.Error())
	}

	user, err := h.userService.GetByID(id)
	if err != nil {
		return utils.NotFoundResponse(c, err.Error())
	}

	return utils.SuccessResponse(c, "User retrieved successfully", user)
}

func (h *UserHandler) Create(c echo.Context) error {
	var req models.UserCreateRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequestResponse(c, "Invalid request data", err.Error())
	}

	// Get creator ID from context
	createdBy := c.Get("user_id").(int)

	user, err := h.userService.Create(&req, createdBy)
	if err != nil {
		return utils.BadRequestResponse(c, err.Error(), nil)
	}

	return utils.CreatedResponse(c, "User created successfully", user)
}

func (h *UserHandler) Update(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequestResponse(c, "Invalid user ID", err.Error())
	}

	var req models.UserUpdateRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequestResponse(c, "Invalid request data", err.Error())
	}

	// Get updater ID from context
	updatedBy := c.Get("user_id").(int)

	user, err := h.userService.Update(id, &req, updatedBy)
	if err != nil {
		return utils.BadRequestResponse(c, err.Error(), nil)
	}

	return utils.SuccessResponse(c, "User updated successfully", user)
}

func (h *UserHandler) Delete(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequestResponse(c, "Invalid user ID", err.Error())
	}

	// Get deleter ID from context
	deletedBy := c.Get("user_id").(int)

	if err := h.userService.Delete(id, deletedBy); err != nil {
		return utils.BadRequestResponse(c, err.Error(), nil)
	}

	return utils.SuccessResponse(c, "User deleted successfully", nil)
}

func (h *UserHandler) ResetPassword(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequestResponse(c, "Invalid user ID", err.Error())
	}

	var req struct {
		NewPassword string `json:"new_password" validate:"required,min=8"`
	}
	if err := c.Bind(&req); err != nil {
		return utils.BadRequestResponse(c, "Invalid request data", err.Error())
	}

	// Get updater ID from context
	updatedBy := c.Get("user_id").(int)

	if err := h.userService.ResetPassword(id, req.NewPassword, updatedBy); err != nil {
		return utils.BadRequestResponse(c, err.Error(), nil)
	}

	return utils.SuccessResponse(c, "Password reset successfully", nil)
}

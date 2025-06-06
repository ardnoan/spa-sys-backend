package handlers

import (
	"strconv"
	"v01_system_backend/apps/models"
	"v01_system_backend/apps/services"
	"v01_system_backend/apps/utils"

	"github.com/labstack/echo/v4"
)

type MenuHandler struct {
	menuService *services.MenuService
}

func NewMenuHandler(menuService *services.MenuService) *MenuHandler {
	return &MenuHandler{menuService: menuService}
}

func (h *MenuHandler) GetAll(c echo.Context) error {
	var pagination models.PaginationRequest
	if err := c.Bind(&pagination); err != nil {
		return utils.BadRequestResponse(c, "Invalid pagination parameters", err.Error())
	}

	response, err := h.menuService.GetAll()
	if err != nil {
		return utils.InternalServerErrorResponse(c, "Failed to get menus", err.Error())
	}

	return utils.SuccessResponse(c, "Menus retrieved successfully", response)
}

func (h *MenuHandler) GetUserMenus(c echo.Context) error {
	userID := c.Get("user_id").(int)

	menus, err := h.menuService.GetUserMenus(userID)
	if err != nil {
		return utils.InternalServerErrorResponse(c, "Failed to get user menus", err.Error())
	}

	return utils.SuccessResponse(c, "User menus retrieved successfully", menus)
}

func (h *MenuHandler) GetByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequestResponse(c, "Invalid menu ID", err.Error())
	}

	menu, err := h.menuService.GetByID(id)
	if err != nil {
		return utils.NotFoundResponse(c, err.Error())
	}

	return utils.SuccessResponse(c, "Menu retrieved successfully", menu)
}

func (h *MenuHandler) Create(c echo.Context) error {
	var req models.MenuCreateRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequestResponse(c, "Invalid request data", err.Error())
	}

	// Get creator ID from context
	createdBy := c.Get("user_id").(int)

	menu, err := h.menuService.Create(&req, createdBy)
	if err != nil {
		return utils.BadRequestResponse(c, err.Error(), nil)
	}

	return utils.CreatedResponse(c, "Menu created successfully", menu)
}

func (h *MenuHandler) Update(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequestResponse(c, "Invalid menu ID", err.Error())
	}

	var req models.MenuUpdateRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequestResponse(c, "Invalid request data", err.Error())
	}

	// Get updater ID from context
	updatedBy := c.Get("user_id").(int)

	menu, err := h.menuService.Update(id, &req, updatedBy)
	if err != nil {
		return utils.BadRequestResponse(c, err.Error(), nil)
	}

	return utils.SuccessResponse(c, "Menu updated successfully", menu)
}

func (h *MenuHandler) Delete(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return utils.BadRequestResponse(c, "Invalid menu ID", err.Error())
	}

	// Get deleter ID from context
	deletedBy := c.Get("user_id").(int)

	if err := h.menuService.Delete(id, deletedBy); err != nil {
		return utils.BadRequestResponse(c, err.Error(), nil)
	}

	return utils.SuccessResponse(c, "Menu deleted successfully", nil)
}

package controllers

import (
	"v01_system_backend/apps/services"

	"github.com/labstack/echo/v4"
)

type MenuController struct {
	menuService *services.MenuService
}

func NewMenuController(menuService *services.MenuService) *MenuController {
	return &MenuController{
		menuService: menuService,
	}
}

func (mc *MenuController) GetAll(c echo.Context) error {
	return c.JSON(200, map[string]string{"message": "GetAll menus endpoint"})
}

func (mc *MenuController) GetMenuTree(c echo.Context) error {
	return c.JSON(200, map[string]string{"message": "GetMenuTree endpoint"})
}

func (mc *MenuController) GetUserMenus(c echo.Context) error {
	return c.JSON(200, map[string]string{"message": "GetUserMenus endpoint"})
}

func (mc *MenuController) GetByID(c echo.Context) error {
	return c.JSON(200, map[string]string{"message": "GetByID menu endpoint"})
}

func (mc *MenuController) Create(c echo.Context) error {
	return c.JSON(200, map[string]string{"message": "Create menu endpoint"})
}

func (mc *MenuController) Update(c echo.Context) error {
	return c.JSON(200, map[string]string{"message": "Update menu endpoint"})
}

func (mc *MenuController) Delete(c echo.Context) error {
	return c.JSON(200, map[string]string{"message": "Delete menu endpoint"})
}

func (mc *MenuController) UpdateOrder(c echo.Context) error {
	return c.JSON(200, map[string]string{"message": "UpdateOrder menu endpoint"})
}

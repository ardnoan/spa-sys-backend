package controllers

import (
	"v01_system_backend/apps/services"

	"github.com/labstack/echo/v4"
)

type RoleController struct {
	roleService *services.RoleService
}

func NewRoleController(roleService *services.RoleService) *RoleController {
	return &RoleController{
		roleService: roleService,
	}
}

func (rc *RoleController) GetAll(c echo.Context) error {
	return c.JSON(200, map[string]string{"message": "GetAll roles endpoint"})
}

func (rc *RoleController) GetByID(c echo.Context) error {
	return c.JSON(200, map[string]string{"message": "GetByID role endpoint"})
}

func (rc *RoleController) Create(c echo.Context) error {
	return c.JSON(200, map[string]string{"message": "Create role endpoint"})
}

func (rc *RoleController) Update(c echo.Context) error {
	return c.JSON(200, map[string]string{"message": "Update role endpoint"})
}

func (rc *RoleController) Delete(c echo.Context) error {
	return c.JSON(200, map[string]string{"message": "Delete role endpoint"})
}

func (rc *RoleController) GetRolePermissions(c echo.Context) error {
	return c.JSON(200, map[string]string{"message": "GetRolePermissions endpoint"})
}

func (rc *RoleController) AssignPermissions(c echo.Context) error {
	return c.JSON(200, map[string]string{"message": "AssignPermissions endpoint"})
}

func (rc *RoleController) RemovePermissions(c echo.Context) error {
	return c.JSON(200, map[string]string{"message": "RemovePermissions endpoint"})
}

func (rc *RoleController) GetRoleMenus(c echo.Context) error {
	return c.JSON(200, map[string]string{"message": "GetRoleMenus endpoint"})
}

func (rc *RoleController) AssignMenus(c echo.Context) error {
	return c.JSON(200, map[string]string{"message": "AssignMenus endpoint"})
}

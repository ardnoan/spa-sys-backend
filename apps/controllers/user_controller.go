package controllers

import (
	"v01_system_backend/apps/services"

	"github.com/labstack/echo/v4"
)

type UserController struct {
	userService *services.UserService
}

func NewUserController(userService *services.UserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

func (uc *UserController) GetAll(c echo.Context) error {
	return c.JSON(200, map[string]string{"message": "GetAll users endpoint"})
}

func (uc *UserController) GetByID(c echo.Context) error {
	return c.JSON(200, map[string]string{"message": "GetByID user endpoint"})
}

func (uc *UserController) Create(c echo.Context) error {
	return c.JSON(200, map[string]string{"message": "Create user endpoint"})
}

func (uc *UserController) Update(c echo.Context) error {
	return c.JSON(200, map[string]string{"message": "Update user endpoint"})
}

func (uc *UserController) Delete(c echo.Context) error {
	return c.JSON(200, map[string]string{"message": "Delete user endpoint"})
}

func (uc *UserController) UpdateStatus(c echo.Context) error {
	return c.JSON(200, map[string]string{"message": "UpdateStatus user endpoint"})
}

func (uc *UserController) GetUserRoles(c echo.Context) error {
	return c.JSON(200, map[string]string{"message": "GetUserRoles endpoint"})
}

func (uc *UserController) AssignRoles(c echo.Context) error {
	return c.JSON(200, map[string]string{"message": "AssignRoles endpoint"})
}

func (uc *UserController) RemoveRoles(c echo.Context) error {
	return c.JSON(200, map[string]string{"message": "RemoveRoles endpoint"})
}

func (uc *UserController) GetUserPermissions(c echo.Context) error {
	return c.JSON(200, map[string]string{"message": "GetUserPermissions endpoint"})
}

func (uc *UserController) GetUserActivities(c echo.Context) error {
	return c.JSON(200, map[string]string{"message": "GetUserActivities endpoint"})
}

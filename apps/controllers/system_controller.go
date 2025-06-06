package controllers

import (
	"v01_system_backend/apps/services"

	"github.com/labstack/echo/v4"
)

type SystemController struct {
	systemService *services.SystemService
}

func NewSystemController(systemService *services.SystemService) *SystemController {
	return &SystemController{
		systemService: systemService,
	}
}

func (sc *SystemController) GetAllSettings(c echo.Context) error {
	return c.JSON(200, map[string]string{"message": "GetAllSettings endpoint"})
}

func (sc *SystemController) GetPublicSettings(c echo.Context) error {
	return c.JSON(200, map[string]string{"message": "GetPublicSettings endpoint"})
}

func (sc *SystemController) GetSettingByKey(c echo.Context) error {
	return c.JSON(200, map[string]string{"message": "GetSettingByKey endpoint"})
}

func (sc *SystemController) CreateSetting(c echo.Context) error {
	return c.JSON(200, map[string]string{"message": "CreateSetting endpoint"})
}

func (sc *SystemController) UpdateSetting(c echo.Context) error {
	return c.JSON(200, map[string]string{"message": "UpdateSetting endpoint"})
}

func (sc *SystemController) DeleteSetting(c echo.Context) error {
	return c.JSON(200, map[string]string{"message": "DeleteSetting endpoint"})
}

func (sc *SystemController) GetAppInfo(c echo.Context) error {
	return c.JSON(200, map[string]string{"message": "GetAppInfo endpoint"})
}

func (sc *SystemController) HealthCheck(c echo.Context) error {
	return c.JSON(200, map[string]string{"message": "HealthCheck endpoint"})
}

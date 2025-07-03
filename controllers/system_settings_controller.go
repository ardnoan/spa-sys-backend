package controller

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

type SystemSettingsController struct {
	DB *sql.DB
}

type SystemSetting struct {
	SystemID     int     `json:"system_id"`
	SettingKey   string  `json:"setting_key"`
	SettingValue *string `json:"setting_value"`
	SettingType  string  `json:"setting_type"`
	Description  *string `json:"description"`
	IsPublic     bool    `json:"is_public"`
	IsActive     bool    `json:"is_active"`
	CreatedAt    string  `json:"created_at"`
	CreatedBy    *string `json:"created_by"`
	UpdatedAt    string  `json:"updated_at"`
	UpdatedBy    *string `json:"updated_by"`
}

func NewSystemSettingsController(db *sql.DB) *SystemSettingsController {
	return &SystemSettingsController{DB: db}
}

func (c *SystemSettingsController) GetAllSystemSettings(ctx echo.Context) error {
	page, _ := strconv.Atoi(ctx.QueryParam("page"))
	limit, _ := strconv.Atoi(ctx.QueryParam("limit"))
	search := ctx.QueryParam("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit

	var countQuery string
	var query string
	var args []interface{}

	if search != "" {
		countQuery = `
			SELECT COUNT(*) 
			FROM system_settings
			WHERE setting_key ILIKE $1 OR setting_value ILIKE $1
		`

		query = `
			SELECT system_id, setting_key, setting_value, setting_type, description,
				   is_public, is_active, created_at, created_by, updated_at, updated_by
			FROM system_settings
			WHERE setting_key ILIKE $1 OR setting_value ILIKE $1
			ORDER BY setting_key
			LIMIT $2 OFFSET $3
		`
		searchParam := "%" + search + "%"
		args = []interface{}{searchParam, limit, offset}
	} else {
		countQuery = `SELECT COUNT(*) FROM system_settings`

		query = `
			SELECT system_id, setting_key, setting_value, setting_type, description,
				   is_public, is_active, created_at, created_by, updated_at, updated_by
			FROM system_settings
			ORDER BY setting_key
			LIMIT $1 OFFSET $2
		`
		args = []interface{}{limit, offset}
	}

	var totalRecords int
	if search != "" {
		err := c.DB.QueryRow(countQuery, "%"+search+"%").Scan(&totalRecords)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to count records"})
		}
	} else {
		err := c.DB.QueryRow(countQuery).Scan(&totalRecords)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to count records"})
		}
	}

	rows, err := c.DB.Query(query, args...)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch system settings"})
	}
	defer rows.Close()

	var settings []SystemSetting
	for rows.Next() {
		var setting SystemSetting
		err := rows.Scan(
			&setting.SystemID,
			&setting.SettingKey,
			&setting.SettingValue,
			&setting.SettingType,
			&setting.Description,
			&setting.IsPublic,
			&setting.IsActive,
			&setting.CreatedAt,
			&setting.CreatedBy,
			&setting.UpdatedAt,
			&setting.UpdatedBy,
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to scan system setting"})
		}
		settings = append(settings, setting)
	}

	totalPages := (totalRecords + limit - 1) / limit

	response := map[string]interface{}{
		"data": settings,
		"pagination": map[string]interface{}{
			"current_page":     page,
			"total_pages":      totalPages,
			"total_records":    totalRecords,
			"records_per_page": limit,
		},
	}

	return ctx.JSON(http.StatusOK, response)
}

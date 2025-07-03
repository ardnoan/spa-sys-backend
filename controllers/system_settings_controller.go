package controller

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
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

func (c *SystemSettingsController) GetAllSystemSettings(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	search := ctx.Query("search")

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
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count records"})
			return
		}
	} else {
		err := c.DB.QueryRow(countQuery).Scan(&totalRecords)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count records"})
			return
		}
	}

	rows, err := c.DB.Query(query, args...)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch system settings"})
		return
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
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan system setting"})
			return
		}
		settings = append(settings, setting)
	}

	totalPages := (totalRecords + limit - 1) / limit

	response := gin.H{
		"data": settings,
		"pagination": gin.H{
			"current_page":     page,
			"total_pages":      totalPages,
			"total_records":    totalRecords,
			"records_per_page": limit,
		},
	}

	ctx.JSON(http.StatusOK, response)
}

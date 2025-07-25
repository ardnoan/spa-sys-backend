package controller

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

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

type CreateSystemSettingRequest struct {
	SettingKey   string  `json:"setting_key" validate:"required"`
	SettingValue *string `json:"setting_value"`
	SettingType  string  `json:"setting_type" validate:"required"`
	Description  *string `json:"description"`
	IsPublic     bool    `json:"is_public"`
	IsActive     bool    `json:"is_active"`
}

type UpdateSystemSettingRequest struct {
	SettingKey   string  `json:"setting_key" validate:"required"`
	SettingValue *string `json:"setting_value"`
	SettingType  string  `json:"setting_type" validate:"required"`
	Description  *string `json:"description"`
	IsPublic     bool    `json:"is_public"`
	IsActive     bool    `json:"is_active"`
}

func NewSystemSettingsController(db *sql.DB) *SystemSettingsController {
	return &SystemSettingsController{DB: db}
}

// GetAllSystemSettings - Get all system settings with pagination and search
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
			WHERE setting_key ILIKE $1 OR setting_value ILIKE $1 OR description ILIKE $1
		`

		query = `
			SELECT system_id, setting_key, setting_value, setting_type, description,
				   is_public, is_active, created_at, created_by, updated_at, updated_by
			FROM system_settings
			WHERE setting_key ILIKE $1 OR setting_value ILIKE $1 OR description ILIKE $1
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
			"current_page": page,
			"total_pages":  totalPages,
			"total_count":  totalRecords,
			"per_page":     limit,
		},
		"success": true,
	}

	return ctx.JSON(http.StatusOK, response)
}

// GetSystemSettingByID - Get system setting by ID
func (c *SystemSettingsController) GetSystemSettingByID(ctx echo.Context) error {
	id := ctx.Param("id")

	query := `
		SELECT system_id, setting_key, setting_value, setting_type, description,
			   is_public, is_active, created_at, created_by, updated_at, updated_by
		FROM system_settings
		WHERE system_id = $1
	`

	var setting SystemSetting
	err := c.DB.QueryRow(query, id).Scan(
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
		if err == sql.ErrNoRows {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "System setting not found"})
		}
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch system setting"})
	}

	response := map[string]interface{}{
		"data":    setting,
		"success": true,
	}

	return ctx.JSON(http.StatusOK, response)
}

// CreateSystemSetting - Create new system setting
func (c *SystemSettingsController) CreateSystemSetting(ctx echo.Context) error {
	var req CreateSystemSettingRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	// Check if setting key already exists
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM system_settings WHERE setting_key = $1)`
	err := c.DB.QueryRow(checkQuery, req.SettingKey).Scan(&exists)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to check setting key"})
	}

	if exists {
		return ctx.JSON(http.StatusConflict, map[string]string{"error": "Setting key already exists"})
	}

	// Insert new system setting
	query := `
		INSERT INTO system_settings (setting_key, setting_value, setting_type, description, is_public, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING system_id, created_at, updated_at
	`

	now := time.Now()
	var setting SystemSetting
	err = c.DB.QueryRow(
		query,
		req.SettingKey,
		req.SettingValue,
		req.SettingType,
		req.Description,
		req.IsPublic,
		req.IsActive,
		now,
		now,
	).Scan(&setting.SystemID, &setting.CreatedAt, &setting.UpdatedAt)

	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create system setting"})
	}

	// Return created setting
	setting.SettingKey = req.SettingKey
	setting.SettingValue = req.SettingValue
	setting.SettingType = req.SettingType
	setting.Description = req.Description
	setting.IsPublic = req.IsPublic
	setting.IsActive = req.IsActive

	response := map[string]interface{}{
		"data":    setting,
		"message": "System setting created successfully",
		"success": true,
	}

	return ctx.JSON(http.StatusCreated, response)
}

// UpdateSystemSetting - Update existing system setting
func (c *SystemSettingsController) UpdateSystemSetting(ctx echo.Context) error {
	id := ctx.Param("id")

	var req UpdateSystemSettingRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	// Check if setting exists
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM system_settings WHERE system_id = $1)`
	err := c.DB.QueryRow(checkQuery, id).Scan(&exists)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to check system setting"})
	}

	if !exists {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": "System setting not found"})
	}

	// Check if setting key already exists for different record
	var keyExists bool
	keyCheckQuery := `SELECT EXISTS(SELECT 1 FROM system_settings WHERE setting_key = $1 AND system_id != $2)`
	err = c.DB.QueryRow(keyCheckQuery, req.SettingKey, id).Scan(&keyExists)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to check setting key"})
	}

	if keyExists {
		return ctx.JSON(http.StatusConflict, map[string]string{"error": "Setting key already exists"})
	}

	// Update system setting
	query := `
		UPDATE system_settings 
		SET setting_key = $1, setting_value = $2, setting_type = $3, description = $4, 
			is_public = $5, is_active = $6, updated_at = $7
		WHERE system_id = $8
		RETURNING system_id, setting_key, setting_value, setting_type, description, 
				  is_public, is_active, created_at, created_by, updated_at, updated_by
	`

	now := time.Now()
	var setting SystemSetting
	err = c.DB.QueryRow(
		query,
		req.SettingKey,
		req.SettingValue,
		req.SettingType,
		req.Description,
		req.IsPublic,
		req.IsActive,
		now,
		id,
	).Scan(
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
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update system setting"})
	}

	response := map[string]interface{}{
		"data":    setting,
		"message": "System setting updated successfully",
		"success": true,
	}

	return ctx.JSON(http.StatusOK, response)
}

// DeleteSystemSetting - Delete system setting
func (c *SystemSettingsController) DeleteSystemSetting(ctx echo.Context) error {
	id := ctx.Param("id")

	// Check if setting exists
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM system_settings WHERE system_id = $1)`
	err := c.DB.QueryRow(checkQuery, id).Scan(&exists)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to check system setting"})
	}

	if !exists {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": "System setting not found"})
	}

	// Delete system setting
	query := `DELETE FROM system_settings WHERE system_id = $1`
	_, err = c.DB.Exec(query, id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete system setting"})
	}

	response := map[string]interface{}{
		"message": "System setting deleted successfully",
		"success": true,
	}

	return ctx.JSON(http.StatusOK, response)
}

// GetPublicSettings - Get only public settings (for frontend)
func (c *SystemSettingsController) GetPublicSettings(ctx echo.Context) error {
	query := `
		SELECT system_id, setting_key, setting_value, setting_type, description
		FROM system_settings
		WHERE is_public = true AND is_active = true
		ORDER BY setting_key
	`

	rows, err := c.DB.Query(query)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch public settings"})
	}
	defer rows.Close()

	settings := make(map[string]interface{})
	for rows.Next() {
		var systemID int
		var settingKey, settingType string
		var settingValue, description *string

		err := rows.Scan(&systemID, &settingKey, &settingValue, &settingType, &description)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to scan public setting"})
		}

		// Parse value based on type
		var parsedValue interface{}
		if settingValue != nil {
			switch settingType {
			case "boolean":
				parsedValue = *settingValue == "true" || *settingValue == "1"
			case "number":
				if val, err := strconv.ParseFloat(*settingValue, 64); err == nil {
					parsedValue = val
				} else {
					parsedValue = *settingValue
				}
			default:
				parsedValue = *settingValue
			}
		}

		settings[settingKey] = parsedValue
	}

	response := map[string]interface{}{
		"data":    settings,
		"success": true,
	}

	return ctx.JSON(http.StatusOK, response)
}

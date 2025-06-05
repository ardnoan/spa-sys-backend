package repositories

import (
	"database/sql"
	"fmt"
	"strings"
	"v01_system_backend/apps/models"

	"github.com/jmoiron/sqlx"
)

type SystemRepository struct {
	db *sqlx.DB
}

type ActivityRepository struct {
	db *sqlx.DB
}

func NewSystemRepository(db *sqlx.DB) *SystemRepository {
	return &SystemRepository{db: db}
}

func NewActivityRepository(db *sqlx.DB) *ActivityRepository {
	return &ActivityRepository{db: db}
}

// System Settings Methods
func (r *SystemRepository) GetSettings(pagination *models.PaginationRequest) ([]models.SystemSetting, int, error) {
	var settings []models.SystemSetting
	var totalRows int

	// Build WHERE clause
	whereClause := "WHERE is_active = true"
	args := []interface{}{}
	argIndex := 1

	if pagination.Search != "" {
		whereClause += fmt.Sprintf(" AND (setting_key ILIKE $%d OR description ILIKE $%d)",
			argIndex, argIndex+1)
		searchPattern := "%" + pagination.Search + "%"
		args = append(args, searchPattern, searchPattern)
		argIndex += 2
	}

	// Count total rows
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM system_settings %s", whereClause)
	if err := r.db.Get(&totalRows, countQuery, args...); err != nil {
		return nil, 0, err
	}

	// Build ORDER BY clause
	orderBy := "ORDER BY setting_key ASC"
	if pagination.SortBy != "" {
		validSortFields := map[string]string{
			"setting_key": "setting_key",
			"created_at":  "created_at",
		}
		if field, exists := validSortFields[pagination.SortBy]; exists {
			orderBy = fmt.Sprintf("ORDER BY %s %s", field, strings.ToUpper(pagination.SortDir))
		}
	}

	// Main query
	query := fmt.Sprintf(`
		SELECT system_id, setting_key, setting_value, setting_type, description,
			   is_public, is_active, created_at, created_by, updated_at, updated_by
		FROM system_settings %s %s
		LIMIT $%d OFFSET $%d`, whereClause, orderBy, argIndex, argIndex+1)

	args = append(args, pagination.GetLimit(), pagination.GetOffset())

	if err := r.db.Select(&settings, query, args...); err != nil {
		return nil, 0, err
	}

	return settings, totalRows, nil
}

func (r *SystemRepository) GetSettingByKey(key string) (*models.SystemSetting, error) {
	var setting models.SystemSetting
	query := `
		SELECT system_id, setting_key, setting_value, setting_type, description,
			   is_public, is_active, created_at, created_by, updated_at, updated_by
		FROM system_settings WHERE setting_key = $1 AND is_active = true`

	if err := r.db.Get(&setting, query, key); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &setting, nil
}

func (r *SystemRepository) CreateSetting(setting *models.SystemSetting, createdBy int) (*models.SystemSetting, error) {
	var systemID int
	query := `
		INSERT INTO system_settings (setting_key, setting_value, setting_type, 
								   description, is_public, created_by)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING system_id`

	if err := r.db.Get(&systemID, query,
		setting.SettingKey, setting.SettingValue, setting.SettingType,
		setting.Description, setting.IsPublic, createdBy); err != nil {
		return nil, err
	}

	return r.GetSettingByID(systemID)
}

func (r *SystemRepository) GetSettingByID(id int) (*models.SystemSetting, error) {
	var setting models.SystemSetting
	query := `
		SELECT system_id, setting_key, setting_value, setting_type, description,
			   is_public, is_active, created_at, created_by, updated_at, updated_by
		FROM system_settings WHERE system_id = $1 AND is_active = true`

	if err := r.db.Get(&setting, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &setting, nil
}

func (r *SystemRepository) UpdateSetting(key string, value string, updatedBy int) error {
	query := `
		UPDATE system_settings SET 
			setting_value = $1, updated_by = $2, updated_at = CURRENT_TIMESTAMP
		WHERE setting_key = $3 AND is_active = true`

	result, err := r.db.Exec(query, value, updatedBy, key)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *SystemRepository) DeleteSetting(key string, deletedBy int) error {
	query := `
		UPDATE system_settings SET 
			is_active = false, updated_by = $1, updated_at = CURRENT_TIMESTAMP
		WHERE setting_key = $2 AND is_active = true`

	result, err := r.db.Exec(query, deletedBy, key)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// Activity Log Methods
func (r *ActivityRepository) Create(activity *models.UserActivityLog) error {
	query := `
		INSERT INTO users_activity_logs (user_id, session_id, action, target_type,
									   target_id, menu_name, description, ip_address,
									   user_agent, request_data, response_status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := r.db.Exec(query,
		activity.UserID, activity.SessionID, activity.Action, activity.TargetType,
		activity.TargetID, activity.MenuName, activity.Description, activity.IPAddress,
		activity.UserAgent, activity.RequestData, activity.ResponseStatus)

	return err
}

func (r *ActivityRepository) GetByUser(userID int, pagination *models.PaginationRequest) ([]models.UserActivityLog, int, error) {
	var activities []models.UserActivityLog
	var totalRows int

	// Build WHERE clause
	whereClause := "WHERE user_id = $1"
	args := []interface{}{userID}
	argIndex := 2

	if pagination.Search != "" {
		whereClause += fmt.Sprintf(" AND (action ILIKE $%d OR description ILIKE $%d OR menu_name ILIKE $%d)",
			argIndex, argIndex+1, argIndex+2)
		searchPattern := "%" + pagination.Search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern)
		argIndex += 3
	}

	// Count total rows
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users_activity_logs %s", whereClause)
	if err := r.db.Get(&totalRows, countQuery, args...); err != nil {
		return nil, 0, err
	}

	// Build ORDER BY clause
	orderBy := "ORDER BY created_at DESC"
	if pagination.SortBy != "" {
		validSortFields := map[string]string{
			"action":     "action",
			"created_at": "created_at",
		}
		if field, exists := validSortFields[pagination.SortBy]; exists {
			orderBy = fmt.Sprintf("ORDER BY %s %s", field, strings.ToUpper(pagination.SortDir))
		}
	}

	// Main query
	query := fmt.Sprintf(`
		SELECT logs_id, user_id, session_id, action, target_type, target_id,
			   menu_name, description, ip_address, user_agent, request_data,
			   response_status, created_at
		FROM users_activity_logs %s %s
		LIMIT $%d OFFSET $%d`, whereClause, orderBy, argIndex, argIndex+1)

	args = append(args, pagination.GetLimit(), pagination.GetOffset())

	if err := r.db.Select(&activities, query, args...); err != nil {
		return nil, 0, err
	}

	return activities, totalRows, nil
}

// System models for settings
type SystemSetting struct {
	SystemID     int     `json:"system_id" db:"system_id"`
	SettingKey   string  `json:"setting_key" db:"setting_key"`
	SettingValue *string `json:"setting_value" db:"setting_value"`
	SettingType  string  `json:"setting_type" db:"setting_type"`
	Description  *string `json:"description" db:"description"`
	IsPublic     bool    `json:"is_public" db:"is_public"`
	IsActive     bool    `json:"is_active" db:"is_active"`
	models.BaseModel
}

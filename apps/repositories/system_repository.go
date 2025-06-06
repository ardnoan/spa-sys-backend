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

func NewSystemRepository(db *sqlx.DB) *SystemRepository {
	return &SystemRepository{db: db}
}

// System Settings Methods
func (r *SystemRepository) GetAll(pagination *models.PaginationRequest) ([]models.SystemSetting, int, error) {
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

func (r *SystemRepository) GetByKey(key string) (*models.SystemSetting, error) {
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

func (r *SystemRepository) Create(req *models.SystemSettingCreateRequest, createdBy int) (*models.SystemSetting, error) {
	var systemID int
	query := `
		INSERT INTO system_settings (setting_key, setting_value, setting_type, 
								   description, is_public, created_by)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING system_id`

	if err := r.db.Get(&systemID, query,
		req.SettingKey, req.SettingValue, req.SettingType,
		req.Description, req.IsPublic, createdBy); err != nil {
		return nil, err
	}

	return r.GetByID(systemID)
}

func (r *SystemRepository) Update(key string, req *models.SystemSettingUpdateRequest, updatedBy int) (*models.SystemSetting, error) {
	query := `
		UPDATE system_settings SET 
			setting_value = $1, setting_type = $2, description = $3, 
			is_public = $4, updated_by = $5, updated_at = CURRENT_TIMESTAMP
		WHERE setting_key = $6 AND is_active = true`

	result, err := r.db.Exec(query, req.SettingValue, req.SettingType, req.Description, req.IsPublic, updatedBy, key)
	if err != nil {
		return nil, err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, sql.ErrNoRows
	}

	return r.GetByKey(key)
}

func (r *SystemRepository) Delete(key string, deletedBy int) error {
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

func (r *SystemRepository) GetByID(id int) (*models.SystemSetting, error) {
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

func (r *SystemRepository) GetPublicSettings() ([]*models.SystemSetting, error) {
	var settings []*models.SystemSetting
	query := `
		SELECT system_id, setting_key, setting_value, setting_type, description,
			   is_public, is_active, created_at, created_by, updated_at, updated_by
		FROM system_settings WHERE is_public = true AND is_active = true
		ORDER BY setting_key ASC`

	if err := r.db.Select(&settings, query); err != nil {
		return nil, err
	}

	return settings, nil
}

// User Status Methods
func (r *SystemRepository) GetUserStatuses() ([]models.UserStatus, error) {
	var statuses []models.UserStatus
	query := `
		SELECT users_application_status_id, status_code, status_name, description,
			   is_active, created_at, created_by, updated_at, updated_by
		FROM users_application_status WHERE is_active = true
		ORDER BY status_name ASC`

	if err := r.db.Select(&statuses, query); err != nil {
		return nil, err
	}

	return statuses, nil
}

// Department Methods
func (r *SystemRepository) GetDepartments() ([]models.Department, error) {
	var departments []models.Department
	query := `
		SELECT department_id, department_name, department_code, parent_id,
			   manager_id, description, is_active, created_at, created_by,
			   updated_at, updated_by
		FROM departments WHERE is_active = true
		ORDER BY department_name ASC`

	if err := r.db.Select(&departments, query); err != nil {
		return nil, err
	}

	return departments, nil
}

func (r *SystemRepository) CreateDepartment(req *models.DepartmentCreateRequest, createdBy int) (*models.Department, error) {
	var departmentID int
	query := `
		INSERT INTO departments (department_name, department_code, parent_id,
							   manager_id, description, created_by)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING department_id`

	if err := r.db.Get(&departmentID, query,
		req.DepartmentName, req.DepartmentCode, req.ParentID,
		req.ManagerID, req.Description, createdBy); err != nil {
		return nil, err
	}

	return r.GetDepartmentByID(departmentID)
}

func (r *SystemRepository) GetDepartmentByID(id int) (*models.Department, error) {
	var department models.Department
	query := `
		SELECT department_id, department_name, department_code, parent_id,
			   manager_id, description, is_active, created_at, created_by,
			   updated_at, updated_by
		FROM departments WHERE department_id = $1 AND is_active = true`

	if err := r.db.Get(&department, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &department, nil
}

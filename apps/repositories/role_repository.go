package repositories

import (
	"database/sql"
	"fmt"
	"strings"
	"v01_system_backend/apps/models"

	"github.com/jmoiron/sqlx"
)

type RoleRepository struct {
	db *sqlx.DB
}

func NewRoleRepository(db *sqlx.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) GetAll(pagination *models.PaginationRequest) ([]models.Role, int, error) {
	var roles []models.Role
	var totalRows int

	// Build WHERE clause
	whereClause := "WHERE is_active = true"
	args := []interface{}{}
	argIndex := 1

	if pagination.Search != "" {
		whereClause += fmt.Sprintf(" AND (roles_name ILIKE $%d OR roles_code ILIKE $%d OR description ILIKE $%d)",
			argIndex, argIndex+1, argIndex+2)
		searchPattern := "%" + pagination.Search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern)
		argIndex += 3
	}

	// Count total rows
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users_roles %s", whereClause)
	if err := r.db.Get(&totalRows, countQuery, args...); err != nil {
		return nil, 0, err
	}

	// Build ORDER BY clause
	orderBy := "ORDER BY created_at DESC"
	if pagination.SortBy != "" {
		validSortFields := map[string]string{
			"roles_name": "roles_name",
			"roles_code": "roles_code",
			"created_at": "created_at",
		}
		if field, exists := validSortFields[pagination.SortBy]; exists {
			orderBy = fmt.Sprintf("ORDER BY %s %s", field, strings.ToUpper(pagination.SortDir))
		}
	}

	// Main query
	query := fmt.Sprintf(`
		SELECT roles_id, roles_name, roles_code, description, is_system_role, is_active,
			   created_at, created_by, updated_at, updated_by
		FROM users_roles %s %s
		LIMIT $%d OFFSET $%d`, whereClause, orderBy, argIndex, argIndex+1)

	args = append(args, pagination.GetLimit(), pagination.GetOffset())

	if err := r.db.Select(&roles, query, args...); err != nil {
		return nil, 0, err
	}

	return roles, totalRows, nil
}

func (r *RoleRepository) GetByID(id int) (*models.Role, error) {
	var role models.Role
	query := `
		SELECT roles_id, roles_name, roles_code, description, is_system_role, is_active,
			   created_at, created_by, updated_at, updated_by
		FROM users_roles WHERE roles_id = $1 AND is_active = true`

	if err := r.db.Get(&role, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &role, nil
}

func (r *RoleRepository) GetByCode(code string) (*models.Role, error) {
	var role models.Role
	query := `SELECT roles_id, roles_name, roles_code FROM users_roles WHERE roles_code = $1 AND is_active = true`

	if err := r.db.Get(&role, query, code); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &role, nil
}

func (r *RoleRepository) Create(role *models.RoleCreateRequest, createdBy int) (*models.Role, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var roleID int
	query := `
		INSERT INTO users_roles (roles_name, roles_code, description, created_by)
		VALUES ($1, $2, $3, $4) RETURNING roles_id`

	if err := tx.Get(&roleID, query, role.RolesName, role.RolesCode, role.Description, createdBy); err != nil {
		return nil, err
	}

	// Assign permissions if provided
	if len(role.PermissionIDs) > 0 {
		for _, permissionID := range role.PermissionIDs {
			_, err := tx.Exec(`
				INSERT INTO role_permissions (role_id, permission_id, granted_by)
				VALUES ($1, $2, $3)`, roleID, permissionID, createdBy)
			if err != nil {
				return nil, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetByID(roleID)
}

func (r *RoleRepository) Update(id int, role *models.RoleUpdateRequest, updatedBy int) (*models.Role, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `
		UPDATE users_roles SET 
			roles_name = $1, roles_code = $2, description = $3,
			updated_by = $4, updated_at = CURRENT_TIMESTAMP
		WHERE roles_id = $5 AND is_active = true AND is_system_role = false`

	result, err := tx.Exec(query, role.RolesName, role.RolesCode, role.Description, updatedBy, id)
	if err != nil {
		return nil, err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, sql.ErrNoRows
	}

	// Update role permissions
	if role.PermissionIDs != nil {
		// Remove existing permissions
		_, err := tx.Exec(`UPDATE role_permissions SET is_active = false WHERE role_id = $1`, id)
		if err != nil {
			return nil, err
		}

		// Add new permissions
		for _, permissionID := range role.PermissionIDs {
			_, err := tx.Exec(`
				INSERT INTO role_permissions (role_id, permission_id, granted_by)
				VALUES ($1, $2, $3)
				ON CONFLICT (role_id, permission_id)
				DO UPDATE SET is_active = true, granted_by = $3, granted_at = CURRENT_TIMESTAMP`,
				id, permissionID, updatedBy)
			if err != nil {
				return nil, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetByID(id)
}

func (r *RoleRepository) Delete(id int, deletedBy int) error {
	query := `
		UPDATE users_roles SET 
			is_active = false, updated_by = $1, updated_at = CURRENT_TIMESTAMP
		WHERE roles_id = $2 AND is_active = true AND is_system_role = false`

	result, err := r.db.Exec(query, deletedBy, id)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *RoleRepository) GetRolePermissions(roleID int) ([]models.Permission, error) {
	var permissions []models.Permission
	query := `
		SELECT p.permissions_id, p.permission_code, p.permission_name, p.description, p.module_name
		FROM permissions p
		INNER JOIN role_permissions rp ON p.permissions_id = rp.permission_id
		WHERE rp.role_id = $1 AND rp.is_active = true AND p.is_active = true`

	if err := r.db.Select(&permissions, query, roleID); err != nil {
		return nil, err
	}

	return permissions, nil
}

func (r *RoleRepository) GetAllPermissions() ([]models.Permission, error) {
	var permissions []models.Permission
	query := `
		SELECT permissions_id, permission_code, permission_name, description, module_name
		FROM permissions WHERE is_active = true ORDER BY module_name, permission_name`

	if err := r.db.Select(&permissions, query); err != nil {
		return nil, err
	}

	return permissions, nil
}

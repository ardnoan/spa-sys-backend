package repositories

import (
	"database/sql"
	"fmt"
	"strings"
	"v01_system_backend/apps/models"

	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetAll(pagination *models.PaginationRequest) ([]models.User, int, error) {
	var users []models.User
	var totalRows int

	// Build WHERE clause for search
	whereClause := "WHERE u.is_active = true"
	args := []interface{}{}
	argIndex := 1

	if pagination.Search != "" {
		whereClause += fmt.Sprintf(" AND (u.username ILIKE $%d OR u.email ILIKE $%d OR u.first_name ILIKE $%d OR u.last_name ILIKE $%d)",
			argIndex, argIndex+1, argIndex+2, argIndex+3)
		searchPattern := "%" + pagination.Search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern, searchPattern)
		argIndex += 4
	}

	// Count total rows
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM users_application u 
		LEFT JOIN users_application_status s ON u.status_id = s.users_application_status_id
		LEFT JOIN departments d ON u.department_id = d.department_id
		%s`, whereClause)

	if err := r.db.Get(&totalRows, countQuery, args...); err != nil {
		return nil, 0, err
	}

	// Build ORDER BY clause
	orderBy := "ORDER BY u.created_at DESC"
	if pagination.SortBy != "" {
		validSortFields := map[string]string{
			"username":   "u.username",
			"email":      "u.email",
			"first_name": "u.first_name",
			"last_name":  "u.last_name",
			"created_at": "u.created_at",
		}
		if field, exists := validSortFields[pagination.SortBy]; exists {
			orderBy = fmt.Sprintf("ORDER BY %s %s", field, strings.ToUpper(pagination.SortDir))
		}
	}

	// Main query with pagination
	query := fmt.Sprintf(`
		SELECT 
			u.user_apps_id, u.username, u.email, u.first_name, u.last_name,
			u.status_id, u.department_id, u.employee_id, u.phone, u.avatar_url,
			u.last_login_at, u.failed_login_attempts, u.locked_until, u.is_active,
			u.created_at, u.created_by, u.updated_at, u.updated_by,
			s.status_name, d.department_name
		FROM users_application u
		LEFT JOIN users_application_status s ON u.status_id = s.users_application_status_id
		LEFT JOIN departments d ON u.department_id = d.department_id
		%s %s
		LIMIT $%d OFFSET $%d`, whereClause, orderBy, argIndex, argIndex+1)

	args = append(args, pagination.GetLimit(), pagination.GetOffset())

	if err := r.db.Select(&users, query, args...); err != nil {
		return nil, 0, err
	}

	return users, totalRows, nil
}

func (r *UserRepository) GetByID(id int) (*models.User, error) {
	var user models.User
	query := `
		SELECT 
			u.user_apps_id, u.username, u.email, u.first_name, u.last_name,
			u.status_id, u.department_id, u.employee_id, u.phone, u.avatar_url,
			u.last_login_at, u.password_changed_at, u.failed_login_attempts, 
			u.locked_until, u.is_active, u.created_at, u.created_by, 
			u.updated_at, u.updated_by, s.status_name, d.department_name
		FROM users_application u
		LEFT JOIN users_application_status s ON u.status_id = s.users_application_status_id
		LEFT JOIN departments d ON u.department_id = d.department_id
		WHERE u.user_apps_id = $1 AND u.is_active = true`

	if err := r.db.Get(&user, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	query := `
		SELECT 
			u.user_apps_id, u.username, u.email, u.password_hash, u.first_name, u.last_name,
			u.status_id, u.department_id, u.failed_login_attempts, u.locked_until, u.is_active,
			u.created_at, u.created_by, u.updated_at, u.updated_by,
			s.status_name, d.department_name
		FROM users_application u
		LEFT JOIN users_application_status s ON u.status_id = s.users_application_status_id
		LEFT JOIN departments d ON u.department_id = d.department_id
		WHERE u.username = $1 AND u.is_active = true`

	if err := r.db.Get(&user, query, username); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	query := `SELECT user_apps_id, username, email FROM users_application WHERE email = $1 AND is_active = true`

	if err := r.db.Get(&user, query, email); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) Create(user *models.UserCreateRequest, hashedPassword string, createdBy int) (*models.User, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var userID int
	query := `
		INSERT INTO users_application (
			username, email, password_hash, first_name, last_name, 
			status_id, department_id, employee_id, phone, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING user_apps_id`

	if err := tx.Get(&userID, query,
		user.Username, user.Email, hashedPassword, user.FirstName, user.LastName,
		user.StatusID, user.DepartmentID, user.EmployeeID, user.Phone, createdBy); err != nil {
		return nil, err
	}

	// Assign roles if provided
	if len(user.RoleIDs) > 0 {
		for _, roleID := range user.RoleIDs {
			_, err := tx.Exec(`
				INSERT INTO user_roles (user_id, role_id, assigned_by) 
				VALUES ($1, $2, $3)`, userID, roleID, createdBy)
			if err != nil {
				return nil, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetByID(userID)
}

func (r *UserRepository) Update(id int, user *models.UserUpdateRequest, updatedBy int) (*models.User, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `
		UPDATE users_application SET 
			username = $1, email = $2, first_name = $3, last_name = $4,
			status_id = $5, department_id = $6, employee_id = $7, phone = $8,
			updated_by = $9, updated_at = CURRENT_TIMESTAMP
		WHERE user_apps_id = $10 AND is_active = true`

	result, err := tx.Exec(query,
		user.Username, user.Email, user.FirstName, user.LastName,
		user.StatusID, user.DepartmentID, user.EmployeeID, user.Phone,
		updatedBy, id)
	if err != nil {
		return nil, err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, sql.ErrNoRows
	}

	// Update user roles
	if user.RoleIDs != nil {
		// Deactivate existing roles
		_, err := tx.Exec(`UPDATE user_roles SET is_active = false WHERE user_id = $1`, id)
		if err != nil {
			return nil, err
		}

		// Add new roles
		for _, roleID := range user.RoleIDs {
			_, err := tx.Exec(`
				INSERT INTO user_roles (user_id, role_id, assigned_by) 
				VALUES ($1, $2, $3)
				ON CONFLICT (user_id, role_id) 
				DO UPDATE SET is_active = true, assigned_by = $3, assigned_at = CURRENT_TIMESTAMP`,
				id, roleID, updatedBy)
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

func (r *UserRepository) Delete(id int, deletedBy int) error {
	query := `
		UPDATE users_application SET 
			is_active = false, updated_by = $1, updated_at = CURRENT_TIMESTAMP
		WHERE user_apps_id = $2 AND is_active = true`

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

func (r *UserRepository) UpdatePassword(userID int, hashedPassword string, updatedBy int) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Save old password to history
	_, err = tx.Exec(`
		INSERT INTO user_password_history (user_id, password_hash)
		SELECT user_apps_id, password_hash FROM users_application WHERE user_apps_id = $1`,
		userID)
	if err != nil {
		return err
	}

	// Update new password
	_, err = tx.Exec(`
		UPDATE users_application SET 
			password_hash = $1, password_changed_at = CURRENT_TIMESTAMP,
			updated_by = $2, updated_at = CURRENT_TIMESTAMP
		WHERE user_apps_id = $3`,
		hashedPassword, updatedBy, userID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *UserRepository) UpdateLastLogin(userID int) error {
	_, err := r.db.Exec(`
		UPDATE users_application SET 
			last_login_at = CURRENT_TIMESTAMP, failed_login_attempts = 0
		WHERE user_apps_id = $1`, userID)
	return err
}

func (r *UserRepository) IncrementFailedAttempts(userID int) error {
	_, err := r.db.Exec(`
		UPDATE users_application SET 
			failed_login_attempts = failed_login_attempts + 1
		WHERE user_apps_id = $1`, userID)
	return err
}

func (r *UserRepository) LockUser(userID int, lockDurationMinutes int) error {
	_, err := r.db.Exec(`
		UPDATE users_application SET 
			locked_until = CURRENT_TIMESTAMP + INTERVAL '%d minutes'
		WHERE user_apps_id = $1`, lockDurationMinutes, userID)
	return err
}

func (r *UserRepository) GetUserRoles(userID int) ([]models.Role, error) {
	var roles []models.Role
	query := `
		SELECT r.roles_id, r.roles_name, r.roles_code, r.description, r.is_system_role
		FROM users_roles r
		INNER JOIN user_roles ur ON r.roles_id = ur.role_id
		WHERE ur.user_id = $1 AND ur.is_active = true AND r.is_active = true`

	if err := r.db.Select(&roles, query, userID); err != nil {
		return nil, err
	}

	return roles, nil
}

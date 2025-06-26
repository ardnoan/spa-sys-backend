package controller

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// Role struct
type Role struct {
	ID           int       `json:"roles_id" db:"roles_id"`
	Name         string    `json:"roles_name" db:"roles_name"`
	Code         string    `json:"roles_code" db:"roles_code"`
	Description  *string   `json:"description" db:"description"`
	IsSystemRole bool      `json:"is_system_role" db:"is_system_role"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	CreatedBy    *string   `json:"created_by" db:"created_by"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	UpdatedBy    *string   `json:"updated_by" db:"updated_by"`
}

// Create Role Request
type CreateRoleRequest struct {
	Name         string  `json:"roles_name" validate:"required,min=3,max=50"`
	Code         string  `json:"roles_code" validate:"required,min=2,max=20,uppercase"`
	Description  *string `json:"description" validate:"omitempty,max=255"`
	IsSystemRole bool    `json:"is_system_role"`
	IsActive     bool    `json:"is_active"`
}

// Update Role Request
type UpdateRoleRequest struct {
	Name         string  `json:"roles_name" validate:"required,min=3,max=50"`
	Description  *string `json:"description" validate:"omitempty,max=255"`
	IsSystemRole bool    `json:"is_system_role"`
	IsActive     bool    `json:"is_active"`
}

// Role Controller
type RoleController struct {
	DB *sql.DB
}

var roleValidate *validator.Validate

func init() {
	roleValidate = validator.New()
	// Register custom validator for uppercase
	roleValidate.RegisterValidation("uppercase", validateUppercase)
}

// Custom validator for uppercase
func validateUppercase(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return value == strings.ToUpper(value)
}

func NewRoleController(db *sql.DB) *RoleController {
	return &RoleController{DB: db}
}

// Response helpers
func (rc *RoleController) successResponse(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

func (rc *RoleController) errorResponse(c echo.Context, code int, message string) error {
	return c.JSON(code, map[string]interface{}{
		"success": false,
		"message": message,
	})
}

// Create Role - FIXED
func (rc *RoleController) CreateRole(c echo.Context) error {
	var req CreateRoleRequest
	if err := c.Bind(&req); err != nil {
		return rc.errorResponse(c, http.StatusBadRequest, "Invalid request format")
	}

	// Validate request
	if err := roleValidate.Struct(&req); err != nil {
		validationErrors := make([]string, 0)
		if validatorErr, ok := err.(validator.ValidationErrors); ok {
			for _, fieldError := range validatorErr {
				switch fieldError.Tag() {
				case "required":
					validationErrors = append(validationErrors, fieldError.Field()+" is required")
				case "min":
					validationErrors = append(validationErrors, fieldError.Field()+" must be at least "+fieldError.Param()+" characters")
				case "max":
					validationErrors = append(validationErrors, fieldError.Field()+" must be at most "+fieldError.Param()+" characters")
				case "uppercase":
					validationErrors = append(validationErrors, fieldError.Field()+" must be uppercase")
				default:
					validationErrors = append(validationErrors, fieldError.Field()+" is invalid")
				}
			}
		}
		return rc.errorResponse(c, http.StatusBadRequest, strings.Join(validationErrors, ", "))
	}

	// Check if role code already exists - FIXED table name
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM users_roles WHERE roles_code = $1)`
	err := rc.DB.QueryRow(checkQuery, req.Code).Scan(&exists)
	if err != nil {
		return rc.errorResponse(c, http.StatusInternalServerError, "Database error")
	}
	if exists {
		return rc.errorResponse(c, http.StatusConflict, "Role code already exists")
	}

	// Check if role name already exists - FIXED table name
	checkNameQuery := `SELECT EXISTS(SELECT 1 FROM users_roles WHERE roles_name = $1)`
	err = rc.DB.QueryRow(checkNameQuery, req.Name).Scan(&exists)
	if err != nil {
		return rc.errorResponse(c, http.StatusInternalServerError, "Database error")
	}
	if exists {
		return rc.errorResponse(c, http.StatusConflict, "Role name already exists")
	}

	// Insert to database - FIXED table name
	query := `INSERT INTO users_roles 
              (roles_name, roles_code, description, is_system_role, is_active, created_by, created_at, updated_at) 
              VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) 
              RETURNING roles_id`

	var roleID int
	err = rc.DB.QueryRow(query,
		req.Name, req.Code, req.Description, req.IsSystemRole, req.IsActive, "system").Scan(&roleID)

	if err != nil {
		return rc.errorResponse(c, http.StatusInternalServerError, "Failed to create role: "+err.Error())
	}

	return rc.successResponse(c, map[string]interface{}{
		"id":      roleID,
		"message": "Role created successfully",
	})
}

// Get Role by ID - FIXED
func (rc *RoleController) GetRole(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return rc.errorResponse(c, http.StatusBadRequest, "Invalid role ID")
	}

	var role Role
	query := `SELECT roles_id, roles_name, roles_code, description, is_system_role, 
              is_active, created_at, created_by, updated_at, updated_by
              FROM users_roles WHERE roles_id = $1`

	err = rc.DB.QueryRow(query, id).Scan(
		&role.ID, &role.Name, &role.Code, &role.Description,
		&role.IsSystemRole, &role.IsActive, &role.CreatedAt,
		&role.CreatedBy, &role.UpdatedAt, &role.UpdatedBy)

	if err != nil {
		if err == sql.ErrNoRows {
			return rc.errorResponse(c, http.StatusNotFound, "Role not found")
		}
		return rc.errorResponse(c, http.StatusInternalServerError, "Failed to fetch role: "+err.Error())
	}

	return rc.successResponse(c, role)
}

// Get All Roles - FIXED
func (rc *RoleController) GetAllRoles(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	search := c.QueryParam("search")
	isActive := c.QueryParam("is_active")
	isSystem := c.QueryParam("is_system_role")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	where := []string{}
	args := []interface{}{}
	index := 1

	if search != "" {
		where = append(where, "(LOWER(roles_name) LIKE $"+strconv.Itoa(index)+" OR LOWER(roles_code) LIKE $"+strconv.Itoa(index)+")")
		args = append(args, "%"+strings.ToLower(search)+"%")
		index++
	}
	if isActive != "" {
		where = append(where, "is_active = $"+strconv.Itoa(index))
		val, _ := strconv.ParseBool(isActive)
		args = append(args, val)
		index++
	}
	if isSystem != "" {
		where = append(where, "is_system_role = $"+strconv.Itoa(index))
		val, _ := strconv.ParseBool(isSystem)
		args = append(args, val)
		index++
	}

	condition := ""
	if len(where) > 0 {
		condition = "WHERE " + strings.Join(where, " AND ")
	}

	// Count query - FIXED table name
	countQuery := "SELECT COUNT(*) FROM users_roles " + condition
	var total int
	err := rc.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return rc.errorResponse(c, http.StatusInternalServerError, "Failed to count users_roles: "+err.Error())
	}

	// Select query - FIXED table name
	query := `SELECT roles_id, roles_name, roles_code, description, is_system_role, is_active, created_at, created_by, updated_at, updated_by
			  FROM users_roles ` + condition + ` ORDER BY created_at DESC LIMIT $` + strconv.Itoa(index) + ` OFFSET $` + strconv.Itoa(index+1)
	args = append(args, limit, offset)

	rows, err := rc.DB.Query(query, args...)
	if err != nil {
		return rc.errorResponse(c, http.StatusInternalServerError, "Failed to fetch roles: "+err.Error())
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var r Role
		err := rows.Scan(&r.ID, &r.Name, &r.Code, &r.Description, &r.IsSystemRole, &r.IsActive, &r.CreatedAt, &r.CreatedBy, &r.UpdatedAt, &r.UpdatedBy)
		if err == nil {
			roles = append(roles, r)
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"roles": roles,
			"pagination": map[string]interface{}{
				"current_page": page,
				"per_page":     limit,
				"total_count":  total,
				"total_pages":  (total + limit - 1) / limit,
				"has_next":     page*limit < total,
				"has_prev":     page > 1,
			},
		},
	})
}

// Update Role - FIXED
func (rc *RoleController) UpdateRole(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return rc.errorResponse(c, http.StatusBadRequest, "Invalid role ID")
	}

	var req UpdateRoleRequest
	if err := c.Bind(&req); err != nil {
		return rc.errorResponse(c, http.StatusBadRequest, "Invalid request format")
	}

	// Validate request
	if err := roleValidate.Struct(&req); err != nil {
		validationErrors := make([]string, 0)
		if validatorErr, ok := err.(validator.ValidationErrors); ok {
			for _, fieldError := range validatorErr {
				switch fieldError.Tag() {
				case "required":
					validationErrors = append(validationErrors, fieldError.Field()+" is required")
				case "min":
					validationErrors = append(validationErrors, fieldError.Field()+" must be at least "+fieldError.Param()+" characters")
				case "max":
					validationErrors = append(validationErrors, fieldError.Field()+" must be at most "+fieldError.Param()+" characters")
				default:
					validationErrors = append(validationErrors, fieldError.Field()+" is invalid")
				}
			}
		}
		return rc.errorResponse(c, http.StatusBadRequest, strings.Join(validationErrors, ", "))
	}

	// Check if role exists - FIXED table name
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM users_roles WHERE roles_id = $1)`
	err = rc.DB.QueryRow(checkQuery, id).Scan(&exists)
	if err != nil || !exists {
		return rc.errorResponse(c, http.StatusNotFound, "Role not found")
	}

	// Check if role name is taken by another role - FIXED table name
	checkNameQuery := `SELECT EXISTS(SELECT 1 FROM users_roles WHERE roles_name = $1 AND roles_id != $2)`
	err = rc.DB.QueryRow(checkNameQuery, req.Name, id).Scan(&exists)
	if err != nil {
		return rc.errorResponse(c, http.StatusInternalServerError, "Database error")
	}
	if exists {
		return rc.errorResponse(c, http.StatusConflict, "Role name already exists")
	}

	// Update role - FIXED table name
	query := `UPDATE users_roles 
              SET roles_name = $1, description = $2, is_system_role = $3, 
                  is_active = $4, updated_by = $5, updated_at = CURRENT_TIMESTAMP
              WHERE roles_id = $6`

	_, err = rc.DB.Exec(query, req.Name, req.Description, req.IsSystemRole,
		req.IsActive, "system", id)
	if err != nil {
		return rc.errorResponse(c, http.StatusInternalServerError, "Failed to update role: "+err.Error())
	}

	return rc.successResponse(c, map[string]string{"message": "Role updated successfully"})
}

// Delete Role (Soft delete) - FIXED
func (rc *RoleController) DeleteRole(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return rc.errorResponse(c, http.StatusBadRequest, "Invalid role ID")
	}

	// Check if role exists and is active - FIXED table name
	var exists bool
	var isSystemRole bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM users_roles WHERE roles_id = $1 AND is_active = true), 
                   COALESCE((SELECT is_system_role FROM users_roles WHERE roles_id = $1), false)`
	err = rc.DB.QueryRow(checkQuery, id).Scan(&exists, &isSystemRole)
	if err != nil || !exists {
		return rc.errorResponse(c, http.StatusNotFound, "Role not found")
	}

	// Prevent deletion of system roles
	if isSystemRole {
		return rc.errorResponse(c, http.StatusForbidden, "Cannot delete system role")
	}

	// Check if role is assigned to any users
	var userCount int
	checkUsersQuery := `SELECT COUNT(*) FROM user_roles WHERE role_id = $1`
	err = rc.DB.QueryRow(checkUsersQuery, id).Scan(&userCount)
	if err != nil {
		return rc.errorResponse(c, http.StatusInternalServerError, "Database error")
	}
	if userCount > 0 {
		return rc.errorResponse(c, http.StatusConflict, "Cannot delete role that is assigned to users")
	}

	// FIXED table name
	query := `UPDATE users_roles 
              SET is_active = false, updated_by = $1, updated_at = CURRENT_TIMESTAMP 
              WHERE roles_id = $2`

	_, err = rc.DB.Exec(query, "system", id)
	if err != nil {
		return rc.errorResponse(c, http.StatusInternalServerError, "Failed to delete role: "+err.Error())
	}

	return rc.successResponse(c, map[string]string{"message": "Role deleted successfully"})
}

// Get Role Options (for dropdowns) - FIXED
func (rc *RoleController) GetRoleOptions(c echo.Context) error {
	query := `SELECT roles_id, roles_name, roles_code 
              FROM users_roles 
              WHERE is_active = true 
              ORDER BY roles_name ASC`

	rows, err := rc.DB.Query(query)
	if err != nil {
		return rc.errorResponse(c, http.StatusInternalServerError, "Failed to fetch role options")
	}
	defer rows.Close()

	var options []map[string]interface{}
	for rows.Next() {
		var id int
		var name, code string
		if err := rows.Scan(&id, &name, &code); err == nil {
			options = append(options, map[string]interface{}{
				"value": id,
				"label": name,
				"code":  code,
			})
		}
	}

	return rc.successResponse(c, options)
}

// Check Role Code Availability - FIXED
func (rc *RoleController) CheckCodeAvailability(c echo.Context) error {
	code := c.QueryParam("code")
	excludeID := c.QueryParam("exclude_id")

	if code == "" {
		return rc.errorResponse(c, http.StatusBadRequest, "Role code is required")
	}

	var exists bool
	var query string
	var args []interface{}

	if excludeID != "" {
		query = `SELECT EXISTS(SELECT 1 FROM users_roles WHERE roles_code = $1 AND roles_id != $2)`
		if excludeIDInt, err := strconv.Atoi(excludeID); err == nil {
			args = []interface{}{code, excludeIDInt}
		} else {
			return rc.errorResponse(c, http.StatusBadRequest, "Invalid exclude ID")
		}
	} else {
		query = `SELECT EXISTS(SELECT 1 FROM users_roles WHERE roles_code = $1)`
		args = []interface{}{code}
	}

	err := rc.DB.QueryRow(query, args...).Scan(&exists)
	if err != nil {
		return rc.errorResponse(c, http.StatusInternalServerError, "Database error")
	}

	return rc.successResponse(c, map[string]interface{}{
		"available": !exists,
		"code":      code,
	})
}

// Search Roles (redirect to GetAllRoles)
func (rc *RoleController) SearchRoles(c echo.Context) error {
	query := c.QueryParam("q")
	if query == "" {
		return rc.errorResponse(c, http.StatusBadRequest, "Search query is required")
	}

	// Redirect to GetAllRoles with search parameter
	c.Request().URL.RawQuery = "search=" + query + "&limit=50"
	return rc.GetAllRoles(c)
}

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

// Permission struct
type Permission struct {
	ID             int       `json:"permissions_id" db:"permissions_id"`
	PermissionCode string    `json:"permission_code" db:"permission_code"`
	PermissionName string    `json:"permission_name" db:"permission_name"`
	Description    *string   `json:"description" db:"description"`
	ModuleName     *string   `json:"module_name" db:"module_name"`
	IsActive       bool      `json:"is_active" db:"is_active"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	CreatedBy      *string   `json:"created_by" db:"created_by"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
	UpdatedBy      *string   `json:"updated_by" db:"updated_by"`
}

// Create Permission Request
type CreatePermissionRequest struct {
	PermissionCode string  `json:"permission_code" validate:"required,min=2,max=50,lowercase_underscore"`
	PermissionName string  `json:"permission_name" validate:"required,min=3,max=100"`
	Description    *string `json:"description" validate:"omitempty,max=500"`
	ModuleName     *string `json:"module_name" validate:"omitempty,max=50"`
	IsActive       bool    `json:"is_active"`
}

// Update Permission Request
type UpdatePermissionRequest struct {
	PermissionName string  `json:"permission_name" validate:"required,min=3,max=100"`
	Description    *string `json:"description" validate:"omitempty,max=500"`
	ModuleName     *string `json:"module_name" validate:"omitempty,max=50"`
	IsActive       bool    `json:"is_active"`
}

// Permission Controller
type PermissionController struct {
	DB *sql.DB
}

var permissionValidate *validator.Validate

func init() {
	permissionValidate = validator.New()
	// Register custom validators
	permissionValidate.RegisterValidation("lowercase_underscore", validateLowercaseUnderscore)
}

// Custom validator for lowercase with underscores
func validateLowercaseUnderscore(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	// Check if contains only lowercase letters, numbers, and underscores
	for _, char := range value {
		if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '_') {
			return false
		}
	}
	return true
}

func NewPermissionController(db *sql.DB) *PermissionController {
	return &PermissionController{DB: db}
}

// Response helpers
func (pc *PermissionController) successResponse(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

func (pc *PermissionController) errorResponse(c echo.Context, code int, message string) error {
	return c.JSON(code, map[string]interface{}{
		"success": false,
		"message": message,
	})
}

// Create Permission
func (pc *PermissionController) CreatePermission(c echo.Context) error {
	var req CreatePermissionRequest
	if err := c.Bind(&req); err != nil {
		return pc.errorResponse(c, http.StatusBadRequest, "Invalid request format")
	}

	// Validate request
	if err := permissionValidate.Struct(&req); err != nil {
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
				case "lowercase_underscore":
					validationErrors = append(validationErrors, fieldError.Field()+" must contain only lowercase letters, numbers, and underscores")
				default:
					validationErrors = append(validationErrors, fieldError.Field()+" is invalid")
				}
			}
		}
		return pc.errorResponse(c, http.StatusBadRequest, strings.Join(validationErrors, ", "))
	}

	// Check if permission code already exists
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM permissions WHERE permission_code = $1)`
	err := pc.DB.QueryRow(checkQuery, req.PermissionCode).Scan(&exists)
	if err != nil {
		return pc.errorResponse(c, http.StatusInternalServerError, "Database error")
	}
	if exists {
		return pc.errorResponse(c, http.StatusConflict, "Permission code already exists")
	}

	// Check if permission name already exists
	checkNameQuery := `SELECT EXISTS(SELECT 1 FROM permissions WHERE permission_name = $1)`
	err = pc.DB.QueryRow(checkNameQuery, req.PermissionName).Scan(&exists)
	if err != nil {
		return pc.errorResponse(c, http.StatusInternalServerError, "Database error")
	}
	if exists {
		return pc.errorResponse(c, http.StatusConflict, "Permission name already exists")
	}

	// Insert to database
	query := `INSERT INTO permissions 
              (permission_code, permission_name, description, module_name, is_active, created_by, created_at, updated_at) 
              VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) 
              RETURNING permissions_id`

	var permissionID int
	err = pc.DB.QueryRow(query,
		req.PermissionCode, req.PermissionName, req.Description, req.ModuleName, req.IsActive, "system").Scan(&permissionID)

	if err != nil {
		return pc.errorResponse(c, http.StatusInternalServerError, "Failed to create permission: "+err.Error())
	}

	return pc.successResponse(c, map[string]interface{}{
		"id":      permissionID,
		"message": "Permission created successfully",
	})
}

// Get Permission by ID
func (pc *PermissionController) GetPermission(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return pc.errorResponse(c, http.StatusBadRequest, "Invalid permission ID")
	}

	var permission Permission
	query := `SELECT permissions_id, permission_code, permission_name, description, module_name, 
              is_active, created_at, created_by, updated_at, updated_by
              FROM permissions WHERE permissions_id = $1`

	err = pc.DB.QueryRow(query, id).Scan(
		&permission.ID, &permission.PermissionCode, &permission.PermissionName, &permission.Description,
		&permission.ModuleName, &permission.IsActive, &permission.CreatedAt,
		&permission.CreatedBy, &permission.UpdatedAt, &permission.UpdatedBy)

	if err != nil {
		if err == sql.ErrNoRows {
			return pc.errorResponse(c, http.StatusNotFound, "Permission not found")
		}
		return pc.errorResponse(c, http.StatusInternalServerError, "Failed to fetch permission: "+err.Error())
	}

	return pc.successResponse(c, permission)
}

// Get All Permissions
func (pc *PermissionController) GetAllPermissions(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	search := c.QueryParam("search")
	isActive := c.QueryParam("is_active")
	moduleName := c.QueryParam("module_name")

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
		where = append(where, "(LOWER(permission_name) LIKE $"+strconv.Itoa(index)+" OR LOWER(permission_code) LIKE $"+strconv.Itoa(index)+")")
		args = append(args, "%"+strings.ToLower(search)+"%")
		index++
	}
	if isActive != "" {
		where = append(where, "is_active = $"+strconv.Itoa(index))
		val, _ := strconv.ParseBool(isActive)
		args = append(args, val)
		index++
	}
	if moduleName != "" {
		where = append(where, "module_name = $"+strconv.Itoa(index))
		args = append(args, moduleName)
		index++
	}

	condition := ""
	if len(where) > 0 {
		condition = "WHERE " + strings.Join(where, " AND ")
	}

	// Count query
	countQuery := "SELECT COUNT(*) FROM permissions " + condition
	var total int
	err := pc.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return pc.errorResponse(c, http.StatusInternalServerError, "Failed to count permissions: "+err.Error())
	}

	// Select query
	query := `SELECT permissions_id, permission_code, permission_name, description, module_name, 
              is_active, created_at, created_by, updated_at, updated_by
			  FROM permissions ` + condition + ` ORDER BY created_at DESC LIMIT $` + strconv.Itoa(index) + ` OFFSET $` + strconv.Itoa(index+1)
	args = append(args, limit, offset)

	rows, err := pc.DB.Query(query, args...)
	if err != nil {
		return pc.errorResponse(c, http.StatusInternalServerError, "Failed to fetch permissions: "+err.Error())
	}
	defer rows.Close()

	var permissions []Permission
	for rows.Next() {
		var p Permission
		err := rows.Scan(&p.ID, &p.PermissionCode, &p.PermissionName, &p.Description, &p.ModuleName,
			&p.IsActive, &p.CreatedAt, &p.CreatedBy, &p.UpdatedAt, &p.UpdatedBy)
		if err == nil {
			permissions = append(permissions, p)
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"permissions": permissions,
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

// Update Permission
func (pc *PermissionController) UpdatePermission(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return pc.errorResponse(c, http.StatusBadRequest, "Invalid permission ID")
	}

	var req UpdatePermissionRequest
	if err := c.Bind(&req); err != nil {
		return pc.errorResponse(c, http.StatusBadRequest, "Invalid request format")
	}

	// Validate request
	if err := permissionValidate.Struct(&req); err != nil {
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
		return pc.errorResponse(c, http.StatusBadRequest, strings.Join(validationErrors, ", "))
	}

	// Check if permission exists
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM permissions WHERE permissions_id = $1)`
	err = pc.DB.QueryRow(checkQuery, id).Scan(&exists)
	if err != nil || !exists {
		return pc.errorResponse(c, http.StatusNotFound, "Permission not found")
	}

	// Check if permission name is taken by another permission
	checkNameQuery := `SELECT EXISTS(SELECT 1 FROM permissions WHERE permission_name = $1 AND permissions_id != $2)`
	err = pc.DB.QueryRow(checkNameQuery, req.PermissionName, id).Scan(&exists)
	if err != nil {
		return pc.errorResponse(c, http.StatusInternalServerError, "Database error")
	}
	if exists {
		return pc.errorResponse(c, http.StatusConflict, "Permission name already exists")
	}

	// Update permission
	query := `UPDATE permissions 
              SET permission_name = $1, description = $2, module_name = $3, 
                  is_active = $4, updated_by = $5, updated_at = CURRENT_TIMESTAMP
              WHERE permissions_id = $6`

	_, err = pc.DB.Exec(query, req.PermissionName, req.Description, req.ModuleName,
		req.IsActive, "system", id)
	if err != nil {
		return pc.errorResponse(c, http.StatusInternalServerError, "Failed to update permission: "+err.Error())
	}

	return pc.successResponse(c, map[string]string{"message": "Permission updated successfully"})
}

// Delete Permission (Soft delete)
func (pc *PermissionController) DeletePermission(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return pc.errorResponse(c, http.StatusBadRequest, "Invalid permission ID")
	}

	// Check if permission exists and is active
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM permissions WHERE permissions_id = $1 AND is_active = true)`
	err = pc.DB.QueryRow(checkQuery, id).Scan(&exists)
	if err != nil || !exists {
		return pc.errorResponse(c, http.StatusNotFound, "Permission not found")
	}

	// Check if permission is assigned to any roles
	var roleCount int
	checkRolesQuery := `SELECT COUNT(*) FROM role_permissions WHERE permission_id = $1`
	err = pc.DB.QueryRow(checkRolesQuery, id).Scan(&roleCount)
	if err != nil {
		return pc.errorResponse(c, http.StatusInternalServerError, "Database error")
	}
	if roleCount > 0 {
		return pc.errorResponse(c, http.StatusConflict, "Cannot delete permission that is assigned to roles")
	}

	// Soft delete permission
	query := `UPDATE permissions 
              SET is_active = false, updated_by = $1, updated_at = CURRENT_TIMESTAMP 
              WHERE permissions_id = $2`

	_, err = pc.DB.Exec(query, "system", id)
	if err != nil {
		return pc.errorResponse(c, http.StatusInternalServerError, "Failed to delete permission: "+err.Error())
	}

	return pc.successResponse(c, map[string]string{"message": "Permission deleted successfully"})
}

// Get Permission Options (for dropdowns)
func (pc *PermissionController) GetPermissionOptions(c echo.Context) error {
	query := `SELECT permissions_id, permission_name, permission_code, module_name
              FROM permissions 
              WHERE is_active = true 
              ORDER BY module_name ASC, permission_name ASC`

	rows, err := pc.DB.Query(query)
	if err != nil {
		return pc.errorResponse(c, http.StatusInternalServerError, "Failed to fetch permission options")
	}
	defer rows.Close()

	var options []map[string]interface{}
	for rows.Next() {
		var id int
		var name, code string
		var moduleName *string
		if err := rows.Scan(&id, &name, &code, &moduleName); err == nil {
			module := ""
			if moduleName != nil {
				module = *moduleName
			}
			options = append(options, map[string]interface{}{
				"value":  id,
				"label":  name,
				"code":   code,
				"module": module,
			})
		}
	}

	return pc.successResponse(c, options)
}

// Get Modules
func (pc *PermissionController) GetModules(c echo.Context) error {
	query := `SELECT DISTINCT module_name 
              FROM permissions 
              WHERE module_name IS NOT NULL AND module_name != '' 
              ORDER BY module_name ASC`

	rows, err := pc.DB.Query(query)
	if err != nil {
		return pc.errorResponse(c, http.StatusInternalServerError, "Failed to fetch modules")
	}
	defer rows.Close()

	var modules []string
	for rows.Next() {
		var module string
		if err := rows.Scan(&module); err == nil {
			modules = append(modules, module)
		}
	}

	return pc.successResponse(c, modules)
}

// Check Permission Code Availability
func (pc *PermissionController) CheckCodeAvailability(c echo.Context) error {
	code := c.QueryParam("code")
	excludeID := c.QueryParam("exclude_id")

	if code == "" {
		return pc.errorResponse(c, http.StatusBadRequest, "Permission code is required")
	}

	var exists bool
	var query string
	var args []interface{}

	if excludeID != "" {
		query = `SELECT EXISTS(SELECT 1 FROM permissions WHERE permission_code = $1 AND permissions_id != $2)`
		if excludeIDInt, err := strconv.Atoi(excludeID); err == nil {
			args = []interface{}{code, excludeIDInt}
		} else {
			return pc.errorResponse(c, http.StatusBadRequest, "Invalid exclude ID")
		}
	} else {
		query = `SELECT EXISTS(SELECT 1 FROM permissions WHERE permission_code = $1)`
		args = []interface{}{code}
	}

	err := pc.DB.QueryRow(query, args...).Scan(&exists)
	if err != nil {
		return pc.errorResponse(c, http.StatusInternalServerError, "Database error")
	}

	return pc.successResponse(c, map[string]interface{}{
		"available": !exists,
		"code":      code,
	})
}

// Search Permissions
func (pc *PermissionController) SearchPermissions(c echo.Context) error {
	query := c.QueryParam("q")
	if query == "" {
		return pc.errorResponse(c, http.StatusBadRequest, "Search query is required")
	}

	// Redirect to GetAllPermissions with search parameter
	c.Request().URL.RawQuery = "search=" + query + "&limit=50"
	return pc.GetAllPermissions(c)
}

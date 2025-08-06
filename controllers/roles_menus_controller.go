package controller

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

type RoleMenusController struct {
	DB *sql.DB
}

type RoleMenu struct {
	RoleMenuID  int       `json:"role_menu_id"`
	RoleID      int       `json:"role_id"`
	RoleName    string    `json:"role_name"`
	MenuID      int       `json:"menu_id"`
	MenuName    string    `json:"menu_name"`
	CanView     bool      `json:"can_view"`
	CanCreate   bool      `json:"can_create"`
	CanModify   bool      `json:"can_modify"`
	CanDelete   bool      `json:"can_delete"`
	CanUpload   bool      `json:"can_upload"`
	CanDownload bool      `json:"can_download"`
	CreatedAt   time.Time `json:"created_at"`
	CreatedBy   *string   `json:"created_by"`
}

type RoleMenuRequest struct {
	RoleID      int  `json:"role_id" validate:"required"`
	MenuID      int  `json:"menu_id" validate:"required"`
	CanView     bool `json:"can_view"`
	CanCreate   bool `json:"can_create"`
	CanModify   bool `json:"can_modify"`
	CanDelete   bool `json:"can_delete"`
	CanUpload   bool `json:"can_upload"`
	CanDownload bool `json:"can_download"`
}

// Using aliases to avoid conflicts with existing structs
type RoleDropdown struct {
	RolesID   int    `json:"roles_id"`
	RolesName string `json:"roles_name"`
}

type MenuDropdown struct {
	MenusID  int    `json:"menus_id"`
	MenuName string `json:"menu_name"`
}

func NewRoleMenusController(db *sql.DB) *RoleMenusController {
	return &RoleMenusController{DB: db}
}

// GetAllRoleMenus handles GET /roles-menus
func (c *RoleMenusController) GetAllRoleMenus(ctx echo.Context) error {
	page, _ := strconv.Atoi(ctx.QueryParam("page"))
	limit, _ := strconv.Atoi(ctx.QueryParam("limit"))
	search := ctx.QueryParam("search")
	roleID := ctx.QueryParam("role_id")
	menuID := ctx.QueryParam("menu_id")

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit

	var countQuery, dataQuery string
	var args []interface{}
	argIndex := 1

	// Build WHERE conditions
	whereConditions := []string{}
	if search != "" {
		searchParam := "%" + search + "%"
		whereConditions = append(whereConditions, "(ur.roles_name ILIKE $"+strconv.Itoa(argIndex)+" OR m.menu_name ILIKE $"+strconv.Itoa(argIndex)+")")
		args = append(args, searchParam)
		argIndex++
	}

	if roleID != "" {
		whereConditions = append(whereConditions, "rm.role_id = $"+strconv.Itoa(argIndex))
		args = append(args, roleID)
		argIndex++
	}

	if menuID != "" {
		whereConditions = append(whereConditions, "rm.menu_id = $"+strconv.Itoa(argIndex))
		args = append(args, menuID)
		argIndex++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	countQuery = `
		SELECT COUNT(*) 
		FROM role_menus rm
		JOIN users_roles ur ON rm.role_id = ur.roles_id
		JOIN menus m ON rm.menu_id = m.menus_id
		` + whereClause

	dataQuery = `
		SELECT rm.role_menu_id, rm.role_id, ur.roles_name, rm.menu_id, m.menu_name,
		       rm.can_view, rm.can_create, rm.can_modify, rm.can_delete, 
		       rm.can_upload, rm.can_download, rm.created_at, rm.created_by
		FROM role_menus rm
		JOIN users_roles ur ON rm.role_id = ur.roles_id
		JOIN menus m ON rm.menu_id = m.menus_id
		` + whereClause + `
		ORDER BY ur.roles_name, m.menu_name
		LIMIT $` + strconv.Itoa(argIndex) + ` OFFSET $` + strconv.Itoa(argIndex+1)

	// Add pagination parameters
	args = append(args, limit, offset)

	var totalRecords int
	countArgs := args[:len(args)-2] // Remove limit and offset for count query
	err := c.DB.QueryRow(countQuery, countArgs...).Scan(&totalRecords)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to count records",
			"message": err.Error(),
		})
	}

	rows, err := c.DB.Query(dataQuery, args...)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to fetch role menus",
			"message": err.Error(),
		})
	}
	defer rows.Close()

	var roleMenus []RoleMenu
	for rows.Next() {
		var roleMenu RoleMenu
		err := rows.Scan(
			&roleMenu.RoleMenuID,
			&roleMenu.RoleID,
			&roleMenu.RoleName,
			&roleMenu.MenuID,
			&roleMenu.MenuName,
			&roleMenu.CanView,
			&roleMenu.CanCreate,
			&roleMenu.CanModify,
			&roleMenu.CanDelete,
			&roleMenu.CanUpload,
			&roleMenu.CanDownload,
			&roleMenu.CreatedAt,
			&roleMenu.CreatedBy,
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
				"error":   "Failed to scan role menu",
				"message": err.Error(),
			})
		}
		roleMenus = append(roleMenus, roleMenu)
	}

	totalPages := (totalRecords + limit - 1) / limit

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"data": roleMenus,
		"pagination": map[string]interface{}{
			"current_page":  page,
			"total_pages":   totalPages,
			"total_count":   totalRecords,
			"per_page":      limit,
		},
	})
}

// GetRoleMenuByID handles GET /roles-menus/:id
func (c *RoleMenusController) GetRoleMenuByID(ctx echo.Context) error {
	id := ctx.Param("id")
	
	query := `
		SELECT rm.role_menu_id, rm.role_id, ur.roles_name, rm.menu_id, m.menu_name,
		       rm.can_view, rm.can_create, rm.can_modify, rm.can_delete, 
		       rm.can_upload, rm.can_download, rm.created_at, rm.created_by
		FROM role_menus rm
		JOIN users_roles ur ON rm.role_id = ur.roles_id
		JOIN menus m ON rm.menu_id = m.menus_id
		WHERE rm.role_menu_id = $1
	`

	var roleMenu RoleMenu
	err := c.DB.QueryRow(query, id).Scan(
		&roleMenu.RoleMenuID,
		&roleMenu.RoleID,
		&roleMenu.RoleName,
		&roleMenu.MenuID,
		&roleMenu.MenuName,
		&roleMenu.CanView,
		&roleMenu.CanCreate,
		&roleMenu.CanModify,
		&roleMenu.CanDelete,
		&roleMenu.CanUpload,
		&roleMenu.CanDownload,
		&roleMenu.CreatedAt,
		&roleMenu.CreatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.JSON(http.StatusNotFound, map[string]interface{}{
				"error":   "Role menu not found",
				"message": "Role menu with the specified ID does not exist",
			})
		}
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to fetch role menu",
			"message": err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"data": roleMenu,
	})
}

// CreateRoleMenu handles POST /roles-menus
func (c *RoleMenusController) CreateRoleMenu(ctx echo.Context) error {
	var req RoleMenuRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid request body",
			"message": err.Error(),
		})
	}

	// Validate required fields
	if req.RoleID == 0 || req.MenuID == 0 {
		return ctx.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Validation failed",
			"message": "role_id and menu_id are required",
		})
	}

	// Check if role exists
	var roleExists bool
	err := c.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users_roles WHERE roles_id = $1)", req.RoleID).Scan(&roleExists)
	if err != nil || !roleExists {
		return ctx.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid role",
			"message": "The specified role does not exist",
		})
	}

	// Check if menu exists
	var menuExists bool
	err = c.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM menus WHERE menus_id = $1)", req.MenuID).Scan(&menuExists)
	if err != nil || !menuExists {
		return ctx.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid menu",
			"message": "The specified menu does not exist",
		})
	}

	// Check if role-menu combination already exists
	var exists bool
	err = c.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM role_menus WHERE role_id = $1 AND menu_id = $2)", req.RoleID, req.MenuID).Scan(&exists)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to check existing role menu",
			"message": err.Error(),
		})
	}
	if exists {
		return ctx.JSON(http.StatusConflict, map[string]interface{}{
			"error":   "Role menu already exists",
			"message": "Permission for this role and menu combination already exists",
		})
	}

	query := `
		INSERT INTO role_menus (role_id, menu_id, can_view, can_create, can_modify, can_delete, can_upload, can_download)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING role_menu_id, created_at
	`

	var roleMenuID int
	var createdAt time.Time
	err = c.DB.QueryRow(query, req.RoleID, req.MenuID, req.CanView, req.CanCreate, req.CanModify, req.CanDelete, req.CanUpload, req.CanDownload).Scan(&roleMenuID, &createdAt)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to create role menu",
			"message": err.Error(),
		})
	}

	return ctx.JSON(http.StatusCreated, map[string]interface{}{
		"success": true,
		"message": "Role menu created successfully",
		"data": map[string]interface{}{
			"role_menu_id": roleMenuID,
			"created_at":   createdAt,
		},
	})
}

// UpdateRoleMenu handles PUT /roles-menus/:id
func (c *RoleMenusController) UpdateRoleMenu(ctx echo.Context) error {
	id := ctx.Param("id")
	
	var req RoleMenuRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid request body",
			"message": err.Error(),
		})
	}

	// Check if role menu exists
	var exists bool
	err := c.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM role_menus WHERE role_menu_id = $1)", id).Scan(&exists)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to check role menu existence",
			"message": err.Error(),
		})
	}
	if !exists {
		return ctx.JSON(http.StatusNotFound, map[string]interface{}{
			"error":   "Role menu not found",
			"message": "Role menu with the specified ID does not exist",
		})
	}

	// Update only the permission fields (role_id and menu_id should not be changed)
	query := `
		UPDATE role_menus 
		SET can_view = $1, can_create = $2, can_modify = $3, can_delete = $4, 
		    can_upload = $5, can_download = $6
		WHERE role_menu_id = $7
	`

	_, err = c.DB.Exec(query, req.CanView, req.CanCreate, req.CanModify, req.CanDelete, req.CanUpload, req.CanDownload, id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to update role menu",
			"message": err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Role menu updated successfully",
	})
}

// DeleteRoleMenu handles DELETE /roles-menus/:id
func (c *RoleMenusController) DeleteRoleMenu(ctx echo.Context) error {
	id := ctx.Param("id")

	// Check if role menu exists
	var exists bool
	err := c.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM role_menus WHERE role_menu_id = $1)", id).Scan(&exists)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to check role menu existence",
			"message": err.Error(),
		})
	}
	if !exists {
		return ctx.JSON(http.StatusNotFound, map[string]interface{}{
			"error":   "Role menu not found",
			"message": "Role menu with the specified ID does not exist",
		})
	}

	query := "DELETE FROM role_menus WHERE role_menu_id = $1"
	_, err = c.DB.Exec(query, id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to delete role menu",
			"message": err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Role menu deleted successfully",
	})
}

// GetAllRoles handles GET /users-roles (for dropdown)
func (c *RoleMenusController) GetAllRoles(ctx echo.Context) error {
	query := `SELECT roles_id, roles_name FROM users_roles`
	
	rows, err := c.DB.Query(query)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to fetch roles",
			"message": err.Error(),
		})
	}
	defer rows.Close()

	var roles []RoleDropdown
	for rows.Next() {
		var role RoleDropdown
		err := rows.Scan(&role.RolesID, &role.RolesName)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
				"error":   "Failed to scan role",
				"message": err.Error(),
			})
		}
		roles = append(roles, role)
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"data": roles,
	})
}

// GetAllMenus handles GET /menus (for dropdown)
func (c *RoleMenusController) GetAllMenus(ctx echo.Context) error {
	query := `SELECT menus_id, menu_name FROM menus WHERE is_active = true ORDER BY menu_name`
	
	rows, err := c.DB.Query(query)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to fetch menus",
			"message": err.Error(),
		})
	}
	defer rows.Close()

	var menus []MenuDropdown
	for rows.Next() {
		var menu MenuDropdown
		err := rows.Scan(&menu.MenusID, &menu.MenuName)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
				"error":   "Failed to scan menu",
				"message": err.Error(),
			})
		}
		menus = append(menus, menu)
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"data": menus,
	})
}

// BulkUpdatePermissions handles POST /roles-menus/bulk-update
func (c *RoleMenusController) BulkUpdatePermissions(ctx echo.Context) error {
	var req struct {
		RoleID      int                    `json:"role_id"`
		Permissions []RoleMenuRequest      `json:"permissions"`
	}
	
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid request body",
			"message": err.Error(),
		})
	}

	tx, err := c.DB.Begin()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to start transaction",
			"message": err.Error(),
		})
	}
	defer tx.Rollback()

	// Delete existing permissions for the role
	_, err = tx.Exec("DELETE FROM role_menus WHERE role_id = $1", req.RoleID)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to delete existing permissions",
			"message": err.Error(),
		})
	}

	// Insert new permissions
	for _, perm := range req.Permissions {
		query := `
			INSERT INTO role_menus (role_id, menu_id, can_view, can_create, can_modify, can_delete, can_upload, can_download)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`
		_, err = tx.Exec(query, req.RoleID, perm.MenuID, perm.CanView, perm.CanCreate, perm.CanModify, perm.CanDelete, perm.CanUpload, perm.CanDownload)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
				"error":   "Failed to insert permission",
				"message": err.Error(),
			})
		}
	}

	if err = tx.Commit(); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to commit transaction",
			"message": err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Permissions updated successfully",
	})
}

// CopyPermissions handles POST /roles-menus/copy-permissions
func (c *RoleMenusController) CopyPermissions(ctx echo.Context) error {
	var req struct {
		FromRoleID int `json:"from_role_id"`
		ToRoleID   int `json:"to_role_id"`
	}
	
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid request body",
			"message": err.Error(),
		})
	}

	tx, err := c.DB.Begin()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to start transaction",
			"message": err.Error(),
		})
	}
	defer tx.Rollback()

	// Delete existing permissions for target role
	_, err = tx.Exec("DELETE FROM role_menus WHERE role_id = $1", req.ToRoleID)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to delete existing permissions",
			"message": err.Error(),
		})
	}

	// Copy permissions from source role to target role
	query := `
		INSERT INTO role_menus (role_id, menu_id, can_view, can_create, can_modify, can_delete, can_upload, can_download)
		SELECT $1, menu_id, can_view, can_create, can_modify, can_delete, can_upload, can_download
		FROM role_menus WHERE role_id = $2
	`
	
	result, err := tx.Exec(query, req.ToRoleID, req.FromRoleID)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to copy permissions",
			"message": err.Error(),
		})
	}

	if err = tx.Commit(); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to commit transaction",
			"message": err.Error(),
		})
	}

	rowsAffected, _ := result.RowsAffected()
	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"success":       true,
		"message":       "Permissions copied successfully",
		"rows_affected": rowsAffected,
	})
}

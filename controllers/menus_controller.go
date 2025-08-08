package controller

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

type MenusController struct {
	DB *sql.DB
}

type Menu struct {
	MenusID    int     `json:"menus_id"`
	MenuCode   string  `json:"menu_code"`
	MenuName   string  `json:"menu_name"`
	ParentID   *int    `json:"parent_id"`
	ParentName *string `json:"parent_name"`
	IconName   *string `json:"icon_name"`
	Route      *string `json:"route"`
	MenuOrder  int     `json:"menu_order"`
	IsVisible  bool    `json:"is_visible"`
	IsActive   bool    `json:"is_active"`
	// Permissions
	CanView     bool `json:"can_view"`
	CanCreate   bool `json:"can_create"`
	CanModify   bool `json:"can_modify"`
	CanDelete   bool `json:"can_delete"`
	CanUpload   bool `json:"can_upload"`
	CanDownload bool `json:"can_download"`
}

type BreadcrumbItem struct {
	MenusID   int    `json:"menus_id"`
	MenuName  string `json:"menu_name"`
	Route     string `json:"route"`
	MenuOrder int    `json:"menu_order"`
	Level     int    `json:"level"`
}

func NewMenusController(db *sql.DB) *MenusController {
	return &MenusController{DB: db}
}

// GetUserMenus - Get all menus for a specific user using procedure
func (c *MenusController) GetUserMenus(ctx echo.Context) error {
	userID := ctx.QueryParam("user_id")
	if userID == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "user_id is required"})
	}

	// Call the stored procedure
	query := `SELECT * FROM security.get_user_menus($1)`

	rows, err := c.DB.Query(query, userID)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch user menus"})
	}
	defer rows.Close()

	var menus []Menu
	for rows.Next() {
		var menu Menu
		err := rows.Scan(
			&menu.MenusID,
			&menu.MenuCode,
			&menu.MenuName,
			&menu.ParentID,
			&menu.IconName,
			&menu.Route,
			&menu.MenuOrder,
			&menu.CanView,
			&menu.CanCreate,
			&menu.CanModify,
			&menu.CanDelete,
			&menu.CanUpload,
			&menu.CanDownload,
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to scan user menu"})
		}
		menus = append(menus, menu)
	}

	response := map[string]interface{}{
		"data":          menus,
		"total_records": len(menus),
	}

	return ctx.JSON(http.StatusOK, response)
}

// GetRootMenusForUser - Get only root menus (parent_id is NULL) for a user
func (c *MenusController) GetRootMenusForUser(ctx echo.Context) error {
	userID := ctx.QueryParam("user_id")
	if userID == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "user_id is required"})
	}

	// Call procedure and filter for root menus only
	query := `
		SELECT * FROM security.get_user_menus($1)
	`

	rows, err := c.DB.Query(query, userID)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch root menus"})
	}
	defer rows.Close()

	var menus []Menu
	for rows.Next() {
		var menu Menu
		err := rows.Scan(
			&menu.MenusID,
			&menu.MenuCode,
			&menu.MenuName,
			&menu.ParentID,
			&menu.IconName,
			&menu.Route,
			&menu.MenuOrder,
			&menu.CanView,
			&menu.CanCreate,
			&menu.CanModify,
			&menu.CanDelete,
			&menu.CanUpload,
			&menu.CanDownload,
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to scan root menu"})
		}
		menus = append(menus, menu)
	}

	response := map[string]interface{}{
		"data":          menus,
		"total_records": len(menus),
	}

	return ctx.JSON(http.StatusOK, response)
}

// GetMenuBreadcrumb - Using procedure for breadcrumb
func (c *MenusController) GetMenuBreadcrumb(ctx echo.Context) error {
	menuID := ctx.Param("id")
	userID := ctx.QueryParam("user_id")

	if userID == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "user_id is required"})
	}

	// First, get the target menu ID
	var targetMenuID int
	var err error

	if id, parseErr := strconv.Atoi(menuID); parseErr == nil {
		targetMenuID = id
	} else {
		// Search by route in user's accessible menus
		query := `
			SELECT menus_id FROM security.get_user_menus($1) 
			WHERE route = $2 LIMIT 1
		`
		err = c.DB.QueryRow(query, userID, menuID).Scan(&targetMenuID)
		if err != nil {
			if err == sql.ErrNoRows {
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Menu not found or not accessible"})
			}
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find menu"})
		}
	}

	// Get breadcrumb using recursive CTE with user permissions
	query := `
		WITH RECURSIVE breadcrumb_path AS (
			-- Base case: start with the target menu
			SELECT 
				um.menus_id,
				um.menu_name,
				COALESCE(um.route, '') as route,
				um.menu_order,
				um.parent_id,
				0 as level
			FROM security.get_user_menus($1) um
			WHERE um.menus_id = $2
			
			UNION ALL
			
			-- Recursive case: get parent menus
			SELECT 
				um.menus_id,
				um.menu_name,
				COALESCE(um.route, '') as route,
				um.menu_order,
				um.parent_id,
				bp.level + 1 as level
			FROM security.get_user_menus($1) um
			INNER JOIN breadcrumb_path bp ON um.menus_id = bp.parent_id
		)
		SELECT menus_id, menu_name, route, menu_order, level
		FROM breadcrumb_path
		ORDER BY level DESC
	`

	rows, err := c.DB.Query(query, userID, targetMenuID)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch breadcrumb"})
	}
	defer rows.Close()

	var breadcrumbs []BreadcrumbItem
	for rows.Next() {
		var item BreadcrumbItem
		err := rows.Scan(
			&item.MenusID,
			&item.MenuName,
			&item.Route,
			&item.MenuOrder,
			&item.Level,
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to scan breadcrumb item"})
		}
		breadcrumbs = append(breadcrumbs, item)
	}

	if len(breadcrumbs) == 0 {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Menu not found or not accessible"})
	}

	response := map[string]interface{}{
		"data": breadcrumbs,
	}

	return ctx.JSON(http.StatusOK, response)
}

// GetChildMenusForUser - Get child menus for a parent using procedure
func (c *MenusController) GetChildMenusForUser(ctx echo.Context) error {
	parentID := ctx.Param("parent_id")
	userID := ctx.QueryParam("user_id")

	if userID == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "user_id is required"})
	}

	query := `
		SELECT * FROM security.get_user_menus($1)
		WHERE parent_id = $2
		ORDER BY menu_order, menu_name
	`

	rows, err := c.DB.Query(query, userID, parentID)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch child menus"})
	}
	defer rows.Close()

	var menus []Menu
	for rows.Next() {
		var menu Menu
		err := rows.Scan(
			&menu.MenusID,
			&menu.MenuCode,
			&menu.MenuName,
			&menu.ParentID,
			&menu.IconName,
			&menu.Route,
			&menu.MenuOrder,
			&menu.CanView,
			&menu.CanCreate,
			&menu.CanModify,
			&menu.CanDelete,
			&menu.CanUpload,
			&menu.CanDownload,
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to scan child menu"})
		}
		menus = append(menus, menu)
	}

	response := map[string]interface{}{
		"data":          menus,
		"total_records": len(menus),
	}

	return ctx.JSON(http.StatusOK, response)
}

// GetAllDescendantsForUser - Get all descendants for a user using procedure
func (c *MenusController) GetAllDescendantsForUser(ctx echo.Context) error {
	parentID := ctx.Param("parent_id")
	userID := ctx.QueryParam("user_id")

	if userID == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "user_id is required"})
	}

	// Get all user menus first, then filter descendants
	query := `
		WITH RECURSIVE menu_descendants AS (
			-- Base case: direct children
			SELECT 
				um.*,
				0 as level,
				CAST(um.menu_order AS text) as path
			FROM security.get_user_menus($1) um
			WHERE um.parent_id = $2
			
			UNION ALL
			
			-- Recursive case: children of children
			SELECT 
				um.*,
				md.level + 1 as level,
				md.path || '.' || CAST(um.menu_order AS text) as path
			FROM security.get_user_menus($1) um
			INNER JOIN menu_descendants md ON um.parent_id = md.menus_id
		)
		SELECT 
			menus_id, menu_code, menu_name, parent_id, icon_name,
			route, menu_order, can_view, can_create, can_modify,
			can_delete, can_upload, can_download
		FROM menu_descendants
		ORDER BY path
	`

	rows, err := c.DB.Query(query, userID, parentID)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch descendants"})
	}
	defer rows.Close()

	var menus []Menu
	for rows.Next() {
		var menu Menu
		err := rows.Scan(
			&menu.MenusID,
			&menu.MenuCode,
			&menu.MenuName,
			&menu.ParentID,
			&menu.IconName,
			&menu.Route,
			&menu.MenuOrder,
			&menu.CanView,
			&menu.CanCreate,
			&menu.CanModify,
			&menu.CanDelete,
			&menu.CanUpload,
			&menu.CanDownload,
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to scan descendant menu"})
		}
		menus = append(menus, menu)
	}

	response := map[string]interface{}{
		"data":          menus,
		"total_records": len(menus),
	}

	return ctx.JSON(http.StatusOK, response)
}

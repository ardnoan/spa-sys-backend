package controller

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

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
	CreatedAt  string  `json:"created_at"`
	CreatedBy  *string `json:"created_by"`
	UpdatedAt  string  `json:"updated_at"`
	UpdatedBy  *string `json:"updated_by"`
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

func (c *MenusController) GetAllMenus(ctx echo.Context) error {
	page, _ := strconv.Atoi(ctx.QueryParam("page"))
	limit, _ := strconv.Atoi(ctx.QueryParam("limit"))
	search := ctx.QueryParam("search")
	parentID := ctx.QueryParam("parent_id")
	onlyActive := ctx.QueryParam("only_active") == "true"
	onlyVisible := ctx.QueryParam("only_visible") == "true"

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	// Base queries
	countQuery := `
		SELECT COUNT(*) 
		FROM menus m
		LEFT JOIN menus p ON m.parent_id = p.menus_id
		WHERE 1=1`
	dataQuery := `
		SELECT m.menus_id, m.menu_code, m.menu_name, m.parent_id, p.menu_name as parent_name,
		       m.icon_name, m.route, m.menu_order, m.is_visible, m.is_active,
		       m.created_at, m.created_by, m.updated_at, m.updated_by
		FROM menus m
		LEFT JOIN menus p ON m.parent_id = p.menus_id
		WHERE 1=1`

	var conditions []string
	var args []interface{}
	argIndex := 1

	// Filter: parent_id
	if parentID != "" {
		if parentID == "null" {
			conditions = append(conditions, "m.parent_id IS NULL")
		} else {
			conditions = append(conditions, fmt.Sprintf("m.parent_id = $%d", argIndex))
			args = append(args, parentID)
			argIndex++
		}
	}

	// Filter: search
	if search != "" {
		cond := fmt.Sprintf(`(
			m.menu_code ILIKE $%d OR 
			m.menu_name ILIKE $%d OR 
			p.menu_name ILIKE $%d
		)`, argIndex, argIndex, argIndex)
		conditions = append(conditions, cond)
		args = append(args, "%"+search+"%")
		argIndex++
	}

	// Filter: is_active
	if onlyActive {
		conditions = append(conditions, fmt.Sprintf("m.is_active = $%d", argIndex))
		args = append(args, true)
		argIndex++
	}

	// Filter: is_visible
	if onlyVisible {
		conditions = append(conditions, fmt.Sprintf("m.is_visible = $%d", argIndex))
		args = append(args, true)
		argIndex++
	}

	// Combine conditions
	conditionString := ""
	if len(conditions) > 0 {
		conditionString = " AND " + strings.Join(conditions, " AND ")
	}

	// Final queries
	finalCountQuery := countQuery + conditionString
	finalDataQuery := dataQuery + conditionString + `
		ORDER BY 
			CASE WHEN m.parent_id IS NULL THEN 0 ELSE 1 END,
			COALESCE(m.parent_id, 0),
			m.menu_order,
			m.menu_name
		LIMIT $` + strconv.Itoa(argIndex) + ` OFFSET $` + strconv.Itoa(argIndex+1)
	args = append(args, limit, offset)

	// Count total
	var totalRecords int
	if err := c.DB.QueryRow(finalCountQuery, args[:len(args)-2]...).Scan(&totalRecords); err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to count records"})
	}

	// Fetch data
	rows, err := c.DB.Query(finalDataQuery, args...)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to fetch menus"})
	}
	defer rows.Close()

	// Scan rows
	var menus []Menu
	for rows.Next() {
		var menu Menu
		if err := rows.Scan(
			&menu.MenusID,
			&menu.MenuCode,
			&menu.MenuName,
			&menu.ParentID,
			&menu.ParentName,
			&menu.IconName,
			&menu.Route,
			&menu.MenuOrder,
			&menu.IsVisible,
			&menu.IsActive,
			&menu.CreatedAt,
			&menu.CreatedBy,
			&menu.UpdatedAt,
			&menu.UpdatedBy,
		); err != nil {
			return ctx.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to scan menu"})
		}
		menus = append(menus, menu)
	}

	// Response
	return ctx.JSON(http.StatusOK, echo.Map{
		"data": menus,
		"pagination": echo.Map{
			"current_page":     page,
			"total_pages":      (totalRecords + limit - 1) / limit,
			"total_records":    totalRecords,
			"records_per_page": limit,
		},
	})
}

// GetRootMenus - Get all root menus (parent_id is NULL)
func (c *MenusController) GetRootMenus(ctx echo.Context) error {
	search := ctx.QueryParam("search")
	onlyActive := ctx.QueryParam("only_active") == "true"
	onlyVisible := ctx.QueryParam("only_visible") == "true"

	var query string
	var args []interface{}
	argIndex := 0

	baseQuery := `
		SELECT m.menus_id, m.menu_code, m.menu_name, m.parent_id, 
		       NULL as parent_name, m.icon_name, m.route, m.menu_order, 
		       m.is_visible, m.is_active, m.created_at, m.created_by, 
		       m.updated_at, m.updated_by
		FROM menus m
		WHERE m.parent_id IS NULL
	`

	conditions := []string{}

	if search != "" {
		argIndex++
		conditions = append(conditions, "(m.menu_code ILIKE $"+strconv.Itoa(argIndex)+" OR m.menu_name ILIKE $"+strconv.Itoa(argIndex)+")")
		args = append(args, "%"+search+"%")
	}

	if onlyActive {
		argIndex++
		conditions = append(conditions, "m.is_active = $"+strconv.Itoa(argIndex))
		args = append(args, true)
	}

	if onlyVisible {
		argIndex++
		conditions = append(conditions, "m.is_visible = $"+strconv.Itoa(argIndex))
		args = append(args, true)
	}

	if len(conditions) > 0 {
		query = baseQuery + " AND " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			query += " AND " + conditions[i]
		}
	} else {
		query = baseQuery
	}

	query += " ORDER BY m.menu_order, m.menu_name"

	rows, err := c.DB.Query(query, args...)
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
			&menu.ParentName,
			&menu.IconName,
			&menu.Route,
			&menu.MenuOrder,
			&menu.IsVisible,
			&menu.IsActive,
			&menu.CreatedAt,
			&menu.CreatedBy,
			&menu.UpdatedAt,
			&menu.UpdatedBy,
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

// GetMenuById - Get specific menu by ID
func (c *MenusController) GetMenuById(ctx echo.Context) error {
	id := ctx.Param("id")

	query := `
		SELECT m.menus_id, m.menu_code, m.menu_name, m.parent_id, p.menu_name as parent_name,
			   m.icon_name, m.route, m.menu_order, m.is_visible, m.is_active,
			   m.created_at, m.created_by, m.updated_at, m.updated_by
		FROM menus m
		LEFT JOIN menus p ON m.parent_id = p.menus_id
		WHERE m.menus_id = $1
	`

	var menu Menu
	err := c.DB.QueryRow(query, id).Scan(
		&menu.MenusID,
		&menu.MenuCode,
		&menu.MenuName,
		&menu.ParentID,
		&menu.ParentName,
		&menu.IconName,
		&menu.Route,
		&menu.MenuOrder,
		&menu.IsVisible,
		&menu.IsActive,
		&menu.CreatedAt,
		&menu.CreatedBy,
		&menu.UpdatedAt,
		&menu.UpdatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Menu not found"})
		}
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch menu"})
	}

	response := map[string]interface{}{
		"data": menu,
	}

	return ctx.JSON(http.StatusOK, response)
}

// GetMenuBreadcrumb - Get breadcrumb path for a specific menu
func (c *MenusController) GetMenuBreadcrumb(ctx echo.Context) error {
	menuID := ctx.Param("id")

	// Find menu by ID first (support both menus_id and route)
	var targetMenuID int
	var err error

	// Try to parse as integer first
	if id, parseErr := strconv.Atoi(menuID); parseErr == nil {
		targetMenuID = id
	} else {
		// If not integer, search by route
		query := `SELECT menus_id FROM menus WHERE route = $1`
		err = c.DB.QueryRow(query, menuID).Scan(&targetMenuID)
		if err != nil {
			if err == sql.ErrNoRows {
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Menu not found"})
			}
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find menu"})
		}
	}

	// Recursive CTE to get breadcrumb path
	query := `
		WITH RECURSIVE breadcrumb_path AS (
			-- Base case: start with the target menu
			SELECT 
				m.menus_id,
				m.menu_name,
				COALESCE(m.route, '') as route,
				m.menu_order,
				m.parent_id,
				0 as level
			FROM menus m
			WHERE m.menus_id = $1
			
			UNION ALL
			
			-- Recursive case: get parent menus
			SELECT 
				m.menus_id,
				m.menu_name,
				COALESCE(m.route, '') as route,
				m.menu_order,
				m.parent_id,
				bp.level + 1 as level
			FROM menus m
			INNER JOIN breadcrumb_path bp ON m.menus_id = bp.parent_id
		)
		SELECT menus_id, menu_name, route, menu_order, level
		FROM breadcrumb_path
		ORDER BY level DESC
	`

	rows, err := c.DB.Query(query, targetMenuID)
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
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Menu not found"})
	}

	response := map[string]interface{}{
		"data": breadcrumbs,
	}

	return ctx.JSON(http.StatusOK, response)
}

// GetChildMenus - Get all child menus for a parent
func (c *MenusController) GetChildMenus(ctx echo.Context) error {
	parentID := ctx.Param("parent_id")
	onlyActive := ctx.QueryParam("only_active") == "true"
	onlyVisible := ctx.QueryParam("only_visible") == "true"

	var query string
	var args []interface{}
	argIndex := 0

	baseQuery := `
		SELECT m.menus_id, m.menu_code, m.menu_name, m.parent_id, p.menu_name as parent_name,
			   m.icon_name, m.route, m.menu_order, m.is_visible, m.is_active,
			   m.created_at, m.created_by, m.updated_at, m.updated_by
		FROM menus m
		LEFT JOIN menus p ON m.parent_id = p.menus_id
		WHERE m.parent_id = $1
	`

	argIndex++
	args = append(args, parentID)

	conditions := []string{}

	if onlyActive {
		argIndex++
		conditions = append(conditions, "m.is_active = $"+strconv.Itoa(argIndex))
		args = append(args, true)
	}

	if onlyVisible {
		argIndex++
		conditions = append(conditions, "m.is_visible = $"+strconv.Itoa(argIndex))
		args = append(args, true)
	}

	if len(conditions) > 0 {
		query = baseQuery + " AND " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			query += " AND " + conditions[i]
		}
	} else {
		query = baseQuery
	}

	query += " ORDER BY m.menu_order, m.menu_name"

	rows, err := c.DB.Query(query, args...)
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
			&menu.ParentName,
			&menu.IconName,
			&menu.Route,
			&menu.MenuOrder,
			&menu.IsVisible,
			&menu.IsActive,
			&menu.CreatedAt,
			&menu.CreatedBy,
			&menu.UpdatedAt,
			&menu.UpdatedBy,
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

// GetHierarchicalMenus - Get complete hierarchical menu structure
func (c *MenusController) GetHierarchicalMenus(ctx echo.Context) error {
	parentID := ctx.QueryParam("parent_id")
	onlyActive := ctx.QueryParam("only_active") == "true"
	onlyVisible := ctx.QueryParam("only_visible") == "true"

	var query string
	var args []interface{}
	argIndex := 0

	baseQuery := `
		WITH RECURSIVE menu_hierarchy AS (
			-- Base case: get root menus or menus with specific parent
			SELECT 
				m.menus_id,
				m.menu_code,
				m.menu_name,
				m.parent_id,
				m.icon_name,
				m.route,
				m.menu_order,
				m.is_visible,
				m.is_active,
				m.created_at,
				m.created_by,
				m.updated_at,
				m.updated_by,
				0 as level,
				CAST(m.menu_order AS text) as path
			FROM menus m
			WHERE 1=1
	`

	conditions := []string{}

	// Add parent filter
	if parentID != "" {
		if parentID == "null" {
			conditions = append(conditions, "m.parent_id IS NULL")
		} else {
			argIndex++
			conditions = append(conditions, "m.parent_id = $"+strconv.Itoa(argIndex))
			args = append(args, parentID)
		}
	} else {
		conditions = append(conditions, "m.parent_id IS NULL")
	}

	// Add active filter
	if onlyActive {
		argIndex++
		conditions = append(conditions, "m.is_active = $"+strconv.Itoa(argIndex))
		args = append(args, true)
	}

	// Add visible filter
	if onlyVisible {
		argIndex++
		conditions = append(conditions, "m.is_visible = $"+strconv.Itoa(argIndex))
		args = append(args, true)
	}

	// Build base query with conditions
	if len(conditions) > 0 {
		baseQuery += " AND " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			baseQuery += " AND " + conditions[i]
		}
	}

	// Add recursive part
	query = baseQuery + `
			UNION ALL
			
			-- Recursive case: get child menus
			SELECT 
				m.menus_id,
				m.menu_code,
				m.menu_name,
				m.parent_id,
				m.icon_name,
				m.route,
				m.menu_order,
				m.is_visible,
				m.is_active,
				m.created_at,
				m.created_by,
				m.updated_at,
				m.updated_by,
				mh.level + 1 as level,
				mh.path || '.' || CAST(m.menu_order AS text) as path
			FROM menus m
			INNER JOIN menu_hierarchy mh ON m.parent_id = mh.menus_id
	`

	// Add filters for recursive part
	if onlyActive {
		query += " WHERE m.is_active = true"
		if onlyVisible {
			query += " AND m.is_visible = true"
		}
	} else if onlyVisible {
		query += " WHERE m.is_visible = true"
	}

	query += `
		)
		SELECT menus_id, menu_code, menu_name, parent_id, icon_name, route, 
			   menu_order, is_visible, is_active, created_at, created_by, 
			   updated_at, updated_by, level
		FROM menu_hierarchy
		ORDER BY path
	`

	rows, err := c.DB.Query(query, args...)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch hierarchical menus"})
	}
	defer rows.Close()

	var menus []Menu
	for rows.Next() {
		var menu Menu
		var level int
		err := rows.Scan(
			&menu.MenusID,
			&menu.MenuCode,
			&menu.MenuName,
			&menu.ParentID,
			&menu.IconName,
			&menu.Route,
			&menu.MenuOrder,
			&menu.IsVisible,
			&menu.IsActive,
			&menu.CreatedAt,
			&menu.CreatedBy,
			&menu.UpdatedAt,
			&menu.UpdatedBy,
			&level,
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to scan hierarchical menu"})
		}
		menus = append(menus, menu)
	}

	response := map[string]interface{}{
		"data":          menus,
		"total_records": len(menus),
	}

	return ctx.JSON(http.StatusOK, response)
}
func (c *MenusController) GetMenuByRoute(ctx echo.Context) error {
	route := ctx.Param("route")

	// URL decode the route parameter
	decodedRoute, err := url.QueryUnescape(route)
	if err != nil {
		decodedRoute = route // fallback to original if decode fails
	}

	query := `
		SELECT m.menus_id, m.menu_code, m.menu_name, m.parent_id, p.menu_name as parent_name,
			   m.icon_name, m.route, m.menu_order, m.is_visible, m.is_active,
			   m.created_at, m.created_by, m.updated_at, m.updated_by
		FROM menus m
		LEFT JOIN menus p ON m.parent_id = p.menus_id
		WHERE m.route = $1 OR m.route = $2
	`

	var menu Menu
	err = c.DB.QueryRow(query, route, decodedRoute).Scan(
		&menu.MenusID,
		&menu.MenuCode,
		&menu.MenuName,
		&menu.ParentID,
		&menu.ParentName,
		&menu.IconName,
		&menu.Route,
		&menu.MenuOrder,
		&menu.IsVisible,
		&menu.IsActive,
		&menu.CreatedAt,
		&menu.CreatedBy,
		&menu.UpdatedAt,
		&menu.UpdatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Menu not found"})
		}
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch menu"})
	}

	response := map[string]interface{}{
		"data": menu,
	}

	return ctx.JSON(http.StatusOK, response)
}

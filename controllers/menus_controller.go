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
	CreatedAt  string  `json:"created_at"`
	CreatedBy  *string `json:"created_by"`
	UpdatedAt  string  `json:"updated_at"`
	UpdatedBy  *string `json:"updated_by"`
}

func NewMenusController(db *sql.DB) *MenusController {
	return &MenusController{DB: db}
}

func (c *MenusController) GetAllMenus(ctx echo.Context) error {
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
			FROM menus m
			LEFT JOIN menus p ON m.parent_id = p.menus_id
			WHERE m.menu_code ILIKE $1 OR m.menu_name ILIKE $1 OR p.menu_name ILIKE $1
		`

		query = `
			SELECT m.menus_id, m.menu_code, m.menu_name, m.parent_id, p.menu_name as parent_name,
				   m.icon_name, m.route, m.menu_order, m.is_visible, m.is_active,
				   m.created_at, m.created_by, m.updated_at, m.updated_by
			FROM menus m
			LEFT JOIN menus p ON m.parent_id = p.menus_id
			WHERE m.menu_code ILIKE $1 OR m.menu_name ILIKE $1 OR p.menu_name ILIKE $1
			ORDER BY m.menu_order, m.menu_name
			LIMIT $2 OFFSET $3
		`
		searchParam := "%" + search + "%"
		args = []interface{}{searchParam, limit, offset}
	} else {
		countQuery = `
			SELECT COUNT(*) 
			FROM menus m
			LEFT JOIN menus p ON m.parent_id = p.menus_id
		`

		query = `
			SELECT m.menus_id, m.menu_code, m.menu_name, m.parent_id, p.menu_name as parent_name,
				   m.icon_name, m.route, m.menu_order, m.is_visible, m.is_active,
				   m.created_at, m.created_by, m.updated_at, m.updated_by
			FROM menus m
			LEFT JOIN menus p ON m.parent_id = p.menus_id
			ORDER BY m.menu_order, m.menu_name
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
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch menus"})
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
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to scan menu"})
		}
		menus = append(menus, menu)
	}

	totalPages := (totalRecords + limit - 1) / limit

	response := map[string]interface{}{
		"data": menus,
		"pagination": map[string]interface{}{
			"current_page":     page,
			"total_pages":      totalPages,
			"total_records":    totalRecords,
			"records_per_page": limit,
		},
	}

	return ctx.JSON(http.StatusOK, response)
}

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
		conditions = append(conditions, "m.menu_code ILIKE $"+strconv.Itoa(argIndex)+" OR m.menu_name ILIKE $"+strconv.Itoa(argIndex))
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

package controller

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type RoleMenusController struct {
	DB *sql.DB
}

type RoleMenu struct {
	RoleMenuID  int    `json:"role_menu_id"`
	RoleID      int    `json:"role_id"`
	RoleName    string `json:"role_name"`
	MenuID      int    `json:"menu_id"`
	MenuName    string `json:"menu_name"`
	CanView     bool   `json:"can_view"`
	CanCreate   bool   `json:"can_create"`
	CanModify   bool   `json:"can_modify"`
	CanDelete   bool   `json:"can_delete"`
	CanUpload   bool   `json:"can_upload"`
	CanDownload bool   `json:"can_download"`
	CreatedAt   string `json:"created_at"`
	CreatedBy   *int   `json:"created_by"`
}

func NewRoleMenusController(db *sql.DB) *RoleMenusController {
	return &RoleMenusController{DB: db}
}

func (c *RoleMenusController) GetAllRoleMenus(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	search := ctx.Query("search")

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
			FROM role_menus rm
			JOIN users_roles ur ON rm.role_id = ur.roles_id
			JOIN menus m ON rm.menu_id = m.menus_id
			WHERE ur.roles_name ILIKE $1 OR m.menu_name ILIKE $1
		`

		query = `
			SELECT rm.role_menu_id, rm.role_id, ur.roles_name, rm.menu_id, m.menu_name,
				   rm.can_view, rm.can_create, rm.can_modify, rm.can_delete, 
				   rm.can_upload, rm.can_download, rm.created_at, rm.created_by
			FROM role_menus rm
			JOIN users_roles ur ON rm.role_id = ur.roles_id
			JOIN menus m ON rm.menu_id = m.menus_id
			WHERE ur.roles_name ILIKE $1 OR m.menu_name ILIKE $1
			ORDER BY ur.roles_name, m.menu_name
			LIMIT $2 OFFSET $3
		`
		searchParam := "%" + search + "%"
		args = []interface{}{searchParam, limit, offset}
	} else {
		countQuery = `
			SELECT COUNT(*) 
			FROM role_menus rm
			JOIN users_roles ur ON rm.role_id = ur.roles_id
			JOIN menus m ON rm.menu_id = m.menus_id
		`

		query = `
			SELECT rm.role_menu_id, rm.role_id, ur.roles_name, rm.menu_id, m.menu_name,
				   rm.can_view, rm.can_create, rm.can_modify, rm.can_delete, 
				   rm.can_upload, rm.can_download, rm.created_at, rm.created_by
			FROM role_menus rm
			JOIN users_roles ur ON rm.role_id = ur.roles_id
			JOIN menus m ON rm.menu_id = m.menus_id
			ORDER BY ur.roles_name, m.menu_name
			LIMIT $1 OFFSET $2
		`
		args = []interface{}{limit, offset}
	}

	var totalRecords int
	if search != "" {
		err := c.DB.QueryRow(countQuery, "%"+search+"%").Scan(&totalRecords)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count records"})
			return
		}
	} else {
		err := c.DB.QueryRow(countQuery).Scan(&totalRecords)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count records"})
			return
		}
	}

	rows, err := c.DB.Query(query, args...)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch role menus"})
		return
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
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan role menu"})
			return
		}
		roleMenus = append(roleMenus, roleMenu)
	}

	totalPages := (totalRecords + limit - 1) / limit

	response := gin.H{
		"data": roleMenus,
		"pagination": gin.H{
			"current_page":     page,
			"total_pages":      totalPages,
			"total_records":    totalRecords,
			"records_per_page": limit,
		},
	}

	ctx.JSON(http.StatusOK, response)
}

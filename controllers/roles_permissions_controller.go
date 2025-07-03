package controller

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type RolePermissionsController struct {
	DB *sql.DB
}

type RolePermission struct {
	RolePermissionID int    `json:"role_permission_id"`
	RoleID           int    `json:"role_id"`
	RoleName         string `json:"role_name"`
	PermissionID     int    `json:"permission_id"`
	PermissionName   string `json:"permission_name"`
	PermissionCode   string `json:"permission_code"`
	GrantedAt        string `json:"granted_at"`
	GrantedBy        *int   `json:"granted_by"`
	IsActive         bool   `json:"is_active"`
}

func NewRolePermissionsController(db *sql.DB) *RolePermissionsController {
	return &RolePermissionsController{DB: db}
}

func (c *RolePermissionsController) GetAllRolePermissions(ctx *gin.Context) {
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
			FROM role_permissions rp
			JOIN users_roles ur ON rp.role_id = ur.roles_id
			JOIN permissions p ON rp.permission_id = p.permissions_id
			WHERE ur.roles_name ILIKE $1 OR p.permission_name ILIKE $1 OR p.permission_code ILIKE $1
		`

		query = `
			SELECT rp.role_permission_id, rp.role_id, ur.roles_name, rp.permission_id, 
				   p.permission_name, p.permission_code, rp.granted_at, rp.granted_by, rp.is_active
			FROM role_permissions rp
			JOIN users_roles ur ON rp.role_id = ur.roles_id
			JOIN permissions p ON rp.permission_id = p.permissions_id
			WHERE ur.roles_name ILIKE $1 OR p.permission_name ILIKE $1 OR p.permission_code ILIKE $1
			ORDER BY ur.roles_name, p.permission_name
			LIMIT $2 OFFSET $3
		`
		searchParam := "%" + search + "%"
		args = []interface{}{searchParam, limit, offset}
	} else {
		countQuery = `
			SELECT COUNT(*) 
			FROM role_permissions rp
			JOIN users_roles ur ON rp.role_id = ur.roles_id
			JOIN permissions p ON rp.permission_id = p.permissions_id
		`

		query = `
			SELECT rp.role_permission_id, rp.role_id, ur.roles_name, rp.permission_id, 
				   p.permission_name, p.permission_code, rp.granted_at, rp.granted_by, rp.is_active
			FROM role_permissions rp
			JOIN users_roles ur ON rp.role_id = ur.roles_id
			JOIN permissions p ON rp.permission_id = p.permissions_id
			ORDER BY ur.roles_name, p.permission_name
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
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch role permissions"})
		return
	}
	defer rows.Close()

	var rolePermissions []RolePermission
	for rows.Next() {
		var rolePermission RolePermission
		err := rows.Scan(
			&rolePermission.RolePermissionID,
			&rolePermission.RoleID,
			&rolePermission.RoleName,
			&rolePermission.PermissionID,
			&rolePermission.PermissionName,
			&rolePermission.PermissionCode,
			&rolePermission.GrantedAt,
			&rolePermission.GrantedBy,
			&rolePermission.IsActive,
		)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan role permission"})
			return
		}
		rolePermissions = append(rolePermissions, rolePermission)
	}

	totalPages := (totalRecords + limit - 1) / limit

	response := gin.H{
		"data": rolePermissions,
		"pagination": gin.H{
			"current_page":     page,
			"total_pages":      totalPages,
			"total_records":    totalRecords,
			"records_per_page": limit,
		},
	}

	ctx.JSON(http.StatusOK, response)
}

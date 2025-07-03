package controller

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

type UserRole struct {
	ID         int       `json:"id" db:"user_role_id"`
	UserID     int       `json:"user_id" db:"user_id"`
	RoleID     int       `json:"role_id" db:"role_id"`
	AssignedAt time.Time `json:"assigned_at" db:"assigned_at"`
	AssignedBy *int      `json:"assigned_by" db:"assigned_by"`
	IsActive   bool      `json:"is_active" db:"is_active"`

	// Joined fields
	Username  string `json:"username" db:"username"`
	FirstName string `json:"first_name" db:"first_name"`
	LastName  string `json:"last_name" db:"last_name"`
	RoleName  string `json:"role_name" db:"roles_name"`
	RoleCode  string `json:"role_code" db:"roles_code"`
}

type UserRoleController struct {
	DB *sql.DB
}

func NewUserRoleController(db *sql.DB) *UserRoleController {
	return &UserRoleController{DB: db}
}

func (urc *UserRoleController) successResponse(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

func (urc *UserRoleController) errorResponse(c echo.Context, code int, message string) error {
	return c.JSON(code, map[string]interface{}{
		"success": false,
		"message": message,
	})
}

func (urc *UserRoleController) GetAllUserRoles(c echo.Context) error {
	page := c.QueryParam("page")
	limit := c.QueryParam("limit")
	search := c.QueryParam("search")
	userID := c.QueryParam("user_id")
	roleID := c.QueryParam("role_id")
	isActive := c.QueryParam("is_active")

	pageInt := 1
	limitInt := 10
	if p, err := strconv.Atoi(page); err == nil && p > 0 {
		pageInt = p
	}
	if l, err := strconv.Atoi(limit); err == nil && l > 0 && l <= 100 {
		limitInt = l
	}
	offset := (pageInt - 1) * limitInt

	whereConditions := []string{}
	args := []interface{}{}
	argIndex := 1

	if search != "" {
		searchPattern := "%" + strings.ToLower(search) + "%"
		condition := `(LOWER(u.username) LIKE $` + strconv.Itoa(argIndex) +
			` OR LOWER(u.first_name) LIKE $` + strconv.Itoa(argIndex) +
			` OR LOWER(u.last_name) LIKE $` + strconv.Itoa(argIndex) +
			` OR LOWER(r.roles_name) LIKE $` + strconv.Itoa(argIndex) + `)`
		whereConditions = append(whereConditions, condition)
		args = append(args, searchPattern)
		argIndex++
	}

	if userID != "" {
		if uid, err := strconv.Atoi(userID); err == nil {
			whereConditions = append(whereConditions, "ur.user_id = $"+strconv.Itoa(argIndex))
			args = append(args, uid)
			argIndex++
		}
	}

	if roleID != "" {
		if rid, err := strconv.Atoi(roleID); err == nil {
			whereConditions = append(whereConditions, "ur.role_id = $"+strconv.Itoa(argIndex))
			args = append(args, rid)
			argIndex++
		}
	}

	if isActive != "" {
		if active, err := strconv.ParseBool(isActive); err == nil {
			whereConditions = append(whereConditions, "ur.is_active = $"+strconv.Itoa(argIndex))
			args = append(args, active)
			argIndex++
		}
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	var totalCount int
	countQuery := `SELECT COUNT(*) FROM user_roles ur
                   JOIN users_application u ON ur.user_id = u.user_apps_id
                   JOIN users_roles r ON ur.role_id = r.roles_id ` + whereClause
	if err := urc.DB.QueryRow(countQuery, args...).Scan(&totalCount); err != nil {
		return urc.errorResponse(c, http.StatusInternalServerError, "Failed to count user roles")
	}

	query := `SELECT ur.user_role_id, ur.user_id, ur.role_id, ur.assigned_at, ur.assigned_by,
              ur.is_active, u.username, u.first_name, u.last_name, r.roles_name, r.roles_code
              FROM user_roles ur
              JOIN users_application u ON ur.user_id = u.user_apps_id
              JOIN users_roles r ON ur.role_id = r.roles_id ` + whereClause + `
              ORDER BY ur.assigned_at DESC
              LIMIT $` + strconv.Itoa(argIndex) + ` OFFSET $` + strconv.Itoa(argIndex+1)

	args = append(args, limitInt, offset)

	rows, err := urc.DB.Query(query, args...)
	if err != nil {
		return urc.errorResponse(c, http.StatusInternalServerError, "Failed to fetch user roles")
	}
	defer rows.Close()

	var userRoles []UserRole
	for rows.Next() {
		var ur UserRole
		if err := rows.Scan(
			&ur.ID, &ur.UserID, &ur.RoleID, &ur.AssignedAt, &ur.AssignedBy,
			&ur.IsActive, &ur.Username, &ur.FirstName, &ur.LastName, &ur.RoleName, &ur.RoleCode,
		); err == nil {
			userRoles = append(userRoles, ur)
		}
	}

	totalPages := (totalCount + limitInt - 1) / limitInt
	return urc.successResponse(c, map[string]interface{}{
		"user_roles": userRoles,
		"pagination": map[string]interface{}{
			"current_page": pageInt,
			"per_page":     limitInt,
			"total_count":  totalCount,
			"total_pages":  totalPages,
			"has_next":     pageInt < totalPages,
			"has_prev":     pageInt > 1,
		},
	})
}

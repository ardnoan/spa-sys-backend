package controller

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

type UserPasswordHistoryController struct {
	DB *sql.DB
}

type UserPasswordHistory struct {
	HistoryID    int    `json:"history_id"`
	UserID       int    `json:"user_id"`
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash"`
	CreatedAt    string `json:"created_at"`
}

func NewUserPasswordHistoryController(db *sql.DB) *UserPasswordHistoryController {
	return &UserPasswordHistoryController{DB: db}
}

func (c *UserPasswordHistoryController) GetAllUserPasswordHistory(ctx echo.Context) error {
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
			FROM user_password_history uph
			JOIN users_application ua ON uph.user_id = ua.user_apps_id
			WHERE ua.username ILIKE $1
		`

		query = `
			SELECT uph.history_id, uph.user_id, ua.username, uph.password_hash, uph.created_at
			FROM user_password_history uph
			JOIN users_application ua ON uph.user_id = ua.user_apps_id
			WHERE ua.username ILIKE $1
			ORDER BY uph.created_at DESC
			LIMIT $2 OFFSET $3
		`
		searchParam := "%" + search + "%"
		args = []interface{}{searchParam, limit, offset}
	} else {
		countQuery = `
			SELECT COUNT(*) 
			FROM user_password_history uph
			JOIN users_application ua ON uph.user_id = ua.user_apps_id
		`

		query = `
			SELECT uph.history_id, uph.user_id, ua.username, uph.password_hash, uph.created_at
			FROM user_password_history uph
			JOIN users_application ua ON uph.user_id = ua.user_apps_id
			ORDER BY uph.created_at DESC
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
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch password history"})
	}
	defer rows.Close()

	var histories []UserPasswordHistory
	for rows.Next() {
		var history UserPasswordHistory
		err := rows.Scan(
			&history.HistoryID,
			&history.UserID,
			&history.Username,
			&history.PasswordHash,
			&history.CreatedAt,
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to scan password history"})
		}
		histories = append(histories, history)
	}

	totalPages := (totalRecords + limit - 1) / limit

	response := map[string]interface{}{
		"data": histories,
		"pagination": map[string]interface{}{
			"current_page":     page,
			"total_pages":      totalPages,
			"total_records":    totalRecords,
			"records_per_page": limit,
		},
	}

	return ctx.JSON(http.StatusOK, response)
}

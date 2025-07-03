package controller

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

type PasswordResetTokensController struct {
	DB *sql.DB
}

type PasswordResetToken struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id"`
	Username  string `json:"username"`
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
	IsUsed    bool   `json:"is_used"`
	CreatedAt string `json:"created_at"`
}

func NewPasswordResetTokensController(db *sql.DB) *PasswordResetTokensController {
	return &PasswordResetTokensController{DB: db}
}

func (c *PasswordResetTokensController) GetAllPasswordResetTokens(ctx echo.Context) error {
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
			FROM password_reset_tokens prt
			JOIN users_application ua ON prt.user_id = ua.user_apps_id
			WHERE ua.username ILIKE $1
		`

		query = `
			SELECT prt.id, prt.user_id, ua.username, prt.token, prt.expires_at, prt.is_used, prt.created_at
			FROM password_reset_tokens prt
			JOIN users_application ua ON prt.user_id = ua.user_apps_id
			WHERE ua.username ILIKE $1
			ORDER BY prt.created_at DESC
			LIMIT $2 OFFSET $3
		`
		searchParam := "%" + search + "%"
		args = []interface{}{searchParam, limit, offset}
	} else {
		countQuery = `
			SELECT COUNT(*) 
			FROM password_reset_tokens prt
			JOIN users_application ua ON prt.user_id = ua.user_apps_id
		`

		query = `
			SELECT prt.id, prt.user_id, ua.username, prt.token, prt.expires_at, prt.is_used, prt.created_at
			FROM password_reset_tokens prt
			JOIN users_application ua ON prt.user_id = ua.user_apps_id
			ORDER BY prt.created_at DESC
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
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch password reset tokens"})
	}
	defer rows.Close()

	var tokens []PasswordResetToken
	for rows.Next() {
		var token PasswordResetToken
		err := rows.Scan(
			&token.ID,
			&token.UserID,
			&token.Username,
			&token.Token,
			&token.ExpiresAt,
			&token.IsUsed,
			&token.CreatedAt,
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to scan password reset token"})
		}
		tokens = append(tokens, token)
	}

	totalPages := (totalRecords + limit - 1) / limit

	response := map[string]interface{}{
		"data": tokens,
		"pagination": map[string]interface{}{
			"current_page":     page,
			"total_pages":      totalPages,
			"total_records":    totalRecords,
			"records_per_page": limit,
		},
	}

	return ctx.JSON(http.StatusOK, response)
}

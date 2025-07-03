package controller

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
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

func (c *UserPasswordHistoryController) GetAllUserPasswordHistory(ctx *gin.Context) {
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
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch password history"})
		return
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
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan password history"})
			return
		}
		histories = append(histories, history)
	}

	totalPages := (totalRecords + limit - 1) / limit

	response := gin.H{
		"data": histories,
		"pagination": gin.H{
			"current_page":     page,
			"total_pages":      totalPages,
			"total_records":    totalRecords,
			"records_per_page": limit,
		},
	}

	ctx.JSON(http.StatusOK, response)
}

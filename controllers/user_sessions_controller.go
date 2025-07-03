package controller

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type UserSessionsController struct {
	DB *sql.DB
}

type UserSession struct {
	SessionID    int     `json:"session_id"`
	UserID       int     `json:"user_id"`
	Username     string  `json:"username"`
	SessionToken string  `json:"session_token"`
	IPAddress    string  `json:"ip_address"`
	UserAgent    string  `json:"user_agent"`
	LoginAt      string  `json:"login_at"`
	LogoutAt     *string `json:"logout_at"`
	ExpiresAt    string  `json:"expires_at"`
	IsActive     bool    `json:"is_active"`
	CreatedAt    string  `json:"created_at"`
}

func NewUserSessionsController(db *sql.DB) *UserSessionsController {
	return &UserSessionsController{DB: db}
}

func (c *UserSessionsController) GetAllUserSessions(ctx *gin.Context) {
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
			FROM user_sessions us
			JOIN users_application ua ON us.user_id = ua.user_apps_id
			WHERE ua.username ILIKE $1 OR us.ip_address::text ILIKE $1
		`

		query = `
			SELECT us.session_id, us.user_id, ua.username, us.session_token, 
				   COALESCE(us.ip_address::text, '') as ip_address, 
				   COALESCE(us.user_agent, '') as user_agent,
				   us.login_at, us.logout_at, us.expires_at, us.is_active, us.created_at
			FROM user_sessions us
			JOIN users_application ua ON us.user_id = ua.user_apps_id
			WHERE ua.username ILIKE $1 OR us.ip_address::text ILIKE $1
			ORDER BY us.created_at DESC
			LIMIT $2 OFFSET $3
		`
		searchParam := "%" + search + "%"
		args = []interface{}{searchParam, limit, offset}
	} else {
		countQuery = `
			SELECT COUNT(*) 
			FROM user_sessions us
			JOIN users_application ua ON us.user_id = ua.user_apps_id
		`

		query = `
			SELECT us.session_id, us.user_id, ua.username, us.session_token, 
				   COALESCE(us.ip_address::text, '') as ip_address, 
				   COALESCE(us.user_agent, '') as user_agent,
				   us.login_at, us.logout_at, us.expires_at, us.is_active, us.created_at
			FROM user_sessions us
			JOIN users_application ua ON us.user_id = ua.user_apps_id
			ORDER BY us.created_at DESC
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
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user sessions"})
		return
	}
	defer rows.Close()

	var sessions []UserSession
	for rows.Next() {
		var session UserSession
		err := rows.Scan(
			&session.SessionID,
			&session.UserID,
			&session.Username,
			&session.SessionToken,
			&session.IPAddress,
			&session.UserAgent,
			&session.LoginAt,
			&session.LogoutAt,
			&session.ExpiresAt,
			&session.IsActive,
			&session.CreatedAt,
		)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan user session"})
			return
		}
		sessions = append(sessions, session)
	}

	totalPages := (totalRecords + limit - 1) / limit

	response := gin.H{
		"data": sessions,
		"pagination": gin.H{
			"current_page":     page,
			"total_pages":      totalPages,
			"total_records":    totalRecords,
			"records_per_page": limit,
		},
	}

	ctx.JSON(http.StatusOK, response)
}

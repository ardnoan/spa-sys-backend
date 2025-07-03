package controller

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type UsersActivityLogsController struct {
	DB *sql.DB
}

type UsersActivityLog struct {
	LogsID         int             `json:"logs_id"`
	UserID         *int            `json:"user_id"`
	Username       *string         `json:"username"`
	SessionID      *int            `json:"session_id"`
	Action         string          `json:"action"`
	TargetType     *string         `json:"target_type"`
	TargetID       *int            `json:"target_id"`
	MenuName       *string         `json:"menu_name"`
	Description    *string         `json:"description"`
	IPAddress      *string         `json:"ip_address"`
	UserAgent      *string         `json:"user_agent"`
	RequestData    json.RawMessage `json:"request_data"`
	ResponseStatus *int            `json:"response_status"`
	CreatedAt      string          `json:"created_at"`
}

func NewUsersActivityLogsController(db *sql.DB) *UsersActivityLogsController {
	return &UsersActivityLogsController{DB: db}
}

func (c *UsersActivityLogsController) GetAllUsersActivityLogs(ctx *gin.Context) {
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
			FROM users_activity_logs ual
			LEFT JOIN users_application ua ON ual.user_id = ua.user_apps_id
			WHERE ua.username ILIKE $1 OR ual.action ILIKE $1 OR ual.menu_name ILIKE $1
		`

		query = `
			SELECT ual.logs_id, ual.user_id, ua.username, ual.session_id, ual.action, 
				   ual.target_type, ual.target_id, ual.menu_name, ual.description,
				   COALESCE(ual.ip_address::text, '') as ip_address, 
				   ual.user_agent, ual.request_data, ual.response_status, ual.created_at
			FROM users_activity_logs ual
			LEFT JOIN users_application ua ON ual.user_id = ua.user_apps_id
			WHERE ua.username ILIKE $1 OR ual.action ILIKE $1 OR ual.menu_name ILIKE $1
			ORDER BY ual.created_at DESC
			LIMIT $2 OFFSET $3
		`
		searchParam := "%" + search + "%"
		args = []interface{}{searchParam, limit, offset}
	} else {
		countQuery = `
			SELECT COUNT(*) 
			FROM users_activity_logs ual
			LEFT JOIN users_application ua ON ual.user_id = ua.user_apps_id
		`

		query = `
			SELECT ual.logs_id, ual.user_id, ua.username, ual.session_id, ual.action, 
				   ual.target_type, ual.target_id, ual.menu_name, ual.description,
				   COALESCE(ual.ip_address::text, '') as ip_address, 
				   ual.user_agent, ual.request_data, ual.response_status, ual.created_at
			FROM users_activity_logs ual
			LEFT JOIN users_application ua ON ual.user_id = ua.user_apps_id
			ORDER BY ual.created_at DESC
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
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch activity logs"})
		return
	}
	defer rows.Close()

	var logs []UsersActivityLog
	for rows.Next() {
		var log UsersActivityLog
		err := rows.Scan(
			&log.LogsID,
			&log.UserID,
			&log.Username,
			&log.SessionID,
			&log.Action,
			&log.TargetType,
			&log.TargetID,
			&log.MenuName,
			&log.Description,
			&log.IPAddress,
			&log.UserAgent,
			&log.RequestData,
			&log.ResponseStatus,
			&log.CreatedAt,
		)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan activity log"})
			return
		}
		logs = append(logs, log)
	}

	totalPages := (totalRecords + limit - 1) / limit

	response := gin.H{
		"data": logs,
		"pagination": gin.H{
			"current_page":     page,
			"total_pages":      totalPages,
			"total_records":    totalRecords,
			"records_per_page": limit,
		},
	}

	ctx.JSON(http.StatusOK, response)
}

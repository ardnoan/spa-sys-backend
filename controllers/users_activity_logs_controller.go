package controller

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

type UsersActivityLogsController struct {
	DB *sql.DB
}

type UsersActivityLog struct {
	LogsID         int              `json:"logs_id"`
	UserID         *int             `json:"user_id"`
	Username       *string          `json:"username"`
	SessionID      *int             `json:"session_id"`
	Action         string           `json:"action"`
	TargetType     *string          `json:"target_type"`
	TargetID       *int             `json:"target_id"`
	MenuName       *string          `json:"menu_name"`
	Description    *string          `json:"description"`
	IPAddress      *string          `json:"ip_address"`
	UserAgent      *string          `json:"user_agent"`
	RequestData    *json.RawMessage `json:"request_data"` // tetap ini
	ResponseStatus *int             `json:"response_status"`
	CreatedAt      time.Time        `json:"created_at"`
}

func NewUsersActivityLogsController(db *sql.DB) *UsersActivityLogsController {
	return &UsersActivityLogsController{DB: db}
}

func (c *UsersActivityLogsController) GetAllUsersActivityLogs(ctx echo.Context) error {
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
		searchParam := "%" + search + "%"
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

	// Count total records
	var totalRecords int
	countArgs := args
	if search == "" {
		countArgs = []interface{}{}
	}
	err := c.DB.QueryRow(countQuery, countArgs...).Scan(&totalRecords)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":  "Failed to count records",
			"detail": err.Error(),
		})
	}

	rows, err := c.DB.Query(query, args...)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":  "Failed to fetch activity logs",
			"detail": err.Error(),
		})
	}
	defer rows.Close()

	var logs []UsersActivityLog
	for rows.Next() {
		var log UsersActivityLog
		var requestData sql.NullString

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
			&requestData, // sementara ke sql.NullString
			&log.ResponseStatus,
			&log.CreatedAt,
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
				"error":  "Failed to scan activity log",
				"detail": err.Error(),
			})
		}

		if requestData.Valid {
			raw := json.RawMessage(requestData.String)
			log.RequestData = &raw
		} else {
			log.RequestData = nil
		}

		logs = append(logs, log)
	}

	totalPages := (totalRecords + limit - 1) / limit

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"data": logs,
		"pagination": map[string]interface{}{
			"current_page":     page,
			"total_pages":      totalPages,
			"total_records":    totalRecords,
			"records_per_page": limit,
		},
	})
}

package controller

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type NotificationsController struct {
	DB *sql.DB
}

type Notification struct {
	NotificationID   int             `json:"notification_id"`
	UserID           int             `json:"user_id"`
	Username         string          `json:"username"`
	NotificationType string          `json:"notification_type"`
	Title            string          `json:"title"`
	Message          string          `json:"message"`
	Data             json.RawMessage `json:"data"`
	IsRead           bool            `json:"is_read"`
	ReadAt           *string         `json:"read_at"`
	ExpiresAt        *string         `json:"expires_at"`
	CreatedAt        string          `json:"created_at"`
}

func NewNotificationsController(db *sql.DB) *NotificationsController {
	return &NotificationsController{DB: db}
}

func (c *NotificationsController) GetAllNotifications(ctx *gin.Context) {
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
			FROM notifications n
			JOIN users_application ua ON n.user_id = ua.user_apps_id
			WHERE ua.username ILIKE $1 OR n.title ILIKE $1 OR n.notification_type ILIKE $1
		`

		query = `
			SELECT n.notification_id, n.user_id, ua.username, n.notification_type, n.title, 
				   n.message, n.data, n.is_read, n.read_at, n.expires_at, n.created_at
			FROM notifications n
			JOIN users_application ua ON n.user_id = ua.user_apps_id
			WHERE ua.username ILIKE $1 OR n.title ILIKE $1 OR n.notification_type ILIKE $1
			ORDER BY n.created_at DESC
			LIMIT $2 OFFSET $3
		`
		searchParam := "%" + search + "%"
		args = []interface{}{searchParam, limit, offset}
	} else {
		countQuery = `
			SELECT COUNT(*) 
			FROM notifications n
			JOIN users_application ua ON n.user_id = ua.user_apps_id
		`

		query = `
			SELECT n.notification_id, n.user_id, ua.username, n.notification_type, n.title, 
				   n.message, n.data, n.is_read, n.read_at, n.expires_at, n.created_at
			FROM notifications n
			JOIN users_application ua ON n.user_id = ua.user_apps_id
			ORDER BY n.created_at DESC
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
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notifications"})
		return
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		var notification Notification
		err := rows.Scan(
			&notification.NotificationID,
			&notification.UserID,
			&notification.Username,
			&notification.NotificationType,
			&notification.Title,
			&notification.Message,
			&notification.Data,
			&notification.IsRead,
			&notification.ReadAt,
			&notification.ExpiresAt,
			&notification.CreatedAt,
		)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan notification"})
			return
		}
		notifications = append(notifications, notification)
	}

	totalPages := (totalRecords + limit - 1) / limit

	response := gin.H{
		"data": notifications,
		"pagination": gin.H{
			"current_page":     page,
			"total_pages":      totalPages,
			"total_records":    totalRecords,
			"records_per_page": limit,
		},
	}

	ctx.JSON(http.StatusOK, response)
}

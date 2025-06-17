package controller

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type Status struct {
	StatusID    int     `json:"users_application_status_id"`
	StatusCode  string  `json:"status_code"`
	StatusName  string  `json:"status_name"`
	Description *string `json:"description"`
	IsActive    bool    `json:"is_active"`
	CreatedAt   string  `json:"created_at"`
	CreatedBy   *string `json:"created_by"`
	UpdatedAt   string  `json:"updated_at"`
	UpdatedBy   *string `json:"updated_by"`
}

type CreateStatusRequest struct {
	StatusCode  string  `json:"status_code" validate:"required"`
	StatusName  string  `json:"status_name" validate:"required"`
	Description *string `json:"description"`
	IsActive    bool    `json:"is_active"`
}

type UpdateStatusRequest struct {
	StatusCode  string  `json:"status_code" validate:"required"`
	StatusName  string  `json:"status_name" validate:"required"`
	Description *string `json:"description"`
	IsActive    bool    `json:"is_active"`
}

type StatusController struct {
	DB *sql.DB
}

func NewStatusController(db *sql.DB) *StatusController {
	return &StatusController{DB: db}
}

// Response helpers
func (sc *StatusController) successResponse(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

func (sc *StatusController) errorResponse(c echo.Context, code int, message string) error {
	return c.JSON(code, map[string]interface{}{
		"success": false,
		"message": message,
	})
}

// Create Status
func (sc *StatusController) CreateStatus(c echo.Context) error {
	var req CreateStatusRequest
	if err := c.Bind(&req); err != nil {
		return sc.errorResponse(c, http.StatusBadRequest, "Invalid request")
	}

	// Check if status code already exists
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM users_application_status WHERE status_code = $1)`
	err := sc.DB.QueryRow(checkQuery, req.StatusCode).Scan(&exists)
	if err != nil {
		return sc.errorResponse(c, http.StatusInternalServerError, "Failed to check status code")
	}
	if exists {
		return sc.errorResponse(c, http.StatusConflict, "Status code already exists")
	}

	// Insert to database
	query := `INSERT INTO users_application_status 
              (status_code, status_name, description, is_active, created_by) 
              VALUES ($1, $2, $3, $4, $5) 
              RETURNING users_application_status_id`

	var statusID int
	err = sc.DB.QueryRow(query,
		req.StatusCode, req.StatusName, req.Description, req.IsActive, "system").Scan(&statusID)

	if err != nil {
		return sc.errorResponse(c, http.StatusInternalServerError, "Failed to create status: "+err.Error())
	}

	return sc.successResponse(c, map[string]interface{}{
		"status_id": statusID,
		"message":   "Status created successfully",
	})
}

// Get Status by ID
func (sc *StatusController) GetStatus(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return sc.errorResponse(c, http.StatusBadRequest, "Invalid status ID")
	}

	var status Status
	query := `SELECT users_application_status_id, status_code, status_name, description, 
              is_active, created_at, created_by, updated_at, updated_by
              FROM users_application_status WHERE users_application_status_id = $1`

	err = sc.DB.QueryRow(query, id).Scan(
		&status.StatusID, &status.StatusCode, &status.StatusName,
		&status.Description, &status.IsActive, &status.CreatedAt,
		&status.CreatedBy, &status.UpdatedAt, &status.UpdatedBy)

	if err != nil {
		if err == sql.ErrNoRows {
			return sc.errorResponse(c, http.StatusNotFound, "Status not found")
		}
		return sc.errorResponse(c, http.StatusInternalServerError, "Failed to fetch status")
	}

	return sc.successResponse(c, status)
}

// Get All Statuses with pagination and filters
func (sc *StatusController) GetAllStatuses(c echo.Context) error {
	// Get query parameters
	page := c.QueryParam("page")
	limit := c.QueryParam("limit")
	search := c.QueryParam("search")
	isActive := c.QueryParam("is_active")

	// Set default values
	pageInt := 1
	limitInt := 10

	if page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			pageInt = p
		}
	}

	if limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 && l <= 100 {
			limitInt = l
		}
	}

	offset := (pageInt - 1) * limitInt

	// Build WHERE clause
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	if search != "" {
		whereClause += " AND (LOWER(status_name) LIKE LOWER($" + strconv.Itoa(argIndex) +
			") OR LOWER(status_code) LIKE LOWER($" + strconv.Itoa(argIndex) +
			") OR LOWER(description) LIKE LOWER($" + strconv.Itoa(argIndex) + "))"
		args = append(args, "%"+search+"%")
		argIndex++
	}

	if isActive != "" {
		if isActive == "true" {
			whereClause += " AND is_active = $" + strconv.Itoa(argIndex)
			args = append(args, true)
		} else if isActive == "false" {
			whereClause += " AND is_active = $" + strconv.Itoa(argIndex)
			args = append(args, false)
		}
		argIndex++
	}

	// Get total count
	var totalCount int
	countQuery := "SELECT COUNT(*) FROM users_application_status " + whereClause
	err := sc.DB.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return sc.errorResponse(c, http.StatusInternalServerError, "Failed to count statuses")
	}

	// Get statuses with pagination
	query := `SELECT users_application_status_id, status_code, status_name, description, 
              is_active, created_at, created_by, updated_at, updated_by
              FROM users_application_status ` + whereClause + ` 
              ORDER BY created_at DESC 
              LIMIT $` + strconv.Itoa(argIndex) + ` OFFSET $` + strconv.Itoa(argIndex+1)

	args = append(args, limitInt, offset)
	rows, err := sc.DB.Query(query, args...)
	if err != nil {
		return sc.errorResponse(c, http.StatusInternalServerError, "Failed to fetch statuses")
	}
	defer rows.Close()

	var statuses []Status
	for rows.Next() {
		var status Status
		err := rows.Scan(&status.StatusID, &status.StatusCode, &status.StatusName,
			&status.Description, &status.IsActive, &status.CreatedAt,
			&status.CreatedBy, &status.UpdatedAt, &status.UpdatedBy)
		if err != nil {
			continue
		}
		statuses = append(statuses, status)
	}

	// Calculate pagination info
	totalPages := (totalCount + limitInt - 1) / limitInt

	return sc.successResponse(c, map[string]interface{}{
		"statuses": statuses,
		"pagination": map[string]interface{}{
			"current_page": pageInt,
			"per_page":     limitInt,
			"total_count":  totalCount,
			"total_pages":  totalPages,
		},
	})
}

// Update Status
func (sc *StatusController) UpdateStatus(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return sc.errorResponse(c, http.StatusBadRequest, "Invalid status ID")
	}

	var req UpdateStatusRequest
	if err := c.Bind(&req); err != nil {
		return sc.errorResponse(c, http.StatusBadRequest, "Invalid request")
	}

	// Check if status exists
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM users_application_status WHERE users_application_status_id = $1)`
	err = sc.DB.QueryRow(checkQuery, id).Scan(&exists)
	if err != nil || !exists {
		return sc.errorResponse(c, http.StatusNotFound, "Status not found")
	}

	// Check if status code already exists (excluding current status)
	var codeExists bool
	codeCheckQuery := `SELECT EXISTS(SELECT 1 FROM users_application_status WHERE status_code = $1 AND users_application_status_id != $2)`
	err = sc.DB.QueryRow(codeCheckQuery, req.StatusCode, id).Scan(&codeExists)
	if err != nil {
		return sc.errorResponse(c, http.StatusInternalServerError, "Failed to check status code")
	}
	if codeExists {
		return sc.errorResponse(c, http.StatusConflict, "Status code already exists")
	}

	query := `UPDATE users_application_status 
              SET status_code = $1, status_name = $2, description = $3, 
                  is_active = $4, updated_by = $5, updated_at = CURRENT_TIMESTAMP
              WHERE users_application_status_id = $6`

	_, err = sc.DB.Exec(query, req.StatusCode, req.StatusName, req.Description,
		req.IsActive, "system", id)

	if err != nil {
		return sc.errorResponse(c, http.StatusInternalServerError, "Failed to update status: "+err.Error())
	}

	return sc.successResponse(c, map[string]string{"message": "Status updated successfully"})
}

// Delete Status (Soft delete)
func (sc *StatusController) DeleteStatus(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return sc.errorResponse(c, http.StatusBadRequest, "Invalid status ID")
	}

	// Check if status exists
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM users_application_status WHERE users_application_status_id = $1)`
	err = sc.DB.QueryRow(checkQuery, id).Scan(&exists)
	if err != nil || !exists {
		return sc.errorResponse(c, http.StatusNotFound, "Status not found")
	}

	// Check if status is being used by users
	var inUse bool
	useCheckQuery := `SELECT EXISTS(SELECT 1 FROM users_application WHERE status_id = $1)`
	err = sc.DB.QueryRow(useCheckQuery, id).Scan(&inUse)
	if err != nil {
		return sc.errorResponse(c, http.StatusInternalServerError, "Failed to check status usage")
	}
	if inUse {
		return sc.errorResponse(c, http.StatusConflict, "Cannot delete status that is being used by users")
	}

	query := `UPDATE users_application_status 
              SET is_active = false, updated_by = $1, updated_at = CURRENT_TIMESTAMP 
              WHERE users_application_status_id = $2`

	_, err = sc.DB.Exec(query, "system", id)
	if err != nil {
		return sc.errorResponse(c, http.StatusInternalServerError, "Failed to delete status")
	}

	return sc.successResponse(c, map[string]string{"message": "Status deleted successfully"})
}

// Search Statuses
func (sc *StatusController) SearchStatuses(c echo.Context) error {
	query := c.QueryParam("q")
	if query == "" {
		return sc.errorResponse(c, http.StatusBadRequest, "Search query is required")
	}

	searchQuery := `SELECT users_application_status_id, status_code, status_name, description, 
                    is_active, created_at, created_by, updated_at, updated_by
                    FROM users_application_status 
                    WHERE is_active = true AND (
                        LOWER(status_name) LIKE LOWER($1) OR 
                        LOWER(status_code) LIKE LOWER($1) OR 
                        LOWER(description) LIKE LOWER($1)
                    )
                    ORDER BY status_name 
                    LIMIT 50`

	searchPattern := "%" + query + "%"
	rows, err := sc.DB.Query(searchQuery, searchPattern)
	if err != nil {
		return sc.errorResponse(c, http.StatusInternalServerError, "Failed to search statuses")
	}
	defer rows.Close()

	var statuses []Status
	for rows.Next() {
		var status Status
		err := rows.Scan(&status.StatusID, &status.StatusCode, &status.StatusName,
			&status.Description, &status.IsActive, &status.CreatedAt,
			&status.CreatedBy, &status.UpdatedAt, &status.UpdatedBy)
		if err != nil {
			continue
		}
		statuses = append(statuses, status)
	}

	return sc.successResponse(c, statuses)
}

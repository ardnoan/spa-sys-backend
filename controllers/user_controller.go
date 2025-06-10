package controller

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           int     `json:"id"`
	Username     string  `json:"username"`
	Email        string  `json:"email"`
	FirstName    string  `json:"first_name"`
	LastName     string  `json:"last_name"`
	StatusID     int     `json:"status_id"`
	DepartmentID *int    `json:"department_id"`
	EmployeeID   *string `json:"employee_id"`
	Phone        *string `json:"phone"`
	IsActive     bool    `json:"is_active"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

type CreateUserRequest struct {
	Username     string  `json:"username" validate:"required"`
	Email        string  `json:"email" validate:"required,email"`
	Password     string  `json:"password" validate:"required,min=6"`
	FirstName    string  `json:"first_name" validate:"required"`
	LastName     string  `json:"last_name" validate:"required"`
	StatusID     int     `json:"status_id" validate:"required"`
	DepartmentID *int    `json:"department_id"`
	EmployeeID   *string `json:"employee_id"`
	Phone        *string `json:"phone"`
}

type UpdateUserRequest struct {
	FirstName    string  `json:"first_name" validate:"required"`
	LastName     string  `json:"last_name" validate:"required"`
	Email        string  `json:"email" validate:"required,email"`
	StatusID     int     `json:"status_id" validate:"required"`
	DepartmentID *int    `json:"department_id"`
	EmployeeID   *string `json:"employee_id"`
	Phone        *string `json:"phone"`
	IsActive     bool    `json:"is_active"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type UserController struct {
	DB *sql.DB
}

func NewUserController(db *sql.DB) *UserController {
	return &UserController{DB: db}
}

// Response helpers
func (uc *UserController) successResponse(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

func (uc *UserController) errorResponse(c echo.Context, code int, message string) error {
	return c.JSON(code, map[string]interface{}{
		"success": false,
		"message": message,
	})
}

// Create User
func (uc *UserController) CreateUser(c echo.Context) error {
	var req CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return uc.errorResponse(c, http.StatusBadRequest, "Invalid request")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return uc.errorResponse(c, http.StatusInternalServerError, "Failed to hash password")
	}

	// Insert to database
	query := `INSERT INTO users_application 
              (username, email, password_hash, first_name, last_name, status_id, department_id, employee_id, phone, created_by) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) 
              RETURNING user_apps_id`

	var userID int
	err = uc.DB.QueryRow(query,
		req.Username, req.Email, string(hashedPassword),
		req.FirstName, req.LastName, req.StatusID, req.DepartmentID,
		req.EmployeeID, req.Phone, "system").Scan(&userID)

	if err != nil {
		return uc.errorResponse(c, http.StatusInternalServerError, "Failed to create user: "+err.Error())
	}

	return uc.successResponse(c, map[string]interface{}{
		"id":      userID,
		"message": "User created successfully",
	})
}

// Get User by ID
func (uc *UserController) GetUser(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return uc.errorResponse(c, http.StatusBadRequest, "Invalid user ID")
	}

	var user User
	query := `SELECT user_apps_id, username, email, first_name, last_name, status_id, 
              department_id, employee_id, phone, is_active, created_at, updated_at
              FROM users_application WHERE user_apps_id = $1`

	err = uc.DB.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.FirstName,
		&user.LastName, &user.StatusID, &user.DepartmentID, &user.EmployeeID,
		&user.Phone, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return uc.errorResponse(c, http.StatusNotFound, "User not found")
		}
		return uc.errorResponse(c, http.StatusInternalServerError, "Failed to fetch user")
	}

	return uc.successResponse(c, user)
}

// Get All Users with pagination
func (uc *UserController) GetAllUsers(c echo.Context) error {
	// Get query parameters for pagination
	page := c.QueryParam("page")
	limit := c.QueryParam("limit")

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

	// Get total count
	var totalCount int
	countQuery := `SELECT COUNT(*) FROM users_application WHERE is_active = true`
	err := uc.DB.QueryRow(countQuery).Scan(&totalCount)
	if err != nil {
		return uc.errorResponse(c, http.StatusInternalServerError, "Failed to count users")
	}

	// Get users with pagination
	query := `SELECT user_apps_id, username, email, first_name, last_name, status_id, 
              department_id, employee_id, phone, is_active, created_at, updated_at
              FROM users_application 
              WHERE is_active = true 
              ORDER BY created_at DESC 
              LIMIT $1 OFFSET $2`

	rows, err := uc.DB.Query(query, limitInt, offset)
	if err != nil {
		return uc.errorResponse(c, http.StatusInternalServerError, "Failed to fetch users")
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.FirstName,
			&user.LastName, &user.StatusID, &user.DepartmentID, &user.EmployeeID,
			&user.Phone, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			continue
		}
		users = append(users, user)
	}

	// Calculate pagination info
	totalPages := (totalCount + limitInt - 1) / limitInt

	return uc.successResponse(c, map[string]interface{}{
		"users": users,
		"pagination": map[string]interface{}{
			"current_page": pageInt,
			"per_page":     limitInt,
			"total_count":  totalCount,
			"total_pages":  totalPages,
		},
	})
}

// Update User
func (uc *UserController) UpdateUser(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return uc.errorResponse(c, http.StatusBadRequest, "Invalid user ID")
	}

	var req UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return uc.errorResponse(c, http.StatusBadRequest, "Invalid request")
	}

	// Check if user exists
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM users_application WHERE user_apps_id = $1)`
	err = uc.DB.QueryRow(checkQuery, id).Scan(&exists)
	if err != nil || !exists {
		return uc.errorResponse(c, http.StatusNotFound, "User not found")
	}

	query := `UPDATE users_application 
              SET first_name = $1, last_name = $2, email = $3, status_id = $4, 
                  department_id = $5, employee_id = $6, phone = $7, is_active = $8,
                  updated_by = $9, updated_at = CURRENT_TIMESTAMP
              WHERE user_apps_id = $10`

	_, err = uc.DB.Exec(query, req.FirstName, req.LastName, req.Email, req.StatusID,
		req.DepartmentID, req.EmployeeID, req.Phone, req.IsActive, "system", id)

	if err != nil {
		return uc.errorResponse(c, http.StatusInternalServerError, "Failed to update user: "+err.Error())
	}

	return uc.successResponse(c, map[string]string{"message": "User updated successfully"})
}

// Delete User (Soft delete)
func (uc *UserController) DeleteUser(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return uc.errorResponse(c, http.StatusBadRequest, "Invalid user ID")
	}

	// Check if user exists
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM users_application WHERE user_apps_id = $1)`
	err = uc.DB.QueryRow(checkQuery, id).Scan(&exists)
	if err != nil || !exists {
		return uc.errorResponse(c, http.StatusNotFound, "User not found")
	}

	query := `UPDATE users_application 
              SET is_active = false, updated_by = $1, updated_at = CURRENT_TIMESTAMP 
              WHERE user_apps_id = $2`

	_, err = uc.DB.Exec(query, "system", id)
	if err != nil {
		return uc.errorResponse(c, http.StatusInternalServerError, "Failed to delete user")
	}

	return uc.successResponse(c, map[string]string{"message": "User deleted successfully"})
}

// Login
func (uc *UserController) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return uc.errorResponse(c, http.StatusBadRequest, "Invalid request")
	}

	var user User
	var passwordHash string
	query := `SELECT user_apps_id, username, email, first_name, last_name, password_hash, status_id
              FROM users_application 
              WHERE username = $1 AND is_active = true`

	err := uc.DB.QueryRow(query, req.Username).Scan(
		&user.ID, &user.Username, &user.Email, &user.FirstName,
		&user.LastName, &passwordHash, &user.StatusID)

	if err != nil {
		if err == sql.ErrNoRows {
			return uc.errorResponse(c, http.StatusUnauthorized, "Invalid credentials")
		}
		return uc.errorResponse(c, http.StatusInternalServerError, "Login failed")
	}

	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password))
	if err != nil {
		return uc.errorResponse(c, http.StatusUnauthorized, "Invalid credentials")
	}

	// Update last login
	updateQuery := `UPDATE users_application SET last_login_at = CURRENT_TIMESTAMP WHERE user_apps_id = $1`
	uc.DB.Exec(updateQuery, user.ID)

	return uc.successResponse(c, map[string]interface{}{
		"user":    user,
		"message": "Login successful",
	})
}

// Get Users by Status
func (uc *UserController) GetUsersByStatus(c echo.Context) error {
	statusID, err := strconv.Atoi(c.Param("status_id"))
	if err != nil {
		return uc.errorResponse(c, http.StatusBadRequest, "Invalid status ID")
	}

	query := `SELECT user_apps_id, username, email, first_name, last_name, status_id, 
              department_id, employee_id, phone, is_active, created_at, updated_at
              FROM users_application 
              WHERE status_id = $1 AND is_active = true 
              ORDER BY created_at DESC`

	rows, err := uc.DB.Query(query, statusID)
	if err != nil {
		return uc.errorResponse(c, http.StatusInternalServerError, "Failed to fetch users")
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.FirstName,
			&user.LastName, &user.StatusID, &user.DepartmentID, &user.EmployeeID,
			&user.Phone, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			continue
		}
		users = append(users, user)
	}

	return uc.successResponse(c, users)
}

// Search Users
func (uc *UserController) SearchUsers(c echo.Context) error {
	query := c.QueryParam("q")
	if query == "" {
		return uc.errorResponse(c, http.StatusBadRequest, "Search query is required")
	}

	searchQuery := `SELECT user_apps_id, username, email, first_name, last_name, status_id, 
                    department_id, employee_id, phone, is_active, created_at, updated_at
                    FROM users_application 
                    WHERE is_active = true AND (
                        LOWER(username) LIKE LOWER($1) OR 
                        LOWER(email) LIKE LOWER($1) OR 
                        LOWER(first_name) LIKE LOWER($1) OR 
                        LOWER(last_name) LIKE LOWER($1) OR
                        LOWER(employee_id) LIKE LOWER($1)
                    )
                    ORDER BY created_at DESC 
                    LIMIT 50`

	searchPattern := "%" + query + "%"
	rows, err := uc.DB.Query(searchQuery, searchPattern)
	if err != nil {
		return uc.errorResponse(c, http.StatusInternalServerError, "Failed to search users")
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.FirstName,
			&user.LastName, &user.StatusID, &user.DepartmentID, &user.EmployeeID,
			&user.Phone, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			continue
		}
		users = append(users, user)
	}

	return uc.successResponse(c, users)
}

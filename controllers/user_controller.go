package controller

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
}

type CreateUserRequest struct {
	Username  string `json:"username" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=6"`
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
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
	query := `INSERT INTO users (username, email, password_hash, first_name, last_name) 
              VALUES ($1, $2, $3, $4, $5) RETURNING id`

	var userID int
	err = uc.DB.QueryRow(query, req.Username, req.Email, string(hashedPassword), req.FirstName, req.LastName).Scan(&userID)
	if err != nil {
		return uc.errorResponse(c, http.StatusInternalServerError, "Failed to create user")
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
	query := `SELECT id, username, email, first_name, last_name, is_active, created_at 
              FROM users WHERE id = $1`

	err = uc.DB.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.FirstName,
		&user.LastName, &user.IsActive, &user.CreatedAt)

	if err != nil {
		return uc.errorResponse(c, http.StatusNotFound, "User not found")
	}

	return uc.successResponse(c, user)
}

// Get All Users
func (uc *UserController) GetAllUsers(c echo.Context) error {
	query := `SELECT id, username, email, first_name, last_name, is_active, created_at 
              FROM users WHERE is_active = true ORDER BY created_at DESC`

	rows, err := uc.DB.Query(query)
	if err != nil {
		return uc.errorResponse(c, http.StatusInternalServerError, "Failed to fetch users")
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.FirstName,
			&user.LastName, &user.IsActive, &user.CreatedAt)
		if err != nil {
			continue
		}
		users = append(users, user)
	}

	return uc.successResponse(c, users)
}

// Update User
func (uc *UserController) UpdateUser(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return uc.errorResponse(c, http.StatusBadRequest, "Invalid user ID")
	}

	var req CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return uc.errorResponse(c, http.StatusBadRequest, "Invalid request")
	}

	query := `UPDATE users SET first_name = $1, last_name = $2, email = $3 
              WHERE id = $4`

	_, err = uc.DB.Exec(query, req.FirstName, req.LastName, req.Email, id)
	if err != nil {
		return uc.errorResponse(c, http.StatusInternalServerError, "Failed to update user")
	}

	return uc.successResponse(c, map[string]string{"message": "User updated successfully"})
}

// Delete User
func (uc *UserController) DeleteUser(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return uc.errorResponse(c, http.StatusBadRequest, "Invalid user ID")
	}

	query := `UPDATE users SET is_active = false WHERE id = $1`
	_, err = uc.DB.Exec(query, id)
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
	query := `SELECT id, username, email, first_name, last_name, password_hash 
              FROM users WHERE username = $1 AND is_active = true`

	err := uc.DB.QueryRow(query, req.Username).Scan(
		&user.ID, &user.Username, &user.Email, &user.FirstName,
		&user.LastName, &passwordHash)

	if err != nil {
		return uc.errorResponse(c, http.StatusUnauthorized, "Invalid credentials")
	}

	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password))
	if err != nil {
		return uc.errorResponse(c, http.StatusUnauthorized, "Invalid credentials")
	}

	return uc.successResponse(c, map[string]interface{}{
		"user":    user,
		"message": "Login successful",
	})
}

package controller

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID                  int        `json:"id" db:"user_apps_id"`
	Username            string     `json:"username" db:"username"`
	Email               string     `json:"email" db:"email"`
	PasswordHash        string     `json:"-" db:"password_hash"` // Hidden from JSON
	FirstName           string     `json:"first_name" db:"first_name"`
	LastName            string     `json:"last_name" db:"last_name"`
	StatusID            int        `json:"status_id" db:"status_id"`
	DepartmentID        *int       `json:"department_id" db:"department_id"`
	EmployeeID          *string    `json:"employee_id" db:"employee_id"`
	Phone               *string    `json:"phone" db:"phone"`
	AvatarURL           *string    `json:"avatar_url" db:"avatar_url"`
	LastLoginAt         *time.Time `json:"last_login_at" db:"last_login_at"`
	PasswordChangedAt   time.Time  `json:"password_changed_at" db:"password_changed_at"`
	FailedLoginAttempts int        `json:"failed_login_attempts" db:"failed_login_attempts"`
	LockedUntil         *time.Time `json:"locked_until" db:"locked_until"`
	IsActive            bool       `json:"is_active" db:"is_active"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	CreatedBy           *string    `json:"created_by" db:"created_by"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
	UpdatedBy           *string    `json:"updated_by" db:"updated_by"`
}

type CreateUserRequest struct {
	Username     string  `json:"username" validate:"required,min=3,max=50"`
	Email        string  `json:"email" validate:"required,email,max=100"`
	Password     string  `json:"password" validate:"required,min=6,max=100"`
	FirstName    string  `json:"first_name" validate:"required,min=1,max=50"`
	LastName     string  `json:"last_name" validate:"required,min=1,max=50"`
	StatusID     int     `json:"status_id" validate:"required,min=1"`
	DepartmentID *int    `json:"department_id" validate:"omitempty,min=1"`
	EmployeeID   *string `json:"employee_id" validate:"omitempty,max=50"`
	Phone        *string `json:"phone" validate:"omitempty,max=20"`
}

type UpdateUserRequest struct {
	FirstName    string  `json:"first_name" validate:"required,min=1,max=100"`
	LastName     string  `json:"last_name" validate:"required,min=1,max=100"`
	Email        string  `json:"email" validate:"required,email,max=255"`
	StatusID     int     `json:"status_id" validate:"required,min=1"`
	DepartmentID *int    `json:"department_id"`
	EmployeeID   *string `json:"employee_id" validate:"omitempty,max=50"`
	Phone        *string `json:"phone" validate:"omitempty,max=20"`
	IsActive     bool    `json:"is_active"`
	Password     *string `json:"password,omitempty" validate:"omitempty,min=6"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	User    User   `json:"user"`
	Message string `json:"message"`
}

type UserController struct {
	DB *sql.DB
}

var validate *validator.Validate

func init() {
	validate = validator.New()
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
		return uc.errorResponse(c, http.StatusBadRequest, "Invalid request format")
	}

	// Validate request menggunakan validator langsung
	if err := validate.Struct(&req); err != nil {
		// Format validation errors
		validationErrors := make([]string, 0)
		if validatorErr, ok := err.(validator.ValidationErrors); ok {
			for _, fieldError := range validatorErr {
				switch fieldError.Tag() {
				case "required":
					validationErrors = append(validationErrors, fieldError.Field()+" is required")
				case "email":
					validationErrors = append(validationErrors, "Invalid email format")
				case "min":
					validationErrors = append(validationErrors, fieldError.Field()+" must be at least "+fieldError.Param()+" characters")
				case "max":
					validationErrors = append(validationErrors, fieldError.Field()+" must be at most "+fieldError.Param()+" characters")
				default:
					validationErrors = append(validationErrors, fieldError.Field()+" is invalid")
				}
			}
		}
		return uc.errorResponse(c, http.StatusBadRequest, strings.Join(validationErrors, ", "))
	}

	// Check if username or email already exists
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM users_application WHERE username = $1 OR email = $2)`
	err := uc.DB.QueryRow(checkQuery, req.Username, req.Email).Scan(&exists)
	if err != nil {
		return uc.errorResponse(c, http.StatusInternalServerError, "Database error")
	}
	if exists {
		return uc.errorResponse(c, http.StatusConflict, "Username or email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return uc.errorResponse(c, http.StatusInternalServerError, "Failed to process password")
	}

	// Insert to database
	query := `INSERT INTO users_application 
              (username, email, password_hash, first_name, last_name, status_id, 
               department_id, employee_id, phone, is_active, created_by, password_changed_at) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, true, $10, CURRENT_TIMESTAMP) 
              RETURNING user_apps_id`

	var userID int
	err = uc.DB.QueryRow(query,
		req.Username, req.Email, string(hashedPassword),
		req.FirstName, req.LastName, req.StatusID, req.DepartmentID,
		req.EmployeeID, req.Phone, "system").Scan(&userID)

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
	query := `SELECT user_apps_id, username, email, first_name, last_name, status_id, 
              department_id, employee_id, phone, avatar_url, last_login_at, 
              password_changed_at, failed_login_attempts, locked_until, is_active, 
              created_at, created_by, updated_at, updated_by
              FROM users_application WHERE user_apps_id = $1`

	err = uc.DB.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.FirstName,
		&user.LastName, &user.StatusID, &user.DepartmentID, &user.EmployeeID,
		&user.Phone, &user.AvatarURL, &user.LastLoginAt, &user.PasswordChangedAt,
		&user.FailedLoginAttempts, &user.LockedUntil, &user.IsActive,
		&user.CreatedAt, &user.CreatedBy, &user.UpdatedAt, &user.UpdatedBy)

	if err != nil {
		if err == sql.ErrNoRows {
			return uc.errorResponse(c, http.StatusNotFound, "User not found")
		}
		return uc.errorResponse(c, http.StatusInternalServerError, "Failed to fetch user")
	}

	return uc.successResponse(c, user)
}
func (uc *UserController) GetAllUsers(c echo.Context) error {
	// Get query parameters
	page := c.QueryParam("page")
	limit := c.QueryParam("limit")
	statusID := c.QueryParam("status_id")
	departmentID := c.QueryParam("department_id")
	search := c.QueryParam("search")

	// Set default pagination
	pageInt := 1
	limitInt := 10
	if p, err := strconv.Atoi(page); err == nil && p > 0 {
		pageInt = p
	}
	if l, err := strconv.Atoi(limit); err == nil && l > 0 && l <= 100 {
		limitInt = l
	}
	offset := (pageInt - 1) * limitInt

	// Build dynamic WHERE clause
	whereConditions := []string{}
	args := []interface{}{}
	argIndex := 1

	if statusID != "" {
		if sid, err := strconv.Atoi(statusID); err == nil {
			whereConditions = append(whereConditions, "status_id = $"+strconv.Itoa(argIndex))
			args = append(args, sid)
			argIndex++
		}
	}

	if departmentID != "" {
		if did, err := strconv.Atoi(departmentID); err == nil {
			whereConditions = append(whereConditions, "department_id = $"+strconv.Itoa(argIndex))
			args = append(args, did)
			argIndex++
		}
	}

	if search != "" {
		searchPattern := "%" + strings.ToLower(search) + "%"
		condition := `(LOWER(username) LIKE $` + strconv.Itoa(argIndex) +
			` OR LOWER(email) LIKE $` + strconv.Itoa(argIndex) +
			` OR LOWER(first_name) LIKE $` + strconv.Itoa(argIndex) +
			` OR LOWER(last_name) LIKE $` + strconv.Itoa(argIndex) +
			` OR LOWER(employee_id) LIKE $` + strconv.Itoa(argIndex) + `)`
		whereConditions = append(whereConditions, condition)
		args = append(args, searchPattern)
		argIndex++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Count total users
	var totalCount int
	countQuery := "SELECT COUNT(*) FROM users_application " + whereClause
	if err := uc.DB.QueryRow(countQuery, args...).Scan(&totalCount); err != nil {
		return uc.errorResponse(c, http.StatusInternalServerError, "Failed to count users")
	}

	// Query user data
	query := `SELECT user_apps_id, username, email, first_name, last_name, status_id, 
              department_id, employee_id, phone, avatar_url, last_login_at, 
              password_changed_at, failed_login_attempts, locked_until, is_active, 
              created_at, created_by, updated_at, updated_by
              FROM users_application ` + whereClause + `
              ORDER BY created_at DESC 
              LIMIT $` + strconv.Itoa(argIndex) + ` OFFSET $` + strconv.Itoa(argIndex+1)

	args = append(args, limitInt, offset)

	rows, err := uc.DB.Query(query, args...)
	if err != nil {
		return uc.errorResponse(c, http.StatusInternalServerError, "Failed to fetch users")
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.FirstName,
			&user.LastName, &user.StatusID, &user.DepartmentID, &user.EmployeeID,
			&user.Phone, &user.AvatarURL, &user.LastLoginAt, &user.PasswordChangedAt,
			&user.FailedLoginAttempts, &user.LockedUntil, &user.IsActive,
			&user.CreatedAt, &user.CreatedBy, &user.UpdatedAt, &user.UpdatedBy,
		); err == nil {
			users = append(users, user)
		}
	}

	totalPages := (totalCount + limitInt - 1) / limitInt
	return uc.successResponse(c, map[string]interface{}{
		"users": users,
		"pagination": map[string]interface{}{
			"current_page": pageInt,
			"per_page":     limitInt,
			"total_count":  totalCount,
			"total_pages":  totalPages,
			"has_next":     pageInt < totalPages,
			"has_prev":     pageInt > 1,
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
		return uc.errorResponse(c, http.StatusBadRequest, "Invalid request format")
	}

	// Use custom validation instead of c.Validate
	if err := uc.validateRequest(&req); err != nil {
		// Format validation errors nicely
		validationErrors := make([]string, 0)
		if validatorErr, ok := err.(validator.ValidationErrors); ok {
			for _, fieldError := range validatorErr {
				switch fieldError.Tag() {
				case "required":
					validationErrors = append(validationErrors, fieldError.Field()+" is required")
				case "email":
					validationErrors = append(validationErrors, "Invalid email format")
				case "min":
					if fieldError.Field() == "Password" {
						validationErrors = append(validationErrors, "Password must be at least 6 characters")
					} else {
						validationErrors = append(validationErrors, fieldError.Field()+" is too short")
					}
				case "max":
					validationErrors = append(validationErrors, fieldError.Field()+" is too long")
				default:
					validationErrors = append(validationErrors, fieldError.Field()+" is invalid")
				}
			}
		}
		return uc.errorResponse(c, http.StatusBadRequest, strings.Join(validationErrors, ", "))
	}

	// Check if user exists
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM users_application WHERE user_apps_id = $1)`
	err = uc.DB.QueryRow(checkQuery, id).Scan(&exists)
	if err != nil || !exists {
		return uc.errorResponse(c, http.StatusNotFound, "User not found")
	}

	// Check if email is taken by another user
	checkEmailQuery := `SELECT EXISTS(SELECT 1 FROM users_application WHERE email = $1 AND user_apps_id != $2)`
	err = uc.DB.QueryRow(checkEmailQuery, req.Email, id).Scan(&exists)
	if err != nil {
		return uc.errorResponse(c, http.StatusInternalServerError, "Database error")
	}
	if exists {
		return uc.errorResponse(c, http.StatusConflict, "Email already exists")
	}

	// Prepare the update query
	var query string
	var args []interface{}

	if req.Password != nil && *req.Password != "" {
		// Update with password
		query = `UPDATE users_application 
				SET first_name = $1, last_name = $2, email = $3, status_id = $4,
					department_id = $5, employee_id = $6, phone = $7, is_active = $8,
					password_hash = $9, updated_by = $10, updated_at = CURRENT_TIMESTAMP
				WHERE user_apps_id = $11`

		// Hash the password here
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return uc.errorResponse(c, http.StatusInternalServerError, "Failed to hash password")
		}

		args = []interface{}{
			req.FirstName, req.LastName, req.Email, req.StatusID,
			req.DepartmentID, req.EmployeeID, req.Phone, req.IsActive,
			string(hashedPassword),
			"system", id,
		}
	} else {
		// Update without password
		query = `UPDATE users_application 
				SET first_name = $1, last_name = $2, email = $3, status_id = $4,
					department_id = $5, employee_id = $6, phone = $7, is_active = $8,
					updated_by = $9, updated_at = CURRENT_TIMESTAMP
				WHERE user_apps_id = $10`

		args = []interface{}{
			req.FirstName, req.LastName, req.Email, req.StatusID,
			req.DepartmentID, req.EmployeeID, req.Phone, req.IsActive,
			"system", id,
		}
	}

	_, err = uc.DB.Exec(query, args...)
	if err != nil {
		return uc.errorResponse(c, http.StatusInternalServerError, "Failed to update user")
	}

	return uc.successResponse(c, map[string]string{"message": "User updated successfully"})
}

// validateRequest validates a struct using the validator package
func (uc *UserController) validateRequest(i interface{}) error {
	return validate.Struct(i)
}

// Delete User (Soft delete)
func (uc *UserController) DeleteUser(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return uc.errorResponse(c, http.StatusBadRequest, "Invalid user ID")
	}

	// Check if user exists and is active
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM users_application WHERE user_apps_id = $1 AND is_active = true)`
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

// Login with enhanced security
func (uc *UserController) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return uc.errorResponse(c, http.StatusBadRequest, "Invalid request format")
	}

	// Validate request
	if err := c.Validate(&req); err != nil {
		return uc.errorResponse(c, http.StatusBadRequest, err.Error())
	}

	var user User
	var passwordHash string
	query := `SELECT user_apps_id, username, email, first_name, last_name, password_hash, 
              status_id, department_id, employee_id, phone, avatar_url, 
              failed_login_attempts, locked_until, is_active
              FROM users_application 
              WHERE username = $1 AND is_active = true`

	err := uc.DB.QueryRow(query, req.Username).Scan(
		&user.ID, &user.Username, &user.Email, &user.FirstName,
		&user.LastName, &passwordHash, &user.StatusID, &user.DepartmentID,
		&user.EmployeeID, &user.Phone, &user.AvatarURL,
		&user.FailedLoginAttempts, &user.LockedUntil, &user.IsActive)

	if err != nil {
		if err == sql.ErrNoRows {
			return uc.errorResponse(c, http.StatusUnauthorized, "Invalid credentials")
		}
		return uc.errorResponse(c, http.StatusInternalServerError, "Login failed")
	}

	// Check if account is locked
	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		return uc.errorResponse(c, http.StatusUnauthorized, "Account is temporarily locked")
	}

	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password))
	if err != nil {
		// Increment failed login attempts
		uc.incrementFailedLogins(user.ID)
		return uc.errorResponse(c, http.StatusUnauthorized, "Invalid credentials")
	}

	// Reset failed login attempts and update last login
	updateQuery := `UPDATE users_application 
                    SET last_login_at = CURRENT_TIMESTAMP, 
                        failed_login_attempts = 0, 
                        locked_until = NULL 
                    WHERE user_apps_id = $1`
	uc.DB.Exec(updateQuery, user.ID)

	return uc.successResponse(c, LoginResponse{
		User:    user,
		Message: "Login successful",
	})
}

// Helper function to increment failed login attempts
func (uc *UserController) incrementFailedLogins(userID int) {
	const maxAttempts = 5
	const lockDuration = 30 * time.Minute

	query := `UPDATE users_application 
              SET failed_login_attempts = failed_login_attempts + 1,
                  locked_until = CASE 
                      WHEN failed_login_attempts + 1 >= $1 THEN $2
                      ELSE locked_until
                  END
              WHERE user_apps_id = $3`

	uc.DB.Exec(query, maxAttempts, time.Now().Add(lockDuration), userID)
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

// Search Users (now integrated into GetAllUsers with search parameter)
func (uc *UserController) SearchUsers(c echo.Context) error {
	query := c.QueryParam("q")
	if query == "" {
		return uc.errorResponse(c, http.StatusBadRequest, "Search query is required")
	}

	// Redirect to GetAllUsers with search parameter
	c.Request().URL.RawQuery = "search=" + query + "&limit=50"
	return uc.GetAllUsers(c)
}

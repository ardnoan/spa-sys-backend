package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	db *sql.DB
}

type User struct {
	UserAppsID   int        `json:"user_apps_id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"` // Hidden from JSON
	FirstName    string     `json:"first_name"`
	LastName     string     `json:"last_name"`
	StatusID     int        `json:"status_id"`
	DepartmentID *int       `json:"department_id"`
	EmployeeID   *string    `json:"employee_id"`
	Phone        *string    `json:"phone"`
	AvatarURL    *string    `json:"avatar_url"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	IsActive     bool       `json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type UserStatus struct {
	StatusID   int    `json:"status_id"`
	StatusCode string `json:"status_code"`
	StatusName string `json:"status_name"`
}

type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type LoginRequest struct {
	Username   string `json:"username" validate:"required"`
	Password   string `json:"password" validate:"required"`
	RememberMe bool   `json:"rememberMe"`
}

type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
	User    *User  `json:"user,omitempty"`
}

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func init() {
	if len(jwtSecret) == 0 {
		jwtSecret = []byte("your-secret-key-here") // Change this in production
	}
}

func NewAuthService(db *sql.DB) *AuthService {
	return &AuthService{db: db}
}

func (s *AuthService) Login(req LoginRequest, ipAddress, userAgent string) (*LoginResponse, error) {
	// Get user from database
	user, err := s.getUserByUsername(req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			return &LoginResponse{
				Success: false,
				Message: "Invalid username or password",
			}, nil
		}
		return nil, fmt.Errorf("database error: %v", err)
	}

	// Check if user is active
	if !user.IsActive {
		return &LoginResponse{
			Success: false,
			Message: "Account is not active. Please contact administrator",
		}, nil
	}

	// Check user status
	userStatus, err := s.getUserStatus(user.StatusID)
	if err != nil {
		return nil, fmt.Errorf("failed to check user status: %v", err)
	}

	if userStatus.StatusCode != "ACTIVE" {
		return &LoginResponse{
			Success: false,
			Message: "Account is not active. Please contact administrator",
		}, nil
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		// Record failed login attempt
		s.recordFailedLogin(user.UserAppsID, ipAddress, userAgent)
		return &LoginResponse{
			Success: false,
			Message: "Invalid username or password",
		}, nil
	}

	// Generate JWT token
	expirationTime := time.Now().Add(24 * time.Hour)
	if req.RememberMe {
		expirationTime = time.Now().Add(7 * 24 * time.Hour) // 7 days
	}

	claims := &Claims{
		UserID:   user.UserAppsID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %v", err)
	}

	// Create session record
	sessionID, err := s.createSession(user.UserAppsID, tokenString, ipAddress, userAgent, expirationTime)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %v", err)
	}

	// Update last login
	s.updateLastLogin(user.UserAppsID)

	// Log successful login
	s.logUserActivity(user.UserAppsID, sessionID, "LOGIN", "USER", user.UserAppsID, "Login", "Successful login", ipAddress, userAgent, nil, 200)

	// Remove sensitive data from response
	user.PasswordHash = ""

	return &LoginResponse{
		Success: true,
		Message: "Login successful",
		Token:   tokenString,
		User:    user,
	}, nil
}

func (s *AuthService) Logout(tokenString, ipAddress, userAgent string) error {
	// Parse token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return fmt.Errorf("invalid token: %v", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return fmt.Errorf("invalid token claims")
	}

	// Mark session as inactive
	s.deactivateSession(tokenString)

	// Log logout activity
	s.logUserActivity(claims.UserID, 0, "LOGOUT", "USER", claims.UserID, "Logout", "User logged out", ipAddress, userAgent, nil, 200)

	return nil
}

func (s *AuthService) GetCurrentUser(tokenString string) (*User, error) {
	// Parse and validate token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %v", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Get user data
	user, err := s.getUserByID(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user data: %v", err)
	}

	// Remove sensitive data
	user.PasswordHash = ""

	return user, nil
}

func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %v", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

func (s *AuthService) RefreshToken(tokenString, ipAddress, userAgent string) (string, error) {
	// Parse existing token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return "", fmt.Errorf("invalid token: %v", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}

	// Check if token is expired
	if time.Now().After(claims.ExpiresAt.Time) {
		return "", fmt.Errorf("token has expired")
	}

	// Create new token with extended expiration
	newExpirationTime := time.Now().Add(24 * time.Hour)
	newClaims := &Claims{
		UserID:   claims.UserID,
		Username: claims.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(newExpirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	newTokenString, err := newToken.SignedString(jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to generate new token: %v", err)
	}

	// Deactivate old session
	s.deactivateSession(tokenString)

	// Create new session
	sessionID, err := s.createSession(claims.UserID, newTokenString, ipAddress, userAgent, newExpirationTime)
	if err != nil {
		return "", fmt.Errorf("failed to create new session: %v", err)
	}

	// Log token refresh
	s.logUserActivity(claims.UserID, sessionID, "TOKEN_REFRESH", "USER", claims.UserID, "Token Refresh", "Token refreshed", ipAddress, userAgent, nil, 200)

	return newTokenString, nil
}

// Database helper methods (sama seperti sebelumnya)
func (s *AuthService) getUserByUsername(username string) (*User, error) {
	query := `
        SELECT user_apps_id, username, email, password_hash, first_name, last_name, 
               status_id, department_id, employee_id, phone, avatar_url, last_login_at, 
               is_active, created_at, updated_at
        FROM users_application 
        WHERE username = $1 OR email = $1
    `

	user := &User{}
	row := s.db.QueryRow(query, username)

	err := row.Scan(
		&user.UserAppsID, &user.Username, &user.Email, &user.PasswordHash,
		&user.FirstName, &user.LastName, &user.StatusID, &user.DepartmentID,
		&user.EmployeeID, &user.Phone, &user.AvatarURL, &user.LastLoginAt,
		&user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) getUserByID(userID int) (*User, error) {
	query := `
        SELECT user_apps_id, username, email, first_name, last_name, 
               status_id, department_id, employee_id, phone, avatar_url, last_login_at, 
               is_active, created_at, updated_at
        FROM users_application 
        WHERE user_apps_id = $1
    `

	user := &User{}
	row := s.db.QueryRow(query, userID)

	err := row.Scan(
		&user.UserAppsID, &user.Username, &user.Email,
		&user.FirstName, &user.LastName, &user.StatusID, &user.DepartmentID,
		&user.EmployeeID, &user.Phone, &user.AvatarURL, &user.LastLoginAt,
		&user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) getUserStatus(statusID int) (*UserStatus, error) {
	query := `
        SELECT users_application_status_id, status_code, status_name
        FROM users_application_status 
        WHERE users_application_status_id = $1
    `

	status := &UserStatus{}
	row := s.db.QueryRow(query, statusID)

	err := row.Scan(&status.StatusID, &status.StatusCode, &status.StatusName)
	if err != nil {
		return nil, err
	}

	return status, nil
}

func (s *AuthService) createSession(userID int, token string, ipAddress string, userAgent string, expiresAt time.Time) (int, error) {
	query := `
        INSERT INTO user_sessions (user_id, session_token, ip_address, user_agent, expires_at)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING session_id
    `

	var sessionID int
	err := s.db.QueryRow(query, userID, token, ipAddress, userAgent, expiresAt).Scan(&sessionID)
	return sessionID, err
}

func (s *AuthService) deactivateSession(token string) error {
	query := `
        UPDATE user_sessions 
        SET is_active = false, logout_at = CURRENT_TIMESTAMP
        WHERE session_token = $1
    `

	_, err := s.db.Exec(query, token)
	return err
}

func (s *AuthService) updateLastLogin(userID int) error {
	query := `
        UPDATE users_application 
        SET last_login_at = CURRENT_TIMESTAMP, failed_login_attempts = 0
        WHERE user_apps_id = $1
    `

	_, err := s.db.Exec(query, userID)
	return err
}

func (s *AuthService) recordFailedLogin(userID int, ipAddress string, userAgent string) error {
	query := `
        UPDATE users_application 
        SET failed_login_attempts = failed_login_attempts + 1
        WHERE user_apps_id = $1
    `

	_, err := s.db.Exec(query, userID)
	return err
}

func (s *AuthService) logUserActivity(userID int, sessionID int, action string, targetType string, targetID int, menuName string, description string, ipAddress string, userAgent string, requestData map[string]interface{}, responseStatus int) error {
	query := `
        INSERT INTO users_activity_logs (user_id, session_id, action, target_type, target_id, menu_name, description, ip_address, user_agent, request_data, response_status)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
    `

	var requestDataJSON []byte
	if requestData != nil {
		requestDataJSON, _ = json.Marshal(requestData)
	}

	_, err := s.db.Exec(query, userID, sessionID, action, targetType, targetID, menuName, description, ipAddress, userAgent, requestDataJSON, responseStatus)
	return err
}

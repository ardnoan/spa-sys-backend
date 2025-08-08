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

type LoginRequest struct {
	Username   string `json:"username" validate:"required"`
	Password   string `json:"password" validate:"required"`
	RememberMe bool   `json:"remember_me"`
}

type LoginResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Token   string          `json:"token,omitempty"`
	User    json.RawMessage `json:"user,omitempty"`
}

type SessionValidation struct {
	Valid    bool   `json:"valid"`
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func init() {
	if len(jwtSecret) == 0 {
		jwtSecret = []byte("your-secret-key-change-in-production")
	}
}

func NewAuthService(db *sql.DB) *AuthService {
	return &AuthService{db: db}
}

func (s *AuthService) Login(req LoginRequest, ipAddress, userAgent string) (*LoginResponse, error) {
	// First, get user for password validation (we need to do this in Go for bcrypt)
	user, err := s.getUserForPasswordValidation(req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			// Record failed login attempt
			s.recordFailedLogin(req.Username, ipAddress, userAgent)
			return &LoginResponse{
				Success: false,
				Message: "Invalid credentials",
			}, nil
		}
		return nil, fmt.Errorf("database error: %v", err)
	}

	// Verify password with bcrypt
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		// Record failed login attempt through procedure
		message, err := s.recordFailedLogin(req.Username, ipAddress, userAgent)
		if err != nil {
			return nil, fmt.Errorf("failed to record login attempt: %v", err)
		}
		return &LoginResponse{
			Success: false,
			Message: message,
		}, nil
	}

	// Call stored procedure for login
	result, err := s.callLoginProcedure(req.Username, ipAddress, userAgent)
	if err != nil {
		return nil, fmt.Errorf("login procedure error: %v", err)
	}

	if !result.Success {
		return &LoginResponse{
			Success: false,
			Message: result.Message,
		}, nil
	}

	// Generate JWT token
	expirationTime := time.Now().Add(24 * time.Hour)
	if req.RememberMe {
		expirationTime = time.Now().Add(7 * 24 * time.Hour)
	}

	claims := &Claims{
		UserID:   result.UserID,
		Username: result.Username,
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

	return &LoginResponse{
		Success: true,
		Message: result.Message,
		Token:   tokenString,
		User:    result.UserInfo,
	}, nil
}

func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
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

func (s *AuthService) ValidateSession(sessionToken string) (*SessionValidation, error) {
	query := `SELECT * FROM security.validate_session($1)`

	var result SessionValidation
	row := s.db.QueryRow(query, sessionToken)

	err := row.Scan(&result.Valid, &result.UserID, &result.Username, &result.Message)
	if err != nil {
		return nil, fmt.Errorf("failed to validate session: %v", err)
	}

	return &result, nil
}

func (s *AuthService) Logout(sessionToken, ipAddress, userAgent string) error {
	query := `SELECT security.logout_user($1, $2, $3)`

	var message string
	err := s.db.QueryRow(query, sessionToken, ipAddress, userAgent).Scan(&message)
	if err != nil {
		return fmt.Errorf("logout failed: %v", err)
	}

	if message != "Logout successful" {
		return fmt.Errorf("logout failed: %s", message)
	}

	return nil
}

// Helper methods (database calls)
type userForPasswordValidation struct {
	UserAppsID   int    `json:"user_apps_id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
}

func (s *AuthService) getUserForPasswordValidation(username string) (*userForPasswordValidation, error) {
	query := `
        SELECT user_apps_id, username, password_hash
        FROM security.users_application 
        WHERE username = $1 OR email = $1
    `

	user := &userForPasswordValidation{}
	row := s.db.QueryRow(query, username)

	err := row.Scan(&user.UserAppsID, &user.Username, &user.PasswordHash)
	if err != nil {
		return nil, err
	}

	return user, nil
}

type procedureLoginResult struct {
	Success  bool            `json:"success"`
	Message  string          `json:"message"`
	UserID   int             `json:"user_id"`
	Username string          `json:"username"`
	UserInfo json.RawMessage `json:"user_info"`
}

func (s *AuthService) callLoginProcedure(username, ipAddress, userAgent string) (*procedureLoginResult, error) {
	query := `
		SELECT success, message, user_id, user_info
		FROM security.login_user($1, ''::text, $2::inet, $3)
	`

	var result procedureLoginResult
	err := s.db.QueryRow(query, username, ipAddress, userAgent).Scan(
		&result.Success,
		&result.Message,
		&result.UserID,
		&result.UserInfo,
	)
	if err != nil {
		return nil, err
	}

	// Extract username from JSON user_info kalau memang disimpan di situ
	var userInfoMap map[string]interface{}
	if err := json.Unmarshal(result.UserInfo, &userInfoMap); err == nil {
		if uname, ok := userInfoMap["username"].(string); ok {
			result.Username = uname
		}
	}

	return &result, nil
}

func (s *AuthService) recordFailedLogin(username, ipAddress, userAgent string) (string, error) {
	query := `SELECT security.record_failed_login($1, $2, $3)`

	var message string
	err := s.db.QueryRow(query, username, ipAddress, userAgent).Scan(&message)
	if err != nil {
		return "Database error", err
	}

	return message, nil
}

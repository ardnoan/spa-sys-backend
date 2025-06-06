package models

import "time"

// Login requests and responses
type LoginRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginResponse struct {
	User         UserProfile `json:"user"`
	Token        string      `json:"token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresAt    int64       `json:"expires_at"`
	TokenType    string      `json:"token_type"`
}

// Token requests and responses
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type RefreshTokenResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
	TokenType    string `json:"token_type"`
}

// Password requests
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8,password"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=NewPassword"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Token           string `json:"token" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8,password"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=NewPassword"`
}

// Profile requests and responses
type UserProfile struct {
	UserID      int        `json:"user_id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	FirstName   string     `json:"first_name"`
	LastName    string     `json:"last_name"`
	Phone       *string    `json:"phone"`
	AvatarURL   *string    `json:"avatar_url"`
	Status      string     `json:"status"`
	Department  *string    `json:"department"`
	LastLoginAt *time.Time `json:"last_login_at"`
	Roles       []string   `json:"roles"`
	Permissions []string   `json:"permissions"`
}

type UpdateProfileRequest struct {
	FirstName string `json:"first_name" validate:"required,max=50"`
	LastName  string `json:"last_name" validate:"required,max=50"`
	Phone     string `json:"phone" validate:"omitempty,phone"`
}

// Session management
type UserSession struct {
	SessionID    int        `json:"session_id" db:"session_id"`
	UserID       int        `json:"user_id" db:"user_id"`
	SessionToken string     `json:"session_token" db:"session_token"`
	IPAddress    *string    `json:"ip_address" db:"ip_address"`
	UserAgent    *string    `json:"user_agent" db:"user_agent"`
	LoginAt      time.Time  `json:"login_at" db:"login_at"`
	LogoutAt     *time.Time `json:"logout_at" db:"logout_at"`
	ExpiresAt    time.Time  `json:"expires_at" db:"expires_at"`
	IsActive     bool       `json:"is_active" db:"is_active"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}

// Password reset tokens
type PasswordResetToken struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	Token     string    `json:"token" db:"token"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	IsUsed    bool      `json:"is_used" db:"is_used"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

package repositories

import (
	"database/sql"
	"v01_system_backend/apps/models"

	"github.com/jmoiron/sqlx"
)

type SessionRepository struct {
	db *sqlx.DB
}

func NewSessionRepository(db *sqlx.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(session *models.UserSession) error {
	query := `
		INSERT INTO user_sessions (user_id, session_token, ip_address, user_agent, login_at, expires_at, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING session_id`

	return r.db.QueryRow(query,
		session.UserID,
		session.SessionToken,
		session.IPAddress,
		session.UserAgent,
		session.LoginAt,
		session.ExpiresAt,
		session.IsActive,
	).Scan(&session.SessionID)
}

func (r *SessionRepository) GetByToken(token string) (*models.UserSession, error) {
	var session models.UserSession
	query := `
		SELECT session_id, user_id, session_token, ip_address, user_agent, 
		       login_at, logout_at, expires_at, is_active, created_at
		FROM user_sessions 
		WHERE session_token = $1`

	err := r.db.Get(&session, query, token)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &session, nil
}

func (r *SessionRepository) Update(session *models.UserSession) error {
	query := `
		UPDATE user_sessions 
		SET session_token = $1, expires_at = $2
		WHERE session_id = $3`

	_, err := r.db.Exec(query, session.SessionToken, session.ExpiresAt, session.SessionID)
	return err
}

func (r *SessionRepository) DeactivateByToken(token string) error {
	query := `
		UPDATE user_sessions 
		SET is_active = false, logout_at = CURRENT_TIMESTAMP
		WHERE session_token = $1`

	_, err := r.db.Exec(query, token)
	return err
}

func (r *SessionRepository) DeactivateAllUserSessions(userID int) error {
	query := `
		UPDATE user_sessions 
		SET is_active = false, logout_at = CURRENT_TIMESTAMP
		WHERE user_id = $1 AND is_active = true`

	_, err := r.db.Exec(query, userID)
	return err
}

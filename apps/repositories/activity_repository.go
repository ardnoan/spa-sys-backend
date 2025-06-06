// apps/repositories/activity_repository.go
package repositories

import (
	"v01_system_backend/apps/models"

	"github.com/jmoiron/sqlx"
)

type ActivityRepository struct {
	db *sqlx.DB
}

func NewActivityRepository(db *sqlx.DB) *ActivityRepository {
	return &ActivityRepository{db: db}
}

func (r *ActivityRepository) Create(activity *models.UserActivityLog) error {
	query := `
		INSERT INTO user_activity_logs (user_id, action, description, ip_address, user_agent, created_at)
		VALUES (:user_id, :action, :description, :ip_address, :user_agent, :created_at)`

	_, err := r.db.NamedExec(query, activity)
	return err
}

func (r *ActivityRepository) GetByUserID(userID int, limit, offset int) ([]models.UserActivityLog, error) {
	var activities []models.UserActivityLog
	query := `
		SELECT * FROM user_activity_logs 
		WHERE user_id = $1 
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3`

	err := r.db.Select(&activities, query, userID, limit, offset)
	return activities, err
}

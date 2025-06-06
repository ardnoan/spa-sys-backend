// apps/repositories/activity_repository.go
package repositories

import (
	"v01_system_backend/apps/models"

	"gorm.io/gorm"
)

type ActivityRepository struct {
	db *gorm.DB
}

func NewActivityRepository(db *gorm.DB) *ActivityRepository {
	return &ActivityRepository{db: db}
}

func (r *ActivityRepository) Create(activity *models.UserActivityLog) error {
	return r.db.Create(activity).Error
}

func (r *ActivityRepository) GetByUserID(userID int, limit, offset int) ([]models.UserActivityLog, error) {
	var activities []models.UserActivityLog
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&activities).Error
	return activities, err
}

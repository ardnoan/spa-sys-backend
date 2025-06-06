// apps/models/user_activity_log.go
package models

import (
	"time"

	"gorm.io/gorm"
)

type UserActivityLog struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	UserID         *int           `json:"user_id" gorm:"index"`
	Action         string         `json:"action" gorm:"size:50;not null"`
	TargetType     *string        `json:"target_type" gorm:"size:50"`
	TargetID       *int           `json:"target_id"`
	MenuName       *string        `json:"menu_name" gorm:"size:100"`
	Description    *string        `json:"description" gorm:"size:255"`
	IPAddress      *string        `json:"ip_address" gorm:"size:45"`
	UserAgent      *string        `json:"user_agent" gorm:"size:255"`
	ResponseStatus *int           `json:"response_status"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (UserActivityLog) TableName() string {
	return "user_activity_logs"
}

package models

import "time"

// SystemSetting represents system configuration settings
type SystemSetting struct {
	SystemID     int        `json:"system_id" db:"system_id"`
	SettingKey   string     `json:"setting_key" db:"setting_key"`
	SettingValue string     `json:"setting_value" db:"setting_value"`
	SettingType  string     `json:"setting_type" db:"setting_type"`
	Description  *string    `json:"description" db:"description"`
	IsPublic     bool       `json:"is_public" db:"is_public"`
	IsActive     bool       `json:"is_active" db:"is_active"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	CreatedBy    string     `json:"created_by" db:"created_by"`
	UpdatedAt    *time.Time `json:"updated_at" db:"updated_at"`
	UpdatedBy    *string    `json:"updated_by" db:"updated_by"`
}

// SystemSettingCreateRequest for creating new system settings
type SystemSettingCreateRequest struct {
	SettingKey   string  `json:"setting_key" validate:"required,max=100"`
	SettingValue string  `json:"setting_value" validate:"required"`
	SettingType  string  `json:"setting_type" validate:"required,oneof=string number boolean json"`
	Description  *string `json:"description"`
	IsPublic     bool    `json:"is_public"`
}

// SystemSettingUpdateRequest for updating system settings
type SystemSettingUpdateRequest struct {
	SettingValue string  `json:"setting_value" validate:"required"`
	SettingType  string  `json:"setting_type" validate:"required,oneof=string number boolean json"`
	Description  *string `json:"description"`
	IsPublic     bool    `json:"is_public"`
}

package services

import (
	"encoding/json"
	"errors"
	"strconv"
	"v01_system_backend/apps/models"
	"v01_system_backend/apps/repositories"
)

type SystemService struct {
	systemRepo *repositories.SystemRepository
}

func NewSystemService(systemRepo *repositories.SystemRepository) *SystemService {
	return &SystemService{systemRepo: systemRepo}
}

func (s *SystemService) GetAllSettings() ([]models.SystemSetting, error) {
	pagination := &models.PaginationRequest{
		Page:     1,
		PageSize: 1000, // Get all settings
	}
	pagination.SetDefaults()

	settings, _, err := s.systemRepo.GetAll(pagination)
	return settings, err
}

func (s *SystemService) GetPublicSettings() ([]*models.SystemSetting, error) {
	return s.systemRepo.GetPublicSettings()
}

func (s *SystemService) GetSettingByKey(key string) (*models.SystemSetting, error) {
	setting, err := s.systemRepo.GetByKey(key)
	if err != nil {
		return nil, err
	}

	if setting == nil {
		return nil, errors.New("setting not found")
	}

	return setting, nil
}

func (s *SystemService) UpdateSetting(key string, value string, updatedBy int) (*models.SystemSetting, error) {
	// Check if setting exists
	existingSetting, err := s.systemRepo.GetByKey(key)
	if err != nil {
		return nil, err
	}
	if existingSetting == nil {
		return nil, errors.New("setting not found")
	}

	// Create update request
	req := &models.SystemSettingUpdateRequest{
		SettingValue: value,
		SettingType:  existingSetting.SettingType,
		Description:  existingSetting.Description,
		IsPublic:     existingSetting.IsPublic,
	}

	// Validate setting value based on type
	if err := s.validateSettingValue(req.SettingValue, req.SettingType); err != nil {
		return nil, err
	}

	return s.systemRepo.Update(key, req, updatedBy)
}

// User Status Methods
func (s *SystemService) GetUserStatuses() ([]models.UserStatus, error) {
	return s.systemRepo.GetUserStatuses()
}

// Department Methods
func (s *SystemService) GetDepartments() ([]models.Department, error) {
	return s.systemRepo.GetDepartments()
}

func (s *SystemService) CreateDepartment(req *models.DepartmentCreateRequest, createdBy int) (*models.Department, error) {
	return s.systemRepo.CreateDepartment(req, createdBy)
}

func (s *SystemService) GetSettingValue(key string) (interface{}, error) {
	setting, err := s.GetSettingByKey(key)
	if err != nil {
		return nil, err
	}

	return s.parseSettingValue(setting)
}

func (s *SystemService) GetStringValue(key string, defaultValue string) string {
	value, err := s.GetSettingValue(key)
	if err != nil {
		return defaultValue
	}

	if str, ok := value.(string); ok {
		return str
	}

	return defaultValue
}

func (s *SystemService) GetIntValue(key string, defaultValue int) int {
	value, err := s.GetSettingValue(key)
	if err != nil {
		return defaultValue
	}

	switch v := value.(type) {
	case int:
		return v
	case float64:
		return int(v)
	case string:
		if intVal, err := strconv.Atoi(v); err == nil {
			return intVal
		}
	}

	return defaultValue
}

func (s *SystemService) GetBoolValue(key string, defaultValue bool) bool {
	value, err := s.GetSettingValue(key)
	if err != nil {
		return defaultValue
	}

	switch v := value.(type) {
	case bool:
		return v
	case string:
		if boolVal, err := strconv.ParseBool(v); err == nil {
			return boolVal
		}
	}

	return defaultValue
}

func (s *SystemService) validateSettingValue(value, settingType string) error {
	switch settingType {
	case "number":
		if _, err := strconv.ParseFloat(value, 64); err != nil {
			return errors.New("invalid number value")
		}
	case "boolean":
		if _, err := strconv.ParseBool(value); err != nil {
			return errors.New("invalid boolean value")
		}
	case "json":
		var js interface{}
		if err := json.Unmarshal([]byte(value), &js); err != nil {
			return errors.New("invalid json value")
		}
	}
	return nil
}

func (s *SystemService) parseSettingValue(setting *models.SystemSetting) (interface{}, error) {
	switch setting.SettingType {
	case "number":
		return strconv.ParseFloat(setting.SettingValue, 64)
	case "boolean":
		return strconv.ParseBool(setting.SettingValue)
	case "json":
		var result interface{}
		err := json.Unmarshal([]byte(setting.SettingValue), &result)
		return result, err
	default:
		return setting.SettingValue, nil
	}
}

// Application configuration methods
func (s *SystemService) GetAppName() string {
	return s.GetStringValue("app_name", "SPA Ardnoan")
}

func (s *SystemService) GetAppVersion() string {
	return s.GetStringValue("app_version", "1.0.0")
}

func (s *SystemService) GetMaxLoginAttempts() int {
	return s.GetIntValue("max_login_attempts", 5)
}

func (s *SystemService) GetAccountLockDuration() int {
	return s.GetIntValue("account_lock_duration_minutes", 30)
}

func (s *SystemService) GetPasswordMinLength() int {
	return s.GetIntValue("password_min_length", 8)
}

func (s *SystemService) GetSessionTimeout() int {
	return s.GetIntValue("session_timeout_hours", 24)
}

func (s *SystemService) IsMaintenanceMode() bool {
	return s.GetBoolValue("maintenance_mode", false)
}

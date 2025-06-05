package models

type PermissionCreateRequest struct {
	PermissionCode string `json:"permission_code" validate:"required,max=50"`
	PermissionName string `json:"permission_name" validate:"required,max=100"`
	Description    string `json:"description"`
	ModuleName     string `json:"module_name" validate:"max=50"`
}

type PermissionUpdateRequest struct {
	PermissionCode string `json:"permission_code" validate:"required,max=50"`
	PermissionName string `json:"permission_name" validate:"required,max=100"`
	Description    string `json:"description"`
	ModuleName     string `json:"module_name" validate:"max=50"`
}

type UserActivityLog struct {
	LogsID         int     `json:"logs_id" db:"logs_id"`
	UserID         *int    `json:"user_id" db:"user_id"`
	SessionID      *int    `json:"session_id" db:"session_id"`
	Action         string  `json:"action" db:"action"`
	TargetType     *string `json:"target_type" db:"target_type"`
	TargetID       *int    `json:"target_id" db:"target_id"`
	MenuName       *string `json:"menu_name" db:"menu_name"`
	Description    *string `json:"description" db:"description"`
	IPAddress      *string `json:"ip_address" db:"ip_address"`
	UserAgent      *string `json:"user_agent" db:"user_agent"`
	RequestData    *string `json:"request_data" db:"request_data"`
	ResponseStatus *int    `json:"response_status" db:"response_status"`
	BaseModel
}

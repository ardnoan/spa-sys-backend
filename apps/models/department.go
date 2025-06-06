package models

import "time"

// Department represents the departments table
type Department struct {
	DepartmentID   int        `json:"department_id" db:"department_id"`
	DepartmentName string     `json:"department_name" db:"department_name"`
	DepartmentCode string     `json:"department_code" db:"department_code"`
	ParentID       *int       `json:"parent_id" db:"parent_id"`
	ManagerID      *int       `json:"manager_id" db:"manager_id"`
	Description    *string    `json:"description" db:"description"`
	IsActive       bool       `json:"is_active" db:"is_active"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	CreatedBy      string     `json:"created_by" db:"created_by"`
	UpdatedAt      *time.Time `json:"updated_at" db:"updated_at"`
	UpdatedBy      *string    `json:"updated_by" db:"updated_by"`
}

// DepartmentCreateRequest for creating new departments
type DepartmentCreateRequest struct {
	DepartmentName string  `json:"department_name" validate:"required,max=100"`
	DepartmentCode string  `json:"department_code" validate:"required,max=20"`
	ParentID       *int    `json:"parent_id"`
	ManagerID      *int    `json:"manager_id"`
	Description    *string `json:"description"`
}

package models

import "time"

type Department struct {
	DepartmentID   int       `json:"department_id" db:"department_id"`
	DepartmentName string    `json:"department_name" db:"department_name"`
	DepartmentCode string    `json:"department_code" db:"department_code"`
	ParentID       *int      `json:"parent_id" db:"parent_id"`
	ManagerID      *int      `json:"manager_id" db:"manager_id"`
	Description    *string   `json:"description" db:"description"`
	IsActive       bool      `json:"is_active" db:"is_active"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	CreatedBy      *string   `json:"created_by" db:"created_by"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
	UpdatedBy      *string   `json:"updated_by" db:"updated_by"`
}

type CreateDepartmentRequest struct {
	DepartmentName string  `json:"department_name" validate:"required,min=3,max=100"`
	DepartmentCode string  `json:"department_code" validate:"required,min=2,max=20"`
	ParentID       *int    `json:"parent_id" validate:"omitempty,min=1"`
	ManagerID      *int    `json:"manager_id" validate:"omitempty,min=1"`
	Description    *string `json:"description" validate:"omitempty,max=500"`
	IsActive       bool    `json:"is_active"`
}

type UpdateDepartmentRequest struct {
	DepartmentName string  `json:"department_name" validate:"required,min=3,max=100"`
	DepartmentCode string  `json:"department_code" validate:"required,min=2,max=20"`
	ParentID       *int    `json:"parent_id" validate:"omitempty,min=1"`
	ManagerID      *int    `json:"manager_id" validate:"omitempty,min=1"`
	Description    *string `json:"description" validate:"omitempty,max=500"`
	IsActive       bool    `json:"is_active"`
}

type DepartmentFilter struct {
	Page     int    `json:"page" validate:"min=1"`
	Limit    int    `json:"limit" validate:"min=1,max=100"`
	Search   string `json:"search" validate:"omitempty,max=100"`
	IsActive *bool  `json:"is_active"`
	ParentID *int   `json:"parent_id"`
}

type DepartmentHierarchy struct {
	DepartmentID   int     `json:"department_id"`
	DepartmentName string  `json:"department_name"`
	DepartmentCode string  `json:"department_code"`
	ParentID       *int    `json:"parent_id"`
	ManagerID      *int    `json:"manager_id"`
	Description    *string `json:"description"`
	IsActive       bool    `json:"is_active"`
	Level          int     `json:"level"`
	IsRoot         bool    `json:"is_root"`
}

type User struct {
	ID         int       `json:"id"`
	Username   string    `json:"username"`
	Email      string    `json:"email"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	EmployeeID *string   `json:"employee_id"`
	Phone      *string   `json:"phone"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
}

type PaginationResponse struct {
	CurrentPage int `json:"current_page"`
	PerPage     int `json:"per_page"`
	TotalCount  int `json:"total_count"`
	TotalPages  int `json:"total_pages"`
}

type DepartmentListResponse struct {
	Departments []Department       `json:"departments"`
	Pagination  PaginationResponse `json:"pagination"`
}

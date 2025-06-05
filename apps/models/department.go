package models

type Department struct {
	DepartmentID   int     `json:"department_id" db:"department_id"`
	DepartmentName string  `json:"department_name" db:"department_name"`
	DepartmentCode string  `json:"department_code" db:"department_code"`
	ParentID       *int    `json:"parent_id" db:"parent_id"`
	ManagerID      *int    `json:"manager_id" db:"manager_id"`
	Description    *string `json:"description" db:"description"`
	IsActive       bool    `json:"is_active" db:"is_active"`
	BaseModel

	// Relations
	ManagerName string       `json:"manager_name,omitempty" db:"manager_name"`
	ParentName  string       `json:"parent_name,omitempty" db:"parent_name"`
	Children    []Department `json:"children,omitempty"`
}

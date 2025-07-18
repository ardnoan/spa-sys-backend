package repositories

import (
	"database/sql"
	"fmt"
	"v01_system_backend/models"
)

type DepartmentRepository interface {
	Create(dept *models.CreateDepartmentRequest, createdBy string) (int, error)
	Update(id int, dept *models.UpdateDepartmentRequest, updatedBy string) error
	Delete(id int, deletedBy string) error

	GetByID(id int) (*models.Department, error)
	GetAll(filter *models.DepartmentFilter) ([]models.Department, int, error)
	GetHierarchy() ([]models.DepartmentHierarchy, error)
	GetUsersByDepartment(departmentID int) ([]models.User, error)
	Search(query string) ([]models.Department, error)
	ExistsByCode(code string) (bool, error)
	ExistsByCodeExcludeID(code string, id int) (bool, error)
	ExistsByID(id int) (bool, error)
	HasActiveChildren(id int) (bool, error)
}

type departmentRepository struct {
	db *sql.DB
}

func NewDepartmentRepository(db *sql.DB) DepartmentRepository {
	return &departmentRepository{db: db}
}

func (r *departmentRepository) Create(dept *models.CreateDepartmentRequest, createdBy string) (int, error) {
	query := `INSERT INTO departments 
              (department_name, department_code, parent_id, manager_id, description, is_active, created_by) 
              VALUES ($1, $2, $3, $4, $5, $6, $7) 
              RETURNING department_id`

	var departmentID int
	err := r.db.QueryRow(query,
		dept.DepartmentName, dept.DepartmentCode, dept.ParentID,
		dept.ManagerID, dept.Description, dept.IsActive, createdBy).Scan(&departmentID)

	if err != nil {
		return 0, fmt.Errorf("failed to create department: %w", err)
	}

	return departmentID, nil
}

func (r *departmentRepository) GetByID(id int) (*models.Department, error) {
	var department models.Department
	query := `SELECT department_id, department_name, department_code, parent_id, 
              manager_id, description, is_active, created_at, created_by, updated_at, updated_by
              FROM departments WHERE department_id = $1`

	err := r.db.QueryRow(query, id).Scan(
		&department.DepartmentID, &department.DepartmentName, &department.DepartmentCode,
		&department.ParentID, &department.ManagerID, &department.Description,
		&department.IsActive, &department.CreatedAt, &department.CreatedBy,
		&department.UpdatedAt, &department.UpdatedBy)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("department not found")
		}
		return nil, fmt.Errorf("failed to get department: %w", err)
	}

	return &department, nil
}

func (r *departmentRepository) GetAll(filter *models.DepartmentFilter) ([]models.Department, int, error) {
	// Build WHERE clause
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	// Default: only show root departments (parent_id IS NULL)
	// Unless specific parent_id is provided
	if filter.ParentID == nil {
		whereClause += " AND parent_id IS NULL"
	} else {
		whereClause += " AND parent_id = $" + fmt.Sprintf("%d", argIndex)
		args = append(args, *filter.ParentID)
		argIndex++
	}

	if filter.Search != "" {
		whereClause += " AND (LOWER(department_name) LIKE LOWER($" + fmt.Sprintf("%d", argIndex) +
			") OR LOWER(department_code) LIKE LOWER($" + fmt.Sprintf("%d", argIndex) +
			") OR LOWER(description) LIKE LOWER($" + fmt.Sprintf("%d", argIndex) + "))"
		args = append(args, "%"+filter.Search+"%")
		argIndex++
	}

	if filter.IsActive != nil {
		whereClause += " AND is_active = $" + fmt.Sprintf("%d", argIndex)
		args = append(args, *filter.IsActive)
		argIndex++
	}

	// Get total count
	var totalCount int
	countQuery := "SELECT COUNT(*) FROM departments " + whereClause
	err := r.db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count departments: %w", err)
	}

	// Get departments with pagination
	offset := (filter.Page - 1) * filter.Limit
	query := `SELECT department_id, department_name, department_code, parent_id, 
              manager_id, description, is_active, created_at, created_by, updated_at, updated_by
              FROM departments ` + whereClause + ` 
              ORDER BY created_at DESC 
              LIMIT $` + fmt.Sprintf("%d", argIndex) + ` OFFSET $` + fmt.Sprintf("%d", argIndex+1)

	args = append(args, filter.Limit, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get departments: %w", err)
	}
	defer rows.Close()

	var departments []models.Department
	for rows.Next() {
		var department models.Department
		err := rows.Scan(&department.DepartmentID, &department.DepartmentName,
			&department.DepartmentCode, &department.ParentID, &department.ManagerID,
			&department.Description, &department.IsActive, &department.CreatedAt,
			&department.CreatedBy, &department.UpdatedAt, &department.UpdatedBy)
		if err != nil {
			continue
		}
		departments = append(departments, department)
	}

	return departments, totalCount, nil
}

func (r *departmentRepository) Update(id int, dept *models.UpdateDepartmentRequest, updatedBy string) error {
	query := `UPDATE departments 
              SET department_name = $1, department_code = $2, parent_id = $3, 
                  manager_id = $4, description = $5, is_active = $6,
                  updated_by = $7, updated_at = CURRENT_TIMESTAMP
              WHERE department_id = $8`

	_, err := r.db.Exec(query, dept.DepartmentName, dept.DepartmentCode, dept.ParentID,
		dept.ManagerID, dept.Description, dept.IsActive, updatedBy, id)

	if err != nil {
		return fmt.Errorf("failed to update department: %w", err)
	}

	return nil
}

func (r *departmentRepository) Delete(id int, deletedBy string) error {
	query := `UPDATE departments 
              SET is_active = false, updated_by = $1, updated_at = CURRENT_TIMESTAMP 
              WHERE department_id = $2`

	_, err := r.db.Exec(query, deletedBy, id)
	if err != nil {
		return fmt.Errorf("failed to delete department: %w", err)
	}

	return nil
}

func (r *departmentRepository) GetHierarchy() ([]models.DepartmentHierarchy, error) {
	query := `WITH RECURSIVE dept_hierarchy AS (
		SELECT department_id, department_name, department_code, parent_id, 
		       manager_id, description, is_active, 0 as level,
		       ARRAY[department_id] as path
		FROM departments 
		WHERE parent_id IS NULL AND is_active = true
		
		UNION ALL
		
		SELECT d.department_id, d.department_name, d.department_code, d.parent_id,
		       d.manager_id, d.description, d.is_active, dh.level + 1,
		       dh.path || d.department_id
		FROM departments d
		INNER JOIN dept_hierarchy dh ON d.parent_id = dh.department_id
		WHERE d.is_active = true
	)
	SELECT department_id, department_name, department_code, parent_id, 
	       manager_id, description, is_active, level FROM dept_hierarchy ORDER BY path`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get department hierarchy: %w", err)
	}
	defer rows.Close()

	var hierarchy []models.DepartmentHierarchy

	// Add hardcoded root node
	rootNode := models.DepartmentHierarchy{
		DepartmentID:   0,
		DepartmentName: "Department Hierarchy",
		DepartmentCode: "ROOT",
		ParentID:       nil,
		ManagerID:      nil,
		Description:    stringPtr("Root department hierarchy"),
		IsActive:       true,
		Level:          -1,
		IsRoot:         true,
	}
	hierarchy = append(hierarchy, rootNode)

	for rows.Next() {
		var dept models.DepartmentHierarchy
		err := rows.Scan(&dept.DepartmentID, &dept.DepartmentName, &dept.DepartmentCode,
			&dept.ParentID, &dept.ManagerID, &dept.Description, &dept.IsActive, &dept.Level)
		if err != nil {
			continue
		}
		dept.IsRoot = false
		hierarchy = append(hierarchy, dept)
	}

	return hierarchy, nil
}

func (r *departmentRepository) GetUsersByDepartment(departmentID int) ([]models.User, error) {
	query := `SELECT user_apps_id, username, email, first_name, last_name, 
              employee_id, phone, is_active, created_at
              FROM users_application 
              WHERE department_id = $1 AND is_active = true 
              ORDER BY first_name, last_name`

	rows, err := r.db.Query(query, departmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get users by department: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.FirstName, &user.LastName,
			&user.EmployeeID, &user.Phone, &user.IsActive, &user.CreatedAt)
		if err != nil {
			continue
		}
		users = append(users, user)
	}

	return users, nil
}

func (r *departmentRepository) Search(query string) ([]models.Department, error) {
	searchQuery := `SELECT department_id, department_name, department_code, parent_id, 
                    manager_id, description, is_active, created_at, created_by, updated_at, updated_by
                    FROM departments 
                    WHERE is_active = true AND (
                        LOWER(department_name) LIKE LOWER($1) OR 
                        LOWER(department_code) LIKE LOWER($1) OR 
                        LOWER(description) LIKE LOWER($1)
                    )
                    ORDER BY department_name 
                    LIMIT 50`

	searchPattern := "%" + query + "%"
	rows, err := r.db.Query(searchQuery, searchPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to search departments: %w", err)
	}
	defer rows.Close()

	var departments []models.Department
	for rows.Next() {
		var department models.Department
		err := rows.Scan(&department.DepartmentID, &department.DepartmentName,
			&department.DepartmentCode, &department.ParentID, &department.ManagerID,
			&department.Description, &department.IsActive, &department.CreatedAt,
			&department.CreatedBy, &department.UpdatedAt, &department.UpdatedBy)
		if err != nil {
			continue
		}
		departments = append(departments, department)
	}

	return departments, nil
}

func (r *departmentRepository) ExistsByCode(code string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM departments WHERE department_code = $1)`
	err := r.db.QueryRow(query, code).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check department code existence: %w", err)
	}
	return exists, nil
}

func (r *departmentRepository) ExistsByCodeExcludeID(code string, id int) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM departments WHERE department_code = $1 AND department_id != $2)`
	err := r.db.QueryRow(query, code, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check department code existence: %w", err)
	}
	return exists, nil
}

func (r *departmentRepository) ExistsByID(id int) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM departments WHERE department_id = $1)`
	err := r.db.QueryRow(query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check department existence: %w", err)
	}
	return exists, nil
}

func (r *departmentRepository) HasActiveChildren(id int) (bool, error) {
	var hasChildren bool
	query := `SELECT EXISTS(SELECT 1 FROM departments WHERE parent_id = $1 AND is_active = true)`
	err := r.db.QueryRow(query, id).Scan(&hasChildren)
	if err != nil {
		return false, fmt.Errorf("failed to check child departments: %w", err)
	}
	return hasChildren, nil
}

// Helper function
func stringPtr(s string) *string {
	return &s
}

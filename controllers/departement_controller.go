package controller

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type Department struct {
	DepartmentID   int     `json:"department_id"`
	DepartmentName string  `json:"department_name"`
	DepartmentCode string  `json:"department_code"`
	ParentID       *int    `json:"parent_id"`
	ManagerID      *int    `json:"manager_id"`
	Description    *string `json:"description"`
	IsActive       bool    `json:"is_active"`
	CreatedAt      string  `json:"created_at"`
	CreatedBy      *string `json:"created_by"`
	UpdatedAt      string  `json:"updated_at"`
	UpdatedBy      *string `json:"updated_by"`
}

type CreateDepartmentRequest struct {
	DepartmentName string  `json:"department_name" validate:"required"`
	DepartmentCode string  `json:"department_code" validate:"required"`
	ParentID       *int    `json:"parent_id"`
	ManagerID      *int    `json:"manager_id"`
	Description    *string `json:"description"`
	IsActive       bool    `json:"is_active"`
}

type UpdateDepartmentRequest struct {
	DepartmentName string  `json:"department_name" validate:"required"`
	DepartmentCode string  `json:"department_code" validate:"required"`
	ParentID       *int    `json:"parent_id"`
	ManagerID      *int    `json:"manager_id"`
	Description    *string `json:"description"`
	IsActive       bool    `json:"is_active"`
}

type DepartmentController struct {
	DB *sql.DB
}

func NewDepartmentController(db *sql.DB) *DepartmentController {
	return &DepartmentController{DB: db}
}

// Response helpers
func (dc *DepartmentController) successResponse(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

func (dc *DepartmentController) errorResponse(c echo.Context, code int, message string) error {
	return c.JSON(code, map[string]interface{}{
		"success": false,
		"message": message,
	})
}

// Create Department
func (dc *DepartmentController) CreateDepartment(c echo.Context) error {
	var req CreateDepartmentRequest
	if err := c.Bind(&req); err != nil {
		return dc.errorResponse(c, http.StatusBadRequest, "Invalid request")
	}

	// Check if department code already exists
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM departments WHERE department_code = $1)`
	err := dc.DB.QueryRow(checkQuery, req.DepartmentCode).Scan(&exists)
	if err != nil {
		return dc.errorResponse(c, http.StatusInternalServerError, "Failed to check department code")
	}
	if exists {
		return dc.errorResponse(c, http.StatusConflict, "Department code already exists")
	}

	// Insert to database
	query := `INSERT INTO departments 
              (department_name, department_code, parent_id, manager_id, description, is_active, created_by) 
              VALUES ($1, $2, $3, $4, $5, $6, $7) 
              RETURNING department_id`

	var departmentID int
	err = dc.DB.QueryRow(query,
		req.DepartmentName, req.DepartmentCode, req.ParentID,
		req.ManagerID, req.Description, req.IsActive, "system").Scan(&departmentID)

	if err != nil {
		return dc.errorResponse(c, http.StatusInternalServerError, "Failed to create department: "+err.Error())
	}

	return dc.successResponse(c, map[string]interface{}{
		"department_id": departmentID,
		"message":       "Department created successfully",
	})
}

// Get Department by ID
func (dc *DepartmentController) GetDepartment(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return dc.errorResponse(c, http.StatusBadRequest, "Invalid department ID")
	}

	var department Department
	query := `SELECT department_id, department_name, department_code, parent_id, 
              manager_id, description, is_active, created_at, created_by, updated_at, updated_by
              FROM departments WHERE department_id = $1`

	err = dc.DB.QueryRow(query, id).Scan(
		&department.DepartmentID, &department.DepartmentName, &department.DepartmentCode,
		&department.ParentID, &department.ManagerID, &department.Description,
		&department.IsActive, &department.CreatedAt, &department.CreatedBy,
		&department.UpdatedAt, &department.UpdatedBy)

	if err != nil {
		if err == sql.ErrNoRows {
			return dc.errorResponse(c, http.StatusNotFound, "Department not found")
		}
		return dc.errorResponse(c, http.StatusInternalServerError, "Failed to fetch department")
	}

	return dc.successResponse(c, department)
}

// Get All Departments with pagination and filters
func (dc *DepartmentController) GetAllDepartments(c echo.Context) error {
	// Get query parameters
	page := c.QueryParam("page")
	limit := c.QueryParam("limit")
	search := c.QueryParam("search")
	isActive := c.QueryParam("is_active")
	parentID := c.QueryParam("parent_id")

	// Set default values
	pageInt := 1
	limitInt := 10

	if page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			pageInt = p
		}
	}

	if limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 && l <= 100 {
			limitInt = l
		}
	}

	offset := (pageInt - 1) * limitInt

	// Build WHERE clause
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	if search != "" {
		whereClause += " AND (LOWER(department_name) LIKE LOWER($" + strconv.Itoa(argIndex) +
			") OR LOWER(department_code) LIKE LOWER($" + strconv.Itoa(argIndex) +
			") OR LOWER(description) LIKE LOWER($" + strconv.Itoa(argIndex) + "))"
		args = append(args, "%"+search+"%")
		argIndex++
	}

	if isActive != "" {
		if isActive == "true" {
			whereClause += " AND is_active = $" + strconv.Itoa(argIndex)
			args = append(args, true)
		} else if isActive == "false" {
			whereClause += " AND is_active = $" + strconv.Itoa(argIndex)
			args = append(args, false)
		}
		argIndex++
	}

	if parentID != "" {
		if parentID == "null" {
			whereClause += " AND parent_id IS NULL"
		} else {
			whereClause += " AND parent_id = $" + strconv.Itoa(argIndex)
			if pid, err := strconv.Atoi(parentID); err == nil {
				args = append(args, pid)
				argIndex++
			}
		}
	}

	// Get total count
	var totalCount int
	countQuery := "SELECT COUNT(*) FROM departments " + whereClause
	err := dc.DB.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return dc.errorResponse(c, http.StatusInternalServerError, "Failed to count departments")
	}

	// Get departments with pagination
	query := `SELECT department_id, department_name, department_code, parent_id, 
              manager_id, description, is_active, created_at, created_by, updated_at, updated_by
              FROM departments ` + whereClause + ` 
              ORDER BY created_at DESC 
              LIMIT $` + strconv.Itoa(argIndex) + ` OFFSET $` + strconv.Itoa(argIndex+1)

	args = append(args, limitInt, offset)
	rows, err := dc.DB.Query(query, args...)
	if err != nil {
		return dc.errorResponse(c, http.StatusInternalServerError, "Failed to fetch departments")
	}
	defer rows.Close()

	var departments []Department
	for rows.Next() {
		var department Department
		err := rows.Scan(&department.DepartmentID, &department.DepartmentName,
			&department.DepartmentCode, &department.ParentID, &department.ManagerID,
			&department.Description, &department.IsActive, &department.CreatedAt,
			&department.CreatedBy, &department.UpdatedAt, &department.UpdatedBy)
		if err != nil {
			continue
		}
		departments = append(departments, department)
	}

	// Calculate pagination info
	totalPages := (totalCount + limitInt - 1) / limitInt

	return dc.successResponse(c, map[string]interface{}{
		"departments": departments,
		"pagination": map[string]interface{}{
			"current_page": pageInt,
			"per_page":     limitInt,
			"total_count":  totalCount,
			"total_pages":  totalPages,
		},
	})
}

// Update Department
func (dc *DepartmentController) UpdateDepartment(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return dc.errorResponse(c, http.StatusBadRequest, "Invalid department ID")
	}

	var req UpdateDepartmentRequest
	if err := c.Bind(&req); err != nil {
		return dc.errorResponse(c, http.StatusBadRequest, "Invalid request")
	}

	// Check if department exists
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM departments WHERE department_id = $1)`
	err = dc.DB.QueryRow(checkQuery, id).Scan(&exists)
	if err != nil || !exists {
		return dc.errorResponse(c, http.StatusNotFound, "Department not found")
	}

	// Check if department code already exists (excluding current department)
	var codeExists bool
	codeCheckQuery := `SELECT EXISTS(SELECT 1 FROM departments WHERE department_code = $1 AND department_id != $2)`
	err = dc.DB.QueryRow(codeCheckQuery, req.DepartmentCode, id).Scan(&codeExists)
	if err != nil {
		return dc.errorResponse(c, http.StatusInternalServerError, "Failed to check department code")
	}
	if codeExists {
		return dc.errorResponse(c, http.StatusConflict, "Department code already exists")
	}

	// Prevent setting parent_id to self
	if req.ParentID != nil && *req.ParentID == id {
		return dc.errorResponse(c, http.StatusBadRequest, "Department cannot be its own parent")
	}

	query := `UPDATE departments 
              SET department_name = $1, department_code = $2, parent_id = $3, 
                  manager_id = $4, description = $5, is_active = $6,
                  updated_by = $7, updated_at = CURRENT_TIMESTAMP
              WHERE department_id = $8`

	_, err = dc.DB.Exec(query, req.DepartmentName, req.DepartmentCode, req.ParentID,
		req.ManagerID, req.Description, req.IsActive, "system", id)

	if err != nil {
		return dc.errorResponse(c, http.StatusInternalServerError, "Failed to update department: "+err.Error())
	}

	return dc.successResponse(c, map[string]string{"message": "Department updated successfully"})
}

// Delete Department (Soft delete)
func (dc *DepartmentController) DeleteDepartment(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return dc.errorResponse(c, http.StatusBadRequest, "Invalid department ID")
	}

	// Check if department exists
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM departments WHERE department_id = $1)`
	err = dc.DB.QueryRow(checkQuery, id).Scan(&exists)
	if err != nil || !exists {
		return dc.errorResponse(c, http.StatusNotFound, "Department not found")
	}

	// Check if department has child departments
	var hasChildren bool
	childCheckQuery := `SELECT EXISTS(SELECT 1 FROM departments WHERE parent_id = $1 AND is_active = true)`
	err = dc.DB.QueryRow(childCheckQuery, id).Scan(&hasChildren)
	if err != nil {
		return dc.errorResponse(c, http.StatusInternalServerError, "Failed to check child departments")
	}
	if hasChildren {
		return dc.errorResponse(c, http.StatusConflict, "Cannot delete department with active child departments")
	}

	query := `UPDATE departments 
              SET is_active = false, updated_by = $1, updated_at = CURRENT_TIMESTAMP 
              WHERE department_id = $2`

	_, err = dc.DB.Exec(query, "system", id)
	if err != nil {
		return dc.errorResponse(c, http.StatusInternalServerError, "Failed to delete department")
	}

	return dc.successResponse(c, map[string]string{"message": "Department deleted successfully"})
}

// Get Department Hierarchy
func (dc *DepartmentController) GetDepartmentHierarchy(c echo.Context) error {
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
	SELECT * FROM dept_hierarchy ORDER BY path`

	rows, err := dc.DB.Query(query)
	if err != nil {
		return dc.errorResponse(c, http.StatusInternalServerError, "Failed to fetch department hierarchy")
	}
	defer rows.Close()

	var hierarchy []map[string]interface{}

	// Add hardcoded root node
	rootNode := map[string]interface{}{
		"department_id":   0,
		"department_name": "Department Hierarchy",
		"department_code": "ROOT",
		"parent_id":       nil,
		"manager_id":      nil,
		"description":     "Root department hierarchy",
		"is_active":       true,
		"level":           -1,
		"is_root":         true,
	}
	hierarchy = append(hierarchy, rootNode)

	for rows.Next() {
		var dept map[string]interface{}
		var departmentID int
		var departmentName, departmentCode string
		var parentID, managerID *int
		var description *string
		var isActive bool
		var level int
		var path string

		err := rows.Scan(&departmentID, &departmentName, &departmentCode,
			&parentID, &managerID, &description, &isActive, &level, &path)
		if err != nil {
			continue
		}

		dept = map[string]interface{}{
			"department_id":   departmentID,
			"department_name": departmentName,
			"department_code": departmentCode,
			"parent_id":       parentID,
			"manager_id":      managerID,
			"description":     description,
			"is_active":       isActive,
			"level":           level,
			"is_root":         false,
		}
		hierarchy = append(hierarchy, dept)
	}

	return dc.successResponse(c, hierarchy)
}

// Get Users by Department
func (dc *DepartmentController) GetUsersByDepartment(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return dc.errorResponse(c, http.StatusBadRequest, "Invalid department ID")
	}

	query := `SELECT user_apps_id, username, email, first_name, last_name, 
              employee_id, phone, is_active, created_at
              FROM users_application 
              WHERE department_id = $1 AND is_active = true 
              ORDER BY first_name, last_name`

	rows, err := dc.DB.Query(query, id)
	if err != nil {
		return dc.errorResponse(c, http.StatusInternalServerError, "Failed to fetch department users")
	}
	defer rows.Close()

	var users []map[string]interface{}
	for rows.Next() {
		var user map[string]interface{}
		var userID int
		var username, email, firstName, lastName string
		var employeeID, phone *string
		var isActive bool
		var createdAt string

		err := rows.Scan(&userID, &username, &email, &firstName, &lastName,
			&employeeID, &phone, &isActive, &createdAt)
		if err != nil {
			continue
		}

		user = map[string]interface{}{
			"id":          userID,
			"username":    username,
			"email":       email,
			"first_name":  firstName,
			"last_name":   lastName,
			"employee_id": employeeID,
			"phone":       phone,
			"is_active":   isActive,
			"created_at":  createdAt,
		}
		users = append(users, user)
	}

	return dc.successResponse(c, users)
}

// Search Departments
func (dc *DepartmentController) SearchDepartments(c echo.Context) error {
	query := c.QueryParam("q")
	if query == "" {
		return dc.errorResponse(c, http.StatusBadRequest, "Search query is required")
	}

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
	rows, err := dc.DB.Query(searchQuery, searchPattern)
	if err != nil {
		return dc.errorResponse(c, http.StatusInternalServerError, "Failed to search departments")
	}
	defer rows.Close()

	var departments []Department
	for rows.Next() {
		var department Department
		err := rows.Scan(&department.DepartmentID, &department.DepartmentName,
			&department.DepartmentCode, &department.ParentID, &department.ManagerID,
			&department.Description, &department.IsActive, &department.CreatedAt,
			&department.CreatedBy, &department.UpdatedAt, &department.UpdatedBy)
		if err != nil {
			continue
		}
		departments = append(departments, department)
	}

	return dc.successResponse(c, departments)
}

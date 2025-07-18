package controller

import (
	"net/http"
	"strconv"

	"v01_system_backend/models"
	"v01_system_backend/services"

	"github.com/labstack/echo/v4"
)

type DepartmentController struct {
	service services.DepartmentService
}

func NewDepartmentController(service services.DepartmentService) *DepartmentController {
	return &DepartmentController{service: service}
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

// @Summary Create a new department
// @Description Create a new department with provided details
// @Tags departments
// @Accept json
// @Produce json
// @Param request body models.CreateDepartmentRequest true "Department creation request"
// @Success 201 {object} map[string]interface{} "Department created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 409 {object} map[string]interface{} "Department code already exists"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /departments [post]
func (dc *DepartmentController) CreateDepartment(c echo.Context) error {
	var req models.CreateDepartmentRequest
	if err := c.Bind(&req); err != nil {
		return dc.errorResponse(c, http.StatusBadRequest, "Invalid request format")
	}

	// You can add validation here using a validation library
	if err := c.Validate(&req); err != nil {
		return dc.errorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
	}

	departmentID, err := dc.service.CreateDepartment(&req, "system")
	if err != nil {
		if err.Error() == "department code already exists" {
			return dc.errorResponse(c, http.StatusConflict, err.Error())
		}
		return dc.errorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"department_id": departmentID,
			"message":       "Department created successfully",
		},
	})
}

// @Summary Get department by ID
// @Description Get a single department by its ID
// @Tags departments
// @Accept json
// @Produce json
// @Param id path int true "Department ID"
// @Success 200 {object} map[string]interface{} "Department details"
// @Failure 400 {object} map[string]interface{} "Invalid department ID"
// @Failure 404 {object} map[string]interface{} "Department not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /departments/{id} [get]
func (dc *DepartmentController) GetDepartment(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return dc.errorResponse(c, http.StatusBadRequest, "Invalid department ID")
	}

	department, err := dc.service.GetDepartmentByID(id)
	if err != nil {
		if err.Error() == "department not found" {
			return dc.errorResponse(c, http.StatusNotFound, "Department not found")
		}
		return dc.errorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return dc.successResponse(c, department)
}

// @Summary Get all departments
// @Description Get all departments with pagination and filtering
// @Tags departments
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search term"
// @Param is_active query boolean false "Filter by active status"
// @Param parent_id query int false "Filter by parent department ID"
// @Success 200 {object} map[string]interface{} "List of departments"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /departments [get]
func (dc *DepartmentController) GetAllDepartments(c echo.Context) error {
	// Parse query parameters
	filter := &models.DepartmentFilter{}

	// Parse page
	if pageStr := c.QueryParam("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filter.Page = page
		} else {
			filter.Page = 1
		}
	} else {
		filter.Page = 1
	}

	// Parse limit
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			filter.Limit = limit
		} else {
			filter.Limit = 10
		}
	} else {
		filter.Limit = 10
	}

	// Parse search
	filter.Search = c.QueryParam("search")

	// Parse is_active
	if isActiveStr := c.QueryParam("is_active"); isActiveStr != "" {
		if isActive, err := strconv.ParseBool(isActiveStr); err == nil {
			filter.IsActive = &isActive
		}
	}

	// Parse parent_id
	if parentIDStr := c.QueryParam("parent_id"); parentIDStr != "" {
		if parentIDStr == "null" {
			filter.ParentID = nil
		} else if parentID, err := strconv.Atoi(parentIDStr); err == nil {
			filter.ParentID = &parentID
		}
	}

	response, err := dc.service.GetAllDepartments(filter)
	if err != nil {
		return dc.errorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return dc.successResponse(c, response)
}

// @Summary Update department
// @Description Update an existing department
// @Tags departments
// @Accept json
// @Produce json
// @Param id path int true "Department ID"
// @Param request body models.UpdateDepartmentRequest true "Department update request"
// @Success 200 {object} map[string]interface{} "Department updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 404 {object} map[string]interface{} "Department not found"
// @Failure 409 {object} map[string]interface{} "Department code already exists"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /departments/{id} [put]
func (dc *DepartmentController) UpdateDepartment(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return dc.errorResponse(c, http.StatusBadRequest, "Invalid department ID")
	}

	var req models.UpdateDepartmentRequest
	if err := c.Bind(&req); err != nil {
		return dc.errorResponse(c, http.StatusBadRequest, "Invalid request format")
	}

	// Validate request
	if err := c.Validate(&req); err != nil {
		return dc.errorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
	}

	err = dc.service.UpdateDepartment(id, &req, "system")
	if err != nil {
		switch err.Error() {
		case "department not found":
			return dc.errorResponse(c, http.StatusNotFound, err.Error())
		case "department code already exists":
			return dc.errorResponse(c, http.StatusConflict, err.Error())
		case "department cannot be its own parent":
			return dc.errorResponse(c, http.StatusBadRequest, err.Error())
		default:
			return dc.errorResponse(c, http.StatusInternalServerError, err.Error())
		}
	}

	return dc.successResponse(c, map[string]string{
		"message": "Department updated successfully",
	})
}

// @Summary Delete department
// @Description Soft delete a department (set is_active to false)
// @Tags departments
// @Accept json
// @Produce json
// @Param id path int true "Department ID"
// @Success 200 {object} map[string]interface{} "Department deleted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid department ID"
// @Failure 404 {object} map[string]interface{} "Department not found"
// @Failure 409 {object} map[string]interface{} "Cannot delete department with active children"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /departments/{id} [delete]
func (dc *DepartmentController) DeleteDepartment(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return dc.errorResponse(c, http.StatusBadRequest, "Invalid department ID")
	}

	err = dc.service.DeleteDepartment(id, "system")
	if err != nil {
		switch err.Error() {
		case "department not found":
			return dc.errorResponse(c, http.StatusNotFound, err.Error())
		case "cannot delete department with active child departments":
			return dc.errorResponse(c, http.StatusConflict, err.Error())
		default:
			return dc.errorResponse(c, http.StatusInternalServerError, err.Error())
		}
	}

	return dc.successResponse(c, map[string]string{
		"message": "Department deleted successfully",
	})
}

// @Summary Get department hierarchy
// @Description Get the complete department hierarchy tree
// @Tags departments
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Department hierarchy"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /departments/hierarchy [get]
func (dc *DepartmentController) GetDepartmentHierarchy(c echo.Context) error {
	hierarchy, err := dc.service.GetDepartmentHierarchy()
	if err != nil {
		return dc.errorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return dc.successResponse(c, hierarchy)
}

// @Summary Get users by department
// @Description Get all users belonging to a specific department
// @Tags departments
// @Accept json
// @Produce json
// @Param id path int true "Department ID"
// @Success 200 {object} map[string]interface{} "List of users in department"
// @Failure 400 {object} map[string]interface{} "Invalid department ID"
// @Failure 404 {object} map[string]interface{} "Department not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /departments/{id}/users [get]
func (dc *DepartmentController) GetUsersByDepartment(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return dc.errorResponse(c, http.StatusBadRequest, "Invalid department ID")
	}

	users, err := dc.service.GetUsersByDepartment(id)
	if err != nil {
		if err.Error() == "department not found" {
			return dc.errorResponse(c, http.StatusNotFound, err.Error())
		}
		return dc.errorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return dc.successResponse(c, users)
}

// @Summary Search departments
// @Description Search departments by name, code, or description
// @Tags departments
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Success 200 {object} map[string]interface{} "Search results"
// @Failure 400 {object} map[string]interface{} "Search query is required"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /departments/search [get]
func (dc *DepartmentController) SearchDepartments(c echo.Context) error {
	query := c.QueryParam("q")
	if query == "" {
		return dc.errorResponse(c, http.StatusBadRequest, "Search query is required")
	}

	departments, err := dc.service.SearchDepartments(query)
	if err != nil {
		return dc.errorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return dc.successResponse(c, departments)
}

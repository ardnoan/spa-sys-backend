package services

import (
	"fmt"
	"v01_system_backend/models"
	"v01_system_backend/repositories"
)

type DepartmentService interface {
	CreateDepartment(req *models.CreateDepartmentRequest, createdBy string) (int, error)
	GetDepartmentByID(id int) (*models.Department, error)
	GetAllDepartments(filter *models.DepartmentFilter) (*models.DepartmentListResponse, error)
	UpdateDepartment(id int, req *models.UpdateDepartmentRequest, updatedBy string) error
	DeleteDepartment(id int, deletedBy string) error
	GetDepartmentHierarchy() ([]models.DepartmentHierarchy, error)
	GetUsersByDepartment(departmentID int) ([]models.User, error)
	SearchDepartments(query string) ([]models.Department, error)
}

type departmentService struct {
	repo repositories.DepartmentRepository
}

func NewDepartmentService(repo repositories.DepartmentRepository) DepartmentService {
	return &departmentService{repo: repo}
}

func (s *departmentService) CreateDepartment(req *models.CreateDepartmentRequest, createdBy string) (int, error) {
	// Validate business rules
	if err := s.validateCreateRequest(req); err != nil {
		return 0, err
	}

	// Check if department code already exists
	exists, err := s.repo.ExistsByCode(req.DepartmentCode)
	if err != nil {
		return 0, fmt.Errorf("failed to check department code: %w", err)
	}
	if exists {
		return 0, fmt.Errorf("department code already exists")
	}

	// Validate parent department exists if provided
	if req.ParentID != nil {
		parentExists, err := s.repo.ExistsByID(*req.ParentID)
		if err != nil {
			return 0, fmt.Errorf("failed to validate parent department: %w", err)
		}
		if !parentExists {
			return 0, fmt.Errorf("parent department does not exist")
		}
	}

	return s.repo.Create(req, createdBy)
}

func (s *departmentService) GetDepartmentByID(id int) (*models.Department, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid department ID")
	}

	return s.repo.GetByID(id)
}

func (s *departmentService) GetAllDepartments(filter *models.DepartmentFilter) (*models.DepartmentListResponse, error) {
	// Set default values
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	departments, totalCount, err := s.repo.GetAll(filter)
	if err != nil {
		return nil, err
	}

	// Calculate pagination info
	totalPages := (totalCount + filter.Limit - 1) / filter.Limit

	response := &models.DepartmentListResponse{
		Departments: departments,
		Pagination: models.PaginationResponse{
			CurrentPage: filter.Page,
			PerPage:     filter.Limit,
			TotalCount:  totalCount,
			TotalPages:  totalPages,
		},
	}

	return response, nil
}

func (s *departmentService) UpdateDepartment(id int, req *models.UpdateDepartmentRequest, updatedBy string) error {
	if id <= 0 {
		return fmt.Errorf("invalid department ID")
	}

	// Validate business rules
	if err := s.validateUpdateRequest(req); err != nil {
		return err
	}

	// Check if department exists
	exists, err := s.repo.ExistsByID(id)
	if err != nil {
		return fmt.Errorf("failed to check department existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("department not found")
	}

	// Check if department code already exists (excluding current department)
	codeExists, err := s.repo.ExistsByCodeExcludeID(req.DepartmentCode, id)
	if err != nil {
		return fmt.Errorf("failed to check department code: %w", err)
	}
	if codeExists {
		return fmt.Errorf("department code already exists")
	}

	// Prevent setting parent_id to self
	if req.ParentID != nil && *req.ParentID == id {
		return fmt.Errorf("department cannot be its own parent")
	}

	// Validate parent department exists if provided
	if req.ParentID != nil {
		parentExists, err := s.repo.ExistsByID(*req.ParentID)
		if err != nil {
			return fmt.Errorf("failed to validate parent department: %w", err)
		}
		if !parentExists {
			return fmt.Errorf("parent department does not exist")
		}
	}

	return s.repo.Update(id, req, updatedBy)
}

func (s *departmentService) DeleteDepartment(id int, deletedBy string) error {
	if id <= 0 {
		return fmt.Errorf("invalid department ID")
	}

	// Check if department exists
	exists, err := s.repo.ExistsByID(id)
	if err != nil {
		return fmt.Errorf("failed to check department existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("department not found")
	}

	// Check if department has child departments
	hasChildren, err := s.repo.HasActiveChildren(id)
	if err != nil {
		return fmt.Errorf("failed to check child departments: %w", err)
	}
	if hasChildren {
		return fmt.Errorf("cannot delete department with active child departments")
	}

	return s.repo.Delete(id, deletedBy)
}

func (s *departmentService) GetDepartmentHierarchy() ([]models.DepartmentHierarchy, error) {
	return s.repo.GetHierarchy()
}

func (s *departmentService) GetUsersByDepartment(departmentID int) ([]models.User, error) {
	if departmentID <= 0 {
		return nil, fmt.Errorf("invalid department ID")
	}

	// Check if department exists
	exists, err := s.repo.ExistsByID(departmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to check department existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("department not found")
	}

	return s.repo.GetUsersByDepartment(departmentID)
}

func (s *departmentService) SearchDepartments(query string) ([]models.Department, error) {
	if query == "" {
		return nil, fmt.Errorf("search query is required")
	}

	return s.repo.Search(query)
}

// Validation methods
func (s *departmentService) validateCreateRequest(req *models.CreateDepartmentRequest) error {
	if req.DepartmentName == "" {
		return fmt.Errorf("department name is required")
	}
	if req.DepartmentCode == "" {
		return fmt.Errorf("department code is required")
	}
	if len(req.DepartmentName) < 3 {
		return fmt.Errorf("department name must be at least 3 characters")
	}
	if len(req.DepartmentCode) < 2 {
		return fmt.Errorf("department code must be at least 2 characters")
	}
	return nil
}

func (s *departmentService) validateUpdateRequest(req *models.UpdateDepartmentRequest) error {
	if req.DepartmentName == "" {
		return fmt.Errorf("department name is required")
	}
	if req.DepartmentCode == "" {
		return fmt.Errorf("department code is required")
	}
	if len(req.DepartmentName) < 3 {
		return fmt.Errorf("department name must be at least 3 characters")
	}
	if len(req.DepartmentCode) < 2 {
		return fmt.Errorf("department code must be at least 2 characters")
	}
	return nil
}

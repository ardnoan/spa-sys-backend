package services

import (
	"errors"
	"v01_system_backend/apps/models"
	"v01_system_backend/apps/repositories"
)

type MenuService struct {
	menuRepo *repositories.MenuRepository
}

func NewMenuService(menuRepo *repositories.MenuRepository) *MenuService {
	return &MenuService{menuRepo: menuRepo}
}

func (s *MenuService) GetAll() ([]*models.Menu, error) {
	return s.menuRepo.GetAll()
}

func (s *MenuService) GetByID(id int) (*models.Menu, error) {
	menu, err := s.menuRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if menu == nil {
		return nil, errors.New("menu not found")
	}

	return menu, nil
}

func (s *MenuService) GetMenuTree() ([]*models.Menu, error) {
	return s.menuRepo.GetMenuTree()
}

func (s *MenuService) GetMenusByRole(roleID int) ([]*models.Menu, error) {
	return s.menuRepo.GetMenusByRole(roleID)
}

func (s *MenuService) GetUserMenus(userID int) ([]*models.Menu, error) {
	return s.menuRepo.GetUserMenus(userID)
}

func (s *MenuService) Create(req *models.MenuCreateRequest, createdBy int) (*models.Menu, error) {
	// Check if menu code already exists
	existingMenu, err := s.menuRepo.GetByCode(req.MenuCode)
	if err != nil {
		return nil, err
	}
	if existingMenu != nil {
		return nil, errors.New("menu code already exists")
	}

	// Validate parent menu if provided
	if req.ParentID != nil {
		parentMenu, err := s.menuRepo.GetByID(*req.ParentID)
		if err != nil {
			return nil, err
		}
		if parentMenu == nil {
			return nil, errors.New("parent menu not found")
		}
	}

	return s.menuRepo.Create(req, createdBy)
}

func (s *MenuService) Update(id int, req *models.MenuUpdateRequest, updatedBy int) (*models.Menu, error) {
	// Check if menu exists
	existingMenu, err := s.menuRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existingMenu == nil {
		return nil, errors.New("menu not found")
	}

	// Check if menu code already exists (excluding current menu)
	menuByCode, err := s.menuRepo.GetByCode(req.MenuCode)
	if err != nil {
		return nil, err
	}
	if menuByCode != nil && menuByCode.MenusID != id {
		return nil, errors.New("menu code already exists")
	}

	// Validate parent menu if provided
	if req.ParentID != nil {
		if *req.ParentID == id {
			return nil, errors.New("menu cannot be parent of itself")
		}
		parentMenu, err := s.menuRepo.GetByID(*req.ParentID)
		if err != nil {
			return nil, err
		}
		if parentMenu == nil {
			return nil, errors.New("parent menu not found")
		}

		// Check for circular reference
		if s.hasCircularReference(id, *req.ParentID) {
			return nil, errors.New("circular reference detected")
		}
	}

	return s.menuRepo.Update(id, req, updatedBy)
}

func (s *MenuService) Delete(id int, deletedBy int) error {
	// Check if menu exists
	menu, err := s.menuRepo.GetByID(id)
	if err != nil {
		return err
	}
	if menu == nil {
		return errors.New("menu not found")
	}

	// Check if menu has children
	children, err := s.menuRepo.GetChildren(id)
	if err != nil {
		return err
	}
	if len(children) > 0 {
		return errors.New("cannot delete menu with children")
	}

	return s.menuRepo.Delete(id, deletedBy)
}

func (s *MenuService) UpdateOrder(id int, newOrder int, updatedBy int) error {
	// Check if menu exists
	menu, err := s.menuRepo.GetByID(id)
	if err != nil {
		return err
	}
	if menu == nil {
		return errors.New("menu not found")
	}

	return s.menuRepo.UpdateOrder(id, newOrder, updatedBy)
}

func (s *MenuService) hasCircularReference(menuID, parentID int) bool {
	// Simple circular reference check
	// In production, you might want a more sophisticated check
	current := parentID
	visited := make(map[int]bool)

	for current != 0 {
		if visited[current] {
			return true
		}
		if current == menuID {
			return true
		}

		visited[current] = true
		parent, err := s.menuRepo.GetByID(current)
		if err != nil || parent == nil || parent.ParentID == nil {
			break
		}
		current = *parent.ParentID
	}

	return false
}

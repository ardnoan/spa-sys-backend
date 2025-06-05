package repositories

import (
	"database/sql"
	"fmt"
	"strings"
	"v01_system_backend/apps/models"

	"github.com/jmoiron/sqlx"
)

type MenuRepository struct {
	db *sqlx.DB
}

func NewMenuRepository(db *sqlx.DB) *MenuRepository {
	return &MenuRepository{db: db}
}

func (r *MenuRepository) GetAll(pagination *models.PaginationRequest) ([]models.Menu, int, error) {
	var menus []models.Menu
	var totalRows int

	// Build WHERE clause
	whereClause := "WHERE is_active = true"
	args := []interface{}{}
	argIndex := 1

	if pagination.Search != "" {
		whereClause += fmt.Sprintf(" AND (menu_name ILIKE $%d OR menu_code ILIKE $%d)",
			argIndex, argIndex+1)
		searchPattern := "%" + pagination.Search + "%"
		args = append(args, searchPattern, searchPattern)
		argIndex += 2
	}

	// Count total rows
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM menus %s", whereClause)
	if err := r.db.Get(&totalRows, countQuery, args...); err != nil {
		return nil, 0, err
	}

	// Build ORDER BY clause
	orderBy := "ORDER BY menu_order ASC, menu_name ASC"
	if pagination.SortBy != "" {
		validSortFields := map[string]string{
			"menu_name":  "menu_name",
			"menu_code":  "menu_code",
			"menu_order": "menu_order",
			"created_at": "created_at",
		}
		if field, exists := validSortFields[pagination.SortBy]; exists {
			orderBy = fmt.Sprintf("ORDER BY %s %s", field, strings.ToUpper(pagination.SortDir))
		}
	}

	// Main query
	query := fmt.Sprintf(`
		SELECT menus_id, menu_code, menu_name, parent_id, icon_name, route,
			   menu_order, is_visible, is_active, created_at, created_by,
			   updated_at, updated_by
		FROM menus %s %s
		LIMIT $%d OFFSET $%d`, whereClause, orderBy, argIndex, argIndex+1)

	args = append(args, pagination.GetLimit(), pagination.GetOffset())

	if err := r.db.Select(&menus, query, args...); err != nil {
		return nil, 0, err
	}

	return menus, totalRows, nil
}

func (r *MenuRepository) GetByID(id int) (*models.Menu, error) {
	var menu models.Menu
	query := `
		SELECT menus_id, menu_code, menu_name, parent_id, icon_name, route,
			   menu_order, is_visible, is_active, created_at, created_by,
			   updated_at, updated_by
		FROM menus WHERE menus_id = $1 AND is_active = true`

	if err := r.db.Get(&menu, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &menu, nil
}

func (r *MenuRepository) GetMenuTree() ([]models.Menu, error) {
	var menus []models.Menu
	query := `
		SELECT menus_id, menu_code, menu_name, parent_id, icon_name, route,
			   menu_order, is_visible, is_active
		FROM menus 
		WHERE is_active = true 
		ORDER BY parent_id NULLS FIRST, menu_order ASC, menu_name ASC`

	if err := r.db.Select(&menus, query); err != nil {
		return nil, err
	}

	return r.buildMenuTree(menus, nil), nil
}

func (r *MenuRepository) GetUserMenus(userID int) ([]models.Menu, error) {
	var menus []models.Menu
	query := `
		SELECT DISTINCT m.menus_id, m.menu_code, m.menu_name, m.parent_id, 
			   m.icon_name, m.route, m.menu_order, m.is_visible,
			   rm.can_view, rm.can_create, rm.can_modify, rm.can_delete,
			   rm.can_upload, rm.can_download
		FROM menus m
		INNER JOIN role_menus rm ON m.menus_id = rm.menu_id
		INNER JOIN user_roles ur ON rm.role_id = ur.role_id
		WHERE ur.user_id = $1 AND ur.is_active = true 
			  AND m.is_active = true AND m.is_visible = true
			  AND rm.can_view = true
		ORDER BY m.parent_id NULLS FIRST, m.menu_order ASC, m.menu_name ASC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var menu models.Menu
		var access models.MenuAccess

		err := rows.Scan(
			&menu.MenusID, &menu.MenuCode, &menu.MenuName, &menu.ParentID,
			&menu.IconName, &menu.Route, &menu.MenuOrder, &menu.IsVisible,
			&access.CanView, &access.CanCreate, &access.CanModify, &access.CanDelete,
			&access.CanUpload, &access.CanDownload,
		)
		if err != nil {
			return nil, err
		}

		menu.Access = &access
		menus = append(menus, menu)
	}

	return r.buildMenuTree(menus, nil), nil
}

func (r *MenuRepository) Create(menu *models.Menu, createdBy int) (*models.Menu, error) {
	var menuID int
	query := `
		INSERT INTO menus (menu_code, menu_name, parent_id, icon_name, route, 
						  menu_order, is_visible, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING menus_id`

	if err := r.db.Get(&menuID, query,
		menu.MenuCode, menu.MenuName, menu.ParentID, menu.IconName,
		menu.Route, menu.MenuOrder, menu.IsVisible, createdBy); err != nil {
		return nil, err
	}

	return r.GetByID(menuID)
}

func (r *MenuRepository) Update(id int, menu *models.Menu, updatedBy int) (*models.Menu, error) {
	query := `
		UPDATE menus SET 
			menu_code = $1, menu_name = $2, parent_id = $3, icon_name = $4,
			route = $5, menu_order = $6, is_visible = $7, updated_by = $8,
			updated_at = CURRENT_TIMESTAMP
		WHERE menus_id = $9 AND is_active = true`

	result, err := r.db.Exec(query,
		menu.MenuCode, menu.MenuName, menu.ParentID, menu.IconName,
		menu.Route, menu.MenuOrder, menu.IsVisible, updatedBy, id)
	if err != nil {
		return nil, err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, sql.ErrNoRows
	}

	return r.GetByID(id)
}

func (r *MenuRepository) Delete(id int, deletedBy int) error {
	query := `
		UPDATE menus SET 
			is_active = false, updated_by = $1, updated_at = CURRENT_TIMESTAMP
		WHERE menus_id = $2 AND is_active = true`

	result, err := r.db.Exec(query, deletedBy, id)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *MenuRepository) buildMenuTree(menus []models.Menu, parentID *int) []models.Menu {
	var result []models.Menu

	for _, menu := range menus {
		if (parentID == nil && menu.ParentID == nil) ||
			(parentID != nil && menu.ParentID != nil && *menu.ParentID == *parentID) {
			menu.Children = r.buildMenuTree(menus, &menu.MenusID)
			result = append(result, menu)
		}
	}

	return result
}

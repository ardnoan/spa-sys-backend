// controllers/email_templates.go
package controller

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

type EmailTemplatesController struct {
	DB *sql.DB
}

type EmailTemplate struct {
	TemplateID   int             `json:"template_id"`
	TemplateCode string          `json:"template_code"`
	TemplateName string          `json:"template_name"`
	Subject      string          `json:"subject"`
	BodyHTML     *string         `json:"body_html"`
	BodyText     *string         `json:"body_text"`
	Variables    json.RawMessage `json:"variables"`
	IsActive     bool            `json:"is_active"`
	CreatedAt    string          `json:"created_at"`
	CreatedBy    *string         `json:"created_by"`
	UpdatedAt    string          `json:"updated_at"`
	UpdatedBy    *string         `json:"updated_by"`
}

func NewEmailTemplatesController(db *sql.DB) *EmailTemplatesController {
	return &EmailTemplatesController{DB: db}
}

func (c *EmailTemplatesController) GetAllEmailTemplates(ctx echo.Context) error {
	page, _ := strconv.Atoi(ctx.QueryParam("page"))
	limit, _ := strconv.Atoi(ctx.QueryParam("limit"))
	search := ctx.QueryParam("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit

	var countQuery string
	var query string
	var args []interface{}

	if search != "" {
		countQuery = `
			SELECT COUNT(*) 
			FROM email_templates
			WHERE template_code ILIKE $1 OR template_name ILIKE $1 OR subject ILIKE $1
		`

		query = `
			SELECT template_id, template_code, template_name, subject, body_html, body_text,
				   variables, is_active, created_at, created_by, updated_at, updated_by
			FROM email_templates
			WHERE template_code ILIKE $1 OR template_name ILIKE $1 OR subject ILIKE $1
			ORDER BY template_name
			LIMIT $2 OFFSET $3
		`
		searchParam := "%" + search + "%"
		args = []interface{}{searchParam, limit, offset}
	} else {
		countQuery = `SELECT COUNT(*) FROM email_templates`

		query = `
			SELECT template_id, template_code, template_name, subject, body_html, body_text,
				   variables, is_active, created_at, created_by, updated_at, updated_by
			FROM email_templates
			ORDER BY template_name
			LIMIT $1 OFFSET $2
		`
		args = []interface{}{limit, offset}
	}

	var totalRecords int
	if search != "" {
		err := c.DB.QueryRow(countQuery, "%"+search+"%").Scan(&totalRecords)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to count records"})
		}
	} else {
		err := c.DB.QueryRow(countQuery).Scan(&totalRecords)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to count records"})
		}
	}

	rows, err := c.DB.Query(query, args...)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch email templates"})
	}
	defer rows.Close()

	var templates []EmailTemplate
	for rows.Next() {
		var template EmailTemplate
		err := rows.Scan(
			&template.TemplateID,
			&template.TemplateCode,
			&template.TemplateName,
			&template.Subject,
			&template.BodyHTML,
			&template.BodyText,
			&template.Variables,
			&template.IsActive,
			&template.CreatedAt,
			&template.CreatedBy,
			&template.UpdatedAt,
			&template.UpdatedBy,
		)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to scan email template"})
		}
		templates = append(templates, template)
	}

	totalPages := (totalRecords + limit - 1) / limit

	response := map[string]interface{}{
		"data": templates,
		"pagination": map[string]interface{}{
			"current_page":     page,
			"total_pages":      totalPages,
			"total_records":    totalRecords,
			"records_per_page": limit,
		},
	}

	return ctx.JSON(http.StatusOK, response)
}

func (c *EmailTemplatesController) GetEmailTemplateByID(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid template ID"})
	}

	query := `
		SELECT template_id, template_code, template_name, subject, body_html, body_text,
			   variables, is_active, created_at, created_by, updated_at, updated_by
		FROM email_templates
		WHERE template_id = $1
	`

	var template EmailTemplate
	err = c.DB.QueryRow(query, id).Scan(
		&template.TemplateID,
		&template.TemplateCode,
		&template.TemplateName,
		&template.Subject,
		&template.BodyHTML,
		&template.BodyText,
		&template.Variables,
		&template.IsActive,
		&template.CreatedAt,
		&template.CreatedBy,
		&template.UpdatedAt,
		&template.UpdatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Template not found"})
		}
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch template"})
	}

	return ctx.JSON(http.StatusOK, template)
}

func (c *EmailTemplatesController) CreateEmailTemplate(ctx echo.Context) error {
	var template EmailTemplate
	if err := ctx.Bind(&template); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request data"})
	}

	// Validate required fields
	if template.TemplateCode == "" || template.TemplateName == "" || template.Subject == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Template code, name, and subject are required"})
	}

	query := `
		INSERT INTO email_templates (template_code, template_name, subject, body_html, body_text, variables, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
		RETURNING template_id
	`

	now := time.Now().Format(time.RFC3339)
	var templateID int
	err := c.DB.QueryRow(query,
		template.TemplateCode,
		template.TemplateName,
		template.Subject,
		template.BodyHTML,
		template.BodyText,
		template.Variables,
		template.IsActive,
		now,
	).Scan(&templateID)

	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create template"})
	}

	return ctx.JSON(http.StatusCreated, map[string]interface{}{
		"message":     "Template created successfully",
		"template_id": templateID,
		"success":     true,
	})
}

func (c *EmailTemplatesController) UpdateEmailTemplate(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid template ID"})
	}

	var template EmailTemplate
	if err := ctx.Bind(&template); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request data"})
	}

	// Validate required fields
	if template.TemplateCode == "" || template.TemplateName == "" || template.Subject == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Template code, name, and subject are required"})
	}

	query := `
		UPDATE email_templates 
		SET template_code = $1, template_name = $2, subject = $3, body_html = $4, 
			body_text = $5, variables = $6, is_active = $7, updated_at = $8
		WHERE template_id = $9
	`

	now := time.Now().Format(time.RFC3339)
	result, err := c.DB.Exec(query,
		template.TemplateCode,
		template.TemplateName,
		template.Subject,
		template.BodyHTML,
		template.BodyText,
		template.Variables,
		template.IsActive,
		now,
		id,
	)

	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update template"})
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Template not found"})
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"message": "Template updated successfully",
		"success": true,
	})
}

func (c *EmailTemplatesController) DeleteEmailTemplate(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid template ID"})
	}

	query := `DELETE FROM email_templates WHERE template_id = $1`
	result, err := c.DB.Exec(query, id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete template"})
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Template not found"})
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"message": "Template deleted successfully",
		"success": true,
	})
}

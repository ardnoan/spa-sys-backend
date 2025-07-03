package controller

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
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

func (c *EmailTemplatesController) GetAllEmailTemplates(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	search := ctx.Query("search")

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
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count records"})
			return
		}
	} else {
		err := c.DB.QueryRow(countQuery).Scan(&totalRecords)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count records"})
			return
		}
	}

	rows, err := c.DB.Query(query, args...)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch email templates"})
		return
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
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan email template"})
			return
		}
		templates = append(templates, template)
	}

	totalPages := (totalRecords + limit - 1) / limit

	response := gin.H{
		"data": templates,
		"pagination": gin.H{
			"current_page":     page,
			"total_pages":      totalPages,
			"total_records":    totalRecords,
			"records_per_page": limit,
		},
	}

	ctx.JSON(http.StatusOK, response)
}

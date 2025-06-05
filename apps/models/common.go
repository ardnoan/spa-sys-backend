package models

import (
	"time"
)

type BaseModel struct {
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	CreatedBy *int       `json:"created_by" db:"created_by"`
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at"`
	UpdatedBy *int       `json:"updated_by" db:"updated_by"`
}

type PaginationRequest struct {
	Page     int    `query:"page" validate:"min=1"`
	PageSize int    `query:"page_size" validate:"min=1,max=100"`
	Search   string `query:"search"`
	SortBy   string `query:"sort_by"`
	SortDir  string `query:"sort_dir" validate:"oneof=asc desc"`
}

type PaginationResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalRows  int         `json:"total_rows"`
	TotalPages int         `json:"total_pages"`
}

func (p *PaginationRequest) SetDefaults() {
	if p.Page == 0 {
		p.Page = 1
	}
	if p.PageSize == 0 {
		p.PageSize = 10
	}
	if p.SortDir == "" {
		p.SortDir = "asc"
	}
}

func (p *PaginationRequest) GetOffset() int {
	return (p.Page - 1) * p.PageSize
}

func (p *PaginationRequest) GetLimit() int {
	return p.PageSize
}

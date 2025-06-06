package models

import (
	"time"
)

type BaseModel struct {
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	CreatedBy *string    `json:"created_by" db:"created_by"`
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at"`
	UpdatedBy *string    `json:"updated_by" db:"updated_by"`
}

type PaginationRequest struct {
	Page     int    `json:"page" query:"page"`
	PageSize int    `json:"page_size" query:"page_size"`
	Search   string `json:"search" query:"search"`
	SortBy   string `json:"sort_by" query:"sort_by"`
	SortDir  string `json:"sort_dir" query:"sort_dir"`
}

func (p *PaginationRequest) SetDefaults() {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 10
	}
	if p.SortDir == "" {
		p.SortDir = "ASC"
	}
}

func (p *PaginationRequest) GetOffset() int {
	return (p.Page - 1) * p.PageSize
}

func (p *PaginationRequest) GetLimit() int {
	return p.PageSize
}

type PaginationResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalRows  int         `json:"total_rows"`
	TotalPages int         `json:"total_pages"`
}

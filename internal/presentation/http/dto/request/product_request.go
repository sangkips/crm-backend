package request

import "github.com/google/uuid"

// CreateProductRequest represents a product creation request
type CreateProductRequest struct {
	CategoryID    *uuid.UUID `json:"category_id"`
	UnitID        *uuid.UUID `json:"unit_id"`
	Name          string     `json:"name" binding:"required,min=2,max=255"`
	Code          string     `json:"code" binding:"omitempty,max=100"`
	Quantity      int        `json:"quantity" binding:"min=0"`
	QuantityAlert int        `json:"quantity_alert" binding:"min=0"`
	BuyingPrice   float64    `json:"buying_price" binding:"min=0"`
	SellingPrice  float64    `json:"selling_price" binding:"min=0"`
	Tax           int        `json:"tax" binding:"min=0,max=100"`
	TaxType       int        `json:"tax_type" binding:"min=0,max=1"`
	Notes         *string    `json:"notes"`
}

// UpdateProductRequest represents a product update request
type UpdateProductRequest struct {
	CategoryID    *uuid.UUID `json:"category_id"`
	UnitID        *uuid.UUID `json:"unit_id"`
	Name          *string    `json:"name" binding:"omitempty,min=2,max=255"`
	Code          *string    `json:"code" binding:"omitempty,min=1,max=100"`
	Quantity      *int       `json:"quantity" binding:"omitempty,min=0"`
	QuantityAlert *int       `json:"quantity_alert" binding:"omitempty,min=0"`
	BuyingPrice   *float64   `json:"buying_price" binding:"omitempty,min=0"`
	SellingPrice  *float64   `json:"selling_price" binding:"omitempty,min=0"`
	Tax           *int       `json:"tax" binding:"omitempty,min=0,max=100"`
	TaxType       *int       `json:"tax_type" binding:"omitempty,min=0,max=1"`
	Notes         *string    `json:"notes"`
}

// ProductFilterRequest represents product filter parameters
type ProductFilterRequest struct {
	Search     string `form:"search"`
	CategoryID string `form:"category_id"`
	UnitID     string `form:"unit_id"`
	LowStock   bool   `form:"low_stock"`
	SortBy     string `form:"sort_by"`
	SortOrder  string `form:"sort_order"`
	Page       int    `form:"page"`
	PerPage    int    `form:"per_page"`
	Limit      int    `form:"limit"` // For cursor-based pagination
}

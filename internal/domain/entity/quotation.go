package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/domain/enum"
	"gorm.io/gorm"
)

// Quotation represents a price quotation for a customer
type Quotation struct {
	ID                 uuid.UUID            `gorm:"type:uuid;primary_key" json:"id"`
	UserID             uuid.UUID            `gorm:"type:uuid;not null;index" json:"user_id"`
	CustomerID         *uuid.UUID           `gorm:"type:uuid;index" json:"customer_id,omitempty"`
	Date               time.Time            `gorm:"type:date;not null" json:"date"`
	Reference          string               `gorm:"size:100;unique;not null" json:"reference"`
	CustomerName       string               `gorm:"size:255" json:"customer_name"`
	TaxPercentage      float64              `gorm:"type:decimal(5,2);default:0" json:"tax_percentage"`
	TaxAmount          float64              `gorm:"type:decimal(15,2);default:0" json:"tax_amount"`
	DiscountPercentage float64              `gorm:"type:decimal(5,2);default:0" json:"discount_percentage"`
	DiscountAmount     float64              `gorm:"type:decimal(15,2);default:0" json:"discount_amount"`
	ShippingAmount     float64              `gorm:"type:decimal(15,2);default:0" json:"shipping_amount"`
	TotalAmount        float64              `gorm:"type:decimal(15,2);default:0" json:"total_amount"`
	Status             enum.QuotationStatus `gorm:"default:0" json:"status"`
	Note               *string              `gorm:"type:text" json:"note,omitempty"`
	CreatedAt          time.Time            `json:"created_at"`
	UpdatedAt          time.Time            `json:"updated_at"`
	DeletedAt          gorm.DeletedAt       `gorm:"index" json:"-"`

	// Relationships
	User     User              `gorm:"foreignKey:UserID" json:"-"`
	Customer *Customer         `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	Details  []QuotationDetail `gorm:"foreignKey:QuotationID" json:"details,omitempty"`
}

// BeforeCreate generates a UUID and reference before creating a new quotation
func (q *Quotation) BeforeCreate(tx *gorm.DB) error {
	if q.ID == uuid.Nil {
		q.ID = uuid.New()
	}
	return nil
}

// TableName returns the table name for the Quotation model
func (Quotation) TableName() string {
	return "quotations"
}

// QuotationDetail represents a line item in a quotation
type QuotationDetail struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	QuotationID uuid.UUID      `gorm:"type:uuid;not null;index" json:"quotation_id"`
	ProductID   uuid.UUID      `gorm:"type:uuid;not null;index" json:"product_id"`
	ProductName string         `gorm:"size:255" json:"product_name"`
	ProductCode string         `gorm:"size:100" json:"product_code"`
	Quantity    int            `gorm:"not null" json:"quantity"`
	UnitPrice   float64        `gorm:"type:decimal(15,2);not null" json:"unit_price"`
	SubTotal    float64        `gorm:"type:decimal(15,2);not null" json:"sub_total"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Quotation Quotation `gorm:"foreignKey:QuotationID" json:"-"`
	Product   Product   `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

// BeforeCreate generates a UUID before creating a new quotation detail
func (qd *QuotationDetail) BeforeCreate(tx *gorm.DB) error {
	if qd.ID == uuid.Nil {
		qd.ID = uuid.New()
	}
	return nil
}

// TableName returns the table name for the QuotationDetail model
func (QuotationDetail) TableName() string {
	return "quotation_details"
}

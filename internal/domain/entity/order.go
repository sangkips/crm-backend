package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/domain/enum"
	"gorm.io/gorm"
)

// Order represents a sales order
type Order struct {
	ID            uuid.UUID        `gorm:"type:uuid;primary_key" json:"id"`
	UserID        uuid.UUID        `gorm:"type:uuid;not null;index" json:"user_id"`
	CustomerID    *uuid.UUID       `gorm:"type:uuid;index" json:"customer_id,omitempty"`
	OrderDate     time.Time        `gorm:"type:date;not null" json:"order_date"`
	OrderStatus   enum.OrderStatus `gorm:"default:0" json:"order_status"`
	TotalProducts int              `gorm:"default:0" json:"total_products"`
	SubTotal      int64            `gorm:"default:0" json:"-"` // Stored in cents, excluded from JSON
	VAT           int64            `gorm:"default:0" json:"-"` // Stored in cents, excluded from JSON
	Total         int64            `gorm:"default:0" json:"-"` // Stored in cents, excluded from JSON
	InvoiceNo     string           `gorm:"size:100;unique;not null" json:"invoice_no"`
	PaymentType   string           `gorm:"size:50" json:"payment_type"`
	Pay           int64            `gorm:"default:0" json:"-"` // Stored in cents, excluded from JSON
	Due           int64            `gorm:"default:0" json:"-"` // Stored in cents, excluded from JSON
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
	DeletedAt     gorm.DeletedAt   `gorm:"index" json:"-"`

	// Relationships
	User     User          `gorm:"foreignKey:UserID" json:"-"`
	Customer *Customer     `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	Details  []OrderDetail `gorm:"foreignKey:OrderID" json:"details,omitempty"`
}

// MarshalJSON custom marshaler to convert cents to decimal for API responses
func (o Order) MarshalJSON() ([]byte, error) {
	type Alias Order
	return json.Marshal(&struct {
		Alias
		SubTotal float64 `json:"sub_total"`
		VAT      float64 `json:"vat"`
		Total    float64 `json:"total"`
		Pay      float64 `json:"pay"`
		Due      float64 `json:"due"`
	}{
		Alias:    Alias(o),
		SubTotal: float64(o.SubTotal) / 100,
		VAT:      float64(o.VAT) / 100,
		Total:    float64(o.Total) / 100,
		Pay:      float64(o.Pay) / 100,
		Due:      float64(o.Due) / 100,
	})
}

// BeforeCreate generates a UUID before creating a new order
func (o *Order) BeforeCreate(tx *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return nil
}

// TableName returns the table name for the Order model
func (Order) TableName() string {
	return "orders"
}

// GetTotalDecimal returns the total as a decimal
func (o *Order) GetTotalDecimal() float64 {
	return float64(o.Total) / 100
}

// GetSubTotalDecimal returns the subtotal as a decimal
func (o *Order) GetSubTotalDecimal() float64 {
	return float64(o.SubTotal) / 100
}

// OrderDetail represents a line item in an order
type OrderDetail struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	OrderID   uuid.UUID      `gorm:"type:uuid;not null;index" json:"order_id"`
	ProductID uuid.UUID      `gorm:"type:uuid;not null;index" json:"product_id"`
	Quantity  int            `gorm:"not null" json:"quantity"`
	UnitCost  int64          `gorm:"not null" json:"-"` // Stored in cents, excluded from JSON
	Total     int64          `gorm:"not null" json:"-"` // Stored in cents, excluded from JSON
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Order   Order   `gorm:"foreignKey:OrderID" json:"-"`
	Product Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

// MarshalJSON custom marshaler to convert cents to decimal for API responses
func (od OrderDetail) MarshalJSON() ([]byte, error) {
	type Alias OrderDetail
	return json.Marshal(&struct {
		Alias
		UnitCost float64 `json:"unit_cost"`
		Total    float64 `json:"total"`
	}{
		Alias:    Alias(od),
		UnitCost: float64(od.UnitCost) / 100,
		Total:    float64(od.Total) / 100,
	})
}

// BeforeCreate generates a UUID before creating a new order detail
func (od *OrderDetail) BeforeCreate(tx *gorm.DB) error {
	if od.ID == uuid.Nil {
		od.ID = uuid.New()
	}
	return nil
}

// TableName returns the table name for the OrderDetail model
func (OrderDetail) TableName() string {
	return "order_details"
}

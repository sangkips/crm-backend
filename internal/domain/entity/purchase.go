package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/domain/enum"
	"gorm.io/gorm"
)

// Purchase represents a purchase from a supplier
type Purchase struct {
	ID            uuid.UUID           `gorm:"type:uuid;primary_key" json:"id"`
	UserID        uuid.UUID           `gorm:"type:uuid;not null;index" json:"user_id"`
	SupplierID    *uuid.UUID          `gorm:"type:uuid;index" json:"supplier_id,omitempty"`
	CreatedByID   *uuid.UUID          `gorm:"type:uuid;column:created_by" json:"created_by,omitempty"`
	UpdatedByID   *uuid.UUID          `gorm:"type:uuid;column:updated_by" json:"updated_by,omitempty"`
	Date          time.Time           `gorm:"type:date;not null" json:"date"`
	PurchaseNo    string              `gorm:"size:100;unique;not null" json:"purchase_no"`
	Status        enum.PurchaseStatus `gorm:"default:0" json:"status"`
	TotalAmount   float64             `gorm:"type:decimal(15,2);default:0" json:"total_amount"`
	TaxPercentage float64             `gorm:"type:decimal(5,2);default:0" json:"tax_percentage"`
	TaxAmount     float64             `gorm:"type:decimal(15,2);default:0" json:"tax_amount"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
	DeletedAt     gorm.DeletedAt      `gorm:"index" json:"-"`

	// Relationships
	User      User             `gorm:"foreignKey:UserID" json:"-"`
	Supplier  *Supplier        `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`
	CreatedBy *User            `gorm:"foreignKey:CreatedByID" json:"created_by_user,omitempty"`
	UpdatedBy *User            `gorm:"foreignKey:UpdatedByID" json:"updated_by_user,omitempty"`
	Details   []PurchaseDetail `gorm:"foreignKey:PurchaseID" json:"details,omitempty"`
}

// BeforeCreate generates a UUID before creating a new purchase
func (p *Purchase) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// TableName returns the table name for the Purchase model
func (Purchase) TableName() string {
	return "purchases"
}

// PurchaseDetail represents a line item in a purchase
type PurchaseDetail struct {
	ID         uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	PurchaseID uuid.UUID      `gorm:"type:uuid;not null;index" json:"purchase_id"`
	ProductID  uuid.UUID      `gorm:"type:uuid;not null;index" json:"product_id"`
	Quantity   int            `gorm:"not null" json:"quantity"`
	UnitCost   int64          `gorm:"not null" json:"unit_cost"` // Stored in cents
	Total      int64          `gorm:"not null" json:"total"`     // Stored in cents
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Purchase Purchase `gorm:"foreignKey:PurchaseID" json:"-"`
	Product  Product  `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

// BeforeCreate generates a UUID before creating a new purchase detail
func (pd *PurchaseDetail) BeforeCreate(tx *gorm.DB) error {
	if pd.ID == uuid.Nil {
		pd.ID = uuid.New()
	}
	return nil
}

// TableName returns the table name for the PurchaseDetail model
func (PurchaseDetail) TableName() string {
	return "purchase_details"
}

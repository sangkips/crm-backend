package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/domain/enum"
	"gorm.io/gorm"
)

// Product represents a product in the inventory
type Product struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	TenantID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"tenant_id"`
	UserID        uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	CategoryID    *uuid.UUID     `gorm:"type:uuid;index" json:"category_id,omitempty"`
	UnitID        *uuid.UUID     `gorm:"type:uuid;index" json:"unit_id,omitempty"`
	Name          string         `gorm:"size:255;not null" json:"name"`
	Slug          string         `gorm:"size:255;unique;not null" json:"slug"`
	Code          string         `gorm:"size:100;unique;not null" json:"code"`
	Quantity      int            `gorm:"default:0" json:"quantity"`
	QuantityAlert int            `gorm:"default:0" json:"quantity_alert"`
	BuyingPrice   int64          `gorm:"default:0" json:"buying_price"`  // Stored in cents
	SellingPrice  int64          `gorm:"default:0" json:"selling_price"` // Stored in cents
	Tax           int            `gorm:"default:0" json:"tax"`
	TaxType       enum.TaxType   `gorm:"default:0" json:"tax_type"`
	Notes         *string        `gorm:"type:text" json:"notes,omitempty"`
	ProductImage  *string        `gorm:"size:255" json:"product_image,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Tenant   Tenant    `gorm:"foreignKey:TenantID" json:"-"`
	User     User      `gorm:"foreignKey:UserID" json:"-"`
	Category *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Unit     *Unit     `gorm:"foreignKey:UnitID" json:"unit,omitempty"`
}

// BeforeCreate generates a UUID before creating a new product
func (p *Product) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// TableName returns the table name for the Product model
func (Product) TableName() string {
	return "products"
}

// GetBuyingPriceDecimal returns the buying price as a decimal (for display)
func (p *Product) GetBuyingPriceDecimal() float64 {
	return float64(p.BuyingPrice) / 100
}

// GetSellingPriceDecimal returns the selling price as a decimal (for display)
func (p *Product) GetSellingPriceDecimal() float64 {
	return float64(p.SellingPrice) / 100
}

// SetBuyingPriceFromDecimal sets the buying price from a decimal value
func (p *Product) SetBuyingPriceFromDecimal(price float64) {
	p.BuyingPrice = int64(price * 100)
}

// SetSellingPriceFromDecimal sets the selling price from a decimal value
func (p *Product) SetSellingPriceFromDecimal(price float64) {
	p.SellingPrice = int64(price * 100)
}

// ProductJSON is a helper struct for JSON marshaling with decimal prices
type ProductJSON struct {
	ID            uuid.UUID    `json:"id"`
	UserID        uuid.UUID    `json:"user_id"`
	CategoryID    *uuid.UUID   `json:"category_id,omitempty"`
	UnitID        *uuid.UUID   `json:"unit_id,omitempty"`
	Name          string       `json:"name"`
	Slug          string       `json:"slug"`
	Code          string       `json:"code"`
	Quantity      int          `json:"quantity"`
	QuantityAlert int          `json:"quantity_alert"`
	BuyingPrice   float64      `json:"buying_price"`  // Decimal value for JSON
	SellingPrice  float64      `json:"selling_price"` // Decimal value for JSON
	Tax           int          `json:"tax"`
	TaxType       enum.TaxType `json:"tax_type"`
	Notes         *string      `json:"notes,omitempty"`
	ProductImage  *string      `json:"product_image,omitempty"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
	Category      *Category    `json:"category,omitempty"`
	Unit          *Unit        `json:"unit,omitempty"`
}

// MarshalJSON converts Product to JSON with decimal prices
func (p Product) MarshalJSON() ([]byte, error) {
	return json.Marshal(ProductJSON{
		ID:            p.ID,
		UserID:        p.UserID,
		CategoryID:    p.CategoryID,
		UnitID:        p.UnitID,
		Name:          p.Name,
		Slug:          p.Slug,
		Code:          p.Code,
		Quantity:      p.Quantity,
		QuantityAlert: p.QuantityAlert,
		BuyingPrice:   p.GetBuyingPriceDecimal(),
		SellingPrice:  p.GetSellingPriceDecimal(),
		Tax:           p.Tax,
		TaxType:       p.TaxType,
		Notes:         p.Notes,
		ProductImage:  p.ProductImage,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
		Category:      p.Category,
		Unit:          p.Unit,
	})
}

// Category represents a product category
type Category struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	TenantID  uuid.UUID      `gorm:"type:uuid;not null;index" json:"tenant_id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	Name      string         `gorm:"size:255;not null" json:"name"`
	Slug      string         `gorm:"size:255;unique;not null" json:"slug"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Tenant   Tenant    `gorm:"foreignKey:TenantID" json:"-"`
	User     User      `gorm:"foreignKey:UserID" json:"-"`
	Products []Product `gorm:"foreignKey:CategoryID" json:"-"`
}

// BeforeCreate generates a UUID before creating a new category
func (c *Category) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// TableName returns the table name for the Category model
func (Category) TableName() string {
	return "categories"
}

// Unit represents a unit of measurement
type Unit struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	TenantID  uuid.UUID      `gorm:"type:uuid;not null;index" json:"tenant_id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	Name      string         `gorm:"size:255;not null" json:"name"`
	Slug      string         `gorm:"size:255;unique;not null" json:"slug"`
	ShortCode string         `gorm:"size:50" json:"short_code"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Tenant   Tenant    `gorm:"foreignKey:TenantID" json:"-"`
	User     User      `gorm:"foreignKey:UserID" json:"-"`
	Products []Product `gorm:"foreignKey:UnitID" json:"-"`
}

// BeforeCreate generates a UUID before creating a new unit
func (u *Unit) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// TableName returns the table name for the Unit model
func (Unit) TableName() string {
	return "units"
}

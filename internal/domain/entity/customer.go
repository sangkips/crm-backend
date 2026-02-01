package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Customer represents a customer in the CRM
type Customer struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	TenantID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"tenant_id"`
	UserID        uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	Name          string         `gorm:"size:255;not null" json:"name"`
	Email         *string        `gorm:"size:255" json:"email,omitempty"`
	Phone         *string        `gorm:"size:50" json:"phone,omitempty"`
	KRAPin        *string        `gorm:"size:50;column:kra_pin" json:"kra_pin,omitempty"`
	Address       *string        `gorm:"type:text" json:"address,omitempty"`
	Photo         *string        `gorm:"size:255" json:"photo,omitempty"`
	AccountHolder *string        `gorm:"size:255" json:"account_holder,omitempty"`
	AccountNumber *string        `gorm:"size:100" json:"account_number,omitempty"`
	BankName      *string        `gorm:"size:255" json:"bank_name,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Tenant     Tenant      `gorm:"foreignKey:TenantID" json:"-"`
	User       User        `gorm:"foreignKey:UserID" json:"-"`
	Orders     []Order     `gorm:"foreignKey:CustomerID" json:"-"`
	Quotations []Quotation `gorm:"foreignKey:CustomerID" json:"-"`
}

// BeforeCreate generates a UUID before creating a new customer
func (c *Customer) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// TableName returns the table name for the Customer model
func (Customer) TableName() string {
	return "customers"
}

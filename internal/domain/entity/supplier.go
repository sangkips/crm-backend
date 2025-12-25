package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/domain/enum"
	"gorm.io/gorm"
)

// Supplier represents a supplier in the CRM
type Supplier struct {
	ID            uuid.UUID         `gorm:"type:uuid;primary_key" json:"id"`
	UserID        uuid.UUID         `gorm:"type:uuid;not null;index" json:"user_id"`
	Name          string            `gorm:"size:255;not null" json:"name"`
	Email         *string           `gorm:"size:255" json:"email,omitempty"`
	Phone         *string           `gorm:"size:50" json:"phone,omitempty"`
	Address       *string           `gorm:"type:text" json:"address,omitempty"`
	ShopName      *string           `gorm:"size:255;column:shopname" json:"shopname,omitempty"`
	KRAPin        *string           `gorm:"size:50;column:kra_pin" json:"kra_pin,omitempty"`
	Type          enum.SupplierType `gorm:"size:50;default:'distributor'" json:"type"`
	Photo         *string           `gorm:"size:255" json:"photo,omitempty"`
	AccountHolder *string           `gorm:"size:255" json:"account_holder,omitempty"`
	AccountNumber *string           `gorm:"size:100" json:"account_number,omitempty"`
	BankName      *string           `gorm:"size:255" json:"bank_name,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	DeletedAt     gorm.DeletedAt    `gorm:"index" json:"-"`

	// Relationships
	User      User       `gorm:"foreignKey:UserID" json:"-"`
	Purchases []Purchase `gorm:"foreignKey:SupplierID" json:"-"`
}

// BeforeCreate generates a UUID before creating a new supplier
func (s *Supplier) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// TableName returns the table name for the Supplier model
func (Supplier) TableName() string {
	return "suppliers"
}

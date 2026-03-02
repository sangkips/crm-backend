package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/domain/enum"
	"gorm.io/gorm"
)

// MpesaTransaction tracks an M-Pesa STK Push payment transaction
type MpesaTransaction struct {
	ID                 uuid.UUID                   `gorm:"type:uuid;primary_key" json:"id"`
	TenantID           uuid.UUID                   `gorm:"type:uuid;not null;index" json:"tenant_id"`
	OrderID            uuid.UUID                   `gorm:"type:uuid;not null;index" json:"order_id"`
	PhoneNumber        string                      `gorm:"size:20;not null" json:"phone_number"`
	Amount             int64                       `gorm:"not null" json:"-"` // Stored in cents
	MerchantRequestID  string                      `gorm:"size:100" json:"merchant_request_id"`
	CheckoutRequestID  string                      `gorm:"size:100;uniqueIndex" json:"checkout_request_id"`
	ResultCode         *int                        `json:"result_code,omitempty"`
	ResultDesc         string                      `gorm:"size:255" json:"result_desc,omitempty"`
	MpesaReceiptNumber string                      `gorm:"size:50" json:"mpesa_receipt_number,omitempty"`
	Status             enum.MpesaTransactionStatus `gorm:"default:0" json:"status"`
	CreatedAt          time.Time                   `json:"created_at"`
	UpdatedAt          time.Time                   `json:"updated_at"`
	DeletedAt          gorm.DeletedAt              `gorm:"index" json:"-"`

	// Relationships
	Tenant Tenant `gorm:"foreignKey:TenantID" json:"-"`
	Order  Order  `gorm:"foreignKey:OrderID" json:"-"`
}

// BeforeCreate generates a UUID before creating a new transaction
func (t *MpesaTransaction) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// TableName returns the table name for the MpesaTransaction model
func (MpesaTransaction) TableName() string {
	return "mpesa_transactions"
}

// GetAmountDecimal returns the amount as a decimal (from cents)
func (t *MpesaTransaction) GetAmountDecimal() float64 {
	return float64(t.Amount) / 100
}

// GetAmountWhole returns the amount as a whole number (KES, no cents)
func (t *MpesaTransaction) GetAmountWhole() int {
	return int(t.Amount / 100)
}

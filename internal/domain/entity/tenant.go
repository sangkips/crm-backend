package entity

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Tenant represents an organization/company in the multitenant system
type Tenant struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	Name      string         `gorm:"size:255;not null" json:"name"`
	Slug      string         `gorm:"size:255;unique;not null" json:"slug"`
	OwnerID   uuid.UUID      `gorm:"type:uuid;not null;index" json:"owner_id"`
	Settings  TenantSettings `gorm:"type:jsonb;serializer:json" json:"settings"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Owner   User               `gorm:"foreignKey:OwnerID" json:"-"`
	Members []TenantMembership `gorm:"foreignKey:TenantID" json:"-"`
}

// BeforeCreate generates a UUID before creating a new tenant
func (t *Tenant) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// TableName returns the table name for the Tenant model
func (Tenant) TableName() string {
	return "tenants"
}

// MemberUser represents a subset of user fields for membership responses
type MemberUser struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
}

// TenantMembership represents a user's membership in a tenant
type TenantMembership struct {
	TenantID  uuid.UUID `gorm:"type:uuid;primaryKey" json:"tenant_id"`
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
	Role      string    `gorm:"size:50;default:'member'" json:"role"` // owner, admin, member
	CreatedAt time.Time `json:"created_at"`

	// Relationships
	Tenant Tenant `gorm:"foreignKey:TenantID" json:"-"`
	User   User   `gorm:"foreignKey:UserID" json:"-"`

	// Computed field for JSON response
	MemberUser *MemberUser `gorm:"-" json:"user,omitempty"`
}

// PopulateUserDetails populates the MemberUser field from the User relationship
func (tm *TenantMembership) PopulateUserDetails() {
	if tm.User.ID != uuid.Nil {
		tm.MemberUser = &MemberUser{
			ID:        tm.User.ID,
			FirstName: tm.User.FirstName,
			LastName:  tm.User.LastName,
			Email:     tm.User.Email,
		}
	}
}

// TableName returns the table name for the TenantMembership model
func (TenantMembership) TableName() string {
	return "tenant_memberships"
}

// TenantSettings holds all customizable tenant configurations
type TenantSettings struct {
	// Branding & Appearance
	LogoURL        string `json:"logo_url,omitempty"`
	FaviconURL     string `json:"favicon_url,omitempty"`
	PrimaryColor   string `json:"primary_color,omitempty"`
	SecondaryColor string `json:"secondary_color,omitempty"`

	// Localization
	Currency   string `json:"currency,omitempty"`
	Timezone   string `json:"timezone,omitempty"`
	Locale     string `json:"locale,omitempty"`
	DateFormat string `json:"date_format,omitempty"`

	// Business Configuration
	TaxRate         float64 `json:"tax_rate,omitempty"`
	TaxLabel        string  `json:"tax_label,omitempty"`
	InvoicePrefix   string  `json:"invoice_prefix,omitempty"`
	QuotationPrefix string  `json:"quotation_prefix,omitempty"`

	// Payment Integrations
	Mpesa    *MpesaIntegration    `json:"mpesa,omitempty"`
	Stripe   *StripeIntegration   `json:"stripe,omitempty"`
	Paystack *PaystackIntegration `json:"paystack,omitempty"`

	// Notification Settings
	EmailNotifications bool   `json:"email_notifications,omitempty"`
	SMSNotifications   bool   `json:"sms_notifications,omitempty"`
	WebhookURL         string `json:"webhook_url,omitempty"`

	// Feature Flags
	Features TenantFeatures `json:"features,omitempty"`
}

// Scan implements the sql.Scanner interface for TenantSettings
func (ts *TenantSettings) Scan(value interface{}) error {
	if value == nil {
		*ts = TenantSettings{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("failed to scan TenantSettings: unsupported type")
	}

	return json.Unmarshal(bytes, ts)
}

// Value implements the driver.Valuer interface for TenantSettings
func (ts TenantSettings) Value() (driver.Value, error) {
	return json.Marshal(ts)
}

// TenantFeatures holds feature flags for a tenant
type TenantFeatures struct {
	EnableInvoicing  bool `json:"invoicing"`
	EnableQuotations bool `json:"quotations"`
	EnableReports    bool `json:"reports"`
	EnableInventory  bool `json:"inventory"`
	EnableMultiUser  bool `json:"multi_user"`
	EnableAPIAccess  bool `json:"api_access"`
}

// MpesaIntegration holds M-Pesa configuration
type MpesaIntegration struct {
	ConsumerKey    string `json:"consumer_key"`
	ConsumerSecret string `json:"consumer_secret"`
	ShortCode      string `json:"short_code"`
	PassKey        string `json:"pass_key"`
	Environment    string `json:"environment"` // sandbox, production
}

// StripeIntegration holds Stripe configuration
type StripeIntegration struct {
	PublishableKey string `json:"publishable_key"`
	SecretKey      string `json:"secret_key"`
	WebhookSecret  string `json:"webhook_secret"`
}

// PaystackIntegration holds Paystack configuration
type PaystackIntegration struct {
	PublicKey string `json:"public_key"`
	SecretKey string `json:"secret_key"`
}

// DefaultTenantSettings returns default settings for new tenants
func DefaultTenantSettings() TenantSettings {
	return TenantSettings{
		Currency:           "KES",
		Timezone:           "Africa/Nairobi",
		Locale:             "en-KE",
		DateFormat:         "DD/MM/YYYY",
		TaxRate:            16.0,
		TaxLabel:           "VAT",
		InvoicePrefix:      "INV-",
		QuotationPrefix:    "QUO-",
		EmailNotifications: true,
		Features: TenantFeatures{
			EnableInvoicing:  true,
			EnableQuotations: true,
			EnableReports:    true,
			EnableInventory:  true,
			EnableMultiUser:  true,
			EnableAPIAccess:  false,
		},
	}
}

package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserSettings represents user-specific application settings
type UserSettings struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex" json:"user_id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// General settings
	Language   string `gorm:"size:10;default:'en'" json:"language"`
	Timezone   string `gorm:"size:50;default:'Africa/Nairobi'" json:"timezone"`
	Currency   string `gorm:"size:10;default:'KES'" json:"currency"`
	DateFormat string `gorm:"size:20;default:'DD/MM/YYYY'" json:"date_format"`

	// Notification settings
	EmailNotifications bool `gorm:"default:true" json:"email_notifications"`
	PushNotifications  bool `gorm:"default:true" json:"push_notifications"`
	OrderAlerts        bool `gorm:"default:true" json:"order_alerts"`
	LowStockAlerts     bool `gorm:"default:true" json:"low_stock_alerts"`
	MarketingEmails    bool `gorm:"default:false" json:"marketing_emails"`

	// Appearance settings
	Theme          string `gorm:"size:20;default:'light'" json:"theme"`
	CompactMode    bool   `gorm:"default:false" json:"compact_mode"`
	ShowAnimations bool   `gorm:"default:true" json:"show_animations"`

	// Security settings
	TwoFactorAuth  bool   `gorm:"default:false" json:"two_factor_auth"`
	SessionTimeout string `gorm:"size:10;default:'30'" json:"session_timeout"`
	LoginAlerts    bool   `gorm:"default:true" json:"login_alerts"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"-"`
}

// BeforeCreate generates a UUID before creating new settings
func (s *UserSettings) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// TableName returns the table name for the UserSettings model
func (UserSettings) TableName() string {
	return "user_settings"
}

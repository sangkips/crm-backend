package entity

import (
	"time"

	"github.com/google/uuid"
)

// IdempotencyKey stores processed requests to prevent duplicates
type IdempotencyKey struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Key          string    `gorm:"uniqueIndex;size:255;not null"` // The idempotency key from client
	UserID       uuid.UUID `gorm:"type:uuid;not null;index"`      // User who made the request
	Endpoint     string    `gorm:"size:255;not null"`             // API endpoint (e.g., "POST /orders")
	RequestHash  string    `gorm:"size:64"`                       // SHA256 hash of request body (optional)
	ResponseCode int       `gorm:"not null"`                      // HTTP status code of original response
	ResponseBody string    `gorm:"type:text"`                     // JSON response body (cached)
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	ExpiresAt    time.Time `gorm:"not null;index"` // Keys expire after 24 hours
}

// TableName returns the table name for IdempotencyKey
func (IdempotencyKey) TableName() string {
	return "idempotency_keys"
}

// IsExpired checks if the idempotency key has expired
func (i *IdempotencyKey) IsExpired() bool {
	return time.Now().After(i.ExpiresAt)
}

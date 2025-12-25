package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/domain/entity"
)

// IdempotencyRepository defines the interface for idempotency key operations
type IdempotencyRepository interface {
	// GetByKey retrieves an idempotency key by its key string and user ID
	GetByKey(ctx context.Context, key string, userID uuid.UUID) (*entity.IdempotencyKey, error)
	// Create stores a new idempotency key
	Create(ctx context.Context, ikey *entity.IdempotencyKey) error
	// DeleteExpired removes expired idempotency keys (for cleanup)
	DeleteExpired(ctx context.Context) error
}

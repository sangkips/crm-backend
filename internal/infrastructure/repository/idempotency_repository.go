package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/domain/entity"
	domainRepo "github.com/sangkips/investify-api/internal/domain/repository"
	"gorm.io/gorm"
)

type idempotencyRepository struct {
	db *gorm.DB
}

// NewIdempotencyRepository creates a new idempotency repository
func NewIdempotencyRepository(db *gorm.DB) domainRepo.IdempotencyRepository {
	return &idempotencyRepository{db: db}
}

func (r *idempotencyRepository) GetByKey(ctx context.Context, key string, userID uuid.UUID) (*entity.IdempotencyKey, error) {
	var ikey entity.IdempotencyKey
	err := r.db.WithContext(ctx).
		Where("key = ? AND user_id = ?", key, userID).
		First(&ikey).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &ikey, err
}

func (r *idempotencyRepository) Create(ctx context.Context, ikey *entity.IdempotencyKey) error {
	return r.db.WithContext(ctx).Create(ikey).Error
}

func (r *idempotencyRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&entity.IdempotencyKey{}).Error
}

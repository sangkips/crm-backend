package repository

import (
	"context"
	"time"

	"github.com/sangkips/investify-api/internal/domain/entity"
	"github.com/sangkips/investify-api/internal/domain/repository"
	"gorm.io/gorm"
)

// emailVerificationTokenRepository implements the EmailVerificationTokenRepository interface
type emailVerificationTokenRepository struct {
	db *gorm.DB
}

// NewEmailVerificationTokenRepository creates a new email verification token repository
func NewEmailVerificationTokenRepository(db *gorm.DB) repository.EmailVerificationTokenRepository {
	return &emailVerificationTokenRepository{db: db}
}

// Create stores a new email verification token
func (r *emailVerificationTokenRepository) Create(ctx context.Context, token *entity.EmailVerificationToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

// GetByToken retrieves a token by its value
func (r *emailVerificationTokenRepository) GetByToken(ctx context.Context, token string) (*entity.EmailVerificationToken, error) {
	var verificationToken entity.EmailVerificationToken
	err := r.db.WithContext(ctx).Where("token = ?", token).First(&verificationToken).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &verificationToken, nil
}

// MarkAsUsed marks a token as used
func (r *emailVerificationTokenRepository) MarkAsUsed(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).
		Model(&entity.EmailVerificationToken{}).
		Where("token = ?", token).
		Update("used", true).Error
}

// DeleteByEmail deletes all tokens for a specific email
func (r *emailVerificationTokenRepository) DeleteByEmail(ctx context.Context, email string) error {
	return r.db.WithContext(ctx).
		Where("email = ?", email).
		Delete(&entity.EmailVerificationToken{}).Error
}

// DeleteExpired deletes all expired tokens
func (r *emailVerificationTokenRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&entity.EmailVerificationToken{}).Error
}

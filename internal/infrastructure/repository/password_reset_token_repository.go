package repository

import (
	"context"
	"time"

	"github.com/sangkips/investify-api/internal/domain/entity"
	"github.com/sangkips/investify-api/internal/domain/repository"
	"gorm.io/gorm"
)

// passwordResetTokenRepository implements the PasswordResetTokenRepository interface
type passwordResetTokenRepository struct {
	db *gorm.DB
}

// NewPasswordResetTokenRepository creates a new password reset token repository
func NewPasswordResetTokenRepository(db *gorm.DB) repository.PasswordResetTokenRepository {
	return &passwordResetTokenRepository{db: db}
}

// Create stores a new password reset token
func (r *passwordResetTokenRepository) Create(ctx context.Context, token *entity.PasswordResetToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

// GetByToken retrieves a token by its value
func (r *passwordResetTokenRepository) GetByToken(ctx context.Context, token string) (*entity.PasswordResetToken, error) {
	var resetToken entity.PasswordResetToken
	err := r.db.WithContext(ctx).Where("token = ?", token).First(&resetToken).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &resetToken, nil
}

// MarkAsUsed marks a token as used
func (r *passwordResetTokenRepository) MarkAsUsed(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).
		Model(&entity.PasswordResetToken{}).
		Where("token = ?", token).
		Update("used", true).Error
}

// DeleteByEmail deletes all tokens for a specific email
func (r *passwordResetTokenRepository) DeleteByEmail(ctx context.Context, email string) error {
	return r.db.WithContext(ctx).
		Where("email = ?", email).
		Delete(&entity.PasswordResetToken{}).Error
}

// DeleteExpired deletes all expired tokens
func (r *passwordResetTokenRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&entity.PasswordResetToken{}).Error
}

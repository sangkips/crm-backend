package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/domain/entity"
	"github.com/sangkips/investify-api/internal/domain/repository"
	"gorm.io/gorm"
)

type settingsRepository struct {
	db *gorm.DB
}

// NewSettingsRepository creates a new settings repository
func NewSettingsRepository(db *gorm.DB) repository.SettingsRepository {
	return &settingsRepository{db: db}
}

// GetByUserID retrieves settings by user ID
func (r *settingsRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.UserSettings, error) {
	var settings entity.UserSettings
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&settings).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &settings, nil
}

// Create creates new user settings
func (r *settingsRepository) Create(ctx context.Context, settings *entity.UserSettings) error {
	return r.db.WithContext(ctx).Create(settings).Error
}

// Update updates existing user settings
func (r *settingsRepository) Update(ctx context.Context, settings *entity.UserSettings) error {
	return r.db.WithContext(ctx).Save(settings).Error
}

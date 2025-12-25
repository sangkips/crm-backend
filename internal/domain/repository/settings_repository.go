package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/domain/entity"
)

// SettingsRepository defines the interface for settings data access
type SettingsRepository interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.UserSettings, error)
	Create(ctx context.Context, settings *entity.UserSettings) error
	Update(ctx context.Context, settings *entity.UserSettings) error
}

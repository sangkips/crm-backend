package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/domain/entity"
	"github.com/sangkips/investify-api/internal/domain/repository"
)

// SettingsService handles settings-related business logic
type SettingsService struct {
	settingsRepo repository.SettingsRepository
}

// NewSettingsService creates a new settings service
func NewSettingsService(settingsRepo repository.SettingsRepository) *SettingsService {
	return &SettingsService{
		settingsRepo: settingsRepo,
	}
}

// GetSettings retrieves user settings, creating defaults if not exists
func (s *SettingsService) GetSettings(ctx context.Context, userID uuid.UUID) (*entity.UserSettings, error) {
	settings, err := s.settingsRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// If no settings exist, create default settings
	if settings == nil {
		settings = &entity.UserSettings{
			UserID:             userID,
			Language:           "en",
			Timezone:           "Africa/Nairobi",
			Currency:           "KES",
			DateFormat:         "DD/MM/YYYY",
			EmailNotifications: true,
			PushNotifications:  true,
			OrderAlerts:        true,
			LowStockAlerts:     true,
			MarketingEmails:    false,
			Theme:              "light",
			CompactMode:        false,
			ShowAnimations:     true,
			TwoFactorAuth:      false,
			SessionTimeout:     "30",
			LoginAlerts:        true,
		}
		if err := s.settingsRepo.Create(ctx, settings); err != nil {
			return nil, err
		}
	}

	return settings, nil
}

// UpdateSettingsInput represents the input for updating settings
type UpdateSettingsInput struct {
	UserID             uuid.UUID
	Language           string
	Timezone           string
	Currency           string
	DateFormat         string
	EmailNotifications bool
	PushNotifications  bool
	OrderAlerts        bool
	LowStockAlerts     bool
	MarketingEmails    bool
	Theme              string
	CompactMode        bool
	ShowAnimations     bool
	TwoFactorAuth      bool
	SessionTimeout     string
	LoginAlerts        bool
}

// UpdateSettings updates user settings
func (s *SettingsService) UpdateSettings(ctx context.Context, input *UpdateSettingsInput) (*entity.UserSettings, error) {
	settings, err := s.settingsRepo.GetByUserID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	// If no settings exist, create new
	if settings == nil {
		settings = &entity.UserSettings{
			UserID: input.UserID,
		}
	}

	// Update fields
	settings.Language = input.Language
	settings.Timezone = input.Timezone
	settings.Currency = input.Currency
	settings.DateFormat = input.DateFormat
	settings.EmailNotifications = input.EmailNotifications
	settings.PushNotifications = input.PushNotifications
	settings.OrderAlerts = input.OrderAlerts
	settings.LowStockAlerts = input.LowStockAlerts
	settings.MarketingEmails = input.MarketingEmails
	settings.Theme = input.Theme
	settings.CompactMode = input.CompactMode
	settings.ShowAnimations = input.ShowAnimations
	settings.TwoFactorAuth = input.TwoFactorAuth
	settings.SessionTimeout = input.SessionTimeout
	settings.LoginAlerts = input.LoginAlerts

	if settings.ID == uuid.Nil {
		if err := s.settingsRepo.Create(ctx, settings); err != nil {
			return nil, err
		}
	} else {
		if err := s.settingsRepo.Update(ctx, settings); err != nil {
			return nil, err
		}
	}

	return settings, nil
}

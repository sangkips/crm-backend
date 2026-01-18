package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sangkips/investify-api/internal/application/service"
	"github.com/sangkips/investify-api/internal/presentation/http/dto/response"
)

// SettingsHandler handles settings-related HTTP requests
type SettingsHandler struct {
	settingsService *service.SettingsService
}

// NewSettingsHandler creates a new settings handler
func NewSettingsHandler(settingsService *service.SettingsService) *SettingsHandler {
	return &SettingsHandler{settingsService: settingsService}
}

// GetSettings retrieves user settings
func (h *SettingsHandler) GetSettings(c *gin.Context) {
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	settings, err := h.settingsService.GetSettings(c.Request.Context(), *userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "Settings retrieved successfully", settings)
}

// UpdateSettings updates user settings
func (h *SettingsHandler) UpdateSettings(c *gin.Context) {
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req struct {
		Language           string `json:"language"`
		Timezone           string `json:"timezone"`
		Currency           string `json:"currency"`
		DateFormat         string `json:"date_format"`
		EmailNotifications bool   `json:"email_notifications"`
		PushNotifications  bool   `json:"push_notifications"`
		OrderAlerts        bool   `json:"order_alerts"`
		LowStockAlerts     bool   `json:"low_stock_alerts"`
		MarketingEmails    bool   `json:"marketing_emails"`
		Theme              string `json:"theme"`
		CompactMode        bool   `json:"compact_mode"`
		ShowAnimations     bool   `json:"show_animations"`
		TwoFactorAuth      bool   `json:"two_factor_auth"`
		SessionTimeout     string `json:"session_timeout"`
		LoginAlerts        bool   `json:"login_alerts"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	settings, err := h.settingsService.UpdateSettings(c.Request.Context(), &service.UpdateSettingsInput{
		UserID:             *userID,
		Language:           req.Language,
		Timezone:           req.Timezone,
		Currency:           req.Currency,
		DateFormat:         req.DateFormat,
		EmailNotifications: req.EmailNotifications,
		PushNotifications:  req.PushNotifications,
		OrderAlerts:        req.OrderAlerts,
		LowStockAlerts:     req.LowStockAlerts,
		MarketingEmails:    req.MarketingEmails,
		Theme:              req.Theme,
		CompactMode:        req.CompactMode,
		ShowAnimations:     req.ShowAnimations,
		TwoFactorAuth:      req.TwoFactorAuth,
		SessionTimeout:     req.SessionTimeout,
		LoginAlerts:        req.LoginAlerts,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "Settings updated successfully", settings)
}

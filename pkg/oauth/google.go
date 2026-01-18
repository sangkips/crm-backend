package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	ErrInvalidCode        = errors.New("invalid authorization code")
	ErrFailedToGetUser    = errors.New("failed to get user info from Google")
	ErrInvalidState       = errors.New("invalid state parameter")
	ErrOAuthNotConfigured = errors.New("Google OAuth is not configured")
)

// GoogleUserInfo represents user information from Google
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

// GoogleOAuthService handles Google OAuth operations
type GoogleOAuthService struct {
	config             *oauth2.Config
	frontendSuccessURL string
	frontendErrorURL   string
}

// GoogleOAuthConfig holds the configuration for Google OAuth
type GoogleOAuthConfig struct {
	ClientID           string
	ClientSecret       string
	RedirectURL        string
	FrontendSuccessURL string
	FrontendErrorURL   string
}

// NewGoogleOAuthService creates a new Google OAuth service
func NewGoogleOAuthService(cfg GoogleOAuthConfig) *GoogleOAuthService {
	config := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	return &GoogleOAuthService{
		config:             config,
		frontendSuccessURL: cfg.FrontendSuccessURL,
		frontendErrorURL:   cfg.FrontendErrorURL,
	}
}

// IsConfigured checks if Google OAuth is properly configured
func (s *GoogleOAuthService) IsConfigured() bool {
	return s.config.ClientID != "" && s.config.ClientSecret != ""
}

// GetAuthURL returns the URL to redirect the user to for Google OAuth consent
func (s *GoogleOAuthService) GetAuthURL(state string) string {
	return s.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// ExchangeCode exchanges the authorization code for tokens
func (s *GoogleOAuthService) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := s.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidCode, err)
	}
	return token, nil
}

// GetUserInfo fetches user information from Google using the access token
func (s *GoogleOAuthService) GetUserInfo(ctx context.Context, token *oauth2.Token) (*GoogleUserInfo, error) {
	client := s.config.Client(ctx, token)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedToGetUser, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: status %d, body: %s", ErrFailedToGetUser, resp.StatusCode, string(body))
	}

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedToGetUser, err)
	}

	return &userInfo, nil
}

// GetFrontendSuccessURL returns the frontend URL to redirect to after successful OAuth
func (s *GoogleOAuthService) GetFrontendSuccessURL() string {
	return s.frontendSuccessURL
}

// GetFrontendErrorURL returns the frontend URL to redirect to after failed OAuth
func (s *GoogleOAuthService) GetFrontendErrorURL() string {
	return s.frontendErrorURL
}

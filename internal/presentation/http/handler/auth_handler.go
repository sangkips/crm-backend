package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/sangkips/investify-api/internal/application/service"
	"github.com/sangkips/investify-api/internal/presentation/http/dto/request"
	"github.com/sangkips/investify-api/internal/presentation/http/dto/response"
	"github.com/sangkips/investify-api/pkg/apperror"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req request.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	output, err := h.authService.Login(c.Request.Context(), &service.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "Login successful", gin.H{
		"user": gin.H{
			"id":          output.User.ID,
			"first_name":  output.User.FirstName,
			"last_name":   output.User.LastName,
			"email":       output.User.Email,
			"username":    output.User.Username,
			"photo":       output.User.Photo,
			"store_name":  output.User.StoreName,
			"roles":       output.User.Roles,
			"permissions": output.User.GetPermissions(),
		},
		"access_token":  output.AccessToken,
		"refresh_token": output.RefreshToken,
		"token_type":    "Bearer",
	})
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req request.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	user, err := h.authService.Register(c.Request.Context(), &service.RegisterInput{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  req.Password,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, "Registration successful", gin.H{
		"user": gin.H{
			"id":         user.ID,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"email":      user.Email,
			"username":   user.Username,
		},
	})
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req request.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	output, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "Token refreshed successfully", gin.H{
		"access_token":  output.AccessToken,
		"refresh_token": output.RefreshToken,
		"token_type":    "Bearer",
	})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// JWT is stateless, so we just return success
	// Client should discard the tokens
	response.OK(c, "Logged out successfully", nil)
}

// GetProfile handles fetching current user profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	user, err := h.authService.GetCurrentUser(c.Request.Context(), *userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "Profile retrieved successfully", gin.H{
		"user": gin.H{
			"id":            user.ID,
			"first_name":    user.FirstName,
			"last_name":     user.LastName,
			"email":         user.Email,
			"username":      user.Username,
			"photo":         user.Photo,
			"store_name":    user.StoreName,
			"store_address": user.StoreAddress,
			"store_phone":   user.StorePhone,
			"store_email":   user.StoreEmail,
			"roles":         user.Roles,
			"permissions":   user.GetPermissions(),
			"created_at":    user.CreatedAt,
			"updated_at":    user.UpdatedAt,
		},
	})
}

// UpdateProfile handles updating user profile
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req struct {
		FirstName    string  `json:"first_name"`
		LastName     string  `json:"last_name"`
		Username     string  `json:"username"`
		Photo        *string `json:"photo"`
		StoreName    *string `json:"store_name"`
		StoreAddress *string `json:"store_address"`
		StorePhone   *string `json:"store_phone"`
		StoreEmail   *string `json:"store_email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	user, err := h.authService.UpdateProfile(c.Request.Context(), &service.UpdateProfileInput{
		UserID:       *userID,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Username:     req.Username,
		Photo:        req.Photo,
		StoreName:    req.StoreName,
		StoreAddress: req.StoreAddress,
		StorePhone:   req.StorePhone,
		StoreEmail:   req.StoreEmail,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "Profile updated successfully", gin.H{
		"user": user,
	})
}

// ChangePassword handles password change
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req request.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	err := h.authService.ChangePassword(c.Request.Context(), &service.ChangePasswordInput{
		UserID:          *userID,
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	})
	if err != nil {
		if apperror.IsAppError(err) {
			response.Error(c, err)
		} else {
			response.BadRequest(c, err.Error())
		}
		return
	}

	response.OK(c, "Password changed successfully", nil)
}

// ForgotPassword handles forgot password request
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req request.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	// Call the service (it handles email enumeration protection internally)
	_ = h.authService.ForgotPassword(c.Request.Context(), &service.ForgotPasswordInput{
		Email: req.Email,
	})

	// Always return success to prevent email enumeration
	response.OK(c, "If the email exists, a reset link has been sent", nil)
}

// ResetPassword handles password reset
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req request.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	err := h.authService.ResetPassword(c.Request.Context(), &service.ResetPasswordInput{
		Email:       req.Email,
		Token:       req.Token,
		NewPassword: req.Password,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "Password reset successfully", nil)
}

// GoogleAuth initiates Google OAuth flow and redirect to Google OAuth consent screen
func (h *AuthHandler) GoogleAuth(c *gin.Context) {
	// Generate a random state parameter for the OAuth flow
	// Google will preserve and return this in the callback
	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		response.InternalServerError(c, "Failed to generate state")
		return
	}
	state := hex.EncodeToString(stateBytes)

	authURL, err := h.authService.GetGoogleAuthURL(state)
	if err != nil {
		response.BadRequest(c, "Google OAuth is not configured")
		return
	}

	c.Redirect(302, authURL)
}

// GoogleCallback handles Google OAuth callback and authenticate user
func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	successURL, errorURL := h.authService.GetGoogleFrontendURLs()

	// Helper function to redirect to error page
	redirectError := func(msg string) {
		errorURLParsed, _ := url.Parse(errorURL)
		q := errorURLParsed.Query()
		q.Set("error", msg)
		errorURLParsed.RawQuery = q.Encode()
		c.Redirect(302, errorURLParsed.String())
	}

	// Check for error from Google first
	if errMsg := c.Query("error"); errMsg != "" {
		redirectError(errMsg)
		return
	}

	// Get authorization code
	code := c.Query("code")
	if code == "" {
		redirectError("No authorization code received")
		return
	}

	// Validate state is present (Google preserves it through the flow)
	state := c.Query("state")
	if state == "" {
		redirectError("Missing state parameter")
		return
	}

	// Authenticate with Google
	output, err := h.authService.GoogleAuth(c.Request.Context(), &service.GoogleAuthInput{
		Code:  code,
		State: state,
	})
	if err != nil {
		redirectError("Authentication failed")
		return
	}

	// Redirect to frontend with tokens
	successURLParsed, _ := url.Parse(successURL)
	q := successURLParsed.Query()
	q.Set("access_token", output.AccessToken)
	q.Set("refresh_token", output.RefreshToken)
	q.Set("token_type", "Bearer")
	if output.IsNewUser {
		q.Set("is_new_user", "true")
	}
	successURLParsed.RawQuery = q.Encode()

	c.Redirect(302, successURLParsed.String())
}

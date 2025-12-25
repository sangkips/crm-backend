package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/domain/entity"
	"github.com/sangkips/investify-api/internal/domain/repository"
	"github.com/sangkips/investify-api/pkg/apperror"
	"github.com/sangkips/investify-api/pkg/utils"
)

// AuthService handles authentication-related operations
type AuthService struct {
	userRepo   repository.UserRepository
	roleRepo   repository.RoleRepository
	jwtManager *utils.JWTManager
}

// NewAuthService creates a new auth service
func NewAuthService(userRepo repository.UserRepository, roleRepo repository.RoleRepository, jwtManager *utils.JWTManager) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		roleRepo:   roleRepo,
		jwtManager: jwtManager,
	}
}

// LoginInput represents the login input
type LoginInput struct {
	Email    string
	Password string
}

// LoginOutput represents the login output
type LoginOutput struct {
	User         *entity.User
	AccessToken  string
	RefreshToken string
}

// Login authenticates a user and returns tokens
func (s *AuthService) Login(ctx context.Context, input *LoginInput) (*LoginOutput, error) {
	user, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, apperror.ErrInvalidCredentials
	}

	if !utils.CheckPasswordHash(input.Password, user.Password) {
		return nil, apperror.ErrInvalidCredentials
	}

	// Get user with roles
	user, err = s.userRepo.GetWithRoles(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	// Generate tokens
	roles := make([]string, 0)
	for _, role := range user.Roles {
		roles = append(roles, role.Name)
	}
	permissions := user.GetPermissions()

	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email, roles, permissions)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &LoginOutput{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// RegisterInput represents the registration input
type RegisterInput struct {
	FirstName string
	LastName  string
	Email     string
	Password  string
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, input *RegisterInput) (*entity.User, error) {
	// Check if email already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, apperror.NewConflictError("Email already registered")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	// Generate username from email (part before @)
	username := input.Email
	if atIdx := len(input.Email); atIdx > 0 {
		for i, c := range input.Email {
			if c == '@' {
				username = input.Email[:i]
				break
			}
		}
	}

	user := &entity.User{
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Username:  username,
		Email:     input.Email,
		Password:  hashedPassword,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Assign default "user" role
	defaultRole, err := s.roleRepo.GetByName(ctx, "user")
	if err != nil {
		// Log error but don't fail registration
		return user, nil
	}
	if defaultRole != nil {
		_ = s.userRepo.AssignRole(ctx, user.ID, defaultRole.ID)
	}

	return user, nil
}

// RefreshToken generates new tokens from a refresh token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*LoginOutput, error) {
	userID, err := s.jwtManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, apperror.ErrInvalidToken
	}

	user, err := s.userRepo.GetWithRoles(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, apperror.ErrNotFound
	}

	roles := make([]string, 0)
	for _, role := range user.Roles {
		roles = append(roles, role.Name)
	}
	permissions := user.GetPermissions()

	newAccessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email, roles, permissions)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &LoginOutput{
		User:         user,
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

// GetCurrentUser returns the current user by ID
func (s *AuthService) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*entity.User, error) {
	user, err := s.userRepo.GetWithRoles(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, apperror.ErrNotFound
	}
	return user, nil
}

// ChangePasswordInput represents the change password input
type ChangePasswordInput struct {
	UserID          uuid.UUID
	CurrentPassword string
	NewPassword     string
}

// ChangePassword changes the user's password
func (s *AuthService) ChangePassword(ctx context.Context, input *ChangePasswordInput) error {
	user, err := s.userRepo.GetByID(ctx, input.UserID)
	if err != nil {
		return err
	}
	if user == nil {
		return apperror.ErrNotFound
	}

	if !utils.CheckPasswordHash(input.CurrentPassword, user.Password) {
		return errors.New("current password is incorrect")
	}

	hashedPassword, err := utils.HashPassword(input.NewPassword)
	if err != nil {
		return err
	}

	user.Password = hashedPassword
	return s.userRepo.Update(ctx, user)
}

// UpdateProfileInput represents the update profile input
type UpdateProfileInput struct {
	UserID       uuid.UUID
	FirstName    string
	LastName     string
	Username     string
	Photo        *string
	StoreName    *string
	StoreAddress *string
	StorePhone   *string
	StoreEmail   *string
}

// UpdateProfile updates the user's profile
func (s *AuthService) UpdateProfile(ctx context.Context, input *UpdateProfileInput) (*entity.User, error) {
	user, err := s.userRepo.GetByID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, apperror.ErrNotFound
	}

	// Check if username is taken by another user
	if input.Username != "" && input.Username != user.Username {
		existingUser, err := s.userRepo.GetByUsername(ctx, input.Username)
		if err != nil {
			return nil, err
		}
		if existingUser != nil && existingUser.ID != user.ID {
			return nil, apperror.NewConflictError("Username already taken")
		}
		user.Username = input.Username
	}

	if input.FirstName != "" {
		user.FirstName = input.FirstName
	}
	if input.LastName != "" {
		user.LastName = input.LastName
	}
	if input.Photo != nil {
		user.Photo = input.Photo
	}
	if input.StoreName != nil {
		user.StoreName = input.StoreName
	}
	if input.StoreAddress != nil {
		user.StoreAddress = input.StoreAddress
	}
	if input.StorePhone != nil {
		user.StorePhone = input.StorePhone
	}
	if input.StoreEmail != nil {
		user.StoreEmail = input.StoreEmail
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

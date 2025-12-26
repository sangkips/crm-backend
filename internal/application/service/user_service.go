package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/domain/entity"
	"github.com/sangkips/investify-api/internal/domain/repository"
	"github.com/sangkips/investify-api/pkg/apperror"
	"github.com/sangkips/investify-api/pkg/pagination"
)

// UserService handles user management operations
type UserService struct {
	userRepo       repository.UserRepository
	roleRepo       repository.RoleRepository
	permissionRepo repository.PermissionRepository
}

// NewUserService creates a new user service
func NewUserService(
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	permissionRepo repository.PermissionRepository,
) *UserService {
	return &UserService{
		userRepo:       userRepo,
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
	}
}

// ListUsersInput represents the input for listing users
type ListUsersInput struct {
	Page    int
	PerPage int
	Search  string
}

// ListUsersOutput represents the output for listing users
type ListUsersOutput struct {
	Users      []entity.User
	Total      int64
	Page       int
	PerPage    int
	TotalPages int
}

// ListUsers returns a paginated list of users with their roles
func (s *UserService) ListUsers(ctx context.Context, input *ListUsersInput) (*ListUsersOutput, error) {
	params := &pagination.PaginationParams{
		Page:    input.Page,
		PerPage: input.PerPage,
	}
	params.Validate()

	users, total, err := s.userRepo.List(ctx, params, input.Search)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / params.PerPage
	if int(total)%params.PerPage > 0 {
		totalPages++
	}

	return &ListUsersOutput{
		Users:      users,
		Total:      total,
		Page:       params.Page,
		PerPage:    params.PerPage,
		TotalPages: totalPages,
	}, nil
}

// GetUser returns a user by ID with roles and permissions
func (s *UserService) GetUser(ctx context.Context, userID uuid.UUID) (*entity.User, error) {
	user, err := s.userRepo.GetWithRoles(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, apperror.ErrNotFound
	}
	return user, nil
}

// UpdateUserRolesInput represents the input for updating user roles
type UpdateUserRolesInput struct {
	UserID  uuid.UUID
	RoleIDs []uint
}

// UpdateUserRoles updates the roles assigned to a user
func (s *UserService) UpdateUserRoles(ctx context.Context, input *UpdateUserRolesInput) (*entity.User, error) {
	// Check if user exists
	user, err := s.userRepo.GetByID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, apperror.ErrNotFound
	}

	// Get current roles
	userWithRoles, err := s.userRepo.GetWithRoles(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	// Create a map of desired role IDs
	desiredRoles := make(map[uint]bool)
	for _, roleID := range input.RoleIDs {
		desiredRoles[roleID] = true
	}

	// Create a map of current role IDs
	currentRoles := make(map[uint]bool)
	for _, role := range userWithRoles.Roles {
		currentRoles[role.ID] = true
	}

	// Remove roles that are no longer desired
	for _, role := range userWithRoles.Roles {
		if !desiredRoles[role.ID] {
			if err := s.userRepo.RemoveRole(ctx, input.UserID, role.ID); err != nil {
				return nil, err
			}
		}
	}

	// Add new roles
	for roleID := range desiredRoles {
		if !currentRoles[roleID] {
			// Verify the role exists
			role, err := s.roleRepo.GetByID(ctx, roleID)
			if err != nil {
				return nil, err
			}
			if role == nil {
				continue // Skip non-existent roles
			}
			if err := s.userRepo.AssignRole(ctx, input.UserID, roleID); err != nil {
				return nil, err
			}
		}
	}

	// Return updated user with roles
	return s.userRepo.GetWithRoles(ctx, input.UserID)
}

// DeleteUser soft deletes a user
func (s *UserService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return apperror.ErrNotFound
	}

	return s.userRepo.Delete(ctx, userID)
}

// ListRoles returns all available roles
func (s *UserService) ListRoles(ctx context.Context) ([]entity.Role, error) {
	return s.roleRepo.List(ctx)
}

// ListPermissions returns all available permissions
func (s *UserService) ListPermissions(ctx context.Context) ([]entity.Permission, error) {
	return s.permissionRepo.List(ctx)
}

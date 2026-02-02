package service

import (
	"context"

	"crypto/rand"
	"encoding/hex"

	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/domain/entity"
	"github.com/sangkips/investify-api/internal/domain/repository"
	"github.com/sangkips/investify-api/pkg/apperror"
)

// TenantService handles tenant-related operations
type TenantService struct {
	tenantRepo repository.TenantRepository
}

// NewTenantService creates a new tenant service
func NewTenantService(tenantRepo repository.TenantRepository) *TenantService {
	return &TenantService{tenantRepo: tenantRepo}
}

// CreateTenantInput represents input for creating a tenant
type CreateTenantInput struct {
	Name     string
	Slug     string
	OwnerID  uuid.UUID
	Settings *entity.TenantSettings
}

// CreateTenant creates a new tenant
func (s *TenantService) CreateTenant(ctx context.Context, input *CreateTenantInput) (*entity.Tenant, error) {
	// Generate slug if not provided
	slug := input.Slug
	if slug == "" {
		slug = s.generateTenantSlug(input.Name)
	}

	// Check if slug already exists
	existing, err := s.tenantRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		if input.Slug != "" {
			return nil, apperror.NewConflictError("Tenant slug already exists")
		}
		// If auto-generated slug exists, try appending randomness
		randomBytes := make([]byte, 4)
		rand.Read(randomBytes)
		slug = slug + "-" + hex.EncodeToString(randomBytes)
	}

	settings := entity.DefaultTenantSettings()
	if input.Settings != nil {
		settings = *input.Settings
	}

	tenant := &entity.Tenant{
		Name:     input.Name,
		Slug:     slug,
		OwnerID:  input.OwnerID,
		Settings: settings,
	}

	if err := s.tenantRepo.Create(ctx, tenant); err != nil {
		return nil, err
	}

	// Add owner as member
	membership := &entity.TenantMembership{
		TenantID: tenant.ID,
		UserID:   input.OwnerID,
		Role:     "owner",
	}
	_ = s.tenantRepo.AddMember(ctx, membership)

	return tenant, nil
}

// generateTenantSlug creates a URL-safe slug from organization name
func (s *TenantService) generateTenantSlug(name string) string {
	slug := ""
	for _, c := range name {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			slug += string(c)
		} else if c >= 'A' && c <= 'Z' {
			slug += string(c - 'A' + 'a')
		} else if c == ' ' || c == '-' {
			if len(slug) > 0 && slug[len(slug)-1] != '-' {
				slug += "-"
			}
		}
	}
	// Trim trailing dash
	if len(slug) > 0 && slug[len(slug)-1] == '-' {
		slug = slug[:len(slug)-1]
	}
	// Ensure slug is unique by appending random suffix
	if len(slug) < 3 {
		slug = "org-" + slug
	}
	return slug
}

// GetTenant retrieves a tenant by ID
func (s *TenantService) GetTenant(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
	tenant, err := s.tenantRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if tenant == nil {
		return nil, apperror.ErrNotFound
	}
	return tenant, nil
}

// GetUserTenants retrieves all tenants a user belongs to
func (s *TenantService) GetUserTenants(ctx context.Context, userID uuid.UUID) ([]entity.Tenant, error) {
	return s.tenantRepo.GetUserTenants(ctx, userID)
}

// UpdateTenantInput represents input for updating a tenant
type UpdateTenantInput struct {
	ID       uuid.UUID
	Name     string
	Settings *entity.TenantSettings
}

// UpdateTenant updates a tenant
func (s *TenantService) UpdateTenant(ctx context.Context, input *UpdateTenantInput) (*entity.Tenant, error) {
	tenant, err := s.tenantRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	if tenant == nil {
		return nil, apperror.ErrNotFound
	}

	if input.Name != "" {
		tenant.Name = input.Name
	}
	if input.Settings != nil {
		tenant.Settings = *input.Settings
	}

	if err := s.tenantRepo.Update(ctx, tenant); err != nil {
		return nil, err
	}

	return tenant, nil
}

// InviteMemberInput represents input for inviting a user to a tenant
type InviteMemberInput struct {
	TenantID uuid.UUID
	UserID   uuid.UUID
	Role     string
}

// InviteMember adds a user to a tenant
func (s *TenantService) InviteMember(ctx context.Context, input *InviteMemberInput) error {
	// Check if user is already a member
	isMember, _ := s.tenantRepo.IsMember(ctx, input.TenantID, input.UserID)
	if isMember {
		return apperror.NewConflictError("User is already a member of this tenant")
	}

	membership := &entity.TenantMembership{
		TenantID: input.TenantID,
		UserID:   input.UserID,
		Role:     input.Role,
	}

	return s.tenantRepo.AddMember(ctx, membership)
}

// RemoveMember removes a user from a tenant
func (s *TenantService) RemoveMember(ctx context.Context, tenantID, userID uuid.UUID) error {
	return s.tenantRepo.RemoveMember(ctx, tenantID, userID)
}

// GetTenantMembers retrieves all members of a tenant
func (s *TenantService) GetTenantMembers(ctx context.Context, tenantID uuid.UUID) ([]entity.TenantMembership, error) {
	members, err := s.tenantRepo.GetMembers(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// Populate user details for JSON response
	for i := range members {
		members[i].PopulateUserDetails()
	}

	return members, nil
}

// UpdateMemberRole updates a member's role in a tenant
func (s *TenantService) UpdateMemberRole(ctx context.Context, tenantID, userID uuid.UUID, role string) error {
	return s.tenantRepo.UpdateMemberRole(ctx, tenantID, userID, role)
}

// ListAllTenants retrieves all tenants (for super admin use)
func (s *TenantService) ListAllTenants(ctx context.Context) ([]entity.Tenant, error) {
	return s.tenantRepo.ListAll(ctx)
}

// AssignUserToTenantInput represents input for assigning a user to a tenant
type AssignUserToTenantInput struct {
	TenantID uuid.UUID
	UserID   uuid.UUID
	Role     string
}

// AssignUserToTenant assigns a user to a tenant (for super admin use)
func (s *TenantService) AssignUserToTenant(ctx context.Context, input *AssignUserToTenantInput) error {
	// Check if tenant exists
	tenant, err := s.tenantRepo.GetByID(ctx, input.TenantID)
	if err != nil {
		return err
	}
	if tenant == nil {
		return apperror.ErrNotFound
	}

	// Check if user is already a member
	isMember, _ := s.tenantRepo.IsMember(ctx, input.TenantID, input.UserID)
	if isMember {
		return apperror.NewConflictError("User is already a member of this tenant")
	}

	// Default role to member if not specified
	role := input.Role
	if role == "" {
		role = "member"
	}

	membership := &entity.TenantMembership{
		TenantID: input.TenantID,
		UserID:   input.UserID,
		Role:     role,
	}

	return s.tenantRepo.AddMember(ctx, membership)
}

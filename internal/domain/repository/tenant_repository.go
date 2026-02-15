package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/domain/entity"
	"github.com/sangkips/investify-api/pkg/pagination"
)

// TenantRepository defines the interface for tenant data operations
type TenantRepository interface {
	// Create creates a new tenant
	Create(ctx context.Context, tenant *entity.Tenant) error

	// GetByID retrieves a tenant by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Tenant, error)

	// GetBySlug retrieves a tenant by slug (subdomain identifier)
	GetBySlug(ctx context.Context, slug string) (*entity.Tenant, error)

	// Update updates an existing tenant
	Update(ctx context.Context, tenant *entity.Tenant) error

	// Delete soft-deletes a tenant
	Delete(ctx context.Context, id uuid.UUID) error

	// GetUserTenants retrieves all tenants a user belongs to with pagination
	GetUserTenants(ctx context.Context, userID uuid.UUID, params *pagination.PaginationParams) ([]entity.Tenant, int64, error)

	// AddMember adds a user as a member of a tenant
	AddMember(ctx context.Context, membership *entity.TenantMembership) error

	// RemoveMember removes a user from a tenant
	RemoveMember(ctx context.Context, tenantID, userID uuid.UUID) error

	// GetMembers retrieves all members of a tenant
	GetMembers(ctx context.Context, tenantID uuid.UUID) ([]entity.TenantMembership, error)

	// IsMember checks if a user is a member of a tenant
	IsMember(ctx context.Context, tenantID, userID uuid.UUID) (bool, error)

	// GetMembership retrieves a specific membership
	GetMembership(ctx context.Context, tenantID, userID uuid.UUID) (*entity.TenantMembership, error)

	// UpdateMemberRole updates a member's role in a tenant
	UpdateMemberRole(ctx context.Context, tenantID, userID uuid.UUID, role string) error

	// SlugExists checks if a slug is already taken
	SlugExists(ctx context.Context, slug string) (bool, error)

	// ListAll retrieves all tenants (for super admin use)
	ListAll(ctx context.Context, params *pagination.PaginationParams) ([]entity.Tenant, int64, error)

	// Count returns the total number of tenants
	Count(ctx context.Context) (int64, error)
}

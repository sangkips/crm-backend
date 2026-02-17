package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/domain/entity"
	domainRepo "github.com/sangkips/investify-api/internal/domain/repository"
	"github.com/sangkips/investify-api/pkg/pagination"
	"gorm.io/gorm"
)

type tenantRepository struct {
	db *gorm.DB
}

// NewTenantRepository creates a new tenant repository
func NewTenantRepository(db *gorm.DB) domainRepo.TenantRepository {
	return &tenantRepository{db: db}
}

func (r *tenantRepository) Create(ctx context.Context, tenant *entity.Tenant) error {
	return r.db.WithContext(ctx).Create(tenant).Error
}

func (r *tenantRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
	var tenant entity.Tenant
	err := r.db.WithContext(ctx).First(&tenant, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &tenant, err
}

func (r *tenantRepository) GetBySlug(ctx context.Context, slug string) (*entity.Tenant, error) {
	var tenant entity.Tenant
	err := r.db.WithContext(ctx).First(&tenant, "slug = ?", slug).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &tenant, err
}

func (r *tenantRepository) Update(ctx context.Context, tenant *entity.Tenant) error {
	return r.db.WithContext(ctx).Save(tenant).Error
}

func (r *tenantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.Tenant{}, "id = ?", id).Error
}

func (r *tenantRepository) GetUserTenants(ctx context.Context, userID uuid.UUID, params *pagination.PaginationParams) ([]entity.Tenant, int64, error) {
	var tenants []entity.Tenant
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Tenant{}).
		Joins("JOIN tenant_memberships ON tenant_memberships.tenant_id = tenants.id").
		Where("tenant_memberships.user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	params.Validate()
	err := query.Offset(params.Offset()).Limit(params.PerPage).Find(&tenants).Error
	return tenants, total, err
}

func (r *tenantRepository) AddMember(ctx context.Context, membership *entity.TenantMembership) error {
	return r.db.WithContext(ctx).Create(membership).Error
}

func (r *tenantRepository) RemoveMember(ctx context.Context, tenantID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Delete(&entity.TenantMembership{}, "tenant_id = ? AND user_id = ?", tenantID, userID).Error
}

func (r *tenantRepository) GetMembers(ctx context.Context, tenantID uuid.UUID) ([]entity.TenantMembership, error) {
	var members []entity.TenantMembership
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("tenant_id = ?", tenantID).
		Find(&members).Error
	return members, err
}

func (r *tenantRepository) IsMember(ctx context.Context, tenantID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.TenantMembership{}).
		Where("tenant_id = ? AND user_id = ?", tenantID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *tenantRepository) GetMembership(ctx context.Context, tenantID, userID uuid.UUID) (*entity.TenantMembership, error) {
	var membership entity.TenantMembership
	err := r.db.WithContext(ctx).
		First(&membership, "tenant_id = ? AND user_id = ?", tenantID, userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &membership, err
}

func (r *tenantRepository) UpdateMemberRole(ctx context.Context, tenantID, userID uuid.UUID, role string) error {
	return r.db.WithContext(ctx).
		Model(&entity.TenantMembership{}).
		Where("tenant_id = ? AND user_id = ?", tenantID, userID).
		Update("role", role).Error
}

func (r *tenantRepository) SlugExists(ctx context.Context, slug string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.Tenant{}).
		Where("slug = ?", slug).
		Count(&count).Error
	return count > 0, err
}

func (r *tenantRepository) ListAll(ctx context.Context, params *pagination.PaginationParams) ([]entity.Tenant, int64, error) {
	var tenants []entity.Tenant
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Tenant{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	params.Validate()
	err := query.Offset(params.Offset()).Limit(params.PerPage).Find(&tenants).Error
	return tenants, total, err
}

func (r *tenantRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Tenant{}).Count(&count).Error
	return count, err
}

func (r *tenantRepository) GetAdminEmails(ctx context.Context, tenantID uuid.UUID) ([]string, error) {
	var emails []string
	err := r.db.WithContext(ctx).
		Model(&entity.TenantMembership{}).
		Select("users.email").
		Joins("JOIN users ON users.id = tenant_memberships.user_id").
		Where("tenant_memberships.tenant_id = ? AND tenant_memberships.role IN ?", tenantID, []string{"owner", "admin"}).
		Scan(&emails).Error
	return emails, err
}

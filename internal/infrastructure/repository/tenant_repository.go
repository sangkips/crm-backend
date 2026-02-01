package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/domain/entity"
	domainRepo "github.com/sangkips/investify-api/internal/domain/repository"
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

func (r *tenantRepository) GetUserTenants(ctx context.Context, userID uuid.UUID) ([]entity.Tenant, error) {
	var tenants []entity.Tenant
	err := r.db.WithContext(ctx).
		Joins("JOIN tenant_memberships ON tenant_memberships.tenant_id = tenants.id").
		Where("tenant_memberships.user_id = ?", userID).
		Find(&tenants).Error
	return tenants, err
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

func (r *tenantRepository) ListAll(ctx context.Context) ([]entity.Tenant, error) {
	var tenants []entity.Tenant
	err := r.db.WithContext(ctx).Find(&tenants).Error
	return tenants, err
}

func (r *tenantRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Tenant{}).Count(&count).Error
	return count, err
}

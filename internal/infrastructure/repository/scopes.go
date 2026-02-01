package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ctxKey string

const (
	// TenantIDKey is the context key for tenant ID
	TenantIDKey ctxKey = "tenant_id"
	// SkipTenantScopeKey is the context key for skipping tenant scope (super admin)
	SkipTenantScopeKey ctxKey = "skip_tenant_scope"
)

// TenantScope returns a GORM scope that filters by tenant
// This should be applied to all queries for tenant-scoped entities
// If SkipTenantScopeKey is true in context (super admin), returns all records
func TenantScope(ctx context.Context) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		// Check if tenant scope should be skipped (super admin)
		if skipScope, ok := ctx.Value(SkipTenantScopeKey).(bool); ok && skipScope {
			return db // Return unfiltered query for super admins
		}

		tenantID, ok := ctx.Value(TenantIDKey).(uuid.UUID)
		if !ok {
			// Fail-safe: return no results if tenant context missing
			// This prevents accidental cross-tenant data access
			return db.Where("1 = 0")
		}
		return db.Where("tenant_id = ?", tenantID)
	}
}

// WithSkipTenantScope adds skip tenant scope flag to context (for super admins)
func WithSkipTenantScope(ctx context.Context, skip bool) context.Context {
	return context.WithValue(ctx, SkipTenantScopeKey, skip)
}

// WithTenant adds tenant ID to context
func WithTenant(ctx context.Context, tenantID uuid.UUID) context.Context {
	return context.WithValue(ctx, TenantIDKey, tenantID)
}

// GetTenantID extracts tenant ID from context
func GetTenantID(ctx context.Context) (uuid.UUID, bool) {
	tenantID, ok := ctx.Value(TenantIDKey).(uuid.UUID)
	return tenantID, ok
}

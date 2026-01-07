package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/domain/entity"
	"github.com/sangkips/investify-api/pkg/pagination"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetByUsername(ctx context.Context, username string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, params *pagination.PaginationParams, search string) ([]entity.User, int64, error)
	GetWithRoles(ctx context.Context, id uuid.UUID) (*entity.User, error)
	AssignRole(ctx context.Context, userID uuid.UUID, roleID uint) error
	RemoveRole(ctx context.Context, userID uuid.UUID, roleID uint) error
}

// RoleRepository defines the interface for role data operations
type RoleRepository interface {
	Create(ctx context.Context, role *entity.Role) error
	GetByID(ctx context.Context, id uint) (*entity.Role, error)
	GetByName(ctx context.Context, name string) (*entity.Role, error)
	Update(ctx context.Context, role *entity.Role) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context) ([]entity.Role, error)
	GetWithPermissions(ctx context.Context, id uint) (*entity.Role, error)
	SyncPermissions(ctx context.Context, roleID uint, permissionIDs []uint) error
}

// PermissionRepository defines the interface for permission data operations
type PermissionRepository interface {
	Create(ctx context.Context, permission *entity.Permission) error
	GetByID(ctx context.Context, id uint) (*entity.Permission, error)
	GetByName(ctx context.Context, name string) (*entity.Permission, error)
	Update(ctx context.Context, permission *entity.Permission) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context) ([]entity.Permission, error)
}

// PasswordResetTokenRepository defines the interface for password reset token operations
type PasswordResetTokenRepository interface {
	Create(ctx context.Context, token *entity.PasswordResetToken) error
	GetByToken(ctx context.Context, token string) (*entity.PasswordResetToken, error)
	MarkAsUsed(ctx context.Context, token string) error
	DeleteByEmail(ctx context.Context, email string) error
	DeleteExpired(ctx context.Context) error
}

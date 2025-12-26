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

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) domainRepo.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).First(&user, "email = ?", email).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).First(&user, "username = ?", username).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.User{}, "id = ?", id).Error
}

func (r *userRepository) List(ctx context.Context, params *pagination.PaginationParams, search string) ([]entity.User, int64, error) {
	var users []entity.User
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.User{})

	if search != "" {
		query = query.Where("name ILIKE ? OR email ILIKE ? OR username ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	params.Validate()
	err := query.Offset(params.Offset()).Limit(params.PerPage).
		Preload("Roles").
		Order("created_at DESC").
		Find(&users).Error

	return users, total, err
}

func (r *userRepository) GetWithRoles(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).
		Preload("Roles.Permissions").
		First(&user, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

func (r *userRepository) AssignRole(ctx context.Context, userID uuid.UUID, roleID uint) error {
	return r.db.WithContext(ctx).Exec(
		"INSERT INTO model_has_roles (model_id, role_id) VALUES (?, ?) ON CONFLICT DO NOTHING",
		userID, roleID,
	).Error
}

func (r *userRepository) RemoveRole(ctx context.Context, userID uuid.UUID, roleID uint) error {
	return r.db.WithContext(ctx).Exec(
		"DELETE FROM model_has_roles WHERE model_id = ? AND role_id = ?",
		userID, roleID,
	).Error
}

type roleRepository struct {
	db *gorm.DB
}

// NewRoleRepository creates a new role repository
func NewRoleRepository(db *gorm.DB) domainRepo.RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) Create(ctx context.Context, role *entity.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *roleRepository) GetByID(ctx context.Context, id uint) (*entity.Role, error) {
	var role entity.Role
	err := r.db.WithContext(ctx).First(&role, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &role, err
}

func (r *roleRepository) GetByName(ctx context.Context, name string) (*entity.Role, error) {
	var role entity.Role
	err := r.db.WithContext(ctx).First(&role, "name = ?", name).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &role, err
}

func (r *roleRepository) Update(ctx context.Context, role *entity.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *roleRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.Role{}, id).Error
}

func (r *roleRepository) List(ctx context.Context) ([]entity.Role, error) {
	var roles []entity.Role
	err := r.db.WithContext(ctx).Preload("Permissions").Find(&roles).Error
	return roles, err
}

func (r *roleRepository) GetWithPermissions(ctx context.Context, id uint) (*entity.Role, error) {
	var role entity.Role
	err := r.db.WithContext(ctx).Preload("Permissions").First(&role, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &role, err
}

func (r *roleRepository) SyncPermissions(ctx context.Context, roleID uint, permissionIDs []uint) error {
	role := &entity.Role{ID: roleID}
	var permissions []entity.Permission
	if len(permissionIDs) > 0 {
		if err := r.db.WithContext(ctx).Find(&permissions, permissionIDs).Error; err != nil {
			return err
		}
	}
	return r.db.WithContext(ctx).Model(role).Association("Permissions").Replace(permissions)
}

type permissionRepository struct {
	db *gorm.DB
}

// NewPermissionRepository creates a new permission repository
func NewPermissionRepository(db *gorm.DB) domainRepo.PermissionRepository {
	return &permissionRepository{db: db}
}

func (r *permissionRepository) Create(ctx context.Context, permission *entity.Permission) error {
	return r.db.WithContext(ctx).Create(permission).Error
}

func (r *permissionRepository) GetByID(ctx context.Context, id uint) (*entity.Permission, error) {
	var permission entity.Permission
	err := r.db.WithContext(ctx).First(&permission, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &permission, err
}

func (r *permissionRepository) GetByName(ctx context.Context, name string) (*entity.Permission, error) {
	var permission entity.Permission
	err := r.db.WithContext(ctx).First(&permission, "name = ?", name).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &permission, err
}

func (r *permissionRepository) Update(ctx context.Context, permission *entity.Permission) error {
	return r.db.WithContext(ctx).Save(permission).Error
}

func (r *permissionRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.Permission{}, id).Error
}

func (r *permissionRepository) List(ctx context.Context) ([]entity.Permission, error) {
	var permissions []entity.Permission
	err := r.db.WithContext(ctx).Order("name ASC").Find(&permissions).Error
	return permissions, err
}

package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	FirstName       string         `gorm:"size:255;not null" json:"first_name"`
	LastName        string         `gorm:"size:255;not null" json:"last_name"`
	Username        string         `gorm:"size:255;unique" json:"username"`
	Email           string         `gorm:"size:255;unique;not null" json:"email"`
	Password        string         `gorm:"size:255" json:"-"`
	Provider        string         `gorm:"size:50;default:'local'" json:"provider"`
	ProviderID      *string        `gorm:"size:255" json:"-"`
	Photo           *string        `gorm:"size:255" json:"photo,omitempty"`
	StoreName       *string        `gorm:"size:255" json:"store_name,omitempty"`
	StoreAddress    *string        `gorm:"type:text" json:"store_address,omitempty"`
	StorePhone      *string        `gorm:"size:50" json:"store_phone,omitempty"`
	StoreEmail      *string        `gorm:"size:255" json:"store_email,omitempty"`
	EmailVerifiedAt *time.Time     `json:"email_verified_at,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Roles      []Role      `gorm:"many2many:model_has_roles;foreignKey:ID;joinForeignKey:model_id;References:ID;joinReferences:role_id" json:"roles,omitempty"`
	Products   []Product   `gorm:"foreignKey:UserID" json:"-"`
	Orders     []Order     `gorm:"foreignKey:UserID" json:"-"`
	Purchases  []Purchase  `gorm:"foreignKey:UserID" json:"-"`
	Quotations []Quotation `gorm:"foreignKey:UserID" json:"-"`
	Customers  []Customer  `gorm:"foreignKey:UserID" json:"-"`
	Suppliers  []Supplier  `gorm:"foreignKey:UserID" json:"-"`
}

// BeforeCreate generates a UUID before creating a new user
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// TableName returns the table name for the User model
func (User) TableName() string {
	return "users"
}

// Role represents a role in the RBAC system
type Role struct {
	ID          uint         `gorm:"primary_key" json:"id"`
	Name        string       `gorm:"size:255;not null" json:"name"`
	GuardName   string       `gorm:"size:255;default:'web'" json:"guard_name"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	Permissions []Permission `gorm:"many2many:role_has_permissions;foreignKey:ID;joinForeignKey:role_id;References:ID;joinReferences:permission_id" json:"permissions,omitempty"`
}

// TableName returns the table name for the Role model
func (Role) TableName() string {
	return "roles"
}

// Permission represents a permission in the RBAC system
type Permission struct {
	ID        uint      `gorm:"primary_key" json:"id"`
	Name      string    `gorm:"size:255;not null" json:"name"`
	GuardName string    `gorm:"size:255;default:'web'" json:"guard_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName returns the table name for the Permission model
func (Permission) TableName() string {
	return "permissions"
}

// HasPermission checks if the user has a specific permission
func (u *User) HasPermission(permissionName string) bool {
	for _, role := range u.Roles {
		for _, permission := range role.Permissions {
			if permission.Name == permissionName {
				return true
			}
		}
	}
	return false
}

// HasRole checks if the user has a specific role
func (u *User) HasRole(roleName string) bool {
	for _, role := range u.Roles {
		if role.Name == roleName {
			return true
		}
	}
	return false
}

// GetPermissions returns all permission names for the user
func (u *User) GetPermissions() []string {
	permissions := make(map[string]bool)
	for _, role := range u.Roles {
		for _, permission := range role.Permissions {
			permissions[permission.Name] = true
		}
	}

	result := make([]string, 0, len(permissions))
	for p := range permissions {
		result = append(result, p)
	}
	return result
}

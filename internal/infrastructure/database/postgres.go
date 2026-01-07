package database

import (
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/config"
	"github.com/sangkips/investify-api/internal/domain/entity"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	logLevel := logger.Info

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  cfg.DSN(),
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL DB to set connection pool settings
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	log.Println("Successfully connected to PostgreSQL database")
	return db, nil
}

// AutoMigrate runs GORM auto-migration for all entities
func AutoMigrate(db *gorm.DB) error {
	log.Println("Running database migrations...")

	err := db.AutoMigrate(
		// User-related entities
		&entity.User{},
		&entity.Role{},
		&entity.Permission{},
		&entity.PasswordResetToken{},

		// Product-related entities
		&entity.Category{},
		&entity.Unit{},
		&entity.Product{},

		// CRM entities
		&entity.Customer{},
		&entity.Supplier{},

		// Transaction entities
		&entity.Order{},
		&entity.OrderDetail{},
		&entity.Purchase{},
		&entity.PurchaseDetail{},
		&entity.Quotation{},
		&entity.QuotationDetail{},

		// System entities
		&entity.IdempotencyKey{},
		&entity.UserSettings{},
	)

	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// SeedDefaultData seeds the database with default data (roles, permissions, admin user)
func SeedDefaultData(db *gorm.DB) error {
	log.Println("Seeding default data...")

	// Create default permissions
	permissions := []entity.Permission{
		{Name: "view-dashboard", GuardName: "web"},
		{Name: "manage-products", GuardName: "web"},
		{Name: "manage-orders", GuardName: "web"},
		{Name: "manage-purchases", GuardName: "web"},
		{Name: "manage-quotations", GuardName: "web"},
		{Name: "manage-customers", GuardName: "web"},
		{Name: "manage-suppliers", GuardName: "web"},
		{Name: "manage-categories", GuardName: "web"},
		{Name: "manage-units", GuardName: "web"},
		{Name: "manage-users", GuardName: "web"},
		{Name: "view-reports", GuardName: "web"},
	}

	for i := range permissions {
		var existing entity.Permission
		if err := db.Where("name = ?", permissions[i].Name).First(&existing).Error; err != nil {
			if err := db.Create(&permissions[i]).Error; err != nil {
				log.Printf("Warning: failed to create permission %s: %v", permissions[i].Name, err)
			}
		}
	}

	// Reload permissions with IDs
	var allPermissions []entity.Permission
	db.Find(&allPermissions)

	// Create super-admin role with all permissions
	var superAdminRole entity.Role
	if err := db.Where("name = ?", "super-admin").First(&superAdminRole).Error; err != nil {
		superAdminRole = entity.Role{
			Name:        "super-admin",
			GuardName:   "web",
			Permissions: allPermissions,
		}
		if err := db.Create(&superAdminRole).Error; err != nil {
			log.Printf("Warning: failed to create super-admin role: %v", err)
		}
	}

	// Create admin role with all permissions
	var adminRole entity.Role
	if err := db.Where("name = ?", "admin").First(&adminRole).Error; err != nil {
		adminRole = entity.Role{
			Name:        "admin",
			GuardName:   "web",
			Permissions: allPermissions,
		}
		if err := db.Create(&adminRole).Error; err != nil {
			log.Printf("Warning: failed to create admin role: %v", err)
		}
	}

	// Create staff role with limited permissions
	staffPermissions := []string{
		"view-dashboard",
		"manage-products",
		"manage-orders",
		"manage-customers",
	}
	var staffPerms []entity.Permission
	for _, name := range staffPermissions {
		for _, p := range allPermissions {
			if p.Name == name {
				staffPerms = append(staffPerms, p)
				break
			}
		}
	}

	var staffRole entity.Role
	if err := db.Where("name = ?", "staff").First(&staffRole).Error; err != nil {
		staffRole = entity.Role{
			Name:        "staff",
			GuardName:   "web",
			Permissions: staffPerms,
		}
		if err := db.Create(&staffRole).Error; err != nil {
			log.Printf("Warning: failed to create staff role: %v", err)
		}
	}

	// Create default user role with basic permissions (for new registrants)
	userPermissions := []string{
		"view-dashboard",
		// "manage-products",
		// "manage-orders",
		// "manage-purchases",
		"manage-customers",
		"manage-suppliers",
		"manage-categories",
		"manage-units",
		// "view-reports",
	}
	var userPerms []entity.Permission
	for _, name := range userPermissions {
		for _, p := range allPermissions {
			if p.Name == name {
				userPerms = append(userPerms, p)
				break
			}
		}
	}

	var userRole entity.Role
	if err := db.Where("name = ?", "user").First(&userRole).Error; err != nil {
		userRole = entity.Role{
			Name:        "user",
			GuardName:   "web",
			Permissions: userPerms,
		}
		if err := db.Create(&userRole).Error; err != nil {
			log.Printf("Warning: failed to create user role: %v", err)
		}
	}

	// Create super admin user if configured via environment variables
	adminEmail := viper.GetString("ADMIN_EMAIL")
	adminPassword := viper.GetString("ADMIN_PASSWORD")
	adminName := viper.GetString("ADMIN_NAME")

	if adminEmail != "" && adminPassword != "" {
		var existingAdmin entity.User
		if err := db.Where("email = ?", adminEmail).First(&existingAdmin).Error; err != nil {
			// Hash the password
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
			if err != nil {
				log.Printf("Warning: failed to hash admin password: %v", err)
			} else {
				// Get super-admin role
				var saRole entity.Role
				if err := db.Where("name = ?", "super-admin").First(&saRole).Error; err == nil {
					if adminName == "" {
						adminName = "Super Admin"
					}
					// Split admin name into first and last name
					firstName := adminName
					lastName := ""
					for i, c := range adminName {
						if c == ' ' {
							firstName = adminName[:i]
							lastName = adminName[i+1:]
							break
						}
					}
					adminUser := entity.User{
						ID:        uuid.New(),
						FirstName: firstName,
						LastName:  lastName,
						Email:     adminEmail,
						Password:  string(hashedPassword),
						Roles:     []entity.Role{saRole},
					}
					if err := db.Create(&adminUser).Error; err != nil {
						log.Printf("Warning: failed to create super admin user: %v", err)
					} else {
						log.Printf("Super admin user created: %s", adminEmail)
					}
				}
			}
		} else {
			log.Printf("Super admin user already exists: %s", adminEmail)
		}
	}

	log.Println("Default data seeding completed")
	return nil
}

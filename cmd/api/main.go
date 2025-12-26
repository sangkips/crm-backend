package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sangkips/investify-api/internal/application/service"
	"github.com/sangkips/investify-api/internal/config"
	"github.com/sangkips/investify-api/internal/infrastructure/database"
	"github.com/sangkips/investify-api/internal/infrastructure/repository"
	"github.com/sangkips/investify-api/internal/presentation/http/handler"
	"github.com/sangkips/investify-api/internal/presentation/http/middleware"
	"github.com/sangkips/investify-api/pkg/utils"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Set Gin mode based on environment
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Connect to database
	db, err := database.NewPostgresDB(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run auto-migrations
	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Seed default data
	if err := database.SeedDefaultData(db); err != nil {
		log.Printf("Warning: Failed to seed default data: %v", err)
	}

	// Initialize JWT manager
	jwtManager := utils.NewJWTManager(
		cfg.JWT.Secret,
		cfg.JWT.ExpiryHours,
		cfg.JWT.RefreshExpiryHours,
	)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	productRepo := repository.NewProductRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	unitRepo := repository.NewUnitRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	orderDetailRepo := repository.NewOrderDetailRepository(db)
	purchaseRepo := repository.NewPurchaseRepository(db)
	purchaseDetailRepo := repository.NewPurchaseDetailRepository(db)
	customerRepo := repository.NewCustomerRepository(db)
	supplierRepo := repository.NewSupplierRepository(db)
	idempotencyRepo := repository.NewIdempotencyRepository(db)
	quotationRepo := repository.NewQuotationRepository(db)
	quotationDetailRepo := repository.NewQuotationDetailRepository(db)
	settingsRepo := repository.NewSettingsRepository(db)
	permissionRepo := repository.NewPermissionRepository(db)

	// Initialize services
	authService := service.NewAuthService(userRepo, roleRepo, jwtManager)
	productService := service.NewProductService(productRepo, categoryRepo, unitRepo)
	categoryService := service.NewCategoryService(categoryRepo)
	unitService := service.NewUnitService(unitRepo)
	orderService := service.NewOrderService(orderRepo, orderDetailRepo, productRepo, customerRepo)
	purchaseService := service.NewPurchaseService(purchaseRepo, purchaseDetailRepo, productRepo, supplierRepo)
	customerService := service.NewCustomerService(customerRepo)
	supplierService := service.NewSupplierService(supplierRepo)
	dashboardService := service.NewDashboardService(orderRepo, purchaseRepo, productRepo, customerRepo)
	quotationService := service.NewQuotationService(quotationRepo, quotationDetailRepo, productRepo, customerRepo)
	settingsService := service.NewSettingsService(settingsRepo)
	userService := service.NewUserService(userRepo, roleRepo, permissionRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	productHandler := handler.NewProductHandler(productService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	unitHandler := handler.NewUnitHandler(unitService)
	orderHandler := handler.NewOrderHandler(orderService)
	purchaseHandler := handler.NewPurchaseHandler(purchaseService)
	customerHandler := handler.NewCustomerHandler(customerService)
	supplierHandler := handler.NewSupplierHandler(supplierService)
	dashboardHandler := handler.NewDashboardHandler(dashboardService)
	quotationHandler := handler.NewQuotationHandler(quotationService)
	settingsHandler := handler.NewSettingsHandler(settingsService)
	userHandler := handler.NewUserHandler(userService)

	// Create Gin router
	router := gin.New()

	// Add global middleware
	router.Use(gin.Recovery())
	router.Use(middleware.LoggerMiddleware())
	router.Use(middleware.CORSMiddleware(&cfg.CORS))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": cfg.App.Name,
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Public routes (no authentication required)
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/register", authHandler.Register)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/forgot-password", authHandler.ForgotPassword)
			auth.POST("/reset-password", authHandler.ResetPassword)
		}

		// Protected routes (authentication required)
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(jwtManager))
		{
			// Auth/Profile routes
			protected.POST("/auth/logout", authHandler.Logout)
			protected.GET("/profile", authHandler.GetProfile)
			protected.PUT("/profile", authHandler.UpdateProfile)
			protected.PUT("/profile/password", authHandler.ChangePassword)

			// Settings
			protected.GET("/settings", settingsHandler.GetSettings)
			protected.PUT("/settings", settingsHandler.UpdateSettings)

			// Dashboard
			protected.GET("/dashboard", dashboardHandler.GetStats)

			// Products
			products := protected.Group("/products")
			products.Use(middleware.RequirePermission("manage-products"))
			{
				products.GET("", productHandler.List)
				products.POST("", productHandler.Create)
				products.GET("/low-stock", productHandler.GetLowStock)
				products.GET("/:slug", productHandler.Get)
				products.PUT("/:slug", productHandler.Update)
				products.DELETE("/:slug", productHandler.Delete)
			}

			// Categories
			categories := protected.Group("/categories")
			categories.Use(middleware.RequirePermission("manage-categories"))
			{
				categories.GET("", categoryHandler.List)
				categories.POST("", categoryHandler.Create)
				categories.PUT("/:id", categoryHandler.Update)
				categories.DELETE("/:id", categoryHandler.Delete)
			}

			// Units
			units := protected.Group("/units")
			units.Use(middleware.RequirePermission("manage-units"))
			{
				units.GET("", unitHandler.List)
				units.POST("", unitHandler.Create)
				units.PUT("/:id", unitHandler.Update)
				units.DELETE("/:id", unitHandler.Delete)
			}

			// Orders
			orders := protected.Group("/orders")
			orders.Use(middleware.RequirePermission("manage-orders"))
			{
				orders.GET("", orderHandler.List)
				// Order creation uses idempotency middleware to prevent duplicates
				orders.POST("", middleware.IdempotencyRequired(middleware.IdempotencyConfig{
					Repo: idempotencyRepo,
				}), orderHandler.Create)
				orders.GET("/due", orderHandler.GetDueOrders)
				orders.GET("/:id", orderHandler.Get)
				orders.PUT("/:id/status", orderHandler.UpdateStatus)
				orders.POST("/:id/cancel", orderHandler.Cancel)
				orders.POST("/:id/pay", orderHandler.PayDue)
			}

			// Purchases
			purchases := protected.Group("/purchases")
			purchases.Use(middleware.RequirePermission("manage-purchases"))
			{
				purchases.GET("", purchaseHandler.List)
				purchases.POST("", purchaseHandler.Create)
				purchases.GET("/pending", purchaseHandler.GetPending)
				purchases.GET("/:id", purchaseHandler.Get)
				purchases.POST("/:id/approve", purchaseHandler.Approve)
				purchases.DELETE("/:id", purchaseHandler.Delete)
			}

			// Customers
			customers := protected.Group("/customers")
			customers.Use(middleware.RequirePermission("manage-customers"))
			{
				customers.GET("", customerHandler.List)
				customers.POST("", customerHandler.Create)
				customers.GET("/:id", customerHandler.Get)
				customers.PUT("/:id", customerHandler.Update)
				customers.DELETE("/:id", customerHandler.Delete)
			}

			// Suppliers
			suppliers := protected.Group("/suppliers")
			suppliers.Use(middleware.RequirePermission("manage-suppliers"))
			{
				suppliers.GET("", supplierHandler.List)
				suppliers.POST("", supplierHandler.Create)
				suppliers.GET("/:id", supplierHandler.Get)
				suppliers.PUT("/:id", supplierHandler.Update)
				suppliers.DELETE("/:id", supplierHandler.Delete)
			}

			// Quotations
			quotations := protected.Group("/quotations")
			quotations.Use(middleware.RequirePermission("manage-quotations"))
			{
				quotations.GET("", quotationHandler.List)
				quotations.POST("", quotationHandler.Create)
				quotations.GET("/:id", quotationHandler.Get)
				quotations.PUT("/:id", quotationHandler.Update)
				quotations.DELETE("/:id", quotationHandler.Delete)
			}

			// Reports (placeholder for now)
			reports := protected.Group("/reports")
			reports.Use(middleware.RequirePermission("view-reports"))
			{
				reports.GET("/orders", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Orders report - Coming soon"})
				})
				reports.GET("/purchases", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Purchases report - Coming soon"})
				})
				reports.GET("/products", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Products report - Coming soon"})
				})
			}

			// Users (Admin)
			users := protected.Group("/users")
			users.Use(middleware.RequirePermission("manage-users"))
			{
				users.GET("", userHandler.List)
				users.GET("/:id", userHandler.Get)
				users.PUT("/:id/roles", userHandler.UpdateRoles)
				users.DELETE("/:id", userHandler.Delete)
			}

			// Roles (Admin)
			roles := protected.Group("/roles")
			roles.Use(middleware.RequirePermission("manage-users"))
			{
				roles.GET("", userHandler.ListRoles)
			}

			// Permissions (Admin)
			permissions := protected.Group("/permissions")
			permissions.Use(middleware.RequirePermission("manage-users"))
			{
				permissions.GET("", userHandler.ListPermissions)
			}
		}
	}

	// Get port from environment or use default
	port := cfg.App.Port
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting %s server on port %s...", cfg.App.Name, port)
	log.Printf("Environment: %s", cfg.App.Env)

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
		os.Exit(1)
	}
}

package main

import (
	"log"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"github.com/sangkips/investify-api/internal/application/service"
	"github.com/sangkips/investify-api/internal/config"
	"github.com/sangkips/investify-api/internal/infrastructure/database"
	"github.com/sangkips/investify-api/internal/infrastructure/repository"
	"github.com/sangkips/investify-api/internal/presentation/http/handler"
	"github.com/sangkips/investify-api/internal/presentation/http/routes"
	"github.com/sangkips/investify-api/pkg/email"
	"github.com/sangkips/investify-api/pkg/oauth"
	"github.com/sangkips/investify-api/pkg/printer"
	"github.com/sangkips/investify-api/pkg/utils"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// GlitchTip is Sentry-protocol compatible; sentry-go sends events to your DSN.
	glitchTipEnabled := false
	if cfg.ErrorReporting.DSN != "" {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:              cfg.ErrorReporting.DSN,
			Environment:      cfg.App.Env,
			Release:          cfg.ErrorReporting.Release,
			EnableTracing:    false,
			TracesSampleRate: 0,
		}); err != nil {
			log.Printf("Warning: error reporting (GlitchTip) init failed: %v", err)
		} else {
			glitchTipEnabled = true
			defer sentry.Flush(5 * time.Second)
			log.Printf("Error reporting enabled (environment=%s)", cfg.App.Env)
		}
	}

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
	tenantRepo := repository.NewTenantRepository(db)
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
	analyticsRepo := repository.NewAnalyticsRepository(db)
	passwordResetRepo := repository.NewPasswordResetTokenRepository(db)
	emailVerificationRepo := repository.NewEmailVerificationTokenRepository(db)
	mpesaTxRepo := repository.NewMpesaTransactionRepository(db)

	// Initialize email service
	emailService := email.NewEmailService(email.EmailConfig{
		SMTPHost:     cfg.Email.SMTPHost,
		SMTPPort:     cfg.Email.SMTPPort,
		SMTPUsername: cfg.Email.SMTPUsername,
		SMTPPassword: cfg.Email.SMTPPassword,
		FromName:     cfg.Email.FromName,
		FromEmail:    cfg.Email.FromEmail,
		FrontendURL:  cfg.Email.FrontendURL,
	})

	// Initialize Google OAuth service
	googleOAuthService := oauth.NewGoogleOAuthService(oauth.GoogleOAuthConfig{
		ClientID:           cfg.OAuth.GoogleClientID,
		ClientSecret:       cfg.OAuth.GoogleClientSecret,
		RedirectURL:        cfg.OAuth.GoogleRedirectURL,
		FrontendSuccessURL: cfg.OAuth.FrontendSuccessURL,
		FrontendErrorURL:   cfg.OAuth.FrontendErrorURL,
	})

	// Initialize services
	authService := service.NewAuthService(userRepo, roleRepo, tenantRepo, passwordResetRepo, emailVerificationRepo, jwtManager, emailService, googleOAuthService)
	tenantService := service.NewTenantService(tenantRepo)
	productService := service.NewProductService(productRepo, categoryRepo, unitRepo)
	categoryService := service.NewCategoryService(categoryRepo)
	unitService := service.NewUnitService(unitRepo)
	orderService := service.NewOrderService(orderRepo, orderDetailRepo, productRepo, customerRepo, emailService, tenantRepo)
	purchaseService := service.NewPurchaseService(purchaseRepo, purchaseDetailRepo, productRepo, supplierRepo)
	customerService := service.NewCustomerService(customerRepo)
	supplierService := service.NewSupplierService(supplierRepo)
	dashboardService := service.NewDashboardService(orderRepo, purchaseRepo, productRepo, customerRepo, analyticsRepo, tenantRepo)
	quotationService := service.NewQuotationService(quotationRepo, quotationDetailRepo, productRepo, customerRepo)
	settingsService := service.NewSettingsService(settingsRepo)
	userService := service.NewUserService(userRepo, roleRepo, permissionRepo)
	mpesaService := service.NewMpesaService(mpesaTxRepo, tenantRepo, orderRepo, orderService)

	// Initialize thermal printer
	thermalPrinter, err := printer.NewPrinterFromConfig(
		cfg.Printer.Type,
		cfg.Printer.USBPath,
		cfg.Printer.Address,
	)
	if err != nil {
		log.Printf("Warning: Failed to initialize printer: %v", err)
		thermalPrinter = printer.NewNullPrinter()
	}
	printerService := service.NewPrinterService(thermalPrinter, orderRepo, quotationRepo, cfg.Printer.Type)

	// Initialize handlers
	handlers := &routes.Handlers{
		Auth:      handler.NewAuthHandler(authService),
		Tenant:    handler.NewTenantHandler(tenantService),
		Product:   handler.NewProductHandler(productService),
		Category:  handler.NewCategoryHandler(categoryService),
		Unit:      handler.NewUnitHandler(unitService),
		Order:     handler.NewOrderHandler(orderService, mpesaService),
		Purchase:  handler.NewPurchaseHandler(purchaseService),
		Customer:  handler.NewCustomerHandler(customerService),
		Supplier:  handler.NewSupplierHandler(supplierService),
		Dashboard: handler.NewDashboardHandler(dashboardService),
		Quotation: handler.NewQuotationHandler(quotationService),
		Settings:  handler.NewSettingsHandler(settingsService),
		User:      handler.NewUserHandler(userService),
		Printer:   handler.NewPrinterHandler(printerService),
		Mpesa:     handler.NewMpesaHandler(mpesaService),
	}

	// Setup routes
	router := routes.Setup(handlers, &routes.Deps{
		JWTManager:       jwtManager,
		Cfg:              cfg,
		IdempotencyRepo:  idempotencyRepo,
		GlitchTipEnabled: glitchTipEnabled,
	})

	// Get port from environment or use default
	port := cfg.App.Port
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting %s server on port %s...", cfg.App.Name, port)
	log.Printf("Environment: %s", cfg.App.Env)

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

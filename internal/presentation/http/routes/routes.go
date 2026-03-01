package routes

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sangkips/investify-api/internal/config"
	domainRepo "github.com/sangkips/investify-api/internal/domain/repository"
	"github.com/sangkips/investify-api/internal/presentation/http/handler"
	"github.com/sangkips/investify-api/internal/presentation/http/middleware"
	"github.com/sangkips/investify-api/pkg/utils"
)

// Handlers holds all the HTTP handlers used for route registration.
type Handlers struct {
	Auth      *handler.AuthHandler
	Tenant    *handler.TenantHandler
	Product   *handler.ProductHandler
	Category  *handler.CategoryHandler
	Unit      *handler.UnitHandler
	Order     *handler.OrderHandler
	Purchase  *handler.PurchaseHandler
	Customer  *handler.CustomerHandler
	Supplier  *handler.SupplierHandler
	Dashboard *handler.DashboardHandler
	Quotation *handler.QuotationHandler
	Settings  *handler.SettingsHandler
	User      *handler.UserHandler
	Printer   *handler.PrinterHandler
}

// Deps holds shared dependencies needed by the routes.
type Deps struct {
	JWTManager      *utils.JWTManager
	Cfg             *config.Config
	IdempotencyRepo domainRepo.IdempotencyRepository
}

// Setup creates the Gin router and registers all routes.
func Setup(h *Handlers, deps *Deps) *gin.Engine {
	router := gin.New()

	// Global middleware
	router.Use(gin.Recovery())
	router.Use(middleware.LoggerMiddleware())
	router.Use(middleware.CORSMiddleware(&deps.Cfg.CORS))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": deps.Cfg.App.Name,
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Public routes (no authentication required)
		registerAuthRoutes(v1, h)

		// Protected routes (authentication required)
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(deps.JWTManager))

		// Per-tenant rate limiter
		rateLimiter := middleware.NewTenantRateLimiter(middleware.RateLimiterConfig{
			RequestsPerSecond: float64(deps.Cfg.RateLimit.Requests) / float64(deps.Cfg.RateLimit.Duration),
			BurstSize:         deps.Cfg.RateLimit.Requests,
			CleanupInterval:   5 * time.Minute,
			EntryTTL:          10 * time.Minute,
		})
		protected.Use(rateLimiter.Middleware())

		registerProtectedRoutes(protected, h, deps)
	}

	return router
}

func registerAuthRoutes(v1 *gin.RouterGroup, h *Handlers) {
	auth := v1.Group("/auth")
	{
		auth.POST("/login", h.Auth.Login)
		auth.POST("/register", h.Auth.Register)
		auth.POST("/refresh", h.Auth.RefreshToken)
		auth.POST("/forgot-password", h.Auth.ForgotPassword)
		auth.POST("/reset-password", h.Auth.ResetPassword)
		// Google OAuth routes
		auth.GET("/google", h.Auth.GoogleAuth)
		auth.GET("/google/callback", h.Auth.GoogleCallback)
	}
}

func registerProtectedRoutes(protected *gin.RouterGroup, h *Handlers, deps *Deps) {
	// Auth/Profile routes
	protected.POST("/auth/logout", h.Auth.Logout)
	protected.GET("/profile", h.Auth.GetProfile)
	protected.PUT("/profile", h.Auth.UpdateProfile)
	protected.PUT("/profile/password", h.Auth.ChangePassword)

	// Settings
	protected.GET("/settings", h.Settings.GetSettings)
	protected.PUT("/settings", h.Settings.UpdateSettings)

	// Dashboard
	protected.GET("/dashboard", h.Dashboard.GetStats)

	// Tenants
	registerTenantRoutes(protected, h)

	// Products
	registerProductRoutes(protected, h)

	// Categories
	registerCategoryRoutes(protected, h)

	// Units
	registerUnitRoutes(protected, h)

	// Orders
	registerOrderRoutes(protected, h, deps)

	// Purchases
	registerPurchaseRoutes(protected, h)

	// Customers
	registerCustomerRoutes(protected, h)

	// Suppliers
	registerSupplierRoutes(protected, h)

	// Quotations
	registerQuotationRoutes(protected, h)

	// Reports
	registerReportRoutes(protected)

	// Users (Admin)
	registerUserRoutes(protected, h)

	// Roles (Admin)
	registerRoleRoutes(protected, h)

	// Permissions (Admin)
	registerPermissionRoutes(protected, h)

	// Super Admin routes
	registerAdminRoutes(protected, h)

	// Printer
	registerPrinterRoutes(protected, h)
}

func registerTenantRoutes(protected *gin.RouterGroup, h *Handlers) {
	tenants := protected.Group("/tenants")
	{
		tenants.GET("", h.Tenant.ListTenants)
		tenants.POST("", h.Tenant.Create)
		tenants.GET("/current", h.Tenant.GetCurrentTenant)
		tenants.PUT("/current", h.Tenant.UpdateTenant)
		tenants.GET("/current/members", h.Tenant.ListMembers)
		tenants.POST("/current/members", h.Tenant.InviteMember)
		tenants.PUT("/current/members/:user_id", h.Tenant.UpdateMemberRole)
		tenants.DELETE("/current/members/:user_id", h.Tenant.RemoveMember)
	}
}

func registerProductRoutes(protected *gin.RouterGroup, h *Handlers) {
	products := protected.Group("/products")
	products.Use(middleware.RequirePermission("manage-products"))
	{
		products.GET("", h.Product.List)
		products.POST("", h.Product.Create)
		products.POST("/import", h.Product.ImportProducts)
		products.GET("/low-stock", h.Product.GetLowStock)
		products.GET("/:slug", h.Product.Get)
		products.PUT("/:slug", h.Product.Update)
		products.DELETE("/:slug", h.Product.Delete)
	}
}

func registerCategoryRoutes(protected *gin.RouterGroup, h *Handlers) {
	categories := protected.Group("/categories")
	categories.Use(middleware.RequirePermission("manage-categories"))
	{
		categories.GET("", h.Category.List)
		categories.POST("", h.Category.Create)
		categories.PUT("/:id", h.Category.Update)
		categories.DELETE("/:id", h.Category.Delete)
	}
}

func registerUnitRoutes(protected *gin.RouterGroup, h *Handlers) {
	units := protected.Group("/units")
	units.Use(middleware.RequirePermission("manage-units"))
	{
		units.GET("", h.Unit.List)
		units.POST("", h.Unit.Create)
		units.PUT("/:id", h.Unit.Update)
		units.DELETE("/:id", h.Unit.Delete)
	}
}

func registerOrderRoutes(protected *gin.RouterGroup, h *Handlers, deps *Deps) {
	orders := protected.Group("/orders")
	orders.Use(middleware.RequirePermission("manage-orders"))
	{
		orders.GET("", h.Order.List)
		// Order creation uses idempotency middleware to prevent duplicates
		orders.POST("", middleware.IdempotencyRequired(middleware.IdempotencyConfig{
			Repo: deps.IdempotencyRepo,
		}), h.Order.Create)
		orders.GET("/due", h.Order.GetDueOrders)
		orders.GET("/:id", h.Order.Get)
		orders.PUT("/:id/status", h.Order.UpdateStatus)
		orders.POST("/:id/cancel", h.Order.Cancel)
		orders.POST("/:id/pay", h.Order.PayDue)
	}
}

func registerPurchaseRoutes(protected *gin.RouterGroup, h *Handlers) {
	purchases := protected.Group("/purchases")
	purchases.Use(middleware.RequirePermission("manage-purchases"))
	{
		purchases.GET("", h.Purchase.List)
		purchases.POST("", h.Purchase.Create)
		purchases.GET("/pending", h.Purchase.GetPending)
		purchases.GET("/:id", h.Purchase.Get)
		purchases.POST("/:id/approve", h.Purchase.Approve)
		purchases.DELETE("/:id", h.Purchase.Delete)
	}
}

func registerCustomerRoutes(protected *gin.RouterGroup, h *Handlers) {
	customers := protected.Group("/customers")
	customers.Use(middleware.RequirePermission("manage-customers"))
	{
		customers.GET("", h.Customer.List)
		customers.POST("", h.Customer.Create)
		customers.GET("/:id", h.Customer.Get)
		customers.PUT("/:id", h.Customer.Update)
		customers.DELETE("/:id", h.Customer.Delete)
	}
}

func registerSupplierRoutes(protected *gin.RouterGroup, h *Handlers) {
	suppliers := protected.Group("/suppliers")
	suppliers.Use(middleware.RequirePermission("manage-suppliers"))
	{
		suppliers.GET("", h.Supplier.List)
		suppliers.POST("", h.Supplier.Create)
		suppliers.GET("/:id", h.Supplier.Get)
		suppliers.PUT("/:id", h.Supplier.Update)
		suppliers.DELETE("/:id", h.Supplier.Delete)
	}
}

func registerQuotationRoutes(protected *gin.RouterGroup, h *Handlers) {
	quotations := protected.Group("/quotations")
	quotations.Use(middleware.RequirePermission("manage-quotations"))
	{
		quotations.GET("", h.Quotation.List)
		quotations.POST("", h.Quotation.Create)
		quotations.GET("/:id", h.Quotation.Get)
		quotations.PUT("/:id", h.Quotation.Update)
		quotations.DELETE("/:id", h.Quotation.Delete)
	}
}

func registerReportRoutes(protected *gin.RouterGroup) {
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
}

func registerUserRoutes(protected *gin.RouterGroup, h *Handlers) {
	users := protected.Group("/users")
	users.Use(middleware.RequirePermission("manage-users"))
	{
		users.GET("", h.User.List)
		users.GET("/:id", h.User.Get)
		users.PUT("/:id/roles", h.User.UpdateRoles)
		users.DELETE("/:id", h.User.Delete)
	}
}

func registerRoleRoutes(protected *gin.RouterGroup, h *Handlers) {
	roles := protected.Group("/roles")
	roles.Use(middleware.RequirePermission("manage-users"))
	{
		roles.GET("", h.User.ListRoles)
	}
}

func registerPermissionRoutes(protected *gin.RouterGroup, h *Handlers) {
	permissions := protected.Group("/permissions")
	permissions.Use(middleware.RequirePermission("manage-users"))
	{
		permissions.GET("", h.User.ListPermissions)
	}
}

func registerAdminRoutes(protected *gin.RouterGroup, h *Handlers) {
	admin := protected.Group("/admin")
	admin.Use(middleware.RequireRole("super-admin"))
	{
		admin.POST("/tenants/assign-user", h.Tenant.AssignUserToTenant)
	}
}

func registerPrinterRoutes(protected *gin.RouterGroup, h *Handlers) {
	printerGroup := protected.Group("/printer")
	{
		printerGroup.GET("/status", h.Printer.GetStatus)
		printerGroup.POST("/test", h.Printer.TestPrint)
		printerGroup.POST("/receipt", h.Printer.PrintReceipt)
	}
}

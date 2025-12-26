package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/domain/enum"
	"github.com/sangkips/investify-api/internal/domain/repository"
	"github.com/sangkips/investify-api/pkg/pagination"
)

// DashboardService provides dashboard statistics
type DashboardService struct {
	orderRepo     repository.OrderRepository
	purchaseRepo  repository.PurchaseRepository
	productRepo   repository.ProductRepository
	customerRepo  repository.CustomerRepository
	analyticsRepo repository.AnalyticsRepository
}

// NewDashboardService creates a new dashboard service
func NewDashboardService(
	orderRepo repository.OrderRepository,
	purchaseRepo repository.PurchaseRepository,
	productRepo repository.ProductRepository,
	customerRepo repository.CustomerRepository,
	analyticsRepo repository.AnalyticsRepository,
) *DashboardService {
	return &DashboardService{
		orderRepo:     orderRepo,
		purchaseRepo:  purchaseRepo,
		productRepo:   productRepo,
		customerRepo:  customerRepo,
		analyticsRepo: analyticsRepo,
	}
}

// DashboardStats represents dashboard statistics
type DashboardStats struct {
	TotalCustomers    int64                `json:"total_customers"`
	TotalProducts     int64                `json:"total_products"`
	TotalOrders       int64                `json:"total_orders"`
	TotalPurchases    int64                `json:"total_purchases"`
	TotalRevenue      float64              `json:"total_revenue"`
	MonthlyRevenue    float64              `json:"monthly_revenue"`
	LowStockCount     int64                `json:"low_stock_count"`
	PendingOrders     int64                `json:"pending_orders"`
	PendingPurchases  int64                `json:"pending_purchases"`
	RevenueGrowth     float64              `json:"revenue_growth"`
	OrdersGrowth      float64              `json:"orders_growth"`
	CustomersGrowth   float64              `json:"customers_growth"`
	DailySalesData    []DailySalesPoint    `json:"daily_sales_data"`
	CategorySalesData []CategorySalesPoint `json:"category_sales_data"`
	TopProducts       []SalesByProduct     `json:"top_products"`
	SalesByCategory   []SalesByCategory    `json:"sales_by_category"`
	TopCustomers      []SalesByCustomer    `json:"top_customers"`
}

// DailySalesPoint represents a daily sales data point
type DailySalesPoint struct {
	Date    string  `json:"date"`
	Revenue float64 `json:"revenue"`
	Profit  float64 `json:"profit"`
}

// CategorySalesPoint represents sales by category
type CategorySalesPoint struct {
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
}

// SalesByProduct represents top selling products
type SalesByProduct struct {
	ProductID    uuid.UUID `json:"product_id"`
	ProductName  string    `json:"product_name"`
	ProductCode  string    `json:"product_code"`
	QuantitySold int       `json:"quantity_sold"`
	Revenue      float64   `json:"revenue"`
}

// SalesByCategory represents category performance
type SalesByCategory struct {
	CategoryID   uuid.UUID `json:"category_id"`
	CategoryName string    `json:"category_name"`
	TotalSales   float64   `json:"total_sales"`
	OrderCount   int       `json:"order_count"`
	Percentage   float64   `json:"percentage"`
}

// SalesByCustomer represents top customers
type SalesByCustomer struct {
	CustomerID   uuid.UUID `json:"customer_id"`
	CustomerName string    `json:"customer_name"`
	TotalSpent   float64   `json:"total_spent"`
	OrderCount   int       `json:"order_count"`
}

// GetDashboardStats returns dashboard statistics using optimized SQL queries
func (s *DashboardService) GetDashboardStats(ctx context.Context, userID uuid.UUID) (*DashboardStats, error) {
	stats := &DashboardStats{}

	// Get counts
	paginationParams := pagination.DefaultPagination()
	paginationParams.PerPage = 1 // We only need the count

	// Customers - show all customers for admin dashboard (skipUserFilter = true)
	_, customerCount, err := s.customerRepo.List(ctx, userID, paginationParams, "", true)
	if err != nil {
		return nil, err
	}
	stats.TotalCustomers = customerCount

	// Products - show all products in dashboard (skip user filter for overview)
	productParams := &repository.ProductFilterParams{
		Pagination:     paginationParams,
		SkipUserFilter: true,
	}
	_, productCount, err := s.productRepo.List(ctx, userID, productParams)
	if err != nil {
		return nil, err
	}
	stats.TotalProducts = productCount

	// Low stock products - show all low stock items
	lowStockParams := &repository.ProductFilterParams{
		Pagination:     &pagination.PaginationParams{Page: 1, PerPage: 1000},
		LowStock:       true,
		SkipUserFilter: true,
	}
	lowStockProducts, _, err := s.productRepo.List(ctx, userID, lowStockParams)
	if err != nil {
		return nil, err
	}
	stats.LowStockCount = int64(len(lowStockProducts))

	// Orders - show all orders for admin dashboard
	orderParams := &repository.OrderFilterParams{
		Pagination:     paginationParams,
		SkipUserFilter: true,
	}
	_, orderCount, err := s.orderRepo.List(ctx, userID, orderParams)
	if err != nil {
		return nil, err
	}
	stats.TotalOrders = orderCount

	// Pending orders - show all pending orders
	pendingStatus := enum.OrderStatusPending
	pendingOrderParams := &repository.OrderFilterParams{
		Pagination:     paginationParams,
		Status:         &pendingStatus,
		SkipUserFilter: true,
	}
	_, pendingOrderCount, err := s.orderRepo.List(ctx, userID, pendingOrderParams)
	if err != nil {
		return nil, err
	}
	stats.PendingOrders = pendingOrderCount

	// Get total revenue using optimized analytics query
	totalRevenue, err := s.analyticsRepo.GetTotalRevenue(ctx)
	if err != nil {
		return nil, err
	}
	stats.TotalRevenue = totalRevenue

	// Get monthly revenue using optimized analytics query
	monthlyRevenue, err := s.analyticsRepo.GetMonthlyRevenue(ctx)
	if err != nil {
		return nil, err
	}
	stats.MonthlyRevenue = monthlyRevenue

	// Purchases - show all purchases for admin dashboard
	purchaseParams := &repository.PurchaseFilterParams{
		Pagination:     paginationParams,
		SkipUserFilter: true,
	}
	_, purchaseCount, err := s.purchaseRepo.List(ctx, userID, purchaseParams)
	if err != nil {
		return nil, err
	}
	stats.TotalPurchases = purchaseCount

	// Pending purchases - show all pending purchases
	pendingPurchaseStatus := enum.PurchaseStatusPending
	pendingPurchaseParams := &repository.PurchaseFilterParams{
		Pagination:     paginationParams,
		Status:         &pendingPurchaseStatus,
		SkipUserFilter: true,
	}
	_, pendingPurchaseCount, err := s.purchaseRepo.List(ctx, userID, pendingPurchaseParams)
	if err != nil {
		return nil, err
	}
	stats.PendingPurchases = pendingPurchaseCount

	// Get daily sales data using optimized analytics query
	dailySales, err := s.analyticsRepo.GetDailySales(ctx, 7)
	if err != nil {
		return nil, err
	}
	stats.DailySalesData = make([]DailySalesPoint, len(dailySales))
	for i, ds := range dailySales {
		stats.DailySalesData[i] = DailySalesPoint{
			Date:    ds.Date.Format("Jan 02"),
			Revenue: ds.Revenue,
			Profit:  ds.Profit,
		}
	}

	// Get top products using optimized analytics query
	topProducts, err := s.analyticsRepo.GetTopProducts(ctx, 10)
	if err != nil {
		return nil, err
	}
	stats.TopProducts = make([]SalesByProduct, len(topProducts))
	for i, p := range topProducts {
		stats.TopProducts[i] = SalesByProduct{
			ProductID:    p.ProductID,
			ProductName:  p.ProductName,
			ProductCode:  p.ProductCode,
			QuantitySold: p.QuantitySold,
			Revenue:      p.Revenue,
		}
	}

	// Get sales by category using optimized analytics query
	salesByCategory, err := s.analyticsRepo.GetSalesByCategory(ctx)
	if err != nil {
		return nil, err
	}
	stats.SalesByCategory = make([]SalesByCategory, len(salesByCategory))
	for i, c := range salesByCategory {
		stats.SalesByCategory[i] = SalesByCategory{
			CategoryID:   c.CategoryID,
			CategoryName: c.CategoryName,
			TotalSales:   c.TotalSales,
			OrderCount:   c.OrderCount,
			Percentage:   c.Percentage,
		}
	}

	// Get top customers using optimized analytics query
	topCustomers, err := s.analyticsRepo.GetTopCustomers(ctx, 10)
	if err != nil {
		return nil, err
	}
	stats.TopCustomers = make([]SalesByCustomer, len(topCustomers))
	for i, c := range topCustomers {
		stats.TopCustomers[i] = SalesByCustomer{
			CustomerID:   c.CustomerID,
			CustomerName: c.CustomerName,
			TotalSpent:   c.TotalSpent,
			OrderCount:   c.OrderCount,
		}
	}

	return stats, nil
}

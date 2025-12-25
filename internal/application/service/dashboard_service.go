package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/domain/enum"
	"github.com/sangkips/investify-api/internal/domain/repository"
	"github.com/sangkips/investify-api/pkg/pagination"
)

// DashboardService provides dashboard statistics
type DashboardService struct {
	orderRepo    repository.OrderRepository
	purchaseRepo repository.PurchaseRepository
	productRepo  repository.ProductRepository
	customerRepo repository.CustomerRepository
}

// NewDashboardService creates a new dashboard service
func NewDashboardService(
	orderRepo repository.OrderRepository,
	purchaseRepo repository.PurchaseRepository,
	productRepo repository.ProductRepository,
	customerRepo repository.CustomerRepository,
) *DashboardService {
	return &DashboardService{
		orderRepo:    orderRepo,
		purchaseRepo: purchaseRepo,
		productRepo:  productRepo,
		customerRepo: customerRepo,
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

// GetDashboardStats returns dashboard statistics
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
	orders, orderCount, err := s.orderRepo.List(ctx, userID, orderParams)
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

	// Calculate total revenue from all completed orders
	completeStatus := enum.OrderStatusComplete
	completeOrderParams := &repository.OrderFilterParams{
		Pagination:     &pagination.PaginationParams{Page: 1, PerPage: 1000},
		Status:         &completeStatus,
		SkipUserFilter: true,
	}
	completeOrders, _, err := s.orderRepo.List(ctx, userID, completeOrderParams)
	if err != nil {
		return nil, err
	}

	var totalRevenue int64
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	var monthlyRevenue int64

	for _, order := range completeOrders {
		totalRevenue += order.Total
		if order.OrderDate.After(startOfMonth) {
			monthlyRevenue += order.Total
		}
	}
	stats.TotalRevenue = float64(totalRevenue) / 100
	stats.MonthlyRevenue = float64(monthlyRevenue) / 100

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

	// Pending purchases - show all pending purchases using List with status filter
	pendingPurchaseStatus := enum.PurchaseStatusPending
	pendingPurchaseParams := &repository.PurchaseFilterParams{
		Pagination:     paginationParams,
		Status:         &pendingPurchaseStatus,
		SkipUserFilter: true,
	}
	pendingPurchases, pendingPurchaseCount, err := s.purchaseRepo.List(ctx, userID, pendingPurchaseParams)
	if err != nil {
		return nil, err
	}
	stats.PendingPurchases = pendingPurchaseCount
	_ = pendingPurchases // Use the variable to avoid compiler warning

	// Calculate daily sales for the last 7 days
	stats.DailySalesData = make([]DailySalesPoint, 0, 7)
	for i := 6; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")

		dayRevenue := int64(0)
		for _, order := range orders {
			if order.OrderDate.Format("2006-01-02") == dateStr {
				dayRevenue += order.Total
			}
		}

		stats.DailySalesData = append(stats.DailySalesData, DailySalesPoint{
			Date:    date.Format("Jan 02"),
			Revenue: float64(dayRevenue) / 100,
			Profit:  float64(dayRevenue) / 100 * 0.2, // Assume 20% profit margin
		})
	}

	return stats, nil
}

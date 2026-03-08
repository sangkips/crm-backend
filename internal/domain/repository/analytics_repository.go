package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// TopProductResult represents a product's sales performance
type TopProductResult struct {
	ProductID    uuid.UUID
	ProductName  string
	ProductCode  string
	QuantitySold int
	Revenue      float64
}

// CategorySalesResult represents sales aggregated by category
type CategorySalesResult struct {
	CategoryID   uuid.UUID
	CategoryName string
	TotalSales   float64
	OrderCount   int
	Percentage   float64
}

// TopCustomerResult represents a customer's spending data
type TopCustomerResult struct {
	CustomerID   uuid.UUID
	CustomerName string
	TotalSpent   float64
	OrderCount   int
}

// DailySalesResult represents sales data for a single day
type DailySalesResult struct {
	Date    time.Time
	Revenue float64
	Profit  float64
}

// DateRange represents an optional date filter applied to analytics queries.
// Both fields must be non-nil for the filter to be applied.
type DateRange struct {
	Start time.Time
	End   time.Time
}

// AnalyticsRepository defines interface for analytics/aggregation queries
type AnalyticsRepository interface {
	// GetTopProducts returns top selling products by revenue
	GetTopProducts(ctx context.Context, limit int, dr *DateRange) ([]TopProductResult, error)

	// GetSalesByCategory returns sales aggregated by category with percentages
	GetSalesByCategory(ctx context.Context, dr *DateRange) ([]CategorySalesResult, error)

	// GetTopCustomers returns top customers by total spending
	GetTopCustomers(ctx context.Context, limit int, dr *DateRange) ([]TopCustomerResult, error)

	// GetDailySales returns daily sales data for the last N days
	GetDailySales(ctx context.Context, days int, dr *DateRange) ([]DailySalesResult, error)

	// GetTotalRevenue returns total revenue from completed orders
	GetTotalRevenue(ctx context.Context, dr *DateRange) (float64, error)

	// GetMonthlyRevenue returns revenue for the current month
	GetMonthlyRevenue(ctx context.Context) (float64, error)
}

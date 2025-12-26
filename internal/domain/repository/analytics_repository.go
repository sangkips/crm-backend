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

// AnalyticsRepository defines interface for analytics/aggregation queries
type AnalyticsRepository interface {
	// GetTopProducts returns top selling products by revenue
	GetTopProducts(ctx context.Context, limit int) ([]TopProductResult, error)

	// GetSalesByCategory returns sales aggregated by category with percentages
	GetSalesByCategory(ctx context.Context) ([]CategorySalesResult, error)

	// GetTopCustomers returns top customers by total spending
	GetTopCustomers(ctx context.Context, limit int) ([]TopCustomerResult, error)

	// GetDailySales returns daily sales data for the last N days
	GetDailySales(ctx context.Context, days int) ([]DailySalesResult, error)

	// GetTotalRevenue returns total revenue from completed orders
	GetTotalRevenue(ctx context.Context) (float64, error)

	// GetMonthlyRevenue returns revenue for the current month
	GetMonthlyRevenue(ctx context.Context) (float64, error)
}

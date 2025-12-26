package repository

import (
	"context"
	"database/sql"
	"time"

	domainRepo "github.com/sangkips/investify-api/internal/domain/repository"
	"gorm.io/gorm"
)

type analyticsRepository struct {
	db *gorm.DB
}

// NewAnalyticsRepository creates a new analytics repository
func NewAnalyticsRepository(db *gorm.DB) domainRepo.AnalyticsRepository {
	return &analyticsRepository{db: db}
}

func (r *analyticsRepository) GetTopProducts(ctx context.Context, limit int) ([]domainRepo.TopProductResult, error) {
	var results []domainRepo.TopProductResult

	err := r.db.WithContext(ctx).Raw(`
		SELECT 
			p.id as product_id,
			p.name as product_name,
			p.code as product_code,
			COALESCE(SUM(od.quantity), 0) as quantity_sold,
			COALESCE(SUM(od.total), 0) / 100.0 as revenue
		FROM order_details od
		JOIN products p ON p.id = od.product_id
		JOIN orders o ON o.id = od.order_id
		WHERE o.order_status = 1
		GROUP BY p.id, p.name, p.code
		ORDER BY revenue DESC
		LIMIT ?
	`, limit).Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (r *analyticsRepository) GetSalesByCategory(ctx context.Context) ([]domainRepo.CategorySalesResult, error) {
	var results []domainRepo.CategorySalesResult

	// First get total sales for percentage calculation
	var totalSales float64
	err := r.db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(od.total), 0) / 100.0
		FROM order_details od
		JOIN orders o ON o.id = od.order_id
		WHERE o.order_status = 1
	`).Scan(&totalSales).Error
	if err != nil {
		return nil, err
	}

	// Get sales by category
	err = r.db.WithContext(ctx).Raw(`
		SELECT 
			COALESCE(c.id, '00000000-0000-0000-0000-000000000000') as category_id,
			COALESCE(c.name, 'Uncategorized') as category_name,
			COALESCE(SUM(od.total), 0) / 100.0 as total_sales,
			COUNT(DISTINCT o.id) as order_count
		FROM order_details od
		JOIN products p ON p.id = od.product_id
		LEFT JOIN categories c ON c.id = p.category_id
		JOIN orders o ON o.id = od.order_id
		WHERE o.order_status = 1
		GROUP BY c.id, c.name
		ORDER BY total_sales DESC
	`).Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// Calculate percentages
	for i := range results {
		if totalSales > 0 {
			results[i].Percentage = (results[i].TotalSales / totalSales) * 100
		}
	}

	return results, nil
}

func (r *analyticsRepository) GetTopCustomers(ctx context.Context, limit int) ([]domainRepo.TopCustomerResult, error) {
	var results []domainRepo.TopCustomerResult

	err := r.db.WithContext(ctx).Raw(`
		SELECT 
			c.id as customer_id,
			c.name as customer_name,
			COALESCE(SUM(o.total), 0) / 100.0 as total_spent,
			COUNT(o.id) as order_count
		FROM orders o
		JOIN customers c ON c.id = o.customer_id
		WHERE o.order_status = 1 AND o.customer_id IS NOT NULL
		GROUP BY c.id, c.name
		ORDER BY total_spent DESC
		LIMIT ?
	`, limit).Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (r *analyticsRepository) GetDailySales(ctx context.Context, days int) ([]domainRepo.DailySalesResult, error) {
	results := make([]domainRepo.DailySalesResult, 0, days)
	now := time.Now()

	// Generate dates for the last N days and get sales for each
	for i := days - 1; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		endOfDay := startOfDay.Add(24 * time.Hour)

		var revenue sql.NullFloat64
		err := r.db.WithContext(ctx).Raw(`
			SELECT COALESCE(SUM(total), 0) / 100.0
			FROM orders
			WHERE order_status = 1 
			AND order_date >= ? AND order_date < ?
		`, startOfDay, endOfDay).Scan(&revenue).Error

		if err != nil {
			return nil, err
		}

		rev := 0.0
		if revenue.Valid {
			rev = revenue.Float64
		}

		results = append(results, domainRepo.DailySalesResult{
			Date:    startOfDay,
			Revenue: rev,
			Profit:  rev * 0.2, // Assume 20% profit margin
		})
	}

	return results, nil
}

func (r *analyticsRepository) GetTotalRevenue(ctx context.Context) (float64, error) {
	var revenue float64
	err := r.db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(total), 0) / 100.0
		FROM orders
		WHERE order_status = 1
	`).Scan(&revenue).Error

	return revenue, err
}

func (r *analyticsRepository) GetMonthlyRevenue(ctx context.Context) (float64, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	var revenue float64
	err := r.db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(total), 0) / 100.0
		FROM orders
		WHERE order_status = 1 AND order_date >= ?
	`, startOfMonth).Scan(&revenue).Error

	return revenue, err
}

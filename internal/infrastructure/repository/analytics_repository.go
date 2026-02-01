package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
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

// getTenantFilter returns the tenant filter clause and args for raw SQL queries
// Returns empty string and nil if tenant scope should be skipped (super admin)
// Returns "1 = 0" if tenant context is missing (fail-safe)
func (r *analyticsRepository) getTenantFilter(ctx context.Context, tableName string) (string, []interface{}) {
	// Check if tenant scope should be skipped (super admin)
	if skipScope, ok := ctx.Value(SkipTenantScopeKey).(bool); ok && skipScope {
		return "", nil
	}

	tenantID, ok := ctx.Value(TenantIDKey).(uuid.UUID)
	if !ok {
		// Fail-safe: return no results if tenant context missing
		return "1 = 0", nil
	}

	if tableName != "" {
		return tableName + ".tenant_id = ?", []interface{}{tenantID}
	}
	return "tenant_id = ?", []interface{}{tenantID}
}

func (r *analyticsRepository) GetTopProducts(ctx context.Context, limit int) ([]domainRepo.TopProductResult, error) {
	var results []domainRepo.TopProductResult

	tenantFilter, tenantArgs := r.getTenantFilter(ctx, "o")
	whereClause := "o.order_status = 1"
	args := []interface{}{}

	if tenantFilter != "" {
		whereClause += " AND " + tenantFilter
		args = append(args, tenantArgs...)
	}
	args = append(args, limit)

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
		WHERE `+whereClause+`
		GROUP BY p.id, p.name, p.code
		ORDER BY revenue DESC
		LIMIT ?
	`, args...).Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (r *analyticsRepository) GetSalesByCategory(ctx context.Context) ([]domainRepo.CategorySalesResult, error) {
	var results []domainRepo.CategorySalesResult

	tenantFilter, tenantArgs := r.getTenantFilter(ctx, "o")
	whereClause := "o.order_status = 1"
	args := []interface{}{}

	if tenantFilter != "" {
		whereClause += " AND " + tenantFilter
		args = append(args, tenantArgs...)
	}

	// First get total sales for percentage calculation
	var totalSales float64
	err := r.db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(od.total), 0) / 100.0
		FROM order_details od
		JOIN orders o ON o.id = od.order_id
		WHERE `+whereClause, args...).Scan(&totalSales).Error
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
		WHERE `+whereClause+`
		GROUP BY c.id, c.name
		ORDER BY total_sales DESC
	`, args...).Scan(&results).Error

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

	tenantFilter, tenantArgs := r.getTenantFilter(ctx, "o")
	whereClause := "o.order_status = 1 AND o.customer_id IS NOT NULL"
	args := []interface{}{}

	if tenantFilter != "" {
		whereClause += " AND " + tenantFilter
		args = append(args, tenantArgs...)
	}
	args = append(args, limit)

	err := r.db.WithContext(ctx).Raw(`
		SELECT 
			c.id as customer_id,
			c.name as customer_name,
			COALESCE(SUM(o.total), 0) / 100.0 as total_spent,
			COUNT(o.id) as order_count
		FROM orders o
		JOIN customers c ON c.id = o.customer_id
		WHERE `+whereClause+`
		GROUP BY c.id, c.name
		ORDER BY total_spent DESC
		LIMIT ?
	`, args...).Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (r *analyticsRepository) GetDailySales(ctx context.Context, days int) ([]domainRepo.DailySalesResult, error) {
	results := make([]domainRepo.DailySalesResult, 0, days)
	now := time.Now()

	tenantFilter, tenantArgs := r.getTenantFilter(ctx, "")
	baseWhereClause := "order_status = 1"
	if tenantFilter != "" {
		baseWhereClause += " AND " + tenantFilter
	}

	// Generate dates for the last N days and get sales for each
	for i := days - 1; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		endOfDay := startOfDay.Add(24 * time.Hour)

		args := []interface{}{}
		args = append(args, tenantArgs...)
		args = append(args, startOfDay, endOfDay)

		var revenue sql.NullFloat64
		err := r.db.WithContext(ctx).Raw(`
			SELECT COALESCE(SUM(total), 0) / 100.0
			FROM orders
			WHERE `+baseWhereClause+`
			AND order_date >= ? AND order_date < ?
		`, args...).Scan(&revenue).Error

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
	tenantFilter, tenantArgs := r.getTenantFilter(ctx, "")
	whereClause := "order_status = 1"
	args := []interface{}{}

	if tenantFilter != "" {
		whereClause += " AND " + tenantFilter
		args = append(args, tenantArgs...)
	}

	var revenue float64
	err := r.db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(total), 0) / 100.0
		FROM orders
		WHERE `+whereClause,
		args...).Scan(&revenue).Error

	return revenue, err
}

func (r *analyticsRepository) GetMonthlyRevenue(ctx context.Context) (float64, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	tenantFilter, tenantArgs := r.getTenantFilter(ctx, "")
	whereClause := "order_status = 1 AND order_date >= ?"
	args := []interface{}{}

	if tenantFilter != "" {
		whereClause += " AND " + tenantFilter
	}
	args = append(args, startOfMonth)
	args = append(args, tenantArgs...)

	var revenue float64
	err := r.db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(total), 0) / 100.0
		FROM orders
		WHERE `+whereClause,
		args...).Scan(&revenue).Error

	return revenue, err
}

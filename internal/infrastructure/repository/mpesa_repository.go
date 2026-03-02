package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/domain/entity"
	"gorm.io/gorm"
)

// MpesaTransactionRepository is a GORM implementation of the MpesaTransactionRepository interface
type MpesaTransactionRepository struct {
	db *gorm.DB
}

// NewMpesaTransactionRepository creates a new MpesaTransactionRepository
func NewMpesaTransactionRepository(db *gorm.DB) *MpesaTransactionRepository {
	return &MpesaTransactionRepository{db: db}
}

// Create creates a new M-Pesa transaction record
func (r *MpesaTransactionRepository) Create(ctx context.Context, tx *entity.MpesaTransaction) error {
	return r.db.WithContext(ctx).Scopes(TenantScope(ctx)).Create(tx).Error
}

// GetByID retrieves a transaction by ID (tenant-scoped)
func (r *MpesaTransactionRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.MpesaTransaction, error) {
	var tx entity.MpesaTransaction
	result := r.db.WithContext(ctx).Scopes(TenantScope(ctx)).Where("id = ?", id).First(&tx)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &tx, nil
}

// GetByCheckoutRequestID retrieves a transaction by Daraja CheckoutRequestID.
// This is intentionally NOT tenant-scoped because Safaricom callbacks don't carry tenant context.
func (r *MpesaTransactionRepository) GetByCheckoutRequestID(ctx context.Context, checkoutRequestID string) (*entity.MpesaTransaction, error) {
	var tx entity.MpesaTransaction
	result := r.db.WithContext(ctx).Where("checkout_request_id = ?", checkoutRequestID).First(&tx)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &tx, nil
}

// GetByOrderID retrieves all transactions for an order (tenant-scoped)
func (r *MpesaTransactionRepository) GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]entity.MpesaTransaction, error) {
	var transactions []entity.MpesaTransaction
	result := r.db.WithContext(ctx).Scopes(TenantScope(ctx)).
		Where("order_id = ?", orderID).
		Order("created_at DESC").
		Find(&transactions)
	if result.Error != nil {
		return nil, result.Error
	}
	return transactions, nil
}

// Update updates an existing transaction
func (r *MpesaTransactionRepository) Update(ctx context.Context, tx *entity.MpesaTransaction) error {
	return r.db.WithContext(ctx).Save(tx).Error
}

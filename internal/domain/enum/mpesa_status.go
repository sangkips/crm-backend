package enum

// MpesaTransactionStatus represents the status of an M-Pesa transaction
type MpesaTransactionStatus int

const (
	// MpesaStatusPending indicates the STK push has been sent and is awaiting user action
	MpesaStatusPending MpesaTransactionStatus = iota
	// MpesaStatusSuccess indicates the payment was completed successfully
	MpesaStatusSuccess
	// MpesaStatusFailed indicates the payment failed
	MpesaStatusFailed
	// MpesaStatusCancelled indicates the user cancelled the payment
	MpesaStatusCancelled
)

// String returns the string representation of the status
func (s MpesaTransactionStatus) String() string {
	switch s {
	case MpesaStatusPending:
		return "pending"
	case MpesaStatusSuccess:
		return "success"
	case MpesaStatusFailed:
		return "failed"
	case MpesaStatusCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

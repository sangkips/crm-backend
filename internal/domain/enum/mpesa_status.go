package enum

import (
	"database/sql/driver"
	"encoding/json"
)

// MpesaTransactionStatus represents the status of an M-Pesa transaction
type MpesaTransactionStatus int

const (
	// MpesaStatusPending indicates the STK push has been sent and is awaiting user action
	MpesaStatusPending MpesaTransactionStatus = 0
	// MpesaStatusSuccess indicates the payment was completed successfully
	MpesaStatusSuccess MpesaTransactionStatus = 1
	// MpesaStatusFailed indicates the payment failed
	MpesaStatusFailed MpesaTransactionStatus = 2
	// MpesaStatusCancelled indicates the user cancelled the payment
	MpesaStatusCancelled MpesaTransactionStatus = 3
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

// MarshalJSON serializes the status as a string for API responses
func (s MpesaTransactionStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// UnmarshalJSON deserializes the status from a string or int
func (s *MpesaTransactionStatus) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		// Try unmarshaling as int
		var i int
		if err := json.Unmarshal(data, &i); err != nil {
			return err
		}
		*s = MpesaTransactionStatus(i)
		return nil
	}
	switch str {
	case "pending":
		*s = MpesaStatusPending
	case "success":
		*s = MpesaStatusSuccess
	case "failed":
		*s = MpesaStatusFailed
	case "cancelled":
		*s = MpesaStatusCancelled
	}
	return nil
}

// Value implements driver.Valuer — stores as integer in the database
func (s MpesaTransactionStatus) Value() (driver.Value, error) {
	return int64(s), nil
}

// Scan implements sql.Scanner — reads integer from database
func (s *MpesaTransactionStatus) Scan(value interface{}) error {
	if value == nil {
		*s = MpesaStatusPending
		return nil
	}
	switch v := value.(type) {
	case int64:
		*s = MpesaTransactionStatus(v)
	case int:
		*s = MpesaTransactionStatus(v)
	}
	return nil
}

package enum

import (
	"database/sql/driver"
	"encoding/json"
)

// PurchaseStatus represents the status of a purchase
type PurchaseStatus int

const (
	PurchaseStatusPending  PurchaseStatus = 0
	PurchaseStatusApproved PurchaseStatus = 1
)

func (s PurchaseStatus) String() string {
	return [...]string{"Pending", "Approved"}[s]
}

func (s PurchaseStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *PurchaseStatus) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		var i int
		if err := json.Unmarshal(data, &i); err != nil {
			return err
		}
		*s = PurchaseStatus(i)
		return nil
	}
	switch str {
	case "Pending":
		*s = PurchaseStatusPending
	case "Approved":
		*s = PurchaseStatusApproved
	}
	return nil
}

func (s PurchaseStatus) Value() (driver.Value, error) {
	return int64(s), nil
}

func (s *PurchaseStatus) Scan(value interface{}) error {
	if value == nil {
		*s = PurchaseStatusPending
		return nil
	}
	switch v := value.(type) {
	case int64:
		*s = PurchaseStatus(v)
	case int:
		*s = PurchaseStatus(v)
	}
	return nil
}

package enum

import (
	"database/sql/driver"
	"encoding/json"
)

// QuotationStatus represents the status of a quotation
type QuotationStatus int

const (
	QuotationStatusPending  QuotationStatus = 0
	QuotationStatusSent     QuotationStatus = 1
	QuotationStatusCanceled QuotationStatus = 2
)

func (s QuotationStatus) String() string {
	return [...]string{"Pending", "Sent", "Canceled"}[s]
}

func (s QuotationStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *QuotationStatus) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		var i int
		if err := json.Unmarshal(data, &i); err != nil {
			return err
		}
		*s = QuotationStatus(i)
		return nil
	}
	switch str {
	case "Pending":
		*s = QuotationStatusPending
	case "Sent":
		*s = QuotationStatusSent
	case "Canceled":
		*s = QuotationStatusCanceled
	}
	return nil
}

func (s QuotationStatus) Value() (driver.Value, error) {
	return int64(s), nil
}

func (s *QuotationStatus) Scan(value interface{}) error {
	if value == nil {
		*s = QuotationStatusPending
		return nil
	}
	switch v := value.(type) {
	case int64:
		*s = QuotationStatus(v)
	case int:
		*s = QuotationStatus(v)
	}
	return nil
}

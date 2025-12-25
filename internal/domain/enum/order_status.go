package enum

import (
	"database/sql/driver"
	"encoding/json"
)

// OrderStatus represents the status of an order
type OrderStatus int

const (
	OrderStatusPending  OrderStatus = 0
	OrderStatusComplete OrderStatus = 1
	OrderStatusCancel   OrderStatus = 2
)

func (s OrderStatus) String() string {
	return [...]string{"Pending", "Complete", "Cancel"}[s]
}

func (s OrderStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *OrderStatus) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		// Try unmarshaling as int
		var i int
		if err := json.Unmarshal(data, &i); err != nil {
			return err
		}
		*s = OrderStatus(i)
		return nil
	}
	switch str {
	case "Pending":
		*s = OrderStatusPending
	case "Complete":
		*s = OrderStatusComplete
	case "Cancel":
		*s = OrderStatusCancel
	}
	return nil
}

func (s OrderStatus) Value() (driver.Value, error) {
	return int64(s), nil
}

func (s *OrderStatus) Scan(value interface{}) error {
	if value == nil {
		*s = OrderStatusPending
		return nil
	}
	switch v := value.(type) {
	case int64:
		*s = OrderStatus(v)
	case int:
		*s = OrderStatus(v)
	}
	return nil
}

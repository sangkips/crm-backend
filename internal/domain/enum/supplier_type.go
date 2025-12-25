package enum

import (
	"database/sql/driver"
	"encoding/json"
)

// SupplierType represents the type of supplier
type SupplierType string

const (
	SupplierTypeDistributor SupplierType = "distributor"
	SupplierTypeWholesaler  SupplierType = "wholesaler"
	SupplierTypeProducer    SupplierType = "producer"
)

func (t SupplierType) String() string {
	return string(t)
}

func (t SupplierType) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(t))
}

func (t *SupplierType) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	*t = SupplierType(str)
	return nil
}

func (t SupplierType) Value() (driver.Value, error) {
	return string(t), nil
}

func (t *SupplierType) Scan(value interface{}) error {
	if value == nil {
		*t = SupplierTypeDistributor
		return nil
	}
	switch v := value.(type) {
	case string:
		*t = SupplierType(v)
	case []byte:
		*t = SupplierType(string(v))
	}
	return nil
}

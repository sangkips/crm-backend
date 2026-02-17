package enum

import (
	"database/sql/driver"
	"encoding/json"
)

// TaxType represents how tax is applied
type TaxType int

const (
	TaxTypeExclusive TaxType = 0
	TaxTypeInclusive TaxType = 1
)

func (t TaxType) String() string {
	names := [...]string{"Exclusive", "Inclusive"}
	if int(t) < 0 || int(t) >= len(names) {
		return "Exclusive"
	}
	return names[t]
}

func (t TaxType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

func (t *TaxType) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		var i int
		if err := json.Unmarshal(data, &i); err != nil {
			return err
		}
		*t = TaxType(i)
		return nil
	}
	switch str {
	case "Exclusive":
		*t = TaxTypeExclusive
	case "Inclusive":
		*t = TaxTypeInclusive
	}
	return nil
}

func (t TaxType) Value() (driver.Value, error) {
	return int64(t), nil
}

func (t *TaxType) Scan(value interface{}) error {
	if value == nil {
		*t = TaxTypeExclusive
		return nil
	}
	switch v := value.(type) {
	case int64:
		*t = TaxType(v)
	case int:
		*t = TaxType(v)
	}
	return nil
}

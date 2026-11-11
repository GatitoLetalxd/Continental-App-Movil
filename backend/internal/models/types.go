package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// StringSlice implementa driver.Valuer, sql.Scanner y json.Marshaler para almacenar slices de strings como JSON en SQLite y garantizar arrays no nulos en JSON
type StringSlice []string

func (s StringSlice) Value() (driver.Value, error) {
	if s == nil {
		return "[]", nil
	}
	bytes, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return string(bytes), nil
}

func (s *StringSlice) Scan(value interface{}) error {
	if value == nil {
		*s = []string{}
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case string:
		bytes = []byte(v)
	case []byte:
		bytes = v
	default:
		return errors.New("tipo no soportado para StringSlice")
	}
	if len(bytes) == 0 {
		*s = []string{}
		return nil
	}
	if err := json.Unmarshal(bytes, s); err != nil {
		*s = []string{}
		return nil
	}
	if *s == nil {
		*s = []string{}
	}
	return nil
}

func (s StringSlice) MarshalJSON() ([]byte, error) {
	if s == nil {
		return []byte("[]"), nil
	}
	return json.Marshal([]string(s))
}

package dbtypes

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// JSONMap is a convenience type for persisting arbitrary JSON payloads to Postgres.
// It implements database/sql interfaces for scanning and writing.
type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, j)
}

// NullTime mirrors sql.NullTime with a Scan that accepts time.Time.
type NullTime struct {
	sql.NullTime
}

func (nt *NullTime) Scan(value interface{}) error {
	if value == nil {
		nt.Valid = false
		return nil
	}
	nt.Valid = true
	switch v := value.(type) {
	case time.Time:
		nt.Time = v
	default:
		return errors.New("type assertion to time.Time failed")
	}
	return nil
}

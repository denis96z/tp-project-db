package models

import (
	"database/sql/driver"
	"encoding/json"
	"github.com/go-openapi/strfmt"
	"github.com/mailru/easyjson"
)

type NullTimestamp struct {
	Valid     bool
	Timestamp strfmt.DateTime
}

func (t *NullTimestamp) MarshalJSON() ([]byte, error) {
	if t.Valid {
		return easyjson.Marshal(t.Timestamp)
	}
	return NullSlice, nil
}

func (t *NullTimestamp) UnmarshalJSON(b []byte) error {
	var tStr string
	err := json.Unmarshal(b, &tStr)
	if tStr == "null" {
		t.Valid = false
		return nil
	}

	t.Timestamp, err = strfmt.ParseDateTime(tStr)
	if err != nil {
		return err
	}

	t.Valid = true
	return nil
}

func (t *NullTimestamp) Scan(value interface{}) error {
	t.Valid = t.Timestamp.Scan(value) == nil
	return nil
}

func (t *NullTimestamp) Value() (driver.Value, error) {
	if t.Valid {
		return t.Timestamp, nil
	}
	return nil, nil
}

package models

import (
	"database/sql"
	"encoding/json"
	"github.com/go-openapi/strfmt"
	"github.com/mailru/easyjson"
)

var (
	nullSlice = []byte("null")
)

type NullString sql.NullString

func (ns *NullString) MarshalJSON() ([]byte, error) {
	if ns.Valid {
		return json.Marshal(ns.String)
	}
	return nullSlice, nil
}

func (ns *NullString) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, &ns.String)
	ns.Valid = err == nil
	return err
}

func (ns *NullString) Scan(value interface{}) error {
	return (*sql.NullString)(ns).Scan(value)
}

type NullTimestamp struct {
	Valid     bool
	Timestamp strfmt.DateTime
}

func (t *NullTimestamp) MarshalJSON() ([]byte, error) {
	if t.Valid {
		return easyjson.Marshal(t.Timestamp)
	}
	return nullSlice, nil
}

func (t *NullTimestamp) UnmarshalJSON(b []byte) error {
	var tStr string
	_ = json.Unmarshal(b, &tStr)

	if tStr == "null" {
		t.Valid = false
		return nil
	}

	t.Timestamp, _ = strfmt.ParseDateTime(tStr) //TODO
	t.Valid = true

	return nil
}

func (t *NullTimestamp) Scan(value interface{}) error {
	t.Valid = t.Timestamp.Scan(value) == nil
	return nil
}

package models

import (
	"encoding/json"
	"time"
)

const (
	TimestampFormat = "2017-01-01T00:00:00.000Z"
)

type Timestamp struct {
	Null  bool
	Value time.Time
}

func (t *Timestamp) MarshalJSON() ([]byte, error) {
	if t.Null {
		return NullSlice, nil
	}
	return []byte(t.Value.Format(TimestampFormat)), nil
}

func (t *Timestamp) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, &t.Value)
	t.Null = err == nil
	return nil
}

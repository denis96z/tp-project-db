package models

import (
	"encoding/json"
	"time"
)

const (
	TimestampFormat = "2017-01-01T00:00:00.000Z"
)

type Timestamp time.Time

func (t *Timestamp) MarshalJSON() ([]byte, error) {
	return []byte((*time.Time)(t).Format(TimestampFormat)), nil
}

func (t *Timestamp) UnmarshalJSON(b []byte) error {
	var str string

	err := json.Unmarshal(b, &str)
	if err != nil {
		return err
	}

	tStamp, err := time.Parse(TimestampFormat, str)
	if err != nil {
		return err
	}

	*t = Timestamp(tStamp)
	return nil
}

package morejson

import (
	"encoding/json"
	"errors"
	"time"
)

// PositiveDuration is a wrapper around time.Duration.
// It restricts its values to positive ones only, and it implements the
// json.Marshaler and json.Unmarshaler interfaces.
type PositiveDuration struct {
	time.Duration
}

// MarshalJSON encodes the time.Duration to a JSON string.
func (d PositiveDuration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

// UnmarshalJSON decodes a JSON string parsable by time.ParseDuration and checks
// if it is positive.
func (d *PositiveDuration) UnmarshalJSON(b []byte) (err error) {
	var s string
	if err = json.Unmarshal(b, &s); err != nil {
		return
	}
	d.Duration, err = time.ParseDuration(s)
	if err == nil && d.Duration.Nanoseconds() < 0 {
		err = errors.New("expected positive duration")
	}
	return
}

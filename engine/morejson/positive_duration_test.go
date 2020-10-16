package morejson

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	Duration PositiveDuration `json:"duration"`
}

func TestPositiveDurationValid(t *testing.T) {
	expected := time.Duration(20) * time.Microsecond
	var s testStruct
	bytes := []byte(`{"duration": "20µs"}`)
	err := json.Unmarshal(bytes, &s)

	assert.Nil(t, err)
	assert.Equal(t, expected, s.Duration.Duration)
}

func TestPositiveDurationNegative(t *testing.T) {
	var s testStruct
	bytes := []byte(`{"duration": "-20h"}`)
	err := json.Unmarshal(bytes, &s)

	assert.NotNil(t, err)
}

func TestPositiveDurationRoundtrip(t *testing.T) {
	base := testStruct{
		Duration: PositiveDuration{time.Duration(15426185700)},
	}

	j, err := json.Marshal(&base)
	assert.Nil(t, err)

	var got testStruct
	err = json.Unmarshal(j, &got)
	assert.Nil(t, err)

	assert.Equal(t, base.Duration.Duration, got.Duration.Duration)
}

func TestPositiveDurationNotString(t *testing.T) {
	var s testStruct
	bytes := []byte(`{"duration": 42}`)
	err := json.Unmarshal(bytes, &s)

	assert.NotNil(t, err)
}

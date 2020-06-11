package executor

import (
	"encoding/json"
)

// Executor matches an executor type with its required configuration
type Executor struct {
	Type              string          `json:"type"`
	BaseConfiguration string          `json:"base_configuration,omitempty"`
	Configuration     json.RawMessage `json:"configuration"`
	BaseOutput        json.RawMessage `json:"base_output"`
	TemplatedOutput   interface{}     `json:"templated_output"`
}

package executor

import (
	"encoding/json"

	"sigs.k8s.io/yaml"
)

// Executor matches an executor type with its required configuration
type Executor struct {
	Type              string          `json:"type"`
	BaseConfiguration string          `json:"base_configuration,omitempty"`
	Configuration     json.RawMessage `json:"configuration"`
	Output            *Output         `json:"output"`
	// XXX: BaseOutput is deprecated to benefits to output with merge strategy
	BaseOutput json.RawMessage `json:"base_output,omitempty"`
}

type InnerExecutor Executor

func (e *Executor) UnmarshalJSON(b []byte) error {
	var inner InnerExecutor

	err := yaml.Unmarshal(b, &inner)
	if err != nil {
		return err
	}

	*e = Executor(inner)

	if len(e.BaseOutput) > 0 && string(e.BaseOutput) != "null" {
		e.Output = &Output{
			Strategy: OutputStrategymerge,
			Format:   e.BaseOutput,
		}
	}

	return nil
}

//go:generate stringer -type=OutputStrategy --trimprefix=OutputStrategy
//go:generate jsonenums -type=OutputStrategy
type OutputStrategy int

const (
	OutputStrategynone OutputStrategy = iota
	OutputStrategymerge
	OutputStrategytemplate
)

type Output struct {
	Format   interface{}    `json:"format"`
	Strategy OutputStrategy `json:"strategy"`
}

package condition

import (
	"fmt"

	"github.com/ovh/utask/engine/values"
)

//go:generate jsonenums -type=CondType --lower --no-stringer
//go:generate stringer -type=CondType
const (
	SKIP  = "skip"
	CHECK = "check"
)

// Condition defines a condition to be evaluated before or after a step's action
type Condition struct {
	Type    string            `json:"type"`
	If      []*Assert         `json:"if"`
	Then    map[string]string `json:"then"`
	Final   bool              `json:"final"`
	Message string            `json:"message"`
}

// Eval runs the condition against a set of values, evaluating the underlying Condition
func (sc *Condition) Eval(v *values.Values, item interface{}, stepName string) error {
	for _, c := range sc.If {
		if err := c.Eval(v, item, stepName); err != nil {
			return err
		}
	}
	msg, err := v.Apply(sc.Message, item, stepName)
	if err != nil {
		sc.Message = fmt.Sprintf("%s (TEMPLATING ERROR: %s)", sc.Message, err.Error())
	} else {
		sc.Message = string(msg)
	}
	return nil
}

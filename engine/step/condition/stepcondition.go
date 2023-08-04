package condition

import (
	"fmt"

	"github.com/juju/errors"
	"github.com/ovh/utask/engine/values"
)

//go:generate jsonenums -type=CondType --lower --no-stringer
//go:generate stringer -type=CondType
const (
	SKIP  = "skip"
	CHECK = "check"
)

const (
	// ForEachChildren executes the condition on the children of a foreach step.
	// The children are created, the condition is copied in them, then run.
	// This is the default value.
	ForEachChildren = "children"
	// ForEachParent executes the condition on the foreach step itself.
	ForEachParent = "parent"
)

// Condition defines a condition to be evaluated before or after a step's action
type Condition struct {
	Type    string            `json:"type"`
	If      []*Assert         `json:"if"`
	Then    map[string]string `json:"then"`
	Final   bool              `json:"final"`
	ForEach string            `json:"foreach"`
	Message string            `json:"message"`
}

// Valid asserts that a condition's definition is valid
// ie. the type and foreach are among the accepted values listed above
func (c *Condition) Valid() error {
	if c == nil {
		return nil
	}

	switch c.Type {
	case SKIP, CHECK:
	default:
		return errors.BadRequestf("Unknown condition type: %s", c.Type)
	}

	switch c.ForEach {
	case "", ForEachChildren, ForEachParent:
	default:
		return errors.BadRequestf("Unknown condition foreach: %s", c.ForEach)
	}

	return nil
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

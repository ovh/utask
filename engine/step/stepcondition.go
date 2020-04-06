package step

import (
	"fmt"

	"github.com/juju/errors"
	"github.com/ovh/utask/engine/values"
	"github.com/ovh/utask/pkg/utils"
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

// Valid asserts that the definition for a StepCondition is valid
func (sc *Condition) Valid(stepName string, steps map[string]*Step) error {
	for _, c := range sc.If {
		if err := c.Valid(); err != nil {
			return err
		}
	}
	for thenStep, thenState := range sc.Then {
		// force the use of "this" for single steps
		// except in the case of "loop" steps: the condition will belong to its children,
		// they should be able to reference their parent (ie. break out and retry loop)
		if thenStep == stepName && steps[stepName].ForEach == "" {
			return errors.BadRequestf("Step condition should not reference itself, use '%s'", stepRefThis)
		}

		if thenStep == stepRefThis {
			thenStep = stepName
		}

		impactedStep, ok := steps[thenStep]
		if !ok {
			return errors.BadRequestf("Step condition points to non-existing step: %s", thenStep)
		}

		// As dependencies are in a known state (parents deps are DONE and childs steps are TODO)
		if !canImpactState(stepName, thenStep, steps) {
			return errors.BadRequestf("Step condition cannot impact the state of step %s, only those who belong to the dependency chain are allowed", thenStep)
		}

		validStates := append(stepConditionValidStates, impactedStep.CustomStates...)
		if !utils.ListContainsString(validStates, thenState) {
			return errors.BadRequestf("Step condition implies invalid state for step %s: %s", thenStep, thenState)
		}
	}
	return nil
}

func canImpactState(sourceStep, destinationStep string, steps map[string]*Step) bool {
	if sourceStep == destinationStep {
		return true
	}
	sourceChain := dependenciesChain(steps, steps[sourceStep].Dependencies)
	destinationChain := dependenciesChain(steps, steps[destinationStep].Dependencies)

	return utils.ListContainsString(sourceChain, destinationStep) || utils.ListContainsString(destinationChain, sourceStep)
}

// dependenciesChain build a chain of dependencies given a list of dependencies, usually
// the step dependencies we want to check.
// This assume that every dependencies does exist.
func dependenciesChain(steps map[string]*Step, dependencies []string) []string {
	// Discard dependencies state
	chain := []string{}
	for _, dep := range dependencies {
		stepName, _ := DependencyParts(dep)

		// No duplicates
		if !utils.ListContainsString(chain, stepName) {
			chain = append(chain, stepName)
		}
	}

	for i := 0; i < len(chain); i++ {
		if steps[chain[i]] == nil {
			// 2nd level dependency may not exist
			// if that happens when validating a step, then that probably means that
			// the direct dependency has not been validated yet (non deterministic order).
			// continue to fail gracefully later
			continue
		}
		for _, stepDep := range steps[chain[i]].Dependencies {
			s, _ := DependencyParts(stepDep)

			// Grow the slice and avoid visited nodes
			if !utils.ListContainsString(chain, s) {
				chain = append(chain, s)
			}
		}
	}
	return chain
}

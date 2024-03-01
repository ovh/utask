package step

import (
	"fmt"

	"github.com/juju/errors"
	"github.com/ovh/utask/engine/step/condition"
	"github.com/ovh/utask/pkg/utils"
)

// ValidCondition asserts that the definition for a StepCondition is valid
func ValidCondition(sc *condition.Condition, stepName string, steps map[string]*Step) error {
	if err := sc.Valid(); err != nil {
		return err
	}

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

		customStates, err := impactedStep.GetCustomStates()
		if err != nil {
			return fmt.Errorf("Step custom states are invalid: %s", err)
		}

		validStates := utils.AppendUniq(stepConditionValidStates, customStates...)
		if !utils.ListContainsString(validStates, thenState) {
			return errors.BadRequestf("Step condition implies invalid state for step %s: %s", thenStep, thenState)
		}
	}

	if sc.ForEach != "" {
		if steps[stepName].ForEach == "" {
			return errors.BadRequestf("Step condition cannot set foreach on a non-foreach step")
		}

		if sc.Type != condition.SKIP {
			return errors.BadRequestf("Step condition can set foreach on a skip condition")
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

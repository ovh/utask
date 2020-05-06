package step

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/juju/errors"
	"github.com/sirupsen/logrus"

	"github.com/ovh/utask"
	"github.com/ovh/utask/engine/values"
	"github.com/ovh/utask/pkg/jsonschema"
	"github.com/ovh/utask/pkg/utils"
)

// retry patterns for a step
const (
	RetrySeconds = "seconds"
	RetryMinutes = "minutes"
	RetryHours   = "hours"
)

// possible states of a step
const (
	StateAny           = "ANY" // wildcard
	StateTODO          = "TODO"
	StateRunning       = "RUNNING"
	StateDone          = "DONE"
	StateClientError   = "CLIENT_ERROR"
	StateServerError   = "SERVER_ERROR"
	StateFatalError    = "FATAL_ERROR"
	StateCrashed       = "CRASHED"
	StatePrune         = "PRUNE"
	StateToRetry       = "TO_RETRY"
	StateAfterrunError = "AFTERRUN_ERROR"

	// steps that carry a foreach list of arguments
	StateExpanded = "EXPANDED"
)

const (
	stepRefThis = utask.This

	defaultMaxRetries = 10000
)

var (
	builtinStates            = []string{StateTODO, StateRunning, StateDone, StateClientError, StateServerError, StateFatalError, StateCrashed, StatePrune, StateToRetry, StateAfterrunError, StateAny, StateExpanded}
	stepConditionValidStates = []string{StateDone, StatePrune, StateToRetry, StateFatalError, StateClientError}
	runnableStates           = []string{StateTODO, StateServerError, StateClientError, StateFatalError, StateCrashed, StateToRetry, StateAfterrunError, StateExpanded} // everything but RUNNING, DONE, PRUNE
	retriableStates          = []string{StateServerError, StateToRetry, StateAfterrunError}
)

// Step describes one unit of work within a task, and its dependency to other steps
// a step contains an action that makes use of an available executor, with a specific parameter set
// The result of a step is stored as its output, and can be validated with json schema
// Any error and metadata returned by the step's executor will also be stored, resulting in a state
// The state of a step can be customized by the author of a template, to account for business-specific
// outcomes (eg. a 404 needn't be an error, it can be called NOT_FOUND and determine execution flow
// without blocking).
// Through the "foreach" parameter, a step can be configured to spawn sub-steps for a list of items:
// the result of such a step will be the collection of results of all sub-steps, which can be fed
// into another "foreach" step
// A step can be configured to evaluate "conditions" before and after the action is performed:
// - a "skip" condition will be run before and might determine that the step's action can be skipped entirely
// - a "check" condition will be run after the action, and can control execution flow by examining
//   the step's result and modifying step states through the entire task's resolution
type Step struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Idempotent  bool   `json:"idempotent"`
	// action
	Action Executor `json:"action"`
	// result
	Schema         json.RawMessage         `json:"json_schema,omitempty"`
	ResultValidate jsonschema.ValidateFunc `json:"-"`
	Output         interface{}             `json:"output,omitempty"`
	Metadata       interface{}             `json:"metadata,omitempty"`
	Children       []interface{}           `json:"children,omitempty"`
	Error          string                  `json:"error,omitempty"`
	State          string                  `json:"state,omitempty"`
	// hints about ETA latency, async, for retrier to define strategy
	// how often VS how many times
	RetryPattern string    `json:"retry_pattern,omitempty"` // seconds, minutes, hours
	TryCount     int       `json:"try_count,omitempty"`
	MaxRetries   int       `json:"max_retries,omitempty"`
	LastRun      time.Time `json:"last_run,omitempty"`

	// flow control
	Dependencies []string     `json:"dependencies,omitempty"`
	CustomStates []string     `json:"custom_states,omitempty"`
	Conditions   []*Condition `json:"conditions,omitempty"`
	skipped      bool
	// loop
	ForEach         string          `json:"foreach,omitempty"`        // "parent" step: expression for list of items
	ChildrenSteps   []string        `json:"children_steps,omitempty"` // list of children names
	ChildrenStepMap map[string]bool `json:"children_steps_map,omitempty"`
	Item            interface{}     `json:"item,omitempty"` // "child" step: item value, issued from foreach

	Resources []string `json:"resources"` // resource limits to enforce

	Tags map[string]string `json:"tags"`
}

// Executor matches an executor type with its required configuration
type Executor struct {
	Type              string          `json:"type"`
	BaseConfiguration string          `json:"base_configuration,omitempty"`
	Configuration     json.RawMessage `json:"configuration"`
	BaseOutput        json.RawMessage `json:"base_output"`
}

// Context provides a step with extra metadata about the task
type Context struct {
	RequesterUsername string    `json:"requester_username"`
	ResolverUsername  string    `json:"resolver_username"`
	TaskID            string    `json:"task_id"`
	Created           time.Time `json:"created"`
}

////

func noopStep(st *Step, stepChan chan<- *Step) {
	stepChan <- st
}

func uniqueSortedList(s []string) []string {
	m := make(map[string]struct{})
	for _, str := range s {
		m[str] = struct{}{}
	}

	ret := make([]string, 0, len(m))
	for k := range m {
		ret = append(ret, k)
	}
	sort.Strings(ret)
	return ret
}

// Run carries out the action defined by a Step, by providing values to its configuration
// - a stepChan channel is provided for committing the result back
// - a stopRunningSteps channel is provided to interrupt execution in flight
// values IS NOT CONCURRENT SAFE, DO NOT SHARE WITH OTHER GOROUTINES
func Run(st *Step, baseConfig map[string]json.RawMessage, values *values.Values, stepChan chan<- *Step, wg *sync.WaitGroup, stopRunningSteps <-chan struct{}) {

	// Step already ran, directly going to afterrun process
	if st.State == StateAfterrunError {
		go noopStep(st, stepChan)
		return
	}

	if st.MaxRetries == 0 {
		st.MaxRetries = defaultMaxRetries
	}
	if st.TryCount > st.MaxRetries {
		st.State = StateFatalError
		st.Error = fmt.Sprintf("Step reached max retries %d: %s", st.MaxRetries, st.Error)
		go noopStep(st, stepChan)
		return
	}

	if st.skipped {
		go noopStep(st, stepChan)
		return
	}

	var baseOutput map[string]interface{}

	if len(st.Action.BaseOutput) > 0 {
		base, err := rawResolveObject(values, st.Action.BaseOutput, st.Item, st.Name)
		if err != nil {
			st.State = StateFatalError
			st.Error = errors.Annotate(err, "failed to template base output").Error()
			go noopStep(st, stepChan)
			return
		}
		if base != nil {
			var ok bool
			baseOutput, ok = base.(map[string]interface{})
			if !ok {
				st.State = StateFatalError
				st.Error = errors.Annotate(errors.New("Base output not a map"), "failed to template base output").Error()
				go noopStep(st, stepChan)
				return
			}
		}
	}

	config, err := resolveObject(values, st.Action.Configuration, st.Item, st.Name)
	if err != nil {
		st.State = StateFatalError
		st.Error = errors.Annotate(err, "failed to template configuration").Error()
		go noopStep(st, stepChan)
		return
	}

	var baseCfgRaw json.RawMessage

	if st.Action.BaseConfiguration != "" {
		base, ok := baseConfig[st.Action.BaseConfiguration]
		if !ok {
			st.State = StateFatalError
			st.Error = fmt.Sprintf("could not find base configuration '%s'", st.Action.BaseConfiguration)
			go noopStep(st, stepChan)
			return
		}
		resolvedBase, err := resolveObject(values, base, st.Item, st.Name)
		if err != nil {
			st.State = StateFatalError
			st.Error = errors.Annotate(err, "failed to template base configuration").Error()
			go noopStep(st, stepChan)
			return
		}
		baseCfgRaw = resolvedBase
	}

	runner, err := getRunner(st.Action.Type)
	if err != nil {
		st.State = StateFatalError
		st.Error = err.Error()
		go noopStep(st, stepChan)
		return
	}

	ctx := runner.Context(st.Name)
	if ctx != nil {
		ctxMarshal, err := utils.JSONMarshal(ctx)
		if err != nil {
			st.State = StateFatalError
			st.Error = fmt.Sprintf("failed to marshal context: %s", err.Error())
			go noopStep(st, stepChan)
			return
		}
		ctxTmpl, err := values.Apply(string(ctxMarshal), st.Item, st.Name)
		if err != nil {
			st.State = StateFatalError
			st.Error = fmt.Sprintf("failed to template context: %s", err.Error())
			go noopStep(st, stepChan)
			return
		}
		err = utils.JSONnumberUnmarshal(bytes.NewReader(ctxTmpl), &ctx)
		if err != nil {
			st.State = StateFatalError
			st.Error = fmt.Sprintf("failed to re-marshal context: %s", err.Error())
			go noopStep(st, stepChan)
			return
		}
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		resources := append(runner.Resources(baseCfgRaw, config), st.Resources...)
		limits := uniqueSortedList(resources)
		for _, limit := range limits {
			utask.AcquireResource(limit)
		}

		select {
		case <-stopRunningSteps:
			st.State = StateToRetry
		default:
			st.Output, st.Metadata, st.Tags, err = runner.Exec(st.Name, baseCfgRaw, config, ctx)
			if baseOutput != nil {
				if st.Output != nil {
					marshaled, err := utils.JSONMarshal(st.Output)
					if err == nil {
						_ = utils.JSONnumberUnmarshal(bytes.NewReader(marshaled), &baseOutput)
					}
				}
				st.Output = baseOutput
			}
			if err != nil {
				if errors.IsBadRequest(err) {
					st.State = StateClientError
				} else {
					st.State = StateServerError
				}
				st.Error = err.Error()
			} else if st.ResultValidate != nil {
				if err := st.ResultValidate(st.Output); err != nil {
					st.Error = err.Error()
					st.State = StateFatalError
				}
			}

			if st.State == StateRunning {
				st.State = StateDone
				st.Error = ""
			}

			st.TryCount++
		}

		for _, limit := range limits {
			utask.ReleaseResource(limit)
		}

		stepChan <- st
	}()
}

// StateSetter is a handle to apply the effects of a condition evaluation
type StateSetter func(step, state, message string)

// PreRun evaluates a step's "skip" conditions before the Step's action has been performed
// and impacts the entire task's execution flow through the provided StateSetter
func PreRun(st *Step, values *values.Values, ss StateSetter, executedSteps map[string]bool) {
	for _, sc := range st.Conditions {
		if sc.Type != SKIP {
			continue
		}
		if err := sc.Eval(values, st.Item, st.Name); err != nil {
			if _, ok := err.(ErrConditionNotMet); ok {
				logrus.Debugf("PreRun: Step [%s] condition eval: %s", st.Name, err)
				continue
			} else { // Templating / strconv errors
				// Putting the step in SERVER_ERROR makes the resolution collectable by the RetryCollector.
				ss(st.Name, StateServerError, err.Error())

				// Do not run the step.
				st.skipped = true
				// inserting current skipped step into executedSteps to avoid being picked-up again in availableSteps candidates
				executedSteps[st.Name] = true
				break
			}
		}
		st.skipped = true
		// inserting current skipped step into executedSteps to avoid being picked-up again in availableSteps candidates
		executedSteps[st.Name] = true
		for step, state := range sc.Then {
			if step == stepRefThis {
				step = st.Name
			}
			ss(step, state, sc.Message)
		}
	}
}

// AfterRun evaluates a step's "check" conditions after the Step's action has been performed
// and impacts the entire task's execution flow through the provided StateSetter
func AfterRun(st *Step, values *values.Values, ss StateSetter) {
	if st.skipped || st.State == StateServerError || st.State == StateFatalError || st.ForEach != "" {
		return
	}
	for _, sc := range st.Conditions {
		if sc.Type != CHECK {
			continue
		}
		if err := sc.Eval(values, st.Item, st.Name); err != nil {
			if _, ok := err.(ErrConditionNotMet); ok {
				logrus.Debugf("AfterRun: Step [%s] condition eval: %s", st.Name, err)
				continue
			} else { // Templating / strconv errors
				// Putting the step in AFTERRUN_ERROR makes the resolution collectable (like step.PreRun),
				// but will skip the nexts step.Run, jumping directly to the after-run logic.
				ss(st.Name, StateAfterrunError, err.Error())
				break
			}
		}
		for step, state := range sc.Then {
			if step == stepRefThis {
				step = st.Name
			}
			ss(step, state, sc.Message)
		}
	}
}

// ValidAndNormalize asserts that a step carries correct configuration
// - checks that executor is registered
// - validates retry pattern
// - validates custom states for the step (no collisions with builtin states)
// - validates conditions
// - validates the provided json schema for result validation
// - checks dependency declaration against the task's execution tree
func (st *Step) ValidAndNormalize(name string, baseConfigs map[string]json.RawMessage, steps map[string]*Step) error {

	if name == stepRefThis {
		return errors.BadRequestf("'%s' step name is reserved", stepRefThis)
	}

	// valid action executor
	if err := validExecutor(baseConfigs, st.Action); err != nil {
		return errors.NewNotValid(err, "Invalid executor action")
	}
	// valid retry pattern, accept empty
	switch st.RetryPattern {
	case "", RetrySeconds, RetryMinutes, RetryHours:
	default:
		return errors.NotValidf("Invalid retry pattern: %s Expecting(%s|%s|%s)", st.RetryPattern, RetrySeconds, RetryMinutes, RetryHours)
	}

	// valid custom states
	for _, cState := range st.CustomStates {
		if utils.ListContainsString(builtinStates, cState) {
			return errors.NewNotValid(nil,
				fmt.Sprintf(`Custom state %q is not allowed as it's a reserved state. Reserved state are: "%s"`,
					cState, strings.Join(builtinStates, `", "`)))
		}
	}

	// valid step conditions
	for _, sc := range st.Conditions {
		if err := sc.Valid(name, steps); err != nil {
			return err
		}
	}

	// normalize and validate json schema
	schema, err := jsonschema.NormalizeAndCompile(name, st.Schema)
	if err != nil {
		return errors.Annotatef(err, "Jsonschema: step %s", name)
	}
	st.Schema = schema

	// valid dependencies
	seenDependencies := map[string]bool{}
	for _, d := range st.Dependencies {
		// no orphan dependencies,
		depStep, depState := DependencyParts(d)
		s, ok := steps[depStep]
		if !ok {
			return errors.NotValidf("Invalid dependency, no step with that name: %q", depStep)
		}
		if _, ok := seenDependencies[depStep]; ok {
			return errors.NotValidf("Invalid dependency, already defined dependency to: %q", depStep)
		}
		if duplicated := utils.HasDupsArray(depState); duplicated {
			return errors.NotValidf("Invalid dependency, duplicated state detected")
		}
		for _, state := range depState {
			switch state {
			case StateDone:
			case StateAny:
				if len(depState) != 1 {
					return errors.NotValidf("Invalid dependency, no other state allowed if ANY is declared")
				}
			default:
				if !utils.ListContainsString(s.CustomStates, state) {
					return errors.NotValidf("Invalid dependency on step %s, step state not allowed: %q", depStep, state)
				}
			}
		}
		seenDependencies[depStep] = true
	}

	// no circular dependencies,
	sourceChain := dependenciesChain(steps, st.Dependencies)
	if utils.ListContainsString(sourceChain, name) {
		return errors.NotValidf("Invalid: circular dependency %v <-> %s", sourceChain, st.Name)
	}

	return nil
}

// IsRunnable asserts that Step is in a runnable state
func (st *Step) IsRunnable() bool {
	return utils.ListContainsString(runnableStates, st.State)
}

// IsRetriable asserts that Step is eligible for retry
func (st *Step) IsRetriable() bool {
	return utils.ListContainsString(retriableStates, st.State)
}

// IsFinal asserts that Step is in a final step (not to be run again)
func (st *Step) IsFinal() bool {
	return (st.State != StateRunning && !st.IsRunnable())
}

// IsChild asserts that Step was spawned by a foreach step
func (st *Step) IsChild() bool {
	return st.Item != nil
}

// ExecutorMetadata returns the step's runner metadata schema
func (st *Step) ExecutorMetadata() json.RawMessage {
	runner, err := getRunner(st.Action.Type)
	if err != nil {
		return []byte{}
	}

	return runner.MetadataSchema()
}

func validExecutor(baseConfigs map[string]json.RawMessage, ex Executor) error {
	r, err := getRunner(ex.Type)
	if err != nil {
		return err
	}

	if len(ex.BaseConfiguration) > 0 {
		if _, ok := baseConfigs[ex.BaseConfiguration]; !ok {
			return errors.New("BaseConfiguration key not found")
		}
	}

	return r.ValidConfig(baseConfigs[ex.BaseConfiguration], ex.Configuration)
}

// DependencyParts de-composes a Step's dependency into its constituent parts: step name + step state
// a dependency expressed only as a step name is equivalent to depending on that step being in DONE state
func DependencyParts(dep string) (string, []string) {
	var depStep string
	var depState []string
	parts := strings.SplitN(dep, ":", 2)
	depStep = parts[0]
	if len(parts) == 1 {
		depState = []string{StateDone}
	} else {
		depState = strings.Split(parts[1], ",")
	}
	return depStep, depState
}

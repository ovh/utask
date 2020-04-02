package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/ghodss/yaml"
	"github.com/juju/errors"
	"github.com/loopfz/gadgeto/zesty"
	"github.com/ovh/configstore"
	"github.com/sirupsen/logrus"

	"github.com/ovh/utask"
	"github.com/ovh/utask/engine/step"
	"github.com/ovh/utask/engine/values"
	"github.com/ovh/utask/models/resolution"
	"github.com/ovh/utask/models/runnerinstance"
	"github.com/ovh/utask/models/task"
	"github.com/ovh/utask/models/tasktemplate"
	"github.com/ovh/utask/pkg/jsonschema"
	"github.com/ovh/utask/pkg/now"
	"github.com/ovh/utask/pkg/utils"
)

var (
	// singleton instance
	eng Engine

	// Used for stopping the current Engine
	stopRunningSteps chan struct{}
	gracePeriodEnd   chan struct{}
)

// Engine is the heart of utask: it is the active process
// that handles the lifecycle of every task resolution.
// All the logic for resolution state changes is expressed here
// - the engine determines which steps in a resolution are elligible
// for execution, and which should be pruned
// - the engine computes step dependencies
// - the conditions expressed in a step (skip/check) that impact resolution
// state are enforced by the engine
type Engine struct {
	// items retrieved from configstore are part of the data
	// available to steps during execution
	// ie. credentials needed for http calls, etc...
	config map[string]interface{}
}

// Init launches the task orchestration engine, providing it with a global context
// and with a store from which to inherit configuration items needed for task execution
func Init(ctx context.Context, store *configstore.Store) error {
	cfg, err := utask.Config(store)
	if err != nil {
		return err
	}
	// get all configuration items
	itemList, err := store.GetItemList()
	if err != nil {
		return err
	}
	// drop those that shouldnt be available for task execution
	// (don't let DB credentials leak, for instance...)
	config, err := filteredConfig(itemList, cfg.ConcealedSecrets...)
	if err != nil {
		return err
	}
	// attempt to deserialize json formatted config items
	// -> make it easier to access internal nodes/values when templating
	eng.config = make(map[string]interface{})
	for k, v := range config {
		var i interface{}
		if v != nil {
			err := yaml.Unmarshal([]byte(*v), &i, func(dec *json.Decoder) *json.Decoder {
				dec.UseNumber()
				return dec
			})
			if err != nil {
				eng.config[k] = v
			} else {
				eng.config[k] = i
			}
		}
	}

	// channels for handling graceful shutdown
	stopRunningSteps = make(chan struct{})
	gracePeriodEnd = make(chan struct{})
	go func() {
		<-ctx.Done()
		// Stop running new steps
		close(stopRunningSteps)

		// Wait for the grace period to end
		time.Sleep(3 * time.Second)

		// Set remaining resolutions to resolution.StateCrashed
		close(gracePeriodEnd)
	}()

	// register an engine instance in DB, for synchronization between collectors
	// -> "acquire" tasks without colliding with other running instances
	// this way utask can be scaled horizontally to cope with a higher volume
	// of tasks, and to remain available under high load through a re-deploy
	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return err
	}
	// make this instance's ID globally available
	utask.InstanceID, err = runnerinstance.Create(dbp)
	if err != nil {
		return err
	}

	// initialize all collectors
	// maintenance mode is meant to ensure that no data can change while we
	// perform administration chores, so collectors are switched off
	if !utask.FMaintenanceMode {

		// init garbage collector (delete tasks completed more than x time ago (x from global config) + delete orphaned batches)
		if err := GarbageCollector(ctx, cfg.CompletedTaskExpiration); err != nil {
			return err
		}
		// init autorun collector (create resolution + run for tasks with state == autorun)
		if err := AutorunCollector(ctx); err != nil {
			return err
		}
		// init crashed instance collector
		if err := InstanceCollector(ctx); err != nil {
			return err
		}
		// init retry collector (retry resolutions with state == error)
		if err := RetryCollector(ctx); err != nil {
			return err
		}
	}
	return nil
}

// filteredConfig takes a configstore item list, drops some items by key
// then reduces the result into a map of key->values
func filteredConfig(list *configstore.ItemList, dropAlias ...string) (map[string]*string, error) {
	cfg := make(map[string]*string)
	for _, i := range list.Items {
		if !utils.ListContainsString(dropAlias, i.Key()) {
			// assume only one value per alias
			if _, ok := cfg[i.Key()]; !ok {
				v, err := i.Value()
				if err != nil {
					return nil, err
				}
				if len(v) > 0 {
					cfg[i.Key()] = &v
				}
			}
		}
	}
	return cfg, nil
}

// GetEngine returns the singleton instance of Engine
func GetEngine() Engine {
	return eng
}

// Resolve launches the asynchronous execution of a resolution, given its ID
func (e Engine) Resolve(publicID string) error {
	_, err := e.launchResolution(publicID, true)
	return err
}

// SyncResolve launches the synchronous execution of a resolution, given its ID
func (e Engine) SyncResolve(publicID string) (*resolution.Resolution, error) {
	return e.launchResolution(publicID, false)
}

func (e Engine) launchResolution(publicID string, async bool) (*resolution.Resolution, error) {

	debugLogger := logrus.WithFields(logrus.Fields{"resolution_id": publicID})
	debugLogger.Debugf("Engine: Resolve() starting for %s", publicID)

	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return nil, err
	}

	// check/update states for all concerned objects
	res, t, err := initialize(dbp, publicID)
	if err != nil {
		debugLogger.Debugf("Engine: Resolve() %s initialize error: %s", publicID, err)
		return nil, err
	}

	// If res is nil, this means we are in a state which needs human intervention.
	// Simply abort the automatic resolution.
	if res == nil {
		debugLogger.Debugf("Engine: Resolve() %s nil res, abort", publicID)
		return nil, nil
	}

	res.Values.SetConfig(e.config)

	// all ready, run remaining steps

	utask.AcquireExecutionSlot()

	recap := make([]string, 0)
	for name, s := range res.Steps {
		recap = append(recap, fmt.Sprintf("step %s = %s", name, s.State))
	}

	if res.InstanceID != nil {
		debugLogger = debugLogger.WithField("instance_id", *res.InstanceID)
	}
	debugLogger.Debugf("Engine: Resolve() %s RECAP BEFORE resolve: state: %s, steps: %s", publicID, res.State, strings.Join(recap, ", "))
	if async {
		go resolve(dbp, res, t, debugLogger)
	} else {
		resolve(dbp, res, t, debugLogger)
	}
	return res, nil
}

func initialize(dbp zesty.DBProvider, publicID string) (*resolution.Resolution, *task.Task, error) {
	sp, err := dbp.TxSavepoint()
	defer dbp.RollbackTo(sp)
	if err != nil {
		return nil, nil, err
	}
	res, err := resolution.LoadLockedFromPublicID(dbp, publicID)
	if err != nil {
		return nil, nil, err
	}

	switch res.State {
	case resolution.StateCancelled:
		return nil, nil, errors.NewBadRequest(nil, "Can't run resolution: cancelled")
	case resolution.StateRunning:
		return nil, nil, errors.NewBadRequest(nil, "Can't run resolution: already running")
	case resolution.StateDone:
		return nil, nil, errors.NewBadRequest(nil, "Can't run resolution: already done")
	case resolution.StateCrashed:
		for _, s := range res.Steps {
			if s.State == step.StateRunning {
				if s.Idempotent {
					// if a crashed step is idempotent, repeat
					s.State = step.StateTODO
				} else {
					// otherwise, block the resolution for human intervention
					s.State = step.StateCrashed
					res.SetState(resolution.StateBlockedToCheck)
				}
			}
		}
		if res.State == resolution.StateBlockedToCheck {
			break
		}
		fallthrough
	default:
		res.SetState(resolution.StateRunning)
		res.SetInstanceID(utask.InstanceID)
		res.SetLastStart(now.Get())
		res.IncrementRunCount()
	}

	if err := res.Update(dbp); err != nil {
		return nil, nil, err
	}

	// if crash recover determined the resolution to be blocked, abort run
	if res.State == resolution.StateBlockedToCheck {
		if err := dbp.Commit(); err != nil {
			return nil, nil, err
		}
		return nil, nil, nil
	}

	// otherwise, proceed
	t, err := task.LoadFromID(dbp, res.TaskID)
	if err != nil {
		return nil, nil, err
	}

	t.SetState(task.StateRunning)
	if err := t.Update(dbp, false, true); err != nil {
		if !errors.IsNotValid(err) {
			// not a validation error -> rollback and let a collector re-handle this
			return nil, nil, err
		}
		// task validation error
		// -> possible race condition with updated template
		// put task on hold
		t.SetState(task.StateBlocked)
		if err := t.Update(dbp, true, true); err != nil {
			return nil, nil, err
		}
	}

	// if task could not be updated owing to validation error, abort run
	if t.State == task.StateBlocked {
		if err := dbp.Commit(); err != nil {
			return nil, nil, err
		}
		return nil, nil, nil
	}

	if err := dbp.Commit(); err != nil {
		dbp.Rollback()
		return nil, nil, err
	}

	tt, err := tasktemplate.LoadFromID(dbp, t.TemplateID)
	if err != nil {
		return nil, nil, err
	}

	// provide the resolution with values
	t.ExportTaskInfos(res.Values)
	res.Values.SetInput(t.Input)
	res.Values.SetResolverInput(res.ResolverInput)
	res.Values.SetVariables(tt.Variables)

	return res, t, nil
}

func resolve(dbp zesty.DBProvider, res *resolution.Resolution, t *task.Task, debugLogger *logrus.Entry) {

	// keep track of steps which get executed during each run, to avoid looping+retrying the same failing step endlessly
	executedSteps := map[string]bool{}
	stepChan := make(chan *step.Step)

	expectedMessages := runAvailableSteps(dbp, map[string]bool{}, res, t, stepChan, executedSteps, []string{}, debugLogger)

	for expectedMessages > 0 {
		debugLogger.Debugf("Engine: resolve() %s loop, %d expected steps", res.PublicID, expectedMessages)
		select {
		case s := <-stepChan:
			s.LastRun = time.Now()

			// "commit" step back into resolution
			res.SetStep(s.Name, s)
			// consolidate its result into live values
			res.Values.SetOutput(s.Name, s.Output)
			res.Values.SetMetadata(s.Name, s.Metadata)
			res.Values.SetChildren(s.Name, s.Children)
			res.Values.SetError(s.Name, s.Error)
			res.Values.SetState(s.Name, s.State)

			// call after-run step logic
			modifiedSteps := map[string]bool{
				s.Name: true,
			}
			step.AfterRun(s, res.Values, resolutionStateSetter(res, modifiedSteps))
			pruneSteps(res, modifiedSteps)

			// loop step: kept in the "available" pool, to collect children's results
			if s.ForEach == "" {
				executedSteps[s.Name] = true
			}

			debugLogger.Debugf("Engine: resolve() %s loop, step %s result: %s", res.PublicID, s.Name, s.State)

			// uptate done step count
			// ignore foreach iterations for global done count
			if s.IsFinal() && !s.IsChild() {
				t.StepsDone++
			}

			// one less step to go
			expectedMessages--
			// state change might unlock more steps for execution
			expectedMessages += runAvailableSteps(dbp, modifiedSteps, res, t, stepChan, executedSteps, []string{}, debugLogger)

			// attempt to persist all changes in db
			if err := commit(dbp, res, t); err != nil {
				debugLogger.Debugf("Engine: resolve() %s loop, FAILED TO COMMIT RESOLUTION: %s", res.PublicID, err)
			} else {
				debugLogger.Debugf("Engine: resolve() %s loop, COMMIT DONE: step: %s = %s", res.PublicID, s.Name, s.State)
			}
		case <-gracePeriodEnd:
			// shutting down, time is up: exit the loop no matter how many steps might be pending
			break
		}
	}

	inShutdown := false
	select {
	case <-stopRunningSteps:
		inShutdown = true
	default:
	}

	// we exited the step selection loop, assume we got to the very end
	allDone := true
	doneCount := 0

	// review all step states, collect potential resolution states
	mapStatus := map[string]bool{}
	for _, s := range res.Steps {
		switch s.State {
		case step.StateClientError:
			mapStatus[resolution.StateBlockedBadRequest] = true
			allDone = false
		case step.StateFatalError:
			mapStatus[resolution.StateBlockedFatal] = true
			allDone = false
		case step.StateServerError, step.StateToRetry, step.StateAfterrunError:
			// setting the resolution to StateError makes it collectable by the retry collector
			mapStatus[resolution.StateError] = true
			allDone = false
		case step.StateRunning:
			mapStatus[resolution.StateCrashed] = true
			allDone = false
		case step.StateTODO:
			// instance is in shutdown mode, the resolution may have been interrupted
			// set to crashed for proper retry
			if inShutdown {
				mapStatus[resolution.StateCrashed] = true
			} else { // otherwise, this points to an unsolvable resolution (unmet dependencies)
				mapStatus[resolution.StateBlockedDeadlock] = true
			}
			allDone = false
		}
		if s.IsFinal() && !s.IsChild() {
			doneCount++
		}
	}
	t.StepsDone = doneCount

	// compute resolution state
	if !allDone {
		// from candidate resolution states, choose a resolution state by priority
		for _, status := range []string{resolution.StateCrashed, resolution.StateBlockedFatal, resolution.StateBlockedBadRequest, resolution.StateError, resolution.StateBlockedDeadlock} {
			if mapStatus[status] {
				res.SetState(status)
				break
			}
		}
	} else {
		// all done -> resolution is done
		res.SetState(resolution.StateDone)
		t.SetState(task.StateDone)
		if err := t.SetResult(res.Values); err != nil {
			debugLogger.Debugf("Engine: resolve() %s loop, task SetResult error: %s", res.PublicID, err)
		}
	}

	// further qualify a resolution in error state -> give hints to collectors, change task state if intervention required
	switch res.State {
	case resolution.StateError, resolution.StateCrashed:
		if res.RunCount >= res.RunMax {
			res.SetState(resolution.StateBlockedMaxRetries)
			t.SetState(task.StateBlocked)
		} else {
			res.NextRetry = nextRetry(res)
		}
	case resolution.StateBlockedBadRequest, resolution.StateBlockedFatal, resolution.StateBlockedDeadlock:
		t.SetState(task.StateBlocked)
	}

	// finalize metadata collection
	res.SetLastStop(now.Get())

	recapLog := make([]string, 0)
	recapLog = append(recapLog, fmt.Sprintf("Engine: resolve() %s END OF LOOP, state: %s", res.PublicID, res.State))
	for stepK, stepV := range res.Steps {
		recapLog = append(recapLog, fmt.Sprintf("step %s = %s", stepK, stepV.State))
	}

	debugLogger.Debugf(strings.Join(recapLog, ", "))

	bkoff := backoff.NewExponentialBackOff()
	bkoff.InitialInterval = time.Second
	bkoff.Multiplier = 2
	bkoff.MaxInterval = 30 * time.Second

	bkoff.Reset()

	for {
		err := commit(dbp, res, t)
		if err != nil {
			debugLogger.Debugf("Engine: resolve() %s final commit error: %s", res.PublicID, err)
		} else {
			debugLogger.Debugf("Engine: resolve() %s final commit done", res.PublicID)
			break
		}
		time.Sleep(bkoff.NextBackOff())
	}

	utask.ReleaseExecutionSlot()
}

func commit(dbp zesty.DBProvider, res *resolution.Resolution, t *task.Task) error {
	sp, err := dbp.TxSavepoint()
	defer dbp.RollbackTo(sp)
	if err != nil {
		return err
	}
	if err := res.Update(dbp); err != nil {
		return err
	}
	if err := t.Update(dbp, false, true); err != nil {
		return err
	}
	return dbp.Commit()
}

func runAvailableSteps(dbp zesty.DBProvider, modifiedSteps map[string]bool, res *resolution.Resolution, t *task.Task, stepChan chan<- *step.Step, executedSteps map[string]bool, expandedSteps []string, debugLogger *logrus.Entry) int {
	av := availableSteps(modifiedSteps, res, executedSteps, expandedSteps, debugLogger)
	preRunModifiedSteps := map[string]bool{}
	expanded := 0

	select {
	case <-stopRunningSteps:
		return 0
	default:
		for name, s := range av {
			// prepare step
			s.Name = name
			if s.ForEach != "" { // loop step
				switch s.State {
				case step.StateTODO:
					expanded++
					expandStep(s, res)
					expandedSteps = append(expandedSteps, s.ChildrenSteps...)
				case step.StateToRetry:
					// attempt contracting step, clean up any children steps
					// any available children have been ignored by availableSteps()
					if s.ChildrenSteps != nil && len(s.ChildrenSteps) > 0 {
						contractStep(s, res)
						s.Output = nil
					} else {
						expanded++
						expandStep(s, res)
						expandedSteps = append(expandedSteps, s.ChildrenSteps...)
					}
				case step.StateExpanded:
					s.State = contractStep(s, res)
				default:
					continue
				}
				// rebuild step dependency tree to include generated loop steps
				res.BuildStepTree()
				commit(dbp, res, t)
				go func() { stepChan <- s }()
			} else { // regular step
				s.ResultValidate = jsonschema.Validator(s.Name, s.Schema)

				// skip prerun
				// TODO fixme, ugly
				// juggling with STATE_AFTERRUN_ERROR should probably only be inside step pkg
				if s.State != step.StateAfterrunError {
					s.State = step.StateRunning
					step.PreRun(s, res.Values, resolutionStateSetter(res, preRunModifiedSteps))
					commit(dbp, res, t)
				}

				// run
				stepCopy := *s
				step.Run(&stepCopy, res.BaseConfigurations, res.Values, stepChan, stopRunningSteps)
			}
		}
	}

	// look for more available steps in case:
	// - prerun impacted states -> dependencies were unlocked
	// - loop step generated new steps
	if len(preRunModifiedSteps) > 0 || expanded > 0 {
		pruneSteps(res, preRunModifiedSteps)
		return len(av) + runAvailableSteps(dbp, preRunModifiedSteps, res, t, stepChan, executedSteps, expandedSteps, debugLogger)
	}

	return len(av)
}

func expandStep(s *step.Step, res *resolution.Resolution) {
	foreach, err := res.Values.Apply(s.ForEach, nil, "")
	if err != nil {
		s.State = step.StateFatalError
		s.Error = err.Error()
		return
	}
	// unmarshal into collection
	var items []interface{}
	if err := utils.JSONnumberUnmarshal(bytes.NewReader(foreach), &items); err != nil {
		s.State = step.StateFatalError
		s.Error = err.Error()
		return
	}
	// generate all children steps
	for i, item := range items {
		childStepName := fmt.Sprintf("%s-%d", s.Name, i)
		res.Steps[childStepName] = &step.Step{
			Name:         childStepName,
			Description:  fmt.Sprintf("%d - %s", i, s.Description),
			Action:       s.Action,
			Schema:       s.Schema,
			State:        step.StateTODO,
			RetryPattern: s.RetryPattern,
			MaxRetries:   s.MaxRetries,
			Dependencies: s.Dependencies,
			CustomStates: s.CustomStates,
			Conditions:   s.Conditions,
			Item:         item,
		}
	}
	// update parent dependencies to wait on children
	s.ChildrenSteps = []string{}
	s.ChildrenStepMap = map[string]bool{}
	for i := range items {
		childStepName := fmt.Sprintf("%s-%d", s.Name, i)
		s.Dependencies = append(s.Dependencies, childStepName+":ANY")
		s.ChildrenSteps = append(s.ChildrenSteps, childStepName)
		s.ChildrenStepMap[childStepName] = true
	}
	s.State = step.StateExpanded
}

func contractStep(s *step.Step, res *resolution.Resolution) string {
	resultingLoopState := step.StateDone

	// collect results, metadata and errors
	collectedChildren := []interface{}{}
	for _, childStepName := range s.ChildrenSteps {
		child, ok := res.Steps[childStepName]
		if ok {
			if child.State != step.StatePrune {
				childM := map[string]interface{}{}
				if child.Output != nil {
					childM[values.OutputKey] = child.Output
				}
				if child.Metadata != nil {
					childM[values.MetadataKey] = child.Metadata
				}
				var i interface{} = childM
				collectedChildren = append(collectedChildren, i)
			}
			delete(res.Steps, childStepName)
		}
	}
	s.Error = ""
	s.Children = collectedChildren

	// clean up dependency on children
	var cleanDependencies []string
	for _, dep := range s.Dependencies {
		stepName, _ := step.DependencyParts(dep)
		if !s.ChildrenStepMap[stepName] {
			cleanDependencies = append(cleanDependencies, dep)
		}
	}
	s.Dependencies = cleanDependencies
	s.ChildrenSteps = nil
	s.ChildrenStepMap = nil

	return resultingLoopState
}

func pruneSteps(res *resolution.Resolution, modifiedSteps map[string]bool) {
	recursiveModif := map[string]bool{}
	for stepName := range modifiedSteps {
		if res.Steps[stepName].State == step.StatePrune {
			// use StepTreeIndexPrune: lists dependencies which should be pruned when their parent is pruned
			for _, dep := range res.StepTreeIndexPrune[stepName] {
				res.Steps[dep].State = step.StatePrune
				recursiveModif[dep] = true
			}
		} else {
			if res.Steps[stepName].IsFinal() {
				// use StepTreeIndexPrune: lists dependencies which should be pruned when their parent is pruned
				for _, childStep := range res.StepTreeIndexPrune[stepName] {
					for _, childDep := range res.Steps[childStep].Dependencies {
						depStep, depState := step.DependencyParts(childDep)
						if depStep == stepName {
							if res.Steps[stepName].State != depState {
								res.Steps[childStep].State = step.StatePrune
								recursiveModif[childStep] = true
							}
						}
					}
				}
			}
		}
	}
	if len(recursiveModif) > 0 {
		pruneSteps(res, recursiveModif)
	}
}

func availableSteps(modifiedSteps map[string]bool, res *resolution.Resolution, executedSteps map[string]bool, expandedSteps []string, debugLogger *logrus.Entry) map[string]*step.Step {

	// pre-filter candidate steps
	// prioritize those which depended on modified steps
	candidateSteps := map[string]struct{}{}
	if len(modifiedSteps) > 0 {
		for modifStep := range modifiedSteps {
			modifDeps, ok := res.StepTreeIndex[modifStep]
			if ok {
				for _, modifDep := range modifDeps {
					candidateSteps[modifDep] = struct{}{}
				}
			}
		}
	} else {
		for _, s := range res.StepList {
			candidateSteps[s] = struct{}{}
		}
	}
	// looping on just created steps from an EXPANDED step, to verify if they are eligible
	// (in case we had modifiedSteps at the same time)
	for _, s := range expandedSteps {
		candidateSteps[s] = struct{}{}
	}

	// look for runnable steps among candidates
	// make sure their dependencies are met
	available := make(map[string]*step.Step)
	availableLoops := make([]*step.Step, 0)
	for name := range candidateSteps {
		s := res.Steps[name]
		if s.IsRunnable() {
			if executedSteps[name] {
				continue
			}
			elligible := true // elligible unless dependencies are not met
			for _, dep := range s.Dependencies {
				depStep, depState := step.DependencyParts(dep)
				if (res.Steps[depStep].State != depState) && // elligibility check: step state matches dependency target
					!(depState == step.StateAny && res.Steps[depStep].IsFinal()) { // eligibility check: depdency on ANY with step in a final state {
					// a loop step which gets retried should ignore previous children.
					// children steps are stored as dependencies, so we check if the unmet dependency is a child.
					// if it is, we ignore it unless it is already running (to avoid weird behavior when the result comes back)
					if s.ForEach != "" && // it's a loop step
						s.State != step.StateExpanded && // expanded is the only state in which it may have a legit dependency on its children
						s.ChildrenStepMap[depStep] && // the unmet dependency is indeed a child of the step
						res.Steps[depStep].State != step.StateRunning { // the child is not running currently
						continue
					}
					// in every other case, an unmet dependency
					elligible = false
					break
				}
			}
			if elligible {
				available[name] = s
				if s.ForEach != "" {
					availableLoops = append(availableLoops, s)
				}
			}
		}
	}

	// when a loop step is considered as available, any children it may have
	// from a previous run can be discarded
	// some of these children may still be eligible (TODO, etc). We force them out by removing them of the
	// list of available steps, because they will get trashed anyway.
	for _, l := range availableLoops {
		for _, ch := range l.ChildrenSteps {
			delete(available, ch)
		}
	}

	recap := make([]string, 0)
	for av, stp := range available {
		recap = append(recap, fmt.Sprintf("step %s = %s", av, stp.State))
	}

	if len(recap) > 0 {
		debugLogger.Debugf("Engine: availableSteps(): %s", strings.Join(recap, ", "))
	}
	return available
}

func nextRetry(res *resolution.Resolution) *time.Time {
	stepsToRetry := []*step.Step{}
	for _, s := range res.Steps {
		if s.IsRetriable() {
			stepsToRetry = append(stepsToRetry, s)
		}
	}

	// find the shortest retry delay among failed steps (default to seconds)
	fromNow := time.Hour
	for _, s := range stepsToRetry {
		switch s.RetryPattern {
		case step.RetrySeconds, "":
			fromNow = minDuration(fromNow, computeDelay(time.Second, s.TryCount))
		case step.RetryMinutes:
			fromNow = minDuration(fromNow, computeDelay(time.Minute, s.TryCount))
		case step.RetryHours:
			fromNow = minDuration(fromNow, computeDelay(time.Hour, s.TryCount))
		}
	}

	nextRetry := now.Get().Add(fromNow)
	return &nextRetry
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

// for retryCount:    1,  2,  3,  4,  5,  6......
// seconds: 	      10, 17, 21, 24, 26, 28.....
// minutes and hours: 1,  1,  7,  11, 14, 16.....
func computeDelay(d time.Duration, rc int) time.Duration {
	// shorter two first retries for "minutes" and "hours" retry patterns
	if d >= time.Minute {
		if rc <= 2 {
			return d
		}
		rc--
	}
	ret := time.Duration(float64(d) * math.Log(float64(rc)) * 10)
	// baseline 10s for "seconds" retry pattern
	if d < time.Minute {
		ret += 10 * d
	}
	return ret
}

func resolutionStateSetter(res *resolution.Resolution, modifiedSteps map[string]bool) step.StateSetter {
	return func(step, state, message string) {
		if _, ok := res.Steps[step]; ok {
			res.Steps[step].State = state
			res.Values.SetState(step, state)
			res.Steps[step].Error = message
			res.Values.SetError(step, message)
			modifiedSteps[step] = true
		}
	}
}

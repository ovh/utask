package handler

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/juju/errors"
	"github.com/loopfz/gadgeto/zesty"
	"github.com/sirupsen/logrus"

	"github.com/ovh/utask"
	"github.com/ovh/utask/engine"
	"github.com/ovh/utask/engine/step"
	"github.com/ovh/utask/models/resolution"
	"github.com/ovh/utask/models/task"
	"github.com/ovh/utask/models/tasktemplate"
	"github.com/ovh/utask/pkg/auth"
	"github.com/ovh/utask/pkg/metadata"
)

type createResolutionIn struct {
	TaskID         string                 `json:"task_id" binding:"required"`
	ResolverInputs map[string]interface{} `json:"resolver_inputs"`
	StartOver      bool                   `json:"start_over"`
}

// CreateResolution handles the creation of a resolution for a given task
// the creator of the resolution (aka "resolver") might have to provide extra inputs
// depending on the task's template definition
func CreateResolution(c *gin.Context, in *createResolutionIn) (*resolution.Resolution, error) {
	metadata.AddActionMetadata(c, metadata.TaskID, in.TaskID)

	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return nil, err
	}

	if err := dbp.Tx(); err != nil {
		return nil, err
	}

	t, err := task.LoadLockedFromPublicID(dbp, in.TaskID)
	if err != nil {
		dbp.Rollback()
		return nil, err
	}

	if in.StartOver && t.Resolution == nil {
		_ = dbp.Rollback()
		return nil, errors.BadRequestf("can't start over a task that hasn't been resolved at least once")
	}

	tt, err := tasktemplate.LoadFromID(dbp, t.TemplateID)
	if err != nil {
		_ = dbp.Rollback()
		return nil, err
	}

	admin := auth.IsAdmin(c) == nil
	resolutionManager := auth.IsResolutionManager(c, tt, t, nil) == nil

	if !admin && !resolutionManager {
		_ = dbp.Rollback()
		return nil, errors.Forbiddenf("You are not allowed to resolve this task")
	} else if !resolutionManager {
		metadata.SetSUDO(c)
	}

	resUser := auth.GetIdentity(c)

	// adding current resolver to task.resolver_usernames, to be able to list resolved tasks
	// as 'resolvable', if current resolver used admins privileges.
	if auth.IsAdmin(c) == nil {
		t.ResolverUsernames = append(t.ResolverUsernames, resUser)
		if err := t.Update(dbp, false, false); err != nil {
			dbp.Rollback()
			return nil, err
		}
	}

	if in.StartOver {
		if !admin && resolutionManager && !tt.AllowTaskStartOver {
			_ = dbp.Rollback()
			return nil, errors.Forbiddenf("You are not allowed to start over this task")
		}

		res, err := resolution.LoadLockedFromPublicID(dbp, *t.Resolution)
		if err != nil {
			_ = dbp.Rollback()
			return nil, err
		}

		if res.State != resolution.StatePaused && res.State != resolution.StateBlockedBadRequest && res.State != resolution.StateCancelled {
			_ = dbp.Rollback()
			return nil, errors.BadRequestf("can't start over a task that isn't in status %q, %q or %q", resolution.StatePaused, resolution.StateBlockedBadRequest, resolution.StateCancelled)
		}

		logrus.WithFields(logrus.Fields{"resolution_id": res.PublicID, "task_id": t.PublicID}).Debugf("Handler CreateResolution: start-over the resolution, deleting old resolution %s", res.PublicID)

		if err := res.Delete(dbp); err != nil {
			_ = dbp.Rollback()
			return nil, err
		}
	}

	r, err := resolution.Create(dbp, t, in.ResolverInputs, resUser, true, nil) // TODO accept delay in handler
	if err != nil {
		dbp.Rollback()
		return nil, err
	}

	metadata.AddActionMetadata(c, metadata.ResolutionID, r.PublicID)
	logrus.WithFields(logrus.Fields{"resolution_id": r.PublicID}).Debugf("Handler CreateResolution: created resolution %s", r.PublicID)

	if err := dbp.Commit(); err != nil {
		dbp.Rollback()
		return nil, err
	}

	return r, nil
}

const (
	resolutionTypeOwn = "own"
	resolutionTypeAll = "all"
)

type listResolutionsIn struct {
	Type       string  `query:"type, default=own"`
	State      *string `query:"state"`
	InstanceID *uint64 `query:"instance_id"`
	PageSize   uint64  `query:"page_size"`
	Last       *string `query:"last"`
}

// ListResolutions returns a list of resolutions, which can be filtered by state and instance ID
// type=own (default option) filters resolutions for which the user is the "resolver"
// type=all returns every resolution, provided that the user is an administrator
// the resolutions are simplified and do not include the content of steps
func ListResolutions(c *gin.Context, in *listResolutionsIn) (rr []*resolution.Resolution, err error) {
	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return nil, err
	}

	resUser := auth.GetIdentity(c)

	in.PageSize = normalizePageSize(in.PageSize)

	switch in.Type {
	case resolutionTypeOwn:
		rr, err = resolution.ListResolutions(dbp, nil, &resUser, in.State, in.InstanceID, in.PageSize, in.Last)
	case resolutionTypeAll:
		if err := auth.IsAdmin(c); err != nil {
			return nil, err
		}
		rr, err = resolution.ListResolutions(dbp, nil, nil, in.State, in.InstanceID, in.PageSize, in.Last)
	default:
		return nil, errors.BadRequestf("Unknown type for listing: '%s'. Was expecting '%s' or '%s'", in.Type, resolutionTypeOwn, resolutionTypeAll)
	}
	if err != nil {
		return nil, err
	}

	if uint64(len(rr)) == in.PageSize {
		lastRes := rr[len(rr)-1].PublicID
		c.Header(
			linkHeader,
			buildResolutionNextLink(in.Type, in.State, in.InstanceID, in.PageSize, lastRes),
		)
	}

	c.Header(pageSizeHeader, fmt.Sprintf("%v", in.PageSize))

	return rr, nil
}

type getResolutionIn struct {
	PublicID string `path:"id, required"`
}

// GetResolution returns a single resolution, with its full content (all step outputs included)
func GetResolution(c *gin.Context, in *getResolutionIn) (*resolution.Resolution, error) {
	metadata.AddActionMetadata(c, metadata.ResolutionID, in.PublicID)

	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return nil, err
	}

	r, err := resolution.LoadFromPublicID(dbp, in.PublicID)
	if err != nil {
		return nil, err
	}

	t, err := task.LoadFromID(dbp, r.TaskID)
	if err != nil {
		return nil, err
	}

	metadata.AddActionMetadata(c, metadata.TaskID, t.PublicID)

	tt, err := tasktemplate.LoadFromID(dbp, t.TemplateID)
	if err != nil {
		return nil, err
	}

	metadata.AddActionMetadata(c, metadata.TemplateName, tt.Name)

	admin := auth.IsAdmin(c) == nil
	requester := auth.IsRequester(c, t) == nil
	watcher := auth.IsWatcher(c, t) == nil
	resolutionManager := auth.IsResolutionManager(c, tt, t, r) == nil

	if !admin && !requester && !watcher && !resolutionManager {
		return nil, errors.Forbiddenf("Can't display resolution details")
	}

	if !resolutionManager && !admin {
		r.ClearOutputs()
	}

	if !resolutionManager && !requester && !watcher {
		metadata.SetSUDO(c)
	}

	return r, nil
}

type updateResolutionIn struct {
	PublicID       string                 `path:"id, required"`
	Steps          map[string]*step.Step  `json:"steps"` // persisted in encrypted blob
	ResolverInputs map[string]interface{} `json:"resolver_inputs"`
}

// UpdateResolution is a special handler reserved to administrators, which allows the
// edition of a live resolution, in case a template mistake needs to be hotfixed
// use sparingly, this opens the door to completely breaking execution
// can only be called when resolution is in state PAUSED
func UpdateResolution(c *gin.Context, in *updateResolutionIn) error {
	metadata.AddActionMetadata(c, metadata.ResolutionID, in.PublicID)

	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return err
	}

	if err := dbp.Tx(); err != nil {
		return err
	}

	r, err := resolution.LoadLockedNoWaitFromPublicID(dbp, in.PublicID)
	if err != nil {
		dbp.Rollback()
		return err
	}

	t, err := task.LoadFromID(dbp, r.TaskID)
	if err != nil {
		dbp.Rollback()
		return err
	}

	metadata.AddActionMetadata(c, metadata.TaskID, t.PublicID)

	tt, err := tasktemplate.LoadFromID(dbp, t.TemplateID)
	if err != nil {
		_ = dbp.Rollback()
		return err
	}

	metadata.AddActionMetadata(c, metadata.TemplateName, tt.Name)

	if r.State != resolution.StatePaused {
		dbp.Rollback()
		return errors.BadRequestf("Cannot update a resolution which is not in state '%s'", resolution.StatePaused)
	}

	if err := auth.IsAdmin(c); err != nil {
		dbp.Rollback()
		return err
	}

	metadata.SetSUDO(c)

	if in.Steps != nil {
		r.Steps = in.Steps

		tt, err := tasktemplate.LoadFromID(dbp, t.TemplateID)
		if err != nil {
			_ = dbp.Rollback()
			return err
		}

		// valid and normalize steps
		for name, st := range r.Steps {
			if err := st.ValidAndNormalize(name, tt.BaseConfigurations, r.Steps); err != nil {
				_ = dbp.Rollback()
				return errors.NewNotValid(err, fmt.Sprintf("invalid step %s", name))
			}

			valid, err := st.CheckIfValidState()
			if err != nil {
				_ = dbp.Rollback()
				return err
			}

			if !valid {
				_ = dbp.Rollback()
				return errors.NewBadRequest(nil, fmt.Sprintf("step %q: invalid state provided: %q is not allowed", name, st.State))
			}
		}
	}

	if in.ResolverInputs != nil {
		r.SetInput(in.ResolverInputs)
	}

	logrus.WithFields(logrus.Fields{"resolution_id": r.PublicID}).Debugf("Handler UpdateResolution: manual update of resolution %s", r.PublicID)

	if err := r.Update(dbp); err != nil {
		dbp.Rollback()
		return err
	}

	reqUsername := auth.GetIdentity(c)
	_, err = task.CreateComment(dbp, t, reqUsername, "manually updated resolution")
	if err != nil {
		dbp.Rollback()
		return err
	}

	if err := dbp.Commit(); err != nil {
		dbp.Rollback()
		return err
	}

	return nil
}

type runResolutionIn struct {
	PublicID string `path:"id, required"`
}

// RunResolution launches the asynchronous execution of a resolution
// the engine determines if resolution is eligible for execution
func RunResolution(c *gin.Context, in *runResolutionIn) error {
	metadata.AddActionMetadata(c, metadata.ResolutionID, in.PublicID)

	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return err
	}

	// not using a LoadLocked here, because GetEngine().Resolve will lock row
	r, err := resolution.LoadFromPublicID(dbp, in.PublicID)
	if err != nil {
		return err
	}

	t, err := task.LoadFromID(dbp, r.TaskID)
	if err != nil {
		return err
	}

	metadata.AddActionMetadata(c, metadata.TaskID, t.PublicID)

	tt, err := tasktemplate.LoadFromID(dbp, t.TemplateID)
	if err != nil {
		return err
	}

	metadata.AddActionMetadata(c, metadata.TemplateName, tt.Name)

	admin := auth.IsAdmin(c) == nil
	resolutionManager := auth.IsResolutionManager(c, tt, t, r) == nil

	if !admin && !resolutionManager {
		return errors.Forbiddenf("You are not allowed to resolve this task")
	} else if !resolutionManager {
		metadata.SetSUDO(c)
	}

	reqUsername := auth.GetIdentity(c)
	_, err = task.CreateComment(dbp, t, reqUsername, "manually ran resolution")
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{"resolution_id": r.PublicID}).Debugf("Handler RunResolution: manual resolve %s", r.PublicID)

	ch := make(chan struct{})
	go func() {
		err = engine.GetEngine().Resolve(in.PublicID, nil)
		close(ch)
	}()

	timeout := time.NewTicker(5 * time.Second)
	defer timeout.Stop()

	// manual resolution can be blocked by a lock acquisition on the Execution pool
	// waiting for 5 seconds to get a response, otherwise let's consider the task will
	// start correctly when the Execution pool gets available, and prevent API thread to be blocked
	select {
	case <-ch:
		return err
	case <-timeout.C:
		return nil
	}
}

type extendResolutionIn struct {
	PublicID string `path:"id, required"`
}

// ExtendResolution increments a resolution's remaining execution retries
// in case it has reached state BLOCKED_MAXRETRIES
func ExtendResolution(c *gin.Context, in *extendResolutionIn) error {
	metadata.AddActionMetadata(c, metadata.ResolutionID, in.PublicID)

	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return err
	}

	if err := dbp.Tx(); err != nil {
		return err
	}

	r, err := resolution.LoadLockedNoWaitFromPublicID(dbp, in.PublicID)
	if err != nil {
		dbp.Rollback()
		return err
	}

	t, err := task.LoadFromID(dbp, r.TaskID)
	if err != nil {
		dbp.Rollback()
		return err
	}

	metadata.AddActionMetadata(c, metadata.TaskID, t.PublicID)

	tt, err := tasktemplate.LoadFromID(dbp, t.TemplateID)
	if err != nil {
		dbp.Rollback()
		return err
	}

	metadata.AddActionMetadata(c, metadata.TemplateName, tt.Name)

	admin := auth.IsAdmin(c) == nil
	resolutionManager := auth.IsResolutionManager(c, tt, t, r) == nil

	if !admin && !resolutionManager {
		dbp.Rollback()
		return errors.Forbiddenf("Not allowed to extend resolution")
	} else if !resolutionManager {
		metadata.SetSUDO(c)
	}

	if r.State != resolution.StateBlockedMaxRetries {
		dbp.Rollback()
		return errors.BadRequestf("Cannot extend a resolution which is not in state '%s'", resolution.StateBlockedMaxRetries)
	}

	reqUsername := auth.GetIdentity(c)
	_, err = task.CreateComment(dbp, t, reqUsername, "manually extended resolution")
	if err != nil {
		dbp.Rollback()
		return err
	}

	if tt.RetryMax != nil {
		r.ExtendRunMax(*tt.RetryMax)
	} else {
		r.ExtendRunMax(utask.DefaultRetryMax)
	}

	r.SetState(resolution.StateError)
	r.SetNextRetry(time.Now())

	if err := r.Update(dbp); err != nil {
		dbp.Rollback()
		return err
	}

	if err := dbp.Commit(); err != nil {
		dbp.Rollback()
		return err
	}

	return nil
}

type cancelResolutionIn struct {
	PublicID string `path:"id, required"`
}

// CancelResolution "kills" a live resolution and its corresponding task,
// rendering it non-runnable (and garbage-collectable)
func CancelResolution(c *gin.Context, in *cancelResolutionIn) error {
	metadata.AddActionMetadata(c, metadata.ResolutionID, in.PublicID)

	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return err
	}

	if err := dbp.Tx(); err != nil {
		return err
	}

	r, err := resolution.LoadLockedNoWaitFromPublicID(dbp, in.PublicID)
	if err != nil {
		dbp.Rollback()
		return err
	}

	t, err := task.LoadFromPublicID(dbp, r.TaskPublicID)
	if err != nil {
		dbp.Rollback()
		return err
	}

	metadata.AddActionMetadata(c, metadata.TaskID, t.PublicID)

	tt, err := tasktemplate.LoadFromID(dbp, t.TemplateID)
	if err != nil {
		dbp.Rollback()
		return err
	}

	metadata.AddActionMetadata(c, metadata.TemplateName, tt.Name)

	admin := auth.IsAdmin(c) == nil
	resolutionManager := auth.IsResolutionManager(c, tt, t, r) == nil

	if !admin && !resolutionManager {
		dbp.Rollback()
		return errors.Forbiddenf("You are not allowed to cancel this task")
	} else if !resolutionManager {
		metadata.SetSUDO(c)
	}

	switch r.State {
	case resolution.StateCancelled, resolution.StateRunning, resolution.StateDone:
		dbp.Rollback()
		return errors.BadRequestf("Can't cancel resolution: state %s", r.State)
	}

	r.SetState(resolution.StateCancelled)

	if err := r.Update(dbp); err != nil {
		dbp.Rollback()
		return err
	}

	t.SetState(task.StateCancelled)

	if err := t.Update(dbp, true, true); err != nil {
		dbp.Rollback()
		return err
	}

	reqUsername := auth.GetIdentity(c)
	_, err = task.CreateComment(dbp, t, reqUsername, "cancelled resolution")
	if err != nil {
		dbp.Rollback()
		return err
	}

	if err := dbp.Commit(); err != nil {
		dbp.Rollback()
		return err
	}

	return nil
}

type pauseResolutionIn struct {
	PublicID string `path:"id, required"`
	Force    bool   `query:"force"`
}

// PauseResolution sets a resolution's state to PAUSED
// allowing for it to be edited
// this action can only be performed by administrators
// and can be "forced" when dealing with exceptions in which a resolution doesn't exit RUNNING state
func PauseResolution(c *gin.Context, in *pauseResolutionIn) error {
	metadata.AddActionMetadata(c, metadata.ResolutionID, in.PublicID)

	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return err
	}

	if err := dbp.Tx(); err != nil {
		return err
	}

	r, err := resolution.LoadLockedNoWaitFromPublicID(dbp, in.PublicID)
	if err != nil {
		dbp.Rollback()
		return err
	}

	t, err := task.LoadFromID(dbp, r.TaskID)
	if err != nil {
		dbp.Rollback()
		return err
	}

	metadata.AddActionMetadata(c, metadata.TaskID, t.PublicID)

	tt, err := tasktemplate.LoadFromID(dbp, t.TemplateID)
	if err != nil {
		dbp.Rollback()
		return err
	}

	metadata.AddActionMetadata(c, metadata.TemplateName, tt.Name)

	admin := auth.IsAdmin(c) == nil
	resolutionManager := auth.IsResolutionManager(c, tt, t, r) == nil

	if !admin && !resolutionManager {
		dbp.Rollback()
		return errors.Forbiddenf("You are not allowed to pause this task")
	} else if !resolutionManager {
		metadata.SetSUDO(c)
	}

	if in.Force {
		if err := auth.IsAdmin(c); err != nil {
			dbp.Rollback()
			return err
		}
	} else {
		switch r.State {
		case resolution.StateCancelled, resolution.StateRunning, resolution.StateDone:
			dbp.Rollback()
			return errors.BadRequestf("Can't pause resolution while in state %s", r.State)
		}
	}

	r.SetState(resolution.StatePaused)

	logrus.WithFields(logrus.Fields{"resolution_id": r.PublicID}).Debugf("Handler PauseResolution: pause of resolution %s", r.PublicID)

	if err := r.Update(dbp); err != nil {
		dbp.Rollback()
		return err
	}

	reqUsername := auth.GetIdentity(c)
	_, err = task.CreateComment(dbp, t, reqUsername, "manually paused resolution")
	if err != nil {
		dbp.Rollback()
		return err
	}

	if err := dbp.Commit(); err != nil {
		dbp.Rollback()
		return err
	}

	return nil
}

type getResolutionStepIn struct {
	PublicID string `path:"id" validate:"required"`
	StepName string `path:"stepName" validate:"required"`
}

// GetResolutionStep returns a single step resolution, with its full content (output included)
func GetResolutionStep(c *gin.Context, in *getResolutionStepIn) (*step.Step, error) {
	metadata.AddActionMetadata(c, metadata.ResolutionID, in.PublicID)

	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return nil, err
	}

	r, err := resolution.LoadFromPublicID(dbp, in.PublicID)
	if err != nil {
		return nil, err
	}

	step, ok := r.Steps[in.StepName]
	if !ok {
		return nil, errors.NotFoundf("given stepName %q for this resolution", in.StepName)
	}

	metadata.AddActionMetadata(c, metadata.StepName, in.StepName)

	t, err := task.LoadFromID(dbp, r.TaskID)
	if err != nil {
		return nil, err
	}

	metadata.AddActionMetadata(c, metadata.TaskID, t.PublicID)

	tt, err := tasktemplate.LoadFromID(dbp, t.TemplateID)
	if err != nil {
		return nil, err
	}

	metadata.AddActionMetadata(c, metadata.TemplateName, tt.Name)

	admin := auth.IsAdmin(c) == nil
	requester := auth.IsRequester(c, t) == nil
	watcher := auth.IsWatcher(c, t) == nil
	resolutionManager := auth.IsResolutionManager(c, tt, t, r) == nil

	if !admin && !requester && !watcher && !resolutionManager {
		return nil, errors.Forbiddenf("Can't display resolution details")
	}

	if !resolutionManager && !admin {
		r.ClearOutputs()
	}

	if !resolutionManager && !requester && !watcher {
		metadata.SetSUDO(c)
	}

	return step, nil
}

type updateResolutionStepIn struct {
	step.Step
	PublicID string `path:"id" validate:"required"`
	StepName string `path:"stepName" validate:"required"`
}

// UpdateResolutionStep is a special handler reserved to administrators, which allows the
// edition of a live resolution. It's equivalent to UpdateResolution, but focus on a singular step
// instead of live patch the whole resolution.
// can only be called when resolution is in state PAUSED
func UpdateResolutionStep(c *gin.Context, in *updateResolutionStepIn) error {
	metadata.AddActionMetadata(c, metadata.ResolutionID, in.PublicID)

	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return err
	}

	if err := dbp.Tx(); err != nil {
		return err
	}

	r, err := resolution.LoadLockedNoWaitFromPublicID(dbp, in.PublicID)
	if err != nil {
		dbp.Rollback()
		return err
	}

	if _, ok := r.Steps[in.StepName]; !ok {
		dbp.Rollback()
		return errors.NotFoundf("given stepName %q for this resolution", in.StepName)
	}

	metadata.AddActionMetadata(c, metadata.StepName, in.StepName)

	t, err := task.LoadFromID(dbp, r.TaskID)
	if err != nil {
		dbp.Rollback()
		return err
	}

	metadata.AddActionMetadata(c, metadata.TaskID, t.PublicID)

	if r.State != resolution.StatePaused {
		dbp.Rollback()
		return errors.BadRequestf("Cannot update a resolution which is not in state '%s'", resolution.StatePaused)
	}

	if err := auth.IsAdmin(c); err != nil {
		dbp.Rollback()
		return err
	}

	metadata.SetSUDO(c)

	tt, err := tasktemplate.LoadFromID(dbp, t.TemplateID)
	if err != nil {
		dbp.Rollback()
		return err
	}

	metadata.AddActionMetadata(c, metadata.TemplateName, tt.Name)

	r.Steps[in.StepName] = &in.Step

	if err := r.Steps[in.StepName].ValidAndNormalize(in.StepName, tt.BaseConfigurations, r.Steps); err != nil {
		dbp.Rollback()
		return err
	}

	valid, err := r.Steps[in.StepName].CheckIfValidState()
	if err != nil {
		_ = dbp.Rollback()
		return err
	}

	if !valid {
		_ = dbp.Rollback()
		return errors.NewBadRequest(nil, fmt.Sprintf("invalid state provided: %q is not allowed", in.State))
	}

	logrus.WithFields(logrus.Fields{"resolution_id": r.PublicID}).Debugf("Handler UpdateResolutionStep: manual update of resolution %s step %s", r.PublicID, in.StepName)

	if err := r.Update(dbp); err != nil {
		dbp.Rollback()
		return err
	}

	reqUsername := auth.GetIdentity(c)
	_, err = task.CreateComment(dbp, t, reqUsername, "manually updated resolution step "+in.StepName)
	if err != nil {
		dbp.Rollback()
		return err
	}

	if err := dbp.Commit(); err != nil {
		dbp.Rollback()
		return err
	}

	return nil
}

type updateResolutionStepStateIn struct {
	PublicID string `path:"id" validate:"required"`
	StepName string `path:"stepName" validate:"required"`
	State    string `json:"state" validate:"required"`
}

// UpdateResolutionStepState allows the edition of a step state.
// Can only be called when the resolution is in state PAUSED, and by the template owners.
func UpdateResolutionStepState(c *gin.Context, in *updateResolutionStepStateIn) error {
	metadata.AddActionMetadata(c, metadata.ResolutionID, in.PublicID)
	metadata.AddActionMetadata(c, metadata.StepName, in.StepName)

	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return err
	}

	if err := dbp.Tx(); err != nil {
		return err
	}

	r, err := resolution.LoadLockedNoWaitFromPublicID(dbp, in.PublicID)
	if err != nil {
		dbp.Rollback()
		return err
	}

	if _, ok := r.Steps[in.StepName]; !ok {
		dbp.Rollback()
		return errors.NotFoundf("given stepName %q for this resolution", in.StepName)
	}

	t, err := task.LoadFromID(dbp, r.TaskID)
	if err != nil {
		dbp.Rollback()
		return err
	}

	metadata.AddActionMetadata(c, metadata.TaskID, t.PublicID)

	if r.State != resolution.StatePaused {
		dbp.Rollback()
		return errors.BadRequestf("Cannot update a resolution which is not in state '%s'", resolution.StatePaused)
	}

	tt, err := tasktemplate.LoadFromID(dbp, t.TemplateID)
	if err != nil {
		dbp.Rollback()
		return err
	}

	metadata.AddActionMetadata(c, metadata.TemplateName, tt.Name)

	admin := auth.IsAdmin(c) == nil
	resolutionManager := auth.IsTemplateOwner(c, tt) == nil

	if !admin && !resolutionManager {
		dbp.Rollback()
		return errors.Forbiddenf("Can't update resolution step state")
	} else if !resolutionManager {
		metadata.SetSUDO(c)
	}

	s := r.Steps[in.StepName]
	oldState := s.State
	s.State = in.State

	valid, err := s.CheckIfValidState()
	if err != nil {
		dbp.Rollback()
		return err
	}

	if !valid {
		dbp.Rollback()
		return errors.NewBadRequest(nil, fmt.Sprintf("invalid state provided: %q is not allowed", in.State))
	}

	logrus.WithFields(logrus.Fields{"resolution_id": r.PublicID}).Debugf("Handler UpdateResolutionStepState: manual update of resolution %s step %s state switched from %s to %s", r.PublicID, in.StepName, oldState, in.State)
	metadata.AddActionMetadata(c, metadata.OldState, oldState)
	metadata.AddActionMetadata(c, metadata.NewState, s.State)

	if err := r.Update(dbp); err != nil {
		dbp.Rollback()
		return err
	}

	reqUsername := auth.GetIdentity(c)
	_, err = task.CreateComment(dbp, t, reqUsername, "manually updated resolution step "+in.StepName+" state from "+oldState+" to "+in.State)
	if err != nil {
		dbp.Rollback()
		return err
	}

	if err := dbp.Commit(); err != nil {
		dbp.Rollback()
		return err
	}

	return nil
}

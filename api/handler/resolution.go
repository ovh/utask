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
)

type createResolutionIn struct {
	TaskID         string                 `json:"task_id" binding:"required"`
	ResolverInputs map[string]interface{} `json:"resolver_inputs"`
}

// CreateResolution handles the creation of a resolution for a given task
// the creator of the resolution (aka "resolver") might have to provide extra inputs
// depending on the task's template definition
func CreateResolution(c *gin.Context, in *createResolutionIn) (*resolution.Resolution, error) {
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

	tt, err := tasktemplate.LoadFromID(dbp, t.TemplateID)
	if err != nil {
		dbp.Rollback()
		return nil, err
	}

	if err := auth.IsResolutionManager(c, tt, t, nil); err != nil {
		dbp.Rollback()
		return nil, errors.Forbiddenf("You are not allowed to resolve this task")
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

	r, err := resolution.Create(dbp, t, in.ResolverInputs, resUser, true, nil) // TODO accept delay in handler
	if err != nil {
		dbp.Rollback()
		return nil, err
	}

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
	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return nil, err
	}

	r, err := resolution.LoadFromPublicID(dbp, in.PublicID)
	if err != nil {
		return nil, err
	}

	if err := auth.IsResolver(c, r); err != nil {
		r.ClearOutputs()
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

	if r.State != resolution.StatePaused {
		dbp.Rollback()
		return errors.NotValidf("Cannot update a resolution which is not in state '%s'", resolution.StatePaused)
	}

	if err := auth.IsAdmin(c); err != nil {
		dbp.Rollback()
		return err
	}

	if in.Steps != nil {
		r.Steps = in.Steps
	}

	if in.ResolverInputs != nil {
		r.SetInput(in.ResolverInputs)
	}

	logrus.WithFields(logrus.Fields{"resolution_id": r.PublicID}).Debugf("Handler UpdateResolution: manual update of resolution %s", r.PublicID)

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

type runResolutionIn struct {
	PublicID string `path:"id, required"`
}

// RunResolution launches the asynchronous execution of a resolution
// the engine determines if resolution is eligible for execution
func RunResolution(c *gin.Context, in *runResolutionIn) error {
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

	tt, err := tasktemplate.LoadFromID(dbp, t.TemplateID)
	if err != nil {
		return err
	}

	if err := auth.IsResolutionManager(c, tt, t, r); err != nil {
		return errors.Forbiddenf("You are not allowed to resolve this task")
	}

	logrus.WithFields(logrus.Fields{"resolution_id": r.PublicID}).Debugf("Handler RunResolution: manual resolve %s", r.PublicID)

	return engine.GetEngine().Resolve(in.PublicID, nil)
}

type extendResolutionIn struct {
	PublicID string `path:"id, required"`
}

// ExtendResolution increments a resolution's remaining execution retries
// in case it has reached state BLOCKED_MAXRETRIES
func ExtendResolution(c *gin.Context, in *extendResolutionIn) error {
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

	tt, err := tasktemplate.LoadFromID(dbp, t.TemplateID)
	if err != nil {
		dbp.Rollback()
		return err
	}

	if err := auth.IsResolutionManager(c, tt, t, r); err != nil {
		dbp.Rollback()
		return err
	}

	if r.State != resolution.StateBlockedMaxRetries {
		dbp.Rollback()
		return errors.NotValidf("Cannot extend a resolution which is not in state '%s'", resolution.StateBlockedMaxRetries)
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

	tt, err := tasktemplate.LoadFromID(dbp, t.TemplateID)
	if err != nil {
		dbp.Rollback()
		return err
	}

	if err := auth.IsResolutionManager(c, tt, t, r); err != nil {
		dbp.Rollback()
		return err
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

	tt, err := tasktemplate.LoadFromID(dbp, t.TemplateID)
	if err != nil {
		dbp.Rollback()
		return err
	}

	if err := auth.IsResolutionManager(c, tt, t, r); err != nil {
		dbp.Rollback()
		return errors.Forbiddenf("You are not allowed to resolve this task")
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

	if err := dbp.Commit(); err != nil {
		dbp.Rollback()
		return err
	}

	return nil
}

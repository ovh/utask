package handler

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/juju/errors"
	"github.com/loopfz/gadgeto/zesty"

	"github.com/ovh/utask"
	"github.com/ovh/utask/engine/step"
	"github.com/ovh/utask/models/resolution"
	"github.com/ovh/utask/models/task"
	"github.com/ovh/utask/models/tasktemplate"
	"github.com/ovh/utask/pkg/auth"
	"github.com/ovh/utask/pkg/utils"
)

type createTaskIn struct {
	TemplateName     string                 `json:"template_name" binding:"required"`
	Input            map[string]interface{} `json:"input" binding:"required"`
	Comment          string                 `json:"comment"`
	WatcherUsernames []string               `json:"watcher_usernames"`
	Delay            *string                `json:"delay"`
}

// CreateTask handles the creation of a new task based on an existing template
// the template determines the expected input
// an initial comment on the task can be provided for context
// watchers will be able to follow the state of this task while having no right to act on it
// a delay can be set to offset this task's execution by a certain amount of time
// delay is expressed according to https://golang.org/pkg/time/#ParseDuration
// A duration string is a possibly signed sequence of decimal numbers,
// each with optional fraction and a unit suffix, such as "300ms", "-1.5h" or "2h45m".
// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
func CreateTask(c *gin.Context, in *createTaskIn) (*task.Task, error) {
	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return nil, err
	}

	tt, err := tasktemplate.LoadFromName(dbp, in.TemplateName)
	if err != nil {
		return nil, err
	}

	if err := dbp.Tx(); err != nil {
		return nil, err
	}

	t, err := transactionTaskCreate(c, dbp, tt, in.WatcherUsernames, in.Input, nil, in.Comment, in.Delay)
	if err != nil {
		dbp.Rollback()
		return nil, err
	}

	if err := dbp.Commit(); err != nil {
		dbp.Rollback()
		return nil, err
	}

	return t, nil
}

func transactionTaskCreate(c *gin.Context, dbp zesty.DBProvider, tt *tasktemplate.TaskTemplate, watcherUsernames []string, input map[string]interface{}, b *task.Batch, comment string, delay *string) (*task.Task, error) {
	reqUsername := auth.GetIdentity(c)

	if tt.Blocked {
		return nil, errors.NewNotValid(nil, "Template not available (blocked)")
	}

	t, err := task.Create(dbp, tt, reqUsername, watcherUsernames, nil, input, b)
	if err != nil {
		return nil, err
	}

	if comment != "" {
		com, err := task.CreateComment(dbp, t, reqUsername, comment)
		if err != nil {
			return nil, err
		}
		t.Comments = []*task.Comment{com}
	}

	if tt.IsAutoRunnable() {
		if err := auth.IsAllowedResolver(c, tt, t.ResolverUsernames); err == nil {
			var delayUntil *time.Time
			if delay != nil {
				delayDuration, _ := time.ParseDuration(*delay)
				delayTime := time.Now().Add(delayDuration)
				delayUntil = &delayTime
			}
			if _, err := resolution.Create(dbp, t, nil, reqUsername, true, delayUntil); err != nil {
				return nil, err
			}
		}
	}

	return t, nil
}

const (
	taskTypeOwn        = "own"
	taskTypeResolvable = "resolvable"
	taskTypeAll        = "all"
)

type listTasksIn struct {
	Type          string     `query:"type, default=own"`
	State         *string    `query:"state"`
	BatchPublicID *string    `query:"batch"`
	PageSize      uint64     `query:"page_size"`
	Last          *string    `query:"last"`
	After         *time.Time `query:"after"`
	Before        *time.Time `query:"before"`
}

// ListTasks returns a list of tasks, which can be filtered by state, batch ID,
// and last activity time (before and/or after)
// type=own (default) returns tasks for which the user is the requester
// type=resolvable returns tasks for which the user is a potential resolver
// type=all returns every task (only available to administrator users)
func ListTasks(c *gin.Context, in *listTasksIn) (t []*task.Task, err error) {
	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return nil, err
	}

	filter := task.ListFilter{
		PageSize: normalizePageSize(in.PageSize),
		Last:     in.Last,
		State:    in.State,
		After:    in.After,
		Before:   in.Before,
	}

	var b *task.Batch
	if in.BatchPublicID != nil {
		b, err = task.LoadBatchFromPublicID(dbp, *in.BatchPublicID)
		if err != nil {
			return nil, err
		}
		filter.Batch = b
	}

	reqUsername := auth.GetIdentity(c)
	var user *string
	if reqUsername != "" {
		user = &reqUsername
	}

	switch in.Type {
	case taskTypeOwn:
		filter.RequesterUser = user
	case taskTypeResolvable:
		// all tasks are resolvable by an admin
		filter.State = utils.StrPtr(task.StateTODO)
		if auth.IsGlobalResolverUser(reqUsername) {
			user = nil
		}
		if err2 := auth.IsAdmin(c); err2 != nil {
			filter.PotentialResolverUser = user
		}
	case taskTypeAll:
		if err2 := auth.IsAdmin(c); err2 != nil {
			return nil, err2
		}
	default:
		return nil, errors.BadRequestf("Unknown type for listing: '%s'. Was expecting '%s', '%s' or '%s'", in.Type, taskTypeOwn, taskTypeResolvable, taskTypeAll)
	}

	t, err = task.ListTasks(dbp, filter)
	if err != nil {
		return nil, err
	}

	if uint64(len(t)) == filter.PageSize {
		lastT := t[len(t)-1].PublicID
		c.Header(
			linkHeader,
			buildTaskNextLink(in.Type, in.State, in.BatchPublicID, filter.PageSize, lastT),
		)
	}

	c.Header(pageSizeHeader, fmt.Sprintf("%v", filter.PageSize))

	return t, nil
}

type getTaskIn struct {
	PublicID string `path:"id, required"`
}

// GetTask returns a single task
// inputs of type password are obfuscated to every user except administrators
func GetTask(c *gin.Context, in *getTaskIn) (*task.Task, error) {
	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return nil, err
	}

	t, err := task.LoadFromPublicID(dbp, in.PublicID)
	if err != nil {
		return nil, err
	}

	tt, err := tasktemplate.LoadFromID(dbp, t.TemplateID)
	if err != nil {
		return nil, err
	}

	admin := auth.IsAdmin(c) == nil
	requester := isRequester(t, c) == nil
	watcher := isWatcher(t, c) == nil
	resolver := isResolver(t, c) == nil

	if !admin && !requester && !watcher && !resolver {
		return nil, errors.Forbiddenf("Can't display task details")
	}
	if !admin {
		t.Input = obfuscateInput(tt.Inputs, t.Input)
	}

	if t.State == task.StateBlocked && t.Resolution != nil {
		res, err := resolution.LoadFromPublicID(dbp, *t.Resolution)
		if err != nil {
			return nil, err
		}
		for _, s := range res.Steps {
			if s.State == step.StateClientError {
				t.Errors = append(t.Errors, task.StepError{
					Step:  s.Description,
					Error: s.Error,
				})
			}
		}
	}

	return t, nil
}

type updateTaskIn struct {
	PublicID         string                 `path:"id, required"`
	Input            map[string]interface{} `json:"input"`
	WatcherUsernames []string               `json:"watcher_usernames"`
}

// UpdateTask modifies a task, allowing it's requester or an administrator
// to fix a broken input, or to add/remove watchers
func UpdateTask(c *gin.Context, in *updateTaskIn) (*task.Task, error) {
	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return nil, err
	}

	t, err := task.LoadFromPublicID(dbp, in.PublicID)
	if err != nil {
		return nil, err
	}

	if err := auth.IsAdmin(c); err != nil {
		if err := isRequester(t, c); err != nil {
			return nil, err
		}
	}

	// avoid secrets being squashed by their obfuscated placeholder
	clearInput := deobfuscateNewInput(t.Input, in.Input)

	t.SetInput(clearInput)
	t.SetWatcherUsernames(in.WatcherUsernames)

	if err := t.Update(dbp,
		false, // do validate task contents
		true,  // change last activity value, bring task bask to top of the list
	); err != nil {
		return nil, err
	}

	return t, nil
}

type deleteTaskIn struct {
	PublicID string `path:"id, required"`
}

// DeleteTask removes a task from the data backend
func DeleteTask(c *gin.Context, in *deleteTaskIn) error {
	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return err
	}

	t, err := task.LoadFromPublicID(dbp, in.PublicID)
	if err != nil {
		return err
	}

	switch t.State {
	case task.StateRunning, task.StateBlocked:
		return errors.BadRequestf("Task can't be deleted while in state %q", t.State)
	}

	return t.Delete(dbp)
}

type wontfixTaskIn struct {
	PublicID string `path:"id, required"`
}

// WontfixTask changes a task's state to prevent it from ever being resolved
func WontfixTask(c *gin.Context, in *wontfixTaskIn) error {
	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return err
	}

	t, err := task.LoadFromPublicID(dbp, in.PublicID)
	if err != nil {
		return err
	}

	if t.State != task.StateTODO {
		return errors.BadRequestf("Can't set task's state to %s: task is in state %s", task.StateWontfix, t.State)
	}

	tt, err := tasktemplate.LoadFromID(dbp, t.TemplateID)
	if err != nil {
		return err
	}

	resolver := auth.IsAllowedResolver(c, tt, t.ResolverUsernames) == nil

	if !resolver {
		return errors.Forbiddenf("Can't set task's state to %s", task.StateWontfix)
	}

	t.SetState(task.StateWontfix)

	return t.Update(dbp,
		false, // skip validation of task contents, task is dead anyway
		true,  // do record mark change with last activity timestamp
	)
}

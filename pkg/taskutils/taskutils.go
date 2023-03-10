package taskutils

import (
	"context"
	"time"

	"github.com/juju/errors"
	"github.com/loopfz/gadgeto/zesty"

	"github.com/ovh/utask/models/resolution"
	"github.com/ovh/utask/models/task"
	"github.com/ovh/utask/models/tasktemplate"
	"github.com/ovh/utask/pkg/auth"
	"github.com/ovh/utask/pkg/constants"
)

// CreateTask creates a task with the given inputs, and creates a resolution if autorunnable
func CreateTask(c context.Context, dbp zesty.DBProvider, tt *tasktemplate.TaskTemplate, watcherUsernames []string, watcherGroups []string, resolverUsernames []string, resolverGroups []string, input map[string]interface{}, b *task.Batch, comment string, delay *string, tags map[string]string) (*task.Task, error) {
	reqUsername := auth.GetIdentity(c)
	reqGroups := auth.GetGroups(c)

	if tt.Blocked {
		return nil, errors.NewNotValid(nil, "Template not available (blocked)")
	}
	delayed := delay != nil
	t, err := task.Create(dbp, tt, reqUsername, reqGroups, watcherUsernames, watcherGroups, resolverUsernames, resolverGroups, input, tags, b, delayed)
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

	if !tt.IsAutoRunnable() && tt.AllowAllResolverUsernames {
		return nil, errors.Errorf("invalid tasktemplate: %q should be auto_runnable", tt.Name)
	} else if !tt.IsAutoRunnable() {
		t.NotifyValidationRequired(tt)
		return t, nil
	}

	// task is AutoRunnable, creating resolution
	admin := auth.IsAdmin(c) == nil
	requester := (auth.IsRequester(c, t) == nil && tt.AllowAllResolverUsernames)
	resolutionManager := auth.IsResolutionManager(c, tt, t, nil) == nil

	if !requester && !resolutionManager && !admin {
		t.NotifyValidationRequired(tt)
		return t, nil
	}

	var delayUntil *time.Time
	if delay != nil {
		delayDuration, err := time.ParseDuration(*delay)
		if err != nil {
			return nil, errors.NewNotValid(err, "delay")
		}
		delayTime := time.Now().Add(delayDuration)
		delayUntil = &delayTime
	}
	if _, err := resolution.Create(dbp, t, nil, reqUsername, true, delayUntil); err != nil {
		return nil, err
	}

	return t, nil
}

func ShouldResumeParentTask(dbp zesty.DBProvider, t *task.Task) (*task.Task, error) {
	switch t.State {
	case task.StateDone, task.StateWontfix, task.StateCancelled:
	default:
		return nil, nil
	}
	if t.Tags == nil {
		return nil, nil
	}
	parentTaskID, ok := t.Tags[constants.SubtaskTagParentTaskID]
	if !ok {
		return nil, nil
	}

	parentTask, err := task.LoadFromPublicID(dbp, parentTaskID)
	if err != nil {
		return nil, err
	}
	switch parentTask.State {
	case task.StateBlocked, task.StateRunning, task.StateWaiting:
	default:
		// not allowed to resume a parent task that is not either Waiting, Running or Blocked.
		// Todo state should not be runned as it might need manual resolution from a granted resolver
		return nil, nil
	}
	if parentTask.Resolution == nil {
		return nil, nil
	}

	r, err := resolution.LoadFromPublicID(dbp, *parentTask.Resolution)
	if err != nil {
		return nil, err
	}

	switch r.State {
	case resolution.StateCrashed, resolution.StatePaused:
		return nil, nil
	}

	return parentTask, nil
}

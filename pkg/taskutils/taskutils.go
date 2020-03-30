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
)

// CreateTask creates a task with the given inputs, and creates a resolution if autorunnable
func CreateTask(c context.Context, dbp zesty.DBProvider, tt *tasktemplate.TaskTemplate, watcherUsernames []string, resolverUsernames []string, input map[string]interface{}, b *task.Batch, comment string, delay *string, tags map[string]string) (*task.Task, error) {
	reqUsername := auth.GetIdentity(c)

	if tt.Blocked {
		return nil, errors.NewNotValid(nil, "Template not available (blocked)")
	}

	t, err := task.Create(dbp, tt, reqUsername, watcherUsernames, resolverUsernames, input, tags, b)
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
		return t, nil
	}

	// task is AutoRunnable, creating resolution
	requester := (auth.IsRequester(c, t) == nil && tt.AllowAllResolverUsernames)
	resolutionManager := auth.IsResolutionManager(c, tt, t, nil) == nil

	if !requester && !resolutionManager {
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

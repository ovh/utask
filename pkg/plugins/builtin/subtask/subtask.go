package pluginsubtask

import (
	"context"
	"fmt"

	"github.com/juju/errors"
	"github.com/loopfz/gadgeto/zesty"
	"github.com/ovh/utask"
	"github.com/ovh/utask/models/resolution"
	"github.com/ovh/utask/models/task"
	"github.com/ovh/utask/models/tasktemplate"
	"github.com/ovh/utask/pkg/auth"
	"github.com/ovh/utask/pkg/plugins/taskplugin"
	"github.com/ovh/utask/pkg/utils"
)

// the subtask plugin spawns a new µTask task, given a template and inputs
// an extra parameter is accepted, not available on API
// resolver usernames can be dynamically set for the task
var (
	Plugin = taskplugin.New("subtask", "0.1", exec,
		taskplugin.WithConfig(validConfig, SubtaskConfig{}),
		taskplugin.WithContextFunc(ctx),
	)
)

// SubtaskConfig is the necessary configuration to spawn a new task
type SubtaskConfig struct {
	Template          string                 `json:"template"`
	Input             map[string]interface{} `json:"input"`
	ResolverUsernames string                 `json:"resolver_usernames"`
}

// SubtaskContext is the metadata inherited from the "parent" task"
type SubtaskContext struct {
	TaskID            string `json:"task_id"`
	RequesterUsername string `json:"requester_username"`
}

func ctx(stepName string) interface{} {
	return &SubtaskContext{
		TaskID:            fmt.Sprintf(`{{ if (index .step "%s" "output") }}{{ index .step "%s" "output" "id" }}{{ end }}`, stepName, stepName),
		RequesterUsername: "{{.task.requester_username}}",
	}
}

func validConfig(config interface{}) error {
	cfg := config.(*SubtaskConfig)

	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return err
	}

	_, err = tasktemplate.LoadFromName(dbp, cfg.Template)
	if err != nil {
		return err
	}

	return nil
}

func exec(stepName string, config interface{}, ctx interface{}) (interface{}, interface{}, error) {
	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return nil, nil, err
	}

	cfg := config.(*SubtaskConfig)
	stepContext := ctx.(*SubtaskContext)

	var t *task.Task
	if stepContext.TaskID != "" {
		// subtask was already launched, retrieve its current state and exit
		t, err = task.LoadFromPublicID(dbp, stepContext.TaskID)
		if err != nil {
			return nil, nil, err
		}
	} else {
		// spawn new subtask

		tt, err := tasktemplate.LoadFromName(dbp, cfg.Template)
		if err != nil {
			return nil, nil, err
		}

		if err := dbp.Tx(); err != nil {
			return nil, nil, err
		}

		var resolverUsernames []string
		if cfg.ResolverUsernames != "" {
			resolverUsernames, err = utils.ConvertJSONRowToSlice(cfg.ResolverUsernames)
			if err != nil {
				return nil, nil, err
			}
		}

		// TODO inherit watchers from parent task
		t, err = task.Create(dbp, tt, stepContext.RequesterUsername, nil, resolverUsernames, cfg.Input, nil)
		if err != nil {
			dbp.Rollback()
			return nil, nil, err
		}

		com, err := task.CreateComment(dbp, t, stepContext.RequesterUsername, "Auto created subtask")
		if err != nil {
			dbp.Rollback()
			return nil, nil, err
		}

		t.Comments = []*task.Comment{com}

		if tt.IsAutoRunnable() {
			ctx := auth.WithIdentity(context.Background(), stepContext.RequesterUsername)
			if err := auth.IsAllowedResolver(ctx, tt, resolverUsernames); err == nil {
				if _, err := resolution.Create(dbp, t, nil, stepContext.RequesterUsername, true, nil); err != nil {
					dbp.Rollback()
					return nil, nil, err
				}
			}
		}

		if err := dbp.Commit(); err != nil {
			dbp.Rollback()
			return nil, nil, err
		}
	}

	var stepError error
	switch t.State {
	case task.StateDone:
		stepError = nil
	case task.StateCancelled:
		// stop retrying if subtask was cancelled
		stepError = errors.BadRequestf("Task '%s' was cancelled", t.PublicID)
	default:
		// keep step running while subtask is not done
		// FIXME, use proper error type
		stepError = fmt.Errorf("Task '%s' not done yet", t.PublicID)
	}
	return map[string]interface{}{
		"id":     t.PublicID,
		"state":  t.State,
		"result": t.Result,
	}, nil, stepError
}

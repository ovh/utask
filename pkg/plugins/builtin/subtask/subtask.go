package pluginsubtask

import (
	"context"
	"fmt"
	"strings"

	"github.com/juju/errors"
	"github.com/loopfz/gadgeto/zesty"

	"github.com/ovh/utask"
	"github.com/ovh/utask/models/task"
	"github.com/ovh/utask/models/tasktemplate"
	"github.com/ovh/utask/pkg/auth"
	"github.com/ovh/utask/pkg/constants"
	"github.com/ovh/utask/pkg/plugins/taskplugin"
	"github.com/ovh/utask/pkg/taskutils"
	"github.com/ovh/utask/pkg/templateimport"
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
	ResolverGroups    string                 `json:"resolver_groups"`
	WatcherUsernames  string                 `json:"watcher_usernames"`
	WatcherGroups     string                 `json:"watcher_groups"`
	Delay             *string                `json:"delay"`
	Tags              map[string]string      `json:"tags"`
}

// SubtaskContext is the metadata inherited from the "parent" task"
type SubtaskContext struct {
	ParentTaskID      string `json:"parent_task_id"`
	TaskID            string `json:"task_id"`
	RequesterUsername string `json:"requester_username"`
	RequesterGroups   string `json:"requester_groups"`
}

func ctx(stepName string) interface{} {
	return &SubtaskContext{
		ParentTaskID:      "{{ .task.task_id }}",
		TaskID:            fmt.Sprintf("{{ if (index .step `%s` ) }}{{ if (index .step `%s` `output`) }}{{ index .step `%s` `output` `id` }}{{ end }}{{ end }}", stepName, stepName, stepName),
		RequesterUsername: "{{.task.requester_username}}",
		RequesterGroups:   "{{ if .task.requester_groups }}{{ .task.requester_groups }}{{ end }}",
	}
}

func validConfig(config interface{}) error {
	cfg := config.(*SubtaskConfig)

	if err := utils.ValidateTags(cfg.Tags); err != nil {
		return err
	}

	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return fmt.Errorf("can't retrieve connexion to DB: %s", err)
	}

	_, err = tasktemplate.LoadFromName(dbp, cfg.Template)
	if err == nil {
		return nil
	}
	if !errors.IsNotFound(err) {
		return fmt.Errorf("can't load template from name: %s", err)
	}

	// searching into currently imported templates
	templates := templateimport.GetTemplates()
	for _, template := range templates {
		if template == cfg.Template {
			return nil
		}
	}

	return errors.NotFoundf("sub-task template %q", cfg.Template)
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

		var requesterGroups, resolverUsernames, resolverGroups, watcherUsernames, watcherGroups []string
		if stepContext.RequesterGroups != "" {
			requesterGroups = strings.Split(stepContext.RequesterGroups, utask.GroupsSeparator)
		}
		if cfg.ResolverUsernames != "" {
			resolverUsernames, err = utils.ConvertJSONRowToSlice(cfg.ResolverUsernames)
			if err != nil {
				return nil, nil, fmt.Errorf("can't convert JSON to row slice: %s", err)
			}
		}
		if cfg.ResolverGroups != "" {
			resolverGroups, err = utils.ConvertJSONRowToSlice(cfg.ResolverGroups)
			if err != nil {
				return nil, nil, fmt.Errorf("can't convert JSON to row slice: %s", err)
			}
		}
		if cfg.WatcherUsernames != "" {
			watcherUsernames, err = utils.ConvertJSONRowToSlice(cfg.WatcherUsernames)
			if err != nil {
				return nil, nil, fmt.Errorf("can't convert JSON to row slice: %s", err)
			}
		}
		if cfg.WatcherGroups != "" {
			watcherGroups, err = utils.ConvertJSONRowToSlice(cfg.WatcherGroups)
			if err != nil {
				return nil, nil, fmt.Errorf("can't convert JSON to row slice: %s", err)
			}
		}

		// TODO inherit watchers from parent task
		ctx := auth.WithIdentity(context.Background(), stepContext.RequesterUsername)
		ctx = auth.WithGroups(context.Background(), requesterGroups)
		if cfg.Tags == nil {
			cfg.Tags = map[string]string{}
		}
		cfg.Tags[constants.SubtaskTagParentTaskID] = stepContext.ParentTaskID
		t, err = taskutils.CreateTask(ctx, dbp, tt, watcherUsernames, watcherGroups, resolverUsernames, resolverGroups, cfg.Input, nil, "Auto created subtask, parent task "+stepContext.ParentTaskID, cfg.Delay, cfg.Tags)
		if err != nil {
			dbp.Rollback()
			return nil, nil, err
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
	case task.StateCancelled, task.StateWontfix, task.StateBlocked:
		// Stop retrying the subtask.
		stepError = errors.BadRequestf("Task '%s' changed state: %s", t.PublicID, strings.ToLower(t.State))
	case task.StateTODO:
		if t.Resolution == nil {
			stepError = errors.NewNotAssigned(fmt.Errorf("Task %q is waiting for human validation", t.PublicID), "")
		} else {
			stepError = errors.NewNotAssigned(fmt.Errorf("Task %q will start shortly", t.PublicID), "")
		}
	case task.StateRunning:
		stepError = errors.NewNotProvisioned(fmt.Errorf("Task %q is currently RUNNING", t.PublicID), "")
	default:
		// keep step running while subtask is not done
		// FIXME, use proper error type
		stepError = fmt.Errorf("Task %q not done yet (current state is %s)", t.PublicID, t.State)
	}
	return map[string]interface{}{
		"id":                 t.PublicID,
		"state":              t.State,
		"result":             t.Result,
		"resolver_username":  t.ResolverUsername,
		"requester_username": t.RequesterUsername,
	}, nil, stepError
}

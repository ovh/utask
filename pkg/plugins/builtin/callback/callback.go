package plugincallback

import (
	"fmt"
	"strings"

	"github.com/juju/errors"
	"github.com/loopfz/gadgeto/zesty"
	"github.com/ovh/utask"
	"github.com/ovh/utask/models/task"
	"github.com/ovh/utask/pkg/plugins/taskplugin"
)

var (
	Plugin = taskplugin.New("callback", "0.1", exec,
		taskplugin.WithConfig(validConfig, CallbackStepConfig{}),
		taskplugin.WithContextFunc(ctx),
	)
)

type CallbackStepConfig struct {
	Action     string `json:"action"`
	BodySchema string `json:"schema,omitempty"`
	ID         string `json:"id"`
}

type CallbackContext struct {
	StepName          string `json:"step"`
	TaskID            string `json:"task_id"`
	RequesterUsername string `json:"requester_username"`
}

func ctx(stepName string) interface{} {
	return &CallbackContext{
		TaskID:            "{{.task.task_id}}",
		RequesterUsername: "{{.task.requester_username}}",
		StepName:          stepName,
	}
}

func validConfig(config interface{}) error {
	cfg := config.(*CallbackStepConfig)

	switch strings.ToLower(cfg.Action) {
	case "create":
		return nil

	case "wait":
		if cfg.ID == "" {
			return fmt.Errorf("missing %q parameter", "id")
		}
		return nil

	default:
		return fmt.Errorf("invalid action %q", cfg.Action)
	}
}

func exec(stepName string, config interface{}, ctx interface{}) (interface{}, interface{}, error) {
	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return nil, nil, err
	}

	cfg := config.(*CallbackStepConfig)
	stepContext := ctx.(*CallbackContext)

	task, err := task.LoadFromPublicID(dbp, stepContext.TaskID)
	if err != nil {
		return nil, nil, err
	}

	switch strings.ToLower(cfg.Action) {
	case "create":
		cb, err := createCallback(dbp, task, stepContext, cfg.BodySchema)
		if err != nil {
			return nil, nil, err
		}

		return map[string]interface{}{
			"id":     cb.PublicID,
			"url":    buildUrl(cb),
			"schema": cb.Schema,
		}, nil, nil

	case "wait":
		cb, err := loadFromPublicID(dbp, cfg.ID, false)
		if err != nil {
			return nil, nil, err
		}

		if cb.Called == nil {
			return nil, nil, errors.NewNotAssigned(fmt.Errorf("task is waiting for a callback"), "")
		}

		return map[string]interface{}{
			"id":   cb.PublicID,
			"date": cb.Called,
			"body": cb.Body,
		}, nil, nil

	default:
		return nil, nil, errors.BadRequestf("invalid action %q", cfg.Action)
	}
}

func buildUrl(cb *callback) string {
	if Init.cfg.PathPrefix == "" {
		return fmt.Sprintf("%s%s/%s?t=%s", Init.cfg.BaseURL, defaultCallbackPathPrefix, cb.PublicID, cb.Secret)
	} else {
		basePath := Init.cfg.PathPrefix
		if !strings.HasPrefix(basePath, "/") {
			basePath = fmt.Sprintf("/%s", basePath)
		}
		basePath = strings.TrimSuffix(basePath, "/")
		return fmt.Sprintf("%s%s/%s?t=%s", Init.cfg.BaseURL, basePath, cb.PublicID, cb.Secret)
	}
}

package task_test

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/loopfz/gadgeto/zesty"
	"github.com/stretchr/testify/assert"

	"github.com/ovh/configstore"
	"github.com/ovh/utask"
	"github.com/ovh/utask/api"
	"github.com/ovh/utask/db"
	"github.com/ovh/utask/engine"
	functionrunner "github.com/ovh/utask/engine/functions/runner"
	"github.com/ovh/utask/engine/step"
	"github.com/ovh/utask/models/task"
	"github.com/ovh/utask/models/tasktemplate"
	compress "github.com/ovh/utask/pkg/compress/init"
	"github.com/ovh/utask/pkg/now"
	"github.com/ovh/utask/pkg/plugins"
	plugincallback "github.com/ovh/utask/pkg/plugins/builtin/callback"
	"github.com/ovh/utask/pkg/plugins/builtin/echo"
	"github.com/ovh/utask/pkg/plugins/builtin/script"
	pluginsubtask "github.com/ovh/utask/pkg/plugins/builtin/subtask"
)

func createTemplates(dbp zesty.DBProvider, prefix string, templates map[string][]string) (map[string]*tasktemplate.TaskTemplate, error) {
	result := make(map[string]*tasktemplate.TaskTemplate)

	for name, groups := range templates {
		tt, err := tasktemplate.Create(dbp, prefix+name, name+" description", nil, nil, nil, nil, groups, nil, false, false, nil, nil, nil, nil, name+" title", nil, false, nil)
		if err != nil {
			return nil, err
		}
		result[name] = tt
	}

	return result, nil
}

func createTasks(dbp zesty.DBProvider, templates map[string]*tasktemplate.TaskTemplate, tasks map[string][]string) (map[string]*task.Task, error) {
	result := make(map[string]*task.Task)

	for name, groups := range tasks {
		template, ok := templates[name]
		if !ok {
			return nil, fmt.Errorf("template %q not found", name)
		}

		task, err := task.Create(dbp, template, "foo", nil, nil, nil, nil, groups, nil, nil, nil, false)
		if err != nil {
			return nil, err
		}

		result[name] = task
	}

	return result, nil
}

func TestMain(m *testing.M) {
	store := configstore.DefaultStore
	store.InitFromEnvironment()

	server := api.NewServer()
	service := &plugins.Service{Store: store, Server: server}

	if err := plugincallback.Init.Init(service); err != nil {
		panic(err)
	}

	if err := db.Init(store); err != nil {
		panic(err)
	}

	if err := now.Init(); err != nil {
		panic(err)
	}

	if err := compress.Register(); err != nil {
		panic(err)
	}

	var wg sync.WaitGroup

	if err := engine.Init(context.Background(), &wg, store); err != nil {
		panic(err)
	}

	if err := functionrunner.Init(); err != nil {
		panic(err)
	}

	step.RegisterRunner(echo.Plugin.PluginName(), echo.Plugin)
	step.RegisterRunner(script.Plugin.PluginName(), script.Plugin)
	step.RegisterRunner(pluginsubtask.Plugin.PluginName(), pluginsubtask.Plugin)
	step.RegisterRunner(plugincallback.Plugin.PluginName(), plugincallback.Plugin)

	os.Exit(m.Run())
}

func TestLoadStateCountResolverGroup(t *testing.T) {
	dbp, err := zesty.NewDBProvider(utask.DBName)
	assert.NoError(t, err)

	err = task.DeleteAllTasks(dbp)
	assert.NoError(t, err)

	tests := []struct {
		name      string
		tasks     map[string][]string
		templates map[string][]string
		wantSc    map[string]map[string]map[string]float64
	}{
		{
			"no-group",
			map[string][]string{"task": nil},
			map[string][]string{"task": nil},
			map[string]map[string]map[string]float64{
				"": {
					"task": {
						task.StateTODO:      1,
						task.StateBlocked:   0,
						task.StateRunning:   0,
						task.StateWontfix:   0,
						task.StateDone:      0,
						task.StateCancelled: 0,
					},
				},
			},
		},
		{
			"no-override",
			map[string][]string{"task": nil},
			map[string][]string{"task": {"foo"}},
			map[string]map[string]map[string]float64{
				"foo": {
					"task": {
						task.StateTODO:      1,
						task.StateBlocked:   0,
						task.StateRunning:   0,
						task.StateWontfix:   0,
						task.StateDone:      0,
						task.StateCancelled: 0,
					},
				},
			},
		},
		{
			"with-override",
			map[string][]string{"task": {"bar"}},
			map[string][]string{"task": {"foo"}},
			map[string]map[string]map[string]float64{
				"bar": {
					"task": {
						task.StateTODO:      1,
						task.StateBlocked:   0,
						task.StateRunning:   0,
						task.StateWontfix:   0,
						task.StateDone:      0,
						task.StateCancelled: 0,
					},
				},
			},
		},
		{
			"no-override-multiple",
			map[string][]string{"task": nil},
			map[string][]string{"task": {"foo", "bar"}},
			map[string]map[string]map[string]float64{
				"foo": {
					"task": {
						task.StateTODO:      1,
						task.StateBlocked:   0,
						task.StateRunning:   0,
						task.StateWontfix:   0,
						task.StateDone:      0,
						task.StateCancelled: 0,
					},
				},
				"bar": {
					"task": {
						task.StateTODO:      1,
						task.StateBlocked:   0,
						task.StateRunning:   0,
						task.StateWontfix:   0,
						task.StateDone:      0,
						task.StateCancelled: 0,
					},
				},
			},
		},
		{
			"with-override-multiple",
			map[string][]string{"task": {"foo", "bar"}},
			map[string][]string{"task": {"dummy"}},
			map[string]map[string]map[string]float64{
				"foo": {
					"task": {
						task.StateTODO:      1,
						task.StateBlocked:   0,
						task.StateRunning:   0,
						task.StateWontfix:   0,
						task.StateDone:      0,
						task.StateCancelled: 0,
					},
				},
				"bar": {
					"task": {
						task.StateTODO:      1,
						task.StateBlocked:   0,
						task.StateRunning:   0,
						task.StateWontfix:   0,
						task.StateDone:      0,
						task.StateCancelled: 0,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := dbp.Tx(); err != nil {
				t.Errorf("Tx() error = %v", err)
				return
			}

			prefix := fmt.Sprintf("task-%d-", time.Now().UnixNano())

			prefixedWantSc := make(map[string]map[string]map[string]float64)

			for group, groupStats := range tt.wantSc {
				prefixedWantSc[group] = map[string]map[string]float64{}

				for template, templateStats := range groupStats {
					prefixedWantSc[group][prefix+template] = map[string]float64{}

					for state, count := range templateStats {
						prefixedWantSc[group][prefix+template][state] = count
					}
				}
			}

			templates, err := createTemplates(dbp, prefix, tt.templates)
			if err != nil {
				t.Errorf("createTemplates() error = %v", err)
				return
			}

			_, err = createTasks(dbp, templates, tt.tasks)
			if err != nil {
				t.Errorf("createTasks() error = %v", err)
				return
			}

			gotSc, err := task.LoadStateCountResolverGroup(dbp)
			if err != nil {
				t.Errorf("LoadStateCountResolverGroup() error = %v, wantErr false", err)
				return
			}

			if !reflect.DeepEqual(gotSc, prefixedWantSc) {
				t.Errorf("LoadStateCountResolverGroup() = %v, want %v", gotSc, prefixedWantSc)
			}

			if err := dbp.Rollback(); err != nil {
				t.Errorf("Rollback() error = %v", err)
				return
			}
		})
	}
}

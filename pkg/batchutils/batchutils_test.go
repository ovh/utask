package batchutils_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/juju/errors"
	"github.com/loopfz/gadgeto/zesty"
	"github.com/ovh/configstore"
	"github.com/stretchr/testify/assert"

	"github.com/ovh/utask"
	"github.com/ovh/utask/db"
	"github.com/ovh/utask/engine/input"
	"github.com/ovh/utask/engine/step"
	"github.com/ovh/utask/engine/step/executor"
	"github.com/ovh/utask/models/task"
	"github.com/ovh/utask/models/tasktemplate"
	"github.com/ovh/utask/pkg/batchutils"
)

func TestRunningTasks(t *testing.T) {
	store := configstore.DefaultStore
	store.InitFromEnvironment()

	if err := db.Init(store); err != nil {
		panic(err)
	}

	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		t.Fatal(err)
	}

	const batchSize int = 10
	batchID, tasks := createBatch(t, batchSize, dbp)

	// Making sure that created tasks running
	running, err := batchutils.RunningTasks(dbp, batchID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, int64(len(tasks)), running)

	// Setting a final state to some tasks in the batch (one per final state)
	for i, state := range batchutils.FinalStates {
		tasks[i].SetState(state)
		if err := tasks[i].Update(dbp, false, false); err != nil {
			t.Fatal(err)
		}
	}

	// Making sure that tasks in final states aren't counted
	running, err = batchutils.RunningTasks(dbp, batchID)
	if err != nil {
		t.Fatal(err)
	}
	expectedRunning := int64(len(tasks) - len(batchutils.FinalStates))
	assert.Equal(t, expectedRunning, running)
}

func createBatch(t *testing.T, amount int, dbp zesty.DBProvider) (int64, []*task.Task) {
	tmpl, err := tasktemplate.LoadFromName(dbp, dummyTemplate.Name)
	if err != nil {
		if !errors.IsNotFound(err) {
			t.Fatal(err)
		}
		tmpl = &dummyTemplate
		if err := dbp.DB().Insert(tmpl); err != nil {
			t.Fatal(err)
		}
	}

	b, err := task.CreateBatch(dbp)
	if err != nil {
		t.Fatal(err)
	}

	tasks := make([]*task.Task, 0, amount)
	for i := 0; i < amount; i++ {
		// Manually populating the batch to prevent cyclic imports
		newTask, err := task.Create(
			dbp,
			tmpl,
			"",
			nil,
			nil,
			nil,
			nil,
			nil,
			map[string]any{"id": fmt.Sprintf("dummyID-%d", i)},
			nil,
			b,
			false,
		)
		if err != nil {
			t.Fatal(err)
		}
		tasks = append(tasks, newTask)
	}

	return b.ID, tasks
}

var dummyTemplate = tasktemplate.TaskTemplate{
	Name:        "dummy-template",
	Description: "does nothing",
	TitleFormat: "this task does nothing at all",
	Inputs: []input.Input{
		{
			Name: "id",
		},
	},
	Steps: map[string]*step.Step{
		"step": {
			Action: executor.Executor{
				Type: "echo",
				Configuration: json.RawMessage(`{
					"output": {"foo":"bar"}
				}`),
			},
		},
	},
}

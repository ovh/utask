package batch

import (
	"context"
	"encoding/json"
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
)

func TestPopulate(t *testing.T) {
	store := configstore.DefaultStore
	store.InitFromEnvironment()

	if err := db.Init(store); err != nil {
		panic(err)
	}

	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		t.Fatal(err)
	}

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

	batchArgs := TaskArgs{
		TemplateName: tmpl.Name,
		Inputs:       []map[string]any{{"id": "dummyID-1"}, {"id": "dummyID-2"}, {"id": "dummyID-3"}},
	}

	taskIDs, err := Populate(context.Background(), b, dbp, batchArgs)
	if err != nil {
		t.Fatal(err)
	}

	// Making sure we returned as many IDs as tasks we created
	assert.Len(t, taskIDs, len(batchArgs.Inputs))

	tasks, err := task.ListTasks(dbp, task.ListFilter{Batch: b})
	if err != nil {
		t.Fatal(err)
	}

	// Making sure the right number of tasks was created in the batch
	assert.Len(t, taskIDs, len(batchArgs.Inputs))

	for i, childTask := range tasks {
		assert.Equal(t, batchArgs.Inputs[i]["id"], childTask.Title)
	}

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

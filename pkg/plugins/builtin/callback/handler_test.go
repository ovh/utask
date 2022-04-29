package plugincallback

import (
	"context"
	"fmt"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/loopfz/gadgeto/zesty"
	"github.com/ovh/configstore"
	"github.com/ovh/utask"
	"github.com/ovh/utask/api"
	"github.com/ovh/utask/db"
	"github.com/ovh/utask/engine"
	"github.com/ovh/utask/engine/input"
	"github.com/ovh/utask/engine/step"
	"github.com/ovh/utask/engine/values"
	"github.com/ovh/utask/models/resolution"
	"github.com/ovh/utask/models/task"
	"github.com/ovh/utask/models/tasktemplate"
	"github.com/ovh/utask/pkg/now"
	"github.com/ovh/utask/pkg/plugins"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	store := configstore.DefaultStore
	store.InitFromEnvironment()

	server := api.NewServer()
	service := &plugins.Service{Store: store, Server: server}

	if err := Init.Init(service); err != nil {
		panic(err)
	}

	if err := db.Init(store); err != nil {
		panic(err)
	}

	if err := now.Init(); err != nil {
		panic(err)
	}

	var wg sync.WaitGroup

	if err := engine.Init(context.Background(), &wg, store); err != nil {
		panic(err)
	}

	step.RegisterRunner(Plugin.PluginName(), Plugin)

	os.Exit(m.Run())
}

func TestHandleCallback(t *testing.T) {
	dbp, err := zesty.NewDBProvider(utask.DBName)
	assert.NoError(t, err)

	tt, err := tasktemplate.Create(dbp, "callback", "callback", nil, nil, []input.Input{}, []input.Input{}, []string{}, []string{}, false, false, nil, []values.Variable{}, nil, nil, "foo", nil, false, nil)
	assert.NoError(t, err)
	assert.NotNil(t, tt)

	tsk, err := task.Create(dbp, tt, "foo", []string{}, []string{}, []string{}, []string{}, []string{}, nil, nil, nil)
	assert.NoError(t, err)
	assert.NotNil(t, t)

	res, err := resolution.Create(dbp, tsk, nil, "bar", true, nil)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	tsk.Resolution = &res.PublicID
	tsk.Update(dbp, true, false)

	cb, err := createCallback(dbp, tsk, &CallbackContext{
		StepName:          "foo",
		TaskID:            tsk.PublicID,
		RequesterUsername: "foo",
	}, `{
		"$schema": "http://json-schema.org/schema#",
		"type": "object",
		"additionalProperties": false,
		"required": ["success"],
		"properties": {
			"success": {
				"type": "boolean"
			}
		}
	}`)
	assert.NoError(t, err)
	assert.NotNil(t, cb)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// invalid state
	out, err := HandleCallback(c, &handleCallbackIn{
		CallbackID:     cb.PublicID,
		CallbackSecret: cb.Secret,
		Body:           map[string]interface{}{},
	})
	assert.EqualError(t, err, "related task is not in a valid state: TODO")
	assert.Nil(t, out)

	// run the task
	tsk.SetState(task.StateRunning)
	tsk.Update(dbp, true, false)

	// Invalid JSON-schema
	out, err = HandleCallback(c, &handleCallbackIn{
		CallbackID:     cb.PublicID,
		CallbackSecret: cb.Secret,
		Body:           map[string]interface{}{},
	})
	assert.EqualError(t, err, fmt.Sprintf("unable to validate body: I[#] S[#] doesn't validate with %q\n  I[#] S[#/required] missing properties: %q", cb.PublicID+"#", "success"))
	assert.Nil(t, out)

	// valid body
	out, err = HandleCallback(c, &handleCallbackIn{
		CallbackID:     cb.PublicID,
		CallbackSecret: cb.Secret,
		Body: map[string]interface{}{
			"success": true,
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, out.Message, "The callback has been resolved")

	// already called
	out, err = HandleCallback(c, &handleCallbackIn{
		CallbackID:     cb.PublicID,
		CallbackSecret: cb.Secret,
		Body: map[string]interface{}{
			"success": true,
		},
	})
	assert.Nil(t, out)
	assert.EqualError(t, err, "callback has already been resolved")
}

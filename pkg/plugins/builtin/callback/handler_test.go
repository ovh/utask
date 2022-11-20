package plugincallback

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/juju/errors"
	"github.com/loopfz/gadgeto/iffy"
	"github.com/loopfz/gadgeto/zesty"
	"github.com/ovh/configstore"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/ovh/utask"
	"github.com/ovh/utask/api"
	"github.com/ovh/utask/db"
	"github.com/ovh/utask/engine"
	"github.com/ovh/utask/engine/step"
	"github.com/ovh/utask/engine/step/executor"
	"github.com/ovh/utask/models/tasktemplate"
	"github.com/ovh/utask/pkg/auth"
	compress "github.com/ovh/utask/pkg/compress/init"
	"github.com/ovh/utask/pkg/now"
	"github.com/ovh/utask/pkg/plugins"
)

const (
	adminUser   = "admin"
	regularUser = "regular"

	usernameHeaderKey = "x-remote-user"
)

var hdl http.Handler

func TestMain(m *testing.M) {
	store := configstore.DefaultStore
	store.InitFromEnvironment()

	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.ErrorLevel)

	step.RegisterRunner(Plugin.PluginName(), Plugin)

	srv := api.NewServer()
	srv.WithAuth(dumbIdentityProvider)
	srv.WithCustomMiddlewares(dummyCustomMiddleware)
	srv.SetDashboardPathPrefix("")
	srv.SetDashboardAPIPathPrefix("")
	srv.SetDashboardSentryDSN("")
	srv.SetEditorPathPrefix("")

	svc := &plugins.Service{Store: store, Server: srv}

	if err := Init.Init(svc); err != nil {
		panic(err)
	}

	if err := db.Init(store); err != nil {
		panic(err)
	}

	if err := now.Init(); err != nil {
		panic(err)
	}

	if err := auth.Init(store); err != nil {
		panic(err)
	}

	if err := compress.Register(); err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	ctx := context.Background()

	if err := engine.Init(ctx, &wg, store); err != nil {
		panic(err)
	}

	go srv.ListenAndServe()
	srvx := &http.Server{Addr: fmt.Sprintf(":%d", utask.FPort)}
	err := srvx.Shutdown(ctx)
	if err != nil {
		panic(err)
	}

	hdl = srv.Handler(ctx)

	os.Exit(m.Run())
}

func dummyCustomMiddleware(c *gin.Context) {
	c.Next()
}

func dumbIdentityProvider(r *http.Request) (string, error) {
	username := r.Header.Get(usernameHeaderKey)

	if username != adminUser && username != regularUser {
		return "", errors.New("unknown user")
	}
	return username, nil
}

func waitChecker(dur time.Duration) iffy.Checker {
	return func(r *http.Response, body string, respObject interface{}) error {
		time.Sleep(dur)
		return nil
	}
}

var (
	adminHeaders = iffy.Headers{
		usernameHeaderKey: adminUser,
	}
	regularHeaders = iffy.Headers{
		usernameHeaderKey: regularUser,
	}
)

func callbackTemplate() (*tasktemplate.TaskTemplate, error) {
	schema := `{
		"$schema": "http://json-schema.org/schema#",
		"type": "object",
		"additionalProperties": false,
		"required": ["success"],
		"properties": {
			"success": {
				"type": "boolean"
			}
		}
	}`
	jsonSchema, err := json.Marshal(schema)
	if err != nil {
		return nil, err
	}

	return &tasktemplate.TaskTemplate{
		Name:        "callback",
		Description: "callback",
		TitleFormat: "callback",
		Steps: map[string]*step.Step{
			"createCb": {
				Action: executor.Executor{
					Type:          "callback",
					Configuration: json.RawMessage(`{"action": "create", "schema": ` + string(jsonSchema) + `}`),
				},
			},
			"waitCb": {
				Dependencies: []string{"createCb"},
				Action: executor.Executor{
					Type:          "callback",
					Configuration: json.RawMessage(`{"action": "wait", "id": "{{.step.createCb.output.id}}"}`),
				},
			},
		},
	}, nil
}

func TestHandleCallback(t *testing.T) {
	tester := iffy.NewTester(t, hdl)

	dbp, err := zesty.NewDBProvider(utask.DBName)
	assert.NoError(t, err)

	callback, err := callbackTemplate()
	if err != nil {
		t.Fatal(err)
	}

	tmpl, err := tasktemplate.LoadFromName(dbp, callback.Name)
	if err != nil {
		if !errors.IsNotFound(err) {
			t.Fatal(err)
		}
		if err := dbp.DB().Insert(callback); err != nil {
			t.Fatal(err)
		}
		tmpl, err = tasktemplate.LoadFromName(dbp, callback.Name)
		if err != nil {
			if !errors.IsNotFound(err) {
				t.Fatal(err)
			}
		}
	}

	tester.AddCall("getTemplate", http.MethodGet, "/template/"+tmpl.Name, "").
		Headers(regularHeaders).
		Checkers(
			iffy.ExpectStatus(200),
		)

	tester.AddCall("newTask", http.MethodPost, "/task", `{"template_name":"{{.getTemplate.name}}", "input": {}}`).
		Headers(regularHeaders).
		Checkers(iffy.ExpectStatus(201))

	tester.AddCall("createResolution", http.MethodPost, "/resolution", `{"task_id":"{{.newTask.id}}"}`).
		Headers(adminHeaders).
		Checkers(iffy.ExpectStatus(201))

	tester.AddCall("runResolution", http.MethodPost, "/resolution/{{.createResolution.id}}/run", "").
		Headers(adminHeaders).
		Checkers(
			iffy.ExpectStatus(204),
			waitChecker(time.Second), // fugly... need to give resolution manager some time to asynchronously finish running
		)

	tester.AddCall("getResolution", http.MethodGet, "/resolution/{{.createResolution.id}}", "").
		Headers(adminHeaders).
		Checkers(
			//iffy.DumpResponse(t),
			iffy.ExpectStatus(200),
			iffy.ExpectJSONBranch("state", "WAITING"),
		)

	tester.AddCall("resolveCallback", http.MethodPost, "/unsecured/callback/{{.getResolution.steps.createCb.output.id}}", `{"success": true}`).
		Checkers(
			//iffy.DumpResponse(t),
			iffy.ExpectStatus(400),
			iffy.ExpectJSONBranch("error", "binding error on field 'CallbackSecret' of type 'handleCallbackIn': missing query parameter: t"),
		)

	tester.AddCall("resolveCallback", http.MethodPost, "/unsecured/callback/{{.getResolution.steps.createCb.output.id}}?t={{.getResolution.steps.createCb.output.token}}", ``).
		Checkers(
			//iffy.DumpResponse(t),
			iffy.ExpectStatus(400),
			iffy.ExpectJSONFields("error"),
		)

	tester.AddCall("resolveCallback", http.MethodPost, "/unsecured/callback/{{.getResolution.steps.createCb.output.id}}?t={{.getResolution.steps.createCb.output.token}}", `{}`).
		Checkers(
			//iffy.DumpResponse(t),
			iffy.ExpectStatus(400),
			iffy.ExpectJSONFields("error"),
		)

	tester.AddCall("resolveCallback", http.MethodPost, "/unsecured/callback/{{.getResolution.steps.createCb.output.id}}?t={{.getResolution.steps.createCb.output.token}}", `{"success": "true"}`).
		Checkers(
			//iffy.DumpResponse(t),
			iffy.ExpectStatus(400),
			iffy.ExpectJSONFields("error"),
		)

	tester.AddCall("resolveCallback", http.MethodPost, "/unsecured/callback/{{.getResolution.steps.createCb.output.id}}?t={{.getResolution.steps.createCb.output.token}}", `{"success": true}`).
		Checkers(
			//iffy.DumpResponse(t),
			iffy.ExpectStatus(200),
			iffy.ExpectJSONBranch("message", "The callback has been resolved"),
			waitChecker(time.Second), // fugly... need to give resolution manager some time to asynchronously finish running
		)

	tester.AddCall("resolveCallback", http.MethodPost, "/unsecured/callback/{{.getResolution.steps.createCb.output.id}}?t={{.getResolution.steps.createCb.output.token}}", `{"success": false}`).
		Checkers(
			//iffy.DumpResponse(t),
			iffy.ExpectStatus(400),
			iffy.ExpectJSONBranch("error", "callback has already been resolved"),
			waitChecker(time.Second), // fugly... need to give resolution manager some time to asynchronously finish running
		)

	tester.AddCall("getResolution", http.MethodGet, "/resolution/{{.createResolution.id}}", "").
		Headers(adminHeaders).
		Checkers(
			//iffy.DumpResponse(t),
			iffy.ExpectStatus(200),
			iffy.ExpectJSONBranch("state", "DONE"),
			iffy.ExpectJSONBranch("steps", "waitCb", "output", "body", "success", "true"),
		)

	tester.Run()
}

package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/juju/errors"
	"github.com/loopfz/gadgeto/iffy"
	"github.com/loopfz/gadgeto/zesty"
	"github.com/ovh/configstore"
	"github.com/sirupsen/logrus"

	"github.com/ovh/utask"
	"github.com/ovh/utask/api"
	"github.com/ovh/utask/db"
	"github.com/ovh/utask/engine"
	"github.com/ovh/utask/engine/input"
	"github.com/ovh/utask/engine/step"
	"github.com/ovh/utask/engine/step/executor"
	"github.com/ovh/utask/models/task"
	"github.com/ovh/utask/models/tasktemplate"
	"github.com/ovh/utask/pkg/auth"
	"github.com/ovh/utask/pkg/now"
	"github.com/ovh/utask/pkg/plugins/builtin/echo"
	"github.com/ovh/utask/pkg/plugins/builtin/script"
	"github.com/ovh/utask/pkg/utils"
)

const (
	adminUser    = "admin"
	regularUser  = "regular"
	resolverUser = "resolver"

	usernameHeaderKey = "x-remote-user"
)

var hdl http.Handler

func TestMain(m *testing.M) {
	store := configstore.DefaultStore
	store.InitFromEnvironment()

	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.ErrorLevel)

	step.RegisterRunner(echo.Plugin.PluginName(), echo.Plugin)
	step.RegisterRunner(script.Plugin.PluginName(), script.Plugin)

	db.Init(store)

	now.Init()

	auth.Init(store)

	ctx := context.Background()
	var wg sync.WaitGroup
	if err := engine.Init(ctx, &wg, store); err != nil {
		panic(err)
	}

	srv := api.NewServer()
	srv.WithAuth(dumbIdentityProvider)
	srv.SetDashboardPathPrefix("")
	srv.SetDashboardAPIPathPrefix("")
	srv.SetDashboardSentryDSN("")
	srv.SetEditorPathPrefix("")

	go srv.ListenAndServe()
	srvx := &http.Server{Addr: fmt.Sprintf(":%d", utask.FPort)}
	err := srvx.Shutdown(ctx)
	if err != nil {
		panic(err)
	}

	hdl = srv.Handler(ctx)

	os.Exit(m.Run())
}

func dumbIdentityProvider(r *http.Request) (string, error) {
	username := r.Header.Get(usernameHeaderKey)

	if username != adminUser && username != regularUser {
		return "", errors.New("unknown user")
	}
	return r.Header.Get(usernameHeaderKey), nil
}

var (
	adminHeaders = iffy.Headers{
		usernameHeaderKey: adminUser,
	}
	regularHeaders = iffy.Headers{
		usernameHeaderKey: regularUser,
	}

	invalidTemplate = tasktemplate.TaskTemplate{
		Name:        "invalid-template-1",
		Description: "Invalid template",
		TitleFormat: "Invalid template",
		Inputs: []input.Input{
			{
				Name:        "input-with-redundant-regex",
				LegalValues: []interface{}{"a", "b", "c"},
				Regex:       strPtr("^d.+$"),
			},
		},
	}
)

func TestUtils(t *testing.T) {
	tester := iffy.NewTester(t, hdl)

	tester.AddCall("testMetrics", http.MethodGet, "/metrics", "").
		Checkers(
			iffy.ExpectStatus(200),
		)

	tester.AddCall("testPing", http.MethodGet, "/unsecured/mon/ping", "").
		Checkers(
			iffy.ExpectStatus(200),
		)

	tester.AddCall("testRootHandler", http.MethodGet, "/", "").
		Headers(regularHeaders).
		Checkers(
			iffy.ExpectStatus(301),
		)

	tester.AddCall("testMeta", http.MethodGet, "/meta", "").
		Headers(regularHeaders).
		Checkers(
			iffy.ExpectStatus(200),
		)

	tester.AddCall("testMaintenanceMod", http.MethodPost, "/task", "").
		Headers(adminHeaders).
		Checkers(iffy.ExpectStatus(405))

	tester.AddCall("testKeyRoate", http.MethodPost, "/key-rotate", "").
		Headers(adminHeaders).
		Checkers(iffy.ExpectStatus(200))

	tester.AddCall("testIsAdmin", http.MethodPost, "/key-rotate", "").
		Headers(regularHeaders).
		Checkers(iffy.ExpectStatus(401))

	utask.FMaintenanceMode = true

	tester.Run()

	utask.FMaintenanceMode = false
}

func TestPasswordInput(t *testing.T) {
	tester := iffy.NewTester(t, hdl)

	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		t.Fatal(err)
	}

	tmpl := templateWithPasswordInput()

	_, err = tasktemplate.LoadFromName(dbp, tmpl.Name)
	if err != nil {
		if !errors.IsNotFound(err) {
			t.Fatal(err)
		}
		if err := dbp.DB().Insert(&tmpl); err != nil {
			t.Fatal(err)
		}
	}

	tester.AddCall("getTemplate", http.MethodGet, "/template/input-password", "").
		Headers(regularHeaders).
		Checkers(
			iffy.ExpectStatus(200),
		)

	tester.AddCall("getTemplateWithoutAuth", http.MethodGet, "/template/input-password", "").
		Checkers(
			iffy.ExpectStatus(401),
		)

	tester.AddCall("newTask", http.MethodPost, "/task", `{"template_name":"input-password","input":{"verysecret":"abracadabra"}}`).
		Headers(regularHeaders).
		Checkers(iffy.ExpectStatus(201))

	tester.AddCall("createComment", http.MethodPost, "/task/{{.newTask.id}}/comment", `{"content":"I'm a pickle rick"}`).
		Headers(regularHeaders).
		Checkers(iffy.ExpectStatus(201))

	tester.AddCall("getObfuscated", http.MethodGet, "/task/{{.newTask.id}}", "").
		Headers(regularHeaders).
		Checkers(
			iffy.ExpectStatus(200),
			iffy.ExpectJSONBranch("input", "verysecret", "**__SECRET__**"),
		)

	tester.AddCall("getClear", http.MethodGet, "/task/{{.newTask.id}}", "").
		Headers(adminHeaders).
		Checkers(
			iffy.ExpectStatus(200),
			iffy.ExpectJSONBranch("input", "verysecret", "abracadabra"),
		)

	tester.AddCall("ignoreUpdate", http.MethodPut, "/task/{{.newTask.id}}", `{"input":{"verysecret":"**__SECRET__**"}}`).
		Headers(regularHeaders).
		Checkers(iffy.ExpectStatus(200))

	tester.AddCall("getClear2", http.MethodGet, "/task/{{.newTask.id}}", "").
		Headers(adminHeaders).
		Checkers(
			iffy.ExpectStatus(200),
			iffy.ExpectJSONBranch("input", "verysecret", "abracadabra"),
		)

	tester.AddCall("realUpdate", http.MethodPut, "/task/{{.newTask.id}}", `{"input":{"verysecret":"expectopatronum"}}`).
		Headers(regularHeaders).
		Checkers(iffy.ExpectStatus(200))

	tester.AddCall("getClear3", http.MethodGet, "/task/{{.newTask.id}}", "").
		Headers(adminHeaders).
		Checkers(
			iffy.ExpectStatus(200),
			iffy.ExpectJSONBranch("input", "verysecret", "expectopatronum"),
		)

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
			iffy.ExpectJSONBranch("state", "DONE"),
		)

	tester.AddCall("getTaskResult", http.MethodGet, "/task/{{.newTask.id}}", "").
		Headers(regularHeaders).
		Checkers(
			iffy.ExpectStatus(200),
			iffy.ExpectJSONBranch("state", "DONE"),
			iffy.ExpectJSONBranch("result", "revealed", "expectopatronum"),
		)

	tester.AddCall("fetchStatistics", http.MethodGet, "/unsecured/stats", "").
		Headers(regularHeaders).
		Checkers(iffy.ExpectStatus(200))

	tester.Run()
}

func TestPagination(t *testing.T) {
	tester := iffy.NewTester(t, hdl)

	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		t.Fatal(err)
	}

	dummy := dummyTemplate()

	tmpl, err := tasktemplate.LoadFromName(dbp, dummy.Name)
	if err != nil {
		if !errors.IsNotFound(err) {
			t.Fatal(err)
		}
		if err := dbp.DB().Insert(&dummy); err != nil {
			t.Fatal(err)
		}
		tmpl, err = tasktemplate.LoadFromName(dbp, dummy.Name)
		if err != nil {
			if !errors.IsNotFound(err) {
				t.Fatal(err)
			}
		}
	}

	cnt := 20
	var midTask task.Task
	for i := 0; i < cnt; i++ {
		tsk, err := task.Create(dbp, tmpl, regularUser, nil, nil, map[string]interface{}{"id": strconv.Itoa(i)}, nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		if i == cnt/2 {
			midTask = *tsk
		}
		time.Sleep(time.Second / 5)
	}

	var tasks []*task.Task
	tester.AddCall("list tasks", http.MethodGet, "/task?page_size=20", "").
		Headers(regularHeaders).
		ResponseObject(&tasks).
		Checkers(
			iffy.ExpectListLength(20),
		)

	tester.Run()

	first := tasks[0]
	last := tasks[len(tasks)-1]

	if first.LastActivity.Before(last.LastActivity) {
		t.Fatal("first list elements should be latest in time")
	}

	tester2 := iffy.NewTester(t, hdl)

	var tasksBefore []*task.Task
	tester2.AddCall("list tasks before midTask", http.MethodGet, "/task?before="+url.QueryEscape(midTask.LastActivity.Format(time.RFC3339Nano)), "").
		Headers(regularHeaders).
		ResponseObject(&tasksBefore).
		Checkers(iffy.ExpectStatus(200))

	var tasksAfter []*task.Task
	tester2.AddCall("list tasks after midTask", http.MethodGet, "/task?after="+url.QueryEscape(midTask.LastActivity.Format(time.RFC3339Nano)), "").
		Headers(regularHeaders).
		ResponseObject(&tasksAfter).
		Checkers(iffy.ExpectStatus(200))

	tester2.Run()

	for _, tsk := range tasksBefore {
		if !tsk.LastActivity.Before(midTask.LastActivity) {
			t.Fatal("All tasks in this list should be before midTask")
		}
	}

	for _, tsk := range tasksAfter {
		if !tsk.LastActivity.After(midTask.LastActivity) {
			t.Fatal("All tasks in the list should be after midTask")
		}
	}
}

const (
	blockedTemplate          = "blocked-template"
	hiddenTemplate           = "hidden-template"
	blockedAndHiddenTemplate = "blocked-and-hidden"
)

var (
	hiddenTemplates  = []string{hiddenTemplate, blockedAndHiddenTemplate}
	blockedTemplates = []string{blockedTemplate, blockedAndHiddenTemplate}
)

func TestBlockedHiddenTemplates(t *testing.T) {
	tester := iffy.NewTester(t, hdl)

	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		t.Fatal(err)
	}

	blockedTmpl := blockedHidden(blockedTemplate, true, false)
	hiddenTmpl := blockedHidden(hiddenTemplate, false, true)
	blockedAndHiddenTmpl := blockedHidden(blockedAndHiddenTemplate, true, true)

	for _, tmpl := range []tasktemplate.TaskTemplate{blockedTmpl, hiddenTmpl, blockedAndHiddenTmpl} {
		_, err = tasktemplate.LoadFromName(dbp, tmpl.Name)
		if err != nil {
			if !errors.IsNotFound(err) {
				t.Fatal(err)
			}
			if err := dbp.DB().Insert(&tmpl); err != nil {
				t.Fatal(err)
			}
		}
	}

	var adminTmplList []*tasktemplate.TaskTemplate
	tester.AddCall("admin list templates", http.MethodGet, "/template", "").
		Headers(adminHeaders).
		ResponseObject(&adminTmplList).
		Checkers(
			iffy.ExpectStatus(200),
		)

	var regularTmplList []*tasktemplate.TaskTemplate
	tester.AddCall("regular list templates", http.MethodGet, "/template", "").
		Headers(regularHeaders).
		ResponseObject(&regularTmplList).
		Checkers(
			iffy.ExpectStatus(200),
		)

	for _, tmpl := range blockedTemplates {
		tester.AddCall("admin-"+tmpl, http.MethodPost, "/task", fmt.Sprintf("{\"template_name\":\"%s\"}", tmpl)).
			Headers(adminHeaders).
			Checkers(iffy.ExpectStatus(400))

		tester.AddCall("regular-"+tmpl, http.MethodPost, "/task", fmt.Sprintf("{\"template_name\":\"%s\"}", tmpl)).
			Headers(regularHeaders).
			Checkers(iffy.ExpectStatus(400))
	}

	tester.Run()

	// every hidden template is found in admin's template list
	adminTmplListNames := []string{}
	for _, tmpl := range adminTmplList {
		adminTmplListNames = append(adminTmplListNames, tmpl.Name)
	}
	for _, tmpl := range hiddenTemplates {
		if !utils.ListContainsString(adminTmplListNames, tmpl) {
			t.Fatalf("%s template should be visible for admin users", tmpl)
		}
	}

	// no hidden template is visible in regular's template list
	for _, tmpl := range regularTmplList {
		if utils.ListContainsString(hiddenTemplates, tmpl.Name) {
			t.Fatalf("%s template should not be visible for regular users", tmpl.Name)
		}
	}
}

func waitChecker(dur time.Duration) iffy.Checker {
	return func(r *http.Response, body string, respObject interface{}) error {
		time.Sleep(dur)
		return nil
	}
}

func templatesWithInvalidInputs() []tasktemplate.TaskTemplate {
	var tt []tasktemplate.TaskTemplate
	for _, inp := range []input.Input{
		{
			Name:        "input-with-redundant-regex",
			LegalValues: []interface{}{"a", "b", "c"},
			Regex:       strPtr("^d.+$"),
		},
		{
			Name:  "input-with-bad-regex",
			Regex: strPtr("^^[d.+$"),
		},
		{
			Name: "input-with-bad-type",
			Type: "bad-type",
		},
		{
			Name:        "input-with-bad-legal-values",
			Type:        "number",
			LegalValues: []interface{}{"a", "b", "c"},
		},
	} {
		tt = append(tt, tasktemplate.TaskTemplate{
			Name:        "invalid-template",
			Description: "Invalid template",
			TitleFormat: "Invalid template",
			Inputs: []input.Input{
				inp,
			},
		})
	}
	return tt
}

func templateWithPasswordInput() tasktemplate.TaskTemplate {
	return tasktemplate.TaskTemplate{
		Name:        "input-password",
		Description: "input-password",
		TitleFormat: "input-password",
		ResultFormat: map[string]interface{}{
			"revealed": "{{.step.stepOne.output.showSecret}}",
		},
		Inputs: []input.Input{
			{
				Name: "verysecret",
				Type: "password",
			},
		},
		Steps: map[string]*step.Step{
			"stepOne": {
				Action: executor.Executor{
					Type: "echo",
					Configuration: json.RawMessage(`{
						"output": {"showSecret":"{{.input.verysecret}}"}
					}`),
				},
			},
		},
	}
}

func dummyTemplate() tasktemplate.TaskTemplate {
	return tasktemplate.TaskTemplate{
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
}

func blockedHidden(name string, blocked, hidden bool) tasktemplate.TaskTemplate {
	return tasktemplate.TaskTemplate{
		Name:        name,
		Description: "does nothing",
		TitleFormat: "this task does nothing at all",
		Blocked:     blocked,
		Hidden:      hidden,
	}
}

func Test_staticMiddleware(t *testing.T) {
	ginEngine := gin.Default()
	ginEngine.
		Group("/", api.StaticFilePatternReplaceMiddleware("static.go", "___test_suite___")).
		StaticFS("/", http.Dir("./"))

	tester := iffy.NewTester(t, ginEngine)

	tester.AddCall("retrieve test folder index and validate replacement OK", http.MethodGet, "/", "").
		Headers(adminHeaders).
		Checkers(
			iffy.ExpectStatus(200),
			expectStringNotPresent("static.go"),
			expectStringPresent("___test_suite___"),
		)

	tester.AddCall("unknown static page", http.MethodGet, "/dsqdzdzodkzdzdz", "").
		Headers(adminHeaders).
		Checkers(
			iffy.ExpectStatus(404),
		)

	tester.Run()
}

func expectStringNotPresent(value string) iffy.Checker {
	return func(r *http.Response, body string, respObject interface{}) error {
		if strings.Contains(body, value) {
			return fmt.Errorf("Response body invalid: should not contains %q, but it does", value)
		}
		return nil
	}
}

func expectStringPresent(value string) iffy.Checker {
	return func(r *http.Response, body string, respObject interface{}) error {
		if !strings.Contains(body, value) {
			return fmt.Errorf("Response body invalid: should contains %q, but it doesn't", value)
		}
		return nil
	}
}

func marshalJSON(t *testing.T, i interface{}) string {
	jsonBytes, err := json.Marshal(i)
	if err != nil {
		t.Fatal(err)
	}
	return string(jsonBytes)
}

func strPtr(s string) *string { return &s }

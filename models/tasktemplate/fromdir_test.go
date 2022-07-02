package tasktemplate_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/loopfz/gadgeto/zesty"
	"github.com/ovh/configstore"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/ovh/utask"
	"github.com/ovh/utask/db"
	"github.com/ovh/utask/engine/step"
	"github.com/ovh/utask/models/task"
	"github.com/ovh/utask/models/tasktemplate"
	"github.com/ovh/utask/pkg/now"
	"github.com/ovh/utask/pkg/plugins/builtin/echo"
	"github.com/ovh/utask/pkg/plugins/builtin/script"
	pluginsubtask "github.com/ovh/utask/pkg/plugins/builtin/subtask"
)

func TestMain(m *testing.M) {
	store := configstore.NewStore()
	store.InitFromEnvironment()

	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.ErrorLevel)

	step.RegisterRunner(echo.Plugin.PluginName(), echo.Plugin)
	step.RegisterRunner(script.Plugin.PluginName(), script.Plugin)
	step.RegisterRunner(pluginsubtask.Plugin.PluginName(), pluginsubtask.Plugin)

	if err := db.Init(store); err != nil {
		panic(err)
	}

	if err := now.Init(); err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

func TestLoadFromDir(t *testing.T) {
	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		t.Fatal(err)
	}

	_, err = dbp.DB().Query("DELETE FROM resolution;")
	assert.Nil(t, err, "Emptying database failed")
	_, err = dbp.DB().Query("DELETE FROM task_comment;")
	assert.Nil(t, err, "Emptying database failed")
	_, err = dbp.DB().Query("DELETE FROM task;")
	assert.Nil(t, err, "Emptying database failed")
	_, err = dbp.DB().Query("DELETE FROM task_template;")
	assert.Nil(t, err, "Emptying database failed")

	err = tasktemplate.LoadFromDir(dbp, "templates_tests")
	assert.Nil(t, err, "LoadFromDir failed")

	taskTemplatesFromDatabase, err := tasktemplate.ListTemplates(dbp, true, 10, nil)
	assert.Nil(t, err, "ListTemplates failed")
	assert.Len(t, taskTemplatesFromDatabase, 2, "wrong size of imported templates")

	tt := tasktemplate.TaskTemplate{}
	tmpl, err := ioutil.ReadFile(path.Join("templates_tests", "hello-world-now.yaml"))
	assert.Nil(t, err, "unable to read file hello-world-now.yaml")
	err = yaml.Unmarshal(tmpl, &tt)
	assert.Nil(t, err, "unable to unmarshal tasktemplate")

	tt.Name = "a WONDERFUL new template"
	tt.Normalize()

	err = tt.Valid()
	assert.Nil(t, err, "unable to valid new template")

	err = dbp.DB().Insert(&tt)
	assert.Nil(t, err, "unable to insert new template")

	taskTemplatesFromDatabase, err = tasktemplate.ListTemplates(dbp, true, 10, nil)
	assert.Nil(t, err, "ListTemplates failed")
	assert.Len(t, taskTemplatesFromDatabase, 3, "wrong size of imported templates")

	err = tasktemplate.LoadFromDir(dbp, "templates_tests")
	assert.Nil(t, err, "LoadFromDir failed")

	taskTemplatesFromDatabase, err = tasktemplate.ListTemplates(dbp, true, 10, nil)
	assert.Nil(t, err, "ListTemplates failed")
	assert.Len(t, taskTemplatesFromDatabase, 2, "wrong size of imported templates")

	tt.ID = 0
	err = dbp.DB().Insert(&tt)
	assert.Nil(t, err, "unable to insert new template")

	_, err = task.Create(dbp, &tt, "admin", []string{}, []string{}, []string{}, map[string]interface{}{}, nil, nil)
	assert.Nil(t, err, "unable to create task")

	err = tasktemplate.LoadFromDir(dbp, "templates_tests")
	assert.Nil(t, err, "LoadFromDir failed")

	taskTemplatesFromDatabase, err = tasktemplate.ListTemplates(dbp, true, 10, nil)
	assert.Nil(t, err, "ListTemplates failed")
	assert.Len(t, taskTemplatesFromDatabase, 3, "wrong size of imported templates")

	tt2, err := tasktemplate.LoadFromName(dbp, tt.Name)
	assert.Nil(t, err, "unable to load tt2")
	assert.False(t, tt.Hidden, "previous template should have not been hidden")
	assert.False(t, tt.Blocked, "previous template should have not been blocked")
	assert.True(t, tt2.Hidden, "template should have been hidden as not existing in dir but have linked task")
	assert.True(t, tt2.Blocked, "template should have been blocked as not existing in dir but have linked task")
}

func TestInvalidVariablesTemplates(t *testing.T) {
	tt := tasktemplate.TaskTemplate{}
	tmpl, err := ioutil.ReadFile(path.Join("templates_errors_tests", "error-variables.yaml"))
	assert.Nil(t, err, "unable to read file error-variables.yaml")
	err = yaml.Unmarshal(tmpl, &tt)
	assert.Nil(t, err, "unable to unmarshal tasktemplate")

	tt.Normalize()
	err = tt.Valid()
	assert.Contains(t, fmt.Sprint(err), "can't have both value and expression defined")

	tt.Variables[0].Name = ""
	tt.Normalize()
	err = tt.Valid()
	assert.Contains(t, fmt.Sprint(err), "variable name can't be empty")

	tt.Variables[0].Name = "toto"
	tt.Variables[0].Expression = ""
	tt.Variables[0].Value = nil
	tt.Normalize()
	err = tt.Valid()
	assert.Contains(t, fmt.Sprint(err), "expression and value can't be empty at the same time")
}

func TestDependenciesValidation(t *testing.T) {
	tt := tasktemplate.TaskTemplate{}
	tmpl, err := ioutil.ReadFile(path.Join("templates_errors_tests", "error-variables.yaml"))
	assert.Nil(t, err, "unable to read file error-variables.yaml")
	err = yaml.Unmarshal(tmpl, &tt)
	assert.Nil(t, err, "unable to unmarshal tasktemplate")
	tt.Variables[0].Expression = ""

	tt.Normalize()
	err = tt.Valid()
	assert.Nil(t, err, "validation failed: %s", err)

	tt.Steps["step2"].Dependencies[0] = "sayHello:NOT_FOUND,DONE"
	tt.Normalize()
	err = tt.Valid()
	assert.Nil(t, err, "validation failed: %s", err)

	tt.Steps["step2"].Dependencies[0] = "sayHello:ANY"
	tt.Normalize()
	err = tt.Valid()
	assert.Nil(t, err, "validation failed: %s", err)

	tt.Steps["step2"].Dependencies[0] = "sayHello:ANY,NOT_FOUND"
	tt.Normalize()
	err = tt.Valid()
	assert.Contains(t, fmt.Sprint(err), "Invalid dependency, no other state allowed if ANY is declared")

	tt.Steps["step2"].Dependencies[0] = "sayHello:DONE,NOT_FOUND,ALREADY_EXISTS,DONE"
	tt.Normalize()
	err = tt.Valid()
	assert.Contains(t, fmt.Sprint(err), "Invalid dependency, duplicated state detected")

	tt.Steps["step2"].Dependencies[0] = "sayHello:ALREADY_EXISTS, NOT_FOUND"
	tt.Normalize()
	err = tt.Valid()
	assert.Contains(t, fmt.Sprint(err), "Invalid dependency on step sayHello, step state not allowed: \" NOT_FOUND\"")

	tt.Steps["step2"].Dependencies[0] = "sayHello:"
	tt.Normalize()
	err = tt.Valid()
	assert.Contains(t, fmt.Sprint(err), "Invalid dependency on step sayHello, step state not allowed: \"\"")

	tt.Steps["step2"].Dependencies[0] = "sayHello:foobar"
	tt.Normalize()
	err = tt.Valid()
	assert.Contains(t, fmt.Sprint(err), "Invalid dependency on step sayHello, step state not allowed: \"foobar\"")

	tt.Steps["step2"].Dependencies[0] = "sayHello:ALREADY_EXISTS"
	tt.Steps["step2"].Dependencies = append(tt.Steps["step2"].Dependencies, "foobar")
	tt.Normalize()
	err = tt.Valid()
	assert.Contains(t, fmt.Sprint(err), "Invalid dependency, no step with that name: \"foobar\"")

	tt.Steps["step2"].Dependencies[0] = "sayHello"
	tt.Steps["step2"].Dependencies[1] = "sayHello:NOT_FOUND"
	tt.Normalize()
	err = tt.Valid()
	assert.Contains(t, fmt.Sprint(err), "Invalid dependency, already defined dependency to: \"sayHello\"")
}

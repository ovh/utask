package engine_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/juju/errors"
	"github.com/loopfz/gadgeto/zesty"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/configstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"

	"github.com/ovh/utask"
	"github.com/ovh/utask/api"
	"github.com/ovh/utask/db"
	"github.com/ovh/utask/db/pgjuju"
	"github.com/ovh/utask/db/sqlgenerator"
	"github.com/ovh/utask/engine"
	"github.com/ovh/utask/engine/functions"
	functionrunner "github.com/ovh/utask/engine/functions/runner"
	"github.com/ovh/utask/engine/step"
	"github.com/ovh/utask/engine/step/condition"
	"github.com/ovh/utask/engine/step/executor"
	"github.com/ovh/utask/engine/values"
	"github.com/ovh/utask/models/resolution"
	"github.com/ovh/utask/models/task"
	"github.com/ovh/utask/models/tasktemplate"
	compress "github.com/ovh/utask/pkg/compress/init"
	"github.com/ovh/utask/pkg/now"
	"github.com/ovh/utask/pkg/plugins"
	pluginbatch "github.com/ovh/utask/pkg/plugins/builtin/batch"
	plugincallback "github.com/ovh/utask/pkg/plugins/builtin/callback"
	"github.com/ovh/utask/pkg/plugins/builtin/echo"
	"github.com/ovh/utask/pkg/plugins/builtin/script"
	pluginsubtask "github.com/ovh/utask/pkg/plugins/builtin/subtask"
	"github.com/ovh/utask/pkg/taskutils"
)

const (
	testDirTemplates = "./templates_tests"
	testDirFunctions = "./functions_tests"
)

var (
	templateList = loadTemplates()
)

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

	if err := functions.LoadFromDir(testDirFunctions); err != nil {
		panic(err)
	}
	if err := functionrunner.Init(); err != nil {
		panic(err)
	}

	step.RegisterRunner(echo.Plugin.PluginName(), echo.Plugin)
	step.RegisterRunner(script.Plugin.PluginName(), script.Plugin)
	step.RegisterRunner(pluginsubtask.Plugin.PluginName(), pluginsubtask.Plugin)
	step.RegisterRunner(pluginbatch.Plugin.PluginName(), pluginbatch.Plugin)
	step.RegisterRunner(plugincallback.Plugin.PluginName(), plugincallback.Plugin)

	os.Exit(m.Run())
}

func loadTemplates() map[string][]byte {
	templateList := map[string][]byte{}
	files, err := os.ReadDir(testDirTemplates)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		info, _ := file.Info()
		if info.Mode().IsRegular() {
			bytes, err := os.ReadFile(filepath.Join(testDirTemplates, file.Name()))
			if err != nil {
				panic(err)
			}
			templateList[file.Name()] = bytes
		}
	}
	return templateList
}

////

func runTask(tmplName string, inputs, resolverInputs map[string]interface{}) (*resolution.Resolution, error) {
	res, err := createResolution(tmplName, inputs, resolverInputs)
	if err != nil {
		return nil, err
	}
	return engine.GetEngine().SyncResolve(res.PublicID, nil)
}

func createResolution(tmplName string, inputs, resolverInputs map[string]interface{}) (*resolution.Resolution, error) {
	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return nil, err
	}
	tmpl, err := templateFromYAML(dbp, tmplName)
	if err != nil {
		return nil, err
	}
	tsk, err := task.Create(dbp, tmpl, "", nil, nil, nil, nil, nil, inputs, nil, nil, false)
	if err != nil {
		return nil, err
	}
	res, err := resolution.Create(dbp, tsk, resolverInputs, "", false, nil)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func updateResolution(res *resolution.Resolution) error {
	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return err
	}
	return res.Update(dbp)
}

func runResolution(res *resolution.Resolution) (*resolution.Resolution, error) {
	if res == nil {
		return nil, errors.New("Nil resolution")
	}
	return engine.GetEngine().SyncResolve(res.PublicID, nil)
}

func templateFromYAML(dbp zesty.DBProvider, filename string) (*tasktemplate.TaskTemplate, error) {
	var tmpl tasktemplate.TaskTemplate

	file, ok := templateList[filename]
	if !ok {
		panic(errors.Errorf("No such file: %s", filename))
	}

	file = bytes.Replace(file, []byte("\t"), []byte("  "), -1)
	if err := yaml.Unmarshal(file, &tmpl); err != nil {
		return nil, err
	}
	if err := tmpl.Valid(); err != nil {
		return nil, err
	}
	tmpl.Normalize()
	if err := dbp.DB().Insert(&tmpl); err != nil {
		intErr := pgjuju.Interpret(err)
		if !errors.IsAlreadyExists(intErr) {
			return nil, intErr
		}
		existing, err := tasktemplate.LoadFromName(dbp, tmpl.Name)
		if err != nil {
			return nil, err
		}
		tmpl.ID = existing.ID
		if _, err := dbp.DB().Update(&tmpl); err != nil {
			return nil, err
		}
	}
	return tasktemplate.LoadFromName(dbp, tmpl.Name)
}

func listBatchTasks(dbp zesty.DBProvider, batchID int64) ([]string, error) {
	query, params, err := sqlgenerator.PGsql.
		Select("public_id").
		From("task").
		Where(squirrel.Eq{"id_batch": batchID}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var taskIDs []string
	_, err = dbp.DB().Select(&taskIDs, query, params...)
	return taskIDs, err
}

func TestSimpleTemplate(t *testing.T) {
	input := map[string]interface{}{
		"foo": "bar",
	}
	res, err := runTask("simple.yaml", input, nil)

	assert.Equal(t, nil, err)
	assert.Equal(t, resolution.StateError, res.State)
	assert.Equal(t, step.StateDone, res.Steps["stepOne"].State)
	assert.Equal(t, step.StateDone, res.Steps["stepTwo"].State)
	assert.Equal(t, step.StateServerError, res.Steps["stepThree"].State)

	assert.Equal(t, "FAIL!", res.Values.GetError("stepThree"))
}

func TestFunction(t *testing.T) {
	input := map[string]interface{}{}
	res, err := runTask("functionEchoHelloWorld.yaml", input, nil)

	require.Nil(t, err)
	assert.Equal(t, map[string]interface{}{
		"value": "Hello toto !",
	}, res.Steps["stepOne"].Output)
}

func TestFunctionBaseOutput(t *testing.T) {
	input := map[string]interface{}{}
	res, err := runTask("functionNested.yaml", input, nil)

	require.Nilf(t, err, "%s", err)
	assert.Equal(t, map[string]interface{}{
		"value":                "Hello foobar !",
		"nested1":              "foo",
		"nested2":              "foo",
		"base_nested":          "nested2",
		"base_output_template": "foo",
	}, res.Steps["stepOne"].Output)
	assert.Equal(t, "CUSTOM_STATE1", res.Steps["stepOne"].State)
}

func TestFunctionCustomState(t *testing.T) {
	input := map[string]interface{}{}
	res, err := runTask("functionCustomState.yaml", input, nil)

	require.Nil(t, err)
	assert.Equal(t, map[string]interface{}{
		"value": "Hello world!",
	}, res.Steps["stepOne"].Output)

	customStates, err := res.Steps["stepOne"].GetCustomStates()
	require.Nil(t, err)
	assert.Equal(t, []string{"STATE_HELLO"}, customStates)
}

func TestFunctionPreHook(t *testing.T) {
	input := map[string]interface{}{}
	res, err := runTask("functionPreHook.yaml", input, nil)

	assert.Equal(t, nil, err)
	assert.Equal(t, res.Steps["stepOne"].Output, map[string]interface{}{
		"value":    "Hello 42 !",
		"coalesce": "Coalesce 42!",
	})
}

func TestFunctionTemplatedOutput(t *testing.T) {
	input := map[string]interface{}{}
	res, err := runTask("functionEchoTemplatedOutput.yaml", input, nil)

	require.Nilf(t, err, "%s", err)
	assert.Equal(t, map[string]interface{}{
		"full_name": "John Doe",
	}, res.Steps["stepOne"].Output)
}

func TestClientError(t *testing.T) {
	res, err := runTask("clientError.yaml", map[string]interface{}{}, nil)

	assert.Equal(t, err, nil)
	assert.Equal(t, resolution.StateBlockedBadRequest, res.State)
	assert.Equal(t, step.StateClientError, res.Steps["stepOne"].State)
}

func TestMaxRetry(t *testing.T) {
	res, err := createResolution("maxRetry.yaml", map[string]interface{}{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 0, res.Steps["stepOne"].TryCount)
	res, err = runResolution(res)

	assert.Equal(t, err, nil)
	assert.Equal(t, resolution.StateBlockedMaxRetries, res.State)
	assert.Equal(t, step.StateServerError, res.Steps["stepOne"].State)
	assert.Equal(t, 1, res.Steps["stepOne"].TryCount)
}

func TestNextRetry(t *testing.T) {
	res, err := createResolution("nextRetry.yaml", map[string]interface{}{}, nil)
	if err != nil {
		t.Fatal(err)
	}

	var expectedNextRetry *time.Time
	assert.Equal(t, 0, res.Steps["stepOne"].TryCount)
	assert.Equal(t, expectedNextRetry, res.NextRetry)

	res, err = runResolution(res)
	assert.Nil(t, err)
	assert.Equal(t, resolution.StateError, res.State)
	assert.Equal(t, step.StateServerError, res.Steps["stepOne"].State)
	assert.Equal(t, 1, res.Steps["stepOne"].TryCount)
	assert.NotEqual(t, &time.Time{}, res.NextRetry)
}

func TestStepMaxRetries(t *testing.T) {
	res, err := createResolution("stepMaxRetries.yaml", map[string]interface{}{}, nil)

	assert.Equal(t, resolution.StateTODO, res.State)
	assert.Nil(t, err)
	assert.Equal(t, 0, res.Steps["stepOne"].TryCount)
	assert.Equal(t, 1, res.Steps["stepOne"].MaxRetries)

	for i := 0; i < 3; i++ {
		res, err = runResolution(res)
		assert.Nil(t, err)
	}

	assert.Equal(t, resolution.StateBlockedFatal, res.State)
	assert.Equal(t, 2, res.Steps["stepOne"].TryCount)
	assert.Equal(t, step.StateFatalError, res.Steps["stepOne"].State)
}

func TestLintingAndValidation(t *testing.T) {
	expectedResult := map[string]struct {
		nilResolution bool
		errstr        string
	}{
		"lintingError.yaml":                   {true, `Variable notfound does not exist`},
		"lintingRootKey.yaml":                 {true, `Variable grault does not exist`},
		"lintingReservedStep.yaml":            {true, `'this' step name is reserved`},
		"customStates.yaml":                   {true, `Custom state "SERVER_ERROR" is not allowed as it's a reserved state`},
		"forbiddenStateImpact.yaml":           {true, `Step condition cannot impact the state of step stepTwo, only those who belong to the dependency chain are allowed`},
		"stepDetailsLintingError.yaml":        {true, `Wrong step key: stepNotFound`},
		"circularDependencies.yaml":           {true, `Invalid: circular dependency [stepOne stepThree stepTwo] <-> step`}, // Last step name is random
		"selfDependency.yaml":                 {true, `Invalid: circular dependency [stepOne] <-> stepOne`},
		"orphanDependencies.yaml":             {true, `Invalid dependency, no step with that name: "stepTwo"`},
		"functionEchoHelloWorldError.yaml":    {true, `Invalid executor action: missing function_args "name"`},
		"conditionForeachSkipOnly.yaml":       {true, `Step condition can set foreach on a skip condition`},
		"conditionForeachInvalid.yaml":        {true, `Unknown condition foreach: invalid`},
		"conditionForeachStepNotForeach.yaml": {true, `Step condition cannot set foreach on a non-foreach step`},

		"lintingInfiniteOk.yaml":           {false, ""},
		"lintingObject.yaml":               {false, ""},
		"allowedStateImpact.yaml":          {false, ""},
		"functionEchoHelloWorld.yaml":      {false, ""},
		"functionCustomState.yaml":         {false, ""},
		"functionPreHook.yaml":             {false, ""},
		"functionEchoTemplatedOutput.yaml": {false, ""},
	}

	for template, testCase := range expectedResult {
		t.Run(template, func(t *testing.T) {
			res, err := createResolution(template, map[string]interface{}{}, nil)

			if testCase.nilResolution {
				assert.Nil(t, res)
			} else {
				assert.NotNil(t, res)
			}

			if testCase.errstr == "" {
				assert.Nil(t, err)
			} else {
				require.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.errstr)
			}
		})
	}
}

func TestComputeversion(t *testing.T) {
	res, err := createResolution("computeVersion.yaml", map[string]interface{}{}, nil)

	assert.Nil(t, err)
	assert.NotNil(t, res)

	expectedResult := map[string]string{
		"stepOne":   "http://json-schema.org/draft-07/schema#",
		"stepTwo":   "http://json-schema.org/draft-06/schema#",
		"stepThree": "http://json-schema.org/draft-04/schema#",
		"stepFour":  "http://json-schema.org/draft-04/schema#",
		"stepFive":  "http://json-schema.org/draft-07/schema#",
		"stepSix":   "http://json-schema.org/draft-07/schema#",
	}

	for name, step := range res.Steps {
		var m map[string]interface{}
		err := json.Unmarshal(step.Schema, &m)
		assert.Nil(t, err)

		v, ok := m["$schema"]
		if !ok {
			t.Errorf("$schema missing on step %q", name)
		}

		schemaVersion := v.(string)

		expectedVersion, ok := expectedResult[name]
		if !ok {
			t.Errorf("missing step %q in expected result", name)
		}

		if expectedVersion != schemaVersion {
			t.Errorf("Step %q, expected $schema to be %q, got %q", name, expectedVersion, schemaVersion)
		}
	}
}

func TestPrune(t *testing.T) {
	expectedResult := map[string]map[string]string{
		"skip": {
			"stepOne":   step.StatePrune,
			"stepTwo":   step.StateDone,
			"stepThree": step.StatePrune,
			"stepFour":  step.StatePrune,
			"stepFive":  step.StatePrune,
		},
		"not_skip": {
			"stepOne":   step.StateDone,
			"stepTwo":   step.StateDone,
			"stepThree": step.StatePrune,
			"stepFour":  step.StatePrune,
			"stepFive":  step.StateDone,
		},
	}

	for input := range expectedResult {
		res, err := runTask("prune.yaml", map[string]interface{}{
			"skipStepOne": input,
		}, nil)

		assert.Nil(t, err)
		assert.NotNil(t, res)

		for name, step := range res.Steps {
			expectedState, ok := expectedResult[input][name]
			if !ok {
				t.Errorf("Step %s not expected", name)
			}
			if step.State != expectedState {
				t.Errorf("Expected step %s to be %s, got %s (input: %s)", name, expectedState, step.State, input)
			}
		}
	}
}

func TestStepConditionStates(t *testing.T) {
	res, err := createResolution("stepCondition.yaml", map[string]interface{}{}, nil)

	assert.NotNil(t, res)
	assert.Nil(t, err)
	assert.Equal(t, step.StateTODO, res.Steps["stepOne"].State)
	assert.Equal(t, step.StateTODO, res.Steps["stepOneFinal"].State)
	assert.Equal(t, resolution.StateTODO, res.State)
	assert.Equal(t, 0, res.Steps["stepOne"].TryCount)
	assert.Equal(t, 0, res.Steps["stepOneFinal"].TryCount)

	res, err = runResolution(res)

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, step.StateToRetry, res.Steps["stepOne"].State)
	assert.Equal(t, step.StatePrune, res.Steps["stepOneFinal"].State)
	assert.Equal(t, step.StateClientError, res.Steps["stepSkip"].State)
	assert.Equal(t, step.StatePrune, res.Steps["stepSkipFinal"].State)
	assert.Equal(t, resolution.StateBlockedBadRequest, res.State)
	assert.Equal(t, 1, res.Steps["stepOne"].TryCount)
	assert.Equal(t, 1, res.Steps["stepOneFinal"].TryCount)

	res, err = runResolution(res)

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, step.StateToRetry, res.Steps["stepOne"].State)
	assert.Equal(t, step.StatePrune, res.Steps["stepOneFinal"].State)
	assert.Equal(t, resolution.StateBlockedBadRequest, res.State)
	assert.Equal(t, 2, res.Steps["stepOne"].TryCount)
	assert.Equal(t, 1, res.Steps["stepOneFinal"].TryCount)

	assert.Equal(t, "REGEXP_MATCH", res.Steps["stepTwo"].State)
	assert.Equal(t, "NOTREGEXP_MATCH", res.Steps["stepTwoBis"].State)
	assert.Equal(t, "FOO", res.Steps["stepThree"].State)
	assert.Equal(t, "changed state matched correctly", res.Steps["stepThree"].Error)

	assert.Equal(t, "OK", res.Steps["stepFour"].State)
	assert.Equal(t, "OK", res.Steps["stepFour"].Error)

	assert.Equal(t, "PRUNE", res.Steps["stepFive"].State)
	assert.Equal(t, "LIST_NOMATCH", res.Steps["stepSix"].State)

	assert.Equal(t, "DONE", res.Steps["stepForeach"].State)
	assert.Len(t, res.Steps["stepForeach"].Children, 3)
	assert.Equal(t, "VALID", res.Steps["stepForeachValidation"].State)
}

func TestResolutionStateCrashed(t *testing.T) {
	res, err := createResolution("stepCondition.yaml", map[string]interface{}{}, nil)
	assert.Nil(t, err)

	res.State = resolution.StateCrashed
	res.SetStepState("stepOne", step.StateRunning)
	err = updateResolution(res)
	assert.Nil(t, err)

	res, err = runResolution(res)
	assert.Nil(t, err)
	assert.Nil(t, res)
}

func TestResolutionStateCancelled(t *testing.T) {
	res, err := createResolution("stepCondition.yaml", map[string]interface{}{}, nil)
	assert.Nil(t, err)

	res.State = resolution.StateCancelled
	err = updateResolution(res)
	assert.Nil(t, err)

	res, err = runResolution(res)
	assert.Nil(t, res)
	assert.NotNil(t, err)
}

func TestResolutionStateDone(t *testing.T) {
	res, err := createResolution("stepCondition.yaml", map[string]interface{}{}, nil)
	assert.Nil(t, err)

	res.State = resolution.StateDone
	err = updateResolution(res)
	assert.Nil(t, err)

	res, err = runResolution(res)
	assert.Nil(t, res)
	assert.NotNil(t, err)
}

func TestResolutionStateRunning(t *testing.T) {
	res, err := createResolution("stepCondition.yaml", map[string]interface{}{}, nil)
	assert.Nil(t, err)

	res.State = resolution.StateRunning
	err = updateResolution(res)
	assert.Nil(t, err)

	res, err = runResolution(res)
	assert.Nil(t, res)
	assert.NotNil(t, err)
}

func TestAsyncResolve(t *testing.T) {
	res, err := createResolution("stepCondition.yaml", map[string]interface{}{}, nil)

	assert.NotNil(t, res)
	assert.Nil(t, err)
	assert.Equal(t, step.StateTODO, res.Steps["stepOne"].State)
	assert.Equal(t, resolution.StateTODO, res.State)
	assert.Equal(t, 0, res.Steps["stepOne"].TryCount)

	err = engine.GetEngine().Resolve(res.PublicID, nil)

	assert.Nil(t, err)
}

func TestInputNumber(t *testing.T) {
	input := map[string]interface{}{
		"quantity": -2.3,
	}
	res, err := createResolution("input.yaml", input, nil)
	assert.NotNil(t, res)
	assert.Nil(t, err)

	res, err = runResolution(res)

	assert.Nil(t, err)
	assert.NotNil(t, res)

	output := res.Steps["stepOne"].Output.(map[string]interface{})
	assert.Equal(t, "-2.3", output["value"])
}

func TestAnyDependency(t *testing.T) {
	res, err := createResolution("anyDependency.yaml", map[string]interface{}{}, nil)
	assert.NotNil(t, res)
	assert.Nil(t, err)

	res, err = runResolution(res)
	assert.NotNil(t, res)
	assert.Nil(t, err)

	output := res.Steps["secondStepRunsIfAny"].Output.(map[string]interface{})
	assert.Equal(t, "yes", output["i_ran_anyway"])

	thirdOKState := res.Steps["thirdStepRunsIfSecondOK"].State
	assert.Equal(t, step.StateDone, thirdOKState)

	thirdPrunedState := res.Steps["thirdStepRunsIfSecondKO"].State
	assert.Equal(t, step.StatePrune, thirdPrunedState)

	finalOKState := res.Steps["fourthStepWillNotBePruned"].State
	assert.Equal(t, step.StateDone, finalOKState)
}

func TestIndirectDependencies(t *testing.T) {
	res, err := createResolution("indirectDependencies.yaml", map[string]interface{}{}, nil)
	assert.NotNil(t, res)
	assert.Nil(t, err)

	res, err = runResolution(res)
	assert.NotNil(t, res)
	assert.Nil(t, err)

	assert.Equal(t, step.StateDone, res.Steps["stepOne"].State)
	assert.Equal(t, step.StateFatalError, res.Steps["stepTwo"].State)
	assert.Equal(t, step.StatePrune, res.Steps["stepThree"].State)
	assert.Equal(t, step.StateTODO, res.Steps["stepFour"].State)
}

func TestMetadata(t *testing.T) {
	res, err := createResolution("metadata.yaml", map[string]interface{}{}, nil)
	assert.NotNil(t, res)
	assert.Nil(t, err)

	res, err = runResolution(res)
	assert.NotNil(t, res)
	assert.Nil(t, err)
	notfoundState := res.Steps["notfound"].State
	assert.Equal(t, "NOTFOUND", notfoundState)
}

func TestRetryNow(t *testing.T) {
	expectedResult := "0 sep 1 sep 1 sep 2 sep 3 sep 5 sep 8 sep 13 sep 21 sep 34 sep 55 sep 89 sep 144"
	res, err := createResolution("retryNowState.yaml", map[string]interface{}{
		"N":         12.0,
		"separator": " sep ",
	}, nil)
	assert.Nil(t, err)
	assert.NotNil(t, res)

	res, err = runResolution(res)
	assert.NotNil(t, res)
	assert.Nil(t, err)
	assert.Equal(t, resolution.StateDone, res.State)

	assert.Equal(t, step.StateDone, res.Steps["fibonacci"].State)
	assert.Equal(t, step.StateDone, res.Steps["join"].State)

	output := res.Steps["join"].Output.(map[string]interface{})
	assert.Equal(t, expectedResult, output["str"])
}

func TestRetryNowMaxRetry(t *testing.T) {
	expected := "42"
	res, err := createResolution("retryNowMaxRetry.yaml", map[string]interface{}{}, nil)
	assert.Nil(t, err)
	assert.NotNil(t, res)

	res, err = runResolution(res)
	assert.NotNil(t, res)
	assert.Nil(t, err)
	assert.Equal(t, resolution.StateBlockedFatal, res.State)

	assert.Equal(t, step.StateFatalError, res.Steps["infinite"].State)
	assert.Equal(t, res.Steps["infinite"].TryCount, res.Steps["infinite"].MaxRetries+1)

	assert.Equal(t, expected, res.Steps["infinite"].Output)
}

func TestForeach(t *testing.T) {
	res, err := createResolution("foreach.yaml", map[string]interface{}{
		"list": []interface{}{"a", "b", "c"},
	}, nil)
	assert.Nil(t, err)
	assert.NotNil(t, res)

	res, err = runResolution(res)
	assert.NotNil(t, res)
	assert.Nil(t, err)
	assert.Equal(t, resolution.StateDone, res.State)

	assert.Equal(t, step.StateDone, res.Steps["emptyLoop"].State) // running on empty collection is ok
	assert.Equal(t, step.StateDone, res.Steps["concatItems"].State)
	assert.Equal(t, step.StateDone, res.Steps["finalStep"].State)
	assert.Equal(t, "B", res.Steps["bStep"].State)

	generateList := res.Steps["generateItems"].Children
	assert.Equal(t, 2, len(generateList))

	concatList := res.Steps["concatItems"].Children
	require.Equal(t, 1, len(concatList))

	firstItem := concatList[0].(map[string]interface{})
	firstItemOutput := firstItem[values.OutputKey].(map[string]interface{})
	assert.Equal(t, "foo-b-bar-b", firstItemOutput["concat"])

	outputExpected := map[string]string{"foo": "foo-b", "bar": "bar-b"}
	metadata, ok := firstItem[values.MetadataKey].(map[string]interface{})
	require.True(t, ok)
	iterator, ok := metadata[values.IteratorKey].(map[string]interface{})
	require.True(t, ok)
	outputInterface, ok := iterator["output"].(map[string]interface{})
	require.True(t, ok)
	output := make(map[string]string)
	for key, value := range outputInterface {
		output[key] = fmt.Sprintf("%v", value)
	}
	assert.Equal(t, outputExpected, output)
}

func TestForeachWithChainedIterations(t *testing.T) {
	assert, require := td.AssertRequire(t)
	res, err := createResolution("foreach.yaml", map[string]interface{}{
		"list": []interface{}{"a", "b", "c", "d", "e"},
	}, nil)
	require.Nil(err)
	require.NotNil(res)

	res.Steps["generateItems"].Conditions[0].Then["this"] = "DONE"
	res.Steps["generateItems"].Conditions = append(
		res.Steps["generateItems"].Conditions,
		&condition.Condition{
			Type: condition.CHECK,
			If: []*condition.Assert{
				{
					Value:    "{{.iterator}}",
					Operator: condition.EQ,
					Expected: "d",
				},
			},
			Then: map[string]string{
				"this": "SERVER_ERROR",
			},
			ForEach: condition.ForEachChildren,
		},
	)
	res.Steps["generateItems"].ForEachStrategy = "sequence"
	err = updateResolution(res)
	require.Nil(err)

	res, err = runResolution(res)
	require.NotNil(res)
	require.Nil(err)
	require.Cmp(res.State, resolution.StateError)

	assert.Cmp(res.Steps["emptyLoop"].State, step.StateDone) // running on empty collection is ok
	assert.Cmp(res.Steps["concatItems"].State, step.StateTODO)
	assert.Cmp(res.Steps["finalStep"].State, step.StateTODO)
	assert.Cmp(res.Steps["bStep"].State, "B")
	assert.Cmp(res.Steps["generateItems-0"].State, step.StateDone)
	assert.Cmp(res.Steps["generateItems-1"].State, step.StateDone)
	assert.Cmp(res.Steps["generateItems-2"].State, step.StateDone)
	assert.Cmp(res.Steps["generateItems-3"].State, step.StateServerError)
	assert.Cmp(res.Steps["generateItems-4"].State, step.StateTODO)
	assert.Len(res.Steps["generateItems-0"].Dependencies, 0)
	assert.Cmp(res.Steps["generateItems-1"].Dependencies, []string{"generateItems-0"})
	assert.Cmp(res.Steps["generateItems-2"].Dependencies, []string{"generateItems-1"})
	assert.Cmp(res.Steps["generateItems-3"].Dependencies, []string{"generateItems-2"})
	assert.Cmp(res.Steps["generateItems-4"].Dependencies, []string{"generateItems-3"})
}

func TestForeachWithChainedIterationsWithDepOnParent(t *testing.T) {
	assert, require := td.AssertRequire(t)
	res, err := createResolution("foreach.yaml", map[string]interface{}{
		"list": []interface{}{"a", "b", "c", "d", "e"},
	}, nil)
	require.Nil(err)
	require.NotNil(res)

	res.Steps["generateItems"].Dependencies = []string{"emptyLoop"}
	res.Steps["generateItems"].Conditions[0].Then["this"] = "DONE"
	res.Steps["generateItems"].Conditions = append(
		res.Steps["generateItems"].Conditions,
		&condition.Condition{
			Type: condition.CHECK,
			If: []*condition.Assert{
				{
					Value:    "{{.iterator}}",
					Operator: condition.EQ,
					Expected: "d",
				},
			},
			Then: map[string]string{
				"this": "SERVER_ERROR",
			},
			ForEach: condition.ForEachChildren,
		},
	)
	res.Steps["generateItems"].ForEachStrategy = "sequence"
	err = updateResolution(res)
	require.Nil(err)

	res, err = runResolution(res)
	require.NotNil(res)
	require.Nil(err)
	require.Cmp(res.State, resolution.StateError)

	assert.Cmp(res.Steps["emptyLoop"].State, step.StateDone) // running on empty collection is ok
	assert.Cmp(res.Steps["concatItems"].State, step.StateTODO)
	assert.Cmp(res.Steps["finalStep"].State, step.StateTODO)
	assert.Cmp(res.Steps["bStep"].State, "B")
	assert.Cmp(res.Steps["generateItems-0"].State, step.StateDone)
	assert.Cmp(res.Steps["generateItems-1"].State, step.StateDone)
	assert.Cmp(res.Steps["generateItems-2"].State, step.StateDone)
	assert.Cmp(res.Steps["generateItems-3"].State, step.StateServerError)
	assert.Cmp(res.Steps["generateItems-4"].State, step.StateTODO)
	assert.Cmp(res.Steps["generateItems"].Dependencies, []string{"emptyLoop", "generateItems-0:ANY", "generateItems-1:ANY", "generateItems-2:ANY", "generateItems-3:ANY", "generateItems-4:ANY"})
	assert.Cmp(res.Steps["generateItems-0"].Dependencies, []string{"emptyLoop"})
	assert.Cmp(res.Steps["generateItems-1"].Dependencies, []string{"emptyLoop", "generateItems-0"})
	assert.Cmp(res.Steps["generateItems-2"].Dependencies, []string{"emptyLoop", "generateItems-1"})
	assert.Cmp(res.Steps["generateItems-3"].Dependencies, []string{"emptyLoop", "generateItems-2"})
	assert.Cmp(res.Steps["generateItems-4"].Dependencies, []string{"emptyLoop", "generateItems-3"})
}

func TestForeachWithPreRun(t *testing.T) {
	for _, switchToToRetry := range []bool{false, true} {
		t.Run(fmt.Sprintf("%s-%t", t.Name(), switchToToRetry), func(t *testing.T) {
			input := map[string]interface{}{}
			res, err := createResolution("foreachAndPreRun.yaml", input, nil)
			require.Nilf(t, err, "expecting nil error, got %s", err)
			require.NotNil(t, res)

			if switchToToRetry {
				for _, st := range []string{"stepForeachPrune", "stepDepOnForeachPrune", "stepForeachPruneParentTask", "stepDepOnForeachPruneParentTask"} {
					res.Steps[st].State = step.StateToRetry
				}
				require.NoError(t, updateResolution(res))
			}

			res, err = runResolution(res)

			require.Nilf(t, err, "got error %s", err)
			require.NotNil(t, res)
			assert.Equal(t, resolution.StateDone, res.State)
			for _, st := range []string{"stepForeachNoDep", "stepSkippedNoDep", "stepNoDep", "stepForeachWithDep", "stepSkippedWithDep"} {
				assert.Equal(t, step.StateDone, res.Steps[st].State)
			}
			for _, st := range []string{"stepDep", "stepDep2"} {
				assert.Equal(t, step.StatePrune, res.Steps[st].State)
			}

			// skip prune on a foreach step's children means:
			// - foreach children are set to prune
			// - the foreach step itself is set to done
			// - the dependencies are not pruned
			assert.Equal(t, step.StateDone, res.Steps["stepForeachPrune"].State)
			assert.Equal(t, step.StateDone, res.Steps["stepDepOnForeachPrune"].State)

			// skip prune on a foreach step itself means:
			// - foreach children are not generated
			// - the foreach step itself is set to prune
			// - the dependencies are pruned
			assert.Equal(t, step.StatePrune, res.Steps["stepForeachPruneParentTask"].State)
			assert.Equal(t, step.StatePrune, res.Steps["stepDepOnForeachPruneParentTask"].State)
		})
	}
}

func TestForeachWithErrors(t *testing.T) {
	res, err := createResolution("foreach.yaml", map[string]interface{}{
		"list": []interface{}{"a", "b", "c"},
	}, nil)
	assert.Nil(t, err)
	assert.NotNil(t, res)

	res.Steps["generateItems"].State = step.StateFatalError
	updateResolution(res)

	res, err = runResolution(res)
	assert.NotNil(t, res)
	assert.Nil(t, err)
	assert.Equal(t, resolution.StateBlockedFatal, res.State)
}

func TestVariables(t *testing.T) {
	res, err := createResolution("variables.yaml", map[string]interface{}{}, nil)
	assert.NotNil(t, res)
	assert.Nil(t, err)

	res, err = runResolution(res)
	assert.NotNil(t, res)
	assert.Nil(t, err)

	output := res.Steps["renderVariables"].Output.(map[string]interface{})
	assert.Equal(t, "4", output["truc"])
	assert.Equal(t, "5", output["bidule"])
	assert.Equal(t, "Hello World!", output["templated"])
	assert.Equal(t, "6", output["cached"])
	output = res.Steps["renderVariablesWithCache"].Output.(map[string]interface{})
	assert.Equal(t, "6", output["cached"])
}

const (
	singleString    = "hello"
	multilineString = `Un,
	Deux,
	Trois,
	Soleil!`
)

func TestJSONTemplating(t *testing.T) {
	res, err := createResolution("jsonTemplating.yaml", map[string]interface{}{
		"singleString":    singleString,
		"multilineString": multilineString,
	}, nil)
	assert.NotNil(t, res)
	assert.Nil(t, err)

	res, err = runResolution(res)
	assert.NotNil(t, res)
	assert.Nil(t, err)
	assert.Equal(t, resolution.StateDone, res.State)

	output := res.Steps["stepOne"].Output.(map[string]interface{})
	assert.Equal(t, multilineString, output["raw-multiline"])
	assert.Equal(t, singleString, output["raw-single"])

	jsonBody := output["my-json-body"].(string)
	body := map[string]interface{}{}
	err = json.Unmarshal([]byte(jsonBody), &body)
	assert.Nil(t, err)
}

func TestJSONNumberTemplating(t *testing.T) {
	res, err := createResolution("jsonnumber.yaml", nil, nil)
	assert.NotNil(t, res)
	assert.Nil(t, err)

	res, err = runResolution(res)
	assert.NotNil(t, res)
	assert.Nil(t, err)
	assert.Equal(t, resolution.StateDone, res.State)

	output := res.Steps["loopStep"].Children
	require.Greater(t, len(output), 0)
	child := output[0].(map[string]interface{})
	assert.Equal(t, "/id/1619464078", child[values.OutputKey].(string))
}

func TestJSONParsing(t *testing.T) {
	res, err := createResolution("jsonParsing.yaml", nil, nil)
	assert.NotNil(t, res)
	assert.Nil(t, err)

	res, err = runResolution(res)
	assert.NotNil(t, res)
	assert.Nil(t, err)
	assert.Equal(t, resolution.StateDone, res.State)

	output := res.Steps["stepOne"].Output.(map[string]interface{})
	assert.Equal(t, "utask", output["a"])
	assert.Equal(t, "666", output["b"])
	assert.Equal(t, "map[k:v]", output["c"])
	assert.Equal(t, "[1 2 3]", output["d"])
}

func TestRetryLoop(t *testing.T) {
	res, err := createResolution("retryloop.yaml", nil, nil)
	assert.NotNil(t, res)
	assert.Nil(t, err)

	res, err = runResolution(res)
	assert.NotNil(t, res)
	assert.Nil(t, err)

	assert.Equal(t, resolution.StateError, res.State)
	assert.Equal(t, step.StateToRetry, res.Steps["generateItems"].State)
	assert.Nil(t, res.Steps["generateItems"].ChildrenSteps)

	// artificially remove the condition that sets the loop step in RETRY state
	res.Steps["generateItems"].Conditions = nil
	err = updateResolution(res)
	assert.Nil(t, err)

	// successfully run resolution
	res, err = runResolution(res)
	assert.NotNil(t, res)
	assert.Nil(t, err)
	assert.Equal(t, resolution.StateDone, res.State)
	assert.Nil(t, res.Steps["generateItems"].ChildrenSteps)

	finalOutput := res.Steps["generateItems"].Children
	assert.Equal(t, 3, len(finalOutput))

	firstItem := finalOutput[0].(map[string]interface{})
	firstItemOutput := firstItem[values.OutputKey].(map[string]interface{})
	assert.Equal(t, "foo-a", firstItemOutput["foo"])
}

func TestBaseOutput(t *testing.T) {
	id := "1234"
	res, err := createResolution("base_output.yaml", map[string]interface{}{"id": id}, nil)
	assert.NotNil(t, res)
	assert.Nil(t, err)

	res, err = runResolution(res)
	assert.NotNil(t, res)
	assert.Nil(t, err)

	output := res.Steps["stepOne"].Output.(map[string]interface{})
	assert.Equal(t, id, output["id"])
	assert.Equal(t, "bar", output["foo"])
}

func TestEmptyStringInput(t *testing.T) {
	input := map[string]interface{}{
		"quantity": -2.3,
		"foo":      "",
	}
	res, err := createResolution("input.yaml", input, nil)
	assert.NotNil(t, res)
	assert.Nil(t, err)

	res, err = runResolution(res)

	require.Nilf(t, err, "got error %s", err)
	require.NotNil(t, res)
	assert.Equal(t, resolution.StateDone, res.State)
	assert.Equal(t, step.StateDone, res.Steps["stepOne"].State)

	output := res.Steps["stepOne"].Output.(map[string]interface{})
	assert.Equal(t, "", output["foo"])
}

func TestBaseOutputNoOutput(t *testing.T) {
	input := map[string]interface{}{}
	res, err := createResolution("no-output.yaml", input, nil)
	require.NotNil(t, res)
	require.Nil(t, err)

	res, err = runResolution(res)

	require.Nilf(t, err, "got error %s", err)
	require.NotNil(t, res)
	assert.Equal(t, resolution.StateDone, res.State)
	assert.Equal(t, step.StateDone, res.Steps["stepOne"].State)

	output := res.Steps["stepOne"].Output.(map[string]interface{})
	assert.Equal(t, "buzz", output["foobar"])
}

func TestOutputTemplatingError(t *testing.T) {
	input := map[string]interface{}{}
	res, err := createResolution("no-output.yaml", input, nil)
	require.NotNil(t, res)
	require.Nil(t, err)

	res.Steps["stepOne"].Action.Output.Strategy = executor.OutputStrategytemplate
	res.Steps["stepOne"].Action.Output.Format = map[string]string{
		"foo2": "{{ index .foo.unknown 3 }}",
	}
	err = updateResolution(res)
	require.Nil(t, err)

	res, err = runResolution(res)

	require.Nilf(t, err, "got error %s", err)
	require.NotNil(t, res)
	assert.Equal(t, resolution.StateBlockedFatal, res.State)
	assert.Equal(t, step.StateFatalError, res.Steps["stepOne"].State)
	assert.Contains(t, res.Steps["stepOne"].Error, "unable to format output: Templating error: template:")

	res, err = createResolution("no-output.yaml", input, nil)
	require.NotNil(t, res)
	require.Nil(t, err)

	res.Steps["stepOne"].Action.Output.Strategy = executor.OutputStrategymerge
	res.Steps["stepOne"].Action.Output.Format = map[string]string{
		"foo2": "{{ index .foo.unknown 3 }}",
	}
	err = updateResolution(res)
	require.Nil(t, err)

	res, err = runResolution(res)

	require.Nilf(t, err, "got error %s", err)
	require.NotNil(t, res)
	assert.Equal(t, resolution.StateBlockedFatal, res.State)
	assert.Equal(t, step.StateFatalError, res.Steps["stepOne"].State)
	assert.Contains(t, res.Steps["stepOne"].Error, "failed to template base output: Templating error: template:")

	res, err = createResolution("no-output.yaml", input, nil)
	require.NotNil(t, res)
	require.Nil(t, err)

	res.Steps["stepOne"].Action.Output.Strategy = executor.OutputStrategytemplate
	res.Steps["stepOne"].Action.Output.Format = map[string]string{
		"foo2": "{{ index .step.stepOne.output.foo.unknown 3 }}",
	}
	err = updateResolution(res)
	require.Nil(t, err)

	res, err = runResolution(res)

	require.Nilf(t, err, "got error %s", err)
	require.NotNil(t, res)
	assert.Equal(t, resolution.StateBlockedFatal, res.State)
	assert.Equal(t, step.StateFatalError, res.Steps["stepOne"].State)
	assert.Contains(t, res.Steps["stepOne"].Error, "unable to format output: Templating error: template:")

	res, err = createResolution("no-output.yaml", input, nil)
	require.NotNil(t, res)
	require.Nil(t, err)

	res.Steps["stepOne"].Action.Output.Strategy = executor.OutputStrategymerge
	res.Steps["stepOne"].Action.Output.Format = map[string]string{
		"foo2": "{{ index .step.stepOne.output.foo.unknown 3 }}",
	}
	err = updateResolution(res)
	require.Nil(t, err)

	res, err = runResolution(res)

	require.Nilf(t, err, "got error %s", err)
	require.NotNil(t, res)
	assert.Equal(t, resolution.StateBlockedFatal, res.State)
	assert.Equal(t, step.StateFatalError, res.Steps["stepOne"].State)
	assert.Contains(t, res.Steps["stepOne"].Error, "failed to template base output: Templating error: template:")
}

func TestBaseOutputNoOutputBackwardCompatibility(t *testing.T) {
	input := map[string]interface{}{}
	res, err := createResolution("no-output-backward.yaml", input, nil)
	require.NotNil(t, res)
	require.Nil(t, err)

	res, err = runResolution(res)

	require.Nilf(t, err, "got error %s", err)
	require.NotNil(t, res)
	assert.Equal(t, resolution.StateDone, res.State)
	assert.Equal(t, step.StateDone, res.Steps["stepOne"].State)

	output := res.Steps["stepOne"].Output.(map[string]interface{})
	assert.Equal(t, "buzz", output["foobar"])
}

func TestScriptPlugin(t *testing.T) {
	argv := "world"
	res, err := createResolution("execScript.yaml", map[string]interface{}{"argv": argv}, nil)
	assert.NotNil(t, res)
	assert.Nil(t, err)

	res, err = runResolution(res)
	assert.NotNil(t, res)
	assert.Nil(t, err)

	output := make(map[string]interface{})
	output["dumb_string"] = fmt.Sprintf("Hello %s!", argv)
	output["random_object"] = map[string]interface{}{"foo": "bar"}

	metadata := map[string]interface{}{
		"exit_code":      "0",
		"process_state":  "exit status 0",
		"output":         "Hello world script\n{\"dumb_string\":\"Hello world!\",\"random_object\":{\"foo\":\"bar\"}}\n",
		"execution_time": "",
		"error":          "",
	}

	// because time can't be consistant through tests
	metadataOutput := res.Steps["stepOne"].Metadata.(map[string]interface{})
	metadataOutput["execution_time"] = ""

	assert.Equal(t, output, res.Steps["stepOne"].Output)
	assert.Equal(t, metadata, metadataOutput)
}

func TestScriptPluginEnvironmentVariables(t *testing.T) {
	res, err := createResolution("execScriptWithEnvironment.yaml", map[string]interface{}{}, nil)
	assert.NotNil(t, res)
	assert.Nil(t, err)

	res, err = runResolution(res)
	assert.NotNil(t, res)
	assert.Nil(t, err)

	assert.Equal(t, step.StateDone, res.State)
	assert.Equal(t, step.StateDone, res.Steps["stepOne"].State)
	t.Log(res.Steps["stepOne"].Error)
	t.Log(res.Steps["stepOne"].Metadata)

	assert.NotNil(t, res.Steps["stepOne"].Output)

	environment := res.Steps["stepOne"].Output.(map[string]interface{})

	assert.NotNil(t, environment["UTASK_TASK_ID"])
	assert.Equal(t, res.TaskPublicID, environment["UTASK_TASK_ID"])
	assert.NotNil(t, environment["UTASK_RESOLUTION_ID"])
	assert.Equal(t, res.PublicID, environment["UTASK_RESOLUTION_ID"])
	assert.NotNil(t, environment["UTASK_STEP_NAME"])
	assert.Equal(t, "stepOne", environment["UTASK_STEP_NAME"])
	assert.NotNil(t, environment["static_value"])
	assert.Equal(t, "foo", environment["static_value"])
	assert.NotNil(t, environment["variable_value"])
	assert.Equal(t, "bar", environment["variable_value"])
}

func TestBaseBaseConfiguration(t *testing.T) {
	res, err := createResolution("base_configuration.yaml", nil, nil)
	assert.NotNil(t, res)
	assert.Nil(t, err)

	res, err = runResolution(res)
	assert.NotNil(t, res)
	assert.Nil(t, err)

	assert.Equal(t, "testingcfg", res.Steps["stepOne"].Action.BaseConfiguration)
	assert.Equal(t, "testingcfg", res.Steps["stepTwo"].Action.BaseConfiguration)

	outputOne := res.Steps["stepOne"].Output.(string)
	outputTwo := res.Steps["stepTwo"].Output.(string)
	bCfg := echo.Config{}
	assert.Nil(t, json.Unmarshal(res.BaseConfigurations["testingcfg"], &bCfg))

	assert.Equal(t, outputOne, outputTwo)
	assert.Equal(t, bCfg.Output, outputTwo)
	assert.Equal(t, bCfg.Output, outputOne)

	assert.NotEqual(t, res.Steps["stepOne"].Error, res.Steps["stepTwo"].Error)
}

func TestResolveSubTask(t *testing.T) {
	dbp, err := zesty.NewDBProvider(utask.DBName)
	require.Nil(t, err)

	_, err = templateFromYAML(dbp, "variables.yaml")
	require.Nil(t, err)

	res, err := createResolution("subtask.yaml", map[string]interface{}{}, nil)
	require.Nil(t, err, "failed to create resolution: %s", err)

	res, err = runResolution(res)
	require.Nil(t, err)
	require.NotNil(t, res)
	assert.Equal(t, resolution.StateWaiting, res.State)

	nextRetryBeforeRun := time.Time{}
	if res.NextRetry != nil {
		nextRetryBeforeRun = *res.NextRetry
	}

	for _, subtaskName := range []string{"subtaskCreation", "jsonInputSubtask", "templatingJsonInputSubtask"} {
		subtaskCreationOutput := res.Steps[subtaskName].Output.(map[string]interface{})
		subtaskPublicID := subtaskCreationOutput["id"].(string)

		subtask, err := task.LoadFromPublicID(dbp, subtaskPublicID)
		require.Nil(t, err)
		assert.Equal(t, task.StateTODO, subtask.State)

		subtaskResolution, err := resolution.Create(dbp, subtask, nil, "", false, nil)
		require.Nil(t, err)

		subtaskResolution, err = runResolution(subtaskResolution)
		require.Nil(t, err)
		assert.Equal(t, task.StateDone, subtaskResolution.State)
		for k, v := range subtaskResolution.Steps {
			assert.Equal(t, step.StateDone, v.State, "not valid state for step %s", k)
		}

		subtask, err = task.LoadFromPublicID(dbp, subtaskPublicID)
		require.Nil(t, err)
		assert.Equal(t, task.StateDone, subtask.State)
		parentTaskToResume, err := taskutils.ShouldResumeParentTask(dbp, subtask)
		require.Nil(t, err)
		require.NotNil(t, parentTaskToResume)
		assert.Equal(t, res.TaskID, parentTaskToResume.ID)
	}

	// checking whether the parent task will be picked up by the RetryCollector after the subtask is resolved.
	res, err = resolution.LoadFromPublicID(dbp, res.PublicID)
	require.Nil(t, err)
	assert.NotNil(t, res.NextRetry)
	assert.False(t, res.NextRetry.IsZero())
	assert.True(t, res.NextRetry.After(nextRetryBeforeRun))
	assert.True(t, res.NextRetry.Before(time.Now()))

	// Starting the RetryCollector to resume the parent task
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	engine.RetryCollector(ctx)

	ti := time.Second
	i := time.Duration(0)
	for i < ti {
		res, err = resolution.LoadFromPublicID(dbp, res.PublicID)
		require.Nil(t, err)
		if res.State == resolution.StateDone {
			break
		}

		time.Sleep(time.Millisecond * 10)
		i += time.Millisecond * 10

	}
	assert.Equal(t, resolution.StateDone, res.State)
}

func TestResolveSubTaskParentTaskPaused(t *testing.T) {
	dbp, err := zesty.NewDBProvider(utask.DBName)
	require.Nil(t, err)

	_, err = templateFromYAML(dbp, "variables.yaml")
	require.Nil(t, err)

	res, err := createResolution("subtask.yaml", map[string]interface{}{}, nil)
	require.Nil(t, err, "failed to create resolution: %s", err)

	res, err = runResolution(res)
	require.Nil(t, err)
	require.NotNil(t, res)
	assert.Equal(t, resolution.StateWaiting, res.State)

	subtaskCreationOutput := res.Steps["subtaskCreation"].Output.(map[string]interface{})
	subtaskPublicID := subtaskCreationOutput["id"].(string)

	// pausing parent task
	res.SetState(resolution.StatePaused)
	res.Update(dbp)

	subtask, err := task.LoadFromPublicID(dbp, subtaskPublicID)
	require.Nil(t, err)
	assert.Equal(t, task.StateTODO, subtask.State)

	subtaskResolution, err := resolution.Create(dbp, subtask, nil, "", false, nil)
	require.Nil(t, err)

	subtaskResolution, err = runResolution(subtaskResolution)
	require.Nil(t, err)
	assert.Equal(t, task.StateDone, subtaskResolution.State)
	for k, v := range subtaskResolution.Steps {
		assert.Equal(t, step.StateDone, v.State, "not valid state for step %s", k)
	}

	subtask, err = task.LoadFromPublicID(dbp, subtaskPublicID)
	require.Nil(t, err)
	assert.Equal(t, task.StateDone, subtask.State)
	parentTaskToResume, err := taskutils.ShouldResumeParentTask(dbp, subtask)
	require.Nil(t, parentTaskToResume)
	require.Nil(t, err)
}

func TestResolveCallback(t *testing.T) {
	res, err := createResolution("callback.yaml", map[string]interface{}{}, nil)
	require.NoError(t, err)

	res, err = runResolution(res)
	require.NoError(t, err)
	require.NotNil(t, res)

	// check steps state
	assert.Equal(t, res.Steps["createCallback"].State, step.StateDone)
	assert.Equal(t, res.Steps["waitCallback"].State, step.StateWaiting)

	// callback has been created, waiting for its resolution
	assert.Equal(t, resolution.StateWaiting, res.State)
}

func TestB64RawEncodeDecode(t *testing.T) {
	res, err := createResolution("rawb64EncodingDecoding.yaml", nil, nil)
	assert.NotNil(t, res)
	assert.Nil(t, err)

	res, err = runResolution(res)
	assert.NotNil(t, res)
	assert.Nil(t, err)
	assert.Equal(t, resolution.StateDone, res.State)

	output := res.Steps["stepOne"].Output.(map[string]interface{})
	assert.Equal(t, "cmF3IG1lc3NhZ2U", output["a"])
	assert.Equal(t, "raw message", output["b"])
}

func TestBatch(t *testing.T) {
	dbp, err := zesty.NewDBProvider(utask.DBName)
	require.Nil(t, err)

	_, err = templateFromYAML(dbp, "batchedTask.yaml")
	require.Nil(t, err)

	_, err = templateFromYAML(dbp, "batch.yaml")
	require.Nil(t, err)

	res, err := createResolution("batch.yaml", map[string]interface{}{}, nil)
	require.Nil(t, err, "failed to create resolution: %s", err)

	res, err = runResolution(res)
	require.Nil(t, err)
	require.NotNil(t, res)
	assert.Equal(t, resolution.StateWaiting, res.State)

	nextRetryBeforeRun := time.Time{}
	if res.NextRetry != nil {
		nextRetryBeforeRun = *res.NextRetry
	}

	for _, batchStepName := range []string{"batchJsonInputs", "batchYamlInputs"} {
		batchStepMetadataRaw, ok := res.Steps[batchStepName].Metadata.(string)
		assert.True(t, ok, "wrong type of metadata for step '%s'", batchStepName)

		assert.Nil(t, res.Steps[batchStepName].Output, "output nil for step '%s'", batchStepName)

		// The plugin formats Metadata in a special way that we need to revert before unmarshalling them
		batchStepMetadataRaw = strings.ReplaceAll(batchStepMetadataRaw, `\"`, `"`)
		var batchStepMetadata map[string]any
		err := json.Unmarshal([]byte(batchStepMetadataRaw), &batchStepMetadata)
		require.Nil(t, err, "metadata unmarshalling of step '%s'", batchStepName)

		batchPublicID := batchStepMetadata["batch_id"].(string)
		assert.NotEqual(t, "", batchPublicID, "wrong batch ID '%s'", batchPublicID)

		b, err := task.LoadBatchFromPublicID(dbp, batchPublicID)
		require.Nil(t, err)

		taskIDs, err := listBatchTasks(dbp, b.ID)
		require.Nil(t, err)
		assert.Len(t, taskIDs, 2)

		for i, publicID := range taskIDs {
			child, err := task.LoadFromPublicID(dbp, publicID)
			require.Nil(t, err)
			assert.Equal(t, task.StateTODO, child.State)

			childResolution, err := resolution.Create(dbp, child, nil, "", false, nil)
			require.Nil(t, err)

			childResolution, err = runResolution(childResolution)
			require.Nil(t, err)
			assert.Equal(t, resolution.StateDone, childResolution.State)

			for k, v := range childResolution.Steps {
				assert.Equal(t, step.StateDone, v.State, "not valid state for step %s", k)
			}

			child, err = task.LoadFromPublicID(dbp, child.PublicID)
			require.Nil(t, err)
			assert.Equal(t, task.StateDone, child.State)

			parentTaskToResume, err := taskutils.ShouldResumeParentTask(dbp, child)
			require.Nil(t, err)
			if i == len(taskIDs)-1 {
				// Only the last child task should resume the parent
				require.NotNil(t, parentTaskToResume)
				assert.Equal(t, res.TaskID, parentTaskToResume.ID)
			} else {
				require.Nil(t, parentTaskToResume)
			}
		}
	}

	// checking whether the parent task will be picked up by the RetryCollector after the subtask is resolved.
	res, err = resolution.LoadFromPublicID(dbp, res.PublicID)
	require.Nil(t, err)
	assert.NotNil(t, res.NextRetry)
	assert.False(t, res.NextRetry.IsZero())
	assert.True(t, res.NextRetry.After(nextRetryBeforeRun))
	assert.True(t, res.NextRetry.Before(time.Now()))

	// Starting the RetryCollector to resume the parent task
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	engine.RetryCollector(ctx)

	ti := time.Second
	i := time.Duration(0)
	for i < ti {
		res, err = resolution.LoadFromPublicID(dbp, res.PublicID)
		require.Nil(t, err)
		if res.State == resolution.StateDone {
			break
		}

		time.Sleep(time.Millisecond * 10)
		i += time.Millisecond * 10
	}
	assert.Equal(t, resolution.StateDone, res.State)
}

func TestWakeParent(t *testing.T) {
	dbp, err := zesty.NewDBProvider(utask.DBName)
	require.Nil(t, err)

	// Create a task that spawns two simple subtasks
	_, err = templateFromYAML(dbp, "no-output.yaml")
	require.Nil(t, err)

	res, err := createResolution("noOutputSubtask.yaml", map[string]interface{}{}, nil)
	require.Nil(t, err, "failed to create resolution: %s", err)

	res, err = runResolution(res)
	require.Nil(t, err)
	require.NotNil(t, res)
	require.Equal(t, resolution.StateWaiting, res.State)
	assert.True(t, res.NextRetry.IsZero())

	// Force the parent task to the RUNNING state
	res.SetState(resolution.StateRunning)
	res.Update(dbp)

	// Create and run one of the subtasks
	subtaskCreationOutput := res.Steps["subtaskCreation"].Output.(map[string]interface{})
	subtaskPublicID := subtaskCreationOutput["id"].(string)

	subtask, err := task.LoadFromPublicID(dbp, subtaskPublicID)
	require.Nil(t, err)
	require.Equal(t, task.StateTODO, subtask.State)

	subtaskResolution, err := resolution.Create(dbp, subtask, nil, "", false, nil)
	require.Nil(t, err)

	beforeRun := time.Now()
	subtaskResolution, err = runResolution(subtaskResolution)
	require.Nil(t, err)
	require.Equal(t, task.StateDone, subtaskResolution.State)
	afterRun := time.Now()

	// Refreshing parent resolution to check its next_retry value
	res, err = resolution.LoadFromPublicID(dbp, res.PublicID)
	require.Nil(t, err)
	assert.Equal(t, res.State, resolution.StateRunning) // Parent should still be RUNNING

	// The parent's next_retry should have been updated so that the RetryCollector would pick it up
	assert.NotNil(t, res.NextRetry)
	assert.True(t, res.NextRetry.After(beforeRun))
	assert.True(t, res.NextRetry.Before(afterRun))
}

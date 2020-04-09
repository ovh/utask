package engine_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ghodss/yaml"
	"github.com/juju/errors"
	"github.com/loopfz/gadgeto/zesty"
	"github.com/ovh/configstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ovh/utask"
	"github.com/ovh/utask/db"
	"github.com/ovh/utask/db/pgjuju"
	"github.com/ovh/utask/engine"
	"github.com/ovh/utask/engine/step"
	"github.com/ovh/utask/engine/values"
	"github.com/ovh/utask/models/resolution"
	"github.com/ovh/utask/models/task"
	"github.com/ovh/utask/models/tasktemplate"
	"github.com/ovh/utask/pkg/now"
	"github.com/ovh/utask/pkg/plugins/builtin/echo"
	"github.com/ovh/utask/pkg/plugins/builtin/script"
)

const (
	testDir = "./templates_tests"
)

var (
	templateList = loadTemplates()
)

func TestMain(m *testing.M) {
	store := configstore.DefaultStore
	store.InitFromEnvironment()

	db.Init(store)

	now.Init()

	if err := engine.Init(context.Background(), store); err != nil {
		panic(err)
	}

	step.RegisterRunner(echo.Plugin.PluginName(), echo.Plugin)
	step.RegisterRunner(script.Plugin.PluginName(), script.Plugin)

	os.Exit(m.Run())
}

func loadTemplates() map[string][]byte {
	templateList := map[string][]byte{}
	files, err := ioutil.ReadDir(testDir)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		if file.Mode().IsRegular() {
			bytes, err := ioutil.ReadFile(filepath.Join(testDir, file.Name()))
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
	return engine.GetEngine().SyncResolve(res.PublicID)
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
	tsk, err := task.Create(dbp, tmpl, "", nil, nil, inputs, nil, nil)
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
	return engine.GetEngine().SyncResolve(res.PublicID)
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

type lintingAndValidationTest struct {
	NilResolution bool
	NilError      bool
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
	expectedResult := map[string]lintingAndValidationTest{
		"lintingError.yaml":            {true, false},
		"lintingRootKey.yaml":          {true, false},
		"lintingReservedStep.yaml":     {true, false},
		"customStates.yaml":            {true, false},
		"forbiddenStateImpact.yaml":    {true, false},
		"stepDetailsLintingError.yaml": {true, false},
		"circularDependencies.yaml":    {true, false},
		"selfDependency.yaml":          {true, false},
		"orphanDependencies.yaml":      {true, false},

		"lintingInfiniteOk.yaml":  {false, true},
		"lintingObject.yaml":      {false, true},
		"allowedStateImpact.yaml": {false, true},
	}

	for template, testCase := range expectedResult {
		t.Run(template, func(t *testing.T) {
			res, err := createResolution(template, map[string]interface{}{}, nil)

			if testCase.NilResolution {
				assert.Nil(t, res)
			} else {
				assert.NotNil(t, res)
			}

			if testCase.NilError {
				assert.Nil(t, err)
			} else {
				assert.NotNil(t, err)
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
	assert.Equal(t, resolution.StateTODO, res.State)
	assert.Equal(t, 0, res.Steps["stepOne"].TryCount)

	res, err = runResolution(res)

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, step.StateToRetry, res.Steps["stepOne"].State)
	assert.Equal(t, resolution.StateError, res.State)
	assert.Equal(t, 1, res.Steps["stepOne"].TryCount)

	res, err = runResolution(res)

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, step.StateToRetry, res.Steps["stepOne"].State)
	assert.Equal(t, resolution.StateError, res.State)
	assert.Equal(t, 2, res.Steps["stepOne"].TryCount)

	assert.Equal(t, "REGEXP_MATCH", res.Steps["stepTwo"].State)
	assert.Equal(t, "FOO", res.Steps["stepThree"].State)
	assert.Equal(t, "changed state matched correctly", res.Steps["stepThree"].Error)
}

func TestResolutionStateCrashed(t *testing.T) {
	res, err := createResolution("stepCondition.yaml", map[string]interface{}{}, nil)
	assert.Nil(t, err)

	res.State = resolution.StateCrashed
	res.Steps["stepOne"].State = step.StateRunning
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

	err = engine.GetEngine().Resolve(res.PublicID)

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
	assert.Equal(t, 1, len(concatList))

	firstItem := concatList[0].(map[string]interface{})
	firstItemOutput := firstItem[values.OutputKey].(map[string]interface{})
	assert.Equal(t, "foo-b-bar-b", firstItemOutput["concat"])
}

func TestForeachWithPreRun(t *testing.T) {
	input := map[string]interface{}{}
	res, err := createResolution("foreachAndPreRun.yaml", input, nil)
	require.Nilf(t, err, "expecting nil error, got %s", err)
	require.NotNil(t, res)

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
	child := output[0].(map[string]interface{})
	assert.Equal(t, "/id/1619464078", child[values.OutputKey].(string))
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

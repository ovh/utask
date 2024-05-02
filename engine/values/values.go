package values

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"text/template"
	"time"

	"github.com/Masterminds/sprig/v3"
	"github.com/gofrs/uuid"
	"github.com/juju/errors"
	"github.com/ovh/utask"
	"github.com/ovh/utask/pkg/utils"
	"github.com/robertkrimen/otto"
)

// keys to store/retrieve data from a Values struct
const (
	InputKey         = "input"
	FunctionsArgsKey = "function_args"
	ResolverInputKey = "resolver_input"
	StepKey          = "step"
	ConfigKey        = "config"
	TaskKey          = "task"
	VarKey           = "var"
	IteratorKey      = "iterator" // reserved for transient one-off values, set/unset when applying values to template

	StateKey      = "state"
	PreHookKey    = "pre_hook"
	OutputKey     = "output"
	MetadataKey   = "metadata"
	ResolutionKey = "resolution"
	ChildrenKey   = "children"
	ErrorKey      = "error"
	TryCountKey   = "try_count"
	MaxRetriesKey = "max_retries"
)

// Values is a container for all the live data of a running task
type Values struct {
	m       map[string]interface{}
	funcMap map[string]interface{}
}

// Variable holds a named variable, with either a JS expression to be evalued
// or a concrete value
type Variable struct {
	Name              string      `json:"name"`
	Expression        string      `json:"expression"`
	ExpressionTimeout string      `json:"expression_timeout"`
	Value             interface{} `json:"value"`
	evalCachedResult  interface{}
}

// NewValues instantiates a new Values holder,
// complete with custom templating functions available to task template authors
func NewValues() *Values {
	v := &Values{
		m: map[string]interface{}{
			InputKey:         map[string]interface{}{},
			ResolverInputKey: map[string]interface{}{},
			FunctionsArgsKey: map[string]interface{}{},
			StepKey:          map[string]interface{}{},
			TaskKey:          map[string]interface{}{},
			ConfigKey:        map[string]interface{}{},
			VarKey:           map[string]*Variable{},
			IteratorKey:      nil,
		},
	}
	v.funcMap = sprig.FuncMap()
	v.funcMap["field"] = v.fieldTmpl
	v.funcMap["fieldFrom"] = fieldFromTmpl
	v.funcMap["eval"] = v.varEval
	v.funcMap["evalCache"] = v.varEvalCache
	v.funcMap["fromJson"] = v.fromJSON
	v.funcMap["mustFromJson"] = v.mustFromJSON
	v.funcMap["uuid"] = uuid.NewV4
	v.funcMap["b64RawEnc"] = v.b64RawEnc
	v.funcMap["b64RawDec"] = v.b64RawDec

	return v
}

// Clone duplicates the values object
func (v *Values) Clone() (*Values, error) {
	n := NewValues()

	for key, val := range v.m {
		if val == nil {
			continue
		}

		ba, err := utils.JSONMarshal(val)
		if err != nil {
			return nil, err
		}

		switch key {
		case VarKey:
			newobj := &map[string]*Variable{}
			err = utils.JSONnumberUnmarshal(bytes.NewReader(ba), newobj)
			if err != nil {
				return nil, err
			}
			n.m[key] = *newobj
		default:
			var newobj interface{}
			err = utils.JSONnumberUnmarshal(bytes.NewReader(ba), &newobj)
			if err != nil {
				return nil, err
			}
			n.m[key] = newobj
		}
	}

	return n, nil
}

// SetInput stores a task's inputs in Values
func (v *Values) SetInput(in map[string]interface{}) {
	v.m[InputKey] = in
}

// SetResolverInput stores a task resolver's inputs in Values
func (v *Values) SetResolverInput(in map[string]interface{}) {
	v.m[ResolverInputKey] = in
}

// SetFunctionsArgs stores a task function args in Values
func (v *Values) SetFunctionsArgs(in map[string]interface{}) {
	v.m[FunctionsArgsKey] = in
}

// SetConfig stores items retrieved from configstore in Values
func (v *Values) SetConfig(cfg map[string]interface{}) {
	v.m[ConfigKey] = cfg
}

// GetOutput returns the output of a named step
func (v *Values) GetOutput(stepName string) interface{} {
	return v.getStepData(stepName, OutputKey)
}

// SetOutput stores a step's output in Values
func (v *Values) SetOutput(stepName string, value interface{}) {
	v.setStepData(stepName, OutputKey, value)
}

// SetPreHook stores a step's prehook in Values
func (v *Values) SetPreHook(output, metadata interface{}) {
	v.m[PreHookKey] = map[string]interface{}{
		OutputKey:   output,
		MetadataKey: metadata,
	}
}

// UnsetOutput empties the output data of a named step
func (v *Values) UnsetOutput(stepName string) {
	v.unsetStepData(stepName, OutputKey)
}

// GetMetadata returns the metadata of a named step
func (v *Values) GetMetadata(stepName string) interface{} {
	return v.getStepData(stepName, MetadataKey)
}

// SetMetadata stores a step's metadata in Values
func (v *Values) SetMetadata(stepName string, value interface{}) {
	v.setStepData(stepName, MetadataKey, value)
}

// UnsetMetadata empties the metadata of a named step
func (v *Values) UnsetMetadata(stepName string) {
	v.unsetStepData(stepName, MetadataKey)
}

// GetChildren returns the collection of results issued from a named "foreach" step
func (v *Values) GetChildren(stepName string) interface{} {
	return v.getStepData(stepName, ChildrenKey)
}

// SetChildren stores the collection of results issued from a named "foreach" step
func (v *Values) SetChildren(stepName string, value interface{}) {
	v.setStepData(stepName, ChildrenKey, value)
}

// UnsetChildren empties results for a named "foreach" step
func (v *Values) UnsetChildren(stepName string) {
	v.unsetStepData(stepName, ChildrenKey)
}

// GetError returns the error resulting from a failed step
func (v *Values) GetError(stepName string) interface{} {
	return v.getStepData(stepName, ErrorKey)
}

// SetError stores the error resulting from a failed step
func (v *Values) SetError(stepName string, value interface{}) {
	v.setStepData(stepName, ErrorKey, value)
}

// UnsetError empties the error from a failed step
func (v *Values) UnsetError(stepName string) {
	v.unsetStepData(stepName, ErrorKey)
}

// GetTryCount returns the try count for a step
func (v *Values) GetTryCount(stepName string) interface{} {
	return v.getStepData(stepName, TryCountKey)
}

// SetTryCount stores the try count for a step
func (v *Values) SetTryCount(stepName string, value interface{}) {
	v.setStepData(stepName, TryCountKey, value)
}

// UnsetTryCount empties the try count for a step
func (v *Values) UnsetTryCount(stepName string) {
	v.unsetStepData(stepName, TryCountKey)
}

// GetMaxRetries returns the max retries for a step
func (v *Values) GetMaxRetries(stepName string) interface{} {
	return v.getStepData(stepName, MaxRetriesKey)
}

// SetMaxRetries stores the max retries for a step
func (v *Values) SetMaxRetries(stepName string, value interface{}) {
	v.setStepData(stepName, MaxRetriesKey, value)
}

// UnsetMaxRetries empties the max retries for a step
func (v *Values) UnsetMaxRetries(stepName string) {
	v.unsetStepData(stepName, MaxRetriesKey)
}

// GetState returns the state of a step
func (v *Values) GetState(stepName string) interface{} {
	return v.getStepData(stepName, StateKey)
}

// SetState stores the state of a step
func (v *Values) SetState(stepName string, value interface{}) {
	v.setStepData(stepName, StateKey, value)
}

// UnsetState empties the state of a step
func (v *Values) UnsetState(stepName string) {
	v.unsetStepData(stepName, StateKey)
}

func (v *Values) getStepData(stepName, field string) interface{} {
	stepmap := v.m[StepKey].(map[string]interface{})
	if stepmap[stepName] == nil {
		return nil
	}
	stepdata := stepmap[stepName].(map[string]interface{})
	if stepdata == nil {
		return nil
	}
	return stepdata[field]
}

func (v *Values) setStepData(stepName, field string, value interface{}) {
	stepmap := v.m[StepKey].(map[string]interface{})
	if stepmap[stepName] == nil {
		stepmap[stepName] = map[string]interface{}{}
	}
	stepdata := stepmap[stepName].(map[string]interface{})
	// replace nil value by an empty map to avoid storing nil elements which could lead
	// to problems when executing long dot notation chains in templates
	if value == nil {
		value = map[string]interface{}{}
	}
	stepdata[field] = value
}

func (v *Values) unsetStepData(stepName, field string) {
	stepmap := v.m[StepKey].(map[string]interface{})
	if stepmap[stepName] != nil {
		stepdata := stepmap[stepName].(map[string]interface{})
		stepdata[field] = nil
	}
}

// SetTaskInfos stores task-related data in Values
func (v *Values) SetTaskInfos(t map[string]interface{}) {
	v.m[TaskKey] = t
}

// SetVariables stores template-defined variables in Values
func (v *Values) SetVariables(vars []Variable) {
	varmap := make(map[string]*Variable)
	for i := range vars {
		varmap[vars[i].Name] = &vars[i]
	}
	v.m[VarKey] = varmap
}

// SetIterator stores the data for the current item in an iteration
func (v *Values) SetIterator(i interface{}) {
	v.m[IteratorKey] = i
}

// UnsetIterator cleans up data on iterator
func (v *Values) UnsetIterator() {
	v.m[IteratorKey] = nil
}

// GetVariables returns all template variables stored in Values
func (v *Values) GetVariables() map[string]*Variable {
	return v.m[VarKey].(map[string]*Variable)
}

// GetSteps returns all consolidated step data stored in Values
func (v *Values) GetSteps() map[string]interface{} {
	return v.m["step"].(map[string]interface{})
}

// Apply takes data from Values to replace templating placeholders in a string
func (v *Values) Apply(templateStr string, item interface{}, stepName string) ([]byte, error) {
	tmpl, err := template.
		New("tmpl").
		Funcs(v.funcMap).
		Parse(templateStr)
	if err != nil {
		return nil, errors.NewBadRequest(err, "Templating error")
	}

	b := new(bytes.Buffer)

	if item != nil {
		v.SetIterator(item)
		defer v.UnsetIterator()
	}

	if stepName != "" {
		v.SetOutput(utask.This, v.GetOutput(stepName))
		defer v.UnsetOutput(utask.This)

		v.SetMetadata(utask.This, v.GetMetadata(stepName))
		defer v.UnsetMetadata(utask.This)

		v.SetChildren(utask.This, v.GetChildren(stepName))
		defer v.UnsetChildren(utask.This)

		v.SetError(utask.This, v.GetError(stepName))
		defer v.UnsetError(utask.This)

		v.SetTryCount(utask.This, v.GetTryCount(stepName))
		defer v.UnsetTryCount(utask.This)

		v.SetMaxRetries(utask.This, v.GetMaxRetries(stepName))
		defer v.UnsetMaxRetries(utask.This)

		v.SetState(utask.This, v.GetState(stepName))
		defer v.UnsetState(utask.This)
	}

	err = tmpl.Execute(b, v.m)
	if err != nil {
		return nil, errors.NewBadRequest(err, "Templating error")
	}

	return b.Bytes(), nil
}

// templating funcs

// zero-value return so that text template will evaluate to false in pipelines
// and print it to <no value> outside of pipelines
var zero reflect.Value

func (v *Values) fieldTmpl(key ...string) reflect.Value {
	var i interface{} //nolint:gosimple

	i = map[string]interface{}(v.m)
	return fieldFn(i, key)
}

func fieldFn(i interface{}, keys []string) reflect.Value {
	var ok bool

	for _, k := range keys {
		switch i.(type) { //nolint:gosimple
		case map[string]interface{}:
			i, ok = i.(map[string]interface{})[k]
			if !ok {
				return zero
			}
		case map[string]string:
			i, ok = i.(map[string]string)[k]
			if !ok {
				return zero
			}
		default:
			return zero
		}
	}
	return reflect.ValueOf(i)
}

func fieldFromTmpl(params ...interface{}) (reflect.Value, error) {
	if len(params) < 2 {
		return zero, errors.New("invalid number of parameters given")
	}
	var i interface{}
	var ok bool
	i, ok = params[len(params)-1].(map[string]interface{})
	if !ok {
		return zero, errors.New("unable to cast given data to a map[string]")
	}

	keys := []string{}
	for j := range params {
		if j >= len(params)-1 {
			break
		}

		item, ok := params[j].(string)
		if !ok {
			return zero, errors.New("foo")
		}
		keys = append(keys, item)
	}

	return fieldFn(i, keys), nil
}

func (v *Values) varEvalCache(varName string) (interface{}, error) {
	i, ok := v.GetVariables()[varName]
	if !ok {
		return nil, fmt.Errorf("Var name not found in template: '%s'", varName)
	}

	if i.evalCachedResult != nil {
		return i.evalCachedResult, nil
	}

	res, err := v.varEval(varName)
	if err != nil {
		return nil, err
	}

	i.evalCachedResult = res
	return i.evalCachedResult, nil
}

func (v *Values) varEval(varName string) (interface{}, error) {
	i, ok := v.GetVariables()[varName]
	if !ok {
		return nil, fmt.Errorf("Var name not found in template: '%s'", varName)
	}

	if i.Value != nil {
		val, ok := i.Value.(string)
		if !ok {
			// Value is not a string, won't be able to template it
			return i.Value, nil
		}
		valS, err := v.Apply(val, nil, "")
		if err != nil {
			return nil, err
		}
		return string(valS), nil
	}

	exp, err := v.Apply(i.Expression, nil, "")
	if err != nil {
		return nil, err
	}

	var timeout = time.Second * 10
	if i.ExpressionTimeout != "" {
		timeout, err = time.ParseDuration(i.ExpressionTimeout)
		if err != nil {
			return nil, err
		}
	}

	res, err := evalUnsafe(exp, timeout)
	if err != nil {
		return nil, err
	}

	return res.String(), nil
}

// fromJSON decodes JSON into a structured value, ignoring errors.
func (v *Values) fromJSON(s string) (reflect.Value, error) {
	output, _ := v.mustFromJSON(s)
	return output, nil
}

// mustFromJSON decodes JSON into a structured value, returning errors.
func (v *Values) mustFromJSON(s string) (reflect.Value, error) {
	var output interface{}
	err := json.Unmarshal([]byte(s), &output)
	return reflect.ValueOf(output), err
}

func (v *Values) b64RawDec(s string) string {
	data, err := base64.RawStdEncoding.DecodeString(s)
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (v *Values) b64RawEnc(s string) string {
	return base64.RawStdEncoding.EncodeToString([]byte(s))
}

var errTimedOut = errors.New("Timed out variable evaluation")

func evalUnsafe(exp []byte, delay time.Duration) (v otto.Value, err error) {

	VM := otto.New()
	VM.Interrupt = make(chan func(), 1)

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Recovered from unsafe JS eval: %v", r)
		}
	}()

	go func() {
		time.Sleep(delay)
		VM.Interrupt <- func() {
			panic(errTimedOut)
		}
	}()

	v, err = VM.Run(exp)

	return
}

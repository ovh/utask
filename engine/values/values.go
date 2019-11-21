package values

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/template"
	"time"

	"github.com/Masterminds/sprig"
	"github.com/juju/errors"
	"github.com/ovh/utask"
	"github.com/robertkrimen/otto"
)

// keys to store/retrieve data from a Values struct
const (
	InputKey         = "input"
	ResolverInputKey = "resolver_input"
	StepKey          = "step"
	ConfigKey        = "config"
	TaskKey          = "task"
	VarKey           = "var"
	IteratorKey      = "iterator" // reserved for transient one-off values, set/unset when applying values to template

	OutputKey   = "output"
	MetadataKey = "metadata"
	ChildrenKey = "children"
	ErrorKey    = "error"
)

// Values is a container for all the live data of a running task
type Values struct {
	m       map[string]interface{}
	funcMap map[string]interface{}
}

// Variable holds a named variable, with either a JS expression to be evalued
// or a concrete value
type Variable struct {
	Name       string      `json:"name"`
	Expression string      `json:"expression"`
	Value      interface{} `json:"value"`
}

// NewValues instantiates a new Values holder,
// complete with custom templating functions available to task template authors
func NewValues() *Values {
	v := &Values{
		m: map[string]interface{}{
			InputKey:         map[string]interface{}{},
			ResolverInputKey: map[string]interface{}{},
			StepKey:          map[string]interface{}{},
			TaskKey:          map[string]interface{}{},
			ConfigKey:        map[string]interface{}{},
			VarKey:           map[string]*Variable{},
			IteratorKey:      nil,
		},
	}
	v.funcMap = sprig.FuncMap()
	v.funcMap["field"] = v.fieldTmpl
	v.funcMap["jsonfield"] = v.jsonFieldTmpl
	v.funcMap["jsonmarshal"] = v.jsonMarshal
	v.funcMap["eval"] = v.varEval
	return v
}

// SetInput stores a task's inputs in Values
func (v *Values) SetInput(in map[string]interface{}) {
	v.m[InputKey] = in
}

// SetResolverInput stores a task resolver's inputs in Values
func (v *Values) SetResolverInput(in map[string]interface{}) {
	v.m[ResolverInputKey] = in
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
	}

	err = tmpl.Execute(b, v.m)
	if err != nil {
		return nil, errors.NewBadRequest(err, "Templating error")
	}

	return b.Bytes(), nil
}

// templating funcs

func (v *Values) fieldTmpl(key ...string) (interface{}, error) {
	var i interface{}

	i = map[string]interface{}(v.m)
	var ok bool
	var previousNotFound string

	for _, k := range key {
		switch i.(type) {
		case map[string]interface{}:
			previousNotFound = ""
			i, ok = i.(map[string]interface{})[k]
			if !ok {
				previousNotFound = k
				i = "<no value>"
			}
		case map[string]string:
			previousNotFound = ""
			i, ok = i.(map[string]string)[k]
			if !ok {
				previousNotFound = k
				i = "<no value>"
			}
		default:
			return nil, fmt.Errorf("cannot dereference %T for key %q; previous key not found %q", i, k, previousNotFound)
		}
	}
	return i, nil
}

func (v *Values) jsonFieldTmpl(key ...string) (interface{}, error) {
	i, err := v.fieldTmpl(key...)
	if err != nil {
		return nil, err
	}
	return v.jsonMarshal(i)
}

func (v *Values) jsonMarshal(i interface{}) (interface{}, error) {
	marshalled, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}
	return string(marshalled), nil
}

func (v *Values) varEval(varName string) (interface{}, error) {
	i, ok := v.GetVariables()[varName]
	if !ok {
		return nil, fmt.Errorf("Var name not found in template: '%s'", varName)
	}

	if i.Value == nil {
		exp, err := v.Apply(i.Expression, nil, "")
		if err != nil {
			return nil, err
		}

		res, err := evalUnsafe(exp, time.Second*5)
		if err != nil {
			return nil, err
		}

		i.Value = res.String()
	}

	return i.Value, nil
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

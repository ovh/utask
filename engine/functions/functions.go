package functions

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/ovh/utask/engine/step/condition"
	"github.com/ovh/utask/engine/step/executor"
	"github.com/ovh/utask/engine/values"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

var (
	functionsImported = make(map[string]*Function)

	functionArgsRegexp = regexp.MustCompile(fmt.Sprintf(`\{\{\s*\.%s\.([a-zA-Z.]+)\s*\}\}`, values.FunctionsArgsKey))
)

// Function describes one reusable action that can be used in steps.
// This function will be resolved as another function or a builtin/plugin
// action. Its configuration will be resolved and can takes parameters
// in the configuration given with templated variables under {{ .functions_args.xxx }}
type Function struct {
	Name         string                 `json:"name"`
	Action       executor.Executor      `json:"action"`
	PreHook      *executor.Executor     `json:"pre_hook,omitempty"`
	Conditions   []*condition.Condition `json:"conditions,omitempty"`
	CustomStates []string               `json:"custom_states,omitempty"`

	fileName    string
	rawFunction json.RawMessage
	args        []string
}

func (f *Function) init() error {
	var config map[string]interface{}

	err := json.Unmarshal(f.Action.Configuration, &config)
	if err != nil {
		return err
	}

	f.args, err = extractArguments("", reflect.ValueOf(config))
	return err
}

// extractArguments goes through the value given, initially the configuration map,
// to find all the names of the variables that needs an input.
func extractArguments(path string, v reflect.Value) ([]string, error) {
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}

	var result = []string{}

	switch v.Kind() {
	case reflect.Map:
		iter := v.MapRange()

		for iter.Next() {
			mv := iter.Value()
			k := iter.Key()

			subpath := path
			if subpath != "" {
				subpath += "."
			}
			subpath += fmt.Sprint(k.Interface())
			if k.Kind() != reflect.String {
				return nil, fmt.Errorf("%s is not a string", subpath)
			}

			ret, err := extractArguments(subpath, mv)
			if err != nil {
				return nil, err
			}
			result = append(result, ret...)
		}

	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			mv := v.Index(i)

			subpath := path
			if subpath != "" {
				subpath += "."
			}
			subpath += fmt.Sprint(i)
			ret, err := extractArguments(subpath, mv)
			if err != nil {
				return nil, err
			}
			result = append(result, ret...)

		}
	case reflect.String:
		s := v.String()
		if functionArgsRegexp.MatchString(s) {
			for _, submatches := range functionArgsRegexp.FindAllStringSubmatch(s, -1) {
				result = append(result, submatches[1])
			}
		}
	}

	return result, nil
}

// Exec is the implementation of the runner.Exec function but does nothing: function runners
// are just place holders to resolve to actual plugin/builtin.
func (f *Function) Exec(stepName string, baseConfig json.RawMessage, config json.RawMessage, ctx interface{}) (interface{}, interface{}, map[string]string, error) {
	return nil, nil, nil, errors.New("functions cannot be executed")
}

// ValidConfig insure that the given configuration resolves all the input needed by the function.
func (f *Function) ValidConfig(baseConfig json.RawMessage, config json.RawMessage) error {
	for _, arg := range f.args {
		result := gjson.GetBytes(config, arg)
		if result.Raw == "" {
			return fmt.Errorf("missing function_args %q", arg)
		}
	}
	return nil
}

// Resources returns the resources used by the config
func (f *Function) Resources(baseConfig json.RawMessage, config json.RawMessage) []string {
	return []string{}
}

// Context is the implementation of the runner.Context function but does nothing: function runners
// are just place holders to resolve to actual plugin/builtin.
func (s *Function) Context(stepName string) interface{} {
	return nil
}

// MetadataSchema returns the configuration schemas of the function
func (s *Function) MetadataSchema() json.RawMessage {
	return s.rawFunction
}

// LoadFromDir loads recursively all the function from a given directory.
func LoadFromDir(directory string) error {
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		logrus.Warnf("Ignoring functions directory %s: %s", directory, err)
		return nil
	}

	for _, file := range files {
		if file.IsDir() {
			if err := LoadFromDir(path.Join(directory, file.Name())); err != nil {
				return err
			}
			continue
		}

		if !strings.HasSuffix(file.Name(), ".yaml") || strings.HasPrefix(file.Name(), ".") {
			continue
		}

		content, err := ioutil.ReadFile(path.Join(directory, file.Name()))
		if err != nil {
			return err
		}

		var function Function
		if err = yaml.Unmarshal(content, &function); err != nil {
			return err
		}
		function.fileName = path.Join(directory, file.Name())

		if err := function.init(); err != nil {
			return err
		}
		if previous, exists := functionsImported[function.Name]; exists {
			return fmt.Errorf("%q: function already exists and was declared in %q", function.fileName, previous.fileName)
		}

		functionsImported[function.Name] = &function
		logrus.Infof("Imported function %q", function.Name)
	}

	return nil
}

// List returns the list of functions imported.
func List() []string {
	var result = []string{}

	for k := range functionsImported {
		result = append(result, k)
	}
	sort.Strings(result)
	return result
}

// Get return the function identified by the name in parameter and whether it exists.
func Get(name string) (*Function, bool) {
	s, exists := functionsImported[name]
	return s, exists
}

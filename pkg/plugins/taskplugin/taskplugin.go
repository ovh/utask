package taskplugin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/juju/errors"
	"github.com/ovh/utask/pkg/jsonschema"
	"github.com/ovh/utask/pkg/utils"
)

// ConfigFunc is a type of function to validate the contents of a configuration payload
type ConfigFunc func(interface{}) error

// ExecFunc is a type of function to be implemented by a plugin to perform an action in a task
type ExecFunc func(string, interface{}, interface{}) (interface{}, interface{}, error)

// PluginExecutor is a structure to generate action executors from different implementations
// builtin or loaded as custom extensions
type PluginExecutor struct {
	configfunc     ConfigFunc
	execfunc       ExecFunc
	configFactory  func() interface{}
	pluginName     string
	pluginVersion  string
	contextFactory func(string) interface{}
	metadataSchema json.RawMessage
	tagsFunc       tagsFunc
}

// Context generates a context payload to pass to Exec()
func (r PluginExecutor) Context(stepName string) interface{} {
	if r.contextFactory != nil {
		return r.contextFactory(stepName)
	}
	return nil
}

// ValidConfig asserts that a given configuration payload complies with the executor's definition
func (r PluginExecutor) ValidConfig(baseConfig json.RawMessage, config json.RawMessage) error {
	if r.configFactory != nil {
		cfg := r.configFactory()
		if len(baseConfig) > 0 {
			err := utils.JSONnumberUnmarshal(bytes.NewReader(baseConfig), cfg)
			if err != nil {
				return errors.Annotate(err, "failed to unmarshal base configuration")
			}
		}
		err := utils.JSONnumberUnmarshal(bytes.NewReader(config), cfg)
		if err != nil {
			return errors.Annotate(err, "failed to unmarshal configuration")
		}
		return r.configfunc(cfg)
	}

	return nil
}

// Exec performs the action implemented by the executor
func (r PluginExecutor) Exec(stepName string, baseConfig json.RawMessage, config json.RawMessage, ctx interface{}) (interface{}, interface{}, map[string]string, error) {
	var cfg interface{}

	if r.configFactory != nil {
		cfg = r.configFactory()
		if len(baseConfig) > 0 {
			err := utils.JSONnumberUnmarshal(bytes.NewReader(baseConfig), cfg)
			if err != nil {
				return nil, nil, nil, errors.Annotate(err, "failed to unmarshal base configuration")
			}
		}
		err := utils.JSONnumberUnmarshal(bytes.NewReader(config), cfg)
		if err != nil {
			return nil, nil, nil, errors.Annotate(err, "failed to unmarshal configuration")
		}
	}
	output, metadata, err := r.execfunc(stepName, cfg, ctx)

	var tags map[string]string
	if r.tagsFunc != nil {
		tags = r.tagsFunc(cfg, ctx, output, metadata, err)
	}
	return output, metadata, tags, err
}

// PluginName returns a plugin's name
func (r PluginExecutor) PluginName() string {
	return r.pluginName
}

// PluginVersion returns a plugin's version
func (r PluginExecutor) PluginVersion() string {
	return r.pluginVersion
}

// MetadataSchema returns json schema to validate the metadata returned on execution
func (r PluginExecutor) MetadataSchema() json.RawMessage {
	return r.metadataSchema
}

type tagsFunc func(config, ctx, output, metadata interface{}, err error) map[string]string

// PluginOpt is a helper struct to customize an action executor
type PluginOpt struct {
	configCheckFunc ConfigFunc
	configObj       interface{}
	contextObj      interface{}
	contextFunc     func(string) interface{}
	metadataFunc    func() string
	tagsFunc        tagsFunc
}

// WithConfig defines the configuration struct and validation function
// for a plugin
func WithConfig(configCheckFunc ConfigFunc, configObj interface{}) func(*PluginOpt) {
	return func(o *PluginOpt) {
		o.configCheckFunc = configCheckFunc
		o.configObj = configObj
	}
}

// WithContext defines the context object expected by the plugin
func WithContext(contextObj interface{}) func(*PluginOpt) {
	return func(o *PluginOpt) {
		o.contextObj = contextObj
	}
}

// WithContextFunc defines a context-generating function
func WithContextFunc(contextFunc func(string) interface{}) func(*PluginOpt) {
	return func(o *PluginOpt) {
		o.contextFunc = contextFunc
	}
}

// WithExecutorMetadata defines a jsonschema-generating function
func WithExecutorMetadata(metadataFunc func() string) func(*PluginOpt) {
	return func(o *PluginOpt) {
		o.metadataFunc = metadataFunc
	}
}

// WithTags defines a function to manipulate the tags of a task.
func WithTags(fn tagsFunc) func(*PluginOpt) {
	return func(o *PluginOpt) {
		o.tagsFunc = fn
	}
}

// New generates a step action executor from a given plugin
func New(pluginName string, pluginVersion string, execfunc ExecFunc, opts ...func(*PluginOpt)) PluginExecutor {

	pOpt := &PluginOpt{}

	for _, o := range opts {
		o(pOpt)
	}

	if pluginName == "" {
		panic("registering plugin without name")
	}
	if execfunc == nil {
		panic(fmt.Sprintf("plugin executor '%s': nil exec function", pluginName))
	}
	if pOpt.configObj != nil && pOpt.configCheckFunc == nil {
		panic(fmt.Sprintf("plugin executor '%s': nil config check function", pluginName))
	}
	if pOpt.contextObj != nil && pOpt.contextFunc != nil {
		panic(fmt.Sprintf("plugin executor '%s': conflicting context object + factory", pluginName))
	}

	var schema json.RawMessage
	if pOpt.metadataFunc != nil {
		metadata := pOpt.metadataFunc()
		s, err := jsonschema.NormalizeAndCompile(pluginName, []byte(metadata))
		if err != nil {
			panic(fmt.Sprintf("plugin executor %q: %s", pluginName, err.Error()))
		}
		schema = s
	}

	var contextFactory func(string) interface{}

	if pOpt.contextFunc != nil {
		contextFactory = pOpt.contextFunc
	} else if pOpt.contextObj != nil {
		v := reflect.ValueOf(pOpt.contextObj)
		for v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		marshaled, err := utils.JSONMarshal(pOpt.contextObj)
		if err != nil {
			panic(fmt.Sprintf("plugin executor '%s': failed to marshal context object: %s", pluginName, err))
		}
		contextFactory = func(stepName string) interface{} {
			i := reflect.New(v.Type()).Interface()
			utils.JSONnumberUnmarshal(bytes.NewReader(marshaled), i)
			return i
		}
	}

	var configFactory func() interface{}

	if pOpt.configObj != nil {
		v := reflect.ValueOf(pOpt.configObj)
		for v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		configFactory = func() interface{} {
			return reflect.New(v.Type()).Interface()
		}
	}

	return PluginExecutor{
		pluginName:     pluginName,
		pluginVersion:  pluginVersion,
		configfunc:     pOpt.configCheckFunc,
		execfunc:       execfunc,
		configFactory:  configFactory,
		contextFactory: contextFactory,
		metadataSchema: schema,
		tagsFunc:       pOpt.tagsFunc,
	}
}

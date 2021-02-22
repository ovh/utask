package plugins

import (
	"fmt"
	"io/ioutil"
	"plugin"
	"reflect"
	"strings"

	"github.com/ovh/configstore"
	"github.com/sirupsen/logrus"

	"github.com/ovh/utask/api"
	"github.com/ovh/utask/engine/step"
)

// TaskPlugin represents the interface for every executor for µtask step actions
type TaskPlugin interface {
	step.Runner
	PluginName() string
	PluginVersion() string
}

// ExecutorsFromFolder loads a collection of TaskPlugin from compiled .so plugins
// found in a folder, then registers each TaskPlugin as a step runner
// to be used by the task execution engine
func ExecutorsFromFolder(path string) error {
	return loadPlugins(path, func(fileName string, p plugin.Symbol) error {
		plugExec, ok := p.(TaskPlugin)
		if !ok {
			return fmt.Errorf("failed to assert type of plugin '%s': expected TaskPlugin got %T", fileName, p)
		}
		if err := step.RegisterRunner(plugExec.PluginName(), plugExec); err != nil {
			return err
		}
		logrus.Infof("Registered plugin '%s' (%s)", plugExec.PluginName(), plugExec.PluginVersion())
		return nil
	})
}

// Service encapsulates the objects accessible to an initialization plugin
// this allows for custom configuration of the api server, and for the declaration
// of additional configstore providers
type Service struct {
	Store  *configstore.Store
	Server *api.Server
}

// InitializerPlugin represents the interface of an initialization plugin
// meant to customize the µtask service
type InitializerPlugin interface {
	Init(service *Service) error
	Description() string
}

// InitializersFromFolder loads initialization plugins compiled as .so files
// from a folder, runs them on a received pointer to a Service
func InitializersFromFolder(path string, service *Service) error {
	return loadPlugins(path, func(fileName string, pluginSymbol plugin.Symbol) error {
		reflectvalue := reflect.ValueOf(pluginSymbol)
		if reflectvalue.Kind() != reflect.Ptr {
			return fmt.Errorf("failed to load Plugin from %s: received a non-pointer object", fileName)
		}
		pluginInterface := reflectvalue.Elem().Interface()
		plug, ok := pluginInterface.(InitializerPlugin)
		if !ok {
			return fmt.Errorf("failed to assert type of plugin '%s': expected InitializerPlugin got %T", fileName, pluginInterface)
		}
		if err := plug.Init(service); err != nil {
			return fmt.Errorf("failed to run initialization plugin: %s", err)
		}
		logrus.Infof("Ran initialization plugin: %s", plug.Description())
		return nil
	})
}

func loadPlugins(path string, load func(string, plugin.Symbol) error) error {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		logrus.Warnf("Ignoring plugin directory %s: %s", path, err)
	} else {
		for _, file := range files {
			if file.IsDir() || !strings.HasSuffix(file.Name(), ".so") {
				continue
			}
			plug, err := plugin.Open(fmt.Sprintf("%s/%s", path, file.Name()))
			if err != nil {
				return fmt.Errorf("failed to load plugin '%s': %s", file.Name(), err)
			}
			plugSym, err := plug.Lookup("Plugin")
			if err != nil {
				return fmt.Errorf("failed to lookup 'Plugin' symbol in plugin '%s'", file.Name())
			}
			if err := load(file.Name(), plugSym); err != nil {
				return err
			}
		}
	}
	return nil
}

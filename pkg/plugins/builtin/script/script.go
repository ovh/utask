package script

import (
	"github.com/ovh/utask/pkg/plugins/taskplugin"
)

// the script plugin execute scripts
var (
	Plugin = taskplugin.New("script", "0.1", exec,
		taskplugin.WithConfig(validConfig, Config{}),
	)
)

// Config is the configuration needed to send a ping
type Config struct {
	File    string
	Argv    []string
	Stdin   string
	Timeout string
}

func validConfig(config interface{}) error {
	cfg := config.(*Config)

	return nil
}

func exec(stepName string, config interface{}, ctx interface{}) (interface{}, interface{}, error) {
	cfg := config.(*Config)

	return nil, nil, nil
}

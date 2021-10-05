package pluginwatcher

import (
	"fmt"
	"strings"

	"github.com/ovh/utask/pkg/plugins/taskplugin"
)

// The watcher plugin allow to update the allowed watcher usernames of a task.
var (
	Plugin = taskplugin.New("watcher", "0.1", exec,
		taskplugin.WithConfig(validConfig, Config{}),
		taskplugin.WithWatchers(watchers),
	)
)

// Config represents the configuration of the plugin.
type Config struct {
	Usernames []string `json:"usernames"`
}

func validConfig(config interface{}) error {
	cfg := config.(*Config)

	for i, v := range cfg.Usernames {
		if strings.TrimSpace(v) == "" {
			return fmt.Errorf("invalid watcher username at position %d", i)
		}
	}
	return nil
}

func exec(stepName string, config interface{}, ctx interface{}) (interface{}, interface{}, error) {
	return nil, nil, nil
}

func watchers(config, _, _, _ interface{}, _ error) []string {
	if config == nil {
		return nil
	}
	cfg := config.(*Config)

	return cfg.Usernames
}

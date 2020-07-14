package plugintag

import (
	"github.com/ovh/utask/pkg/plugins/taskplugin"
	"github.com/ovh/utask/pkg/utils"
)

// The tag plugin allow to update the tags of a task.
var (
	Plugin = taskplugin.New("tag", "0.1", exec,
		taskplugin.WithConfig(validConfig, Config{}),
		taskplugin.WithTags(tags),
	)
)

// Config represents the configuration of the plugin.
type Config struct {
	Tags map[string]string `json:"tags"`
}

func validConfig(config interface{}) error {
	cfg := config.(*Config)

	if err := utils.ValidateTags(cfg.Tags); err != nil {
		return err
	}

	return nil
}

func exec(stepName string, config interface{}, ctx interface{}) (interface{}, interface{}, error) {
	return nil, nil, nil
}

func tags(config, _, _, _ interface{}, _ error) map[string]string {
	if config == nil {
		return nil
	}
	cfg := config.(*Config)

	return cfg.Tags
}

package echo

import (
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/juju/errors"
	"github.com/ovh/utask/pkg/plugins/taskplugin"
	"github.com/ovh/utask/pkg/utils"
)

// the echo plugin is used to "manually" build result outputs
// allowing to aggregate several results in a consolidated structure
var (
	Plugin = taskplugin.New("echo", "0.1", exec,
		taskplugin.WithConfig(validConfig, Config{}),
	)
)

// Config describes transparently the outcome of execution
// output:   an arbitrary object, equivalent to a successful return
// metadata: the metadata returned by execution, if any
// unmarshal: defines whether unmarshal the output if it's a string or byte array before returning
// error_message: the outcome of a non-successful execution
// error_type:    choose between client|server, to trigger different behavior (blocked VS retry)
type Config struct {
	Output       interface{}            `json:"output"`
	Metadata     map[string]interface{} `json:"metadata"`
	Unmarshal    bool                   `json:"unmarshal"`
	ErrorMessage string                 `json:"error_message"`
	ErrorType    string                 `json:"error_type"` // default if empty: server -> ie. retry
}

func validConfig(config interface{}) error {
	cfg := config.(*Config)
	switch cfg.ErrorType {
	case "client", "server", "":
	default:
		return errors.New("Wrong error type: expecting 'client' or 'server'")
	}
	return nil
}

func exec(stepName string, config interface{}, ctx interface{}) (interface{}, interface{}, error) {
	cfg := config.(*Config)
	var resultErr error
	if cfg.ErrorMessage != "" {
		switch cfg.ErrorType {
		case "client":
			resultErr = errors.NewBadRequest(nil, cfg.ErrorMessage)
		default:
			resultErr = errors.New(cfg.ErrorMessage)
		}
	}

	var output interface{} = cfg.Output
	if cfg.Unmarshal {
		var content []byte
		switch v := cfg.Output.(type) {
		case string:
			content = []byte(v)
		case []byte:
			content = v
		default:
			return nil, nil, fmt.Errorf("cannot unmarshal: invalid data type (%T)", cfg.Output)
		}

		if err := yaml.Unmarshal(content, &output, utils.JSONUseNumber); err != nil {
			return nil, nil, fmt.Errorf("failed to unmarshal output: %s", err)
		}
	}

	return output, cfg.Metadata, resultErr
}

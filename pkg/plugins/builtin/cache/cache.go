package plugincache

import (
	"encoding/json"
	"fmt"

	"github.com/juju/errors"
	"github.com/loopfz/gadgeto/zesty"

	"github.com/ovh/utask"
	"github.com/ovh/utask/pkg/plugins/taskplugin"
)

var (
	Plugin = taskplugin.New("cache", "0.1", exec,
		taskplugin.WithConfig(validConfig, Config{}),
	)
)

// Config holds the configuration for a cache action.
// Action: "set", "get", or "delete"
// Key: the cache key (required)
// Value: the value to store (required for "set", ignored otherwise)
// TTL: time-to-live in seconds (0 means no expiration, only used with "set")
type Config struct {
	Action string      `json:"action"`
	Key    string      `json:"key"`
	Value  interface{} `json:"value,omitempty"`
	TTL    int64       `json:"ttl,omitempty"`
}

func validConfig(config interface{}) error {
	cfg := config.(*Config)

	if cfg.Key == "" {
		return fmt.Errorf("missing required parameter %q", "key")
	}

	switch cfg.Action {
	case "set":
		return nil
	case "get":
		return nil
	case "delete":
		return nil
	default:
		return fmt.Errorf("invalid action %q: expected \"set\", \"get\", or \"delete\"", cfg.Action)
	}
}

func exec(stepName string, config interface{}, ctx interface{}) (interface{}, interface{}, error) {
	cfg := config.(*Config)

	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return nil, nil, err
	}

	switch cfg.Action {
	case "set":
		return execSet(dbp, cfg)
	case "get":
		return execGet(dbp, cfg)
	case "delete":
		return execDelete(dbp, cfg)
	default:
		return nil, nil, errors.BadRequestf("invalid action %q", cfg.Action)
	}
}

func execSet(dbp zesty.DBProvider, cfg *Config) (interface{}, interface{}, error) {
	valueBytes, err := json.Marshal(cfg.Value)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal value: %s", err)
	}

	if err := setCacheEntry(dbp, cfg.Key, valueBytes, cfg.TTL); err != nil {
		return nil, nil, err
	}

	return map[string]interface{}{
		"key":    cfg.Key,
		"cached": true,
	}, nil, nil
}

func execGet(dbp zesty.DBProvider, cfg *Config) (interface{}, interface{}, error) {
	entry, err := getCacheEntry(dbp, cfg.Key)
	if err != nil {
		if errors.Is(err, errors.NotFound) {
			return map[string]interface{}{
				"key":   cfg.Key,
				"hit":   false,
				"value": nil,
			}, nil, nil
		}
		return nil, nil, err
	}

	var value interface{}
	if err := json.Unmarshal(entry.Value, &value); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal cached value: %s", err)
	}

	return map[string]interface{}{
		"key":   cfg.Key,
		"hit":   true,
		"value": value,
	}, nil, nil
}

func execDelete(dbp zesty.DBProvider, cfg *Config) (interface{}, interface{}, error) {
	if err := deleteCacheEntry(dbp, cfg.Key); err != nil {
		return nil, nil, err
	}

	return map[string]interface{}{
		"key":     cfg.Key,
		"deleted": true,
	}, nil, nil
}

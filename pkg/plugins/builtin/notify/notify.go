package notify

import (
	"fmt"
	"sort"

	"github.com/ovh/utask"
	"github.com/ovh/utask/pkg/notify"
	"github.com/ovh/utask/pkg/plugins/taskplugin"
)

// the notify plugin broadcasts a message over all registered notification senders
var (
	Plugin = taskplugin.New("notify", "0.1", exec,
		taskplugin.WithConfig(validConfig, Config{}))
)

// Config is the configuration needed to send a notification
// consisting of a message and extra fields
// implements notify.Payload
type Config struct {
	Msg      string   `json:"message"`
	Flds     []string `json:"fields"`
	Backends []string `json:"backends"`
}

// Message returns the config's message
func (nc *Config) Message() string { return nc.Msg }

// Fields returns the config's fields
func (nc *Config) Fields() []string { return nc.Flds }

func validConfig(config interface{}) error {
	cfg := config.(*Config)
	if len(cfg.Backends) == 0 {
		return nil
	}

	snames := notify.ListSendersNames()
	// The slice must be sorted in ascending order
	// From https://golang.org/pkg/sort/#SearchStrings
	sort.Strings(snames)

	for _, backend := range cfg.Backends {
		i := sort.SearchStrings(snames, backend)
		if i >= len(snames) && snames[i] != backend {
			return fmt.Errorf("Can't find backend name: %s", backend)
		}
	}

	return nil
}

func exec(stepName string, config interface{}, ctx interface{}) (interface{}, interface{}, error) {
	cfg := config.(*Config)
	notify.Send(cfg, utask.NotifyActionsParameters{
		Disabled:       false,
		NotifyBackends: cfg.Backends,
	})
	return nil, nil, nil
}

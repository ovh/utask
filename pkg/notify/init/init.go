package init

import (
	"encoding/json"
	"fmt"

	"github.com/ovh/configstore"
	"github.com/ovh/utask"
	"github.com/ovh/utask/pkg/notify"
	"github.com/ovh/utask/pkg/notify/tatnotify"
)

// Init aims to inject user defined cfg around notify
func Init(store *configstore.Store) error {
	cfg, err := utask.Config(store)
	if err != nil {
		return err
	}
	for name, ncfg := range cfg.NotifyConfig {
		switch ncfg.Type {
		case tatnotify.Tat:
			f := utask.NotifyBackendTat{}
			if err := json.Unmarshal(ncfg.Config, &f); err != nil {
				return fmt.Errorf("Failed to retrieve Tat cfg: %s", name)
			}
			tn, err := tatnotify.NewTatNotificationSender(
				f.URL,
				f.Username,
				f.Password,
				f.Topic,
			)
			if err != nil {
				return fmt.Errorf("Failed to instantiate tat notification sender: %s", err)
			}
			notify.RegisterSender(tn, name)
		default:
			return fmt.Errorf("Failed to identify backend type: %s", ncfg.Type)
		}
	}

	notify.RegisterActions(cfg.NotifyActions)

	return nil
}

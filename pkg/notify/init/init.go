package init

import (
	"encoding/json"
	"fmt"

	"github.com/ovh/configstore"
	"github.com/ovh/utask"
	"github.com/ovh/utask/pkg/notify"
	"github.com/ovh/utask/pkg/notify/slack"
	"github.com/ovh/utask/pkg/notify/tat"
	"github.com/ovh/utask/pkg/notify/webhook"
)

const (
	errRetrieveCfg string = "Failed to retrieve cfg"
)

// Init aims to inject user defined cfg around notify
func Init(store *configstore.Store) error {
	cfg, err := utask.Config(store)
	if err != nil {
		return err
	}

	for name, ncfg := range cfg.NotifyConfig {
		switch ncfg.Type {
		case tat.Type:
			f := utask.NotifyBackendTat{}
			if err := json.Unmarshal(ncfg.Config, &f); err != nil {
				return fmt.Errorf("%s: %s, %s: %s", errRetrieveCfg, ncfg.Type, name, err)
			}
			tn, err := tat.NewTatNotificationSender(
				f.URL,
				f.Username,
				f.Password,
				f.Topic,
			)
			if err != nil {
				return fmt.Errorf("Failed to instantiate tat notification sender: %s", err)
			}
			notify.RegisterSender(tn, name)

		case slack.Type:
			f := utask.NotifyBackendSlack{}
			if err := json.Unmarshal(ncfg.Config, &f); err != nil {
				return fmt.Errorf("%s: %s, %s: %s", errRetrieveCfg, ncfg.Type, name, err)
			}
			sn := slack.NewSlackNotificationSender(f.WebhookURL)
			notify.RegisterSender(sn, name)

		case webhook.Type:
			f := utask.NotifyBackendWebhook{}
			if err := json.Unmarshal(ncfg.Config, &f); err != nil {
				return fmt.Errorf("%s: %s, %s: %s", errRetrieveCfg, ncfg.Type, name, err)
			}
			sn := webhook.NewWebhookNotificationSender(f.WebhookURL, f.Username, f.Password, f.Headers)
			notify.RegisterSender(sn, name)

		default:
			return fmt.Errorf("Failed to identify backend type: %s", ncfg.Type)
		}
	}

	notify.RegisterActions(cfg.NotifyActions)

	return nil
}

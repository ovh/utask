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
		switch ncfg.DefaultNotificationStrategy {
		case utask.NotificationStrategyAlways, utask.NotificationStrategySilent, utask.NotificationStrategyFailureOnly:
		case "":
			ncfg.DefaultNotificationStrategy = utask.NotificationStrategyAlways
		default:
			return fmt.Errorf("invalid default_notification_strategy: %q is not a valid value", ncfg.DefaultNotificationStrategy)
		}

		for _, strat := range ncfg.TemplateNotificationStrategies {
			switch strat.NotificationStrategy {
			case utask.NotificationStrategyAlways, utask.NotificationStrategySilent, utask.NotificationStrategyFailureOnly:
			default:
				return fmt.Errorf("invalid notification_strategy for templates %#v: %q is not a valid value", strat.Templates, strat.NotificationStrategy)
			}
		}

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
			notify.RegisterSender(name, tn, ncfg.DefaultNotificationStrategy, ncfg.TemplateNotificationStrategies)

		case slack.Type:
			f := utask.NotifyBackendSlack{}
			if err := json.Unmarshal(ncfg.Config, &f); err != nil {
				return fmt.Errorf("%s: %s, %s: %s", errRetrieveCfg, ncfg.Type, name, err)
			}
			sn := slack.NewSlackNotificationSender(f.WebhookURL)
			notify.RegisterSender(name, sn, ncfg.DefaultNotificationStrategy, ncfg.TemplateNotificationStrategies)

		case webhook.Type:
			f := utask.NotifyBackendWebhook{}
			if err := json.Unmarshal(ncfg.Config, &f); err != nil {
				return fmt.Errorf("%s: %s, %s: %s", errRetrieveCfg, ncfg.Type, name, err)
			}
			sn := webhook.NewWebhookNotificationSender(f.WebhookURL, f.Username, f.Password, f.Headers)
			notify.RegisterSender(name, sn, ncfg.DefaultNotificationStrategy, ncfg.TemplateNotificationStrategies)

		default:
			return fmt.Errorf("Failed to identify backend type: %s", ncfg.Type)
		}
	}

	notify.RegisterActions(cfg.NotifyActions)

	return nil
}

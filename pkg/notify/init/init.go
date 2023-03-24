package init

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ovh/configstore"

	"github.com/ovh/utask"
	"github.com/ovh/utask/pkg/notify"
	"github.com/ovh/utask/pkg/notify/opsgenie"
	"github.com/ovh/utask/pkg/notify/slack"
	"github.com/ovh/utask/pkg/notify/tat"
	"github.com/ovh/utask/pkg/notify/webhook"
)

const (
	errRetrieveCfg string = "failed to retrieve cfg"
)

// Init aims to inject user defined cfg around notify
func Init(store *configstore.Store) error {
	cfg, err := utask.Config(store)
	if err != nil {
		return err
	}

	for name, ncfg := range cfg.NotifyConfig {
		newncfg, err := validateAndNormalizeNotificationStrategy(ncfg)
		if err != nil {
			return err
		}

		// save normalisation modifications
		ncfg.DefaultNotificationStrategy = newncfg.DefaultNotificationStrategy

		switch ncfg.Type {
		case opsgenie.Type:
			f := utask.NotifyBackendOpsGenie{}
			if err := json.Unmarshal(ncfg.Config, &f); err != nil {
				return fmt.Errorf("%s: %s, %s: %s", errRetrieveCfg, ncfg.Type, name, err)
			}
			ogns, err := opsgenie.NewOpsGenieNotificationSender(
				f.Zone,
				f.APIKey,
				f.Timeout,
				f.PersistAlert,
			)
			if err != nil {
				return fmt.Errorf("failed to instantiate opsgenie notification sender: %s", err)
			}
			notify.RegisterSender(name, ogns, ncfg.DefaultNotificationStrategy, ncfg.TemplateNotificationStrategies)

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
				return fmt.Errorf("failed to instantiate tat notification sender: %s", err)
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
			return fmt.Errorf("failed to identify backend type: %s", ncfg.Type)
		}
	}

	notify.RegisterActions(cfg.NotifyActions)

	return nil
}

func validateAndNormalizeNotificationStrategy(ncfg utask.NotifyBackend) (utask.NotifyBackend, error) {
	for action := range ncfg.DefaultNotificationStrategy {
		if !validateActionName(action) {
			return ncfg, fmt.Errorf("invalid action in default_notification_strategy: %q is not allowed value", action)
		}
	}

	for _, action := range []string{notify.TaskValidationKey, notify.TaskStateUpdateKey, notify.TaskStepUpdateKey} {
		if ncfg.DefaultNotificationStrategy == nil {
			ncfg.DefaultNotificationStrategy = make(map[string]string)
		}

		defaultStrategy, ok := ncfg.DefaultNotificationStrategy[action]
		if !ok {
			ncfg.DefaultNotificationStrategy[action] = utask.NotificationStrategyAlways
			defaultStrategy = utask.NotificationStrategyAlways
		}

		switch validateStrategyForAction(action, defaultStrategy) {
		case errNotAllowed:
			return ncfg, fmt.Errorf("invalid default_notification_strategy for action %q: %q is not allowed for this action", action, defaultStrategy)
		case errUnknown:
			return ncfg, fmt.Errorf("invalid default_notification_strategy: %q is not a valid value", ncfg.DefaultNotificationStrategy)
		}

		for action, strats := range ncfg.TemplateNotificationStrategies {
			if !validateActionName(action) {
				return ncfg, fmt.Errorf("invalid action name %q found in notification_strategy: %q is not a valid value", action, action)
			}
			for _, strat := range strats { //nolint:misspell // misspell believes that we wanted to use start instead of strat
				switch validateStrategyForAction(action, strat.NotificationStrategy) {
				case errNotAllowed:
					return ncfg, fmt.Errorf("invalid notification_strategy for templates %#v and action %q: %q is not allowed for this action", strat.Templates, action, strat.NotificationStrategy)
				case errUnknown:
					return ncfg, fmt.Errorf("invalid notification_strategy for templates %#v: %q is not a valid value", strat.Templates, strat.NotificationStrategy)
				}
			}
		}
	}

	return ncfg, nil
}

var (
	errNotAllowed = errors.New("strategy not allowed")
	errUnknown    = errors.New("strategy unknown")
)

func validateStrategyForAction(action, strategy string) error {
	switch strategy {
	case utask.NotificationStrategyAlways, utask.NotificationStrategySilent:
	case utask.NotificationStrategyFailureOnly:
		if action == notify.TaskValidationKey {
			return errNotAllowed
		}
	default:
		return errUnknown
	}

	return nil
}

func validateActionName(action string) bool {
	switch action {
	case notify.TaskValidationKey, notify.TaskStateUpdateKey, notify.TaskStepUpdateKey:
		return true
	default:
		return false
	}
}

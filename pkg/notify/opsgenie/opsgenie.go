package opsgenie

import (
	"github.com/juju/errors"
	"github.com/opsgenie/opsgenie-go-sdk-v2/alert"
	"github.com/opsgenie/opsgenie-go-sdk-v2/client"

	"github.com/ovh/utask/pkg/notify"
)

const (
	// Type represents OpsGenie as notify backend
	Type = "opsgenie"

	// Zone are opsgenie api zones
	ZoneSandbox = "sandbox"
	ZoneDefault = "global"
	ZoneEU      = "eu"
)

// NotificationSender is a notify.NotificationSender implementation
// capable of sending formatted notifications over OpsGenie (https://www.atlassian.com/software/opsgenie)
type NotificationSender struct {
	opsGenieZone   string
	opsGenieAPIKey string
	client         *alert.Client
}

// NewOpsGenieNotificationSender instantiates a NotificationSender
func NewOpsGenieNotificationSender(zone, apikey string) (*NotificationSender, error) {
	zonesToApiUrls := map[string]client.ApiUrl{
		ZoneDefault: client.API_URL,
		ZoneEU:      client.API_URL_EU,
		ZoneSandbox: client.API_URL_SANDBOX,
	}
	apiUrl, present := zonesToApiUrls[zone]
	if !present {
		return nil, errors.NotFoundf("opsgenie zone %q", zone)
	}
	client, err := alert.NewClient(&client.Config{
		ApiKey:         apikey,
		OpsGenieAPIURL: apiUrl,
	})
	if err != nil {
		return nil, err
	}
	return &NotificationSender{
		opsGenieZone:   zone,
		opsGenieAPIKey: apikey,
		client:         client,
	}, nil
}

// Send dispatches a notify.Message to OpsGenie
func (ns *NotificationSender) Send(m *notify.Message, name string) {
	req := &alert.CreateAlertRequest{
		Message:     m.MainMessage,
		Description: m.MainMessage,
		Details:     m.Fields,
	}

	_, err := ns.client.Create(nil, req)
	if err != nil {
		notify.WrappedSendError(Type, err.Error())
		return
	}
}

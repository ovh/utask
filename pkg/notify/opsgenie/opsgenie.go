package opsgenie

import (
	"context"
	"time"

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
	opsGenieZone    string
	opsGenieAPIKey  string
	opsGenieTimeout time.Duration
	client          *alert.Client
}

// NewOpsGenieNotificationSender instantiates a NotificationSender
func NewOpsGenieNotificationSender(zone, apikey, timeout string) (*NotificationSender, error) {
	zonesToAPIUrls := map[string]client.ApiUrl{
		ZoneDefault: client.API_URL,
		ZoneEU:      client.API_URL_EU,
		ZoneSandbox: client.API_URL_SANDBOX,
	}
	apiURL, present := zonesToAPIUrls[zone]
	if !present {
		return nil, errors.NotFoundf("opsgenie zone %q", zone)
	}
	client, err := alert.NewClient(&client.Config{
		ApiKey:         apikey,
		OpsGenieAPIURL: apiURL,
	})
	if err != nil {
		return nil, err
	}
	timeoutDuration := 30 * time.Second
	if timeout != "" {
		timeoutDuration, err = time.ParseDuration(timeout)
		if err != nil {
			return nil, err
		}
	}
	return &NotificationSender{
		opsGenieZone:    zone,
		opsGenieAPIKey:  apikey,
		opsGenieTimeout: timeoutDuration,
		client:          client,
	}, nil
}

// Send dispatches a notify.Message to OpsGenie
func (ns *NotificationSender) Send(m *notify.Message, name string) {
	req := &alert.CreateAlertRequest{
		Message:     m.MainMessage,
		Description: m.MainMessage,
		Details:     m.Fields,
	}

	ctx, cancel := context.WithTimeout(context.Background(), ns.opsGenieTimeout)
	defer cancel()

	_, err := ns.client.Create(ctx, req)
	if err != nil {
		notify.WrappedSendError(Type, err.Error())
		return
	}
}

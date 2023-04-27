package opsgenie

import (
	"context"
	"encoding/json"
	"time"

	"github.com/juju/errors"
	"github.com/opsgenie/opsgenie-go-sdk-v2/alert"
	"github.com/opsgenie/opsgenie-go-sdk-v2/client"

	"github.com/ovh/utask/models/task"
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
func (ns *NotificationSender) Send(msg *notify.Message, name string) {
	ctx, cancel := context.WithTimeout(context.Background(), ns.opsGenieTimeout)
	defer cancel()

	var err error

	// Generate an alias to support alert deduplication
	// cf. https://support.atlassian.com/opsgenie/docs/what-is-alert-de-duplication/
	alias := msg.TaskID()

	if msg.TaskState() == task.StateDone {
		_, err = ns.client.Close(ctx, &alert.CloseAlertRequest{
			IdentifierType:  alert.ALIAS,
			IdentifierValue: alias,
		})
	} else {
		req := &alert.CreateAlertRequest{
			Message:     msg.MainMessage,
			Description: msg.MainMessage,
			Details:     msg.Fields,
			Alias:       alias,
		}
		msgContent, _ := json.Marshal(msg.Fields)
		if msgContent != nil {
			req.Note = string(msgContent)
		}
		_, err = ns.client.Create(ctx, req)
	}
	if err != nil {
		notify.WrappedSendError(err, msg, Type, name)
	}
}

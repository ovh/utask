package opsgenie

import (
	"context"
	"time"

	"github.com/juju/errors"
	"github.com/loopfz/gadgeto/zesty"
	"github.com/opsgenie/opsgenie-go-sdk-v2/alert"
	"github.com/opsgenie/opsgenie-go-sdk-v2/client"

	"github.com/ovh/utask"
	"github.com/ovh/utask/models/task"
	"github.com/ovh/utask/pkg/notify"
)

const (
	// Type represents OpsGenie as notify backend
	Type = "opsgenie"

	MetadataDBKey = "notify_opsgenie"

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
	dbp             zesty.DBProvider
}

// NewOpsGenieNotificationSender instantiates a NotificationSender
func NewOpsGenieNotificationSender(zone, apikey, timeout string, persistDB bool) (*NotificationSender, error) {
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
	var dbp zesty.DBProvider
	if persistDB {
		dbp, err = zesty.NewDBProvider(utask.DBName)
		if err != nil {
			return nil, err
		}
	}
	return &NotificationSender{
		opsGenieZone:    zone,
		opsGenieAPIKey:  apikey,
		opsGenieTimeout: timeoutDuration,
		client:          client,
		dbp:             dbp,
	}, nil
}

type Metadata struct {
	AlertID string `json:"alertId,omitempty"`
	Status  string `json:"status,omitempty"`
}

func (ns *NotificationSender) getMetadata(taskID string) (*task.Metadata[*Metadata], error) {
	if ns.dbp != nil && taskID != "" {
		return task.LoadMetadataFromTaskIDAndKey[*Metadata](ns.dbp, taskID, MetadataDBKey)
	}
	return nil, nil
}

func (ns *NotificationSender) sendAndStoreMetadata(msg *notify.Message, name string, m *task.Metadata[*Metadata]) error {
	req := &alert.CreateAlertRequest{
		Message:     msg.MainMessage,
		Description: msg.MainMessage,
		Details:     msg.Fields,
	}

	ctx, cancel := context.WithTimeout(context.Background(), ns.opsGenieTimeout)
	defer cancel()

	res, err := ns.client.Create(ctx, req)
	if err != nil {
		return err
	}
	alert, err := res.RetrieveStatus(ctx)
	if err != nil {
		return err
	}
	if ns.dbp != nil && msg.TaskID() != "" {
		if m == nil {
			_, err = task.CreateMetadata(ns.dbp, msg.TaskID(), MetadataDBKey, &Metadata{
				AlertID: alert.AlertID,
				Status:  alert.Status,
			})
			if err != nil {
				notify.WrappedSendError(err, msg, Type, name)
			}
		} else {
			m.Value = &Metadata{
				AlertID: alert.AlertID,
				Status:  alert.Status,
			}
			if err = task.UpdateMetadata(ns.dbp, m); err != nil {
				notify.WrappedSendError(err, msg, Type, name)
			}
		}
	}
	return nil
}

func (ns *NotificationSender) updateOrCloseAlertStatus(msg *notify.Message, name string, m *task.Metadata[*Metadata]) error {
	ctx, cancel := context.WithTimeout(context.Background(), ns.opsGenieTimeout)

	cur, err := ns.client.Get(ctx, &alert.GetAlertRequest{
		IdentifierType:  alert.ALERTID,
		IdentifierValue: m.Value.AlertID,
	})
	if err != nil {
		defer cancel()
		return err
	}

	ctx, cancel = context.WithTimeout(context.Background(), ns.opsGenieTimeout)
	defer cancel()

	var res *alert.AsyncAlertResult
	if msg.Fields["state"] == task.StateDone {
		if cur.Status == "closed" {
			if msg.Fields["state"] != task.StateDone {
				// Current task in metadata is closed while the task is not done. Create another alert
				res, err = ns.client.Create(ctx, &alert.CreateAlertRequest{
					Message:     msg.MainMessage,
					Description: msg.MainMessage,
					Details:     msg.Fields,
				})
			} else {
				// Update alert details
				res, err = ns.client.AddDetails(ctx, &alert.AddDetailsRequest{
					IdentifierType:  alert.ALERTID,
					IdentifierValue: m.Value.AlertID,
					Details:         msg.Fields,
				})
			}
		} else {
			// alert is not closed. Try to close it
			res, err = ns.client.Close(ctx, &alert.CloseAlertRequest{
				IdentifierType:  alert.ALERTID,
				IdentifierValue: m.Value.AlertID,
			})
		}
	} else {
		// Update alert details
		res, err = ns.client.AddDetails(ctx, &alert.AddDetailsRequest{
			IdentifierType:  alert.ALERTID,
			IdentifierValue: m.Value.AlertID,
			Details:         msg.Fields,
		})
	}
	if err != nil {
		return err
	}
	alert, err := res.RetrieveStatus(ctx)
	if err == nil {
		m.Value = &Metadata{
			AlertID: alert.AlertID,
			Status:  alert.Status,
		}
		if err := task.UpdateMetadata(ns.dbp, m); err != nil {
			notify.WrappedSendError(err, msg, Type, name)
		}
	}
	return nil
}

// Send dispatches a notify.Message to OpsGenie
func (ns *NotificationSender) Send(msg *notify.Message, name string) {
	m, err := ns.getMetadata(msg.TaskID())
	if err != nil || m == nil || m != nil && m.Value == nil {
		if err != nil {
			notify.WrappedSendError(err, msg, Type, name)
		}
		if err := ns.sendAndStoreMetadata(msg, name, m); err != nil {
			notify.WrappedSendError(err, msg, Type, name)
		}
	} else {
		if err := ns.updateOrCloseAlertStatus(msg, name, m); err != nil {
			notify.WrappedSendError(err, msg, Type, name)
			if err := ns.sendAndStoreMetadata(msg, name, m); err != nil {
				notify.WrappedSendError(err, msg, Type, name)
			}
		}
	}
}

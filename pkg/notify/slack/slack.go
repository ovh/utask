package slack

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/ovh/utask/pkg/notify"
)

const (
	// Type represents Slack as notify backend
	Type string = "slack"
)

// NotificationSender is a notify.NotificationSender implementation
// capable of sending formatted notifications over Slack
type NotificationSender struct {
	webhookURL string
	httpClient *http.Client
}

type slackFormattedBody struct {
	Text   string `json:"text,omitempty"`
	Blocks []struct {
		Type    string `json:"type"`
		BlockID string `json:"block_id"`
		Text    struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"text"`
		Accessory struct {
			Type     string `json:"type"`
			ImageURL string `json:"image_url"`
			AltText  string `json:"alt_text"`
		} `json:"accessory"`
	} `json:"blocks,omitempty"`
}

// NewSlackNotificationSender instantiates a NotificationSender
func NewSlackNotificationSender(webhookURL string) *NotificationSender {
	return &NotificationSender{
		webhookURL: webhookURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// Send dispatches a notify.Payload to Slack
func (sn *NotificationSender) Send(p notify.Payload) {
	slackfb := &slackFormattedBody{
		Text: p.Message(),
	}

	slackBody, _ := json.Marshal(slackfb)

	req, err := http.NewRequest(http.MethodPost, sn.webhookURL, bytes.NewBuffer(slackBody))
	if err != nil {
		notify.WrappedSendError(Type, err.Error())
		return
	}

	req.Header.Add("Content-Type", "application/json")
	resp, err := sn.httpClient.Do(req)
	if err != nil {
		notify.WrappedSendError(Type, err.Error())
		return
	}

	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	if buf.String() != "ok" {
		notify.WrappedSendError(Type, "Non-ok response returned from Slack")
		return
	}
}

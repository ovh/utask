package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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

type formattedSlackRequest struct {
	Blocks []blockSlackRequest `json:"blocks"`
}

type blockSlackRequest struct {
	Type     string                `json:"type"`
	Text     textSlackRequest      `json:"text,omitempty"`
	Fields   []fieldSlackRequest   `json:"fields,omitempty"`
	Elements []elementSlackRequest `json:"elements,omitempty"`
}

type elementSlackRequest fieldSlackRequest
type textSlackRequest fieldSlackRequest
type fieldSlackRequest struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// NewSlackNotificationSender instantiates a NotificationSender
func NewSlackNotificationSender(webhookURL string) *NotificationSender {
	return &NotificationSender{
		webhookURL: webhookURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// Send dispatches a notify.Message to Slack
func (sn *NotificationSender) Send(m *notify.Message, name string) {
	slackfb := formatSendRequest(m, name)

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

func formatSendRequest(m *notify.Message, name string) *formattedSlackRequest {
	var fsr formattedSlackRequest

	sec := "section"
	mrk := "mrkdwn"

	fsr.Blocks = make([]blockSlackRequest, 4)

	// First line title
	fsr.Blocks[0].Type = sec
	fsr.Blocks[0].Text.Type = mrk
	fsr.Blocks[0].Text.Text = m.MainMessage

	// Fields
	fsr.Blocks[1].Type = sec
	fsr.Blocks[1].Fields = make([]fieldSlackRequest, 0)
	for key, value := range m.Fields {
		if len(value) > 0 {
			trimStr := strings.Replace(key, "_", " ", -1)
			fsr.Blocks[1].Fields = append(
				fsr.Blocks[1].Fields,
				fieldSlackRequest{
					Type: mrk,
					Text: fmt.Sprintf("*%s:*\n%s", strings.Title(trimStr), value),
				})
		}
	}

	// Separator
	fsr.Blocks[2].Type = "divider"

	// Sent context
	fsr.Blocks[3].Type = "context"
	fsr.Blocks[3].Elements = make([]elementSlackRequest, 1)
	fsr.Blocks[3].Elements[0].Type = mrk
	fsr.Blocks[3].Elements[0].Text = fmt.Sprintf("🚀 Sent from %s", name)

	return &fsr
}

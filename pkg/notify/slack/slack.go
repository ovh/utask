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
	Blocks []struct {
		Type string `json:"type"`
		Text struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"text,omitempty"`
		Fields []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"fields,omitempty"`
		Elements []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"elements,omitempty"`
	} `json:"blocks"`
}

// NewSlackNotificationSender instantiates a NotificationSender
func NewSlackNotificationSender(webhookURL string) *NotificationSender {
	return &NotificationSender{
		webhookURL: webhookURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// Send dispatches a notify.Payload to Slack
func (sn *NotificationSender) Send(p notify.Payload, name string) {
	slackfb := formatSendRequest(p.MessageFields(), name)

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

func formatSendRequest(tsu *notify.TaskStateUpdate, name string) *formattedSlackRequest {
	var fsr formattedSlackRequest

	sec := "section"
	mrk := "mrkdwn"

	// First line title
	fsr.Blocks[0].Type = sec
	fsr.Blocks[0].Text.Type = mrk
	fsr.Blocks[0].Text.Text = tsu.Title

	// Fields
	fid := 0
	// ID
	fsr.Blocks[1].Type = sec
	fsr.Blocks[1].Fields[fid].Type = mrk
	fsr.Blocks[1].Fields[0].Text = fmt.Sprintf("*ID:*\n%s", tsu.PublicID)
	fid++

	// Status
	fsr.Blocks[1].Type = sec
	fsr.Blocks[1].Fields[fid].Type = mrk
	fsr.Blocks[1].Fields[fid].Text = fmt.Sprintf("*Status:*\n%s", tsu.State)
	fid++

	// Template
	fsr.Blocks[1].Type = sec
	fsr.Blocks[1].Fields[fid].Type = mrk
	fsr.Blocks[1].Fields[fid].Text = fmt.Sprintf("*Template:*\n%s", tsu.TemplateName)
	fid++

	// Steps
	fsr.Blocks[1].Type = sec
	fsr.Blocks[1].Fields[fid].Type = mrk
	fsr.Blocks[1].Fields[fid].Text = fmt.Sprintf("*Steps:*\n%d/%d", tsu.StepsDone, tsu.StepsTotal)
	fid++

	// Requester
	if tsu.RequesterUsername != "" {
		fsr.Blocks[1].Type = sec
		fsr.Blocks[1].Fields[fid].Type = mrk
		fsr.Blocks[1].Fields[fid].Text = fmt.Sprintf("*Requester:*\n%s", tsu.RequesterUsername)
		fid++
	}

	// Resolver
	if tsu.ResolverUsername != nil && *tsu.ResolverUsername != "" {
		fsr.Blocks[1].Type = sec
		fsr.Blocks[1].Fields[fid].Type = mrk
		fsr.Blocks[1].Fields[fid].Text = fmt.Sprintf("*Resolver:*\n%s", *tsu.ResolverUsername)
		fid++
	}

	// Potential Resolvers
	if tsu.PotentialResolvers != nil && len(tsu.PotentialResolvers) > 0 {
		fsr.Blocks[1].Type = sec
		fsr.Blocks[1].Fields[fid].Type = mrk
		fsr.Blocks[1].Fields[fid].Text = fmt.Sprintf("*Potential Resolvers:*\n%s", strings.Join(tsu.PotentialResolvers, " "))
	}

	// Separator
	fsr.Blocks[2].Type = "divider"

	// Sent context
	fsr.Blocks[3].Type = "context"
	fsr.Blocks[3].Elements[0].Type = mrk
	fsr.Blocks[3].Elements[0].Text = fmt.Sprintf("🚀 Sent from %s", name)

	return &fsr
}

package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/ovh/utask/pkg/notify"
)

const (
	// Type represents Webhook as notify backend
	Type string = "webhook"
)

// NotificationSender is a notify.NotificationSender implementation
// capable of sending notifications to a webhook
type NotificationSender struct {
	webhookURL string
	username   string
	password   string
	headers    map[string]string
	httpClient *http.Client
}

// NewWebhookNotificationSender instantiates a NotificationSender
func NewWebhookNotificationSender(webhookURL, username, password string, headers map[string]string) *NotificationSender {
	return &NotificationSender{
		webhookURL: webhookURL,
		username:   username,
		password:   password,
		headers:    headers,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// Send is the implementation for triggering a webhook to send the notification
func (w *NotificationSender) Send(m *notify.Message, name string) {
	msg := map[string]string{
		"message": m.MainMessage,
	}

	for k, v := range m.Fields {
		msg[k] = v
	}

	b, err := json.Marshal(msg)
	if err != nil {
		fmt.Println(err)
		return
	}

	req, err := http.NewRequest("POST", w.webhookURL, bytes.NewBuffer(b))
	if err != nil {
		fmt.Println(err)
		return
	}

	for k, v := range w.headers {
		req.Header.Set(k, v)
	}

	if w.username != "" && w.password != "" {
		req.SetBasicAuth(w.username, w.password)
	}

	res, err := w.httpClient.Do(req)
	if err != nil {
		log.Println(err)
		return
	}

	defer res.Body.Close()

	if res.StatusCode >= 400 {
		resBody, err := ioutil.ReadAll(res.Body)
		log.Printf("failed to send notification using %q: backend returned with status code %d\n", name, res.StatusCode)
		if err == nil {
			log.Printf("webhook %q response body: %s\n", name, string(resBody))
		}
		return
	}
}

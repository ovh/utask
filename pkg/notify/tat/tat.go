package tat

import (
	"fmt"

	tatlib "github.com/ovh/tat"

	"github.com/ovh/utask/pkg/notify"
)

const (
	// Type represents Tat as notify backend
	Type string = "tat"
)

var labelColors = []string{
	"#2E0927",
	"#FF8C00",
	"#04756F",
	"#D90000",
	"#FF2D00",
	"#0080FF",
}

// NotificationSender is a notify.NotificationSender implementation
// capable of sending formatted notifications over TaT (github.com/ovh/tat)
type NotificationSender struct {
	tatURL      string
	tatUser     string
	tatPassword string
	tatTopic    string
}

// NewTatNotificationSender instantiates a NotificationSender
func NewTatNotificationSender(url, user, pass, topic string) (*NotificationSender, error) {
	tn := &NotificationSender{
		tatURL:      url,
		tatUser:     user,
		tatPassword: pass,
		tatTopic:    topic,
	}
	_, err := tn.spawnTatClient()
	if err != nil {
		return nil, err
	}
	return tn, nil
}

// Send dispatches a notify.Message to TaT
func (tn *NotificationSender) Send(m *notify.Message, name string) {
	client, err := tn.spawnTatClient()
	if err != nil {
		fmt.Println(err)
		return
	}

	labels := formatSendRequest(m, name)

	_, err = client.MessageAdd(
		tatlib.MessageJSON{
			Text:   m.MainMessage,
			Labels: labels,
			Topic:  tn.tatTopic,
		},
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	// TODO create message for task creation
	// TODO update message afterwards, selecting on #id:xxx
}

func (tn *NotificationSender) spawnTatClient() (*tatlib.Client, error) {
	return tatlib.NewClient(tatlib.Options{
		URL:      tn.tatURL,
		Username: tn.tatUser,
		Password: tn.tatPassword,
	})
}

func taskLabels(fields []string) []tatlib.Label {
	l := make([]tatlib.Label, 0)
	for i, f := range fields {
		l = append(l, tatlib.Label{
			Text:  f,
			Color: labelColors[i%len(labelColors)],
		})
	}
	return l
}

func formatSendRequest(m *notify.Message, name string) []tatlib.Label {
	labels := []string{}

	for key, value := range m.Fields {
		if len(value) > 0 {
			labels = append(labels,
				fmt.Sprintf("%s:%s", key, value))
		}
	}

	labels = append(labels, fmt.Sprintf("backend_name:%s", name))

	return taskLabels(labels)
}

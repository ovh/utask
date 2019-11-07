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

// Send dispatches a notify.Payload to TaT
func (tn *NotificationSender) Send(p notify.Payload) {
	client, err := tn.spawnTatClient()
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = client.MessageAdd(
		tatlib.MessageJSON{
			Text:   p.Message(),
			Labels: taskLabels(p.Fields()),
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

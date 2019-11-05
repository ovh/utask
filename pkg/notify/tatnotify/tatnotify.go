package tatnotify

import (
	"fmt"

	"github.com/ovh/tat"

	"github.com/ovh/utask/pkg/notify"
)

const (
	// Tat represents tat as notify backend
	Tat string = "tat"
)

var labelColors = []string{
	"#2E0927",
	"#FF8C00",
	"#04756F",
	"#D90000",
	"#FF2D00",
}

// TatNotificationSender is a notify.NotificationSender implementation
// capable of sending formatted notifications over TaT (github.com/ovh/tat)
type TatNotificationSender struct {
	tatURL      string
	tatUser     string
	tatPassword string
	tatTopic    string
}

// NewTatNotificationSender instantiates a TatNotificationSender
func NewTatNotificationSender(url, user, pass, topic string) (*TatNotificationSender, error) {
	tn := &TatNotificationSender{
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
func (tn *TatNotificationSender) Send(p notify.Payload) {
	client, err := tn.spawnTatClient()
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = client.MessageAdd(
		tat.MessageJSON{
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

func (tn *TatNotificationSender) spawnTatClient() (*tat.Client, error) {
	return tat.NewClient(tat.Options{
		URL:      tn.tatURL,
		Username: tn.tatUser,
		Password: tn.tatPassword,
	})
}

func taskLabels(fields []string) []tat.Label {
	l := make([]tat.Label, 0)
	for i, f := range fields {
		l = append(l, tat.Label{
			Text:  f,
			Color: labelColors[i%len(labelColors)],
		})
	}
	return l
}

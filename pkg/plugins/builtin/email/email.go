package email

import (
	"bytes"
	"fmt"
	"net/smtp"
	"strings"
	"text/template"

	"github.com/ovh/utask/pkg/plugins/taskplugin"
)

// the email plugin send email
var (
	Plugin = taskplugin.New("email", "0.1", exec,
		taskplugin.WithConfig(validConfig, Config{}),
	)
)

// Config is the configuration needed to send an email
type Config struct {
	SMTPUsername string
	SMTPPassword string
	SMTPPort     uint
	SMTPHost     string
	From         string
	To           []string
	Subject      string
	Body         string
}

const emailTemplate = `From: {{.From}}<br />
To: {{.To}}<br />
Subject: {{.Subject}}<br />
MIME-version: 1.0<br />
Content-Type: text/html; charset="UTF-8"<br />
<br />
{{.Body}}`

func validConfig(config interface{}) error {
	return nil
}

func exec(stepName string, config interface{}, ctx interface{}) (interface{}, interface{}, error) {
	cfg := config.(*Config)

	parameters := struct {
		From    string
		To      string
		Subject string
		Body    string
	}{
		cfg.From,
		strings.Join(cfg.To, ","),
		cfg.Subject,
		cfg.Body,
	}

	buffer := new(bytes.Buffer)

	template := template.Must(template.New("emailTemplate").Parse(emailTemplate))
	template.Execute(buffer, &parameters)

	auth := smtp.PlainAuth("", cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPHost)

	err := smtp.SendMail(
		fmt.Sprintf("%s:%d", cfg.SMTPHost, int(cfg.SMTPPort)),
		auth,
		cfg.SMTPUsername,
		cfg.To,
		buffer.Bytes())
	if err != nil {
		return nil, nil, fmt.Errorf("Send email failed: %s", err.Error())
	}

	return nil, nil, nil
}

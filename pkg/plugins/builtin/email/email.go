package email

import (
	"bytes"
	"errors"
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
	SMTPUsername string   `json:"smtp_username"`
	SMTPPassword string   `json:"smtp_password"`
	SMTPPort     uint     `json:"smtp_port,omitempty"`
	SMTPHostname string   `json:"smtp_hostname"`
	From         string   `json:"from"`
	To           []string `json:"to"`
	Subject      string   `json:"subject"`
	Body         string   `json:"body"`
}

const emailTemplate = `From: {{.From}}<br />
To: {{.To}}<br />
Subject: {{.Subject}}<br />
MIME-version: 1.0<br />
Content-Type: text/html; charset="UTF-8"<br />
<br />
{{.Body}}`

func validConfig(config interface{}) error {
	cfg := config.(*Config)

	if cfg.SMTPUsername == "" {
		return errors.New("smtp_username is missing")
	}
	if cfg.SMTPPassword == "" {
		return errors.New("smtp_password is missing")
	}
	if cfg.From == "" {
		return errors.New("from is missing")
	}
	if len(cfg.To) == 0 {
		return errors.New("to is missing")
	}
	if cfg.SMTPHostname == "" {
		return errors.New("smtp_hostname is missing")
	}
	if cfg.Subject == "" {
		return errors.New("subject is missing")
	}
	if cfg.Body == "" {
		return errors.New("body is missing")
	}

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

	auth := smtp.PlainAuth("", cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPHostname)

	port := cfg.SMTPPort
	if port == 0 {
		port = 25
	}

	err := smtp.SendMail(
		fmt.Sprintf("%s:%d", cfg.SMTPHostname, int(cfg.SMTPPort)),
		auth,
		cfg.SMTPUsername,
		cfg.To,
		buffer.Bytes())
	if err != nil {
		return nil, nil, fmt.Errorf("Send email failed: %s", err.Error())
	}

	return nil, nil, nil
}

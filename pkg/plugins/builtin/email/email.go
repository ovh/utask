package email

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"

	"github.com/ovh/utask/pkg/plugins/taskplugin"
	mail "gopkg.in/mail.v2"
)

// the email plugin send email
var (
	Plugin = taskplugin.New("email", "0.2", exec,
		taskplugin.WithConfig(validConfig, Config{}),
	)
)

// Config is the configuration needed to send an email
type Config struct {
	SMTPUsername      string   `json:"smtp_username"`
	SMTPPassword      string   `json:"smtp_password"`
	SMTPPort          uint16   `json:"smtp_port"`
	SMTPHostname      string   `json:"smtp_hostname"`
	SMTPSkipTLSVerify bool     `json:"smtp_skip_tls_verify,omitempty"`
	FromAddress       string   `json:"from_address"`
	FromName          string   `json:"from_name,omitempty"`
	To                []string `json:"to"`
	Subject           string   `json:"subject"`
	Body              string   `json:"body"`
}

func validConfig(config interface{}) error {
	cfg := config.(*Config)

	if cfg.SMTPUsername == "" {
		return errors.New("smtp_username is missing")
	}
	if cfg.SMTPPassword == "" {
		return errors.New("smtp_password is missing")
	}
	if cfg.SMTPPort == 0 {
		return errors.New("smtp_port is missing")
	}
	if cfg.SMTPHostname == "" {
		return errors.New("smtp_hostname is missing")
	}
	if cfg.FromAddress == "" {
		return errors.New("from_address is missing")
	}
	if len(cfg.To) == 0 {
		return errors.New("to is missing")
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

	message := mail.NewMessage()

	recipients := make([]string, len(cfg.To))
	for i, recipient := range cfg.To {
		recipients[i] = message.FormatAddress(recipient, "")
	}

	message.SetAddressHeader("From", cfg.FromAddress, cfg.FromName)
	message.SetHeader("To", recipients...)
	message.SetHeader("Subject", cfg.Subject)
	message.SetBody(
		http.DetectContentType([]byte(cfg.Body)),
		cfg.Body,
	)

	d := mail.NewDialer(cfg.SMTPHostname, int(cfg.SMTPPort), cfg.SMTPUsername, cfg.SMTPPassword)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: cfg.SMTPSkipTLSVerify}
	if err := d.DialAndSend(message); err != nil {
		fmt.Errorf("Send email failed: %s", err.Error())
	}

	// to reuse configuration
	parameters := struct {
		FromAddress string   `json:"from_address"`
		FromName    string   `json:"from_name"`
		To          []string `json:"to"`
		Subject     string   `json:"subject"`
		Body        string   `json:"body"`
	}{
		cfg.FromAddress,
		cfg.FromName,
		cfg.To,
		cfg.Subject,
		cfg.Body,
	}

	return &parameters, nil, nil
}

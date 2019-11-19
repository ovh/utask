package email

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	mail "gopkg.in/mail.v2"

	"github.com/ovh/utask/pkg/plugins/taskplugin"
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
	SMTPPort          string   `json:"smtp_port"`
	SMTPHostname      string   `json:"smtp_hostname"`
	SMTPSkipTLSVerify string   `json:"smtp_skip_tls_verify,omitempty"`
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

	smtpp, err := strconv.Atoi(cfg.SMTPPort)
	if smtpp <= 0 || err != nil || cfg.SMTPPort == "" {
		return fmt.Errorf("smtp_port is missing or wrong %s", err)
	}

	if cfg.SMTPHostname == "" {
		return errors.New("smtp_hostname is missing")
	}

	if cfg.SMTPSkipTLSVerify != "" {
		if _, err := strconv.ParseBool(cfg.SMTPSkipTLSVerify); err != nil {
			return fmt.Errorf("smtp_skip_tls_verify is wrong %s", err)
		}
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

	port, _ := strconv.Atoi(cfg.SMTPPort)
	skipTLS, _ := strconv.ParseBool(cfg.SMTPSkipTLSVerify)

	d := mail.NewDialer(cfg.SMTPHostname, port, cfg.SMTPUsername, cfg.SMTPPassword)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: skipTLS}
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

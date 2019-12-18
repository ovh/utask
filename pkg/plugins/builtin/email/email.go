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

type mailParameters struct {
	FromAddress string   `json:"from_address"`
	FromName    string   `json:"from_name,omitempty"`
	To          []string `json:"to"`
	Subject     string   `json:"subject"`
	Body        string   `json:"body"`
}

// Config is the configuration needed to send an email
type Config struct {
	SMTPUsername      string `json:"smtp_username"`
	SMTPPassword      string `json:"smtp_password"`
	SMTPPort          string `json:"smtp_port"`
	SMTPHostname      string `json:"smtp_hostname"`
	SMTPSkipTLSVerify string `json:"smtp_skip_tls_verify,omitempty"`
	mailParameters
}

func validConfig(config interface{}) error {
	cfg := config.(*Config)

	if cfg.SMTPUsername == "" {
		return errors.New("smtp_username is missing")
	}

	if cfg.SMTPPassword == "" {
		return errors.New("smtp_password is missing")
	}

	if cfg.SMTPPort == "" {
		return errors.New("smtp_password is missing")
	}

	if _, err := strconv.ParseUint(cfg.SMTPPort, 10, 64); err != nil {
		return fmt.Errorf("can't parse smtp_port field %q: %s", cfg.SMTPPort, err.Error())
	}

	if cfg.SMTPHostname == "" {
		return errors.New("smtp_hostname is missing")
	}

	if cfg.SMTPSkipTLSVerify != "" {
		if _, err := strconv.ParseBool(cfg.SMTPSkipTLSVerify); err != nil {
			return fmt.Errorf("can't parse smtp_skip_tls_verify field %q: %s", cfg.SMTPSkipTLSVerify, err)
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

	message.SetAddressHeader("From", cfg.FromAddress, cfg.FromName)
	message.SetHeader("To", cfg.To...)
	message.SetHeader("Subject", cfg.Subject)
	message.SetBody(
		http.DetectContentType([]byte(cfg.Body)),
		cfg.Body,
	)

	// port and skipTLS already checked at validConfig() lvl
	// values must be correct so errors are not evaluated
	port, _ := strconv.ParseUint(cfg.SMTPPort, 10, 64)     // no defaults, must be set by user
	skipTLS, _ := strconv.ParseBool(cfg.SMTPSkipTLSVerify) // defaults to false

	d := mail.NewDialer(cfg.SMTPHostname, int(port), cfg.SMTPUsername, cfg.SMTPPassword)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: skipTLS, ServerName: cfg.SMTPHostname}
	if err := d.DialAndSend(message); err != nil {
		return nil, nil, fmt.Errorf("can't send email: %s", err.Error())
	}

	return &cfg.mailParameters, nil, nil
}

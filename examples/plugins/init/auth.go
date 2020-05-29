package main

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"

	"github.com/juju/errors"
	"github.com/ovh/utask"
	"github.com/ovh/utask/pkg/plugins"
)

var Plugin = NewSillyAuth()

type SillyAuth struct{}

func NewSillyAuth() SillyAuth {
	return SillyAuth{}
}

// Init function will be called at uTask startup, and will configure uTask instance with a custom authentication provider
func (sa SillyAuth) Init(s *plugins.Service) error {
	utaskCfg, err := utask.Config(s.Store)
	if err != nil {
		return errors.Annotate(err, "unable to retrieve utaskCfg")
	}

	// add a custom ID provider
	auth, err := authProvider(utaskCfg.AdminUsernames)
	if err != nil {
		return fmt.Errorf("Failed to load auth provider: %s", err)
	}
	s.Server.WithAuth(auth)

	return nil
}

func (sa SillyAuth) Description() string {
	return `This plugin will configure a silly authentication system, based on sha256 of the username and source IP address.`
}

func authProvider(admins []string) (func(r *http.Request) (string, error), error) {
	return func(r *http.Request) (string, error) {
		remoteUser, remotePass, ok := r.BasicAuth()
		if !ok {
			return "", errors.Forbiddenf("missing or invalid Authorization header")
		}

		h := sha256.New()
		h.Write([]byte(remoteUser + "+foobar"))
		if sum := fmt.Sprintf("%x", h.Sum(nil)); sum != remotePass {
			return "", errors.Forbiddenf("invalid authentication")
		}

		for _, admin := range admins {
			if admin != remoteUser {
				continue
			}
			if !strings.HasPrefix(r.RemoteAddr, "127.0.0.1:") {
				return "", errors.Forbiddenf("admin account should authenticate from localhost")
			}
		}

		return remoteUser, nil
	}, nil
}

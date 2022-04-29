package plugincallback

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_validConfig(t *testing.T) {
	// create step - valid
	cfg := CallbackStepConfig{
		Action: "create",
	}
	cfgJSON, err := json.Marshal(cfg)
	assert.NoError(t, err)
	assert.NoError(t, Plugin.ValidConfig(json.RawMessage(""), json.RawMessage(cfgJSON)))

	// create step - valid with JSON schema
	cfg.BodySchema = "{}"
	cfgJSON, err = json.Marshal(cfg)
	assert.NoError(t, err)
	assert.NoError(t, Plugin.ValidConfig(json.RawMessage(""), json.RawMessage(cfgJSON)))

	// wait step - invalid missing id
	cfg = CallbackStepConfig{
		Action: "wait",
	}
	cfgJSON, err = json.Marshal(cfg)
	assert.NoError(t, err)
	assert.Error(t, Plugin.ValidConfig(json.RawMessage(""), json.RawMessage(cfgJSON)), "missing \"id\" parameter")

	// wait step - valid
	cfg.ID = "foo"
	cfgJSON, err = json.Marshal(cfg)
	assert.NoError(t, err)
	assert.NoError(t, Plugin.ValidConfig(json.RawMessage(""), json.RawMessage(cfgJSON)))

	// unknown action
	cfg = CallbackStepConfig{
		Action: "foo",
	}
	cfgJSON, err = json.Marshal(cfg)
	assert.NoError(t, err)
	assert.Error(t, Plugin.ValidConfig(json.RawMessage(""), json.RawMessage(cfgJSON)), "invalid action \"foo\"")
}

func Test_buildUrl(t *testing.T) {
	cb := &callback{
		ID:       42,
		PublicID: "foobar",
		Secret:   "s3cr3t",
	}

	Init.cfg.BaseURL = "http://utask.example.com"
	Init.cfg.PathPrefix = ""
	assert.Equal(t, buildUrl(cb), fmt.Sprintf("http://utask.example.com%s/foobar?t=s3cr3t", defaultCallbackPathPrefix))

	Init.cfg.BaseURL = "http://utask.example.com"
	Init.cfg.PathPrefix = "/foo"
	assert.Equal(t, buildUrl(cb), "http://utask.example.com/foo/foobar?t=s3cr3t")

	Init.cfg.BaseURL = "http://utask.example.com"
	Init.cfg.PathPrefix = "/bar/"
	assert.Equal(t, buildUrl(cb), "http://utask.example.com/bar/foobar?t=s3cr3t")

	Init.cfg.BaseURL = "http://utask.example.com"
	Init.cfg.PathPrefix = "/"
	assert.Equal(t, buildUrl(cb), "http://utask.example.com/foobar?t=s3cr3t")
}

package step

import (
	"encoding/json"

	"github.com/juju/errors"
)

// Hook represents the hook data structure
type Hook struct {
	Name            string      `json:"name"`
	Description     string      `json:"description""`
	LongDescription *string     `json:"long_description,omitempty"`
	DocLink         *string     `json:"doc_link,omitempty"`
	Actions         HookActions `json:"actions"`
}

// HookActions is just a binding over []json.RawMessage type
// TODO: switch to Executor
type HookActions []json.RawMessage

// MapAllHooks represents a static map of all hooks
var MapAllHooks map[string]Hook = make(map[string]Hook)

func getHook(name string) (*Hook, error) {
	h, ok := MapAllHooks[name]
	if ok {
		return &h, nil
	}

	return nil, errors.New("hook not found")
}

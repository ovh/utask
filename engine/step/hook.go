package step

import (
	"github.com/juju/errors"
)

// Hook represents the hook data structure
type Hook struct {
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	LongDescription *string                `json:"long_description,omitempty"`
	DocLink         *string                `json:"doc_link,omitempty"`
	Action          Executor               `json:"action"`
	Result          map[string]interface{} `json:"result"`
}

// MapAllHooks represents a static map of all hooks
var MapAllHooks map[string]Hook = make(map[string]Hook)

func getHook(name string) (*Hook, error) {
	h, ok := MapAllHooks[name]
	if ok {
		return &h, nil
	}

	return nil, errors.New("hook not found")
}

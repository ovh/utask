package hook

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/juju/errors"
)

var mapAllHooks map[string]Hook = make(map[string]Hook)

type Hook struct {
	Name            string      `json:"name"`
	Description     string      `json:"description""`
	LongDescription *string     `json:"long_description,omitempty"`
	DocLink         *string     `json:"doc_link,omitempty"`
	Actions         HookActions `json:"actions"`
}

type HookActions []json.RawMessage

// LoadFromDir reads yaml-formatted task templates from a folder
func LoadFromDir(dir string) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("Failed to open template directory %s: %s", dir, err)
	}
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".yaml") {
			continue
		}
		tmpl, err := ioutil.ReadFile(path.Join(dir, file.Name()))
		if err != nil {
			return err
		}
		var h Hook
		if err := yaml.Unmarshal(tmpl, &h); err != nil {
			return fmt.Errorf("failed to unmarshal '%s': '%s'", file.Name(), err)
		}

		mapAllHooks[h.Name] = h
	}
	return nil
}

func GetHook(name string) (*Hook, error) {
	h, ok := mapAllHooks[name]
	if ok {
		return &h, nil
	}

	return nil, errors.New("hook not found")
}

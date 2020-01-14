package hook

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/ghodss/yaml"

	"github.com/ovh/utask/engine/step"
)

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
		var h step.Hook
		if err := yaml.Unmarshal(tmpl, &h); err != nil {
			return fmt.Errorf("failed to unmarshal '%s': '%s'", file.Name(), err)
		}

		step.MapAllHooks[h.Name] = h
	}
	return nil
}

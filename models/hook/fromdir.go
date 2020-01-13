package hook

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/juju/errors"
	"github.com/loopfz/gadgeto/zesty"
	"github.com/sirupsen/logrus"
)

// LoadFromDir reads yaml-formatted task templates
// from a folder and upserts them in database
func LoadFromDir(dbp zesty.DBProvider, dir string) error {
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

		verb := "Created"
		existing, err := LoadFromName(dbp, h.Name)
		if err != nil {
			if !errors.IsNotFound(err) {
				return err
			}
			if err := create(dbp, &h); err != nil {
				return err
			}
		} else {
			verb = "Updated"
			h.ID = existing.ID
			if err := update(dbp, &h); err != nil {
				return err
			}
		}
		logrus.Infof("%s hook '%s'", verb, h.Name)

		mapAllHooks[h.Name] = h
	}
	return nil
}

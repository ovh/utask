package tasktemplate

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
		var tt TaskTemplate
		if err := yaml.Unmarshal(tmpl, &tt); err != nil {
			return fmt.Errorf("failed to unmarshal '%s': '%s'", file.Name(), err)
		}
		verb := "Created"
		existing, err := LoadFromName(dbp, tt.Name)
		if err != nil {
			if !errors.IsNotFound(err) {
				return err
			}
			if _, err := create(dbp, &tt); err != nil {
				return err
			}
		} else {
			verb = "Updated"
			tt.ID = existing.ID
			if err := update(dbp, &tt); err != nil {
				return err
			}
		}
		logrus.Infof("%s task template '%s'", verb, tt.Name)
	}
	return nil
}

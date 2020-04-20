package tasktemplate

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/juju/errors"
	"github.com/loopfz/gadgeto/zesty"
	"github.com/ovh/utask/pkg/templateimport"
	"github.com/sirupsen/logrus"
)

var (
	discoveredTemplates []TaskTemplate = []TaskTemplate{}
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
			return fmt.Errorf("failed to read template '%s': %s", file.Name(), err)
		}
		var tt TaskTemplate
		if err := yaml.Unmarshal(tmpl, &tt); err != nil {
			return fmt.Errorf("failed to unmarshal template '%s': '%s'", file.Name(), err)
		}

		tt.Normalize()

		discoveredTemplates = append(discoveredTemplates, tt)
		templateimport.AddTemplate(tt.Name)
	}

	for _, tt := range discoveredTemplates {
		verb := "Created"
		existing, err := LoadFromName(dbp, tt.Name)
		if err != nil {
			if !errors.IsNotFound(err) {
				return err
			}
			if _, err := create(dbp, &tt); err != nil {
				return fmt.Errorf("failed to create template '%s': %s", tt.Name, err)
			}
		} else {
			verb = "Updated"
			tt.ID = existing.ID
			if err := update(dbp, &tt); err != nil {
				return fmt.Errorf("failed to update template '%s': %s", tt.Name, err)
			}
		}
		logrus.Infof("%s task template '%s'", verb, tt.Name)
	}

	// removing or archiving old task_templates
	var last *string
	currentTemplates := []*TaskTemplate{}
	for {
		taskTemplatesFromDatabase, err := ListTemplates(dbp, true, 100, last)
		if err != nil {
			logrus.Fatalf("unable to remove old templates: %s", err)
		}
		if len(taskTemplatesFromDatabase) == 0 {
			break
		}
		currentTemplates = append(currentTemplates, taskTemplatesFromDatabase...)
		lastI := taskTemplatesFromDatabase[len(taskTemplatesFromDatabase)-1].Name
		last = &lastI
	}

	importedTemplates := templateimport.GetTemplates()
	for _, tt := range currentTemplates {
		found := false
		for _, importedTT := range importedTemplates {
			if tt.Name == importedTT {
				found = true
				break
			}
		}
		if !found {
			if err = tt.Delete(dbp); err == nil {
				logrus.Infof("Deleted task template %q", tt.Name)
				continue
			}
			// unable to delete TaskTemplate, probably some old Tasks still in database, archiving it
			tt, err = LoadFromID(dbp, tt.ID)
			if err != nil {
				return fmt.Errorf("unable to load template %q for archiving: %s", tt.Name, err)
			}
			tt.Hidden = true
			tt.Blocked = true
			if err := update(dbp, tt); err != nil {
				return fmt.Errorf("unable to archive template %q: %s", tt.Name, err)
			}
			logrus.Infof("Archived task template %q", tt.Name)
		}
	}

	templateimport.CleanTemplates()
	return nil
}

package tasktemplate

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/juju/errors"
	"github.com/ovh/utask"
	"github.com/ovh/utask/engine/step"
	"github.com/ovh/utask/engine/values"
	"github.com/ovh/utask/pkg/jsonschema"
	"github.com/ovh/utask/pkg/utils"
)

const (
	variableDoesNotExist = "Variable %s does not exist"
)

var (
	tmplRegex = regexp.MustCompile(`{{[^}\.]*(\.[A-Za-z_\.]+)[^{]*}}`)
)

func validTemplate(template string, inputs, resolverInputs []string, steps map[string]*step.Step) error {
	// Ranging over tmplRegex.FindAllStringSubmatch does not match all "should-match" values, so
	// we split the indented json line by line, and match each lines.
	matches := make([][]string, 0)
	for _, s := range strings.Split(template, "\n") {
		matches = append(matches, tmplRegex.FindAllStringSubmatch(s, -1)...)
	}

	stepNames := stepNames(steps)
	taskInfoKeys := []string{"resolver_username", "created", "requester_username", "task_id", "region"}
	for _, m := range matches {
		parts := strings.Split(m[1], ".")
		if len(parts) >= 3 {
			valueType := parts[1]
			key := parts[2]
			switch valueType {
			case values.StepKey:
				if !utils.ListContainsString(stepNames, key) {
					return fmt.Errorf("step: Wrong step key: %s", key)
				}
				if len(parts) > 3 {
					stepdata := parts[3]
					switch stepdata {
					case "output":
						// "this" doesnt give enough context to fetch jsonschema, no linting -> only check conditions impacted
						// TODO make sure that "this" is only used in check condition
						if key != utask.This && len(parts) >= 5 {
							if err := lintTemplate(steps, parts[2:]); err != nil {
								return errors.Annotatef(err, "Linting error: step %s", key)
							}
						}
					case "metadata":
						if key != utask.This && len(parts) >= 5 {
							if err := lintStepDetails(steps, parts[2:]); err != nil {
								return errors.Annotatef(err, "Linting error: step details for step %q", key)
							}
						}
					}
				}
			case values.InputKey:
				if !utils.ListContainsString(inputs, key) {
					return fmt.Errorf("Wrong input key: %s", key)
				}
			case values.ResolverInputKey:
				if !utils.ListContainsString(resolverInputs, key) {
					return fmt.Errorf("Wrong input key: %s", key)
				}
			case values.ConfigKey:
				// TODO... not sure how to check this... against global secret store?
			case values.TaskKey:
				if !utils.ListContainsString(taskInfoKeys, key) {
					return fmt.Errorf("Wrong task key: %s", key)
				}
			default:
				// other templating handles might fall within a template "range"
				// or such other constructs, not our common use case
				// this won't leak anything -> template author proceeds at his own risk
			}
		}
	}
	return nil
}

// expect parts to contain:
// - {stepname}
// - "output"
// - ... list of keys within object
func lintTemplate(steps map[string]*step.Step, parts []string) error {
	stepName := parts[0]

	v, ok := steps[stepName]
	if !ok {
		return errors.Errorf("Unknown step %s", stepName)
	}

	return lint(stepName, v.Schema, parts[2:]...)
}

// expect parts to contain:
// - {stepname}
// - "metadata"
// - ... list of keys within object
func lintStepDetails(steps map[string]*step.Step, parts []string) error {
	stepName := parts[0]

	v, ok := steps[stepName]
	if !ok {
		return errors.Errorf("Unknown step %s", stepName)
	}

	schema := v.ExecutorMetadata()

	return lint(stepName, schema, parts[2:]...)
}

func lint(url string, schema json.RawMessage, parts ...string) error {
	if len(schema) == 0 {
		return nil
	}
	properties, err := jsonschema.ExtractProperty(url, schema)
	if err != nil {
		return err
	}
	return tryVariablePath(properties, parts)
}

func stepNames(stepMap map[string]*step.Step) []string {
	stepNames := []string{utask.This} // accept "this" as a valid step reference in values map
	for name := range stepMap {
		stepNames = append(stepNames, name)
	}
	return stepNames
}

// tryVariablePath tries to match a chain of variables with the given properties.
// For example:
//      given properties = map[foo:[bar] bar:[bar] qux:[foo] utaskRootKey:[qux]]
//      and parts = [qux foo bar bar bar]
//      "qux.foo.bar.bar.bar" is valid since "qux" makes "foo", "foo" makes "bar"
//      and "bar" makes "bar".
//      "qux.foo.bar.foo" is not valid since we cannot make "foo" from "bar".
//      "foo.bar.bar" is not valid either since we start looping on parts using
//      utaskRootKey map, which contains root properties of the json schema.
func tryVariablePath(properties map[string][]string, parts []string) error {
	// start with root properties
	lastKey := jsonschema.RootKey

	for _, p := range parts {
		v, ok := properties[lastKey]
		if !ok {
			return errors.Errorf(variableDoesNotExist, strings.Join(parts, "."))
		}
		if !utils.ListContainsString(v, p) {
			return errors.Errorf(variableDoesNotExist, strings.Join(parts, "."))
		}
		lastKey = p
	}

	return nil
}

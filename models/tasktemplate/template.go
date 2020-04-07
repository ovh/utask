package tasktemplate

import (
	"encoding/json"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/juju/errors"
	"github.com/loopfz/gadgeto/zesty"
	"github.com/ovh/utask/db/pgjuju"
	"github.com/ovh/utask/db/sqlgenerator"
	"github.com/ovh/utask/engine/input"
	"github.com/ovh/utask/engine/step"
	"github.com/ovh/utask/engine/values"
	"github.com/ovh/utask/pkg/utils"
)

// TaskTemplate holds the formal description for a process
// that can be executed by µTask
// It describes:
// - needed inputs and validation rules on them
// - a collection of named steps, full with their configurations and interdependencies
// - rules for execution rights (allowed resolvers, auto run, blocked), API exposition (hidden)
// - a format for result consolidation in tasks derived from the template
type TaskTemplate struct {
	ID              int64                  `json:"-" db:"id"`
	Name            string                 `json:"name" db:"name"`
	Description     string                 `json:"description" db:"description"`
	LongDescription *string                `json:"long_description,omitempty" db:"long_description"`
	DocLink         *string                `json:"doc_link,omitempty" db:"doc_link"`
	TitleFormat     string                 `json:"title_format,omitempty" db:"title_format"`
	ResultFormat    map[string]interface{} `json:"result_format,omitempty" db:"result_format"`

	AllowedResolverUsernames  []string `json:"allowed_resolver_usernames" db:"allowed_resolver_usernames"`
	AllowAllResolverUsernames bool     `json:"allow_all_resolver_usernames" db:"allow_all_resolver_usernames"`
	AutoRunnable              bool     `json:"auto_runnable" db:"auto_runnable"`
	Blocked                   bool     `json:"blocked" db:"blocked"`
	Hidden                    bool     `json:"hidden" db:"hidden"`
	RetryMax                  *int     `json:"retry_max,omitempty" db:"retry_max"`

	Inputs             []input.Input              `json:"inputs,omitempty" db:"inputs"`
	ResolverInputs     []input.Input              `json:"resolver_inputs,omitempty" db:"resolver_inputs"`
	Variables          []values.Variable          `json:"variables,omitempty" db:"variables"`
	Tags               map[string]string          `json:"tags,omitempty" db:"tags"`
	Steps              map[string]*step.Step      `json:"steps,omitempty" db:"steps"`
	BaseConfigurations map[string]json.RawMessage `json:"base_configurations" db:"base_configurations"`
}

// Create inserts a new task template in DB
func Create(dbp zesty.DBProvider,
	name, description string,
	longDescription,
	docLink *string,
	inputs, resolverInputs []input.Input,
	allowedResolverUsernames []string,
	allowAllResolverUsernames, autoRunnable bool,
	steps map[string]*step.Step,
	variables []values.Variable,
	tags map[string]string,
	resultFormat map[string]interface{},
	titleFormat string,
	retryMax *int,
	baseConfig map[string]json.RawMessage) (tt *TaskTemplate, err error) {

	defer errors.DeferredAnnotatef(&err, "Failed to insert task template")

	tt = &TaskTemplate{
		Name:                      name,
		Description:               description,
		LongDescription:           longDescription,
		DocLink:                   docLink,
		Steps:                     steps,
		Inputs:                    inputs,
		ResolverInputs:            resolverInputs,
		Variables:                 variables,
		Tags:                      tags,
		AllowedResolverUsernames:  allowedResolverUsernames,
		AllowAllResolverUsernames: allowAllResolverUsernames,
		AutoRunnable:              autoRunnable,
		Blocked:                   false,
		Hidden:                    false,
		ResultFormat:              resultFormat,
		TitleFormat:               titleFormat,
		RetryMax:                  retryMax,
		BaseConfigurations:        baseConfig,
	}

	tt, err = create(dbp, tt)
	return tt, err
}

func create(dbp zesty.DBProvider, tt *TaskTemplate) (*TaskTemplate, error) {
	tt.Normalize()

	if err := tt.Valid(); err != nil {
		return nil, err
	}

	if err := dbp.DB().Insert(tt); err != nil {
		return nil, pgjuju.Interpret(err)
	}

	return tt, nil
}

// LoadFromName returns a task template, given its unique human-readable identifier
func LoadFromName(dbp zesty.DBProvider, name string) (tt *TaskTemplate, err error) {
	defer errors.DeferredAnnotatef(&err, "Failed to load template %q from name", name)

	query, params, err := ttSelector.Where(
		squirrel.Eq{`"task_template".name`: name},
	).ToSql()
	if err != nil {
		return nil, err
	}

	err = dbp.DB().SelectOne(&tt, query, params...)
	if err != nil {
		return nil, pgjuju.Interpret(err)
	}

	return tt, nil
}

// LoadFromID returns a task template, given its "private" identifier
// A shortcut only used internally, not exposed through API
func LoadFromID(dbp zesty.DBProvider, id int64) (tt *TaskTemplate, err error) {
	defer errors.DeferredAnnotatef(&err, "Failed to load template from ID %d", id)

	query, params, err := ttSelector.Where(
		squirrel.Eq{`"task_template".id`: id},
	).ToSql()
	if err != nil {
		return nil, err
	}

	err = dbp.DB().SelectOne(&tt, query, params...)
	if err != nil {
		return nil, pgjuju.Interpret(err)
	}

	return tt, nil
}

// ListTemplates returns a list of task templates, in a simplified form (steps not included)
func ListTemplates(dbp zesty.DBProvider, includeHidden bool, pageSize uint64, last *string) (tt []*TaskTemplate, err error) {
	defer errors.DeferredAnnotatef(&err, "Failed to list templates")

	sel := ttBasicSelector.OrderBy(
		`"task_template".id`,
	).Limit(
		pageSize,
	)

	if !includeHidden {
		sel = sel.Where(squirrel.Eq{`"task_template".hidden`: false})
	}

	if last != nil {
		lastTT, err := LoadFromName(dbp, *last)
		if err != nil {
			return nil, err
		}
		sel = sel.Where(`"task_template".id > ?`, lastTT.ID)
	}

	query, params, err := sel.ToSql()
	if err != nil {
		return nil, err
	}

	_, err = dbp.DB().Select(&tt, query, params...)
	if err != nil {
		return nil, pgjuju.Interpret(err)
	}

	return tt, nil
}

// Update introduces changes to a template in DB
func (tt *TaskTemplate) Update(dbp zesty.DBProvider,
	description, longDescription, docLink *string,
	inputs, resolverInputs []input.Input,
	allowedResolverUsernames []string,
	allowAllResolverUsernames, autoRunnable, blocked, hidden *bool,
	steps map[string]*step.Step,
	variables []values.Variable,
	tags map[string]string,
	resultFormat map[string]interface{},
	titleFormat *string,
	retryMax *int,
	baseConfig map[string]json.RawMessage) (err error) {

	defer errors.DeferredAnnotatef(&err, "Failed to update template")

	tt.LongDescription = longDescription
	tt.DocLink = docLink

	if description != nil {
		tt.Description = *description
	}
	if inputs != nil {
		tt.Inputs = inputs
	}
	if resolverInputs != nil {
		tt.ResolverInputs = resolverInputs
	}
	if allowedResolverUsernames != nil {
		tt.AllowedResolverUsernames = allowedResolverUsernames
	}
	if allowAllResolverUsernames != nil {
		tt.AllowAllResolverUsernames = *allowAllResolverUsernames
	}
	if autoRunnable != nil {
		tt.AutoRunnable = *autoRunnable
	}
	if blocked != nil {
		tt.Blocked = *blocked
	}
	if hidden != nil {
		tt.Hidden = *hidden
	}
	if steps != nil {
		tt.Steps = steps
	}
	if variables != nil {
		tt.Variables = variables
	}
	if tags != nil {
		tt.Tags = tags
	}
	if resultFormat != nil {
		tt.ResultFormat = resultFormat
	}
	if titleFormat != nil {
		tt.TitleFormat = *titleFormat
	}
	tt.RetryMax = retryMax
	if baseConfig != nil {
		tt.BaseConfigurations = baseConfig
	}

	tt.Normalize()

	err = update(dbp, tt)

	return
}

func update(dbp zesty.DBProvider, tt *TaskTemplate) error {
	tt.Normalize()

	if err := tt.Valid(); err != nil {
		return err
	}

	rows, err := dbp.DB().Update(tt)
	if err != nil {
		return pgjuju.Interpret(err)
	} else if rows == 0 {
		return errors.NotFoundf("No such template to update: %s", tt.Name)
	}

	return nil
}

// Delete removes a template from DB
func (tt *TaskTemplate) Delete(dbp zesty.DBProvider) (err error) {
	defer errors.DeferredAnnotatef(&err, "Failed to delete template")

	rows, err := dbp.DB().Delete(tt)
	if err != nil {
		return pgjuju.Interpret(err)
	} else if rows == 0 {
		return errors.NotFoundf("No such template to delete: %s", tt.Name)
	}

	return nil
}

// Normalize transforms a template's name into a standard format
func (tt *TaskTemplate) Normalize() {
	tt.Name = utils.NormalizeName(tt.Name)
}

// Valid asserts that the content of a task template is correct:
// - metadata (name, description, etc...) is valid
// - inputs are correctly expressed
// - steps are coherent (dependency graph, templating handles)
func (tt *TaskTemplate) Valid() (err error) {
	defer errors.DeferredAnnotatef(&err, "Invalid task template")

	if err := utils.ValidString("template name", tt.Name); err != nil {
		return err
	}

	if err := utils.ValidString("template description", tt.Description); err != nil {
		return err
	}

	if err := utils.ValidString("template title format", tt.TitleFormat); err != nil {
		return err
	}

	if tt.DocLink != nil {
		if err := utils.ValidString("template doc link", *tt.DocLink); err != nil {
			return err
		}
	}

	if tt.LongDescription != nil {
		if err := utils.ValidText("template long description", *tt.LongDescription); err != nil {
			return err
		}
	}

	if tt.AutoRunnable && len(tt.ResolverInputs) > 0 {
		return errors.NotValidf("A template can't be auto runnable if it has resolver inputs")
	}

	if tt.AllowAllResolverUsernames && !tt.AutoRunnable {
		return errors.NotValidf("A template that can be resolved by everybody have to be auto-runnable")
	}

	inputNames, err := validateInputs(tt.Inputs)
	if err != nil {
		return err
	}

	resolverInputNames, err := validateInputs(tt.ResolverInputs)
	if err != nil {
		return err
	}

	if err := validateVariables(tt.Variables); err != nil {
		return err
	}

	// valid and normalize steps:
	for name, st := range tt.Steps {
		if err := st.ValidAndNormalize(name, tt.BaseConfigurations, tt.Steps); err != nil {
			return errors.NewNotValid(err, fmt.Sprintf("Invalid step %s", name))
		}
	}

	// MarshalIndent as it's easier to read line by line
	tmplJSON, err := utils.JSONMarshalIndent(tt, "", " ")
	if err != nil {
		return err
	}

	if err := validTemplate(string(tmplJSON), inputNames, resolverInputNames, tt.Steps); err != nil {
		return errors.NewNotValid(err, "Invalid text-template handles within task template")
	}

	return nil
}

// IsAutoRunnable asserts that a task issued from this template can
// be executed directly, ie. a resolution can be created and launched automatically
func (tt *TaskTemplate) IsAutoRunnable() bool {
	return tt.AutoRunnable && len(tt.ResolverInputs) == 0
}

// ValidateResolverInputs asserts that input values provided by a task's resolver
// conform to the template's spec for resolver inputs
func (tt *TaskTemplate) ValidateResolverInputs(inputValues map[string]interface{}) error {
	return validateInputsValues(tt.ResolverInputs, inputValues)
}

// ValidateInputs asserts that input values provided by a task's requester
// conform to the template's spec for requester inputs
func (tt *TaskTemplate) ValidateInputs(inputValues map[string]interface{}) error {
	return validateInputsValues(tt.Inputs, inputValues)
}

func validateInputsValues(inputs []input.Input, inputValues map[string]interface{}) error {
	for _, i := range inputs {
		val, ok := inputValues[i.Name]
		if !ok || val == nil || val == "" {
			if i.Default != nil {
				inputValues[i.Name] = i.Default
				continue
			}
			if !i.Optional {
				return errors.NotValidf("Missing input '%s'", i.Name)
			}
		} else {
			if err := i.CheckValue(val); err != nil {
				return err
			}
		}
	}
	return nil
}

// FilterInputs drops received inputs that are not declared by a template
func (tt *TaskTemplate) FilterInputs(inputValues map[string]interface{}) map[string]interface{} {
	filtered := make(map[string]interface{})
	for _, i := range tt.Inputs {
		if val, ok := inputValues[i.Name]; ok {
			filtered[i.Name] = val
		}
	}
	return filtered
}

// validateInputs performs sanity checks on a input set before template creation
func validateInputs(inputs []input.Input) ([]string, error) {
	inputNames := make([]string, 0)
	for _, i := range inputs {
		if err := i.Valid(); err != nil {
			return nil, err
		}
		inputNames = append(inputNames, i.Name)
	}
	return inputNames, nil
}

func validateVariables(variables []values.Variable) error {
	for _, variable := range variables {
		if variable.Name == "" {
			return errors.BadRequestf("variable name can't be empty")
		}
		if variable.Value != nil && variable.Expression != "" {
			return errors.BadRequestf("variable %q can't have both value and expression defined", variable.Name)
		}
		if variable.Value == nil && variable.Expression == "" {
			return errors.BadRequestf("variable %q expression and value can't be empty at the same time", variable.Name)
		}
	}

	return nil
}

var (
	ttBasicSelector = sqlgenerator.PGsql.Select(
		`"task_template".id, "task_template".name, "task_template".description, "task_template".long_description, "task_template".doc_link, "task_template".allowed_resolver_usernames, "task_template".allow_all_resolver_usernames, "task_template".auto_runnable, "task_template".blocked, "task_template".hidden, "task_template".retry_max, "task_template".inputs, "task_template".resolver_inputs, "task_template".base_configurations, "task_template".tags`,
	).From(
		`"task_template"`,
	).OrderBy(
		`"task_template".id`,
	)

	ttSelector = ttBasicSelector.Columns(
		`"task_template".steps, "task_template".variables, "task_template".result_format, "task_template".title_format`,
	)
)

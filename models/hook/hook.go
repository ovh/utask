package hook

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/juju/errors"
	"github.com/loopfz/gadgeto/zesty"
	"github.com/ovh/utask/db/pgjuju"
	"github.com/ovh/utask/db/sqlgenerator"
	"github.com/ovh/utask/pkg/utils"
)

type Hook struct {
	ID              int64       `json:"-" db:"id"`
	Name            string      `json:"name" db:"name"`
	Description     string      `json:"description" db:"description"`
	LongDescription *string     `json:"long_description,omitempty" db:"long_description"`
	DocLink         *string     `json:"doc_link,omitempty" db:"doc_link"`
	Actions         HookActions `json:"actions" db:"actions"`
}

type HookActions []json.RawMessage

// Value returns driver.Value from HookActions.
func (a HookActions) Value() (driver.Value, error) {
	j, err := json.Marshal(a)
	return j, err
}

// Scan HookActions.
func (a *HookActions) Scan(src interface{}) error {
	if src == nil {
		return nil
	}
	source, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("type assertion .([]byte) failed (%T)", src)
	}
	return json.Unmarshal(source, a)
}

// Normalize transforms a hooks's name into a standard format
func (h *Hook) Normalize() {
	h.Name = utils.NormalizeName(h.Name)
}

// Valid asserts that the content of a task template is correct
func (h *Hook) Valid() (err error) {
	defer errors.DeferredAnnotatef(&err, "Invalid hook")

	if err := utils.ValidString("hook name", h.Name); err != nil {
		return err
	}

	if err := utils.ValidString("hook description", h.Description); err != nil {
		return err
	}

	return nil
}

func create(dbp zesty.DBProvider, h *Hook) error {
	h.Normalize()

	if err := h.Valid(); err != nil {
		return err
	}

	if err := dbp.DB().Insert(h); err != nil {
		return pgjuju.Interpret(err)
	}

	return nil
}

func update(dbp zesty.DBProvider, h *Hook) error {
	h.Normalize()

	if err := h.Valid(); err != nil {
		return err
	}

	rows, err := dbp.DB().Update(h)
	if err != nil {
		return pgjuju.Interpret(err)
	} else if rows == 0 {
		return errors.NotFoundf("No such template to update: %s", h.Name)
	}

	return nil
}

// LoadFromName returns a hook, given its unique human-readable identifier
func LoadFromName(dbp zesty.DBProvider, name string) (h *Hook, err error) {
	defer errors.DeferredAnnotatef(&err, "Failed to load template from name")

	query, params, err := hSelector.Where(
		squirrel.Eq{`"hook".name`: name},
	).ToSql()
	if err != nil {
		return nil, err
	}

	err = dbp.DB().SelectOne(&h, query, params...)
	if err != nil {
		return nil, pgjuju.Interpret(err)
	}

	return h, nil
}

var (
	hBasicSelector = sqlgenerator.PGsql.Select(
		`"hook".id, "hook".name, "hook".description, "hook".long_description, "hook".doc_link`,
	).From(
		`"hook"`,
	).OrderBy(
		`"hook".id`,
	)

	hSelector = hBasicSelector.Columns(
		`"hook".actions`,
	)
	mapAllHooks map[string]Hook = make(map[string]Hook)
)

func GetHook(name string) (*Hook, error) {
	h, ok := mapAllHooks[name]
	if ok {
		return &h, nil
	}

	return nil, errors.New("hook not found")
}

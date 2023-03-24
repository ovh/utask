package task

import (
	"encoding/json"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/gofrs/uuid"
	"github.com/juju/errors"
	"github.com/loopfz/gadgeto/zesty"
	"github.com/ovh/utask/db/pgjuju"
	"github.com/ovh/utask/db/sqlgenerator"
	"github.com/ovh/utask/pkg/now"
)

type Metadata[T any] struct {
	ID       int64     `json:"-" db:"id"`
	PublicID string    `json:"id" db:"public_id"`
	TaskID   string    `json:"-" db:"id_task"`
	Created  time.Time `json:"created" db:"created"`
	Updated  time.Time `json:"updated" db:"updated"`
	Key      string    `json:"key" db:"key"`
	Value    T         `json:"value" db:"value"`
}

// CreateMetadata inserts a new metadata in DB
func CreateMetadata[T any](dbp zesty.DBProvider, taskID, key string, value T) (m *Metadata[T], err error) {
	defer errors.DeferredAnnotatef(&err, "Failed to create metadata")

	m = &Metadata[T]{
		PublicID: uuid.Must(uuid.NewV4()).String(),
		TaskID:   taskID,
		Created:  now.Get(),
		Updated:  now.Get(),
		Key:      key,
		Value:    value,
	}

	err = m.Valid()
	if err != nil {
		return nil, err
	}

	err = dbp.DB().Insert(m)
	if err != nil {
		return nil, pgjuju.Interpret(err)
	}

	return m, nil
}

// LoadMetadataFromTaskIDAndKey returns a single metadata value related to a task and a key
func LoadMetadataFromTaskIDAndKey[T any](dbp zesty.DBProvider, taskID, key string) (m *Metadata[T], err error) {
	defer errors.DeferredAnnotatef(&err, "Failed to load metadata from public id")

	query, params, err := mSelector.Where(
		squirrel.Eq{`"task_metadata".id_task`: taskID, `"task_metadata".key`: key},
	).ToSql()
	if err != nil {
		return nil, err
	}

	err = dbp.DB().SelectOne(&m, query, params...)
	if err != nil {
		return m, pgjuju.Interpret(err)
	}

	return m, nil
}

// LoadMetadatasFromTaskID returns the list of metadatas related to a task
func LoadMetadatasFromTaskID[T any](dbp zesty.DBProvider, taskID string) (m []*Metadata[T], err error) {
	defer errors.DeferredAnnotatef(&err, "Failed to load metadata from task id")

	query, params, err := mSelector.Where(
		squirrel.Eq{`"task_metadata".id_task`: taskID},
	).ToSql()
	if err != nil {
		return nil, err
	}

	_, err = dbp.DB().Select(&m, query, params...)
	if err != nil {
		return nil, pgjuju.Interpret(err)
	}

	return m, nil
}

func UpdateMetadata[T any](dbp zesty.DBProvider, m *Metadata[T]) (err error) {
	defer errors.DeferredAnnotatef(&err, "Failed to update metadata")

	m.Updated = now.Get()

	err = m.Valid()
	if err != nil {
		return err
	}

	rows, err := dbp.DB().Update(m)
	if err != nil {
		return pgjuju.Interpret(err)
	} else if rows == 0 {
		return errors.NotFoundf("No such metadata to update: %s", m.PublicID)
	}

	return nil
}

// Valid asserts that the key and value are registered metadata content
func (m *Metadata[T]) Valid() error {
	_, err := json.Marshal(m.Value)
	if err != nil {
		return errors.Errorf("The metadata value is not JSON serializable: %s", err)
	}

	return nil
}

var (
	mSelector = sqlgenerator.PGsql.Select(
		`"task_metadata".id`, `"task_metadata".public_id`, `"task_metadata".id_task`,
		`"task_metadata".created`, `"task_metadata".updated`,
		`"task_metadata".key`, `"task_metadata".value`,
	).From(
		`"task_metadata"`,
	)
)

package task

import (
	"github.com/Masterminds/squirrel"
	"github.com/gofrs/uuid"
	"github.com/juju/errors"
	"github.com/loopfz/gadgeto/zesty"
	"github.com/ovh/utask/db/pgjuju"
	"github.com/ovh/utask/db/sqlgenerator"
)

// Batch represents a group of tasks, created under a common identifier
type Batch struct {
	BatchDBModel
	//	Tasks []*Task `json:"tasks,omitempty" db:"-"`
}

// BatchDBModel is a Batch's representation in DB
type BatchDBModel struct {
	ID       int64  `json:"-" db:"id"`
	PublicID string `json:"id" db:"public_id"`
}

// CreateBatch inserts a new batch in DB
func CreateBatch(dbp zesty.DBProvider) (b *Batch, err error) {
	defer errors.DeferredAnnotatef(&err, "Failed to create batch")

	b = &Batch{
		BatchDBModel: BatchDBModel{
			PublicID: uuid.Must(uuid.NewV4()).String(),
		},
	}

	err = dbp.DB().Insert(&b.BatchDBModel)
	if err != nil {
		return nil, pgjuju.Interpret(err)
	}

	return b, nil
}

// LoadBatchFromPublicID returns a task batch, loaded from DB given its ID
func LoadBatchFromPublicID(dbp zesty.DBProvider, publicID string) (b *Batch, err error) {
	defer errors.DeferredAnnotatef(&err, "Failed to load batch from public id")

	query, params, err := sqlgenerator.PGsql.Select(
		`"batch".id, "batch".public_id`,
	).From(
		`"batch"`,
	).Where(
		squirrel.Eq{`"batch".public_id`: publicID},
	).ToSql()
	if err != nil {
		return nil, err
	}

	err = dbp.DB().SelectOne(&b, query, params...)
	if err != nil {
		return nil, pgjuju.Interpret(err)
	}

	return b, nil
}

// Delete removes a task batch from DB
func (b *Batch) Delete(dbp zesty.DBProvider) (err error) {
	defer errors.DeferredAnnotatef(&err, "Failed to delete batch")

	rows, err := dbp.DB().Delete(b)
	if err != nil {
		return pgjuju.Interpret(err)
	} else if rows == 0 {
		return errors.NotFoundf("No such batch to delete: %s", b.PublicID)
	}

	return nil
}

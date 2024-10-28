package batchutils

import (
	"github.com/Masterminds/squirrel"
	"github.com/loopfz/gadgeto/zesty"

	"github.com/ovh/utask/db/sqlgenerator"
	"github.com/ovh/utask/models/task"
)

// FinalStates hold the states in which a task won't ever be run again
var FinalStates = []string{task.StateDone, task.StateCancelled, task.StateWontfix}

// RunningTasks returns the amount of running tasks sharing the same given batchId.
func RunningTasks(dbp zesty.DBProvider, batchID int64) (int64, error) {
	query, params, err := sqlgenerator.PGsql.
		Select("count (*)").
		From("task t").
		Join("batch b on b.id = t.id_batch").
		Where(squirrel.Eq{"b.id": batchID}).
		Where(squirrel.NotEq{"t.state": FinalStates}).
		ToSql()
	if err != nil {
		return -1, err
	}

	return dbp.DB().SelectInt(query, params...)
}

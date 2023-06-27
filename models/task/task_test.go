package task

import (
	"github.com/loopfz/gadgeto/zesty"
	"github.com/ovh/utask/db/pgjuju"
)

func DeleteAllTasks(dbp zesty.DBProvider) error {
	var tasks []*Task

	query, params, err := tSelector.ToSql()
	if err != nil {
		return err
	}

	_, err = dbp.DB().Select(&tasks, query, params...)
	if err != nil {
		return pgjuju.Interpret(err)
	}

	for _, tsk := range tasks {
		if err := tsk.Delete(dbp); err != nil {
			return err
		}
	}

	return nil
}

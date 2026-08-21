package task

import (
	"testing"

	"github.com/loopfz/gadgeto/zesty"
	"github.com/maxatome/go-testdeep/td"
	"github.com/ovh/utask/db/pgjuju"
	"github.com/ovh/utask/engine/values"
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

func TestExportTaskInfos(t *testing.T) {
	resolverUsername := "foo-resolver"
	resolutionPublicID := "9b1f4f2e-3d1a-4f2b-8a1e-0c7d5f6a2b3c"
	tsk := &Task{
		DBModel: DBModel{
			PublicID:          "3f2d0b1c-5a6e-4c7d-9e8f-1a2b3c4d5e6f",
			RequesterUsername: "bar-requester",
		},
		Resolution:       &resolutionPublicID,
		ResolverUsername: &resolverUsername,
	}

	v := values.NewValues()
	tsk.ExportTaskInfos(v)

	output, err := v.Apply("{{ .task.resolver_username | toString }}", nil, "")
	td.CmpNil(t, err)
	td.Cmp(t, string(output), resolverUsername)

	output, err = v.Apply("{{ .task.resolution_id | toString }}", nil, "")
	td.CmpNil(t, err)
	td.Cmp(t, string(output), resolutionPublicID)
}

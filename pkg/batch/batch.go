package batch

import (
	"context"

	"github.com/juju/errors"
	"github.com/loopfz/gadgeto/zesty"

	"github.com/ovh/utask/models/task"
	"github.com/ovh/utask/models/tasktemplate"
	"github.com/ovh/utask/pkg/taskutils"
)

// TaskArgs holds arguments needed to create tasks in a batch
type TaskArgs struct {
	TemplateName     string                   // Mandatory
	Inputs           []map[string]interface{} // Mandatory
	CommonInput      map[string]interface{}   // Optional
	Comment          string                   // Optional
	WatcherUsernames []string                 // Optional
	WatcherGroups    []string                 // Optional
	Tags             map[string]string        // Optional
}

// Populate creates and adds new tasks to a given batch.
// All tasks share a common batchID which can be used as a listing filter.
// The [constants.SubtaskTagParentTaskID] tag can be set in the Tags to link the newly created tasks to another
// existing task, making it the parent of the batch. A parent task is resumed everytime a child task finishes.
func Populate(ctx context.Context, batch *task.Batch, dbp zesty.DBProvider, args TaskArgs) ([]string, error) {
	tt, err := tasktemplate.LoadFromName(dbp, args.TemplateName)
	if err != nil {
		return nil, err
	}

	taskIDs := make([]string, 0, len(args.Inputs))
	for _, inp := range args.Inputs {
		input, err := mergeMaps(args.CommonInput, inp)
		if err != nil {
			return nil, err
		}

		t, err := taskutils.CreateTask(
			ctx,
			dbp,
			tt,
			args.WatcherUsernames,
			args.WatcherGroups,
			[]string{},
			[]string{},
			input,
			batch,
			args.Comment,
			nil,
			args.Tags,
		)
		if err != nil {
			return nil, err
		}
		taskIDs = append(taskIDs, t.PublicID)
	}
	return taskIDs, nil
}

func mergeMaps(common, particular map[string]interface{}) (map[string]interface{}, error) {
	merged := make(map[string]interface{}, len(common)+len(particular))
	for key, value := range particular {
		merged[key] = value
	}

	for key, value := range common {
		if _, ok := merged[key]; ok {
			return nil, errors.NewBadRequest(nil, "Conflicting keys in input maps")
		}
		merged[key] = value
	}
	return merged, nil
}

package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/juju/errors"
	"github.com/loopfz/gadgeto/zesty"

	"github.com/ovh/utask"
	"github.com/ovh/utask/models/task"
	"github.com/ovh/utask/models/tasktemplate"
	"github.com/ovh/utask/pkg/taskutils"
)

type createBatchIn struct {
	TemplateName     string                   `json:"template_name" binding:"required"`
	CommonInput      map[string]interface{}   `json:"common_input"`
	Inputs           []map[string]interface{} `json:"inputs" binding:"required"`
	Comment          string                   `json:"comment"`
	WatcherUsernames []string                 `json:"watcher_usernames"`
	Tags             map[string]string        `json:"tags"`
}

// CreateBatch handles the creation of a collection of tasks based on the same template
// one task is created for each element in the "inputs" slice
// all tasks share a common "batchID" which can be used as a listing filter on /task
func CreateBatch(c *gin.Context, in *createBatchIn) (*task.Batch, error) {
	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return nil, err
	}

	tt, err := tasktemplate.LoadFromName(dbp, in.TemplateName)
	if err != nil {
		return nil, err
	}

	if err := dbp.Tx(); err != nil {
		return nil, err
	}

	b, err := task.CreateBatch(dbp)
	if err != nil {
		dbp.Rollback()
		return nil, err
	}

	for _, inp := range in.Inputs {
		input, err := conjMap(in.CommonInput, inp)
		if err != nil {
			dbp.Rollback()
			return nil, err
		}

		_, err = taskutils.CreateTask(c, dbp, tt, in.WatcherUsernames, []string{}, input, b, in.Comment, nil, in.Tags)
		if err != nil {
			dbp.Rollback()
			return nil, err
		}
	}

	if err := dbp.Commit(); err != nil {
		dbp.Rollback()
		return nil, err
	}

	return b, nil
}

func conjMap(common, particular map[string]interface{}) (map[string]interface{}, error) {
	conj := make(map[string]interface{})
	if particular != nil {
		for key, value := range particular {
			conj[key] = value
		}
	}
	if common != nil {
		for key, value := range common {
			if _, ok := conj[key]; ok {
				return nil, errors.NewBadRequest(nil, "Conflicting keys in input maps")
			}
			conj[key] = value
		}
	}
	return conj, nil
}

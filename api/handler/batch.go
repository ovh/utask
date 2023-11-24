package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/loopfz/gadgeto/zesty"

	"github.com/ovh/utask"
	"github.com/ovh/utask/models/task"
	"github.com/ovh/utask/pkg/batch"
	"github.com/ovh/utask/pkg/metadata"
	"github.com/ovh/utask/pkg/utils"
)

type createBatchIn struct {
	TemplateName     string                   `json:"template_name" binding:"required"`
	CommonInput      map[string]interface{}   `json:"common_input"`
	Inputs           []map[string]interface{} `json:"inputs" binding:"required"`
	Comment          string                   `json:"comment"`
	WatcherUsernames []string                 `json:"watcher_usernames"`
	WatcherGroups    []string                 `json:"watcher_groups"`
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

	metadata.AddActionMetadata(c, metadata.TemplateName, in.TemplateName)

	if err := utils.ValidateTags(in.Tags); err != nil {
		return nil, err
	}

	if err := dbp.Tx(); err != nil {
		return nil, err
	}

	b, err := task.CreateBatch(dbp)
	if err != nil {
		_ = dbp.Rollback()
		return nil, err
	}

	metadata.AddActionMetadata(c, metadata.BatchID, b.PublicID)

	_, err = batch.Populate(c, b, dbp, batch.TaskArgs{
		TemplateName:     in.TemplateName,
		Inputs:           in.Inputs,
		CommonInput:      in.CommonInput,
		Comment:          in.Comment,
		WatcherUsernames: in.WatcherUsernames,
		WatcherGroups:    in.WatcherGroups,
		Tags:             in.Tags,
	})
	if err != nil {
		_ = dbp.Rollback()
		return nil, err
	}

	if err := dbp.Commit(); err != nil {
		_ = dbp.Rollback()
		return nil, err
	}

	return b, nil
}

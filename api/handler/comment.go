package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/juju/errors"
	"github.com/loopfz/gadgeto/zesty"

	"github.com/ovh/utask"
	"github.com/ovh/utask/models/resolution"
	"github.com/ovh/utask/models/task"
	"github.com/ovh/utask/models/tasktemplate"
	"github.com/ovh/utask/pkg/auth"
)

type createCommentIn struct {
	TaskID  string `path:"id, required"`
	Content string `json:"content"`
}

// CreateComment create a comment related to a task
func CreateComment(c *gin.Context, in *createCommentIn) (*task.Comment, error) {
	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return nil, err
	}

	t, err := task.LoadFromPublicID(dbp, in.TaskID)
	if err != nil {
		return nil, err
	}

	tt, err := tasktemplate.LoadFromID(dbp, t.TemplateID)
	if err != nil {
		return nil, err
	}

	var res *resolution.Resolution
	if t.Resolution != nil {
		res, err = resolution.LoadFromPublicID(dbp, *t.Resolution)
		if err != nil {
			return nil, err
		}
	}

	requester := auth.IsRequester(c, t) == nil
	watcher := auth.IsWatcher(c, t) == nil
	resolutionManager := auth.IsResolutionManager(c, tt, t, res) == nil

	if !requester && !watcher && !resolutionManager {
		return nil, errors.Forbiddenf("Can't create comment")
	}

	reqUsername := auth.GetIdentity(c)

	comment, err := task.CreateComment(dbp, t, reqUsername, in.Content)
	if err != nil {
		return nil, err
	}

	return comment, nil
}

type getCommentIn struct {
	TaskID    string `path:"id, required"`
	CommentID string `path:"commentid, required"`
}

// GetComment return a specific comment related to a task
func GetComment(c *gin.Context, in *getCommentIn) (*task.Comment, error) {
	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return nil, err
	}

	comment, err := task.LoadCommentFromPublicID(dbp, in.CommentID)
	if err != nil {
		return nil, err
	}

	t, err := task.LoadFromPublicID(dbp, in.TaskID)
	if err != nil {
		return nil, err
	}

	tt, err := tasktemplate.LoadFromID(dbp, t.TemplateID)
	if err != nil {
		return nil, err
	}

	var res *resolution.Resolution
	if t.Resolution != nil {
		res, err = resolution.LoadFromPublicID(dbp, *t.Resolution)
		if err != nil {
			return nil, err
		}
	}

	requester := auth.IsRequester(c, t) == nil
	watcher := auth.IsWatcher(c, t) == nil
	resolutionManager := auth.IsResolutionManager(c, tt, t, res) == nil

	if !requester && !watcher && !resolutionManager {
		return nil, errors.Forbiddenf("Can't get comment")
	}

	return comment, nil
}

type listCommentsIn struct {
	TaskID string `path:"id, required"`
}

// ListComments return a list of comments related to a task
func ListComments(c *gin.Context, in *listCommentsIn) ([]*task.Comment, error) {
	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return nil, err
	}

	t, err := task.LoadFromPublicID(dbp, in.TaskID)
	if err != nil {
		return nil, err
	}

	tt, err := tasktemplate.LoadFromID(dbp, t.TemplateID)
	if err != nil {
		return nil, err
	}

	var res *resolution.Resolution
	if t.Resolution != nil {
		res, err = resolution.LoadFromPublicID(dbp, *t.Resolution)
		if err != nil {
			return nil, err
		}
	}

	requester := auth.IsRequester(c, t) == nil
	watcher := auth.IsWatcher(c, t) == nil
	resolutionManager := auth.IsResolutionManager(c, tt, t, res) == nil

	if !requester && !watcher && !resolutionManager {
		return nil, errors.Forbiddenf("Can't list comment")
	}

	comments, err := task.LoadCommentsFromTaskID(dbp, t.ID)
	if err != nil {
		return nil, err
	}

	return comments, nil
}

type updateCommentIn struct {
	TaskID    string `path:"id, required"`
	CommentID string `path:"commentid, required"`
	Content   string `json:"content"`
}

// UpdateComment update a specific comment related to a task
func UpdateComment(c *gin.Context, in *updateCommentIn) (*task.Comment, error) {
	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return nil, err
	}

	t, err := task.LoadFromPublicID(dbp, in.TaskID)
	if err != nil {
		return nil, err
	}

	comment, err := task.LoadCommentFromPublicID(dbp, in.CommentID)
	if err != nil {
		return nil, err
	}

	reqUsername := auth.GetIdentity(c)

	if auth.IsAdmin(c) != nil && reqUsername != comment.Username {
		return nil, errors.Forbiddenf("Can't update comment")
	}

	if comment.TaskID != t.ID {
		return nil, errors.BadRequestf("Comment and task don't match")
	}

	err = comment.Update(dbp, in.Content)
	if err != nil {
		return nil, err
	}

	return comment, nil
}

type deleteCommentIn struct {
	TaskID    string `path:"id, required"`
	CommentID string `path:"commentid, required"`
}

// DeleteComment delete a specific comment related to a task
func DeleteComment(c *gin.Context, in *deleteCommentIn) error {
	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return err
	}

	t, err := task.LoadFromPublicID(dbp, in.TaskID)
	if err != nil {
		return err
	}

	comment, err := task.LoadCommentFromPublicID(dbp, in.CommentID)
	if err != nil {
		return err
	}

	reqUsername := auth.GetIdentity(c)

	if auth.IsAdmin(c) != nil && reqUsername != comment.Username {
		return errors.Forbiddenf("Can't update comment")
	}

	if comment.TaskID != t.ID {
		return errors.BadRequestf("Comment and task don't match")
	}

	return comment.Delete(dbp)
}

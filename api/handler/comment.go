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
	"github.com/ovh/utask/pkg/metadata"
)

type createCommentIn struct {
	TaskID  string `path:"id, required"`
	Content string `json:"content"`
}

// CreateComment create a comment related to a task
func CreateComment(c *gin.Context, in *createCommentIn) (*task.Comment, error) {
	metadata.AddActionMetadata(c, metadata.TaskID, in.TaskID)

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

	metadata.AddActionMetadata(c, metadata.TemplateName, tt.Name)

	var res *resolution.Resolution
	if t.Resolution != nil {
		res, err = resolution.LoadFromPublicID(dbp, *t.Resolution)
		if err != nil {
			return nil, err
		}

		metadata.AddActionMetadata(c, metadata.ResolutionID, res.PublicID)
	}

	admin := auth.IsAdmin(c) == nil
	requester := auth.IsRequester(c, t) == nil
	watcher := auth.IsWatcher(c, t) == nil
	resolutionManager := auth.IsResolutionManager(c, tt, t, res) == nil

	if !requester && !watcher && !resolutionManager && !admin {
		return nil, errors.Forbiddenf("Can't create comment")
	} else if !requester && !watcher && !resolutionManager {
		metadata.SetSUDO(c)
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
	metadata.AddActionMetadata(c, metadata.TaskID, in.TaskID)

	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return nil, err
	}

	comment, err := task.LoadCommentFromPublicID(dbp, in.CommentID)
	if err != nil {
		return nil, err
	}

	metadata.AddActionMetadata(c, metadata.CommentID, comment.PublicID)

	t, err := task.LoadFromPublicID(dbp, in.TaskID)
	if err != nil {
		return nil, err
	}

	tt, err := tasktemplate.LoadFromID(dbp, t.TemplateID)
	if err != nil {
		return nil, err
	}

	metadata.AddActionMetadata(c, metadata.TemplateName, tt.Name)

	var res *resolution.Resolution
	if t.Resolution != nil {
		res, err = resolution.LoadFromPublicID(dbp, *t.Resolution)
		if err != nil {
			return nil, err
		}

		metadata.AddActionMetadata(c, metadata.ResolutionID, res.PublicID)
	}

	admin := auth.IsAdmin(c) == nil
	requester := auth.IsRequester(c, t) == nil
	watcher := auth.IsWatcher(c, t) == nil
	resolutionManager := auth.IsResolutionManager(c, tt, t, res) == nil

	if !requester && !watcher && !resolutionManager && !admin {
		return nil, errors.Forbiddenf("Can't get comment")
	} else if !requester && !watcher && !resolutionManager {
		metadata.SetSUDO(c)
	}

	return comment, nil
}

type listCommentsIn struct {
	TaskID string `path:"id, required"`
}

// ListComments return a list of comments related to a task
func ListComments(c *gin.Context, in *listCommentsIn) ([]*task.Comment, error) {
	metadata.AddActionMetadata(c, metadata.TaskID, in.TaskID)

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

	metadata.AddActionMetadata(c, metadata.TemplateName, tt.Name)

	var res *resolution.Resolution
	if t.Resolution != nil {
		res, err = resolution.LoadFromPublicID(dbp, *t.Resolution)
		if err != nil {
			return nil, err
		}

		metadata.AddActionMetadata(c, metadata.ResolutionID, res.PublicID)
	}

	admin := auth.IsAdmin(c) == nil
	requester := auth.IsRequester(c, t) == nil
	watcher := auth.IsWatcher(c, t) == nil
	resolutionManager := auth.IsResolutionManager(c, tt, t, res) == nil

	if !requester && !watcher && !resolutionManager && !admin {
		return nil, errors.Forbiddenf("Can't list comment")
	} else if !requester && !watcher && !resolutionManager {
		metadata.SetSUDO(c)
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
	metadata.AddActionMetadata(c, metadata.TaskID, in.TaskID)
	metadata.AddActionMetadata(c, metadata.CommentID, in.CommentID)

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

	admin := auth.IsAdmin(c) == nil
	commentAuthor := reqUsername == comment.Username

	if !commentAuthor && !admin {
		return nil, errors.Forbiddenf("Not allowed to update comment")
	} else if !commentAuthor {
		metadata.SetSUDO(c)
	}

	if t.Resolution != nil {
		res, err := resolution.LoadFromPublicID(dbp, *t.Resolution)
		if err != nil {
			return nil, err
		}

		metadata.AddActionMetadata(c, metadata.ResolutionID, res.PublicID)
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
	metadata.AddActionMetadata(c, metadata.TaskID, in.TaskID)
	metadata.AddActionMetadata(c, metadata.CommentID, in.CommentID)

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

	if t.Resolution != nil {
		res, err := resolution.LoadFromPublicID(dbp, *t.Resolution)
		if err != nil {
			return err
		}

		metadata.AddActionMetadata(c, metadata.ResolutionID, res.PublicID)
	}

	reqUsername := auth.GetIdentity(c)

	admin := auth.IsAdmin(c) == nil
	commentAuthor := reqUsername == comment.Username

	if !commentAuthor && !admin {
		return errors.Forbiddenf("Not allowed to delete comment")
	} else if !commentAuthor {
		metadata.SetSUDO(c)
	}

	if comment.TaskID != t.ID {
		return errors.BadRequestf("Comment and task don't match")
	}

	return comment.Delete(dbp)
}

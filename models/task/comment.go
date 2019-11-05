package task

import (
	"time"

	"github.com/ovh/utask/db/pgjuju"
	"github.com/ovh/utask/db/sqlgenerator"
	"github.com/ovh/utask/pkg/now"
	"github.com/ovh/utask/pkg/utils"

	"github.com/Masterminds/squirrel"
	"github.com/gofrs/uuid"
	"github.com/juju/errors"
	"github.com/loopfz/gadgeto/zesty"
)

// Comment is the structure representing a comment made on a task
type Comment struct {
	ID       int64     `json:"-" db:"id"`
	PublicID string    `json:"id" db:"public_id"`
	TaskID   int64     `json:"-" db:"id_task"`
	Username string    `json:"username" db:"username"`
	Created  time.Time `json:"created" db:"created"`
	Updated  time.Time `json:"updated" db:"updated"`
	Content  string    `json:"content" db:"content"`
}

// CreateComment inserts a new comment in DB
func CreateComment(dbp zesty.DBProvider, t *Task, user, content string) (c *Comment, err error) {
	defer errors.DeferredAnnotatef(&err, "Failed to create comment")

	c = &Comment{
		PublicID: uuid.Must(uuid.NewV4()).String(),
		TaskID:   t.ID,
		Username: user,
		Created:  now.Get(),
		Updated:  now.Get(),
		Content:  content,
	}

	err = c.Valid()
	if err != nil {
		return nil, err
	}

	err = dbp.DB().Insert(c)
	if err != nil {
		return nil, pgjuju.Interpret(err)
	}

	return c, nil
}

// LoadCommentFromPublicID returns a single comment, given its ID
func LoadCommentFromPublicID(dbp zesty.DBProvider, publicID string) (c *Comment, err error) {
	defer errors.DeferredAnnotatef(&err, "Failed to load comment from public id")

	query, params, err := cSelector.Where(
		squirrel.Eq{`"task_comment".public_id`: publicID},
	).ToSql()

	err = dbp.DB().SelectOne(&c, query, params...)
	if err != nil {
		return nil, pgjuju.Interpret(err)
	}

	return c, nil
}

// LoadCommentsFromTaskID returns the list of comments related to a task
func LoadCommentsFromTaskID(dbp zesty.DBProvider, taskID int64) (c []*Comment, err error) {
	defer errors.DeferredAnnotatef(&err, "Failed to load comments from task id")

	query, params, err := cSelector.Where(
		squirrel.Eq{`"task_comment".id_task`: taskID},
	).ToSql()

	_, err = dbp.DB().Select(&c, query, params...)
	if err != nil {
		return nil, pgjuju.Interpret(err)
	}

	return c, nil
}

// Update changes the content of a comment in DB
func (c *Comment) Update(dbp zesty.DBProvider, content string) (err error) {
	defer errors.DeferredAnnotatef(&err, "Failed to update comment")

	c.Content = content
	c.Updated = now.Get()

	err = c.Valid()
	if err != nil {
		return err
	}

	rows, err := dbp.DB().Update(c)
	if err != nil {
		return pgjuju.Interpret(err)
	} else if rows == 0 {
		return errors.NotFoundf("No such comment to update: %s", c.PublicID)
	}

	return nil
}

// Delete removes a comment from DB
func (c *Comment) Delete(dbp zesty.DBProvider) (err error) {
	defer errors.DeferredAnnotatef(&err, "Failed to delete comment")

	rows, err := dbp.DB().Delete(c)
	if err != nil {
		return pgjuju.Interpret(err)
	} else if rows == 0 {
		return errors.NotFoundf("No such comment to delete: %s", c.PublicID)
	}

	return nil
}

// Valid asserts that the content of a message is whithin min/max character bounds
func (c *Comment) Valid() error {
	return utils.ValidText("task comment", c.Content)
}

var (
	cSelector = sqlgenerator.PGsql.Select(
		`"task_comment".id, "task_comment".public_id, "task_comment".id_task, "task_comment".username, "task_comment".created, "task_comment".updated, "task_comment".content`,
	).From(
		`"task_comment"`,
	)
)

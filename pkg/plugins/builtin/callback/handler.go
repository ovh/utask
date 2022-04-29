package plugincallback

import (
	"github.com/gin-gonic/gin"
	"github.com/juju/errors"
	"github.com/loopfz/gadgeto/zesty"
	"github.com/ovh/utask"
	"github.com/ovh/utask/engine"
	"github.com/ovh/utask/models/task"
	"github.com/ovh/utask/pkg/jsonschema"
	"github.com/ovh/utask/pkg/metadata"
	"github.com/sirupsen/logrus"
)

const (
	CallbackID     = "callback_id"
	CallbackSecret = "callback_secret"
	CallbackBody   = "callback_body"
)

type handleCallbackIn struct {
	CallbackID     string                 `path:"id, required"`
	CallbackSecret string                 `query:"t, required"`
	Body           map[string]interface{} `body:""`
}

type handleCallbackOut struct {
	Message string `json:"message"`
}

func HandleCallback(c *gin.Context, in *handleCallbackIn) (res *handleCallbackOut, err error) {
	metadata.AddActionMetadata(c, CallbackID, in.CallbackID)
	metadata.AddActionMetadata(c, CallbackSecret, in.CallbackSecret)
	metadata.AddActionMetadata(c, CallbackBody, in.Body)

	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return nil, err
	}

	if err = dbp.Tx(); err != nil {
		return nil, err
	}

	cb, err := loadFromPublicID(dbp, in.CallbackID, true)
	if err != nil {
		dbp.Rollback()
		return nil, err
	}

	if cb.Secret != in.CallbackSecret {
		dbp.Rollback()
		return nil, errors.NotFoundf("failed to load callback from public id: callback")
	}

	if cb.Called != nil {
		dbp.Rollback()
		return nil, errors.BadRequestf("callback has already been resolved")
	}

	t, err := task.LoadFromID(dbp, cb.TaskID)
	if err != nil {
		dbp.Rollback()
		return nil, errors.BadRequestf("unable to fetch related task")
	}

	// Check the state of the related task
	switch t.State {
	case task.StateBlocked, task.StateRunning, task.StateWaiting:
	default:
		dbp.Rollback()
		return nil, errors.BadRequestf("related task is not in a valid state: %s", t.State)
	}

	if cb.Schema != nil {
		s, err := jsonschema.NormalizeAndCompile(in.CallbackID, cb.Schema)
		if err != nil {
			dbp.Rollback()
			return nil, errors.BadRequestf("unable to validate body: %s", err)
		}
		vc := jsonschema.Validator(in.CallbackID, s)
		if err := vc(in.Body); err != nil {
			dbp.Rollback()
			return nil, errors.BadRequestf("unable to validate body: %s", err)
		}
	}

	if err = cb.SetCalled(dbp, in.Body); err != nil {
		dbp.Rollback()
		return nil, err
	}

	if err := dbp.Commit(); err != nil {
		dbp.Rollback()
		return nil, err
	}

	logrus.Debugf("resuming task %q resolution %q", t.PublicID, *t.Resolution)
	logrus.WithFields(logrus.Fields{"task_id": t.PublicID, "resolution_id": *t.Resolution}).Debugf("resuming resolution %q as callback %q has been called", *t.Resolution, cb.PublicID)

	// We ignore the potential error because the caller don't care about it
	_ = engine.GetEngine().Resolve(*t.Resolution, nil)

	res = &handleCallbackOut{
		Message: "The callback has been resolved",
	}

	return res, nil
}

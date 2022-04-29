package plugincallback

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/gofrs/uuid"
	"github.com/juju/errors"
	"github.com/loopfz/gadgeto/zesty"
	"github.com/ovh/utask"
	"github.com/ovh/utask/db/pgjuju"
	"github.com/ovh/utask/db/sqlgenerator"
	"github.com/ovh/utask/models"
	"github.com/ovh/utask/models/resolution"
	"github.com/ovh/utask/models/task"
	"github.com/ovh/utask/pkg/jsonschema"
	"github.com/ovh/utask/pkg/now"
)

type callback struct {
	ID               int64           `json:"-" db:"id"`
	PublicID         string          `json:"id" db:"public_id"`
	TaskID           int64           `json:"-" db:"id_task"`
	ResolutionID     int64           `json:"-" db:"id_resolution"`
	Created          time.Time       `json:"created" db:"created"`
	Updated          time.Time       `json:"updated" db:"updated"`
	Called           *time.Time      `json:"called" db:"called"`
	ResolverUsername string          `json:"resolver" db:"resolver_username"`
	EncryptedSchema  []byte          `json:"-" db:"encrypted_schema"`
	EncryptedSecret  []byte          `json:"-" db:"encrypted_secret"`
	EncryptedBody    []byte          `json:"-" db:"encrypted_body"`
	Secret           string          `json:"-" db:"-"`
	Body             json.RawMessage `json:"body" db:"-"`
	Schema           json.RawMessage `json:"schema" db:"-"`
}

func createCallback(dbp zesty.DBProvider, task *task.Task, ctx *CallbackContext, schemaJSON string) (cb *callback, err error) {
	defer errors.DeferredAnnotatef(&err, "Failed to create callback")

	resolution, err := resolution.LoadFromPublicID(dbp, *task.Resolution)
	if err != nil {
		return nil, err
	}

	cb = &callback{
		PublicID:         uuid.Must(uuid.NewV4()).String(),
		TaskID:           task.ID,
		ResolutionID:     resolution.ID,
		Created:          now.Get(),
		Updated:          now.Get(),
		ResolverUsername: ctx.RequesterUsername,
		Secret:           uuid.Must(uuid.NewV4()).String(),
	}

	if schemaJSON != "" {
		cb.Schema, err = jsonschema.NormalizeAndCompile(ctx.StepName, json.RawMessage(schemaJSON))

		if err != nil {
			return nil, errors.NewBadRequest(fmt.Errorf("unable to parse provided schema: %s", err), "")
		}
	}

	cb.EncryptedSchema, err = models.EncryptionKey.Encrypt(cb.Schema, []byte(cb.PublicID))
	if err != nil {
		return nil, err
	}

	cb.EncryptedSecret, err = models.EncryptionKey.Encrypt([]byte(cb.Secret), []byte(cb.PublicID))
	if err != nil {
		return nil, err
	}

	cb.EncryptedBody, err = models.EncryptionKey.Encrypt([]byte{}, []byte(cb.PublicID))
	if err != nil {
		return nil, err
	}

	err = dbp.DB().Insert(cb)
	if err != nil {
		return nil, pgjuju.Interpret(err)
	}

	return cb, nil
}

func (cb *callback) update(dbp zesty.DBProvider) (err error) {
	defer errors.DeferredAnnotatef(&err, "failed to update callback")

	cb.EncryptedSchema, err = models.EncryptionKey.Encrypt(cb.Schema, []byte(cb.PublicID))
	if err != nil {
		return err
	}

	cb.EncryptedSecret, err = models.EncryptionKey.Encrypt([]byte(cb.Secret), []byte(cb.PublicID))
	if err != nil {
		return err
	}

	cb.EncryptedBody, err = models.EncryptionKey.Encrypt(cb.Body, []byte(cb.PublicID))
	if err != nil {
		return err
	}

	rows, err := dbp.DB().Update(&cb)
	if err != nil {
		return pgjuju.Interpret(err)
	} else if rows == 0 {
		return errors.NotFoundf("no such callback to update: %s", cb.PublicID)
	}

	return nil
}

func (cb *callback) SetCalled(dbp zesty.DBProvider, body interface{}) (err error) {
	defer errors.DeferredAnnotatef(&err, "failed to update callback")

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	cb.EncryptedBody, err = models.EncryptionKey.Encrypt(bodyBytes, []byte(cb.PublicID))
	if err != nil {
		return err
	}

	nowTime := now.Get()
	cb.Called = &nowTime

	rows, err := dbp.DB().Update(cb)
	if err != nil {
		return pgjuju.Interpret(err)
	} else if rows == 0 {
		return errors.NotFoundf("no such task to update: %s", cb.PublicID)
	}

	return nil
}

func loadFromPublicID(dbp zesty.DBProvider, publicID string, forUpdate bool) (*callback, error) {
	cb, err := load(dbp, publicID, forUpdate)
	if err != nil {
		return nil, err
	}

	schemaBytes, err := models.EncryptionKey.Decrypt(cb.EncryptedSchema, []byte(cb.PublicID))
	if err != nil {
		return nil, err
	}
	cb.Schema = schemaBytes

	secretBytes, err := models.EncryptionKey.Decrypt(cb.EncryptedSecret, []byte(cb.PublicID))
	if err != nil {
		return nil, err
	}
	cb.Secret = string(secretBytes)

	bodyBytes, err := models.EncryptionKey.Decrypt(cb.EncryptedBody, []byte(cb.PublicID))
	if err != nil {
		return nil, err
	}
	cb.Body = bodyBytes

	return cb, err
}

func load(dbp zesty.DBProvider, publicID string, locked bool) (cb *callback, err error) {
	defer errors.DeferredAnnotatef(&err, "failed to load callback from public id")

	sel := rSelector

	if locked {
		sel = sel.Suffix(`FOR NO KEY UPDATE OF "callback"`)
	}

	query, params, err := sel.
		Where(squirrel.Eq{`"callback".public_id`: publicID}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var rows []*callback
	_, err = dbp.DB().Select(&rows, query, params...)
	if err != nil {
		return nil, pgjuju.Interpret(err)
	} else if len(rows) != 1 {
		return nil, errors.NotFoundf("callback")
	} else {
		cb = rows[0]
	}

	return cb, nil
}

func listCallbacks(dbp zesty.DBProvider, pageSize uint64, last *string, locked bool) (r []*callback, err error) {
	defer errors.DeferredAnnotatef(&err, "failed to list callbacks")

	sel := rSelector.OrderBy(
		`"resolution".id`,
	).Limit(
		pageSize,
	)

	if locked {
		sel = sel.Suffix(`FOR NO KEY UPDATE OF "callback"`)
	}

	if last != nil {
		lastR, err := loadFromPublicID(dbp, *last, locked)
		if err != nil {
			return nil, err
		}
		sel = sel.Where(`"callback".id > ?`, lastR.ID)
	}

	query, params, err := sel.ToSql()
	if err != nil {
		return nil, err
	}

	_, err = dbp.DB().Select(&r, query, params...)
	if err != nil {
		return nil, pgjuju.Interpret(err)
	}

	return r, nil
}

// RotateEncryptionKeys loads all callbacks stored in DB and makes sure
// that their cyphered content has been handled with the latest
// available storage key
func RotateEncryptionKeys(dbp zesty.DBProvider) (err error) {
	defer errors.DeferredAnnotatef(&err, "Failed to rotate encrypted callbacks to new key")

	var last string
	for {
		var lastID *string
		if last != "" {
			lastID = &last
		}
		// load all callbacks
		callbacks, err := listCallbacks(dbp, utask.MaxPageSize, lastID, false)
		if err != nil {
			return err
		}
		if len(callbacks) == 0 {
			break
		}
		last = callbacks[len(callbacks)-1].PublicID

		for _, c := range callbacks {
			sp, err := dbp.TxSavepoint()
			if err != nil {
				return err
			}
			// load callback locked
			cb, err := loadFromPublicID(dbp, c.PublicID, true)
			if err != nil {
				dbp.RollbackTo(sp)
				return err
			}
			// update callback (encrypt)
			if err := cb.update(dbp); err != nil {
				dbp.RollbackTo(sp)
				return err
			}
			// commit
			if err := dbp.Commit(); err != nil {
				return err
			}
		}
	}

	return nil
}

var rSelector = sqlgenerator.PGsql.Select(
	`"callback".id, "callback".public_id, "callback".id_task, "callback".id_resolution, "callback".resolver_username, "callback".created, "callback".updated, "callback".called, "callback".encrypted_schema, "callback".encrypted_body, "callback".encrypted_secret`,
).From(
	`"callback"`,
).OrderBy(
	`"callback".id`,
)

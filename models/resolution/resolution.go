package resolution

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/gofrs/uuid"
	"github.com/juju/errors"
	"github.com/loopfz/gadgeto/zesty"

	"github.com/ovh/utask"
	"github.com/ovh/utask/db/pgjuju"
	"github.com/ovh/utask/db/sqlgenerator"
	"github.com/ovh/utask/engine/step"
	"github.com/ovh/utask/engine/values"
	"github.com/ovh/utask/models"
	"github.com/ovh/utask/models/task"
	"github.com/ovh/utask/models/tasktemplate"
	"github.com/ovh/utask/pkg/compress"
	"github.com/ovh/utask/pkg/now"
	"github.com/ovh/utask/pkg/utils"
)

// all valid resolution states
const (
	// non runnable

	StateRunning   = "RUNNING"
	StateDone      = "DONE"
	StateCancelled = "CANCELLED"

	// runnable / cancellable

	StateTODO              = "TODO"               // default on creation
	StatePaused            = "PAUSED"             // pause execution in order to make safe updates
	StateBlockedToCheck    = "BLOCKED_TOCHECK"    // blocked by a crash, needs human intervention
	StateBlockedBadRequest = "BLOCKED_BADREQUEST" // blocked by client error, bad input, etc..
	StateBlockedDeadlock   = "BLOCKED_DEADLOCK"   // blocked by unsolvable dependencies
	StateBlockedMaxRetries = "BLOCKED_MAXRETRIES" // has reached max retries, still failing
	StateBlockedFatal      = "BLOCKED_FATAL"      // encountered a fatal non-client error

	// collectable

	StateWaiting          = "WAITING"
	StateCrashed          = "CRASHED"
	StateRetry            = "RETRY"
	StateError            = "ERROR" // a step failed, we'll retry, keep the resolution running
	StateToAutorun        = "TO_AUTORUN"
	StateToAutorunDelayed = "TO_AUTORUN_DELAYED"
	StateAutorunning      = "AUTORUNNING"
)

// Resolution is the full representation of a task's resolution process
// composed from data from "resolution" table and "task" table
// All intermediary state of execution will be held by this structure
type Resolution struct {
	DBModel
	TaskPublicID                     string                 `json:"task_id" db:"task_public_id"`
	TaskTitle                        string                 `json:"task_title" db:"task_title"`
	Values                           *values.Values         `json:"-" db:"-"`                         // never persisted: rebuilt on instantiation
	Steps                            map[string]*step.Step  `json:"steps,omitempty" db:"-"`           // persisted in encrypted blob
	ResolverInput                    map[string]interface{} `json:"resolver_inputs,omitempty" db:"-"` // persisted in encrypted blob
	StepTreeIndex                    map[string][]string    `json:"-" db:"-"`
	StepTreeIndexPrune               map[string][]string    `json:"-" db:"-"`
	StepList                         []string               `json:"-" db:"-"`
	ForeachChildrenAlreadyContracted map[string]bool        `json:"-" db:"-"`
}

// DBModel is a resolution's representation in DB
type DBModel struct {
	ID               int64  `json:"-" db:"id"`
	PublicID         string `json:"id" db:"public_id"`
	TaskID           int64  `json:"-" db:"id_task"`
	ResolverUsername string `json:"resolver_username" db:"resolver_username"`

	State      string     `json:"state" db:"state"`
	InstanceID *uint64    `json:"instance_id,omitempty" db:"instance_id"`
	Created    time.Time  `json:"created,omitempty" db:"created"`
	LastStart  *time.Time `json:"last_start,omitempty" db:"last_start"`
	LastStop   *time.Time `json:"last_stop,omitempty" db:"last_stop"`
	NextRetry  *time.Time `json:"next_retry,omitempty" db:"next_retry"`
	RunCount   int        `json:"run_count" db:"run_count"`
	RunMax     int        `json:"run_max" db:"run_max"`

	CryptKey            []byte `json:"-" db:"crypt_key"` // key for encrypting steps (itself encrypted with master key)
	EncryptedInput      []byte `json:"-" db:"encrypted_resolver_input"`
	EncryptedSteps      []byte `json:"-" db:"encrypted_steps"`       // encrypted Steps map
	StepsCompressionAlg string `json:"-" db:"steps_compression_alg"` // compression algorithm used

	BaseConfigurations map[string]json.RawMessage `json:"base_configurations" db:"base_configurations"`
}

// Create inserts a new resolution in DB
func Create(dbp zesty.DBProvider, t *task.Task, resolverInputs map[string]interface{}, resUser string, autorun bool, delayedUntil *time.Time) (r *Resolution, err error) {
	defer errors.DeferredAnnotatef(&err, "Failed to create Resolution")

	if t.State == task.StateWontfix {
		err = errors.BadRequestf("Task is in state %s", task.StateWontfix)
		return nil, err
	}

	r = &Resolution{
		DBModel: DBModel{
			PublicID:         uuid.Must(uuid.NewV4()).String(),
			TaskID:           t.ID,
			ResolverUsername: resUser,
			State:            StateTODO,
			Created:          now.Get(),
		},
		TaskPublicID: t.PublicID,
		Values:       values.NewValues(),
	}

	if autorun {
		// delayed autorun, make use of the retry collector for simplicity
		if delayedUntil != nil {
			r.State = StateToAutorunDelayed
			r.NextRetry = delayedUntil
		} else {
			r.State = StateToAutorun
		}
	}

	// force empty to stop using old crypto code
	r.CryptKey = []byte{}

	tt, err := tasktemplate.LoadFromID(dbp, t.TemplateID)
	if err != nil {
		return nil, err
	}

	r.setSteps(tt.Steps)
	for stepName := range r.Steps {
		r.Steps[stepName].Name = stepName
		r.SetStepState(stepName, step.StateTODO)
	}

	if tt.RetryMax != nil {
		r.RunMax = *tt.RetryMax
	} else {
		r.RunMax = utask.DefaultRetryMax
	}

	r.BaseConfigurations = tt.BaseConfigurations

	c, err := compress.Get(utask.StepsCompressionAlg)
	if err != nil {
		return nil, err
	}

	r.StepsCompressionAlg = utask.StepsCompressionAlg

	jsonSteps, err := json.Marshal(r.Steps)
	if err != nil {
		return nil, err
	}

	compressedSteps, err := c.Compress(jsonSteps)
	if err != nil {
		return nil, err
	}

	encryptedSteps, err := models.EncryptionKey.Encrypt(compressedSteps, []byte(r.PublicID))
	if err != nil {
		return nil, err
	}

	dst := make([]byte, hex.EncodedLen(len(encryptedSteps)))
	hex.Encode(dst, encryptedSteps)
	r.EncryptedSteps = encryptedSteps

	err = tt.ValidateResolverInputs(resolverInputs)
	if err != nil {
		return nil, err
	}

	r.SetInput(resolverInputs)
	encrInput, err := models.EncryptionKey.EncryptMarshal(r.ResolverInput, []byte(r.PublicID))
	if err != nil {
		return nil, err
	}
	r.EncryptedInput = []byte(encrInput)

	err = dbp.DB().Insert(&r.DBModel)
	if err != nil {
		return nil, pgjuju.Interpret(err)
	}

	// register validation duration
	task.RegisterValidationTime(tt.Name, t.Created)

	return r, nil
}

// LoadFromPublicID returns a single task resolution given its public ID
func LoadFromPublicID(dbp zesty.DBProvider, publicID string) (*Resolution, error) {
	return load(dbp, publicID, false, false)
}

// LoadLockedFromPublicID returns a single task resolution given its public ID
// while acquiring a lock on its DB row, to ensure only this instance keeps access to it
// until the surrounding transaction is done (ensure that only this instance of µTask
// collects this resolution for execution). If another instance already has a lock, it will
// wait until the other instance release it.
func LoadLockedFromPublicID(dbp zesty.DBProvider, publicID string) (*Resolution, error) {
	return load(dbp, publicID, true, false)
}

// LoadLockedNoWaitFromPublicID returns a single task resolution given its public ID
// while acquiring a lock on its DB row, to ensure only this instance keeps access to it
// until the surrounding transaction is done (ensure that only this instance of µTask
// collects this resolution for execution). If another instance already has a lock, it will
// directly return an error.
func LoadLockedNoWaitFromPublicID(dbp zesty.DBProvider, publicID string) (*Resolution, error) {
	return load(dbp, publicID, true, true)
}

func load(dbp zesty.DBProvider, publicID string, locked bool, lockNoWait bool) (r *Resolution, err error) {
	defer errors.DeferredAnnotatef(&err, "Failed to load resolution from public id")

	sel := rSelector
	if locked && lockNoWait {
		sel = sel.Suffix(
			`FOR NO KEY UPDATE OF "resolution" NOWAIT`,
		)
	} else if locked {
		sel = sel.Suffix(
			`FOR NO KEY UPDATE OF "resolution"`,
		)
	}

	query, params, err := sel.Where(
		squirrel.Eq{`"resolution".public_id`: publicID},
	).ToSql()
	if err != nil {
		return nil, err
	}

	err = dbp.DB().SelectOne(&r, query, params...)
	if err != nil {
		return nil, pgjuju.Interpret(err)
	}

	r.Values = values.NewValues()

	c, err := compress.Get(r.StepsCompressionAlg)
	if err != nil {
		return nil, err
	}

	dst := make([]byte, hex.DecodedLen(len(r.EncryptedSteps)))

	// if we can't hex Decode, we might be in the case of a Resolution row in database that was
	// created between the v1.21.1 and v1.21.3 that was bugged, and failed to hex Encode/Decode the
	// ciphered data. We need to keep backward compatibility for those, but this should not happen
	// often.
	// See https://github.com/ovh/utask/commit/bf23fbb10b62bb487ac4ea01b1e519f85480e58b and migration
	// from symmecrypt.Key.DecryptMarshal to symmecrypt.Key.Decrypt
	if _, err = hex.Decode(dst, r.EncryptedSteps); err != nil {
		dst = r.EncryptedSteps
	}

	compressedSteps, err := models.EncryptionKey.Decrypt(dst, []byte(r.PublicID))
	if err != nil {
		return nil, err
	}

	jsonSteps, err := c.Decompress(compressedSteps)
	if err != nil {
		return nil, err
	}

	st := make(map[string]*step.Step)
	if err := utils.JSONnumberUnmarshal(bytes.NewReader(jsonSteps), &st); err != nil {
		return nil, err
	}
	r.setSteps(st)

	input := make(map[string]interface{})
	err = models.EncryptionKey.DecryptMarshal(string(r.EncryptedInput), &input, []byte(r.PublicID))
	if err != nil {
		return nil, err
	}
	r.SetInput(input)

	r.BuildStepTree()

	return r, nil
}

// BuildStepTree re-generates a dependency graph for the steps
// useful for determining elligibility of any given step for execution
func (r *Resolution) BuildStepTree() {
	treeIdx := map[string][]string{}
	treeIdxPrune := map[string][]string{}
	stepList := []string{}
	for name, s := range r.Steps {
		for _, dep := range s.Dependencies {
			dName, dState := step.DependencyParts(dep)
			if treeIdx[dName] == nil {
				treeIdx[dName] = []string{}
			}
			treeIdx[dName] = append(treeIdx[dName], name)
			if dState[0] != step.StateAny {
				if treeIdxPrune[dName] == nil {
					treeIdxPrune[dName] = []string{}
				}
				treeIdxPrune[dName] = append(treeIdxPrune[dName], name)
			}
		}
		stepList = append(stepList, name)
	}

	r.StepList = stepList
	r.StepTreeIndex = treeIdx
	r.StepTreeIndexPrune = treeIdxPrune
	if r.ForeachChildrenAlreadyContracted == nil {
		r.ForeachChildrenAlreadyContracted = map[string]bool{}
	}
}

// ListResolutions returns a collection of existing task resolutions
// optionally filtered by task, resolver username, state or instance ID
// a page size can be passed to limit the size of the collection, and also
// a pointer to the previous page's last element
func ListResolutions(dbp zesty.DBProvider, t *task.Task, resolverUsername *string, state *string, instanceID *uint64, pageSize uint64, last *string) (r []*Resolution, err error) {
	defer errors.DeferredAnnotatef(&err, "Failed to list resolutions")

	sel := rSelector.OrderBy(
		`"resolution".id`,
	).Limit(
		pageSize,
	)

	if t != nil {
		sel = sel.Where(squirrel.Eq{`"task".public_id`: t.PublicID})
	}

	if resolverUsername != nil {
		sel = sel.Where(squirrel.Eq{`"resolution".resolver_username`: *resolverUsername})
	}

	if state != nil {
		sel = sel.Where(squirrel.Eq{`"resolution".state`: *state})
	}

	if instanceID != nil {
		sel = sel.Where(squirrel.Eq{`"resolution".instance_id`: *instanceID})
	}

	if last != nil {
		lastR, err := LoadFromPublicID(dbp, *last)
		if err != nil {
			return nil, err
		}
		sel = sel.Where(`"resolution".id > ?`, lastR.ID)
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

// Update commits any changes of state in Resolution to DB
func (r *Resolution) Update(dbp zesty.DBProvider) (err error) {
	defer errors.DeferredAnnotatef(&err, "Failed to update resolution")

	// TODO tasktemplate.ValidateResolverInput !!

	c, err := compress.Get(r.StepsCompressionAlg)
	if err != nil {
		return err
	}

	jsonSteps, err := json.Marshal(r.Steps)
	if err != nil {
		return err
	}

	compressedSteps, err := c.Compress(jsonSteps)
	if err != nil {
		return err
	}

	dst := make([]byte, hex.EncodedLen(len(compressedSteps)))
	hex.Encode(dst, compressedSteps)
	encryptedSteps, err := models.EncryptionKey.Encrypt(compressedSteps, []byte(r.PublicID))
	if err != nil {
		return err
	}
	r.EncryptedSteps = encryptedSteps

	encrInput, err := models.EncryptionKey.EncryptMarshal(r.ResolverInput, []byte(r.PublicID))
	if err != nil {
		return err
	}
	r.EncryptedInput = []byte(encrInput)

	// force empty to stop using old crypto code
	r.CryptKey = []byte{}

	var nextRetry time.Time
	if r.NextRetry == nil || r.NextRetry.IsZero() {
		// No local next_retry value. It may have been set in DB since the last read, so we need to refresh it
		nextRetry, err = getNextRetry(dbp, r.ID)
		if err != nil {
			return err
		}
	} else {
		// Updating the next_retry individually to prevent overriding it with a nil value or a later date
		nextRetry, err = r.UpdateNextRetry(dbp, *r.NextRetry)
		if err != nil {
			return errors.Annotatef(err, "failed to update resolution's next_retry")
		}
	}
	r.NextRetry = &nextRetry

	rows, err := dbp.DB().Update(&r.DBModel)
	if err != nil {
		return pgjuju.Interpret(err)
	} else if rows == 0 {
		return errors.NotFoundf("No such resolution to update: %s", r.PublicID)
	}

	return nil
}

// UpdateNextRetry updates the Resolution's next_retry field while respecting the current next_retry value in DB. It
// can only shorten the time before the resolution will be retried next, not increase it.
func (r *Resolution) UpdateNextRetry(dbp zesty.DBProvider, newNextRetry time.Time) (time.Time, error) {
	sp, err := dbp.TxSavepoint()
	if err != nil {
		return time.Time{}, err
	}
	defer dbp.RollbackTo(sp) //nolint:errcheck

	// Using the Resolution's ID to soft lock the update and prevent concurrent updates
	if _, err := dbp.DB().Exec(`SELECT pg_advisory_xact_lock($1)`, r.ID); err != nil {
		return time.Time{}, err
	}

	query, params, err := sqlgenerator.PGsql.
		Update("resolution").
		Where(squirrel.Eq{"public_id": r.PublicID}).
		// Is the new next_retry valid
		Where("? > last_start", newNextRetry).
		// And, is the current next_retry outdated or further in time than the new one
		Where("(next_retry < last_start OR ? < next_retry OR next_retry IS NULL)", newNextRetry).
		Set("next_retry", newNextRetry).
		Suffix("RETURNING next_retry").
		ToSql()
	if err != nil {
		return time.Time{}, err
	}

	res, err := dbp.DB().Query(query, params...)
	if err != nil {
		return time.Time{}, err
	}
	defer res.Close()

	if !res.Next() {
		// Update returned no result
		if err := res.Err(); err != nil {
			// An error happened when reading the query's result
			return time.Time{}, err
		}

		// The nextRetry wasn't updated, we need to fetch its current value
		nextRetry, err := getNextRetry(dbp, r.ID)
		if err != nil {
			return time.Time{}, err
		}
		return nextRetry, nil
	}

	var nextRetry time.Time
	if err := res.Scan(&nextRetry); err != nil {
		return time.Time{}, err
	}

	return nextRetry, dbp.Commit()
}

// Delete removes the Resolution from DB
func (r *Resolution) Delete(dbp zesty.DBProvider) (err error) {
	defer errors.DeferredAnnotatef(&err, "Failed to update resolution")

	rows, err := dbp.DB().Delete(&r.DBModel)
	if err != nil {
		return pgjuju.Interpret(err)
	} else if rows == 0 {
		return errors.NotFoundf("No such resolution to delete: %s", r.PublicID)
	}

	return nil
}

// RotateResolutions loads all resolutions stored in DB and makes sure
// that their cyphered content has been handled with the latest
// available storage key
func RotateResolutions(dbp zesty.DBProvider) (err error) {
	defer errors.DeferredAnnotatef(&err, "Failed to rotate encrypted resolutions to new key")

	var last string
	for {
		var lastID *string
		if last != "" {
			lastID = &last
		}
		resolutions, err := ListResolutions(dbp,
			nil, // task
			nil, // resolverUsername
			nil, // state
			nil, // instanceID
			utask.MaxPageSize,
			lastID)
		if err != nil {
			return err
		}
		if len(resolutions) == 0 {
			break
		}
		last = resolutions[len(resolutions)-1].PublicID

		for _, r := range resolutions {
			sp, err := dbp.TxSavepoint()
			if err != nil {
				return err
			}
			// load resolution locked (decrypt)
			res, err := LoadLockedFromPublicID(dbp, r.PublicID)
			if err != nil {
				dbp.RollbackTo(sp)
				return err
			}
			// update resolution (encrypt)
			if err := res.Update(dbp); err != nil {
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

// SetState changes the Resolution's state
func (r *Resolution) SetState(state string) {
	r.State = state
}

// SetInstanceID assigns an instance ID to the Resolution
// In other words, the current running instance of µTask
// 'acquires' this resolution, ensuring that other instances
// will not perform conflicting work on the Resolution
func (r *Resolution) SetInstanceID(id uint64) {
	r.InstanceID = &id
}

// SetLastStart records the last time an execution cycle began
func (r *Resolution) SetLastStart(t time.Time) {
	r.LastStart = &t
}

// SetLastStop records the last time an execution cycle ended
func (r *Resolution) SetLastStop(t time.Time) {
	r.LastStop = &t
}

// SetStep re-assigns a named step, with its updated state and data
func (r *Resolution) SetStep(name string, s *step.Step) {
	r.Steps[name] = s
}

// IncrementRunCount records that a new execution has been performed for this resolution
// incrementing the total count of executions (relevant to keep track of RunCount < MaxRetries)
func (r *Resolution) IncrementRunCount() {
	r.RunCount++
}

// SetNextRetry assigns a point in time when the resolution will become eligible for execution
func (r *Resolution) SetNextRetry(t time.Time) {
	r.NextRetry = &t
}

// ExtendRunMax adds an arbitraty amount to the number of allowed executions for the resolution
func (r *Resolution) ExtendRunMax(i int) {
	r.RunMax += i
}

// ClearOutputs empties the sensitive content of steps
// -> renders a simplified view of a resolution
func (r *Resolution) ClearOutputs() {
	for _, s := range r.Steps {
		s.Output = nil
		s.Metadata = nil
		s.Children = nil
		s.Item = nil
	}
	r.ResolverInput = map[string]interface{}{}
}

///

func (r *Resolution) setSteps(st map[string]*step.Step) {
	r.Steps = st
	for name, s := range r.Steps {
		r.Values.SetOutput(name, s.Output)
		r.Values.SetMetadata(name, s.Metadata)
		r.Values.SetChildren(name, s.Children)
		r.Values.SetError(s.Name, s.Error)
		r.Values.SetState(s.Name, s.State)
	}
}

func (r *Resolution) SetStepState(stepName, state string) {
	oldState := r.Steps[stepName].State
	r.Steps[stepName].State = state

	// notify about step state modification
	if oldState != state {
		if dbp, err := zesty.NewDBProvider(utask.DBName); err == nil {
			if t, err := task.LoadFromID(dbp, r.TaskID); err == nil {
				t.NotifyStepState(stepName, state)
			}
		}
	}

	if r.Values == nil {
		return
	}
	r.Values.SetState(stepName, state)
}

// SetInput stores the inputs provided by the task's resolver
func (r *Resolution) SetInput(input map[string]interface{}) {
	r.ResolverInput = input
}

// getNextRetry fetches from the database the current next_retry value of the resolution with given ID
func getNextRetry(dbp zesty.DBProvider, resolutionID int64) (time.Time, error) {
	var tmpRes Resolution
	err := dbp.DB().SelectOne(&tmpRes, `SELECT next_retry FROM resolution WHERE id = $1`, resolutionID)
	if err != nil {
		return time.Time{}, err
	}

	if tmpRes.NextRetry == nil {
		return time.Time{}, nil
	}

	return *tmpRes.NextRetry, nil
}

var rSelector = sqlgenerator.PGsql.Select(
	`"resolution".id, "resolution".public_id, "resolution".id_task, "resolution".resolver_username, "resolution".state, "resolution".instance_id, "resolution".created, "resolution".last_start, "resolution".last_stop, "resolution".next_retry, "resolution".run_count, "resolution".run_max, "resolution".crypt_key, "resolution".encrypted_steps, "resolution".steps_compression_alg, "resolution".encrypted_resolver_input, "resolution".base_configurations, "task".public_id as task_public_id, "task".title as task_title`,
).From(
	`"resolution"`,
).OrderBy(
	`"resolution".id`,
).Join(
	`"task" on "task".id = "resolution".id_task`,
)

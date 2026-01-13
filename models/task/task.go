package task

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/gofrs/uuid"
	"github.com/juju/errors"
	"github.com/lib/pq"
	"github.com/loopfz/gadgeto/zesty"

	"github.com/ovh/utask"
	"github.com/ovh/utask/db/pgjuju"
	"github.com/ovh/utask/db/sqlgenerator"
	"github.com/ovh/utask/engine/input"
	"github.com/ovh/utask/engine/values"
	"github.com/ovh/utask/models"
	"github.com/ovh/utask/models/tasktemplate"
	"github.com/ovh/utask/pkg/notify"
	"github.com/ovh/utask/pkg/now"
	"github.com/ovh/utask/pkg/utils"
)

// possible task states
const (
	StateRunning   = "RUNNING"
	StateDone      = "DONE"
	StateTODO      = "TODO"    // default on creation
	StateBlocked   = "BLOCKED" // not automatically retriable, 400 bad requests, etc..
	StateDelayed   = "DELAYED"
	StateCancelled = "CANCELLED"
	StateWontfix   = "WONTFIX"
	StateWaiting   = "WAITING"
)

// StepError holds an error and the name of the step from where it originated
type StepError struct {
	Step  string `json:"step"`
	Error string `json:"error"`
}

// Task is the full representation of a requested process on µTask
// A task is necessarily derived from a template, the formal description of the process
// to be executed, plus inputs provided by the requester
// The execution of the task will be handled through another structure, its resolution
// When the resolution is finalized, results will be committed back to the Task structure
// Comments can be added to a task, by all parties involved (requester and resolver)
// A task can be made visible to third parties by adding their usernames to the watcher_usernames list
type Task struct {
	DBModel
	TemplateName     string                 `json:"template_name" db:"template_name"`
	Input            map[string]interface{} `json:"input" db:"-"`
	Result           map[string]interface{} `json:"result,omitempty" db:"-"`
	ResultStr        string                 `json:"-" db:"-"`
	Resolution       *string                `json:"resolution,omitempty" db:"resolution_public_id"`
	LastStart        *time.Time             `json:"last_start,omitempty" db:"last_start"`
	LastStop         *time.Time             `json:"last_stop,omitempty" db:"last_stop"`
	ResolverUsername *string                `json:"resolver_username,omitempty" db:"resolver_username"`
	Comments         []*Comment             `json:"comments,omitempty" db:"-"`
	Batch            *string                `json:"batch,omitempty" db:"batch_public_id"`
	Errors           []StepError            `json:"errors,omitempty" db:"-"`
	ResolverInputs   []input.Input          `json:"resolver_inputs,omitempty" db:"resolver_inputs"`
}

// DBModel is the "strict" representation of a task in DB, as expressed in SQL schema
type DBModel struct {
	ID                int64             `json:"-" db:"id"`
	PublicID          string            `json:"id" db:"public_id"`
	Title             string            `json:"title" db:"title"`
	TemplateID        int64             `json:"-" db:"id_template"`
	BatchID           *int64            `json:"-" db:"id_batch"`
	RequesterUsername string            `json:"requester_username" db:"requester_username"`
	RequesterGroups   []string          `json:"requester_groups,omitempty" db:"requester_groups"`
	WatcherUsernames  []string          `json:"watcher_usernames,omitempty" db:"watcher_usernames"`
	WatcherGroups     []string          `json:"watcher_groups,omitempty" db:"watcher_groups"`
	ResolverUsernames []string          `json:"resolver_usernames,omitempty" db:"resolver_usernames"`
	ResolverGroups    []string          `json:"resolver_groups,omitempty" db:"resolver_groups"`
	Created           time.Time         `json:"created" db:"created"`
	State             string            `json:"state" db:"state"`
	StepsDone         int               `json:"steps_done" db:"steps_done"`
	StepsTotal        int               `json:"steps_total" db:"steps_total"`
	LastActivity      time.Time         `json:"last_activity" db:"last_activity"`
	Tags              map[string]string `json:"tags,omitempty" db:"tags"`

	CryptKey        []byte `json:"-" db:"crypt_key"` // key for encrypting steps (itself encrypted with master key)
	EncryptedInput  []byte `json:"-" db:"encrypted_input"`
	EncryptedResult []byte `json:"-" db:"encrypted_result"` // encrypted Result
}

// Create inserts a new Task in DB
func Create(dbp zesty.DBProvider, tt *tasktemplate.TaskTemplate, reqUsername string, reqGroups []string, watcherUsernames []string, watcherGroups []string, resolverUsernames []string, resolverGroups []string, input map[string]interface{}, tags map[string]string, b *Batch, delayed bool) (t *Task, err error) {
	defer errors.DeferredAnnotatef(&err, "Failed to create new Task")

	initState := StateTODO
	if delayed {
		initState = StateDelayed
	}
	t = &Task{
		DBModel: DBModel{
			PublicID:          uuid.Must(uuid.NewV4()).String(),
			TemplateID:        tt.ID,
			RequesterUsername: reqUsername,
			RequesterGroups:   reqGroups,
			WatcherUsernames:  watcherUsernames,
			WatcherGroups:     watcherGroups,
			ResolverUsernames: resolverUsernames,
			ResolverGroups:    resolverGroups,
			Created:           now.Get(),
			LastActivity:      now.Get(),
			StepsTotal:        len(tt.Steps),
			State:             initState,
		},
		TemplateName: tt.Name,
		Result:       tt.ResultFormat,
		Input:        tt.FilterInputs(input),
	}

	if b != nil {
		t.BatchID = &b.ID
	}

	// force empty to stop using old crypto code
	t.CryptKey = []byte{}

	resultB, err := utils.JSONMarshal(t.Result)
	if err != nil {
		return nil, err
	}
	t.ResultStr = string(resultB)

	t.EncryptedResult, err = models.EncryptionKey.Encrypt([]byte(t.ResultStr), []byte(t.PublicID))
	if err != nil {
		return nil, err
	}

	encrInput, err := models.EncryptionKey.EncryptMarshal(t.Input, []byte(t.PublicID))
	if err != nil {
		return nil, err
	}
	t.EncryptedInput = []byte(encrInput)

	err = t.Valid(tt)
	if err != nil {
		return nil, err
	}

	// title can be computed if input values are valid
	v := values.NewValues()
	v.SetInput(input)
	v.SetVariables(tt.Variables)
	t.ExportTaskInfos(v) // make task-specific info available for title
	title, err := v.Apply(tt.TitleFormat, nil, "")
	if err != nil {
		return nil, err
	}
	t.Title = string(title)

	// Merge input tags into template tags.
	mergedTags := make(map[string]string)
	for k, v := range tt.Tags {
		mergedTags[k] = v
	}
	for k, v := range tags {
		mergedTags[k] = v
	}
	if err := t.SetTags(mergedTags, v); err != nil {
		return nil, err
	}

	err = dbp.DB().Insert(&t.DBModel)
	if err != nil {
		return nil, pgjuju.Interpret(err)
	}

	notificationAllowedResolverUsernames := []string{}
	notificationAllowedResolverUsernames = append(notificationAllowedResolverUsernames, tt.AllowedResolverUsernames...)
	if tt.AllowAllResolverUsernames {
		notificationAllowedResolverUsernames = append(notificationAllowedResolverUsernames, t.RequesterUsername)
	}
	t.notifyState(notificationAllowedResolverUsernames)

	return t, nil
}

// LoadFromPublicID returns a single task, given its public ID
func LoadFromPublicID(dbp zesty.DBProvider, publicID string) (t *Task, err error) {
	return loadFromPublicID(dbp, publicID, false, true)
}

// LoadLockedFromPublicID returns a single task, given its ID,
// locked for an update transaction, so that only one instance of µTask can
// make a claim on it and avoid collisions with other instances
func LoadLockedFromPublicID(dbp zesty.DBProvider, publicID string) (t *Task, err error) {
	return loadFromPublicID(dbp, publicID, true, true)
}

func loadFromPublicID(dbp zesty.DBProvider, publicID string, locked, withComments bool) (t *Task, err error) {
	defer errors.DeferredAnnotatef(&err, "Failed to load task from public id")

	sel := tSelector
	if locked {
		sel = sel.Suffix(
			`FOR NO KEY UPDATE OF "task"`,
		)
	}

	query, params, err := sel.Where(
		squirrel.Eq{`"task".public_id`: publicID},
	).ToSql()
	if err != nil {
		return nil, err
	}

	err = dbp.DB().SelectOne(&t, query, params...)
	if err != nil {
		return nil, pgjuju.Interpret(err)
	}

	err = loadDetails(dbp, t, withComments)
	if err != nil {
		return nil, err
	}

	return t, nil
}

// LoadFromID returns a single task, given its "private" ID
// only used internally, this ID is not exposed through µTask's API
func LoadFromID(dbp zesty.DBProvider, ID int64) (t *Task, err error) {
	defer errors.DeferredAnnotatef(&err, "Failed to load task from id")

	query, params, err := tSelector.Where(
		squirrel.Eq{`"task".id`: ID},
	).ToSql()
	if err != nil {
		return nil, err
	}

	err = dbp.DB().SelectOne(&t, query, params...)
	if err != nil {
		return nil, pgjuju.Interpret(err)
	}

	err = loadDetails(dbp, t, true)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func loadDetails(dbp zesty.DBProvider, t *Task, withComments bool) (err error) {
	resBytes, err := models.EncryptionKey.Decrypt(t.EncryptedResult, []byte(t.PublicID))
	if err != nil {
		return err
	}
	t.ResultStr = string(resBytes)

	t.Result = make(map[string]interface{})
	err = utils.JSONnumberUnmarshal(strings.NewReader(t.ResultStr), &t.Result)
	if err != nil {
		return err
	}

	input := make(map[string]interface{})
	err = models.EncryptionKey.DecryptMarshal(string(t.EncryptedInput), &input, []byte(t.PublicID))
	if err != nil {
		return err
	}
	t.Input = input

	if withComments {
		t.Comments, err = LoadCommentsFromTaskID(dbp, t.ID)
	}

	return err
}

// ListFilter holds parameters for filtering a list of tasks
type ListFilter struct {
	RequesterUser                      *string
	PotentialResolverUser              *string
	PotentialResolverGroups            []string
	RequesterOrPotentialResolverUser   *string
	RequesterOrPotentialResolverGroups []string
	Last                               *string
	State                              *string
	Batch                              *Batch
	PageSize                           uint64
	Before                             *time.Time
	After                              *time.Time
	Tags                               map[string]string
	Template                           *string
}

// ListTasks returns a list of tasks, optionally filtered on one or several criteria
func ListTasks(dbp zesty.DBProvider, filter ListFilter) (t []*Task, err error) {
	defer errors.DeferredAnnotatef(&err, "Failed to list tasks")

	sel := tSelector.Limit(
		filter.PageSize,
	).OrderBy(
		`"task".last_activity DESC`,
	)

	if filter.Last != nil {
		lastT, err := LoadFromPublicID(dbp, *filter.Last)
		if err != nil {
			return nil, err
		}
		sel = sel.Where(squirrel.Lt{`"task".last_activity`: lastT.LastActivity})
	}

	if filter.Before != nil {
		sel = sel.Where(squirrel.Lt{`"task".last_activity`: *filter.Before})
	}

	if filter.After != nil {
		sel = sel.Where(squirrel.Gt{`"task".last_activity`: *filter.After})
	}

	if filter.RequesterUser != nil {
		sel = sel.Where(squirrel.Or{
			squirrel.Eq{`"task".requester_username`: *filter.RequesterUser},
			squirrel.Expr(`"task".watcher_usernames @> ?::jsonb`, strconv.Quote(*filter.RequesterUser)),
		})
	}

	if filter.PotentialResolverGroups != nil && len(filter.PotentialResolverGroups) > 0 {
		argGroups, err := pq.Array(filter.PotentialResolverGroups).Value()
		if err != nil {
			return nil, err
		}

		if filter.PotentialResolverUser != nil {
			argUser := strconv.Quote(*filter.PotentialResolverUser)
			sel = sel.Where(squirrel.Or{
				// User conditions
				squirrel.Expr(`"task_template".allowed_resolver_usernames @> ?::jsonb`, argUser),
				squirrel.Expr(`"task".resolver_usernames @> ?::jsonb`, argUser),
				// Group conditions
				// To escape "?", insert two "?" (see https://github.com/Masterminds/squirrel/pull/32)
				squirrel.Expr(`"task_template".allowed_resolver_groups ??| ?`, argGroups),
				squirrel.Expr(`"task".resolver_groups ??| ?`, argGroups),
			})
		} else {
			sel = sel.Where(squirrel.Or{
				// To escape "?", insert two "?" (see https://github.com/Masterminds/squirrel/pull/32)
				squirrel.Expr(`"task_template".allowed_resolver_groups ??| ?`, argGroups),
				squirrel.Expr(`"task".resolver_groups ??| ?`, argGroups),
			})
		}
	} else if filter.PotentialResolverUser != nil {
		arg := strconv.Quote(*filter.PotentialResolverUser)
		sel = sel.Where(squirrel.Or{
			squirrel.Expr(`"task_template".allowed_resolver_usernames @> ?::jsonb`, arg),
			squirrel.Expr(`"task".resolver_usernames @> ?::jsonb`, arg),
		})
	}

	if filter.RequesterOrPotentialResolverGroups != nil && len(filter.RequesterOrPotentialResolverGroups) > 0 {
		argGroups, err := pq.Array(filter.RequesterOrPotentialResolverGroups).Value()
		if err != nil {
			return nil, err
		}

		if filter.RequesterOrPotentialResolverUser != nil {
			argUser := strconv.Quote(*filter.RequesterOrPotentialResolverUser)
			sel = sel.Where(squirrel.Or{
				// User conditions
				squirrel.Eq{`"task".requester_username`: *filter.RequesterOrPotentialResolverUser},
				squirrel.Expr(`"task".watcher_usernames @> ?::jsonb`, argUser),
				squirrel.Expr(`"task_template".allowed_resolver_usernames @> ?::jsonb`, argUser),
				squirrel.Expr(`"task".resolver_usernames @> ?::jsonb`, argUser),
				// Group conditions
				// To escape "?", insert two "?" (see https://github.com/Masterminds/squirrel/pull/32)
				squirrel.Expr(`"task".watcher_groups ??| ?`, argGroups),
				squirrel.Expr(`"task_template".allowed_resolver_groups ??| ?`, argGroups),
				squirrel.Expr(`"task".resolver_groups ??| ?`, argGroups),
			})
		} else {
			sel = sel.Where(squirrel.Or{
				// To escape "?", insert two "?" (see https://github.com/Masterminds/squirrel/pull/32)
				squirrel.Expr(`"task".watcher_groups ??| ?`, argGroups),
				squirrel.Expr(`"task_template".allowed_resolver_groups ??| ?`, argGroups),
				squirrel.Expr(`"task".resolver_groups ??| ?`, argGroups),
			})
		}
	} else if filter.RequesterOrPotentialResolverUser != nil {
		arg := strconv.Quote(*filter.RequesterOrPotentialResolverUser)
		sel = sel.Where(squirrel.Or{
			squirrel.Eq{`"task".requester_username`: *filter.RequesterOrPotentialResolverUser},
			squirrel.Expr(`"task".watcher_usernames @> ?::jsonb`, arg),
			squirrel.Expr(`"task_template".allowed_resolver_usernames @> ?::jsonb`, arg),
			squirrel.Expr(`"task".resolver_usernames @> ?::jsonb`, arg),
		})
	}

	if filter.State != nil {
		sel = sel.Where(squirrel.Eq{`"task".state`: *filter.State})
	}

	if filter.Batch != nil {
		sel = sel.Where(squirrel.Eq{`"task".id_batch`: filter.Batch.ID})
	}

	if filter.Tags != nil && len(filter.Tags) > 0 {
		b, err := json.Marshal(filter.Tags)
		if err != nil {
			return nil, err
		}
		sel = sel.Where(`"task".tags @> ?::jsonb`, string(b))
	}

	if filter.Template != nil {
		sel = sel.Where(squirrel.Eq{`"task_template".name`: *filter.Template})
	}

	query, params, err := sel.ToSql()
	if err != nil {
		return nil, err
	}

	_, err = dbp.DB().Select(&t, query, params...)
	if err != nil {
		return nil, pgjuju.Interpret(err)
	}

	return t, nil
}

// Update commits changes to a task's state to DB
// A flag allows to skip validation: only exposed internally for special
// situations, where a task could be out of sync with its template
func (t *Task) Update(dbp zesty.DBProvider, skipValidation, recordLastActivity bool) (err error) {
	defer errors.DeferredAnnotatef(&err, "Failed to update task")

	resultB, err := utils.JSONMarshal(t.Result)
	if err != nil {
		return err
	}
	t.ResultStr = string(resultB)

	t.EncryptedResult, err = models.EncryptionKey.Encrypt([]byte(t.ResultStr), []byte(t.PublicID))
	if err != nil {
		return err
	}

	encrInput, err := models.EncryptionKey.EncryptMarshal(t.Input, []byte(t.PublicID))
	if err != nil {
		return err
	}
	t.EncryptedInput = []byte(encrInput)

	// force empty to stop using old crypto code
	t.CryptKey = []byte{}

	tt, err := tasktemplate.LoadFromID(dbp, t.TemplateID)
	if err != nil {
		return err
	}

	if !skipValidation {
		err = t.Valid(tt)
		if err != nil {
			return err
		}

		// re-template task title in case inputs were updated
		v := values.NewValues()
		v.SetInput(t.Input)
		v.SetVariables(tt.Variables)
		t.ExportTaskInfos(v)
		title, err := v.Apply(tt.TitleFormat, nil, "")
		if err != nil {
			return err
		}
		t.Title = string(title)
	}

	if recordLastActivity {
		t.LastActivity = now.Get()
	}

	rows, err := dbp.DB().Update(&t.DBModel)
	if err != nil {
		return pgjuju.Interpret(err)
	} else if rows == 0 {
		return errors.NotFoundf("No such task to update: %s", t.PublicID)
	}

	return nil
}

// Delete removes a task from DB
func (t *Task) Delete(dbp zesty.DBProvider) (err error) {
	defer errors.DeferredAnnotatef(&err, "Failed to delete task")

	rows, err := dbp.DB().Delete(&t.DBModel)
	if err != nil {
		return pgjuju.Interpret(err)
	} else if rows == 0 {
		return errors.NotFoundf("No such task to delete: %s", t.PublicID)
	}

	return nil
}

// RotateTasks loads all tasks stored in DB and makes sure
// that their cyphered content has been handled with the latest
// available storage key
func RotateTasks(dbp zesty.DBProvider) (err error) {
	defer errors.DeferredAnnotatef(&err, "Failed to rotate encrypted tasks to new key")

	var last string
	for {
		var lastID *string
		if last != "" {
			lastID = &last
		}
		// load all tasks
		tasks, err := ListTasks(dbp, ListFilter{
			PageSize: utask.MaxPageSize,
			Last:     lastID,
		})
		if err != nil {
			return err
		}
		if len(tasks) == 0 {
			break
		}
		last = tasks[len(tasks)-1].PublicID

		for _, t := range tasks {
			sp, err := dbp.TxSavepoint()
			if err != nil {
				return err
			}
			// load task locked without comments (decrypt)
			tsk, err := loadFromPublicID(dbp, t.PublicID, true, false)
			if err != nil {
				dbp.RollbackTo(sp)
				return err
			}
			// update task (encrypt)
			if err := tsk.Update(dbp,
				true,  // skip validation
				false, // do not change lastActivity value
			); err != nil {
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

// SetWatcherUsernames sets the list of watchers for the task
func (t *Task) SetWatcherUsernames(watcherUsernames []string) {
	t.WatcherUsernames = watcherUsernames
}

// SetWatcherGroups sets the list of watcher groups for the task
func (t *Task) SetWatcherGroups(watcherGroups []string) {
	t.WatcherGroups = watcherGroups
}

// SetInput sets the provided input for the task
func (t *Task) SetInput(input map[string]interface{}) {
	t.Input = input
}

// SetResult consolidates values collected during resolution into the task's final result
func (t *Task) SetResult(values *values.Values) error {
	return applyTemplateToMap(t.Result, values)
}

func applyTemplateToMap(m map[string]interface{}, values *values.Values) error {
	// templating on map keys
	for k, v := range m {
		tempk, err := values.Apply(k, nil, "")
		if err != nil {
			return fmt.Errorf("failed to template: %s", err.Error())
		}
		if string(tempk) != k {
			m[string(tempk)] = v
			delete(m, k)
		}
	}
	// templating on map values
	for k, v := range m {
		switch v := v.(type) {
		case map[string]interface{}:
			if err := applyTemplateToMap(v, values); err != nil {
				return err
			}
		case string:
			tempv, err := values.Apply(v, nil, "")
			if err != nil {
				return fmt.Errorf("failed to template: %s", err.Error())
			}
			m[k] = string(tempv)
		}
	}
	return nil
}

// SetState updates the task's state
func (t *Task) SetState(s string) {
	var notify bool
	if t.State != s {
		notify = true
	}

	t.State = s

	if notify {
		t.notifyState(nil)
	}
}

func (t *Task) SetTags(tags map[string]string, values *values.Values) error {
	t.Tags = tags
	if values == nil {
		return nil
	}
	for k, v := range t.Tags {
		tempv, err := values.Apply(v, nil, "")
		if err != nil {
			return fmt.Errorf("failed to template: %s", err.Error())
		}
		t.Tags[k] = string(tempv)
	}
	return nil
}

// Valid asserts that the task holds valid data: the state is among accepted states,
// and input is present and valid given the template spec
func (t *Task) Valid(tt *tasktemplate.TaskTemplate) error {
	switch t.State {
	case StateTODO, StateRunning, StateDone, StateBlocked, StateCancelled, StateWontfix, StateWaiting, StateDelayed:
		break
	default:
		return errors.BadRequestf("Wrong state: %s", t.State)
	}
	if t.Input == nil {
		return errors.BadRequestf("Missing input")
	}

	return tt.ValidateInputs(t.Input)
}

// ExportTaskInfos records task-specific data to a Values structure
func (t *Task) ExportTaskInfos(values *values.Values) {
	m := make(map[string]interface{})

	m["task_id"] = t.PublicID
	m["created"] = t.Created
	m["requester_username"] = t.RequesterUsername
	if len(t.RequesterGroups) > 0 {
		m["requester_groups"] = strings.Join(t.RequesterGroups, utask.GroupsSeparator)
	}
	if t.ResolverUsername != nil {
		m["resolver_username"] = t.ResolverUsername
	}
	if t.ResolverUsername != nil && len(t.ResolverGroups) > 0 {
		m["resolver_groups"] = strings.Join(t.ResolverGroups, utask.GroupsSeparator)
	}
	if len(t.WatcherUsernames) > 0 {
		m["watcher_usernames"] = strings.Join(t.WatcherUsernames, utask.UsernamesSeparator)
	}
	if len(t.WatcherGroups) > 0 {
		m["watcher_groups"] = strings.Join(t.WatcherGroups, utask.GroupsSeparator)
	}
	m["last_activity"] = t.LastActivity
	m["region"] = utask.FRegion
	if t.Resolution != nil {
		m["resolution_id"] = t.Resolution
	}

	values.SetTaskInfos(m)
}

var (
	tSelector = sqlgenerator.PGsql.Select(
		`"task".id, "task".public_id, "task".title, "task".id_template, "task".id_batch, "task".requester_username, "task".requester_groups, "task".watcher_usernames, "task".watcher_groups, "task".created, "task".state, "task".tags, "task".steps_done, "task".steps_total, "task".crypt_key, "task".encrypted_input, "task".encrypted_result, "task".last_activity, "task".resolver_usernames, "task".resolver_groups, "task_template".name as template_name, "task_template".resolver_inputs as resolver_inputs, "resolution".public_id as resolution_public_id, "resolution".last_start as last_start, "resolution".last_stop as last_stop, "resolution".resolver_username as resolver_username, "batch".public_id as batch_public_id`,
	).From(
		`"task"`,
	).Join(
		`"task_template" ON "task_template".id = "task".id_template`,
	).LeftJoin(
		`"resolution" ON "resolution".id_task = "task".id`,
	).LeftJoin(
		`"batch" ON "batch".id = "task".id_batch`,
	)
)

func (t *Task) notifyState(potentialResolvers []string) {
	tsu := &notify.TaskStateUpdate{
		Title:              t.Title,
		PublicID:           t.PublicID,
		State:              t.State,
		TemplateName:       t.TemplateName,
		PotentialResolvers: potentialResolvers,
		RequesterUsername:  t.RequesterUsername,
		ResolverUsername:   t.ResolverUsername,
		StepsDone:          t.StepsDone,
		StepsTotal:         t.StepsTotal,
		Tags:               t.Tags,
	}
	if t.Resolution != nil {
		tsu.ResolutionPublicID = *t.Resolution
	}

	notify.Send(
		notify.WrapTaskStateUpdate(tsu),
		notify.ListActions().TaskStateUpdateAction,
	)
}

func (t *Task) NotifyValidationRequired(tt *tasktemplate.TaskTemplate) {
	notificationAllowedResolverUsernames := []string{}
	if tt != nil {
		notificationAllowedResolverUsernames = append(notificationAllowedResolverUsernames, tt.AllowedResolverUsernames...)
	}
	if tt.AllowAllResolverUsernames {
		notificationAllowedResolverUsernames = append(notificationAllowedResolverUsernames, t.RequesterUsername)
	}

	tv := &notify.TaskValidation{
		Title:              t.Title,
		PublicID:           t.PublicID,
		State:              t.State,
		TemplateName:       t.TemplateName,
		PotentialResolvers: notificationAllowedResolverUsernames,
		RequesterUsername:  t.RequesterUsername,
		Tags:               t.Tags,
	}

	notify.Send(
		notify.WrapTaskValidation(tv),
		notify.ListActions().TaskValidationAction,
	)
}

func (t *Task) NotifyStepState(stepName, stepState string) {
	if t.Resolution == nil || t.ResolverUsername == nil {
		// matches mainly the period where the task is getting created and all steps states are assigned to TODO
		return
	}

	tsu := &notify.TaskStepUpdate{
		Title:              t.Title,
		PublicID:           t.PublicID,
		State:              t.State,
		TemplateName:       t.TemplateName,
		RequesterUsername:  t.RequesterUsername,
		ResolverUsername:   *t.ResolverUsername,
		StepsDone:          t.StepsDone,
		StepsTotal:         t.StepsTotal,
		Tags:               t.Tags,
		StepName:           stepName,
		StepState:          stepState,
		ResolutionPublicID: *t.Resolution,
	}

	notify.Send(
		notify.WrapTaskStepUpdate(tsu),
		notify.ListActions().TaskStepUpdateAction,
	)
}

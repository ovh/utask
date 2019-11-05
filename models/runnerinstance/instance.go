package runnerinstance

import (
	"time"

	"github.com/juju/errors"
	"github.com/loopfz/gadgeto/zesty"
	"github.com/ovh/utask/db/pgjuju"
	"github.com/ovh/utask/db/sqlgenerator"
	"github.com/ovh/utask/pkg/now"
)

// HeartbeatInterval is the duration between an instance's heartbeats
// a heartbeat is a sign of life, committed to DB, for an instance to
// assert that it is still active
// useful to discriminate tasks acquired by a defunct instance
var HeartbeatInterval = time.Second * 30

// Instance represents one active instance of µTask
// with its ID acquired from DB, and its latest heartbeat timestamp
type Instance struct {
	ID        uint64    `db:"id"`
	Heartbeat time.Time `db:"heartbeat"`
}

// Create inserts a new instance object in DB, sets its heartbeat routine in motion
func Create(dbp zesty.DBProvider) (id uint64, err error) {
	defer errors.DeferredAnnotatef(&err, "Failed to create new runner instance")

	i := &Instance{
		Heartbeat: now.Get(),
	}

	err = dbp.DB().Insert(i)
	if err != nil {
		return 0, pgjuju.Interpret(err)
	}

	go func() {
		for {
			time.Sleep(HeartbeatInterval)
			i.heartBeat(dbp)
		}
	}()

	id = i.ID

	return
}

// ListInstances returns a list of µTask instances from DB
func ListInstances(dbp zesty.DBProvider) (ii []*Instance, err error) {
	defer errors.DeferredAnnotatef(&err, "Failed to list runner instances")

	query, params, err := sqlgenerator.PGsql.Select(
		`"runner_instance".id, "runner_instance".heartbeat`,
	).From(
		`"runner_instance"`,
	).ToSql()
	if err != nil {
		return nil, err
	}

	_, err = dbp.DB().Select(&ii, query, params...)
	if err != nil {
		return nil, pgjuju.Interpret(err)
	}

	return
}

// IsDead asserts that an instance is dead (hasn't emitted a heartbeat
// for longer than twice the heartbeat interval)
func (i *Instance) IsDead() bool {
	// leave some margin for declaring an instance dead (twice past its due heartbeat)
	return now.Get().Sub(i.Heartbeat) > 2*HeartbeatInterval
}

func (i *Instance) heartBeat(dbp zesty.DBProvider) error {

	i.Heartbeat = now.Get()

	rows, err := dbp.DB().Update(i)
	if err != nil {
		return pgjuju.Interpret(err)
	} else if rows == 0 {
		return errors.NotFoundf("No such runner instance to update: %d", i.ID)
	}

	return nil
}

// Delete removes a µTask instance from DB
func (i *Instance) Delete(dbp zesty.DBProvider) error {

	rows, err := dbp.DB().Delete(i)
	if err != nil {
		return pgjuju.Interpret(err)
	} else if rows == 0 {
		return errors.NotFoundf("No such runner instance to delete: %d", i.ID)
	}

	return nil
}

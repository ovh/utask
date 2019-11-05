package task

import (
	"github.com/juju/errors"
	"github.com/loopfz/gadgeto/zesty"
	"github.com/ovh/utask/db/pgjuju"
	"github.com/ovh/utask/db/sqlgenerator"
)

type stateCount struct {
	State string  `db:"state"`
	Count float64 `db:"state_count"`
}

// LoadStateCount returns a map containing the count of tasks grouped by state
func LoadStateCount(dbp zesty.DBProvider) (sc map[string]float64, err error) {
	defer errors.DeferredAnnotatef(&err, "Failed to load task stats")

	query, params, err := sqlgenerator.PGsql.Select(`state, count(state) as state_count`).
		From(`"task"`).
		GroupBy(`state`).
		ToSql()
	if err != nil {
		return nil, err
	}

	s := []stateCount{}
	if _, err := dbp.DB().Select(&s, query, params...); err != nil {
		return nil, pgjuju.Interpret(err)
	}

	sc = map[string]float64{
		StateTODO:      0,
		StateBlocked:   0,
		StateRunning:   0,
		StateWontfix:   0,
		StateDone:      0,
		StateCancelled: 0,
	}
	for _, c := range s {
		sc[c.State] = c.Count
	}

	return sc, nil
}

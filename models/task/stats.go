package task

import (
	"time"

	"github.com/juju/errors"
	"github.com/loopfz/gadgeto/zesty"
	"github.com/ovh/utask/db/pgjuju"
	"github.com/ovh/utask/db/sqlgenerator"
	"github.com/ovh/utask/pkg/now"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	validationTimes = promauto.NewSummaryVec(prometheus.SummaryOpts{Name: "utask_valid_times"}, []string{"template"})
	completionTimes = promauto.NewSummaryVec(prometheus.SummaryOpts{Name: "utask_complete_times"}, []string{"template"})
	executionTimes  = promauto.NewSummaryVec(prometheus.SummaryOpts{Name: "utask_exec_times"}, []string{"template"})
)

type stateCount struct {
	State string  `db:"state"`
	Count float64 `db:"state_count"`
}

// RegisterValidationTime computes the duration between the task creation and
// the associated resolution's creation. This metric is then pushed to Prometheus.
func RegisterValidationTime(templateName string, taskCreation time.Time) {
	duration := now.Get().Sub(taskCreation).Seconds()
	validationTimes.WithLabelValues(templateName).Observe(duration)
}

// RegisterTaskTime computes the execution duration and the complete duration
// (from creation to completion) of a task. These metrics are then pushed to Prometheus.
func RegisterTaskTime(templateName string, taskCreation, resCreation time.Time) {
	currentTime := now.Get()

	// Time taken since task creation
	completeTime := currentTime.Sub(taskCreation).Seconds()
	completionTimes.WithLabelValues(templateName).Observe(completeTime)

	// Time taken since resolution was created
	executionTime := currentTime.Sub(resCreation).Seconds()
	executionTimes.WithLabelValues(templateName).Observe(executionTime)
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

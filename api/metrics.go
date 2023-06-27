package api

import (
	"context"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/juju/errors"
	"github.com/loopfz/gadgeto/zesty"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"

	"github.com/ovh/utask"
	"github.com/ovh/utask/models/task"
)

var (
	metrics = promauto.NewGaugeVec(prometheus.GaugeOpts{Name: "utask_task_state"}, []string{"status", "group"})
)

func updateMetrics(dbp zesty.DBProvider) {
	// utask_task_state_per_resolver_group
	statsResolverGroup, err := task.LoadStateCountResolverGroup(dbp)
	if err != nil {
		logrus.Warn(err)
	}
	for group, groupStats := range statsResolverGroup {
		for state, count := range groupStats {
			metrics.WithLabelValues(state, group).Set(count)
		}
	}
}

func collectMetrics(ctx context.Context) {
	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		logrus.Warn(err)
		return
	}

	tick := time.NewTicker(5 * time.Second)

	updateMetrics(dbp)

	go func() {
		for {
			select {
			case <-tick.C:
				updateMetrics(dbp)
			case <-ctx.Done():
				tick.Stop()
				return
			}
		}
	}()
}

type StatsIn struct {
	Tags []string `query:"tag" explode:"true"`
}

// StatsOut aggregates different business stats:
// - a map of task states and their count
type StatsOut struct {
	TaskStates map[string]float64 `json:"task_states"`
}

// Stats handles the http request to fetch µtask statistics
// common to all instances
func Stats(c *gin.Context, in *StatsIn) (*StatsOut, error) {
	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return nil, err
	}

	tags := make(map[string]string, len(in.Tags))
	for _, t := range in.Tags {
		parts := strings.Split(t, "=")
		if len(parts) != 2 {
			return nil, errors.BadRequestf("invalid tag %s", t)
		}
		if parts[0] == "" || parts[1] == "" {
			return nil, errors.BadRequestf("invalid tag %s", t)
		}
		tags[parts[0]] = parts[1]
	}

	out := StatsOut{}
	out.TaskStates, err = task.LoadStateCount(dbp, tags)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

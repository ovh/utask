package api

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/loopfz/gadgeto/zesty"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"

	"github.com/ovh/utask"
	"github.com/ovh/utask/models/task"
)

var (
	metrics = promauto.NewGaugeVec(prometheus.GaugeOpts{Name: "utask_task_state"}, []string{"status"})
)

func collectMetrics(ctx context.Context) {
	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		logrus.Warn(err)
		return
	}

	tick := time.NewTicker(5 * time.Second)
	go func() {
		for {
			select {
			case <-tick.C:
				stats, err := task.LoadStateCount(dbp)
				if err != nil {
					logrus.Warn(err)
				}
				for state, count := range stats {
					metrics.WithLabelValues(state).Set(count)
				}
			case <-ctx.Done():
				tick.Stop()
				return
			}
		}
	}()
}

// StatsOut aggregates different business stats:
// - a map of task states and their count
type StatsOut struct {
	TaskStates map[string]float64 `json:"task_states"`
}

// Stats handles the http request to fetch µtask statistics
// common to all instances
func Stats(c *gin.Context) (*StatsOut, error) {
	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return nil, err
	}

	out := StatsOut{}
	out.TaskStates, err = task.LoadStateCount(dbp)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

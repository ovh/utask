package engine

import (
	"context"
	"time"

	"github.com/juju/errors"
	"github.com/loopfz/gadgeto/zesty"
	"github.com/ovh/utask"
	"github.com/ovh/utask/db/pgjuju"
	"github.com/ovh/utask/models/resolution"
	"github.com/ovh/utask/models/runnerinstance"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"
)

// InstanceCollector launches a process that retrieves resolutions
// which might have been running on a dead instance and marks them as
// crashed, for examination
func InstanceCollector(ctx context.Context, maxConcurrentExecutions int, waitDuration time.Duration) error {
	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return err
	}

	var sm *semaphore.Weighted
	if maxConcurrentExecutions >= 0 {
		sm = semaphore.NewWeighted(int64(maxConcurrentExecutions))
	}

	go func() {
		// Start immediately
		collect(dbp, sm, waitDuration)

		for running := true; running; {
			// wake up every minute
			time.Sleep(time.Minute)

			select {
			case <-ctx.Done():
				running = false
			default:
				collect(dbp, sm, waitDuration)
			}
		}
	}()

	return nil
}

func collect(dbp zesty.DBProvider, sm *semaphore.Weighted, waitDuration time.Duration) error {
	// get a list of all instances
	instances, err := runnerinstance.ListInstances(dbp)
	if err != nil {
		return err
	}
	log := logrus.WithFields(logrus.Fields{"instance_id": utask.InstanceID, "collector": "instance_collector"})
	for _, i := range instances {
		// if an instance is dead
		if i.IsDead() {
			// loop while there are running resolutions from this instance
			for {
				r, err := getUpdateRunningResolution(dbp, i)
				if err != nil {
					// no more resolutions found, break out of loop
					if errors.IsNotFound(err) {
						break
					}
				} else {
					// run found resolution
					log.WithFields(logrus.Fields{"resolution_id": r.PublicID}).Debugf("collected crashed resolution %s", r.PublicID)
					_ = GetEngine().Resolve(r.PublicID, sm)

					// waiting between two resolve, so others instances can also select tasks
					time.Sleep(waitDuration)
				}
			}
			// no resolutions left to retry, delete instance
			if remaining, err := getRemainingResolution(dbp, i); err == nil && remaining == 0 {
				log.Infof("collected all resolution from %d, deleting instance from instance list", i.ID)
				i.Delete(dbp)
			}
		}
	}

	return nil
}

func getUpdateRunningResolution(dbp zesty.DBProvider, i *runnerinstance.Instance) (*resolution.Resolution, error) {
	sqlStmt := `UPDATE "resolution"
		SET instance_id = $1, state = $2
		WHERE id IN
		(
			SELECT id
			FROM "resolution"
			WHERE ((instance_id = $3 AND state IN ($2,$4,$5,$6)) OR
				   (instance_id = $1 AND state = $2))
     	 	AND pg_try_advisory_xact_lock(id)
			LIMIT 1
			FOR UPDATE
		)
		RETURNING id, public_id`

	var r resolution.Resolution

	if err := dbp.DB().SelectOne(&r, sqlStmt,
		utask.InstanceID,
		resolution.StateCrashed,
		i.ID,
		resolution.StateRunning,
		resolution.StateRetry,
		resolution.StateAutorunning,
	); err != nil {
		return nil, pgjuju.Interpret(err)
	}

	return &r, nil
}

func getRemainingResolution(dbp zesty.DBProvider, i *runnerinstance.Instance) (int64, error) {
	sqlStmt := `SELECT COUNT(id)
			FROM "resolution"
			WHERE instance_id = $1 AND state IN ($2,$3,$4,$5)`

	return dbp.DB().SelectInt(sqlStmt,
		i.ID,
		resolution.StateCrashed,
		resolution.StateRunning,
		resolution.StateRetry,
		resolution.StateAutorunning,
	)
}

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
)

// InstanceCollector launches a process that retrieves resolutions
// which might have been running on a dead instance and marks them as
// crashed, for examination
func InstanceCollector(ctx context.Context) error {
	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return err
	}

	go func() {
		// Start immediately
		collect(dbp)

		for running := true; running; {
			// wake up every minute
			time.Sleep(time.Minute)

			select {
			case <-ctx.Done():
				running = false
			default:
				collect(dbp)
			}
		}
	}()

	return nil
}

func collect(dbp zesty.DBProvider) error {
	// get a list of all instances
	instances, err := runnerinstance.ListInstances(dbp)
	if err != nil {
		return err
	}
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
					logrus.WithFields(logrus.Fields{"resolution_id": r.PublicID}).Debugf("Instance Collector: collected crashed resolution %s", r.PublicID)
					_ = GetEngine().Resolve(r.PublicID)
				}
			}
			// no resolutions left to retry, delete instance
			i.Delete(dbp)
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

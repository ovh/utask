package engine

import (
	"context"

	"github.com/loopfz/gadgeto/zesty"
	"github.com/sirupsen/logrus"

	"github.com/ovh/utask"
	"github.com/ovh/utask/db/pgjuju"
	"github.com/ovh/utask/models/resolution"
)

// RetryCollector launches a process that collects all resolutions
// eligible for a new run and passes them to the engine for execution
func RetryCollector(ctx context.Context) error {
	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return err
	}

	sl := newSleeper()

	go func() {
		for running := true; running; {
			sl.sleep()

			select {
			case <-ctx.Done():
				running = false
			default:
				r, _ := getUpdateErrorResolution(dbp)
				if r != nil {
					sl.wakeup()
					logrus.WithFields(logrus.Fields{"resolution_id": r.PublicID}).Debugf("Retry Collector: collected resolution %s", r.PublicID)
					_ = GetEngine().Resolve(r.PublicID, nil)
				}
			}
		}
	}()

	return nil
}

func getUpdateErrorResolution(dbp zesty.DBProvider) (*resolution.Resolution, error) {
	sqlStmt := `UPDATE "resolution"
		SET instance_id = $1, state = $2
		WHERE id IN
		(
			SELECT id
			FROM "resolution"
			WHERE ((instance_id = $1 AND state = $2) OR
				  ((state = $3 OR state = $4) AND next_retry < NOW()))
			AND pg_try_advisory_xact_lock(id)
			LIMIT 1
			FOR UPDATE
		)
		RETURNING id, public_id`

	var r resolution.Resolution

	instanceID := utask.InstanceID
	if err := dbp.DB().SelectOne(&r, sqlStmt, instanceID, resolution.StateRetry, resolution.StateError, resolution.StateToAutorunDelayed); err != nil {
		return nil, pgjuju.Interpret(err)
	}

	logrus.WithFields(logrus.Fields{
		"resolution_id": r.PublicID,
		"instance_id":   instanceID,
	}).Debugf("Retry Collector: set resolution %s with instanceID %d", r.PublicID, instanceID)
	return &r, nil
}

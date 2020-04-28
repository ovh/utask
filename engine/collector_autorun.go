package engine

import (
	"context"

	"github.com/loopfz/gadgeto/zesty"
	"github.com/sirupsen/logrus"

	"github.com/ovh/utask"
	"github.com/ovh/utask/db/pgjuju"
	"github.com/ovh/utask/models/resolution"
)

// AutorunCollector launches a process that looks for existing resolutions
// with state TO_AUTORUN, and passes them to the engine for execution
func AutorunCollector(ctx context.Context) error {
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
				r, _ := getUpdateAutorunResolution(dbp)
				if r != nil {
					sl.wakeup()
					logrus.WithFields(logrus.Fields{"resolution_id": r.PublicID}).Debugf("Autorun Collector: collected resolution %s", r.PublicID)
					_ = GetEngine().Resolve(r.PublicID, nil)
				}
			}
		}
	}()

	return nil
}

func getUpdateAutorunResolution(dbp zesty.DBProvider) (*resolution.Resolution, error) {
	sqlStmt := `UPDATE "resolution"
		SET instance_id = $1, state = $2
		WHERE id IN
		(
			SELECT id
			FROM "resolution"
			WHERE (state = $3 OR
				  (instance_id = $1 AND state = $2))
			AND pg_try_advisory_xact_lock(id)
			LIMIT 1
			FOR UPDATE
		)
		RETURNING id, public_id`

	var r resolution.Resolution

	instanceID := utask.InstanceID
	if err := dbp.DB().SelectOne(&r, sqlStmt, instanceID, resolution.StateAutorunning, resolution.StateToAutorun); err != nil {
		return nil, pgjuju.Interpret(err)
	}

	logrus.WithFields(logrus.Fields{
		"resolution_id": r.PublicID,
		"instance_id":   instanceID,
	}).Debugf("Autorun Collector: set resolution %s with instanceID %d", r.PublicID, instanceID)
	return &r, nil
}

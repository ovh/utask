package engine

import (
	"context"
	"time"

	"github.com/loopfz/gadgeto/zesty"
	"github.com/ovh/utask"
	"github.com/ovh/utask/db/pgjuju"
	"github.com/ovh/utask/models/task"
	"github.com/ovh/utask/pkg/now"
)

const thresholdStrDefault = "720h" // 1 month

// GarbageCollector launches a process that cleans up finished tasks
// (ie are in a final state) older than a given threshold
func GarbageCollector(ctx context.Context, completedTaskExpiration string) error {
	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return err
	}

	thresholdStr := completedTaskExpiration
	if thresholdStr == "" {
		thresholdStr = thresholdStrDefault // default fallback
	}
	threshold, err := time.ParseDuration(thresholdStr)
	if err != nil {
		return err
	}

	// sleep 24h: run once a day
	// delete old completed/blocked/cancelled/wontfix tasks
	go func() {
		// Run it immediately and wait for new tick
		deleteOldTasks(dbp, threshold)

		for running := true; running; {
			time.Sleep(24 * time.Hour)

			select {
			case <-ctx.Done():
				running = false
			default:
				deleteOldTasks(dbp, threshold)
			}
		}
	}()

	// delete un-referenced batches
	go func() {
		// Run it immediately and wait for new tick
		deleteOrphanBatches(dbp)

		for running := true; running; {
			time.Sleep(24 * time.Hour)

			select {
			case <-ctx.Done():
				running = false
			default:
				deleteOrphanBatches(dbp)
			}
		}
	}()

	return nil
}

// cascade delete task comments and task resolution
func deleteOldTasks(dbp zesty.DBProvider, perishedThreshold time.Duration) error {
	sqlStmt := `DELETE FROM "task"
		WHERE "task".state IN ($1,$2,$3)
		AND   "task".last_activity < $4`

	if _, err := dbp.DB().Exec(sqlStmt,
		// final task states, cannot run anymore
		task.StateDone,
		task.StateCancelled,
		task.StateWontfix,
		now.Get().Add(-perishedThreshold),
	); err != nil {
		return pgjuju.Interpret(err)
	}

	return nil
}

func deleteOrphanBatches(dbp zesty.DBProvider) error {
	sqlStmt := `DELETE FROM "batch"
		WHERE id IN (
			SELECT "batch".id
			FROM "batch"
			LEFT JOIN "task" ON "batch".id = "task".id_batch
			WHERE "task".id IS NULL
		)`

	if _, err := dbp.DB().Exec(sqlStmt); err != nil {
		return pgjuju.Interpret(err)
	}

	return nil
}

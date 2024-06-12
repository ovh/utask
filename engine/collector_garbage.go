package engine

import (
	"context"
	"log"
	"time"

	"github.com/loopfz/gadgeto/zesty"

	"github.com/ovh/utask"
	"github.com/ovh/utask/db/pgjuju"
	"github.com/ovh/utask/models/task"
	"github.com/ovh/utask/pkg/constants"
	"github.com/ovh/utask/pkg/now"
)

const (
	thresholdStrDefault  = "720h" // 1 month
	sleepDurationDefault = 24 * time.Hour
)

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

	sleepDuration := sleepDurationDefault
	if threshold < sleepDurationDefault {
		sleepDuration = threshold
	}

	// delete old completed/cancelled/wontfix tasks
	go func() {
		// Run it immediately and wait for new tick
		if err := deleteOldTasks(dbp, threshold); err != nil {
			log.Printf("GarbageCollector: failed to trash old tasks: %s", err)
		}

		for running := true; running; {
			time.Sleep(sleepDuration)

			select {
			case <-ctx.Done():
				running = false
			default:
				if err := deleteOldTasks(dbp, threshold); err != nil {
					log.Printf("GarbageCollector: failed to trash old tasks: %s", err)
				}
			}
		}
	}()

	// delete un-referenced batches
	go func() {
		// Run it immediately and wait for new tick
		if err := deleteOrphanBatches(dbp); err != nil {
			log.Printf("GarbageCollector: failed to trash old batches: %s", err)
		}

		for running := true; running; {
			time.Sleep(sleepDuration)

			select {
			case <-ctx.Done():
				running = false
			default:
				if err := deleteOrphanBatches(dbp); err != nil {
					log.Printf("GarbageCollector: failed to trash old batches: %s", err)
				}
			}
		}
	}()

	return nil
}

// cascade delete task comments and task resolution
// Subtasks are kept until the parent task is in a final state.
func deleteOldTasks(dbp zesty.DBProvider, perishedThreshold time.Duration) error {
	sqlStmt := `DELETE FROM "task"
		WHERE "task".state IN ($1,$2,$3)
		AND   "task".last_activity < $4
		AND   (NOT "task".tags ? $5 OR EXISTS (
		        SELECT 1 FROM "task" AS parentTask
		        WHERE 	parentTask.public_id = ("task".tags->>$5)::uuid
				AND 	parentTask.state IN ($1,$2,$3)
		    )
		)`

	if _, err := dbp.DB().Exec(sqlStmt,
		// final task states, cannot run anymore
		task.StateDone,
		task.StateCancelled,
		task.StateWontfix,
		now.Get().Add(-perishedThreshold),
		constants.SubtaskTagParentTaskID,
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

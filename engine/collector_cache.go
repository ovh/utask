package engine

import (
	"context"
	"log"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/loopfz/gadgeto/zesty"
	"github.com/ovh/utask"
	"github.com/ovh/utask/db/pgjuju"
	"github.com/ovh/utask/db/sqlgenerator"
	"github.com/ovh/utask/pkg/now"
)

// CacheCollector launches a process that cleans up expired entries from the cache plugin
func CacheCollector(ctx context.Context) error {
	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return err
	}

	// Using the same duration as the GarbageCollector
	sleepDuration := sleepDurationDefault

	// Delete expired entries from the cache plugin
	go func() {
		// Run it immediately and wait for new tick
		if purged, err := purgeExpiredEntries(dbp); err != nil {
			log.Printf("CacheCollector: failed to trash expired entries: %s", err)
		} else if purged > 0 {
			log.Printf("CacheCollector: purged %d expired entries at startup", purged)
		}

		for running := true; running; {
			time.Sleep(sleepDuration)

			select {
			case <-ctx.Done():
				running = false
			default:
				if purged, err := purgeExpiredEntries(dbp); err != nil {
					log.Printf("CacheCollector: failed to trash expired entries: %s", err)
				} else if purged > 0 {
					log.Printf("CacheCollector: purged %d expired entries", purged)
				}
			}
		}
	}()

	return nil
}

func purgeExpiredEntries(dbp zesty.DBProvider) (int64, error) {
	query, args, err := sqlgenerator.PGsql.
		Delete(`"cache"`).
		Where(squirrel.And{
			squirrel.NotEq{`"expires_at"`: nil},
			squirrel.Lt{`"expires_at"`: now.Get()},
		}).
		ToSql()
	if err != nil {
		return 0, err
	}

	res, err := dbp.DB().Exec(query, args...)
	if err != nil {
		return 0, pgjuju.Interpret(err)
	}

	return res.RowsAffected()
}

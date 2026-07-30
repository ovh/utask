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

// How many orphan entries we can fetch to purge at once
const maxOrphanEntriesToPurge uint64 = 100

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
		expired, orphan := purgeCache(dbp)
		if expired+orphan > 0 {
			log.Printf("CacheCollector: purged %d expired and %d orphan entrie(s) at startup", expired, orphan)
		}

		for running := true; running; {
			time.Sleep(sleepDuration)

			select {
			case <-ctx.Done():
				running = false
			default:
				expired, orphan := purgeCache(dbp)
				if expired+orphan > 0 {
					log.Printf("CacheCollector: purged %d expired and %d orphan entrie(s)", expired, orphan)
				}
			}
		}
	}()

	return nil
}

// purgeCache purges all expired and orphan entries in the "cache" table (See the "cache" plugin)
func purgeCache(dbp zesty.DBProvider) (int64, int64) {
	purgedExpired, err := purgeExpiredEntries(dbp)
	if err != nil {
		log.Printf("CacheCollector: failed to trash expired entries: %s", err)
	}

	purgedOrdphan, err := purgeOrphanEntries(dbp)
	if err != nil {
		log.Printf("CacheCollector: failed to trash orphan entries: %s", err)
	}

	return purgedExpired, purgedOrdphan
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

func purgeOrphanEntries(dbp zesty.DBProvider) (int64, error) {
	// Fetch orphan entries
	query, args, err := sqlgenerator.PGsql.
		Select(`"cache"."key"`).
		From(`"cache"`).
		LeftJoin(`"task" on "cache"."owner" = "task"."public_id"`).
		Where(squirrel.And{
			squirrel.Eq{`"cache"."expires_at"`: nil},
			squirrel.NotEq{`"cache"."owner"`: nil},
			squirrel.Eq{`"task"."public_id"`: nil},
		}).
		Limit(maxOrphanEntriesToPurge).
		ToSql()
	if err != nil {
		return 0, err
	}

	cacheKeys := make([]string, 0)
	_, err = dbp.DB().Select(&cacheKeys, query, args...)
	if err != nil {
		return 0, pgjuju.Interpret(err)
	}

	query, args, err = sqlgenerator.PGsql.
		Delete(`"cache"`).
		Where(squirrel.Eq{`"key"`: cacheKeys}).
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

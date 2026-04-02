package plugincache

import (
	"database/sql"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/juju/errors"
	"github.com/loopfz/gadgeto/zesty"

	"github.com/ovh/utask/db/pgjuju"
	"github.com/ovh/utask/db/sqlgenerator"
	"github.com/ovh/utask/pkg/now"
)

type cacheEntry struct {
	Key       string     `db:"key"`
	Value     []byte     `db:"value"`
	ExpiresAt *time.Time `db:"expires_at"`
}

func (c *cacheEntry) isExpired() bool {
	return c.ExpiresAt != nil && now.Get().After(*c.ExpiresAt)
}

func setCacheEntry(dbp zesty.DBProvider, key string, value []byte, ttl int64) error {
	var expiresAt *time.Time
	if ttl > 0 {
		t := now.Get().Add(time.Duration(ttl) * time.Second)
		expiresAt = &t
	}

	query, args, err := sqlgenerator.PGsql.
		Insert(`"cache"`).
		Columns(`"key"`, `"value"`, `"expires_at"`).
		Values(key, value, expiresAt).
		Suffix(`ON CONFLICT ("key") DO UPDATE SET "value" = EXCLUDED."value", "expires_at" = EXCLUDED."expires_at"`).
		ToSql()
	if err != nil {
		return err
	}

	_, err = dbp.DB().Exec(query, args...)
	if err != nil {
		return pgjuju.Interpret(err)
	}

	return nil
}

func getCacheEntry(dbp zesty.DBProvider, key string) (*cacheEntry, error) {
	query, args, err := sqlgenerator.PGsql.
		Select(`"key"`, `"value"`, `"expires_at"`).
		From(`"cache"`).
		Where(squirrel.Eq{`"key"`: key}).
		ToSql()
	if err != nil {
		return nil, err
	}

	entry := &cacheEntry{}
	err = dbp.DB().SelectOne(entry, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFoundf("cache key %q", key)
		}
		return nil, pgjuju.Interpret(err)
	}

	if entry.isExpired() {
		_ = deleteCacheEntry(dbp, key)
		return nil, errors.NotFoundf("cache key %q", key)
	}

	return entry, nil
}

func deleteCacheEntry(dbp zesty.DBProvider, key string) error {
	query, args, err := sqlgenerator.PGsql.
		Delete(`"cache"`).
		Where(squirrel.Eq{`"key"`: key}).
		ToSql()
	if err != nil {
		return err
	}

	_, err = dbp.DB().Exec(query, args...)
	if err != nil {
		return pgjuju.Interpret(err)
	}

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

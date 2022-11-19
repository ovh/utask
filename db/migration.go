package db

import (
	"database/sql"
	"fmt"

	"github.com/Masterminds/squirrel"
	_ "github.com/lib/pq" // postgresql driver
	"github.com/loopfz/gadgeto/zesty"

	"github.com/ovh/utask"
	"github.com/ovh/utask/db/sqlgenerator"
)

const (
	expectedVersion = "v1.21.1-migration010"
)

var (
	baseQuery = sqlgenerator.PGsql.Select(
		`"utask_sql_migrations".current_migration_applied`,
	).From(
		`"utask_sql_migrations"`,
	)
)

// migrationChecker make sure that the latest SQL migration is correctly applied
// otherwise, fails to start µTask.
// SQL migrations are supposed to be added each time the database schema evolves. Previous migrations
// are not supposed to be removed from the database, unless the SQL migration is a breaking change.
// Keeping previous migration will avoid breaking the current running instance to break if the SQL migration
// is applied, while µTask haven't been updated yet.
// In case of a SQL breaking change, SQL migration should consider removing all previous SQL migration from the table
// before adding the latest SQL migration version, to prevent old version to start with incompatible SQL schema.
func migrationChecker() error {
	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return err
	}

	sel := baseQuery.Where(
		squirrel.Eq{`"utask_sql_migrations".current_migration_applied`: expectedVersion},
	).Limit(1)

	query, params, err := sel.ToSql()
	if err != nil {
		return err
	}

	row := dbp.DB().QueryRow(query, params...)
	if err := row.Err(); err != nil {
		return fmt.Errorf("unable to start µTask: can't fetch latest SQL migration: %s", err)
	}

	var version string
	if err := row.Scan(&version); err != nil && err == sql.ErrNoRows {
		lastVersionApplied, _ := lastKnownVersion(dbp)
		if lastVersionApplied == "" {
			return fmt.Errorf("unable to start µTask: migration table is empty, can't determine which SQL migration were already applied")
		}

		return fmt.Errorf("unable to start µTask: SQL database doesn't contains the latest SQL migration (%s). Last SQL migration applied according to the database: %s. Please apply all SQL migrations before starting the latest µTask version", expectedVersion, lastVersionApplied)
	}

	if version != expectedVersion {
		return fmt.Errorf("unable to start µTask: missing latest SQL migration: expected %q, found %q", expectedVersion, version)
	}

	return nil
}

func lastKnownVersion(dbp zesty.DBProvider) (string, error) {
	sel := baseQuery.OrderBy("current_migration_applied DESC").Limit(1)

	query, params, err := sel.ToSql()
	if err != nil {
		return "", err
	}

	row := dbp.DB().QueryRow(query, params...)
	if err := row.Err(); err != nil {
		return "", fmt.Errorf("unable to start µTask: missing latest SQL migration: %s", err)
	}

	var version string
	if err := row.Scan(&version); err != nil && err == sql.ErrNoRows {
		return "", nil
	}

	return version, nil
}

package db

import (
	"fmt"

	_ "github.com/lib/pq" // postgresql driver
	"github.com/loopfz/gadgeto/zesty"

	"github.com/ovh/utask"
	"github.com/ovh/utask/db/sqlgenerator"
)

const (
	expectedVersion = "v1.10.0-migration005"
)

// migrationChecker make sure that the latest SQL migration is correctly applied
// otherwise, fails to start µTask.
func migrationChecker() error {
	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return err
	}

	sel := sqlgenerator.PGsql.Select(
		`"utask_sql_migrations".current_migration_applied`,
	).From(
		`"utask_sql_migrations"`,
	).Limit(1)

	query, params, err := sel.ToSql()
	if err != nil {
		return err
	}

	row := dbp.DB().QueryRow(query, params...)
	if err := row.Err(); err != nil {
		return fmt.Errorf("unable to start µTask: missing latest SQL migration: %s", err)
	}

	var version string
	if err := row.Scan(&version); err != nil {
		return fmt.Errorf("unable to start µTask: missing latest SQL migration: %s", err)
	}
	if version != expectedVersion {
		return fmt.Errorf("unable to start µTask: missing latest SQL migration: expected %q, found %q", expectedVersion, version)
	}

	return nil
}

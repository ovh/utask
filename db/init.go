package db

import (
	"database/sql"
	"time"

	"github.com/go-gorp/gorp"
	_ "github.com/lib/pq" // postgresql driver
	"github.com/loopfz/gadgeto/zesty"
	"github.com/ovh/configstore"
	"github.com/sirupsen/logrus"

	"github.com/ovh/utask"
	"github.com/ovh/utask/models"
	"github.com/ovh/utask/models/hook"
	"github.com/ovh/utask/models/resolution"
	"github.com/ovh/utask/models/runnerinstance"
	"github.com/ovh/utask/models/task"
	"github.com/ovh/utask/models/tasktemplate"
	"github.com/ovh/utask/pkg/now"
)

const (
	databaseCfgKey = "database"

	defaultMaxOpenConns    = 50
	defaultMaxIdleConns    = 30
	defaultConnMaxLifetime = 60 // In seconds
)

type tableModel struct {
	Model   interface{}
	Name    string
	Keys    []string
	Autoinc bool
}

var schema = []tableModel{
	{tasktemplate.TaskTemplate{}, "task_template", []string{"id"}, true},
	{task.DBModel{}, "task", []string{"id"}, true},
	{task.Comment{}, "task_comment", []string{"id"}, true},
	{task.BatchDBModel{}, "batch", []string{"id"}, true},
	{resolution.DBModel{}, "resolution", []string{"id"}, true},
	{runnerinstance.Instance{}, "runner_instance", []string{"id"}, true},
	{hook.Hook{}, "hook", []string{"id"}, true},
}

// Init takes a connection string and a configuration struct
// and registers a new postgres DB connection in zesty,
// under utask.DBName -> accessible from api handlers and engine collectors
func Init(store *configstore.Store) error {
	config, err := utask.Config(store)
	if err != nil {
		return err
	}
	cfg := config.DatabaseConfig
	dbConn, err := configstore.Filter().Slice(databaseCfgKey).Squash().Store(store).MustGetFirstItem().Value()
	if err != nil {
		return err
	}
	db, err := sql.Open("postgres", dbConn)
	if err != nil {
		return err
	}

	if cfg == nil {
		cfg = &utask.DatabaseConfig{
			MaxOpenConns:    defaultMaxOpenConns,
			MaxIdleConns:    defaultMaxIdleConns,
			ConnMaxLifetime: defaultConnMaxLifetime,
		}
	} else {
		cfg.MaxOpenConns = normalize(cfg.MaxOpenConns, defaultMaxOpenConns)
		cfg.MaxIdleConns = normalize(cfg.MaxIdleConns, defaultMaxIdleConns)
		cfg.ConnMaxLifetime = normalize(cfg.ConnMaxLifetime, defaultConnMaxLifetime)
	}

	logrus.Infof("[DatabaseConfig] Using %d max open connections, %d max idle connections, %d seconds timeout", cfg.MaxOpenConns, cfg.MaxIdleConns, cfg.ConnMaxLifetime)
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	dbmap, err := getDbMap(db, schema, typeConverter{})
	if err != nil {
		return err
	}
	if err := zesty.RegisterDB(
		zesty.NewDB(dbmap),
		utask.DBName,
	); err != nil {
		return err
	}
	if err := now.Init(); err != nil {
		return err
	}
	return models.Init(store)
}

func getDbMap(db *sql.DB, schema []tableModel, tc gorp.TypeConverter) (*gorp.DbMap, error) {
	dbmap := &gorp.DbMap{
		Db:            db,
		Dialect:       gorp.PostgresDialect{},
		TypeConverter: tc,
	}

	for _, m := range schema {
		dbmap.AddTableWithName(m.Model, m.Name).SetKeys(m.Autoinc, m.Keys...)
	}

	return dbmap, nil
}

func normalize(current, fallback int) int {
	if current < 0 {
		return fallback
	}
	return current
}

package now

import (
	"time"

	"github.com/loopfz/gadgeto/zesty"
	"github.com/ovh/utask"
)

var timeDelta time.Duration

// Init establishes a synchronization mechanism for all utask instances
// by comparing time.Now() to the database's NOW value
func Init() error {
	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return err
	}

	dbNow, err := getDbTimeNow(dbp)
	if err != nil {
		return err
	}

	timeDelta = dbNow.Sub(time.Now())

	return nil
}

// Get returns the synchronized Now() value
// equal among all utask instances connected to the same DB
// this is meant to replace time.Now() in all packages of utask
func Get() time.Time {
	return time.Now().Add(timeDelta)
}

func getDbTimeNow(dbp zesty.DBProvider) (*time.Time, error) {
	now := struct {
		Now *time.Time `db:"now"`
	}{}

	if err := dbp.DB().SelectOne(&now, `SELECT NOW()`); err != nil {
		return nil, err
	}
	return now.Now, nil
}

package plugincache

import (
	"github.com/loopfz/gadgeto/zesty"
	"github.com/ovh/utask"
	"github.com/ovh/utask/pkg/plugins"
	"github.com/sirupsen/logrus"
)

var (
	Init = &CacheInit{}
)

// CacheInit handles the plugin initialization: table registration
// and periodic cleanup of expired entries.
type CacheInit struct{}

func (ci *CacheInit) Init(s *plugins.Service) error {
	// No table model registration needed: we use raw SQL queries
	// with squirrel, not gorp ORM mapping.

	// Purge expired entries at startup
	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		logrus.Warnf("cache plugin: unable to get DB provider for initial purge: %s", err)
		return nil
	}

	purged, err := purgeExpiredEntries(dbp)
	if err != nil {
		logrus.Warnf("cache plugin: initial purge failed: %s", err)
	} else if purged > 0 {
		logrus.Infof("cache plugin: purged %d expired entries at startup", purged)
	}

	return nil
}

func (ci *CacheInit) Description() string {
	return "Cache plugin: key-value store with optional TTL expiration"
}

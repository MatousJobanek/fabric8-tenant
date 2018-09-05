package dbsupport

import (
	"github.com/jinzhu/gorm"
	"time"
	"github.com/fabric8-services/fabric8-tenant/configuration"
	"github.com/fabric8-services/fabric8-tenant/log"
	"github.com/fabric8-services/fabric8-tenant/dbsupport/migration"
	"github.com/sirupsen/logrus"
)

func Connect(config *configuration.Data, log log.Logger) *gorm.DB {
	var err error
	var db *gorm.DB
	for {
		db, err = gorm.Open("postgres", config.GetPostgresConfigString())
		if err != nil {
			log.Errorf("ERROR: Unable to open connection to database %v", err)
			log.Infof("Retrying to connect in %v...", config.GetPostgresConnectionRetrySleep())
			time.Sleep(config.GetPostgresConnectionRetrySleep())
		} else {
			break
		}
	}

	if config.IsDeveloperModeEnabled() {
		db = db.Debug()
	}

	if config.GetPostgresConnectionMaxIdle() > 0 {
		log.Infof("Configured connection pool max idle %v", config.GetPostgresConnectionMaxIdle())
		db.DB().SetMaxIdleConns(config.GetPostgresConnectionMaxIdle())
	}
	if config.GetPostgresConnectionMaxOpen() > 0 {
		log.Infof("Configured connection pool max open %v", config.GetPostgresConnectionMaxOpen())
		db.DB().SetMaxOpenConns(config.GetPostgresConnectionMaxOpen())
	}
	return db
}

func Migrate(db *gorm.DB,config *configuration.Data, log *logrus.Entry) {
	// Migrate the schema
	err := migration.Migrate(db.DB(), config.GetPostgresDatabase())
	if err != nil {
		log.Panic(nil, map[string]interface{}{
			"err": err,
		}, "failed migration")
	}
}
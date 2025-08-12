package database

import (
	"fmt"
	"sync"
	"time"

	"authz/internal/cfg"
	"authz/internal/engine/rbac"
	"authz/internal/pkg"

	"github.com/rs/zerolog/log"
	"github.com/skyrocket-qy/erx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var initOnce sync.Once

type zerologWriter struct{}

func (z *zerologWriter) Printf(format string, v ...any) {
	log.Info().Msgf(format, v...)
}

func New(lc pkg.Lifecycle) (db *gorm.DB, err error) {
	initOnce.Do(func() {
		log.Info().Msg("New db")

		config := gorm.Config{
			NamingStrategy: schema.NamingStrategy{
				// NoLowerCase: true,
			},
			Logger: logger.New(
				&zerologWriter{},
				logger.Config{
					SlowThreshold:             time.Second,
					LogLevel:                  logger.Warn,
					IgnoreRecordNotFoundError: false,
					ParameterizedQueries:      true,
					Colorful:                  true,
				},
			),
		}

		log.Info().Msg("Connecting to Postgres")
		dbCfg := cfg.Cfg.Db
		connStr := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s TimeZone=%s",
			dbCfg.Host,
			dbCfg.Port,
			dbCfg.User,
			dbCfg.Password,
			dbCfg.Db,
			"Asia/Taipei",
		)
		db, err = gorm.Open(postgres.Open(connStr), &config)

		if err != nil {
			err = erx.W(err).SetCode(pkg.ErrDBUnavailable)

			return
		}
	})

	lc.Add(func() error {
		if db == nil {
			return nil
		}

		sqlDB, err := db.DB()
		if err != nil {
			return err
		}

		return sqlDB.Close()
	})

	db.AutoMigrate(&rbac.Tuple{})
	// db.AutoMigrate(
	// 	&model.Org{},
	// 	&model.User{},
	// 	&model.UserAuth{},
	// 	&model.Resource{},
	// 	&model.Role{},
	// 	&model.Tuple{},
	// 	&model.ChangeLog{},
	// 	&model.GraphCheckpoint{},
	// )

	return db, err
}

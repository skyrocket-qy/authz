package database

import (
	"context"
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

func New(lc *pkg.LifecycleParallel) (db *gorm.DB, err error) {
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
			"UTC",
		)

		db, err = gorm.Open(postgres.Open(connStr), &config)
		if err != nil {
			err = erx.W(err).SetCode(pkg.ErrDBUnavailable)

			return
		}
	})

	lc.Add(db, func(c context.Context) error {
		if db == nil {
			return nil
		}

		sqlDB, err := db.DB()
		if err != nil {
			return err
		}

		return sqlDB.Close()
	})

	if err := db.AutoMigrate(&rbac.Tuple{}, &rbac.GraphCheckpoint{}); err != nil {
		return nil, err
	}

	return db, err
}

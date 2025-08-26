package database

import (
	"context"
	"fmt"
	"time"

	"authz/internal/config"
	"authz/internal/engine/rbac"
	"authz/internal/util"

	"github.com/rs/zerolog/log"
	"github.com/skyrocket-qy/erx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type zerologWriter struct{}

func (z *zerologWriter) Printf(format string, v ...any) {
	log.Info().Msgf(format, v...)
}

func New(lc *util.LifecycleParallel) (db *gorm.DB, err error) {
	log.Info().Msg("New db")

	gormConf := gorm.Config{
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

	dbCfg := config.Conf.Db
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s TimeZone=%s",
		dbCfg.Host,
		dbCfg.Port,
		dbCfg.User,
		dbCfg.Password,
		dbCfg.Db,
		"UTC",
	)

	db, err = gorm.Open(postgres.Open(connStr), &gormConf)
	if err != nil {
		err = erx.W(err).SetCode(util.ErrDBUnavailable)

		return db, err
	}

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

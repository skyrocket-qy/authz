package database

import (
	"context"
	"fmt"
	"time"

	"authz/internal/config"
	"authz/internal/entity/model"
	"authz/internal/util"
	"github.com/rs/zerolog/log"
	"github.com/skyrocket-qy/erx"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
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

	var dialector gorm.Dialector

	switch dbCfg.Driver {
	case "postgres":
		connStr := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s TimeZone=%s",
			dbCfg.Host,
			dbCfg.Port,
			dbCfg.User,
			dbCfg.Password,
			dbCfg.Db,
			"UTC",
		)
		log.Print(connStr)
		dialector = postgres.Open(connStr)
	case "sqlite":
		dialector = sqlite.Open(dbCfg.Host)
	default:
		return nil, erx.Newf(util.ErrBadRequest, "unsupported db driver: %s", dbCfg.Driver)
	}

	db, err = gorm.Open(dialector, &gormConf)
	if err != nil {
		return db, erx.W(err).SetCode(util.ErrDBUnavailable)
	}

	lc.Add(db, func(c context.Context) error {
		if db == nil {
			return nil
		}

		sqlDB, err := db.DB()
		if err != nil {
			return erx.W(err)
		}

		return sqlDB.Close()
	})

	if err := db.AutoMigrate(&model.Tuple{}, &model.GraphCheckpoint{}); err != nil {
		return nil, erx.W(err)
	}

	return db, nil
}

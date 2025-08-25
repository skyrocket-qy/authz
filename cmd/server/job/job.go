package job

import (
	"context"

	"authz/internal/engine/rbac"
	"authz/internal/pkg"
	"authz/internal/service"
	"authz/internal/service/database"
	"authz/internal/service/logx"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "job",
	Short: "run single job service",
	// Long:  `The longer description`,
	Run: start,
}

func start(cmd *cobra.Command, args []string) {
	if err := pkg.NewConfig(); err != nil {
		log.Err(err).Msg("Failed to load config")

		return
	}

	if err := logx.InitLogger(); err != nil {
		log.Err(err).Msg("Failed to init logger")

		return
	}

	lc := pkg.NewLifecycleParallel()

	db, err := database.New(lc)
	if err != nil {
		log.Err(err).Msg("Failed to init db")

		return
	}

	kafkaR := service.NewKafkaReader(lc)

	zm, err := rbac.NewZanzibarMemory(context.TODO(), lc, db, nil, kafkaR)
	if err != nil {
		log.Err(err).Msg("Failed to init rbac engine")

		return
	}

	zm.SyncGraphCheckpoint(context.TODO())
}

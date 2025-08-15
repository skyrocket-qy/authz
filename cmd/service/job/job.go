package job

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "job",
	Short: "run single job service",
	// Long:  `The longer description`,
	Run: start,
}

func start(cmd *cobra.Command, args []string) {
	// if err := pkg.NewConfig(); err != nil {
	// 	logx.Error(err.Error())

	// 	return
	// }

	// logger.InitLogger()
	// lc := pkg.NewLifecycleParallel()

	// db, err := database.New(lc)
	// if err != nil {
	// 	logx.Error(err.Error())
	// 	return
	// }
	// kafkaR := service.NewKafkaReader(lc)

	// zm, err := rbac.NewZanzibarMemory(context.TODO(), lc, db, nil, kafkaR)
	// if err != nil {
	// 	logx.Error(err.Error())
	// 	return
	// }

	// zm.SyncGraphCheckpoint(context.TODO())
}

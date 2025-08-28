//go:build wireinject
// +build wireinject

package wire

import (
	"authz/internal/engine/rbac"
	"authz/internal/schema"
	"authz/internal/service"
	"authz/internal/service/redis"
	"authz/internal/util"
	"authz/internal/zanzibar"
	"context"

	"github.com/google/wire"
	"gorm.io/gorm"
)

func NewRbacHandler(context.Context, *util.LifecycleParallel, *gorm.DB) (*rbac.Handler, error) {
	wire.Build(
		redis.New,
		schema.NewSchema,

		service.NewKafkaReader,

		zanzibar.NewZanzibarMemory,
		wire.Bind(new(zanzibar.ZanzibarMemory), new(*zanzibar.ZanzibarMemoryImpl)),
		zanzibar.NewZanzibarLogic,
		wire.Bind(new(zanzibar.ZanzibarLogic), new(*zanzibar.ZanzibarLogicImpl)),
		rbac.NewRbacLogic,
		wire.Bind(new(rbac.RbacLogic), new(*rbac.RbacLogicImpl)),
		rbac.NewHandler,
	)
	return nil, nil
}

// func NewJobService(context.Context, *util.LifecycleParallel) (*service.JobService, error) {
// 	wire.Build(
// 		service.NewJobService,
// 	)
// 	return nil, nil
// }

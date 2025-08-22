//go:build wireinject
// +build wireinject

package wire

import (
	"authz/internal/engine/rbac"
	"authz/internal/pkg"
	"authz/internal/schema"
	"authz/internal/service"
	"authz/internal/service/redis"
	"context"

	"github.com/google/wire"
	"gorm.io/gorm"
)

func NewRbacHandler(context.Context, *pkg.LifecycleParallel, *gorm.DB) (*rbac.Handler, error) {
	wire.Build(
		redis.New,
		schema.NewSchema,

		service.NewKafkaReader,

		rbac.NewZanzibarMemory,
		wire.Bind(new(rbac.ZanzibarMemory), new(*rbac.ZanzibarMemoryImpl)),
		rbac.NewZanzibarLogic,
		wire.Bind(new(rbac.ZanzibarLogic), new(*rbac.ZanzibarLogicImpl)),
		rbac.NewRbacLogic,
		wire.Bind(new(rbac.RbacLogic), new(*rbac.RbacLogicImpl)),
		rbac.NewHandler,
	)
	return nil, nil
}

// func NewJobService(context.Context, *pkg.LifecycleParallel) (*service.JobService, error) {
// 	wire.Build(
// 		service.NewJobService,
// 	)
// 	return nil, nil
// }

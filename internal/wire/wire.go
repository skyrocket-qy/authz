//go:build wireinject
// +build wireinject

package wire

import (
	"authz/internal/engine"
	"authz/internal/handler/connect"
	"authz/internal/logic"
	"authz/internal/pkg"
	"authz/internal/service/database"
	"authz/internal/service/redis"
	"context"

	"github.com/google/wire"
)

// func NewRestHandler(pkg.Lifecycle) (*rest.Handler, error) {
// 	wire.Build(
// 		database.New,
// 		redis.New,
// 		logic.NewZanzibarLogic,
// 		wire.Bind(new(logic.ZanzibarLogic), new(*logic.ZanzibarLogicImpl)),
// 		logic.NewRbacLogic,
// 		wire.Bind(new(logic.RbacLogic), new(*logic.RbacLogicImpl)),
// 		rest.NewHandler,
// 	)
// 	return nil, nil
// }

func NewConnectHandler(context.Context, pkg.Lifecycle) (*connect.Handler, error) {
	wire.Build(
		database.New,
		redis.New,

		engine.NewZanzibarEngine,
		wire.Bind(new(engine.ZanzibarEngine), new(*engine.ZanzibarEngineImpl)),
		logic.NewZanzibarLogic,
		wire.Bind(new(logic.ZanzibarLogic), new(*logic.ZanzibarLogicImpl)),
		logic.NewRbacLogic,
		wire.Bind(new(logic.RbacLogic), new(*logic.RbacLogicImpl)),
		connect.NewHandler,
	)
	return nil, nil
}

//go:build wireinject
// +build wireinject

package wire

import (
	"authz/internal/handler/rest"
	"authz/internal/logic"
	"authz/internal/pkg"
	"authz/internal/service/database"
	"authz/internal/service/redis"

	"github.com/google/wire"
)

func NewHandler(pkg.Lifecycle) (*rest.Handler, error) {
	wire.Build(
		database.New,
		redis.New,
		logic.NewZanzibarLogic,
		wire.Bind(new(logic.ZanzibarLogic), new(*logic.ZanzibarLogicImpl)),
		logic.NewRbacLogic,
		wire.Bind(new(logic.RbacLogic), new(*logic.RbacLogicImpl)),
		rest.NewHandler,
	)
	return nil, nil
}

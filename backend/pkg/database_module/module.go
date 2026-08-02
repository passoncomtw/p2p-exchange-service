package databasemodule

import (
	postgresql "p2p-exchange/pkg/database_module/postgresql"
	redis "p2p-exchange/pkg/database_module/redis"

	"go.uber.org/fx"
)

var Module = fx.Module("databasemodule",
	postgresql.Module,
	redis.Module,
)

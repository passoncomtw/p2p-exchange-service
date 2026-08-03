package repository

import (
	userrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/user"

	"go.uber.org/fx"
)

var Module = fx.Module("repository",
	fx.Provide(userrepo.New),
)

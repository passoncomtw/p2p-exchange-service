package main

import (
	"flag"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/config"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/server"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/service"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/wsserver"
	databaseModule "p2p-exchange/pkg/database_module"
	"p2p-exchange/pkg/mq"
	notificationModule "p2p-exchange/pkg/notification_module"
	"p2p-exchange/pkg/schedule"

	"go.uber.org/fx"
)

func main() {
	flag.Parse()

	fx.New(
		config.Module,
		databaseModule.Module,
		mq.Module,
		repository.Module,
		service.Module,
		server.Module,
		wsserver.Module,
		notificationModule.Module,
		schedule.Module,
	).Run()
}

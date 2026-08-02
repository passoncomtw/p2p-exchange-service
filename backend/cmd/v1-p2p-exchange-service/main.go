package main

import (
	"flag"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/config"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/server"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/wsserver"
	notificationModule "p2p-exchange/pkg/notification_module"

	"go.uber.org/fx"
)

func main() {
	flag.Parse()

	fx.New(
		config.Module,
		server.Module,
		wsserver.Module,
		notificationModule.Module,
	).Run()
}

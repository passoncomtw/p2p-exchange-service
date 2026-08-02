package main

import (
	"flag"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/config"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/server"

	"go.uber.org/fx"
)

func main() {
	flag.Parse()

	fx.New(
		config.Module,
		server.Module,
	).Run()
}

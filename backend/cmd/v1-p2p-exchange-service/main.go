package main

import (
	"flag"
	"log"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/config"

	"github.com/joho/godotenv"
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

// @title						P2P Exchange API v1
// @version					1.0.0
// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						Authorization
// @description				輸入 JWT 本身（不含 Bearer 前綴）
func main() {
	flag.Parse()
	if err := godotenv.Load(); err != nil {
		log.Fatalf("error: .env file not found: %v", err)
	}

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

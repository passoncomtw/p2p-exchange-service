package server

import (
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/server/handlers"

	"github.com/zeromicro/go-zero/rest"
	"go.uber.org/fx"
)

var Module = fx.Module("server",
	fx.Provide(handlers.NewLoginHandler),
	fx.Provide(handlers.NewProfileHandler),
	fx.Provide(handlers.NewPaymentMethodHandler),
	fx.Provide(handlers.NewListingHandler),
	fx.Provide(handlers.NewOrderHandler),
	fx.Provide(handlers.NewBackendHandler),
	fx.Provide(handlers.NewV1Handler),
	fx.Provide(NewServer),
	fx.Invoke(func(*rest.Server) {}),
)

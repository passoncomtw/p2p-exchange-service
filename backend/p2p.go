package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"

	apierrors "p2p-exchange/internal/errors"
	"p2p-exchange/internal/config"
	"p2p-exchange/internal/handler"
	"p2p-exchange/internal/job"
	"p2p-exchange/internal/response"
	"p2p-exchange/internal/swagger"
	"p2p-exchange/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

var configFile = flag.String("f", "etc/config.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	httpx.SetOkHandler(func(_ context.Context, v any) any {
		return response.Success(v)
	})

	httpx.SetErrorHandler(func(err error) (int, any) {
		if e, ok := err.(*apierrors.AppError); ok {
			return e.Code, response.Fail(e.Code, e.Message)
		}
		return http.StatusBadRequest, response.Fail(http.StatusBadRequest, err.Error())
	})

	server := rest.MustNewServer(c.RestConf,
		rest.WithCors("*"),
		rest.WithUnauthorizedCallback(func(w http.ResponseWriter, r *http.Request, _ error) {
			httpx.WriteJsonCtx(r.Context(), w, http.StatusUnauthorized,
				response.Fail(http.StatusUnauthorized, apierrors.ErrUnauthorized.Message))
		}),
	)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	rootCtx := context.Background()
	job.StartScheduler(rootCtx, ctx.MQ)
	job.StartExpiredOrderConsumer(ctx.MQ, job.ExpiredOrderDeps{
		RDB:       ctx.RDB,
		Order:     ctx.Order,
		Wallet:    ctx.Wallet,
		Listing:   ctx.Listing,
		StatusLog: ctx.OrderStatusLog,
		DB:        ctx.DB,
	})
	job.StartPushNotificationConsumer(ctx.MQ, job.PushNotificationDeps{
		AppUser: ctx.AppUser,
	})

	if c.Mode != "pro" {
		swagger.RegisterRoutes(server)
		fmt.Printf("Swagger UI: http://localhost:%d/swagger\n", c.Port)
	}

	fmt.Printf("Starting server at %s:%d (mode=%s)...\n", c.Host, c.Port, c.Mode)
	server.Start()
}

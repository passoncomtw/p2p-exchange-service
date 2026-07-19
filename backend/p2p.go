package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	apierrors "p2p-exchange/internal/errors"
	"p2p-exchange/internal/config"
	"p2p-exchange/internal/handler"
	"p2p-exchange/internal/job"
	"p2p-exchange/internal/response"
	"p2p-exchange/internal/swagger"
	"p2p-exchange/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
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

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	server.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/healthz",
		Handler: healthzHandler(ctx),
	})

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
	job.StartTronScannerJob(rootCtx, c.Tron, job.TronScannerDeps{
		DB:            ctx.DB,
		RDB:           ctx.RDB,
		CryptoDeposit: ctx.CryptoDeposit,
		AppUser:       ctx.AppUser,
		Wallet:        ctx.Wallet,
	})
	job.StartTronWithdrawJob(rootCtx, c.Tron, job.TronWithdrawDeps{
		RDB:            ctx.RDB,
		CryptoWithdraw: ctx.CryptoWithdraw,
		Wallet:         ctx.Wallet,
	})

	if c.Mode != "pro" {
		swagger.RegisterRoutes(server)
		fmt.Printf("Swagger UI: http://localhost:%d/swagger\n", c.Port)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		fmt.Printf("Starting server at %s:%d (mode=%s)...\n", c.Host, c.Port, c.Mode)
		server.Start()
	}()

	<-quit
	logx.Info("shutting down...")

	server.Stop()

	if ctx.MQ != nil {
		ctx.MQ.Close()
		logx.Info("nats drained")
	}
	if ctx.RDB != nil {
		_ = ctx.RDB.Close()
		logx.Info("redis closed")
	}

	logx.Info("shutdown complete")
}

func healthzHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		checks := map[string]string{}
		httpStatus := http.StatusOK

		if ctx.RDB != nil {
			if err := ctx.RDB.Ping(r.Context()); err != nil {
				checks["redis"] = "unhealthy: " + err.Error()
				httpStatus = http.StatusServiceUnavailable
			} else {
				checks["redis"] = "ok"
			}
		}

		if ctx.MQ != nil {
			if err := ctx.MQ.Ping(); err != nil {
				checks["nats"] = "unhealthy: " + err.Error()
				httpStatus = http.StatusServiceUnavailable
			} else {
				checks["nats"] = "ok"
			}
		}

		var one int
		if err := ctx.DB.QueryRowCtx(r.Context(), &one, "select 1"); err != nil {
			checks["database"] = "unhealthy: " + err.Error()
			httpStatus = http.StatusServiceUnavailable
		} else {
			checks["database"] = "ok"
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(httpStatus)
		_ = json.NewEncoder(w).Encode(checks)
	}
}

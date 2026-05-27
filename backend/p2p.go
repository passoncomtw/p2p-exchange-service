// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"

	apierrors "p2p-exchange/internal/errors"
	"p2p-exchange/internal/config"
	"p2p-exchange/internal/handler"
	"p2p-exchange/internal/response"
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

	server.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/swagger",
		Handler: swaggerUIHandler(),
	})
	server.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/swagger/doc.yaml",
		Handler: swaggerDocHandler(),
	})

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	fmt.Printf("Swagger UI: http://%s:%d/swagger\n", c.Host, c.Port)
	server.Start()
}

func swaggerUIHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		specURL := fmt.Sprintf("http://%s/swagger/doc.yaml", r.Host)
		html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <title>P2P Exchange API</title>
  <meta charset="utf-8"/>
  <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
<script>
  SwaggerUIBundle({ url: "%s", dom_id: '#swagger-ui' })
</script>
</body>
</html>`, specURL)
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, html)
	}
}

func swaggerDocHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := os.Stat("docs/swagger.json"); err == nil {
			w.Header().Set("Content-Type", "application/json")
			http.ServeFile(w, r, "docs/swagger.json")
			return
		}
		w.Header().Set("Content-Type", "application/yaml")
		http.ServeFile(w, r, "docs/swagger.yaml")
	}
}

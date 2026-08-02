package wsserver

import (
	"net/http"

	"go.uber.org/fx"
)

var Module = fx.Module("wsserver",
	fx.Provide(New),
	fx.Invoke(func(*http.Server) {}),
)

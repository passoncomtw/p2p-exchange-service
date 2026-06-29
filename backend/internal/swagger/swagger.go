package swagger

import (
	"embed"
	"net/http"
	"path"
	"strings"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/pathvar"
)

//go:embed dist
var dist embed.FS

func RegisterRoutes(server *rest.Server) {
	server.AddRoutes([]rest.Route{
		{
			Method:  http.MethodGet,
			Path:    "/swagger",
			Handler: serveIndex(),
		},
		{
			Method:  http.MethodGet,
			Path:    "/swagger/:file",
			Handler: serveFile(),
		},
	})
}

func serveIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := dist.ReadFile("dist/index.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	}
}

func serveFile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := pathvar.Vars(r)
		name := vars["file"]
		if name == "" || strings.Contains(name, "..") {
			http.NotFound(w, r)
			return
		}

		data, err := dist.ReadFile("dist/" + name)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		switch strings.ToLower(path.Ext(name)) {
		case ".html":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		case ".css":
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
		case ".js":
			w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		case ".json":
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
		default:
			w.Header().Set("Content-Type", "application/octet-stream")
		}
		w.Write(data)
	}
}

package handlers

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
)

const version = "v1.0.0"

func VersionHandler(w http.ResponseWriter, r *http.Request) {
	httpx.OkJson(w, map[string]string{"version": version})
}

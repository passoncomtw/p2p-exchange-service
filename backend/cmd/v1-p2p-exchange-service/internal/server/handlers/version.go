package handlers

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
)

const version = "v1.0.0"

// VersionHandler 查詢服務版本。
//
//	@Summary		查詢服務版本
//	@Description	直接回傳 {"version":"v1.0.0"}，不含 {code,message,data} 包裝。
//	@Tags			system
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Router			/api/v1/version [get]
func VersionHandler(w http.ResponseWriter, r *http.Request) {
	httpx.OkJson(w, map[string]string{"version": version})
}

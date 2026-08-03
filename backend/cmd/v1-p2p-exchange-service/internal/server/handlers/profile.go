package handlers

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	app_interface "p2p-exchange/cmd/v1-p2p-exchange-service/internal/interfaces/app"
	"p2p-exchange/internal/response"
)

type ProfileHandler struct{}

func NewProfileHandler() *ProfileHandler {
	return &ProfileHandler{}
}

func (h *ProfileHandler) Handle(w http.ResponseWriter, r *http.Request) {
	payload, _ := r.Context().Value("payload").(map[string]interface{})
	username, _ := payload["username"].(string)

	httpx.WriteJsonCtx(r.Context(), w, http.StatusOK, response.Success(app_interface.ProfileResponse{
		Username: username,
	}))
}

package handlers

import (
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
	app_interface "p2p-exchange/cmd/v1-p2p-exchange-service/internal/interfaces/app"
	userrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/user"
	apierrors "p2p-exchange/internal/errors"
	"p2p-exchange/internal/response"
)

type ProfileHandler struct {
	userRepo userrepo.UserRepository
}

func NewProfileHandler(userRepo userrepo.UserRepository) *ProfileHandler {
	return &ProfileHandler{userRepo: userRepo}
}

func (h *ProfileHandler) Handle(w http.ResponseWriter, r *http.Request) {
	payload, _ := r.Context().Value("payload").(map[string]interface{})
	username, _ := payload["username"].(string)

	httpx.WriteJsonCtx(r.Context(), w, http.StatusOK, response.Success(app_interface.ProfileResponse{
		Username: username,
	}))
}

// UpdatePushToken 處理 PUT /app/profile/push-token。
// 更新對象一律取自 JWT context 的 uid，不接受 body 指定使用者。
// token 為空字串時直接回成功且不寫入 DB（維持 legacy 行為：清空 token 的情境尚未支援）。
func (h *ProfileHandler) UpdatePushToken(w http.ResponseWriter, r *http.Request) {
	var req app_interface.RegisterPushTokenRequest
	if err := httpx.Parse(r, &req); err != nil {
		httpx.WriteJsonCtx(r.Context(), w, http.StatusBadRequest, response.Fail(http.StatusBadRequest, err.Error()))
		return
	}

	if req.Token == "" {
		httpx.OkJsonCtx(r.Context(), w, response.Success(nil))
		return
	}

	uid := ctxUID(r)
	if err := h.userRepo.UpdatePushToken(r.Context(), uid, req.Token); err != nil {
		// 直接呼叫 repository，錯誤是原始 DB 錯誤：只記錄不外露，避免洩漏內部細節。
		logx.WithContext(r.Context()).Errorf("update push token for user %d failed: %v", uid, err)
		httpx.WriteJsonCtx(r.Context(), w, http.StatusInternalServerError,
			response.Fail(http.StatusInternalServerError, apierrors.ErrInternal.Message))
		return
	}

	httpx.OkJsonCtx(r.Context(), w, response.Success(nil))
}

package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"p2p-exchange/internal/infra/rdb"
	"p2p-exchange/pkg/mq"
)

// HealthzHandler 檢查 Redis/NATS/DB 三項依賴是否健康，供 k8s liveness/readiness probe 使用。
// 任一必要依賴（DB）不通即回 503；Redis/NATS 為選用依賴（未設定時為 nil），未設定時不計入檢查。
type HealthzHandler struct {
	db  sqlx.SqlConn
	rdb *rdb.Client
	mq  *mq.Client
}

func NewHealthzHandler(db sqlx.SqlConn, rdbClient *rdb.Client, mqClient *mq.Client) *HealthzHandler {
	return &HealthzHandler{db: db, rdb: rdbClient, mq: mqClient}
}

func (h *HealthzHandler) Handle(w http.ResponseWriter, r *http.Request) {
	checks := map[string]string{}
	httpStatus := http.StatusOK

	if h.rdb != nil {
		if err := h.rdb.Ping(r.Context()); err != nil {
			checks["redis"] = "unhealthy: " + err.Error()
			httpStatus = http.StatusServiceUnavailable
		} else {
			checks["redis"] = "ok"
		}
	}

	if h.mq != nil {
		if err := h.mq.Ping(); err != nil {
			checks["nats"] = "unhealthy: " + err.Error()
			httpStatus = http.StatusServiceUnavailable
		} else {
			checks["nats"] = "ok"
		}
	}

	var one int
	if err := h.db.QueryRowCtx(r.Context(), &one, "select 1"); err != nil {
		checks["database"] = "unhealthy: " + err.Error()
		httpStatus = http.StatusServiceUnavailable
	} else {
		checks["database"] = "ok"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(checks)
}

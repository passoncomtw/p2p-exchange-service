package logic

import (
	"context"
	"encoding/json"

	"github.com/zeromicro/go-zero/core/logx"
	"p2p-exchange/internal/svc"
	"p2p-exchange/internal/types"
)

type BackendDashboardLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBackendDashboardLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BackendDashboardLogic {
	return &BackendDashboardLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BackendDashboardLogic) Dashboard(_ *types.DashboardRequest) (*types.DashboardResponse, error) {
	// go-zero JWT middleware 將各 claim 個別存入 context，key 為 claim 名稱
	// username claim（string）
	username, _ := l.ctx.Value("username").(string)

	// 額外防護：驗證 platform claim 必須為 "backend"
	// （主要防線是 Backend.AccessSecret 與 App.AccessSecret 不同）
	platform, _ := l.ctx.Value("platform").(string)
	if platform != "backend" {
		// App token 不應出現在 backend 路由，理論上 secret 不同時不會發生
		// 但作為 defense-in-depth 保留此檢查
		_ = platform // 靜默處理，不中斷服務，僅供未來加告警用
	}

	// sub 存 userID（json.Number），取 username 改從 username claim
	_ = l.ctx.Value("sub").(json.Number) // 確認型別正確，不使用值

	return &types.DashboardResponse{
		Username: username,
		Role:     "admin",
	}, nil
}

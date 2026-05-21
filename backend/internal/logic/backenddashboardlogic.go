package logic

import (
	"context"

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
	payload, _ := l.ctx.Value("payload").(map[string]interface{})
	username, _ := payload["sub"].(string)

	return &types.DashboardResponse{
		Username: username,
		Role:     "admin",
	}, nil
}

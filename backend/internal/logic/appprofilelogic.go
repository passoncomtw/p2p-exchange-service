package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"p2p-exchange/internal/svc"
	"p2p-exchange/internal/types"
)

type AppProfileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppProfileLogic {
	return &AppProfileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AppProfileLogic) Profile(_ *types.ProfileRequest) (*types.ProfileResponse, error) {
	payload, _ := l.ctx.Value("payload").(map[string]interface{})
	username, _ := payload["sub"].(string)

	return &types.ProfileResponse{
		Username: username,
	}, nil
}

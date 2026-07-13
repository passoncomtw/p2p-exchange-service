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
	username, _ := payload["username"].(string)

	return &types.ProfileResponse{
		Username: username,
	}, nil
}

type RegisterPushTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterPushTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterPushTokenLogic {
	return &RegisterPushTokenLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *RegisterPushTokenLogic) Register(uid int64, req *types.RegisterPushTokenRequest) error {
	if req.Token == "" {
		return nil
	}
	return l.svcCtx.AppUser.UpdatePushToken(l.ctx, uid, req.Token)
}

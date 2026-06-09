// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"p2p-exchange/internal/svc"
	"p2p-exchange/internal/types"
)

type P2pLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewP2pLogic(ctx context.Context, svcCtx *svc.ServiceContext) *P2pLogic {
	return &P2pLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *P2pLogic) P2p(req *types.Request) (*types.Response, error) {
	return &types.Response{
		Message: "Hello " + req.Name,
	}, nil
}

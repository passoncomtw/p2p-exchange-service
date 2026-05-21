// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"

	"p2p-exchange/internal/svc"
	"p2p-exchange/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
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

func (l *P2pLogic) P2p(req *types.Request) (resp *types.Response, err error) {
	// todo: add your logic here and delete this line

	return
}

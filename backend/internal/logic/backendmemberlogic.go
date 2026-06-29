package logic

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	apierrors "p2p-exchange/internal/errors"
	"p2p-exchange/internal/svc"
	"p2p-exchange/internal/types"
)

type BackendListMembersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBackendListMembersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BackendListMembersLogic {
	return &BackendListMembersLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *BackendListMembersLogic) List(req *types.BackendListMembersRequest) (*types.BackendListMembersResponse, error) {
	rows, err := l.svcCtx.AppUser.List(l.ctx, req.Keyword, req.Limit, req.Offset)
	if err != nil {
		return nil, apierrors.ErrInternal
	}

	total, err := l.svcCtx.AppUser.Count(l.ctx, req.Keyword)
	if err != nil {
		return nil, apierrors.ErrInternal
	}

	items := make([]types.MemberItem, 0, len(rows))
	for _, r := range rows {
		email := ""
		if r.Email != nil {
			email = *r.Email
		}
		items = append(items, types.MemberItem{
			ID:        r.ID,
			Username:  r.Username,
			Email:     email,
			CreatedAt: r.CreatedAt.Format(time.RFC3339),
			UpdatedAt: r.UpdatedAt.Format(time.RFC3339),
		})
	}

	return &types.BackendListMembersResponse{List: items, Total: total}, nil
}

package backenduserrepo

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/model/entity"
)

type BackendUserRepository interface {
	FindByUsername(ctx context.Context, username string) (*entity.BackendUser, error)
}

type backendUserRepository struct {
	db sqlx.SqlConn
}

func New(db sqlx.SqlConn) BackendUserRepository {
	return &backendUserRepository{db: db}
}

func (r *backendUserRepository) FindByUsername(ctx context.Context, username string) (*entity.BackendUser, error) {
	var u entity.BackendUser
	err := r.db.QueryRowCtx(ctx, &u,
		`SELECT id, username, password_hash, role FROM backend_users WHERE username = $1`,
		username,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

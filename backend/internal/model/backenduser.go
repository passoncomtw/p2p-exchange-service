package model

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type BackendUser struct {
	ID           int64  `db:"id"`
	Username     string `db:"username"`
	PasswordHash string `db:"password_hash"`
	Role         string `db:"role"`
}

type BackendUserModel struct {
	conn sqlx.SqlConn
}

func NewBackendUserModel(conn sqlx.SqlConn) *BackendUserModel {
	return &BackendUserModel{conn: conn}
}

func (m *BackendUserModel) FindByUsername(ctx context.Context, username string) (*BackendUser, error) {
	var user BackendUser
	err := m.conn.QueryRowCtx(ctx, &user,
		`SELECT id, username, password_hash, role FROM backend_users WHERE username = $1`,
		username,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

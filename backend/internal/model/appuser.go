package model

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type AppUser struct {
	ID           int64  `db:"id"`
	Username     string `db:"username"`
	PasswordHash string `db:"password_hash"`
}

type AppUserModel struct {
	conn sqlx.SqlConn
}

func NewAppUserModel(conn sqlx.SqlConn) *AppUserModel {
	return &AppUserModel{conn: conn}
}

func (m *AppUserModel) FindByUsername(ctx context.Context, username string) (*AppUser, error) {
	var user AppUser
	err := m.conn.QueryRowCtx(ctx, &user,
		`SELECT id, username, password_hash FROM app_users WHERE username = $1`,
		username,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

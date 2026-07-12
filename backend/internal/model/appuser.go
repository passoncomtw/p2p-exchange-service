package model

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type AppUser struct {
	ID           int64     `db:"id"`
	Username     string    `db:"username"`
	PasswordHash string    `db:"password_hash"`
	Email        *string   `db:"email"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

type AppUserModel struct {
	conn sqlx.SqlConn
}

func NewAppUserModel(conn sqlx.SqlConn) *AppUserModel {
	return &AppUserModel{conn: conn}
}

func (m *AppUserModel) Create(ctx context.Context, username, passwordHash string) (*AppUser, error) {
	var user AppUser
	err := m.conn.QueryRowCtx(ctx, &user,
		`INSERT INTO app_users (username, password_hash)
		 VALUES ($1, $2)
		 RETURNING id, username, password_hash, email, created_at, updated_at`,
		username, passwordHash,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (m *AppUserModel) FindByUsername(ctx context.Context, username string) (*AppUser, error) {
	var user AppUser
	err := m.conn.QueryRowCtx(ctx, &user,
		`SELECT id, username, password_hash, email, created_at, updated_at FROM app_users WHERE username = $1`,
		username,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (m *AppUserModel) List(ctx context.Context, keyword string, limit, offset int64) ([]AppUser, error) {
	query := `SELECT id, username, password_hash, email, created_at, updated_at FROM app_users`
	args := []any{}
	argIdx := 1

	if keyword != "" {
		query += fmt.Sprintf(` WHERE username ILIKE $%d OR email ILIKE $%d`, argIdx, argIdx+1)
		like := "%" + keyword + "%"
		args = append(args, like, like)
		argIdx += 2
	}

	query += ` ORDER BY id ASC`
	query += fmt.Sprintf(` LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
	args = append(args, limit, offset)

	var users []AppUser
	err := m.conn.QueryRowsCtx(ctx, &users, query, args...)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (m *AppUserModel) Count(ctx context.Context, keyword string) (int64, error) {
	query := `SELECT COUNT(*) FROM app_users`
	args := []any{}

	if keyword != "" {
		query += ` WHERE username ILIKE $1 OR email ILIKE $2`
		like := "%" + keyword + "%"
		args = append(args, like, like)
	}

	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, args...)
	if err != nil {
		return 0, err
	}
	return count, nil
}

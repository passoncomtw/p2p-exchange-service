package userrepo

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/model/entity"
)

type UserRepository interface {
	FindByUsername(ctx context.Context, username string) (*entity.AppUser, error)
	GetPushToken(ctx context.Context, userID int64) (string, error)
}

type userRepository struct {
	db sqlx.SqlConn
}

func New(db sqlx.SqlConn) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetPushToken(ctx context.Context, userID int64) (string, error) {
	var token *string
	err := r.db.QueryRowCtx(ctx, &token,
		`SELECT expo_push_token FROM app_users WHERE id = $1`,
		userID,
	)
	if err != nil || token == nil {
		return "", err
	}
	return *token, nil
}

func (r *userRepository) FindByUsername(ctx context.Context, username string) (*entity.AppUser, error) {
	var user entity.AppUser
	err := r.db.QueryRowCtx(ctx, &user,
		`SELECT id, username, password_hash, email, expo_push_token, created_at, updated_at
		 FROM app_users WHERE username = $1`,
		username,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

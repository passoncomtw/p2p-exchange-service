package userrepo

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/model/entity"
)

type UserRepository interface {
	FindByUsername(ctx context.Context, username string) (*entity.AppUser, error)
	GetPushToken(ctx context.Context, userID int64) (string, error)
	// Create 建立 App 使用者並回傳完整資料。
	// passwordHash 必須是呼叫端已雜湊過的密碼，本方法不做任何雜湊或明文處理。
	// username 的唯一性由 app_users.username 的 UNIQUE 約束保證。
	Create(ctx context.Context, username, passwordHash string) (*entity.AppUser, error)
	// UpdatePushToken 更新使用者的 Expo 推播 token。
	UpdatePushToken(ctx context.Context, userID int64, token string) error
	// FindByID 依主鍵查詢使用者；查無資料時回傳 sqlx.ErrNotFound。
	FindByID(ctx context.Context, id int64) (*entity.AppUser, error)
	// List 分頁查詢使用者，keyword 非空時比對 username / email（不分大小寫的模糊比對）。
	List(ctx context.Context, keyword string, limit, offset int64) ([]*entity.AppUser, error)
	// Count 回傳與 List 相同過濾條件下的總筆數。
	Count(ctx context.Context, keyword string) (int64, error)
}

// defaultMemberPageSize List 未指定（或給了非正數）limit 時的預設分頁筆數。
const defaultMemberPageSize = 10

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

// userColumns 使用者查詢的共用欄位清單。
const userColumns = `id, username, password_hash, email, expo_push_token, created_at, updated_at`

// keywordFilter 關鍵字過濾條件：keyword 為空字串時不過濾。
// 關鍵字一律以查詢參數傳入（不做字串拼接），避免 SQL injection；
// $1 明確標註 ::text，避免 PostgreSQL 在與空字串比較時無法推導參數型別。
const keywordFilter = `WHERE ($1::text = '' OR username ILIKE '%' || $1::text || '%' OR email ILIKE '%' || $1::text || '%')`

func (r *userRepository) FindByID(ctx context.Context, id int64) (*entity.AppUser, error) {
	var user entity.AppUser
	err := r.db.QueryRowCtx(ctx, &user,
		`SELECT `+userColumns+` FROM app_users WHERE id = $1`,
		id,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) List(ctx context.Context, keyword string, limit, offset int64) ([]*entity.AppUser, error) {
	if limit <= 0 {
		limit = defaultMemberPageSize
	}
	if offset < 0 {
		offset = 0
	}

	var rows []*entity.AppUser
	err := r.db.QueryRowsCtx(ctx, &rows,
		fmt.Sprintf(`SELECT %s FROM app_users %s ORDER BY id DESC LIMIT $2 OFFSET $3`, userColumns, keywordFilter),
		keyword, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *userRepository) Count(ctx context.Context, keyword string) (int64, error) {
	var count int64
	err := r.db.QueryRowCtx(ctx, &count,
		`SELECT COUNT(*) FROM app_users `+keywordFilter,
		keyword,
	)
	return count, err
}

func (r *userRepository) Create(ctx context.Context, username, passwordHash string) (*entity.AppUser, error) {
	var user entity.AppUser
	err := r.db.QueryRowCtx(ctx, &user,
		`INSERT INTO app_users (username, password_hash)
		 VALUES ($1, $2)
		 RETURNING `+userColumns,
		username, passwordHash,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) UpdatePushToken(ctx context.Context, userID int64, token string) error {
	_, err := r.db.ExecCtx(ctx,
		`UPDATE app_users SET expo_push_token = $1, updated_at = NOW() WHERE id = $2`,
		token, userID,
	)
	return err
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

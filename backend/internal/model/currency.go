package model

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type CurrencyModel struct {
	conn sqlx.SqlConn
}

func NewCurrencyModel(conn sqlx.SqlConn) *CurrencyModel {
	return &CurrencyModel{conn: conn}
}

func (m *CurrencyModel) Exists(ctx context.Context, code string) (bool, error) {
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count,
		`SELECT COUNT(*) FROM currencies WHERE code = $1 AND is_active = TRUE`,
		code,
	)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

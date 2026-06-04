package model

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type PaymentMethod struct {
	ID            int64     `db:"id"`
	UserID        int64     `db:"user_id"`
	Type          string    `db:"type"`
	BankName      string    `db:"bank_name"`
	AccountName   string    `db:"account_name"`
	AccountNumber string    `db:"account_number"`
	IsActive      bool      `db:"is_active"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

type PaymentMethodModel struct {
	conn sqlx.SqlConn
}

func NewPaymentMethodModel(conn sqlx.SqlConn) *PaymentMethodModel {
	return &PaymentMethodModel{conn: conn}
}

func (m *PaymentMethodModel) Create(ctx context.Context, pm *PaymentMethod) (int64, error) {
	var id int64
	err := m.conn.QueryRowCtx(ctx, &id,
		`INSERT INTO payment_methods (user_id, type, bank_name, account_name, account_number, is_active, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		 RETURNING id`,
		pm.UserID, pm.Type, pm.BankName, pm.AccountName, pm.AccountNumber, pm.IsActive,
	)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (m *PaymentMethodModel) FindByUserID(ctx context.Context, userID int64) ([]*PaymentMethod, error) {
	var rows []*PaymentMethod
	err := m.conn.QueryRowsCtx(ctx, &rows,
		`SELECT id, user_id, type, bank_name, account_name, account_number, is_active, created_at, updated_at
		 FROM payment_methods
		 WHERE user_id = $1 AND is_active = true
		 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (m *PaymentMethodModel) FindByID(ctx context.Context, id int64) (*PaymentMethod, error) {
	var pm PaymentMethod
	err := m.conn.QueryRowCtx(ctx, &pm,
		`SELECT id, user_id, type, bank_name, account_name, account_number, is_active, created_at, updated_at
		 FROM payment_methods
		 WHERE id = $1`,
		id,
	)
	if err != nil {
		return nil, err
	}
	return &pm, nil
}

func (m *PaymentMethodModel) SoftDelete(ctx context.Context, id, userID int64) error {
	_, err := m.conn.ExecCtx(ctx,
		`UPDATE payment_methods SET is_active = false, updated_at = NOW() WHERE id = $1 AND user_id = $2`,
		id, userID,
	)
	return err
}

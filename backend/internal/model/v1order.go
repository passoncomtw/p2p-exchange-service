package model

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// V1OrderRow 為 v1 掛單的精簡視圖，沿用 listings 表並關聯 app_users 取得建立者名稱。
type V1OrderRow struct {
	ID                 int64     `db:"id"`
	Username           string    `db:"username"`
	Type               string    `db:"type"`
	CryptoCurrency     string    `db:"crypto_currency"`
	FiatCurrency       string    `db:"fiat_currency"`
	Price              float64   `db:"price"`
	Quantity           float64   `db:"quantity"`
	PaymentMethodLabel *string   `db:"payment_method_label"`
	Status             string    `db:"status"`
	CreatedAt          time.Time `db:"created_at"`
	UpdatedAt          time.Time `db:"updated_at"`
}

// V1CreateParams 為建立 v1 掛單所需的最小參數。
type V1CreateParams struct {
	UserID             int64
	Type               string
	CryptoCurrency     string
	FiatCurrency       string
	Price              float64
	Quantity           float64
	PaymentMethodLabel string
}

type V1OrderModel struct {
	conn sqlx.SqlConn
}

func NewV1OrderModel(conn sqlx.SqlConn) *V1OrderModel {
	return &V1OrderModel{conn: conn}
}

const v1SelectColumns = `l.id, u.username AS username, l.type, l.crypto_currency, l.fiat_currency,
	l.price, l.total_amount AS quantity, l.payment_method_label,
	l.status, l.created_at, l.updated_at`

// Create 建立一筆 v1 掛單。沿用 listings 表，未使用的欄位以 v1 預設值填入。
func (m *V1OrderModel) Create(ctx context.Context, p *V1CreateParams) (int64, error) {
	maxOrderFiat := p.Price * p.Quantity
	var id int64
	err := m.conn.QueryRowCtx(ctx, &id,
		`INSERT INTO listings (user_id, type, crypto_currency, fiat_currency, total_amount, remaining_amount, price,
		  min_order_fiat, max_order_fiat, payment_method_label, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $5, $6, 0, $7, $8, 'active', NOW(), NOW())
		 RETURNING id`,
		p.UserID, p.Type, p.CryptoCurrency, p.FiatCurrency, p.Quantity, p.Price, maxOrderFiat, p.PaymentMethodLabel,
	)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// FindByID 依 id 取得單筆 v1 掛單。
func (m *V1OrderModel) FindByID(ctx context.Context, id int64) (*V1OrderRow, error) {
	var row V1OrderRow
	query := fmt.Sprintf(
		`SELECT %s FROM listings l JOIN app_users u ON u.id = l.user_id WHERE l.id = $1`,
		v1SelectColumns,
	)
	if err := m.conn.QueryRowCtx(ctx, &row, query, id); err != nil {
		return nil, err
	}
	return &row, nil
}

// ListByUserID 取得指定使用者的所有 v1 掛單。
func (m *V1OrderModel) ListByUserID(ctx context.Context, userID int64) ([]*V1OrderRow, error) {
	var rows []*V1OrderRow
	query := fmt.Sprintf(
		`SELECT %s FROM listings l JOIN app_users u ON u.id = l.user_id
		 WHERE l.user_id = $1 ORDER BY l.created_at DESC`,
		v1SelectColumns,
	)
	if err := m.conn.QueryRowsCtx(ctx, &rows, query, userID); err != nil {
		return nil, err
	}
	return rows, nil
}

// ListAll 取得全部 v1 掛單，dbStatus 為空字串時不篩選。
func (m *V1OrderModel) ListAll(ctx context.Context, dbStatus string) ([]*V1OrderRow, error) {
	var rows []*V1OrderRow
	query := fmt.Sprintf(
		`SELECT %s FROM listings l JOIN app_users u ON u.id = l.user_id`,
		v1SelectColumns,
	)
	args := []interface{}{}
	if dbStatus != "" {
		query += " WHERE l.status = $1"
		args = append(args, dbStatus)
	}
	query += " ORDER BY l.created_at DESC"
	if err := m.conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
		return nil, err
	}
	return rows, nil
}

// UpdateStatus 更新掛單狀態。
func (m *V1OrderModel) UpdateStatus(ctx context.Context, id int64, dbStatus string) error {
	_, err := m.conn.ExecCtx(ctx,
		`UPDATE listings SET status = $1, updated_at = NOW() WHERE id = $2`,
		dbStatus, id,
	)
	return err
}

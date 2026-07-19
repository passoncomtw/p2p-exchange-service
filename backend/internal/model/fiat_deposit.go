package model

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type FiatDeposit struct {
	ID              int64      `db:"id"`
	UserID          int64      `db:"user_id"`
	Currency        string     `db:"currency"`
	Amount          string     `db:"amount"`
	EcpayOrderNo    *string    `db:"ecpay_order_no"`
	MerchantTradeNo string     `db:"merchant_trade_no"`
	Status          string     `db:"status"`
	PaymentType     *string    `db:"payment_type"`
	PaidAt          *time.Time `db:"paid_at"`
	CreatedAt       time.Time  `db:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at"`
}

type FiatDepositModel struct{ conn sqlx.SqlConn }

func NewFiatDepositModel(conn sqlx.SqlConn) *FiatDepositModel {
	return &FiatDepositModel{conn: conn}
}

func (m *FiatDepositModel) Create(ctx context.Context, d *FiatDeposit) error {
	return m.conn.QueryRowCtx(ctx, &d.ID,
		`INSERT INTO fiat_deposits (user_id, currency, amount, merchant_trade_no, status)
		 VALUES ($1, $2, $3, $4, 'pending') RETURNING id`,
		d.UserID, d.Currency, d.Amount, d.MerchantTradeNo,
	)
}

func (m *FiatDepositModel) FindByMerchantTradeNo(ctx context.Context, tradeNo string) (*FiatDeposit, error) {
	var d FiatDeposit
	err := m.conn.QueryRowCtx(ctx, &d,
		`SELECT id, user_id, currency, amount::text, ecpay_order_no, merchant_trade_no, status, payment_type, paid_at, created_at, updated_at
		 FROM fiat_deposits WHERE merchant_trade_no = $1`,
		tradeNo,
	)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (m *FiatDepositModel) UpdatePaid(ctx context.Context, id int64, ecpayOrderNo, paymentType string, paidAt time.Time) error {
	_, err := m.conn.ExecCtx(ctx,
		`UPDATE fiat_deposits SET status='paid', ecpay_order_no=$2, payment_type=$3, paid_at=$4, updated_at=NOW() WHERE id=$1`,
		id, ecpayOrderNo, paymentType, paidAt,
	)
	return err
}

func (m *FiatDepositModel) UpdateFailed(ctx context.Context, id int64) error {
	_, err := m.conn.ExecCtx(ctx,
		`UPDATE fiat_deposits SET status='failed', updated_at=NOW() WHERE id=$1`,
		id,
	)
	return err
}

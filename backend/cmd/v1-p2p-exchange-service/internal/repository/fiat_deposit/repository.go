package fiatdepositrepo

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/model/entity"
)

// fiatDepositColumns 查詢共用欄位；amount 以 ::text 取出避免浮點誤差。
const fiatDepositColumns = `id, user_id, currency, amount::text, ecpay_order_no, merchant_trade_no, status, payment_type, paid_at, created_at, updated_at`

// FiatDepositRepository 存取 fiat_deposits 表（ECPay TWD 入金記錄）。
type FiatDepositRepository interface {
	// Create 新增一筆 pending 入金記錄，回傳新的主鍵 ID。
	Create(ctx context.Context, d *entity.FiatDeposit) (int64, error)
	// ListByUserID 分頁查詢使用者的入金記錄，依建立時間新到舊排序。
	ListByUserID(ctx context.Context, userID int64, limit, offset int64) ([]*entity.FiatDeposit, error)
	// CountByUserID 回傳使用者入金記錄總筆數。
	CountByUserID(ctx context.Context, userID int64) (int64, error)
	// FindByMerchantTradeNo 依平台產生的交易號查詢；查無資料時回傳 sqlx.ErrNotFound。
	// merchant_trade_no 有 UNIQUE 約束，最多只會有一筆。
	FindByMerchantTradeNo(ctx context.Context, tradeNo string) (*entity.FiatDeposit, error)
	// ConfirmPaidInTx 在呼叫端傳入的交易 session 內把狀態由 pending 轉為 paid，
	// 同時記錄 ECPay 回傳的訂單編號、付款方式與付款時間，回傳異動筆數。
	// 回傳 0 代表該筆已被處理過（ECPay 重送通知），呼叫端必須視為冪等成功並「不再入帳」，
	// 否則同一筆入金會被重複計入餘額。
	ConfirmPaidInTx(ctx context.Context, session sqlx.Session, id int64, ecpayOrderNo, paymentType string, paidAt time.Time) (int64, error)
	// UpdateFailed 把 pending 記錄轉為 failed（ECPay 回報付款失敗）。
	// 不涉及資金異動所以不需要交易；仍帶 WHERE status='pending' 守衛，
	// 避免已入帳（paid）的記錄被延遲抵達的失敗通知覆寫成 failed。
	UpdateFailed(ctx context.Context, id int64) error
}

type fiatDepositRepository struct {
	db sqlx.SqlConn
}

func New(db sqlx.SqlConn) FiatDepositRepository {
	return &fiatDepositRepository{db: db}
}

func (r *fiatDepositRepository) Create(ctx context.Context, d *entity.FiatDeposit) (int64, error) {
	var id int64
	err := r.db.QueryRowCtx(ctx, &id,
		`INSERT INTO fiat_deposits (user_id, currency, amount, merchant_trade_no, status)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id`,
		d.UserID, d.Currency, d.Amount, d.MerchantTradeNo, d.Status,
	)
	return id, err
}

func (r *fiatDepositRepository) ListByUserID(ctx context.Context, userID int64, limit, offset int64) ([]*entity.FiatDeposit, error) {
	var rows []*entity.FiatDeposit
	err := r.db.QueryRowsCtx(ctx, &rows,
		`SELECT id, user_id, currency, amount::text, ecpay_order_no, merchant_trade_no, status, payment_type, paid_at, created_at, updated_at
		 FROM fiat_deposits WHERE user_id = $1
		 ORDER BY created_at DESC
		 LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	return rows, err
}

func (r *fiatDepositRepository) CountByUserID(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := r.db.QueryRowCtx(ctx, &count,
		`SELECT COUNT(*) FROM fiat_deposits WHERE user_id = $1`,
		userID,
	)
	return count, err
}

func (r *fiatDepositRepository) FindByMerchantTradeNo(ctx context.Context, tradeNo string) (*entity.FiatDeposit, error) {
	var d entity.FiatDeposit
	err := r.db.QueryRowCtx(ctx, &d,
		`SELECT `+fiatDepositColumns+` FROM fiat_deposits WHERE merchant_trade_no = $1`,
		tradeNo,
	)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *fiatDepositRepository) ConfirmPaidInTx(ctx context.Context, session sqlx.Session, id int64, ecpayOrderNo, paymentType string, paidAt time.Time) (int64, error) {
	// 條件式 UPDATE：WHERE 帶 status = 'pending'，ECPay 重送通知時不會異動任何列，
	// 由 RowsAffected 讓呼叫端判斷是否真的完成了 pending → paid 的狀態轉換。
	// legacy 的 UpdatePaid 沒有這道守衛（且入帳在交易外），兩個幾乎同時抵達的通知會重複入帳。
	res, err := session.ExecCtx(ctx,
		`UPDATE fiat_deposits SET status = 'paid', ecpay_order_no = $2, payment_type = $3, paid_at = $4, updated_at = NOW()
		 WHERE id = $1 AND status = 'pending'`,
		id, ecpayOrderNo, paymentType, paidAt,
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *fiatDepositRepository) UpdateFailed(ctx context.Context, id int64) error {
	_, err := r.db.ExecCtx(ctx,
		`UPDATE fiat_deposits SET status = 'failed', updated_at = NOW()
		 WHERE id = $1 AND status = 'pending'`,
		id,
	)
	return err
}

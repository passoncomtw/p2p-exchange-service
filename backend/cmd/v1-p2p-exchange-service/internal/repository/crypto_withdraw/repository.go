package cryptowithdrawrepo

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/model/entity"
)

// defaultScanLimit ListPending / ListBroadcasting 未指定（或給了非正數）limit 時的預設上限。
const defaultScanLimit = 100

// cryptoWithdrawalColumns 查詢共用欄位；amount 以 ::text 取出避免浮點誤差。
const cryptoWithdrawalColumns = `id, user_id, currency, amount::text, to_address, tx_hash, status, broadcast_at, confirmed_at, created_at, updated_at`

// CryptoWithdrawRepository 存取 crypto_withdrawals 表（鏈上 USDT 提領記錄）。
//
// 狀態機：pending（待廣播）→ broadcasting（已認領/已廣播）→ confirmed（已確認並扣款）；
// 廣播失敗則 broadcasting → failed（餘額解凍）。所有狀態轉換一律用條件式 UPDATE
// 帶上「目前狀態」守衛，並以 RowsAffected 判斷是否真的完成轉換。
type CryptoWithdrawRepository interface {
	// CreateInTx 在呼叫端傳入的交易 session 內建立提領申請（狀態固定 'pending'），回傳寫入後的完整記錄。
	// 提領申請必須與餘額凍結在同一個交易內，避免凍結成功卻沒有對應單據（或反之）。
	CreateInTx(ctx context.Context, session sqlx.Session, w *entity.CryptoWithdrawal) (*entity.CryptoWithdrawal, error)
	// ListPending 取出待廣播的提領記錄（建立時間由舊到新），limit <= 0 時使用預設上限。
	ListPending(ctx context.Context, limit int) ([]*entity.CryptoWithdrawal, error)
	// ListBroadcasting 取出已廣播、待鏈上確認的提領記錄，limit <= 0 時使用預設上限。
	ListBroadcasting(ctx context.Context, limit int) ([]*entity.CryptoWithdrawal, error)
	// ClaimForBroadcastInTx 原子認領一筆 pending 記錄準備廣播（pending → broadcasting），回傳異動筆數。
	//
	// 回傳 0 代表該筆已被其他流程（其他副本、legacy job、上一輪殘留）認領，
	// 呼叫端「必須」直接跳過，不得再呼叫任何鏈上 API：
	// 鏈上轉帳不可回滾，重複廣播等於熱錢包真實對外多付一次錢。
	// 這是唯一可信的併發防線，Redis 鎖只是減少無謂重試的最佳化。
	ClaimForBroadcastInTx(ctx context.Context, session sqlx.Session, id int64) (int64, error)
	// UpdateBroadcastedInTx 在認領成功、鏈上廣播完成後補寫 tx_hash 與 broadcast_at。
	// txHash 為空字串時寫入 NULL（無法從鏈上回應解析出 txID 的情況），
	// 避免多筆記錄寫入相同的佔位值而撞上 tx_hash 的 UNIQUE 約束。
	UpdateBroadcastedInTx(ctx context.Context, session sqlx.Session, id int64, txHash string, broadcastAt time.Time) error
	// UpdateFailedInTx 把已認領（broadcasting）的記錄轉為 failed，回傳異動筆數。
	// 回傳 0 代表狀態已被其他流程推進，呼叫端必須視為冪等成功並「不再解凍餘額」，
	// 否則同一筆提領會被重複解凍。
	UpdateFailedInTx(ctx context.Context, session sqlx.Session, id int64) (int64, error)
	// ConfirmInTx 在呼叫端傳入的交易 session 內把狀態由 broadcasting 轉為 confirmed，回傳異動筆數。
	// 回傳 0 代表該筆已被其他流程確認過，呼叫端必須視為冪等成功並「不再扣款」，
	// 否則同一筆提領會被重複自凍結餘額扣除。
	ConfirmInTx(ctx context.Context, session sqlx.Session, id int64, confirmedAt time.Time) (int64, error)
	// ListByUserID 分頁查詢使用者的提領記錄，依建立時間新到舊排序。
	ListByUserID(ctx context.Context, userID int64, limit, offset int64) ([]*entity.CryptoWithdrawal, error)
	// CountByUserID 回傳使用者提領記錄總筆數。
	CountByUserID(ctx context.Context, userID int64) (int64, error)
}

type cryptoWithdrawRepository struct {
	db sqlx.SqlConn
}

func New(db sqlx.SqlConn) CryptoWithdrawRepository {
	return &cryptoWithdrawRepository{db: db}
}

func (r *cryptoWithdrawRepository) CreateInTx(ctx context.Context, session sqlx.Session, w *entity.CryptoWithdrawal) (*entity.CryptoWithdrawal, error) {
	// status 由 SQL 固定為 'pending'，不採用呼叫端傳入的值：
	// 後續狀態只能由 Job 的認領／確認流程推進。
	var row entity.CryptoWithdrawal
	err := session.QueryRowCtx(ctx, &row,
		`INSERT INTO crypto_withdrawals (user_id, currency, amount, to_address, status)
		 VALUES ($1, $2, $3, $4, 'pending')
		 RETURNING `+cryptoWithdrawalColumns,
		w.UserID, w.Currency, w.Amount, w.ToAddress,
	)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *cryptoWithdrawRepository) ListPending(ctx context.Context, limit int) ([]*entity.CryptoWithdrawal, error) {
	if limit <= 0 {
		limit = defaultScanLimit
	}
	var rows []*entity.CryptoWithdrawal
	err := r.db.QueryRowsCtx(ctx, &rows,
		`SELECT `+cryptoWithdrawalColumns+` FROM crypto_withdrawals
		 WHERE status = 'pending' ORDER BY created_at ASC LIMIT $1`,
		limit,
	)
	return rows, err
}

func (r *cryptoWithdrawRepository) ListBroadcasting(ctx context.Context, limit int) ([]*entity.CryptoWithdrawal, error) {
	if limit <= 0 {
		limit = defaultScanLimit
	}
	var rows []*entity.CryptoWithdrawal
	err := r.db.QueryRowsCtx(ctx, &rows,
		`SELECT `+cryptoWithdrawalColumns+` FROM crypto_withdrawals
		 WHERE status = 'broadcasting' ORDER BY created_at ASC LIMIT $1`,
		limit,
	)
	return rows, err
}

func (r *cryptoWithdrawRepository) ClaimForBroadcastInTx(ctx context.Context, session sqlx.Session, id int64) (int64, error) {
	// 條件式 UPDATE（WHERE status = 'pending'）：同一筆記錄同時被兩個執行緒認領時，
	// PostgreSQL 的列鎖讓第二個交易看到已轉為 'broadcasting' 的值，異動 0 列。
	// 呼叫端必須在觸發任何鏈上 API 之前先取得 RowsAffected = 1。
	res, err := session.ExecCtx(ctx,
		`UPDATE crypto_withdrawals SET status = 'broadcasting', updated_at = NOW()
		 WHERE id = $1 AND status = 'pending'`,
		id,
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *cryptoWithdrawRepository) UpdateBroadcastedInTx(ctx context.Context, session sqlx.Session, id int64, txHash string, broadcastAt time.Time) error {
	// 帶 status = 'broadcasting' 守衛純屬保險：正常路徑只有認領成功的執行緒會走到這裡。
	// tx_hash 撞 UNIQUE（極端情況下鏈上回傳同一個 txID）時原樣往上拋，由呼叫端記錄後人工處理。
	_, err := session.ExecCtx(ctx,
		`UPDATE crypto_withdrawals SET tx_hash = NULLIF($2, ''), broadcast_at = $3, updated_at = NOW()
		 WHERE id = $1 AND status = 'broadcasting'`,
		id, txHash, broadcastAt,
	)
	return err
}

func (r *cryptoWithdrawRepository) UpdateFailedInTx(ctx context.Context, session sqlx.Session, id int64) (int64, error) {
	res, err := session.ExecCtx(ctx,
		`UPDATE crypto_withdrawals SET status = 'failed', updated_at = NOW()
		 WHERE id = $1 AND status = 'broadcasting'`,
		id,
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *cryptoWithdrawRepository) ConfirmInTx(ctx context.Context, session sqlx.Session, id int64, confirmedAt time.Time) (int64, error) {
	// 條件式 UPDATE：WHERE 帶 status = 'broadcasting'，重複觸發時不會異動任何列，
	// 由 RowsAffected 讓呼叫端判斷是否真的完成了 broadcasting → confirmed 的狀態轉換。
	// legacy 版本沒有這道守衛且扣款在交易外，重跑確認流程會重複扣款（P0-3）。
	res, err := session.ExecCtx(ctx,
		`UPDATE crypto_withdrawals SET status = 'confirmed', confirmed_at = $2, updated_at = NOW()
		 WHERE id = $1 AND status = 'broadcasting'`,
		id, confirmedAt,
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *cryptoWithdrawRepository) ListByUserID(ctx context.Context, userID int64, limit, offset int64) ([]*entity.CryptoWithdrawal, error) {
	var rows []*entity.CryptoWithdrawal
	err := r.db.QueryRowsCtx(ctx, &rows,
		`SELECT `+cryptoWithdrawalColumns+` FROM crypto_withdrawals
		 WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	return rows, err
}

func (r *cryptoWithdrawRepository) CountByUserID(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := r.db.QueryRowCtx(ctx, &count,
		`SELECT COUNT(*) FROM crypto_withdrawals WHERE user_id = $1`,
		userID,
	)
	return count, err
}

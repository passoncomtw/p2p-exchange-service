package orderrepo

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/model/entity"
)

const orderSelectCols = `id, order_no, listing_id, listing_type, seller_id, buyer_id, crypto_currency, fiat_currency,
  crypto_amount, price, fiat_amount, platform_fee_base, platform_fee_amount, payment_fee_base, payment_fee_amount,
  total_fee, total_amount, payment_method_id, status, payment_deadline, paid_at, confirmed_at, completed_at,
  cancelled_at, cancel_reason, created_at, updated_at`

// activeOrderStatusesSQL 為訂單的非終態，SQL 片段與 p2p-listing 規格「進行中訂單」的定義一致。
const activeOrderStatusesSQL = `('matched', 'paid', 'releasing', 'disputed')`

const (
	insertOrderQuery = `INSERT INTO orders (order_no, listing_id, listing_type, seller_id, buyer_id, crypto_currency, fiat_currency,
		  crypto_amount, price, fiat_amount, platform_fee_base, platform_fee_amount, payment_fee_base, payment_fee_amount,
		  total_fee, total_amount, payment_method_id, status, payment_deadline, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, NOW(), NOW())
		 RETURNING id`

	// hasActiveOrdersByListingQuery 刻意不加任何鎖子句（FOR UPDATE / FOR SHARE）：
	// 呼叫端已持有掛單列的 FOR UPDATE 鎖，建單與釋放都先鎖同一列，
	// 在 READ COMMITTED 下這個一般 SELECT 能看到所有已提交的訂單，不會漏掉並發建立的訂單。
	hasActiveOrdersByListingQuery = `SELECT EXISTS (SELECT 1 FROM orders WHERE listing_id = $1 AND status IN ` + activeOrderStatusesSQL + `)`
)

type OrderRepository interface {
	Create(ctx context.Context, o *entity.Order) (int64, error)
	// CreateInTx 與 Create 相同，但在呼叫端傳入的交易 session 內寫入，
	// 讓「凍結 + 寫入訂單 + 扣減剩餘量 + 託管紀錄 + 狀態日誌」是同一個原子交易。
	CreateInTx(ctx context.Context, session sqlx.Session, o *entity.Order) (int64, error)
	// HasActiveOrdersByListingInTx 在交易內查詢該掛單是否仍有非終態訂單
	// （matched、paid、releasing、disputed）。必須在取得掛單列的 FOR UPDATE 鎖之後呼叫。
	HasActiveOrdersByListingInTx(ctx context.Context, session sqlx.Session, listingID int64) (bool, error)
	FindByID(ctx context.Context, id int64) (*entity.Order, error)
	FindExpired(ctx context.Context) ([]*entity.Order, error)
	List(ctx context.Context, userID int64, role, status string, limit, offset int64) ([]*entity.Order, error)
	BackendList(ctx context.Context, status string, limit, offset int64) ([]*entity.Order, error)
	// UpdateStatusInTx 在呼叫端傳入的交易 session 內，以條件式 UPDATE（WHERE status = fromStatus）
	// 把訂單從 fromStatus 轉為 toStatus，回傳影響列數。影響 0 列代表訂單狀態已被其他請求改變，
	// 呼叫端必須據此中止後續的錢包操作。
	UpdateStatusInTx(ctx context.Context, session sqlx.Session, id int64, fromStatus, toStatus string, extras map[string]interface{}) (int64, error)
	// UpdateStatus 與 UpdateStatusInTx 相同的條件式狀態轉換，但不在交易內執行。
	UpdateStatus(ctx context.Context, id int64, fromStatus, toStatus string, extras map[string]interface{}) (int64, error)
}

type orderRepository struct {
	db sqlx.SqlConn
}

func New(db sqlx.SqlConn) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(ctx context.Context, o *entity.Order) (int64, error) {
	var id int64
	err := r.db.QueryRowCtx(ctx, &id, insertOrderQuery, insertOrderArgs(o)...)
	return id, err
}

func (r *orderRepository) CreateInTx(ctx context.Context, session sqlx.Session, o *entity.Order) (int64, error) {
	var id int64
	err := session.QueryRowCtx(ctx, &id, insertOrderQuery, insertOrderArgs(o)...)
	return id, err
}

func insertOrderArgs(o *entity.Order) []interface{} {
	return []interface{}{
		o.OrderNo, o.ListingID, o.ListingType, o.SellerID, o.BuyerID, o.CryptoCurrency, o.FiatCurrency,
		o.CryptoAmount, o.Price, o.FiatAmount, o.PlatformFeeBase, o.PlatformFeeAmount, o.PaymentFeeBase, o.PaymentFeeAmount,
		o.TotalFee, o.TotalAmount, o.PaymentMethodID, o.Status, o.PaymentDeadline,
	}
}

func (r *orderRepository) HasActiveOrdersByListingInTx(ctx context.Context, session sqlx.Session, listingID int64) (bool, error) {
	var exists bool
	err := session.QueryRowCtx(ctx, &exists, hasActiveOrdersByListingQuery, listingID)
	return exists, err
}

func (r *orderRepository) FindByID(ctx context.Context, id int64) (*entity.Order, error) {
	var o entity.Order
	err := r.db.QueryRowCtx(ctx, &o,
		`SELECT `+orderSelectCols+` FROM orders WHERE id = $1`,
		id,
	)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *orderRepository) List(ctx context.Context, userID int64, role, status string, limit, offset int64) ([]*entity.Order, error) {
	query := `SELECT ` + orderSelectCols + ` FROM orders WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	switch role {
	case "buyer":
		query += fmt.Sprintf(" AND buyer_id = $%d", argIdx)
		args = append(args, userID)
		argIdx++
	case "seller":
		query += fmt.Sprintf(" AND seller_id = $%d", argIdx)
		args = append(args, userID)
		argIdx++
	default:
		query += fmt.Sprintf(" AND (buyer_id = $%d OR seller_id = $%d)", argIdx, argIdx)
		args = append(args, userID, userID)
		argIdx++
	}

	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	var rows []*entity.Order
	err := r.db.QueryRowsCtx(ctx, &rows, query, args...)
	return rows, err
}

func (r *orderRepository) BackendList(ctx context.Context, status string, limit, offset int64) ([]*entity.Order, error) {
	query := `SELECT ` + orderSelectCols + ` FROM orders WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	var rows []*entity.Order
	err := r.db.QueryRowsCtx(ctx, &rows, query, args...)
	return rows, err
}

func (r *orderRepository) UpdateStatus(ctx context.Context, id int64, fromStatus, toStatus string, extras map[string]interface{}) (int64, error) {
	return buildStatusUpdate(r.db, ctx, id, fromStatus, toStatus, extras)
}

func (r *orderRepository) UpdateStatusInTx(ctx context.Context, session sqlx.Session, id int64, fromStatus, toStatus string, extras map[string]interface{}) (int64, error) {
	query, args := buildStatusUpdateQuery(id, fromStatus, toStatus, extras)
	result, err := session.ExecCtx(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	n, _ := result.RowsAffected()
	return n, nil
}

func (r *orderRepository) FindExpired(ctx context.Context) ([]*entity.Order, error) {
	var rows []*entity.Order
	err := r.db.QueryRowsCtx(ctx, &rows,
		`SELECT `+orderSelectCols+` FROM orders WHERE status = 'matched' AND payment_deadline < NOW()`,
	)
	return rows, err
}

func buildStatusUpdate(db sqlx.SqlConn, ctx context.Context, id int64, fromStatus, toStatus string, extras map[string]interface{}) (int64, error) {
	query, args := buildStatusUpdateQuery(id, fromStatus, toStatus, extras)
	result, err := db.ExecCtx(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// buildStatusUpdateQuery 組出條件式狀態轉換的 UPDATE 語句。
// WHERE 必須同時比對 id 與 fromStatus：讀取訂單（FindByID）到寫入之間若有其他請求先改了狀態，
// 這次 UPDATE 只會影響 0 列，避免同一筆訂單被重複放幣、解凍或退款。
func buildStatusUpdateQuery(id int64, fromStatus, toStatus string, extras map[string]interface{}) (string, []interface{}) {
	query := "UPDATE orders SET status = $1, updated_at = NOW()"
	args := []interface{}{toStatus}
	argIdx := 2

	for key, val := range extras {
		switch key {
		case "paid_at", "confirmed_at", "completed_at", "cancelled_at", "cancel_reason":
			query += fmt.Sprintf(", %s = $%d", key, argIdx)
			args = append(args, val)
			argIdx++
		}
	}

	query += fmt.Sprintf(" WHERE id = $%d AND status = $%d", argIdx, argIdx+1)
	args = append(args, id, fromStatus)
	return query, args
}

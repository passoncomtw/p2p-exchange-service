package order_service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/model/entity"
	escrowrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/escrow_record"
	listingrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/listing"
	orderrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/order"
	orderstatuslogrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/order_status_log"
	paymentrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/payment_method"
	walletrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/wallet"
	apierrors "p2p-exchange/internal/errors"
)

// errInjected 為故障注入回傳的錯誤，用來模擬資料庫在任一步失敗。
var errInjected = errors.New("injected failure")

// errOutsideTx 代表 repository 的 InTx 方法收到的不是交易 session（呼叫端在交易外執行了寫入）。
var errOutsideTx = errors.New("called outside transaction")

type walletKey struct {
	userID   int64
	currency string
}

type balance struct {
	available, frozen decimal.Decimal
}

func (b balance) equals(o balance) bool {
	return b.available.Equal(o.available) && b.frozen.Equal(o.frozen)
}

// toDecimal 把測試用的數值字面量轉成 decimal；測試輸入以 float64 字面量書寫較易閱讀，正式程式碼不使用 float64。
func toDecimal(v float64) decimal.Decimal {
	return decimal.NewFromFloat(v)
}

// ledgerState 是會隨交易回滾的資料狀態，模擬資料庫內容。
type ledgerState struct {
	listing     entity.Listing
	orders      map[int64]entity.Order
	nextOrderID int64
	wallets     map[walletKey]balance
	escrows     []escrowrepo.EscrowRecord
	logs        []orderstatuslogrepo.OrderStatusLog
}

func (s ledgerState) clone() ledgerState {
	c := s
	c.orders = make(map[int64]entity.Order, len(s.orders))
	for k, v := range s.orders {
		c.orders[k] = v
	}
	c.wallets = make(map[walletKey]balance, len(s.wallets))
	for k, v := range s.wallets {
		c.wallets[k] = v
	}
	c.escrows = append([]escrowrepo.EscrowRecord(nil), s.escrows...)
	c.logs = append([]orderstatuslogrepo.OrderStatusLog(nil), s.logs...)
	return c
}

// fakeStore 持有可回滾的 state，以及不隨交易回滾的觀測紀錄與故障注入設定。
type fakeStore struct {
	state ledgerState

	inTx       bool
	rollbacks  int
	calls      []string
	failStep   string
	publishLog []bool // 每次發布 NATS 時是否仍在交易內

	// forceOrderRows 不為 nil 時覆寫訂單條件式 UPDATE 的影響列數（模擬並發請求已先改變狀態）。
	forceOrderRows *int64
	// forceDeductRows 不為 nil 時覆寫 DeductAmountInTx 的影響列數。
	forceDeductRows *int64

	freezeAmounts   []string
	unfreezeAmounts []decimal.Decimal
}

func newFakeStore(listing entity.Listing) *fakeStore {
	return &fakeStore{state: ledgerState{
		listing:     listing,
		orders:      map[int64]entity.Order{},
		nextOrderID: testOrderID,
		wallets:     map[walletKey]balance{},
	}}
}

// step 記錄一次 repository 呼叫；若為指定的故障步驟則回傳 errInjected。
func (st *fakeStore) step(name string) error {
	st.calls = append(st.calls, name)
	if st.failStep == name {
		return errInjected
	}
	return nil
}

func (st *fakeStore) requireTx(session sqlx.Session) error {
	if _, ok := session.(fakeSession); !ok || !st.inTx {
		return errOutsideTx
	}
	return nil
}

func (st *fakeStore) wallet(userID int64) balance {
	return st.state.wallets[walletKey{userID, "USDT"}]
}

func (st *fakeStore) setWallet(userID int64, available, frozen float64) {
	st.state.wallets[walletKey{userID, "USDT"}] = balance{available: toDecimal(available), frozen: toDecimal(frozen)}
}

// totalUSDT 回傳所有錢包 available + frozen 的總和，用來驗證資金守恆。
func (st *fakeStore) totalUSDT() decimal.Decimal {
	sum := decimal.Zero
	for k, b := range st.state.wallets {
		if k.currency == "USDT" {
			sum = sum.Add(b.available).Add(b.frozen)
		}
	}
	return sum
}

func (st *fakeStore) hasCall(name string) bool {
	for _, c := range st.calls {
		if c == name {
			return true
		}
	}
	return false
}

// conditionalOrderUpdate 模擬 `UPDATE orders ... WHERE id = ? AND status = from`。
func (st *fakeStore) conditionalOrderUpdate(id int64, from, to string) int64 {
	if st.forceOrderRows != nil {
		return *st.forceOrderRows
	}
	o, ok := st.state.orders[id]
	if !ok || o.Status != from {
		return 0
	}
	o.Status = to
	st.state.orders[id] = o
	return 1
}

// fakeConn 只實作 TransactCtx：fn 失敗時把 state 還原成交易開始前的快照，模擬資料庫回滾。
// 其餘方法沿用內嵌的 nil 介面，走到未預期的路徑會直接 panic。
type fakeConn struct {
	sqlx.SqlConn
	store *fakeStore
}

func (c fakeConn) TransactCtx(ctx context.Context, fn func(context.Context, sqlx.Session) error) error {
	snapshot := c.store.state.clone()
	c.store.inTx = true
	err := fn(ctx, fakeSession{store: c.store})
	c.store.inTx = false
	if err != nil {
		c.store.state = snapshot
		c.store.rollbacks++
	}
	return err
}

// fakeSession 只實作 ExecCtx，供逾時消費者直接在 session 上執行的條件式 UPDATE 使用。
type fakeSession struct {
	sqlx.Session
	store *fakeStore
}

type fakeResult struct {
	rows int64
}

func (r fakeResult) LastInsertId() (int64, error) { return 0, nil }
func (r fakeResult) RowsAffected() (int64, error) { return r.rows, nil }

func (s fakeSession) ExecCtx(_ context.Context, query string, args ...any) (sql.Result, error) {
	q := strings.Join(strings.Fields(query), " ")
	if !strings.HasPrefix(q, "UPDATE orders SET status = 'timeout'") || !strings.Contains(q, "WHERE id = $1 AND status = 'matched'") {
		return nil, fmt.Errorf("unexpected exec: %s", q)
	}
	if err := s.store.step("updateOrder:matched->timeout"); err != nil {
		return nil, err
	}
	return fakeResult{rows: s.store.conditionalOrderUpdate(args[0].(int64), "matched", "timeout")}, nil
}

type fakeOrderRepo struct {
	orderrepo.OrderRepository
	store *fakeStore
}

func (r *fakeOrderRepo) FindByID(_ context.Context, id int64) (*entity.Order, error) {
	o, ok := r.store.state.orders[id]
	if !ok {
		return nil, sqlx.ErrNotFound
	}
	return &o, nil
}

func (r *fakeOrderRepo) CreateInTx(_ context.Context, session sqlx.Session, o *entity.Order) (int64, error) {
	if err := r.store.requireTx(session); err != nil {
		return 0, err
	}
	if err := r.store.step("createOrder"); err != nil {
		return 0, err
	}
	id := r.store.state.nextOrderID
	r.store.state.nextOrderID++
	stored := *o
	stored.ID = id
	r.store.state.orders[id] = stored
	return id, nil
}

func (r *fakeOrderRepo) UpdateStatus(_ context.Context, id int64, fromStatus, toStatus string, _ map[string]interface{}) (int64, error) {
	if err := r.store.step("updateOrder:" + fromStatus + "->" + toStatus); err != nil {
		return 0, err
	}
	return r.store.conditionalOrderUpdate(id, fromStatus, toStatus), nil
}

func (r *fakeOrderRepo) UpdateStatusInTx(_ context.Context, session sqlx.Session, id int64, fromStatus, toStatus string, _ map[string]interface{}) (int64, error) {
	if err := r.store.requireTx(session); err != nil {
		return 0, err
	}
	if err := r.store.step("updateOrder:" + fromStatus + "->" + toStatus); err != nil {
		return 0, err
	}
	return r.store.conditionalOrderUpdate(id, fromStatus, toStatus), nil
}

// fakeListingRepo 刻意不實作 FindByID：建單與釋放若改用不加鎖的讀取會直接 panic。
type fakeListingRepo struct {
	listingrepo.ListingRepository
	store *fakeStore
}

func (r *fakeListingRepo) FindByIDForUpdateInTx(_ context.Context, session sqlx.Session, id int64) (*entity.Listing, error) {
	if err := r.store.requireTx(session); err != nil {
		return nil, err
	}
	if err := r.store.step("lockListing"); err != nil {
		return nil, err
	}
	if r.store.state.listing.ID != id {
		return nil, sqlx.ErrNotFound
	}
	l := r.store.state.listing
	return &l, nil
}

func (r *fakeListingRepo) DeductAmountInTx(_ context.Context, session sqlx.Session, id int64, amount decimal.Decimal) (int64, error) {
	if err := r.store.requireTx(session); err != nil {
		return 0, err
	}
	if err := r.store.step("deductListing"); err != nil {
		return 0, err
	}
	if r.store.forceDeductRows != nil {
		return *r.store.forceDeductRows, nil
	}
	l := &r.store.state.listing
	if l.ID != id || l.Status != "active" || l.RemainingAmount.LessThan(amount) {
		return 0, nil
	}
	l.RemainingAmount = l.RemainingAmount.Sub(amount)
	return 1, nil
}

func (r *fakeListingRepo) RestoreAmountInTx(_ context.Context, session sqlx.Session, id int64, amount decimal.Decimal) error {
	if err := r.store.requireTx(session); err != nil {
		return err
	}
	if err := r.store.step("restoreListing"); err != nil {
		return err
	}
	if r.store.state.listing.ID == id {
		r.store.state.listing.RemainingAmount = r.store.state.listing.RemainingAmount.Add(amount)
	}
	return nil
}

// fakeWalletRepo 以 decimal 模擬錢包：FreezeInTx 檢查可用餘額；UnfreezeInTx / TransferInTx
// 與真實實作相同不檢查凍結餘額，讓重複解凍會直接反映成負的凍結餘額。
type fakeWalletRepo struct {
	walletrepo.WalletRepository
	store *fakeStore
}

func (r *fakeWalletRepo) AcquireLocks(_ context.Context, _ []int64, _ string) (func(), error) {
	return func() {}, nil
}

func (r *fakeWalletRepo) FreezeInTx(_ context.Context, session sqlx.Session, userID int64, currency string, amount string) error {
	if err := r.store.requireTx(session); err != nil {
		return err
	}
	if err := r.store.step("freeze"); err != nil {
		return err
	}
	r.store.freezeAmounts = append(r.store.freezeAmounts, amount)
	amt, err := decimal.NewFromString(amount)
	if err != nil || !amt.IsPositive() {
		return apierrors.New(400, "invalid amount")
	}
	k := walletKey{userID, currency}
	b, ok := r.store.state.wallets[k]
	if !ok {
		return apierrors.New(400, "wallet not found for currency "+currency)
	}
	if b.available.LessThan(amt) {
		return apierrors.New(400, "insufficient balance")
	}
	b.available = b.available.Sub(amt)
	b.frozen = b.frozen.Add(amt)
	r.store.state.wallets[k] = b
	return nil
}

func (r *fakeWalletRepo) UnfreezeInTx(_ context.Context, session sqlx.Session, userID int64, currency string, amount decimal.Decimal) error {
	if err := r.store.requireTx(session); err != nil {
		return err
	}
	if err := r.store.step("unfreeze"); err != nil {
		return err
	}
	r.store.unfreezeAmounts = append(r.store.unfreezeAmounts, amount)
	k := walletKey{userID, currency}
	b := r.store.state.wallets[k]
	b.frozen = b.frozen.Sub(amount)
	b.available = b.available.Add(amount)
	r.store.state.wallets[k] = b
	return nil
}

func (r *fakeWalletRepo) TransferInTx(_ context.Context, session sqlx.Session, sellerID, buyerID int64, currency string, amount decimal.Decimal, _ string) error {
	if err := r.store.requireTx(session); err != nil {
		return err
	}
	if err := r.store.step("transfer"); err != nil {
		return err
	}
	sk, bk := walletKey{sellerID, currency}, walletKey{buyerID, currency}
	s := r.store.state.wallets[sk]
	s.frozen = s.frozen.Sub(amount)
	r.store.state.wallets[sk] = s
	b := r.store.state.wallets[bk]
	b.available = b.available.Add(amount)
	r.store.state.wallets[bk] = b
	return nil
}

type fakeEscrowRepo struct {
	escrowrepo.EscrowRepository
	store *fakeStore
}

func (r *fakeEscrowRepo) CreateInTx(_ context.Context, session sqlx.Session, rec *escrowrepo.EscrowRecord) error {
	if err := r.store.requireTx(session); err != nil {
		return err
	}
	if err := r.store.step("escrow"); err != nil {
		return err
	}
	r.store.state.escrows = append(r.store.state.escrows, *rec)
	return nil
}

type fakeStatusLogRepo struct {
	orderstatuslogrepo.OrderStatusLogRepository
	store *fakeStore
}

func (r *fakeStatusLogRepo) Append(_ context.Context, log *orderstatuslogrepo.OrderStatusLog) error {
	r.store.calls = append(r.store.calls, "statusLog")
	r.store.state.logs = append(r.store.state.logs, *log)
	return nil
}

func (r *fakeStatusLogRepo) AppendInTx(_ context.Context, session sqlx.Session, log *orderstatuslogrepo.OrderStatusLog) error {
	if err := r.store.requireTx(session); err != nil {
		return err
	}
	if err := r.store.step("statusLog"); err != nil {
		return err
	}
	r.store.state.logs = append(r.store.state.logs, *log)
	return nil
}

// fakePaymentRepo 走連線池（非交易 session）；在交易內被呼叫視為錯誤（會與交易搶連線池）。
type fakePaymentRepo struct {
	paymentrepo.PaymentMethodRepository
	store *fakeStore
	none  bool
}

func (r *fakePaymentRepo) FindByUserID(_ context.Context, userID int64) ([]*entity.PaymentMethod, error) {
	if r.store.inTx {
		return nil, errors.New("payment method lookup must not run inside the transaction")
	}
	if r.none {
		return nil, nil
	}
	return []*entity.PaymentMethod{{ID: 9, UserID: userID}}, nil
}

// fakePublisher 記錄每次發布時是否仍在交易內，用來驗證 NATS 只在交易提交後發布。
type fakePublisher struct {
	store *fakeStore
}

func (p fakePublisher) Publish(_ context.Context, _ string, _ []byte) error {
	p.store.publishLog = append(p.store.publishLog, p.store.inTx)
	return nil
}

func (p fakePublisher) PublishAsync(_ string, _ []byte) {
	p.store.publishLog = append(p.store.publishLog, p.store.inTx)
}

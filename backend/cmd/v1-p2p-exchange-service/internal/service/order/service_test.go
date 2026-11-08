package order_service

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"

	app_interface "p2p-exchange/cmd/v1-p2p-exchange-service/internal/interfaces/app"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/model/entity"
	escrowrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/escrow_record"
	listingrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/listing"
	orderrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/order"
	orderstatuslogrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/order_status_log"
	walletrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/wallet"
	apierrors "p2p-exchange/internal/errors"
)

const (
	testListingID   int64 = 10
	testOrderID     int64 = 20
	testSellerID    int64 = 1001
	testBuyerID     int64 = 2002
	testCryptoValue       = 10.0
)

// fakeConn 只實作 TransactCtx（直接執行 fn），其餘方法沿用內嵌的 nil 介面：
// 測試若走到未預期的路徑會直接 panic，不會靜默通過。
type fakeConn struct {
	sqlx.SqlConn
}

func (fakeConn) TransactCtx(ctx context.Context, fn func(context.Context, sqlx.Session) error) error {
	return fn(ctx, fakeSession{})
}

type fakeSession struct {
	sqlx.Session
}

type statusUpdate struct {
	from, to string
}

// fakeOrderRepo 以 rowsAffected 模擬條件式 UPDATE 的結果：0 代表訂單狀態已被其他請求改變。
type fakeOrderRepo struct {
	orderrepo.OrderRepository

	order        *entity.Order
	rowsAffected int64
	updates      []statusUpdate
	createCalls  int
}

func (r *fakeOrderRepo) FindByID(_ context.Context, _ int64) (*entity.Order, error) {
	o := *r.order
	return &o, nil
}

func (r *fakeOrderRepo) UpdateStatus(_ context.Context, _ int64, fromStatus, toStatus string, _ map[string]interface{}) (int64, error) {
	r.updates = append(r.updates, statusUpdate{from: fromStatus, to: toStatus})
	return r.rowsAffected, nil
}

func (r *fakeOrderRepo) UpdateStatusInTx(_ context.Context, _ sqlx.Session, _ int64, fromStatus, toStatus string, _ map[string]interface{}) (int64, error) {
	r.updates = append(r.updates, statusUpdate{from: fromStatus, to: toStatus})
	return r.rowsAffected, nil
}

func (r *fakeOrderRepo) Create(_ context.Context, _ *entity.Order) (int64, error) {
	r.createCalls++
	return testOrderID, nil
}

func (r *fakeOrderRepo) DeductListingAmount(_ context.Context, _ int64, _ float64) error {
	return nil
}

type fakeListingRepo struct {
	listingrepo.ListingRepository
	listing *entity.Listing
}

func (r *fakeListingRepo) FindByID(_ context.Context, _ int64) (*entity.Listing, error) {
	l := *r.listing
	return &l, nil
}

// fakeWalletRepo 記錄轉帳與解凍次數，用來驗證「狀態轉換失敗時完全不動錢包」。
type fakeWalletRepo struct {
	walletrepo.WalletRepository
	transferCalls int
	unfreezeCalls int
}

func (r *fakeWalletRepo) AcquireLocks(_ context.Context, _ []int64, _ string) (func(), error) {
	return func() {}, nil
}

func (r *fakeWalletRepo) TransferInTx(_ context.Context, _ sqlx.Session, _, _ int64, _ string, _ float64, _ string) error {
	r.transferCalls++
	return nil
}

func (r *fakeWalletRepo) UnfreezeInTx(_ context.Context, _ sqlx.Session, _ int64, _ string, _ float64) error {
	r.unfreezeCalls++
	return nil
}

func (r *fakeWalletRepo) movements() int {
	return r.transferCalls + r.unfreezeCalls
}

type fakeEscrowRepo struct {
	escrowrepo.EscrowRepository
}

func (fakeEscrowRepo) Create(_ context.Context, _ *escrowrepo.EscrowRecord) error {
	return nil
}

type fakeStatusLogRepo struct {
	orderstatuslogrepo.OrderStatusLogRepository
	appendCalls int
}

func (r *fakeStatusLogRepo) Append(_ context.Context, _ *orderstatuslogrepo.OrderStatusLog) error {
	r.appendCalls++
	return nil
}

type testDeps struct {
	orders    *fakeOrderRepo
	listings  *fakeListingRepo
	wallet    *fakeWalletRepo
	statusLog *fakeStatusLogRepo
	svc       OrderService
}

func newTestDeps(order *entity.Order, listing *entity.Listing, rowsAffected int64) *testDeps {
	d := &testDeps{
		orders:    &fakeOrderRepo{order: order, rowsAffected: rowsAffected},
		listings:  &fakeListingRepo{listing: listing},
		wallet:    &fakeWalletRepo{},
		statusLog: &fakeStatusLogRepo{},
	}
	d.svc = New(Params{
		DB:            fakeConn{},
		OrderRepo:     d.orders,
		ListingRepo:   d.listings,
		WalletRepo:    d.wallet,
		EscrowRepo:    fakeEscrowRepo{},
		StatusLogRepo: d.statusLog,
	})
	return d
}

func activeSellListing() *entity.Listing {
	paymentMethodID := int64(5)
	return &entity.Listing{
		ID:               testListingID,
		UserID:           testSellerID,
		Type:             "sell",
		CryptoCurrency:   "USDT",
		FiatCurrency:     "TWD",
		TotalAmount:      100,
		RemainingAmount:  100,
		Price:            32,
		MinOrderFiat:     0,
		MaxOrderFiat:     100000,
		PaymentTimeLimit: 15,
		PaymentMethodID:  &paymentMethodID,
		Status:           "active",
	}
}

func orderInStatus(status string) *entity.Order {
	return &entity.Order{
		ID:              testOrderID,
		OrderNo:         "P2P20261108000001",
		ListingID:       testListingID,
		ListingType:     "sell",
		SellerID:        testSellerID,
		BuyerID:         testBuyerID,
		CryptoCurrency:  "USDT",
		CryptoAmount:    testCryptoValue,
		Status:          status,
		PaymentDeadline: time.Now().Add(15 * time.Minute),
	}
}

func errCode(t *testing.T, err error) int {
	t.Helper()
	var appErr *apierrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected *apierrors.AppError, got %T: %v", err, err)
	}
	return appErr.Code
}

// TestCreateOrderRejectsInvalidCryptoAmount 迴歸測試：cryptoAmount 必須是大於零的有限值。
// 負數 cryptoAmount 會讓 `remaining_amount < cryptoAmount` 的檢查形同虛設，
// 後續 DeductListingAmount 反而增加掛單剩餘量；NaN 會讓所有比較皆為 false 而通過上下限檢查。
func TestCreateOrderRejectsInvalidCryptoAmount(t *testing.T) {
	for _, amount := range []float64{-1, -0.000001, 0, math.NaN(), math.Inf(1), math.Inf(-1)} {
		d := newTestDeps(nil, activeSellListing(), 1)

		_, err := d.svc.Create(context.Background(), testBuyerID, app_interface.CreateOrderRequest{
			ListingID:    testListingID,
			CryptoAmount: amount,
		})
		if code := errCode(t, err); code != 400 {
			t.Errorf("cryptoAmount=%v: expected 400, got %d (%v)", amount, code, err)
		}
		if d.orders.createCalls != 0 {
			t.Errorf("cryptoAmount=%v: order must not be created, got %d", amount, d.orders.createCalls)
		}
	}
}

// TestCreateOrderRejectsOwnListing 迴歸測試：掛單者不得吃自己的單（單一帳號自行造幣的路徑）。
func TestCreateOrderRejectsOwnListing(t *testing.T) {
	d := newTestDeps(nil, activeSellListing(), 1)

	_, err := d.svc.Create(context.Background(), testSellerID, app_interface.CreateOrderRequest{
		ListingID:    testListingID,
		CryptoAmount: testCryptoValue,
	})
	if code := errCode(t, err); code != 400 {
		t.Fatalf("expected 400, got %d (%v)", code, err)
	}
	if err.Error() != "cannot take own listing" {
		t.Fatalf("unexpected message: %q", err.Error())
	}
	if d.orders.createCalls != 0 {
		t.Fatalf("order must not be created, got %d", d.orders.createCalls)
	}
}

// TestCreateOrderRejectsBuyListingTemporarily 驗證臨時開關：對買幣掛單下單一律回 503 且不建立訂單。
// 此為臨時行為，H2 修正買幣掛單的凍結流程後必須連同 service.go 的判斷一併移除。
func TestCreateOrderRejectsBuyListingTemporarily(t *testing.T) {
	listing := activeSellListing()
	listing.Type = "buy"
	listing.UserID = testBuyerID
	listing.PaymentMethodID = nil
	d := newTestDeps(nil, listing, 1)

	_, err := d.svc.Create(context.Background(), testSellerID, app_interface.CreateOrderRequest{
		ListingID:    testListingID,
		CryptoAmount: testCryptoValue,
	})
	if code := errCode(t, err); code != 503 {
		t.Fatalf("expected 503, got %d (%v)", code, err)
	}
	if err.Error() != "buy listings are temporarily unavailable" {
		t.Fatalf("unexpected message: %q", err.Error())
	}
	if d.orders.createCalls != 0 {
		t.Fatalf("order must not be created, got %d", d.orders.createCalls)
	}
}

// TestCreateOrderAcceptsValidInput 合法輸入：其他使用者以正數金額吃賣幣掛單，成功建立一筆訂單。
func TestCreateOrderAcceptsValidInput(t *testing.T) {
	d := newTestDeps(nil, activeSellListing(), 1)

	resp, err := d.svc.Create(context.Background(), testBuyerID, app_interface.CreateOrderRequest{
		ListingID:    testListingID,
		CryptoAmount: testCryptoValue,
	})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if resp == nil || resp.ID != testOrderID {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if d.orders.createCalls != 1 {
		t.Fatalf("expected exactly 1 order create, got %d", d.orders.createCalls)
	}
}

// statusTransitionCase 描述一個使用者端的狀態轉換操作與其前置狀態。
type statusTransitionCase struct {
	name       string
	fromStatus string
	toStatus   string
	// wantMovements 為狀態轉換成功（影響 1 列）時應發生的錢包異動次數。
	wantMovements int
	run           func(svc OrderService) error
}

func statusTransitionCases() []statusTransitionCase {
	ctx := context.Background()
	return []statusTransitionCase{
		{
			name: "Pay", fromStatus: "matched", toStatus: "paid", wantMovements: 0,
			run: func(svc OrderService) error { return svc.Pay(ctx, testBuyerID, testOrderID) },
		},
		{
			name: "Confirm", fromStatus: "paid", toStatus: "completed", wantMovements: 1,
			run: func(svc OrderService) error { return svc.Confirm(ctx, testSellerID, testOrderID) },
		},
		{
			name: "Cancel", fromStatus: "matched", toStatus: "cancelled", wantMovements: 1,
			run: func(svc OrderService) error {
				return svc.Cancel(ctx, testBuyerID, app_interface.CancelOrderRequest{ID: testOrderID, Reason: "changed mind"})
			},
		},
		{
			name: "Dispute", fromStatus: "paid", toStatus: "disputed", wantMovements: 0,
			run: func(svc OrderService) error {
				return svc.Dispute(ctx, testBuyerID, app_interface.DisputeOrderRequest{ID: testOrderID, Reason: "not received"})
			},
		},
	}
}

// TestStatusTransitionConflictSkipsWalletOps 迴歸測試：條件式狀態轉換影響 0 列時必須回 409，
// 且不得執行任何錢包操作、不得寫狀態紀錄。
//
// 重現場景：同一筆 paid 訂單同時收到賣方 Confirm 與另一個請求（重送、或 Dispute 後的後台裁決）。
// legacy 的 UPDATE 只比對 id，兩個請求在 FindByID 都看到舊狀態後各自成功改寫，
// 並各自執行一次轉帳或解凍，造成同一筆託管資金被重複放出。
func TestStatusTransitionConflictSkipsWalletOps(t *testing.T) {
	for _, tc := range statusTransitionCases() {
		t.Run(tc.name, func(t *testing.T) {
			d := newTestDeps(orderInStatus(tc.fromStatus), nil, 0)

			err := tc.run(d.svc)
			if code := errCode(t, err); code != 409 {
				t.Fatalf("expected 409, got %d (%v)", code, err)
			}
			if err.Error() != "order status changed" {
				t.Fatalf("unexpected message: %q", err.Error())
			}
			if len(d.orders.updates) != 1 || d.orders.updates[0] != (statusUpdate{from: tc.fromStatus, to: tc.toStatus}) {
				t.Fatalf("expected one conditional update %s -> %s, got %v", tc.fromStatus, tc.toStatus, d.orders.updates)
			}
			if got := d.wallet.movements(); got != 0 {
				t.Fatalf("wallet must not be touched on conflict, got transfer=%d unfreeze=%d", d.wallet.transferCalls, d.wallet.unfreezeCalls)
			}
			if d.statusLog.appendCalls != 0 {
				t.Fatalf("status log must not be appended on conflict, got %d", d.statusLog.appendCalls)
			}
		})
	}
}

// TestStatusTransitionSucceedsWhenRowUpdated 對照組：影響 1 列時操作成功，錢包異動與狀態紀錄各一次。
func TestStatusTransitionSucceedsWhenRowUpdated(t *testing.T) {
	for _, tc := range statusTransitionCases() {
		t.Run(tc.name, func(t *testing.T) {
			d := newTestDeps(orderInStatus(tc.fromStatus), nil, 1)

			if err := tc.run(d.svc); err != nil {
				t.Fatalf("expected success, got %v", err)
			}
			if got := d.wallet.movements(); got != tc.wantMovements {
				t.Fatalf("expected %d wallet movements, got %d", tc.wantMovements, got)
			}
			if d.statusLog.appendCalls != 1 {
				t.Fatalf("expected exactly 1 status log, got %d", d.statusLog.appendCalls)
			}
		})
	}
}

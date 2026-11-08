package order_service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"reflect"
	"testing"
	"time"

	app_interface "p2p-exchange/cmd/v1-p2p-exchange-service/internal/interfaces/app"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/model/entity"
	apierrors "p2p-exchange/internal/errors"
)

const (
	testListingID   int64 = 10
	testOrderID     int64 = 20
	testSellerID    int64 = 1001
	testBuyerID     int64 = 2002
	testCryptoValue       = 10.0
	testListingSize       = 100.0
)

type testDeps struct {
	store   *fakeStore
	svc     OrderService
	expired ExpiredOrderConsumerDeps
}

func newTestDeps(listing entity.Listing) *testDeps {
	st := newFakeStore(listing)
	conn := fakeConn{store: st}
	orders := &fakeOrderRepo{store: st}
	listings := &fakeListingRepo{store: st}
	wallets := &fakeWalletRepo{store: st}
	logs := &fakeStatusLogRepo{store: st}
	return &testDeps{
		store: st,
		svc: New(Params{
			DB:            conn,
			MQPublisher:   fakePublisher{store: st},
			OrderRepo:     orders,
			ListingRepo:   listings,
			PaymentRepo:   fakePaymentRepo{},
			WalletRepo:    wallets,
			EscrowRepo:    &fakeEscrowRepo{store: st},
			StatusLogRepo: logs,
		}),
		expired: ExpiredOrderConsumerDeps{
			DB:            conn,
			OrderRepo:     orders,
			ListingRepo:   listings,
			WalletRepo:    wallets,
			StatusLogRepo: logs,
		},
	}
}

func baseListing(listingType string, ownerID int64) entity.Listing {
	return entity.Listing{
		ID:               testListingID,
		UserID:           ownerID,
		Type:             listingType,
		CryptoCurrency:   "USDT",
		FiatCurrency:     "TWD",
		TotalAmount:      testListingSize,
		RemainingAmount:  testListingSize,
		Price:            32,
		MinOrderFiat:     0,
		MaxOrderFiat:     100000,
		PaymentTimeLimit: 15,
		Status:           "active",
	}
}

// newSellDeps：賣方掛 100 USDT 的 sell 掛單，掛單時已凍結 100。
func newSellDeps() *testDeps {
	l := baseListing("sell", testSellerID)
	paymentMethodID := int64(5)
	l.PaymentMethodID = &paymentMethodID
	d := newTestDeps(l)
	d.store.setWallet(testSellerID, 0, testListingSize)
	d.store.setWallet(testBuyerID, 0, 0)
	return d
}

// newBuyDeps：買方掛 buy 掛單（不凍結），吃單的賣方可用餘額 100。
func newBuyDeps() *testDeps {
	d := newTestDeps(baseListing("buy", testBuyerID))
	d.store.setWallet(testSellerID, testListingSize, 0)
	d.store.setWallet(testBuyerID, 0, 0)
	return d
}

func newDepsFor(listingType string) *testDeps {
	if listingType == "buy" {
		return newBuyDeps()
	}
	return newSellDeps()
}

// takerOf 回傳吃單者：sell 掛單由買方吃單，buy 掛單由賣方吃單。
func takerOf(listingType string) int64 {
	if listingType == "buy" {
		return testSellerID
	}
	return testBuyerID
}

func (d *testDeps) create(uid int64, amount float64) (*app_interface.CreateOrderResponse, error) {
	return d.svc.Create(context.Background(), uid, app_interface.CreateOrderRequest{
		ListingID:    testListingID,
		CryptoAmount: amount,
	})
}

func errCode(t *testing.T, err error) int {
	t.Helper()
	var appErr *apierrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected *apierrors.AppError, got %T: %v", err, err)
	}
	return appErr.Code
}

func assertWallet(t *testing.T, d *testDeps, userID int64, available, frozen float64) {
	t.Helper()
	if got := d.store.wallet(userID); got != (balance{available: available, frozen: frozen}) {
		t.Fatalf("wallet of %d = %+v, want available=%v frozen=%v", userID, got, available, frozen)
	}
}

func assertRemaining(t *testing.T, d *testDeps, want float64) {
	t.Helper()
	if got := d.store.state.listing.RemainingAmount; got != want {
		t.Fatalf("listing remaining = %v, want %v", got, want)
	}
}

// TestCreateOrderRejectsInvalidCryptoAmount 迴歸測試：cryptoAmount 必須是大於零的有限值。
// 負數 cryptoAmount 會讓 `remaining_amount < cryptoAmount` 的檢查形同虛設，
// 扣量反而增加掛單剩餘量；NaN 會讓所有比較皆為 false 而通過上下限檢查。
func TestCreateOrderRejectsInvalidCryptoAmount(t *testing.T) {
	for _, amount := range []float64{-1, -0.000001, 0, math.NaN(), math.Inf(1), math.Inf(-1)} {
		d := newSellDeps()

		_, err := d.create(testBuyerID, amount)
		if code := errCode(t, err); code != 400 {
			t.Errorf("cryptoAmount=%v: expected 400, got %d (%v)", amount, code, err)
		}
		if len(d.store.state.orders) != 0 || len(d.store.calls) != 0 {
			t.Errorf("cryptoAmount=%v: must be rejected before touching the database, calls=%v", amount, d.store.calls)
		}
	}
}

// TestCreateOrderRejectsUnstorableCryptoAmount 迴歸測試（S8）：低於 1e-18，或以 18 位小數格式化後
// 不再是原值的數量必須拒絕；否則訂單量、凍結量與釋放量會被各自捨入，甚至凍結 0 卻成立非 0 的訂單。
func TestCreateOrderRejectsUnstorableCryptoAmount(t *testing.T) {
	for _, amount := range []float64{1e-19, 5e-19, 1.5e-18} {
		for _, listingType := range []string{"sell", "buy"} {
			d := newDepsFor(listingType)
			before := d.store.state.clone()

			_, err := d.create(takerOf(listingType), amount)
			if code := errCode(t, err); code != 400 {
				t.Fatalf("%s cryptoAmount=%v: expected 400, got %d (%v)", listingType, amount, code, err)
			}
			if !reflect.DeepEqual(d.store.state, before) {
				t.Fatalf("%s cryptoAmount=%v: state must not change", listingType, amount)
			}
			if d.store.hasCall("freeze") || d.store.hasCall("createOrder") {
				t.Fatalf("%s cryptoAmount=%v: must be rejected before freeze/create, calls=%v", listingType, amount, d.store.calls)
			}
		}
	}

	for _, amount := range []float64{1e-18, 2.5e-17, 0.1, 0.3, 12.345678} {
		d := newBuyDeps()
		if _, err := d.create(testSellerID, amount); err != nil {
			t.Fatalf("cryptoAmount=%v: expected success, got %v", amount, err)
		}
	}
}

// TestCreateOrderRejectsOwnListing 迴歸測試：掛單者不得吃自己的單（單一帳號自行造幣的路徑）。
func TestCreateOrderRejectsOwnListing(t *testing.T) {
	d := newSellDeps()

	_, err := d.create(testSellerID, testCryptoValue)
	if code := errCode(t, err); code != 400 {
		t.Fatalf("expected 400, got %d (%v)", code, err)
	}
	if err.Error() != "cannot take own listing" {
		t.Fatalf("unexpected message: %q", err.Error())
	}
	if len(d.store.state.orders) != 0 {
		t.Fatalf("order must not be created, got %d", len(d.store.state.orders))
	}
}

// TestCreateOrderRevalidatesInsideLock 迴歸測試：掛單狀態與剩餘量必須以 FOR UPDATE 鎖內讀到的值驗證。
// legacy 在交易外讀掛單後才建單，掛單被取消或被並發吃單吃光後仍能成立訂單（超賣）。
func TestCreateOrderRevalidatesInsideLock(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(*entity.Listing)
		wantMsg string
	}{
		{name: "cancelled listing", mutate: func(l *entity.Listing) { l.Status = "cancelled" }, wantMsg: "listing is not active"},
		{name: "paused listing", mutate: func(l *entity.Listing) { l.Status = "paused" }, wantMsg: "listing is not active"},
		{name: "insufficient remaining", mutate: func(l *entity.Listing) { l.RemainingAmount = 5 }, wantMsg: "insufficient remaining amount in listing"},
		{name: "below minimum", mutate: func(l *entity.Listing) { l.MinOrderFiat = 1000 }, wantMsg: "order amount below minimum"},
		{name: "above maximum", mutate: func(l *entity.Listing) { l.MaxOrderFiat = 100 }, wantMsg: "order amount above maximum"},
	}
	for _, listingType := range []string{"sell", "buy"} {
		for _, tc := range cases {
			t.Run(listingType+"/"+tc.name, func(t *testing.T) {
				d := newDepsFor(listingType)
				tc.mutate(&d.store.state.listing)
				before := d.store.state.clone()

				_, err := d.create(takerOf(listingType), testCryptoValue)
				if code := errCode(t, err); code != 400 || err.Error() != tc.wantMsg {
					t.Fatalf("expected 400 %q, got %d %q", tc.wantMsg, code, err.Error())
				}
				if len(d.store.calls) == 0 || d.store.calls[0] != "lockListing" {
					t.Fatalf("listing must be read with FOR UPDATE first, calls=%v", d.store.calls)
				}
				if !reflect.DeepEqual(d.store.state, before) {
					t.Fatalf("state must not change")
				}
			})
		}
	}
}

// TestCreateOrderSellListingDoesNotFreeze 吃 sell 掛單：額度已在掛單時凍結，建單不動錢包，只扣減剩餘量；
// 所有寫入在同一交易內依序完成，NATS 在提交後才發布。
func TestCreateOrderSellListingDoesNotFreeze(t *testing.T) {
	d := newSellDeps()

	resp, err := d.create(testBuyerID, testCryptoValue)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if resp == nil || resp.ID != testOrderID || resp.OrderNo == "" {
		t.Fatalf("unexpected response: %+v", resp)
	}
	wantCalls := []string{"lockListing", "createOrder", "deductListing", "escrow", "statusLog"}
	if !reflect.DeepEqual(d.store.calls, wantCalls) {
		t.Fatalf("calls = %v, want %v", d.store.calls, wantCalls)
	}
	assertWallet(t, d, testSellerID, 0, testListingSize)
	assertWallet(t, d, testBuyerID, 0, 0)
	assertRemaining(t, d, testListingSize-testCryptoValue)

	o := d.store.state.orders[testOrderID]
	if o.SellerID != testSellerID || o.BuyerID != testBuyerID || o.ListingType != "sell" || o.Status != "matched" || o.PaymentMethodID != 5 {
		t.Fatalf("unexpected order: %+v", o)
	}
	if len(d.store.state.escrows) != 1 || d.store.state.escrows[0].Action != "lock" || d.store.state.escrows[0].Amount != testCryptoValue {
		t.Fatalf("expected one escrow lock record, got %+v", d.store.state.escrows)
	}
	if len(d.store.state.logs) != 1 || d.store.state.logs[0].ToStatus != "matched" {
		t.Fatalf("expected one matched status log, got %+v", d.store.state.logs)
	}
	if len(d.store.publishLog) == 0 {
		t.Fatalf("expected order.status.changed to be published")
	}
	for _, inTx := range d.store.publishLog {
		if inTx {
			t.Fatalf("NATS must be published after commit, not inside the transaction")
		}
	}
}

// TestCreateOrderBuyListingFreezesTaker 迴歸測試：吃 buy 掛單時，吃單者（賣方）在同一交易內凍結 cryptoAmount。
// legacy 完全不凍結，Confirm 卻從賣方 frozen_balance 轉出，凍結餘額被扣成負數、買方憑空收幣。
func TestCreateOrderBuyListingFreezesTaker(t *testing.T) {
	d := newBuyDeps()

	if _, err := d.create(testSellerID, testCryptoValue); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	wantCalls := []string{"lockListing", "freeze", "createOrder", "deductListing", "escrow", "statusLog"}
	if !reflect.DeepEqual(d.store.calls, wantCalls) {
		t.Fatalf("calls = %v, want %v", d.store.calls, wantCalls)
	}
	if want := []string{"10.000000000000000000"}; !reflect.DeepEqual(d.store.freezeAmounts, want) {
		t.Fatalf("freeze amounts = %v, want %v", d.store.freezeAmounts, want)
	}
	assertWallet(t, d, testSellerID, testListingSize-testCryptoValue, testCryptoValue)
	assertWallet(t, d, testBuyerID, 0, 0)
	assertRemaining(t, d, testListingSize-testCryptoValue)

	o := d.store.state.orders[testOrderID]
	if o.SellerID != testSellerID || o.BuyerID != testBuyerID || o.ListingType != "buy" || o.PaymentMethodID != 9 {
		t.Fatalf("unexpected order: %+v", o)
	}
}

// TestCreateOrderBuyListingInsufficientBalance 規格：吃 buy 掛單餘額不足回 400 insufficient balance，
// 不建立訂單，掛單剩餘量不變。
func TestCreateOrderBuyListingInsufficientBalance(t *testing.T) {
	d := newBuyDeps()
	d.store.setWallet(testSellerID, 5, 0)
	before := d.store.state.clone()

	_, err := d.create(testSellerID, testCryptoValue)
	if code := errCode(t, err); code != 400 || err.Error() != "insufficient balance" {
		t.Fatalf("expected 400 insufficient balance, got %d %q", code, err.Error())
	}
	if !reflect.DeepEqual(d.store.state, before) {
		t.Fatalf("state must not change")
	}
	if len(d.store.publishLog) != 0 {
		t.Fatalf("nothing must be published on failure")
	}
}

// TestCreateOrderRollsBackOnAnyStepFailure 迴歸測試：建單任一步失敗都必須整筆回滾，
// 不留下訂單、凍結、扣量、託管紀錄或狀態日誌，並回 500（不外洩資料庫錯誤訊息）。
// legacy 的託管紀錄、狀態日誌與扣量都在交易外執行且忽略錯誤。
func TestCreateOrderRollsBackOnAnyStepFailure(t *testing.T) {
	steps := map[string][]string{
		"sell": {"lockListing", "createOrder", "deductListing", "escrow", "statusLog"},
		"buy":  {"lockListing", "freeze", "createOrder", "deductListing", "escrow", "statusLog"},
	}
	for listingType, names := range steps {
		for _, step := range names {
			t.Run(listingType+"/"+step, func(t *testing.T) {
				d := newDepsFor(listingType)
				d.store.failStep = step
				before := d.store.state.clone()

				_, err := d.create(takerOf(listingType), testCryptoValue)
				if code := errCode(t, err); code != 500 || err.Error() != "internal server error" {
					t.Fatalf("expected 500 internal server error, got %d %q", code, err.Error())
				}
				if d.store.rollbacks != 1 {
					t.Fatalf("expected the transaction to roll back once, got %d", d.store.rollbacks)
				}
				if !reflect.DeepEqual(d.store.state, before) {
					t.Fatalf("state must be fully rolled back")
				}
				if len(d.store.publishLog) != 0 {
					t.Fatalf("nothing must be published on failure")
				}
			})
		}
	}
}

// TestCreateOrderDeductZeroRowsRollsBack 迴歸測試：條件式扣量影響 0 列（掛單已非 active 或剩餘量不足）時
// 回 400 並整筆回滾，buy 掛單已做的凍結一併還原。legacy 忽略扣量結果，訂單照樣成立（超賣）。
func TestCreateOrderDeductZeroRowsRollsBack(t *testing.T) {
	for _, listingType := range []string{"sell", "buy"} {
		t.Run(listingType, func(t *testing.T) {
			d := newDepsFor(listingType)
			zero := int64(0)
			d.store.forceDeductRows = &zero
			before := d.store.state.clone()

			_, err := d.create(takerOf(listingType), testCryptoValue)
			if code := errCode(t, err); code != 400 || err.Error() != "insufficient remaining amount in listing" {
				t.Fatalf("expected 400 insufficient remaining, got %d %q", code, err.Error())
			}
			if !d.store.hasCall("deductListing") {
				t.Fatalf("expected deduct to be attempted, calls=%v", d.store.calls)
			}
			if d.store.rollbacks != 1 || !reflect.DeepEqual(d.store.state, before) {
				t.Fatalf("state must be fully rolled back (rollbacks=%d)", d.store.rollbacks)
			}
		})
	}
}

// TestCreateOrderNotFound 掛單不存在回 404。
func TestCreateOrderNotFound(t *testing.T) {
	d := newSellDeps()
	_, err := d.svc.Create(context.Background(), testBuyerID, app_interface.CreateOrderRequest{ListingID: 999, CryptoAmount: 1})
	if code := errCode(t, err); code != 404 {
		t.Fatalf("expected 404, got %d (%v)", code, err)
	}
}

func seedOrder(d *testDeps, status string) {
	l := d.store.state.listing
	o := entity.Order{
		ID:              testOrderID,
		OrderNo:         "P2P20261108000001",
		ListingID:       l.ID,
		ListingType:     l.Type,
		SellerID:        testSellerID,
		BuyerID:         testBuyerID,
		CryptoCurrency:  "USDT",
		CryptoAmount:    testCryptoValue,
		Status:          status,
		PaymentDeadline: time.Now().Add(15 * time.Minute),
	}
	d.store.state.orders[o.ID] = o
	d.store.state.nextOrderID = o.ID + 1
}

// statusTransitionCase 描述一個使用者端的狀態轉換操作與其前置狀態。
type statusTransitionCase struct {
	name       string
	fromStatus string
	toStatus   string
	run        func(svc OrderService) error
}

func statusTransitionCases() []statusTransitionCase {
	ctx := context.Background()
	return []statusTransitionCase{
		{
			name: "Pay", fromStatus: "matched", toStatus: "paid",
			run: func(svc OrderService) error { return svc.Pay(ctx, testBuyerID, testOrderID) },
		},
		{
			name: "Confirm", fromStatus: "paid", toStatus: "completed",
			run: func(svc OrderService) error { return svc.Confirm(ctx, testSellerID, testOrderID) },
		},
		{
			name: "Cancel", fromStatus: "matched", toStatus: "cancelled",
			run: func(svc OrderService) error {
				return svc.Cancel(ctx, testBuyerID, app_interface.CancelOrderRequest{ID: testOrderID, Reason: "changed mind"})
			},
		},
		{
			name: "Dispute", fromStatus: "paid", toStatus: "disputed",
			run: func(svc OrderService) error {
				return svc.Dispute(ctx, testBuyerID, app_interface.DisputeOrderRequest{ID: testOrderID, Reason: "not received"})
			},
		},
	}
}

// accountingCalls 為所有會異動錢包或掛單剩餘量（或為此鎖住掛單）的呼叫。
var accountingCalls = []string{"lockListing", "freeze", "unfreeze", "transfer", "restoreListing", "deductListing"}

func accountingCallsOf(st *fakeStore) []string {
	var out []string
	for _, c := range st.calls {
		for _, a := range accountingCalls {
			if c == a {
				out = append(out, c)
			}
		}
	}
	return out
}

// TestStatusTransitionConflictSkipsWalletOps 迴歸測試：條件式狀態轉換影響 0 列時必須回 409，
// 且不得執行任何錢包或掛單剩餘量操作（也不得鎖掛單）、不得寫狀態紀錄。
//
// 重現場景：同一筆 paid 訂單同時收到賣方 Confirm 與另一個請求（重送、或 Dispute 後的後台裁決）。
// legacy 的 UPDATE 只比對 id，兩個請求在 FindByID 都看到舊狀態後各自成功改寫，
// 並各自執行一次轉帳或解凍，造成同一筆託管資金被重複放出。
func TestStatusTransitionConflictSkipsWalletOps(t *testing.T) {
	for _, listingType := range []string{"sell", "buy"} {
		for _, tc := range statusTransitionCases() {
			t.Run(listingType+"/"+tc.name, func(t *testing.T) {
				d := newDepsFor(listingType)
				seedOrder(d, tc.fromStatus)
				zero := int64(0)
				d.store.forceOrderRows = &zero
				before := d.store.state.clone()

				err := tc.run(d.svc)
				if code := errCode(t, err); code != 409 {
					t.Fatalf("expected 409, got %d (%v)", code, err)
				}
				if err.Error() != "order status changed" {
					t.Fatalf("unexpected message: %q", err.Error())
				}
				wantCalls := []string{"updateOrder:" + tc.fromStatus + "->" + tc.toStatus}
				if !reflect.DeepEqual(d.store.calls, wantCalls) {
					t.Fatalf("calls = %v, want only %v", d.store.calls, wantCalls)
				}
				if !reflect.DeepEqual(d.store.state, before) {
					t.Fatalf("state must not change on conflict")
				}
			})
		}
	}
}

// TestStatusTransitionSucceedsWhenRowUpdated 對照組：影響 1 列時操作成功，狀態紀錄一次。
func TestStatusTransitionSucceedsWhenRowUpdated(t *testing.T) {
	for _, tc := range statusTransitionCases() {
		t.Run(tc.name, func(t *testing.T) {
			d := newSellDeps()
			seedOrder(d, tc.fromStatus)
			d.store.state.listing.RemainingAmount = testListingSize - testCryptoValue

			if err := tc.run(d.svc); err != nil {
				t.Fatalf("expected success, got %v", err)
			}
			if got := d.store.state.orders[testOrderID].Status; got != tc.toStatus {
				t.Fatalf("order status = %q, want %q", got, tc.toStatus)
			}
			if len(d.store.state.logs) != 1 {
				t.Fatalf("expected exactly 1 status log, got %d", len(d.store.state.logs))
			}
		})
	}
}

// formatAmount 與錢包 repository 內部的金額格式相同。
func formatAmount(v float64) string {
	return fmt.Sprintf("%.18f", v)
}

func (d *testDeps) cancel(uid int64) error {
	return d.svc.Cancel(context.Background(), uid, app_interface.CancelOrderRequest{ID: testOrderID, Reason: "changed mind"})
}

func (d *testDeps) timeout() error {
	o, err := d.expired.OrderRepo.FindByID(context.Background(), testOrderID)
	if err != nil {
		return err
	}
	return cancelExpiredOrder(context.Background(), d.expired, o)
}

func (d *testDeps) payAndConfirm() error {
	if err := d.svc.Pay(context.Background(), testBuyerID, testOrderID); err != nil {
		return err
	}
	return d.svc.Confirm(context.Background(), testSellerID, testOrderID)
}

type walletWant struct {
	sellerAvailable, sellerFrozen, buyerAvailable, buyerFrozen, remaining float64
}

// TestAccountingPaths 依 p2p-order「吃單的凍結規則」驗證 buy/sell × 建單後取消/逾時/確認 的記帳結果，
// 並驗證全部錢包 available + frozen 總和守恆、凍結餘額不為負。（refund 於 backend_admin 驗證。）
//
// 迴歸重點：legacy 對 sell 掛單的取消/逾時會解凍賣方並還原剩餘量，額度被釋放兩次；
// 對 buy 掛單則建單不凍結、Confirm 卻從賣方凍結餘額轉出。
func TestAccountingPaths(t *testing.T) {
	afterCreate := map[string]walletWant{
		"sell": {sellerAvailable: 0, sellerFrozen: 100, buyerAvailable: 0, buyerFrozen: 0, remaining: 90},
		"buy":  {sellerAvailable: 90, sellerFrozen: 10, buyerAvailable: 0, buyerFrozen: 0, remaining: 90},
	}
	cases := []struct {
		path string
		run  func(d *testDeps) error
		want map[string]walletWant
	}{
		{
			path: "cancel by buyer",
			run:  func(d *testDeps) error { return d.cancel(testBuyerID) },
			want: map[string]walletWant{
				"sell": {sellerAvailable: 0, sellerFrozen: 100, remaining: 100},
				"buy":  {sellerAvailable: 100, sellerFrozen: 0, remaining: 100},
			},
		},
		{
			path: "cancel by seller",
			run:  func(d *testDeps) error { return d.cancel(testSellerID) },
			want: map[string]walletWant{
				"sell": {sellerAvailable: 0, sellerFrozen: 100, remaining: 100},
				"buy":  {sellerAvailable: 100, sellerFrozen: 0, remaining: 100},
			},
		},
		{
			path: "timeout",
			run:  func(d *testDeps) error { return d.timeout() },
			want: map[string]walletWant{
				"sell": {sellerAvailable: 0, sellerFrozen: 100, remaining: 100},
				"buy":  {sellerAvailable: 100, sellerFrozen: 0, remaining: 100},
			},
		},
		{
			path: "pay and confirm",
			run:  func(d *testDeps) error { return d.payAndConfirm() },
			want: map[string]walletWant{
				"sell": {sellerAvailable: 0, sellerFrozen: 90, buyerAvailable: 10, remaining: 90},
				"buy":  {sellerAvailable: 90, sellerFrozen: 0, buyerAvailable: 10, remaining: 90},
			},
		},
	}

	for _, listingType := range []string{"sell", "buy"} {
		for _, tc := range cases {
			t.Run(listingType+"/"+tc.path, func(t *testing.T) {
				d := newDepsFor(listingType)
				total := d.store.totalUSDT()

				if _, err := d.create(takerOf(listingType), testCryptoValue); err != nil {
					t.Fatalf("create: %v", err)
				}
				w := afterCreate[listingType]
				assertWallet(t, d, testSellerID, w.sellerAvailable, w.sellerFrozen)
				assertWallet(t, d, testBuyerID, w.buyerAvailable, w.buyerFrozen)
				assertRemaining(t, d, w.remaining)

				if err := tc.run(d); err != nil {
					t.Fatalf("%s: %v", tc.path, err)
				}
				w = tc.want[listingType]
				assertWallet(t, d, testSellerID, w.sellerAvailable, w.sellerFrozen)
				assertWallet(t, d, testBuyerID, w.buyerAvailable, w.buyerFrozen)
				assertRemaining(t, d, w.remaining)
				if got := d.store.totalUSDT(); got != total {
					t.Fatalf("total balance changed from %v to %v", total, got)
				}
			})
		}
	}
}

// TestRepeatedSellCancelThenListingCancelNeverNegative 迴歸測試（F2）：sell 掛單連續 3 輪「吃單 → 取消/逾時」後，
// 賣方錢包不變、掛單剩餘量回到 100；此時取消掛單只會解凍 remaining 100，凍結歸零而不為負。
// legacy 每輪都會先解凍賣方再還原剩餘量，三輪後取消掛單會讓凍結餘額變成 -30。
func TestRepeatedSellCancelThenListingCancelNeverNegative(t *testing.T) {
	d := newSellDeps()
	for i := 0; i < 3; i++ {
		if _, err := d.create(testBuyerID, testCryptoValue); err != nil {
			t.Fatalf("round %d create: %v", i, err)
		}
		o := d.store.state.orders[d.store.state.nextOrderID-1]
		var err error
		if i%2 == 0 {
			err = d.svc.Cancel(context.Background(), testBuyerID, app_interface.CancelOrderRequest{ID: o.ID, Reason: "r"})
		} else {
			err = cancelExpiredOrder(context.Background(), d.expired, &o)
		}
		if err != nil {
			t.Fatalf("round %d release: %v", i, err)
		}
		assertWallet(t, d, testSellerID, 0, testListingSize)
		assertRemaining(t, d, testListingSize)
	}
	if d.store.hasCall("unfreeze") {
		t.Fatalf("sell order release must not unfreeze the seller while the listing is active, calls=%v", d.store.calls)
	}

	// 取消掛單在 listing service 驗證；這裡只確認凍結額度與 remaining 一致，解凍 remaining 後恰好歸零。
	if got := d.store.wallet(testSellerID).frozen - d.store.state.listing.RemainingAmount; got != 0 {
		t.Fatalf("seller frozen must equal listing remaining before listing cancel, diff=%v", got)
	}
}

// TestReleaseAfterSellListingCancelled 舊資料情境：H2 之前可以「先取消 sell 掛單、訂單仍在途」，
// 掛單取消時只解凍了當時的 remaining，在途訂單的額度仍凍結在賣方。取消或逾時這筆訂單時必須解凍賣方，
// 且不得還原已取消掛單的剩餘量。
func TestReleaseAfterSellListingCancelled(t *testing.T) {
	for _, path := range []string{"cancel", "timeout"} {
		t.Run(path, func(t *testing.T) {
			d := newSellDeps()
			d.store.state.listing.Status = "cancelled"
			d.store.state.listing.RemainingAmount = 0
			d.store.setWallet(testSellerID, 90, testCryptoValue)
			seedOrder(d, "matched")

			var err error
			if path == "cancel" {
				err = d.cancel(testBuyerID)
			} else {
				err = d.timeout()
			}
			if err != nil {
				t.Fatalf("expected success, got %v", err)
			}
			assertWallet(t, d, testSellerID, testListingSize, 0)
			assertRemaining(t, d, 0)
		})
	}
}

// TestReleaseLockOrder 驗證取消與逾時的鎖序固定為「訂單列條件式 UPDATE → 掛單列 FOR UPDATE → 錢包/剩餘量」，
// 與建單、取消掛單「先鎖掛單列」不形成相反鎖序；且釋放只在條件式 UPDATE 影響 1 列之後發生。
func TestReleaseLockOrder(t *testing.T) {
	cases := []struct {
		listingType, path string
		want              []string
	}{
		{"buy", "cancel", []string{"updateOrder:matched->cancelled", "lockListing", "unfreeze", "restoreListing"}},
		{"sell", "cancel", []string{"updateOrder:matched->cancelled", "lockListing", "restoreListing"}},
		{"buy", "timeout", []string{"updateOrder:matched->timeout", "lockListing", "unfreeze", "restoreListing", "statusLog"}},
		{"sell", "timeout", []string{"updateOrder:matched->timeout", "lockListing", "restoreListing", "statusLog"}},
	}
	for _, tc := range cases {
		t.Run(tc.listingType+"/"+tc.path, func(t *testing.T) {
			d := newDepsFor(tc.listingType)
			seedOrder(d, "matched")
			if tc.listingType == "buy" {
				d.store.setWallet(testSellerID, 90, testCryptoValue)
			}

			var err error
			if tc.path == "cancel" {
				err = d.cancel(testBuyerID)
			} else {
				err = d.timeout()
			}
			if err != nil {
				t.Fatalf("expected success, got %v", err)
			}
			got := d.store.calls
			if tc.path == "cancel" {
				// 取消的狀態日誌在交易提交後寫入（既有行為），不屬於交易內的鎖序。
				got = got[:len(got)-1]
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("calls = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestTimeoutSkipsWhenOrderNoLongerMatched 與付款或取消競速：逾時的條件式 UPDATE 影響 0 列時直接結束，
// 不鎖掛單、不動錢包與剩餘量、不寫狀態日誌。
func TestTimeoutSkipsWhenOrderNoLongerMatched(t *testing.T) {
	for _, listingType := range []string{"sell", "buy"} {
		t.Run(listingType, func(t *testing.T) {
			d := newDepsFor(listingType)
			seedOrder(d, "paid")
			before := d.store.state.clone()

			if err := d.timeout(); err != nil {
				t.Fatalf("expected nil, got %v", err)
			}
			if want := []string{"updateOrder:matched->timeout"}; !reflect.DeepEqual(d.store.calls, want) {
				t.Fatalf("calls = %v, want %v", d.store.calls, want)
			}
			if !reflect.DeepEqual(d.store.state, before) {
				t.Fatalf("state must not change")
			}
		})
	}
}

// TestReleaseFailureRollsBackStatusChange 釋放任一步失敗時，訂單狀態轉換必須一併回滾，額度不變。
func TestReleaseFailureRollsBackStatusChange(t *testing.T) {
	for _, path := range []string{"cancel", "timeout"} {
		for _, step := range []string{"lockListing", "unfreeze", "restoreListing"} {
			t.Run(path+"/"+step, func(t *testing.T) {
				d := newBuyDeps()
				d.store.setWallet(testSellerID, 90, testCryptoValue)
				seedOrder(d, "matched")
				d.store.failStep = step
				before := d.store.state.clone()

				var err error
				if path == "cancel" {
					err = d.cancel(testBuyerID)
				} else {
					err = d.timeout()
				}
				if !errors.Is(err, errInjected) {
					t.Fatalf("expected injected error, got %v", err)
				}
				if !reflect.DeepEqual(d.store.state, before) {
					t.Fatalf("state must be fully rolled back")
				}
			})
		}
	}
}

// TestFreezeAndReleaseAmountFormatConsistent 迴歸測試（S7）：buy 掛單吃單凍結的金額字串必須與
// 取消/逾時/確認時錢包以 %.18f 格式化的金額完全相同，小數金額（0.1、0.3）才不會留下凍結殘差。
func TestFreezeAndReleaseAmountFormatConsistent(t *testing.T) {
	for _, amount := range []float64{0.1, 0.3, 1.1, 33.333333} {
		for _, path := range []string{"cancel", "timeout"} {
			t.Run(fmt.Sprintf("%v/%s", amount, path), func(t *testing.T) {
				d := newBuyDeps()
				if _, err := d.create(testSellerID, amount); err != nil {
					t.Fatalf("create: %v", err)
				}
				var err error
				if path == "cancel" {
					err = d.cancel(testSellerID)
				} else {
					err = d.timeout()
				}
				if err != nil {
					t.Fatalf("%s: %v", path, err)
				}

				if len(d.store.freezeAmounts) != 1 || len(d.store.unfreezeAmounts) != 1 {
					t.Fatalf("expected one freeze and one unfreeze, got %v / %v", d.store.freezeAmounts, d.store.unfreezeAmounts)
				}
				frozen := d.store.freezeAmounts[0]
				if frozen != formatAmount(amount) {
					t.Fatalf("freeze amount %q, want %q", frozen, formatAmount(amount))
				}
				if released := formatAmount(d.store.unfreezeAmounts[0]); released != frozen {
					t.Fatalf("release amount %q differs from frozen amount %q", released, frozen)
				}
				if got := d.store.wallet(testSellerID).frozen; got != 0 {
					t.Fatalf("seller frozen must return to 0, got %v", got)
				}
			})
		}
	}
}

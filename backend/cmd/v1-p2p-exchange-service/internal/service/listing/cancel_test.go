package listing_service

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/model/entity"
	listingrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/listing"
	orderrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/order"
	walletrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/wallet"
)

const testCancelListingID int64 = 10

// cancelState 是會隨交易回滾的狀態：掛單狀態、剩餘量與掛單者錢包。
type cancelState struct {
	status    string
	remaining float64
	available float64
	frozen    float64
}

// cancelStore 共用狀態與呼叫紀錄；fakeTxConn 在交易失敗時把 state 還原，模擬資料庫回滾。
type cancelStore struct {
	state cancelState
	calls []string

	// preRead 為交易外 FindByID 回傳的掛單（可與鎖內狀態不同，用來驗證鎖內再驗證）。
	preRead entity.Listing
	// locked 為鎖內 FindByIDForUpdateInTx 回傳掛單的基底（狀態與剩餘量取自 state）。
	locked entity.Listing

	hasActiveOrders bool
	// forceCancelRows 不為 nil 時覆寫條件式取消的影響列數。
	forceCancelRows *int64
	unfreezeAmounts []float64
}

type fakeTxConn struct {
	sqlx.SqlConn
	store *cancelStore
}

func (c fakeTxConn) TransactCtx(ctx context.Context, fn func(context.Context, sqlx.Session) error) error {
	snapshot := c.store.state
	err := fn(ctx, cancelSession{})
	if err != nil {
		c.store.state = snapshot
	}
	return err
}

type cancelSession struct {
	sqlx.Session
}

type cancelListingRepo struct {
	listingrepo.ListingRepository
	store *cancelStore
}

func (r *cancelListingRepo) FindByID(_ context.Context, _ int64) (*entity.Listing, error) {
	r.store.calls = append(r.store.calls, "findListing")
	l := r.store.preRead
	return &l, nil
}

func (r *cancelListingRepo) FindByIDForUpdateInTx(_ context.Context, session sqlx.Session, _ int64) (*entity.Listing, error) {
	if _, ok := session.(cancelSession); !ok {
		return nil, errors.New("called outside transaction")
	}
	r.store.calls = append(r.store.calls, "lockListing")
	l := r.store.locked
	l.Status = r.store.state.status
	l.RemainingAmount = r.store.state.remaining
	return &l, nil
}

func (r *cancelListingRepo) CancelInTx(_ context.Context, _ sqlx.Session, _ int64) (int64, error) {
	r.store.calls = append(r.store.calls, "cancelListing")
	if r.store.forceCancelRows != nil {
		return *r.store.forceCancelRows, nil
	}
	if r.store.state.status != "active" && r.store.state.status != "paused" {
		return 0, nil
	}
	r.store.state.status = "cancelled"
	return 1, nil
}

type cancelOrderRepo struct {
	orderrepo.OrderRepository
	store *cancelStore
}

func (r *cancelOrderRepo) HasActiveOrdersByListingInTx(_ context.Context, session sqlx.Session, _ int64) (bool, error) {
	if _, ok := session.(cancelSession); !ok {
		return false, errors.New("called outside transaction")
	}
	r.store.calls = append(r.store.calls, "hasActiveOrders")
	return r.store.hasActiveOrders, nil
}

type cancelWalletRepo struct {
	walletrepo.WalletRepository
	store *cancelStore
}

func (r *cancelWalletRepo) UnfreezeInTx(_ context.Context, _ sqlx.Session, _ int64, _ string, amount float64) error {
	r.store.calls = append(r.store.calls, "unfreeze")
	r.store.unfreezeAmounts = append(r.store.unfreezeAmounts, amount)
	r.store.state.frozen -= amount
	r.store.state.available += amount
	return nil
}

// newCancelDeps：掛單者 testUID 的 sell 掛單，total 100，鎖內 remaining 60，錢包凍結 100（40 屬於在途訂單）。
// 交易外讀到的 remaining 刻意設為過期值 100，用來驗證解凍額度取自鎖內讀值。
func newCancelDeps(listingType string) (*cancelStore, ListingService) {
	base := entity.Listing{ID: testCancelListingID, UserID: testUID, Type: listingType, CryptoCurrency: "USDT", TotalAmount: 100}
	pre := base
	pre.Status, pre.RemainingAmount = "active", 100
	st := &cancelStore{
		state:   cancelState{status: "active", remaining: 60, available: 0, frozen: 100},
		preRead: pre,
		locked:  base,
	}
	svc := New(fakeTxConn{store: st}, nil, &cancelListingRepo{store: st}, &cancelOrderRepo{store: st}, &cancelWalletRepo{store: st})
	return st, svc
}

// TestCancelListingRejectsActiveOrders 迴歸測試（p2p-listing 取消掛單）：掛單仍有非終態訂單
// （matched/paid/releasing/disputed）時回 400 `listing has active orders`，掛單與錢包不變。
// legacy 直接取消並解凍 remaining，在途訂單之後取消/逾時又解凍一次，或 Confirm 從已解凍的額度轉出。
func TestCancelListingRejectsActiveOrders(t *testing.T) {
	for _, listingType := range []string{"sell", "buy"} {
		t.Run(listingType, func(t *testing.T) {
			st, svc := newCancelDeps(listingType)
			st.hasActiveOrders = true
			before := st.state

			err := svc.Cancel(context.Background(), testUID, testCancelListingID)
			if code := errCode(t, err); code != 400 || err.Error() != "listing has active orders" {
				t.Fatalf("expected 400 listing has active orders, got %d %q", code, err.Error())
			}
			if st.state != before {
				t.Fatalf("listing and wallet must not change, got %+v want %+v", st.state, before)
			}
			if want := []string{"findListing", "lockListing", "hasActiveOrders"}; !reflect.DeepEqual(st.calls, want) {
				t.Fatalf("calls = %v, want %v", st.calls, want)
			}
		})
	}
}

// TestCancelListingWithoutActiveOrders 沒有進行中訂單：先鎖掛單再查訂單，條件式取消後以鎖內讀到的
// remaining（60，而非交易外讀到的 100）解凍；buy 掛單不解凍。
func TestCancelListingWithoutActiveOrders(t *testing.T) {
	cases := []struct {
		listingType string
		wantCalls   []string
		wantState   cancelState
	}{
		{
			listingType: "sell",
			wantCalls:   []string{"findListing", "lockListing", "hasActiveOrders", "cancelListing", "unfreeze"},
			wantState:   cancelState{status: "cancelled", remaining: 60, available: 60, frozen: 40},
		},
		{
			listingType: "buy",
			wantCalls:   []string{"findListing", "lockListing", "hasActiveOrders", "cancelListing"},
			wantState:   cancelState{status: "cancelled", remaining: 60, available: 0, frozen: 100},
		},
	}
	for _, tc := range cases {
		t.Run(tc.listingType, func(t *testing.T) {
			st, svc := newCancelDeps(tc.listingType)

			if err := svc.Cancel(context.Background(), testUID, testCancelListingID); err != nil {
				t.Fatalf("expected success, got %v", err)
			}
			if !reflect.DeepEqual(st.calls, tc.wantCalls) {
				t.Fatalf("calls = %v, want %v", st.calls, tc.wantCalls)
			}
			if st.state != tc.wantState {
				t.Fatalf("state = %+v, want %+v", st.state, tc.wantState)
			}
		})
	}
}

// TestCancelListingConditionalUpdateZeroRows 條件式取消影響 0 列（掛單已被其他請求改變狀態）時回 409，
// 不解凍、整筆回滾。
func TestCancelListingConditionalUpdateZeroRows(t *testing.T) {
	st, svc := newCancelDeps("sell")
	zero := int64(0)
	st.forceCancelRows = &zero
	before := st.state

	err := svc.Cancel(context.Background(), testUID, testCancelListingID)
	if code := errCode(t, err); code != 409 {
		t.Fatalf("expected 409, got %d (%v)", code, err)
	}
	if len(st.unfreezeAmounts) != 0 || st.state != before {
		t.Fatalf("must not unfreeze on conflict, unfreeze=%v state=%+v", st.unfreezeAmounts, st.state)
	}
}

// TestCancelListingRevalidatesStatusInsideLock 交易外讀到 active，但鎖內已是 cancelled（並發取消已提交）：
// 回 400 且不查訂單、不解凍，避免同一掛單被解凍兩次。
func TestCancelListingRevalidatesStatusInsideLock(t *testing.T) {
	st, svc := newCancelDeps("sell")
	st.state.status = "cancelled"
	before := st.state

	err := svc.Cancel(context.Background(), testUID, testCancelListingID)
	if code := errCode(t, err); code != 400 || err.Error() != "only active or paused listings can be cancelled" {
		t.Fatalf("expected 400 only active or paused, got %d %q", code, err.Error())
	}
	if want := []string{"findListing", "lockListing"}; !reflect.DeepEqual(st.calls, want) {
		t.Fatalf("calls = %v, want %v", st.calls, want)
	}
	if st.state != before {
		t.Fatalf("state must not change")
	}
}

// TestCancelListingForbiddenForOthers 非掛單者回 403，不進入交易。
func TestCancelListingForbiddenForOthers(t *testing.T) {
	st, svc := newCancelDeps("sell")

	err := svc.Cancel(context.Background(), testUID+1, testCancelListingID)
	if code := errCode(t, err); code != 403 {
		t.Fatalf("expected 403, got %d (%v)", code, err)
	}
	if want := []string{"findListing"}; !reflect.DeepEqual(st.calls, want) {
		t.Fatalf("calls = %v, want %v", st.calls, want)
	}
}

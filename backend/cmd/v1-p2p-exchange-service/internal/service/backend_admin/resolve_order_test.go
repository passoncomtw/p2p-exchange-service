package backend_admin_service

import (
	"context"
	"testing"

	"github.com/zeromicro/go-zero/core/stores/sqlx"

	backend_interface "p2p-exchange/cmd/v1-p2p-exchange-service/internal/interfaces/backend"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/model/entity"
	listingrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/listing"
	orderrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/order"
	walletrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/wallet"
)

const testDisputedOrderID int64 = 77

// resolveOrderRepo 以 rowsAffected 模擬條件式 UPDATE 的結果：0 代表爭議訂單已被其他請求裁決。
type resolveOrderRepo struct {
	orderrepo.OrderRepository

	rowsAffected int64
	fromStatuses []string
}

func (r *resolveOrderRepo) FindByID(_ context.Context, id int64) (*entity.Order, error) {
	return &entity.Order{
		ID:             id,
		OrderNo:        "P2P20261108000077",
		ListingID:      10,
		SellerID:       1001,
		BuyerID:        2002,
		CryptoCurrency: "USDT",
		CryptoAmount:   10,
		Status:         "disputed",
	}, nil
}

func (r *resolveOrderRepo) UpdateStatusInTx(_ context.Context, _ sqlx.Session, _ int64, fromStatus, _ string, _ map[string]interface{}) (int64, error) {
	r.fromStatuses = append(r.fromStatuses, fromStatus)
	return r.rowsAffected, nil
}

// resolveWalletRepo 記錄轉帳與解凍次數，用來驗證「狀態轉換失敗時完全不動錢包」。
type resolveWalletRepo struct {
	walletrepo.WalletRepository
	transferCalls int
	unfreezeCalls int
}

func (r *resolveWalletRepo) AcquireLocks(_ context.Context, _ []int64, _ string) (func(), error) {
	return func() {}, nil
}

func (r *resolveWalletRepo) TransferInTx(_ context.Context, _ sqlx.Session, _, _ int64, _ string, _ float64, _ string) error {
	r.transferCalls++
	return nil
}

func (r *resolveWalletRepo) UnfreezeInTx(_ context.Context, _ sqlx.Session, _ int64, _ string, _ float64) error {
	r.unfreezeCalls++
	return nil
}

type resolveListingRepo struct {
	listingrepo.ListingRepository
	restoreCalls int
}

func (r *resolveListingRepo) RestoreAmountInTx(_ context.Context, _ sqlx.Session, _ int64, _ float64) error {
	r.restoreCalls++
	return nil
}

type resolveDeps struct {
	orders   *resolveOrderRepo
	wallet   *resolveWalletRepo
	listings *resolveListingRepo
	svc      *backendAdminService
}

func newResolveDeps(rowsAffected int64) *resolveDeps {
	d := &resolveDeps{
		orders:   &resolveOrderRepo{rowsAffected: rowsAffected},
		wallet:   &resolveWalletRepo{},
		listings: &resolveListingRepo{},
	}
	d.svc = &backendAdminService{
		db:          fakeConn{},
		orderRepo:   d.orders,
		walletRepo:  d.wallet,
		listingRepo: d.listings,
	}
	return d
}

// TestResolveOrderConflictSkipsWalletOps 迴歸測試：爭議裁決的條件式狀態轉換影響 0 列時必須回 409，
// 且不得轉帳、解凍或還原掛單剩餘量。
//
// 重現場景：兩位後台人員同時裁決同一筆 disputed 訂單（或同一人連點、前端重送）。
// legacy 忽略 UpdateStatusInTx 的影響列數，且 UPDATE 只比對 id，兩個請求都會執行錢包操作；
// 一個 complete 一個 refund 時，賣方被解凍退款的同時買方也收到幣，同一筆託管資金被放出兩次。
func TestResolveOrderConflictSkipsWalletOps(t *testing.T) {
	for _, action := range []string{"complete", "refund"} {
		t.Run(action, func(t *testing.T) {
			d := newResolveDeps(0)

			err := d.svc.ResolveOrder(context.Background(), backend_interface.ResolveOrderRequest{
				ID:     testDisputedOrderID,
				Action: action,
				Reason: "duplicate resolution",
			})
			if code := errCode(t, err); code != 409 {
				t.Fatalf("expected 409, got %d (%v)", code, err)
			}
			if err.Error() != "order status changed" {
				t.Fatalf("unexpected message: %q", err.Error())
			}
			if len(d.orders.fromStatuses) != 1 || d.orders.fromStatuses[0] != "disputed" {
				t.Fatalf("expected one conditional update from disputed, got %v", d.orders.fromStatuses)
			}
			if d.wallet.transferCalls != 0 || d.wallet.unfreezeCalls != 0 {
				t.Fatalf("wallet must not be touched on conflict, got transfer=%d unfreeze=%d", d.wallet.transferCalls, d.wallet.unfreezeCalls)
			}
			if d.listings.restoreCalls != 0 {
				t.Fatalf("listing amount must not be restored on conflict, got %d", d.listings.restoreCalls)
			}
		})
	}
}

// TestResolveOrderSucceedsWhenRowUpdated 對照組：影響 1 列時 complete 轉帳一次、refund 解凍並還原各一次。
func TestResolveOrderSucceedsWhenRowUpdated(t *testing.T) {
	cases := []struct {
		action                                  string
		wantTransfer, wantUnfreeze, wantRestore int
	}{
		{action: "complete", wantTransfer: 1},
		{action: "refund", wantUnfreeze: 1, wantRestore: 1},
	}

	for _, tc := range cases {
		t.Run(tc.action, func(t *testing.T) {
			d := newResolveDeps(1)

			if err := d.svc.ResolveOrder(context.Background(), backend_interface.ResolveOrderRequest{
				ID:     testDisputedOrderID,
				Action: tc.action,
				Reason: "resolved",
			}); err != nil {
				t.Fatalf("expected success, got %v", err)
			}
			if d.wallet.transferCalls != tc.wantTransfer || d.wallet.unfreezeCalls != tc.wantUnfreeze || d.listings.restoreCalls != tc.wantRestore {
				t.Fatalf("unexpected movements: transfer=%d unfreeze=%d restore=%d", d.wallet.transferCalls, d.wallet.unfreezeCalls, d.listings.restoreCalls)
			}
		})
	}
}

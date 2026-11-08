package backend_admin_service

import (
	"context"
	"reflect"
	"testing"

	"github.com/zeromicro/go-zero/core/stores/sqlx"

	backend_interface "p2p-exchange/cmd/v1-p2p-exchange-service/internal/interfaces/backend"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/model/entity"
	listingrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/listing"
	orderrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/order"
	walletrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/wallet"
)

const (
	testDisputedOrderID int64 = 77
	testResolveListing  int64 = 10
	testResolveSeller   int64 = 1001
	testResolveBuyer    int64 = 2002
	testResolveAmount         = 10.0
)

// resolveCall 記錄一次會影響帳務或鎖的呼叫，用來驗證動作與順序。
type resolveCall struct {
	op     string
	userID int64
	amount float64
}

// resolveOrderRepo 以 rowsAffected 模擬條件式 UPDATE 的結果：0 代表爭議訂單已被其他請求裁決。
type resolveOrderRepo struct {
	orderrepo.OrderRepository

	listingType  string
	rowsAffected int64
	fromStatuses []string
	toStatuses   []string
	calls        *[]resolveCall
}

func (r *resolveOrderRepo) FindByID(_ context.Context, id int64) (*entity.Order, error) {
	return &entity.Order{
		ID:             id,
		OrderNo:        "P2P20261108000077",
		ListingID:      testResolveListing,
		ListingType:    r.listingType,
		SellerID:       testResolveSeller,
		BuyerID:        testResolveBuyer,
		CryptoCurrency: "USDT",
		CryptoAmount:   testResolveAmount,
		Status:         "disputed",
	}, nil
}

func (r *resolveOrderRepo) UpdateStatusInTx(_ context.Context, _ sqlx.Session, _ int64, fromStatus, toStatus string, _ map[string]interface{}) (int64, error) {
	r.fromStatuses = append(r.fromStatuses, fromStatus)
	r.toStatuses = append(r.toStatuses, toStatus)
	*r.calls = append(*r.calls, resolveCall{op: "updateOrder"})
	return r.rowsAffected, nil
}

// resolveWalletRepo 記錄轉帳與解凍，用來驗證「狀態轉換失敗時完全不動錢包」與 refund 的記帳分支。
type resolveWalletRepo struct {
	walletrepo.WalletRepository
	calls *[]resolveCall
}

func (r *resolveWalletRepo) AcquireLocks(_ context.Context, _ []int64, _ string) (func(), error) {
	return func() {}, nil
}

func (r *resolveWalletRepo) TransferInTx(_ context.Context, _ sqlx.Session, sellerID, _ int64, _ string, amount float64, _ string) error {
	*r.calls = append(*r.calls, resolveCall{op: "transfer", userID: sellerID, amount: amount})
	return nil
}

func (r *resolveWalletRepo) UnfreezeInTx(_ context.Context, _ sqlx.Session, userID int64, _ string, amount float64) error {
	*r.calls = append(*r.calls, resolveCall{op: "unfreeze", userID: userID, amount: amount})
	return nil
}

// resolveListingRepo 刻意不實作 FindByID：refund 的釋放必須先以 FOR UPDATE 鎖掛單。
type resolveListingRepo struct {
	listingrepo.ListingRepository
	listing entity.Listing
	calls   *[]resolveCall
}

func (r *resolveListingRepo) FindByIDForUpdateInTx(_ context.Context, _ sqlx.Session, _ int64) (*entity.Listing, error) {
	*r.calls = append(*r.calls, resolveCall{op: "lockListing"})
	l := r.listing
	return &l, nil
}

func (r *resolveListingRepo) RestoreAmountInTx(_ context.Context, _ sqlx.Session, _ int64, amount float64) error {
	*r.calls = append(*r.calls, resolveCall{op: "restoreListing", amount: amount})
	return nil
}

type resolveDeps struct {
	orders *resolveOrderRepo
	calls  *[]resolveCall
	svc    *backendAdminService
}

func newResolveDeps(listingType, listingStatus string, rowsAffected int64) *resolveDeps {
	calls := &[]resolveCall{}
	ownerID := testResolveSeller
	if listingType == "buy" {
		ownerID = testResolveBuyer
	}
	d := &resolveDeps{
		orders: &resolveOrderRepo{listingType: listingType, rowsAffected: rowsAffected, calls: calls},
		calls:  calls,
	}
	d.svc = &backendAdminService{
		db:         fakeConn{},
		orderRepo:  d.orders,
		walletRepo: &resolveWalletRepo{calls: calls},
		listingRepo: &resolveListingRepo{
			listing: entity.Listing{ID: testResolveListing, UserID: ownerID, Type: listingType, Status: listingStatus},
			calls:   calls,
		},
	}
	return d
}

// TestResolveOrderConflictSkipsWalletOps 迴歸測試：爭議裁決的條件式狀態轉換影響 0 列時必須回 409，
// 且不得鎖掛單、轉帳、解凍或還原掛單剩餘量。
//
// 重現場景：兩位後台人員同時裁決同一筆 disputed 訂單（或同一人連點、前端重送）。
// legacy 忽略 UpdateStatusInTx 的影響列數，且 UPDATE 只比對 id，兩個請求都會執行錢包操作；
// 一個 complete 一個 refund 時，賣方被解凍退款的同時買方也收到幣，同一筆託管資金被放出兩次。
func TestResolveOrderConflictSkipsWalletOps(t *testing.T) {
	for _, listingType := range []string{"sell", "buy"} {
		for _, action := range []string{"complete", "refund"} {
			t.Run(listingType+"/"+action, func(t *testing.T) {
				d := newResolveDeps(listingType, "active", 0)

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
				if want := []resolveCall{{op: "updateOrder"}}; !reflect.DeepEqual(*d.calls, want) {
					t.Fatalf("calls = %+v, want only the conditional update", *d.calls)
				}
			})
		}
	}
}

// TestResolveOrderSucceedsWhenRowUpdated 影響 1 列時依 p2p-dispute 規則記帳：
// complete 一律由訂單賣方的凍結轉給買方；refund 與取消相同，經 orderallocation.ReleaseInTx：
// buy 掛單解凍吃單者並還原剩餘量；sell 掛單 active 只還原剩餘量（額度留在掛單凍結內）；
// sell 掛單已取消（H2 前的舊資料）只解凍賣方。refund 的狀態字串維持 "refunded"。
//
// 迴歸重點：legacy 的 refund 對 sell 掛單同時解凍賣方又還原剩餘量，額度之後隨掛單取消再被解凍一次。
func TestResolveOrderSucceedsWhenRowUpdated(t *testing.T) {
	cases := []struct {
		name, listingType, listingStatus, action, wantTo string
		want                                             []resolveCall
	}{
		{
			name: "sell complete", listingType: "sell", listingStatus: "active", action: "complete", wantTo: "completed",
			want: []resolveCall{{op: "updateOrder"}, {op: "transfer", userID: testResolveSeller, amount: testResolveAmount}},
		},
		{
			name: "buy complete", listingType: "buy", listingStatus: "active", action: "complete", wantTo: "completed",
			want: []resolveCall{{op: "updateOrder"}, {op: "transfer", userID: testResolveSeller, amount: testResolveAmount}},
		},
		{
			name: "buy refund", listingType: "buy", listingStatus: "active", action: "refund", wantTo: "refunded",
			want: []resolveCall{
				{op: "updateOrder"}, {op: "lockListing"},
				{op: "unfreeze", userID: testResolveSeller, amount: testResolveAmount},
				{op: "restoreListing", amount: testResolveAmount},
			},
		},
		{
			name: "sell refund on active listing", listingType: "sell", listingStatus: "active", action: "refund", wantTo: "refunded",
			want: []resolveCall{{op: "updateOrder"}, {op: "lockListing"}, {op: "restoreListing", amount: testResolveAmount}},
		},
		{
			name: "sell refund on cancelled listing", listingType: "sell", listingStatus: "cancelled", action: "refund", wantTo: "refunded",
			want: []resolveCall{{op: "updateOrder"}, {op: "lockListing"}, {op: "unfreeze", userID: testResolveSeller, amount: testResolveAmount}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := newResolveDeps(tc.listingType, tc.listingStatus, 1)

			if err := d.svc.ResolveOrder(context.Background(), backend_interface.ResolveOrderRequest{
				ID:     testDisputedOrderID,
				Action: tc.action,
				Reason: "resolved",
			}); err != nil {
				t.Fatalf("expected success, got %v", err)
			}
			if !reflect.DeepEqual(*d.calls, tc.want) {
				t.Fatalf("calls = %+v, want %+v", *d.calls, tc.want)
			}
			if len(d.orders.toStatuses) != 1 || d.orders.toStatuses[0] != tc.wantTo {
				t.Fatalf("target status = %v, want %q", d.orders.toStatuses, tc.wantTo)
			}
		})
	}
}

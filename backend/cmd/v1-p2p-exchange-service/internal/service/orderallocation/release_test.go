package orderallocation

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/model/entity"
	listingrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/listing"
	walletrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/wallet"
)

const (
	testListingID int64 = 10
	testOrderID   int64 = 20
	testSellerID  int64 = 1001
	testBuyerID   int64 = 2002
	testAmount          = 10.0
)

type movement struct {
	op     string
	userID int64
	amount float64
}

// fakeListingRepo 刻意不實作 FindByID：釋放若改用不加鎖的讀取會直接 panic。
type fakeListingRepo struct {
	listingrepo.ListingRepository
	listing *entity.Listing
	log     *[]movement
}

func (r *fakeListingRepo) FindByIDForUpdateInTx(_ context.Context, _ sqlx.Session, id int64) (*entity.Listing, error) {
	*r.log = append(*r.log, movement{op: "lockListing"})
	if r.listing == nil || r.listing.ID != id {
		return nil, sqlx.ErrNotFound
	}
	l := *r.listing
	return &l, nil
}

func (r *fakeListingRepo) RestoreAmountInTx(_ context.Context, _ sqlx.Session, _ int64, amount float64) error {
	*r.log = append(*r.log, movement{op: "restoreListing", amount: amount})
	return nil
}

type fakeWalletRepo struct {
	walletrepo.WalletRepository
	log *[]movement
}

func (r *fakeWalletRepo) UnfreezeInTx(_ context.Context, _ sqlx.Session, userID int64, _ string, amount float64) error {
	*r.log = append(*r.log, movement{op: "unfreeze", userID: userID, amount: amount})
	return nil
}

type fakeSession struct {
	sqlx.Session
}

func orderFor(listingType string) *entity.Order {
	return &entity.Order{
		ID:             testOrderID,
		ListingID:      testListingID,
		ListingType:    listingType,
		SellerID:       testSellerID,
		BuyerID:        testBuyerID,
		CryptoCurrency: "USDT",
		CryptoAmount:   testAmount,
	}
}

func release(listing *entity.Listing, order *entity.Order) ([]movement, error) {
	var log []movement
	err := ReleaseInTx(context.Background(), fakeSession{},
		&fakeListingRepo{listing: listing, log: &log},
		&fakeWalletRepo{log: &log},
		order,
	)
	return log, err
}

// TestReleaseInTxBranches 依 p2p-order 凍結規則驗證每個分支的記帳動作，且一律先以 FOR UPDATE 鎖掛單。
//
// 迴歸重點：legacy 的取消/逾時/refund 一律「解凍賣方 + 還原剩餘量」。對 sell 掛單而言額度同時回到
// 賣方可用餘額與掛單剩餘量，之後取消掛單又再解凍一次，同一筆額度被釋放兩次（frozen 變負、可重複提領）。
func TestReleaseInTxBranches(t *testing.T) {
	cases := []struct {
		name          string
		listingType   string
		listingStatus string
		want          []movement
	}{
		{
			name: "buy listing unfreezes taker and restores remaining", listingType: "buy", listingStatus: "active",
			want: []movement{{op: "lockListing"}, {op: "unfreeze", userID: testSellerID, amount: testAmount}, {op: "restoreListing", amount: testAmount}},
		},
		{
			name: "buy listing already cancelled still unfreezes taker", listingType: "buy", listingStatus: "cancelled",
			want: []movement{{op: "lockListing"}, {op: "unfreeze", userID: testSellerID, amount: testAmount}, {op: "restoreListing", amount: testAmount}},
		},
		{
			name: "sell active listing only restores remaining", listingType: "sell", listingStatus: "active",
			want: []movement{{op: "lockListing"}, {op: "restoreListing", amount: testAmount}},
		},
		{
			name: "sell paused listing only restores remaining", listingType: "sell", listingStatus: "paused",
			want: []movement{{op: "lockListing"}, {op: "restoreListing", amount: testAmount}},
		},
		{
			name: "sell cancelled listing only unfreezes seller", listingType: "sell", listingStatus: "cancelled",
			want: []movement{{op: "lockListing"}, {op: "unfreeze", userID: testSellerID, amount: testAmount}},
		},
		{
			name: "sell completed listing only unfreezes seller", listingType: "sell", listingStatus: "completed",
			want: []movement{{op: "lockListing"}, {op: "unfreeze", userID: testSellerID, amount: testAmount}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			listing := &entity.Listing{ID: testListingID, UserID: testSellerID, Type: tc.listingType, Status: tc.listingStatus}
			if tc.listingType == "buy" {
				listing.UserID = testBuyerID
			}

			got, err := release(listing, orderFor(tc.listingType))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("movements = %+v, want %+v", got, tc.want)
			}
		})
	}
}

// TestReleaseInTxFailsClosed 不認得的掛單類型或狀態、訂單與掛單類型不一致、掛單不存在時必須回錯誤，
// 且不得做任何錢包或剩餘量異動（呼叫端交易會因此回滾）。
func TestReleaseInTxFailsClosed(t *testing.T) {
	cases := []struct {
		name    string
		listing *entity.Listing
		order   *entity.Order
	}{
		{name: "unknown sell listing status", listing: &entity.Listing{ID: testListingID, Type: "sell", Status: "frozen"}, order: orderFor("sell")},
		{name: "unknown listing type", listing: &entity.Listing{ID: testListingID, Type: "swap", Status: "active"}, order: orderFor("swap")},
		{name: "order and listing type mismatch", listing: &entity.Listing{ID: testListingID, Type: "sell", Status: "active"}, order: orderFor("buy")},
		{name: "listing not found", listing: nil, order: orderFor("sell")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := release(tc.listing, tc.order)
			if err == nil {
				t.Fatalf("expected an error")
			}
			if tc.listing == nil && !errors.Is(err, sqlx.ErrNotFound) {
				t.Fatalf("expected ErrNotFound, got %v", err)
			}
			if want := []movement{{op: "lockListing"}}; !reflect.DeepEqual(got, want) {
				t.Fatalf("movements = %+v, want only the listing lock", got)
			}
		})
	}
}

// TestTransactionIsolationLevelUnchanged 迴歸測試（S6）：建單、釋放與取消掛單的正確性依賴 PostgreSQL 預設的
// READ COMMITTED（取得掛單列鎖後的一般 SELECT 看得到已提交的並發訂單）。程式碼不得自行指定交易隔離級別；
// 改用 REPEATABLE READ / SERIALIZABLE 會讓鎖內查詢讀到交易開始時的快照而漏看並發建立的訂單。
func TestTransactionIsolationLevelUnchanged(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate test file")
	}
	serviceDir := filepath.Dir(file)
	roots := []string{
		// cmd/v1-p2p-exchange-service/internal：repository 與 service。
		filepath.Clean(filepath.Join(serviceDir, "..", "..")),
		// backend/internal：資料庫連線等基礎設施。
		filepath.Clean(filepath.Join(serviceDir, "..", "..", "..", "..", "..", "internal")),
	}
	forbidden := regexp.MustCompile(`(?i)ISOLATION\s+LEVEL|default_transaction_isolation|sql\.TxOptions|sql\.Level(Serializable|RepeatableRead|Snapshot|Linearizable)|BeginTx\(`)

	scanned := 0
	for _, root := range roots {
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			scanned++
			src, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if loc := forbidden.FindIndex(src); loc != nil {
				t.Errorf("%s sets a transaction isolation level: %q", path, src[loc[0]:loc[1]])
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", root, err)
		}
	}
	if scanned == 0 {
		t.Fatalf("no source files scanned under %v", roots)
	}
}

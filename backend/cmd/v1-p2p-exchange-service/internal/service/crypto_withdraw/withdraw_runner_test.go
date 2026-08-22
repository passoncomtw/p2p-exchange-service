package cryptowithdraw_service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/model/entity"
	cryptowithdrawrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/crypto_withdraw"
	walletrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/wallet"
	"p2p-exchange/pkg/tron"
)

// fakeConn 只實作 TransactCtx（直接執行 fn），其餘方法沿用內嵌的 nil 介面：
// 測試若不小心走到未預期的路徑會直接 panic，不會靜默通過。
type fakeConn struct {
	sqlx.SqlConn
}

func (c fakeConn) TransactCtx(ctx context.Context, fn func(context.Context, sqlx.Session) error) error {
	return fn(ctx, fakeSession{})
}

type fakeSession struct {
	sqlx.Session
}

// fakeWithdrawRepo 以互斥鎖保護的 status 欄位模擬 DB 的條件式 UPDATE 語意：
// 只有看到期望狀態的呼叫才會「異動 1 列」，其餘一律 0 列。
type fakeWithdrawRepo struct {
	cryptowithdrawrepo.CryptoWithdrawRepository

	mu     sync.Mutex
	status string

	claimCalls       int
	broadcastedCalls int
	failedCalls      int
	confirmCalls     int

	confirmErr error
}

func (r *fakeWithdrawRepo) ClaimForBroadcastInTx(_ context.Context, _ sqlx.Session, _ int64) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.claimCalls++
	if r.status != "pending" {
		return 0, nil
	}
	r.status = "broadcasting"
	return 1, nil
}

func (r *fakeWithdrawRepo) UpdateBroadcastedInTx(_ context.Context, _ sqlx.Session, _ int64, _ string, _ time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.broadcastedCalls++
	return nil
}

func (r *fakeWithdrawRepo) UpdateFailedInTx(_ context.Context, _ sqlx.Session, _ int64) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.failedCalls++
	if r.status != "broadcasting" {
		return 0, nil
	}
	r.status = "failed"
	return 1, nil
}

func (r *fakeWithdrawRepo) ConfirmInTx(_ context.Context, _ sqlx.Session, _ int64, _ time.Time) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.confirmCalls++
	if r.confirmErr != nil {
		return 0, r.confirmErr
	}
	if r.status != "broadcasting" {
		return 0, nil
	}
	r.status = "confirmed"
	return 1, nil
}

// fakeWalletRepo 記錄扣款與解凍被呼叫幾次，用來驗證「不重複扣款／不重複解凍」。
type fakeWalletRepo struct {
	walletrepo.WalletRepository

	mu            sync.Mutex
	deductCalls   int
	unfreezeCalls int
}

func (r *fakeWalletRepo) DeductFrozenBalanceWithLedgerTypeInTx(_ context.Context, _ sqlx.Session, _ int64, _, _, ledgerType string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if ledgerType != ledgerTypeCryptoWithdraw {
		return errors.New("unexpected ledger type: " + ledgerType)
	}
	r.deductCalls++
	return nil
}

func (r *fakeWalletRepo) UnfreezeBalance(_ context.Context, _ int64, _ string, _ string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.unfreezeCalls++
	return nil
}

// newTestRunner 組出不依賴 DB / Redis / 鏈上節點的 runner；
// broadcast 由各測試自行注入，用來計算「鏈上 API 被呼叫幾次」。
func newTestRunner(
	repo *fakeWithdrawRepo,
	wallet *fakeWalletRepo,
	broadcast func(ctx context.Context, client *tron.Client, w *entity.CryptoWithdrawal) (string, error),
) *TronWithdrawRunner {
	return &TronWithdrawRunner{
		db:           fakeConn{},
		withdrawRepo: repo,
		walletRepo:   wallet,
		broadcast:    broadcast,
	}
}

func testWithdrawal() *entity.CryptoWithdrawal {
	txHash := "0xabc"
	return &entity.CryptoWithdrawal{
		ID:       7,
		UserID:   1001,
		Currency: usdtCurrency,
		Amount:   "25.000000000000000000",
		TxHash:   &txHash,
		Status:   "broadcasting",
	}
}

// TestClaimForBroadcastAllowsSingleWinner 是 P0-5 的迴歸測試。
//
// 重現場景：同一筆 pending 提領被兩個執行緒同時取出並處理（多副本部署、Redis 鎖失效、
// legacy/v1 併行）。legacy 在廣播前只有 Redis 鎖（fail-open、無 ownership token），
// 鎖一失效就會對同一筆提領廣播兩次 —— 熱錢包真實對外多付一次錢且鏈上不可回滾。
// 修正後：條件式 UPDATE（WHERE status='pending'）只讓一個執行緒認領成功，
// 因此鏈上廣播必定只發生一次。
func TestClaimForBroadcastAllowsSingleWinner(t *testing.T) {
	repo := &fakeWithdrawRepo{status: "pending"}
	wallet := &fakeWalletRepo{}

	var mu sync.Mutex
	broadcastCalls := 0
	runner := newTestRunner(repo, wallet, func(_ context.Context, _ *tron.Client, _ *entity.CryptoWithdrawal) (string, error) {
		mu.Lock()
		defer mu.Unlock()
		broadcastCalls++
		return "0xdeadbeef", nil
	})

	w := &entity.CryptoWithdrawal{ID: 7, UserID: 1001, Currency: usdtCurrency, Amount: "25.000000000000000000", Status: "pending"}

	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			runner.processWithdrawal(context.Background(), nil, w)
		}()
	}
	close(start)
	wg.Wait()

	if repo.claimCalls != 2 {
		t.Fatalf("expected 2 claim attempts, got %d", repo.claimCalls)
	}
	if broadcastCalls != 1 {
		t.Fatalf("double broadcast: expected exactly 1 on-chain broadcast, got %d", broadcastCalls)
	}
	if repo.broadcastedCalls != 1 {
		t.Fatalf("expected exactly 1 tx_hash record, got %d", repo.broadcastedCalls)
	}
}

// TestProcessWithdrawalSkipsChainCallWhenAlreadyClaimed 認領不到時必須完全不碰鏈上 API。
func TestProcessWithdrawalSkipsChainCallWhenAlreadyClaimed(t *testing.T) {
	// 記錄已是 broadcasting（其他流程認領過）。
	repo := &fakeWithdrawRepo{status: "broadcasting"}
	wallet := &fakeWalletRepo{}

	broadcastCalls := 0
	runner := newTestRunner(repo, wallet, func(_ context.Context, _ *tron.Client, _ *entity.CryptoWithdrawal) (string, error) {
		broadcastCalls++
		return "0xdeadbeef", nil
	})

	runner.processWithdrawal(context.Background(), nil, testWithdrawal())

	if broadcastCalls != 0 {
		t.Fatalf("expected no on-chain broadcast when claim failed, got %d", broadcastCalls)
	}
	if repo.broadcastedCalls != 0 {
		t.Fatalf("expected no tx_hash record when claim failed, got %d", repo.broadcastedCalls)
	}
}

// TestProcessWithdrawalMarksFailedAndUnfreezes 廣播失敗：記錄轉 failed 並解凍一次。
func TestProcessWithdrawalMarksFailedAndUnfreezes(t *testing.T) {
	repo := &fakeWithdrawRepo{status: "pending"}
	wallet := &fakeWalletRepo{}

	runner := newTestRunner(repo, wallet, func(_ context.Context, _ *tron.Client, _ *entity.CryptoWithdrawal) (string, error) {
		return "", errors.New("tron rpc unreachable")
	})

	w := &entity.CryptoWithdrawal{ID: 7, UserID: 1001, Currency: usdtCurrency, Amount: "25.000000000000000000", Status: "pending"}
	runner.processWithdrawal(context.Background(), nil, w)

	if repo.failedCalls != 1 {
		t.Fatalf("expected 1 UpdateFailed call, got %d", repo.failedCalls)
	}
	if wallet.unfreezeCalls != 1 {
		t.Fatalf("expected exactly 1 unfreeze, got %d", wallet.unfreezeCalls)
	}
	if repo.broadcastedCalls != 0 {
		t.Fatalf("expected no tx_hash record on broadcast failure, got %d", repo.broadcastedCalls)
	}
}

// TestMarkFailedDoesNotUnfreezeTwice 失敗處理被重複觸發時不得重複解凍
// （第二次的條件式 UPDATE 異動 0 列）。
func TestMarkFailedDoesNotUnfreezeTwice(t *testing.T) {
	repo := &fakeWithdrawRepo{status: "broadcasting"}
	wallet := &fakeWalletRepo{}
	runner := newTestRunner(repo, wallet, nil)

	w := testWithdrawal()
	runner.markFailedAndUnfreeze(context.Background(), w)
	runner.markFailedAndUnfreeze(context.Background(), w)

	if wallet.unfreezeCalls != 1 {
		t.Fatalf("double unfreeze: expected 1 unfreeze call, got %d", wallet.unfreezeCalls)
	}
}

// TestConfirmAndDeductDeductsOnceWhenBroadcasting 正常路徑：broadcasting → confirmed 成功轉換時才扣款。
func TestConfirmAndDeductDeductsOnceWhenBroadcasting(t *testing.T) {
	repo := &fakeWithdrawRepo{status: "broadcasting"}
	wallet := &fakeWalletRepo{}
	runner := newTestRunner(repo, wallet, nil)

	if err := runner.confirmAndDeduct(context.Background(), testWithdrawal(), time.Now()); err != nil {
		t.Fatalf("confirmAndDeduct: %v", err)
	}
	if wallet.deductCalls != 1 {
		t.Fatalf("expected exactly 1 deduct, got %d", wallet.deductCalls)
	}
}

// TestConfirmAndDeductSkipsDeductWhenAlreadyConfirmed 是 P0-3 的迴歸測試。
//
// 重現場景：同一筆 broadcasting 記錄的確認流程被觸發兩次（Job 重啟、重複輪詢、legacy/v1 併行）。
// legacy 的 UpdateConfirmed 沒有 WHERE status 守衛，且扣款是交易外的第二趟往返，
// 第二次會重複自凍結餘額扣款；修正後 RowsAffected=0 必須直接冪等結束，完全不呼叫扣款。
func TestConfirmAndDeductSkipsDeductWhenAlreadyConfirmed(t *testing.T) {
	repo := &fakeWithdrawRepo{status: "broadcasting"}
	wallet := &fakeWalletRepo{}
	runner := newTestRunner(repo, wallet, nil)

	w := testWithdrawal()
	if err := runner.confirmAndDeduct(context.Background(), w, time.Now()); err != nil {
		t.Fatalf("first confirmAndDeduct: %v", err)
	}
	// 第二次觸發：狀態已是 confirmed，條件式 UPDATE 不會異動任何列。
	if err := runner.confirmAndDeduct(context.Background(), w, time.Now()); err != nil {
		t.Fatalf("expected idempotent success on repeat, got: %v", err)
	}

	if repo.confirmCalls != 2 {
		t.Fatalf("expected 2 ConfirmInTx calls, got %d", repo.confirmCalls)
	}
	if wallet.deductCalls != 1 {
		t.Fatalf("double deduct: expected 1 deduct call across 2 confirm attempts, got %d", wallet.deductCalls)
	}
}

// TestConfirmAndDeductPropagatesConfirmError 狀態轉換失敗時不得扣款，錯誤要往上拋讓交易回滾。
func TestConfirmAndDeductPropagatesConfirmError(t *testing.T) {
	confirmErr := errors.New("update failed")
	repo := &fakeWithdrawRepo{status: "broadcasting", confirmErr: confirmErr}
	wallet := &fakeWalletRepo{}
	runner := newTestRunner(repo, wallet, nil)

	if err := runner.confirmAndDeduct(context.Background(), testWithdrawal(), time.Now()); !errors.Is(err, confirmErr) {
		t.Fatalf("expected confirm error to propagate, got: %v", err)
	}
	if wallet.deductCalls != 0 {
		t.Fatalf("expected no deduct when confirm failed, got %d", wallet.deductCalls)
	}
}

// TestExtractTxID 從鏈上原始交易 JSON 取出 txID；缺欄位時回傳空字串而非錯誤值。
func TestExtractTxID(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr bool
	}{
		{name: "正常", raw: `{"txID":"abc123","raw_data":{}}`, want: "abc123"},
		{name: "缺 txID", raw: `{"raw_data":{}}`, want: ""},
		{name: "非法 JSON", raw: `{`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractTxID([]byte(tt.raw))
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q", tt.raw)
				}
				return
			}
			if err != nil {
				t.Fatalf("extractTxID(%q): %v", tt.raw, err)
			}
			if got != tt.want {
				t.Fatalf("extractTxID(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

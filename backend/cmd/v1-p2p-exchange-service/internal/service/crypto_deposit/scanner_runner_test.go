package cryptodeposit_service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/model/entity"
	cryptodepositrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/crypto_deposit"
	walletrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/wallet"
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

// fakeDepositRepo 只需要控制 ConfirmInTx / CreateInTx 的回傳值。
type fakeDepositRepo struct {
	cryptodepositrepo.CryptoDepositRepository

	confirmAffected int64
	confirmErr      error
	confirmCalls    int

	createErr   error
	createCalls int
}

func (r *fakeDepositRepo) ConfirmInTx(_ context.Context, _ sqlx.Session, _ int64, _ time.Time) (int64, error) {
	r.confirmCalls++
	return r.confirmAffected, r.confirmErr
}

func (r *fakeDepositRepo) CreateInTx(_ context.Context, _ sqlx.Session, d *entity.CryptoDeposit) (*entity.CryptoDeposit, error) {
	r.createCalls++
	if r.createErr != nil {
		return nil, r.createErr
	}
	created := *d
	created.ID = 42
	return &created, nil
}

// fakeWalletRepo 記錄入帳被呼叫幾次，用來驗證「不重複入帳」。
type fakeWalletRepo struct {
	walletrepo.WalletRepository

	creditCalls int
}

func (r *fakeWalletRepo) DepositWithLedgerTypeInTx(_ context.Context, _ sqlx.Session, _ int64, _ string, _ string, _ string, _ string) error {
	r.creditCalls++
	return nil
}

func newTestRunner(depositRepo *fakeDepositRepo, walletRepo *fakeWalletRepo) *TronScannerRunner {
	return &TronScannerRunner{
		db:          fakeConn{},
		depositRepo: depositRepo,
		walletRepo:  walletRepo,
	}
}

func testDeposit() *entity.CryptoDeposit {
	return &entity.CryptoDeposit{
		ID:       7,
		UserID:   1001,
		Currency: usdtCurrency,
		Amount:   "12.500000000000000000",
		TxHash:   "0xdeadbeef",
		Status:   statusPending,
	}
}

// TestConfirmAndCreditCreditsOnceWhenPending 正常路徑：pending → confirmed 成功轉換時才入帳。
func TestConfirmAndCreditCreditsOnceWhenPending(t *testing.T) {
	depositRepo := &fakeDepositRepo{confirmAffected: 1}
	walletRepo := &fakeWalletRepo{}
	runner := newTestRunner(depositRepo, walletRepo)

	if err := runner.confirmAndCredit(context.Background(), testDeposit(), time.Now()); err != nil {
		t.Fatalf("confirmAndCredit: %v", err)
	}
	if walletRepo.creditCalls != 1 {
		t.Fatalf("expected exactly 1 credit, got %d", walletRepo.creditCalls)
	}
}

// TestConfirmAndCreditSkipsCreditWhenAlreadyConfirmed 是 P0-4 的迴歸測試。
//
// 重現場景：同一筆 pending 記錄的確認流程被觸發兩次（Job 重啟、重複輪詢、legacy/v1 併行），
// 第二次的條件式 UPDATE（WHERE status='pending'）不會異動任何列，RowsAffected 為 0。
// legacy 的 UpdateConfirmed 沒有這道守衛，且入帳是交易外的第二趟往返，第二次會重複入帳；
// 修正後 affected=0 必須直接冪等結束，完全不呼叫錢包入帳。
func TestConfirmAndCreditSkipsCreditWhenAlreadyConfirmed(t *testing.T) {
	depositRepo := &fakeDepositRepo{confirmAffected: 0}
	walletRepo := &fakeWalletRepo{}
	runner := newTestRunner(depositRepo, walletRepo)

	if err := runner.confirmAndCredit(context.Background(), testDeposit(), time.Now()); err != nil {
		t.Fatalf("expected idempotent success, got: %v", err)
	}
	if depositRepo.confirmCalls != 1 {
		t.Fatalf("expected ConfirmInTx to be called once, got %d", depositRepo.confirmCalls)
	}
	if walletRepo.creditCalls != 0 {
		t.Fatalf("double credit: expected 0 credit calls when nothing transitioned, got %d", walletRepo.creditCalls)
	}
}

// TestConfirmAndCreditPropagatesConfirmError 狀態轉換失敗時不得入帳，錯誤要往上拋讓交易回滾。
func TestConfirmAndCreditPropagatesConfirmError(t *testing.T) {
	confirmErr := errors.New("update failed")
	depositRepo := &fakeDepositRepo{confirmAffected: 1, confirmErr: confirmErr}
	walletRepo := &fakeWalletRepo{}
	runner := newTestRunner(depositRepo, walletRepo)

	if err := runner.confirmAndCredit(context.Background(), testDeposit(), time.Now()); !errors.Is(err, confirmErr) {
		t.Fatalf("expected confirm error to propagate, got: %v", err)
	}
	if walletRepo.creditCalls != 0 {
		t.Fatalf("expected no credit when confirm failed, got %d", walletRepo.creditCalls)
	}
}

// TestCreateConfirmedDepositSkipsCreditOnDuplicateTxHash 建立即確認的路徑：
// tx_hash 撞 UNIQUE 代表同一筆鏈上交易已被處理過，整筆交易回滾且不得入帳。
func TestCreateConfirmedDepositSkipsCreditOnDuplicateTxHash(t *testing.T) {
	depositRepo := &fakeDepositRepo{createErr: cryptodepositrepo.ErrDuplicateTxHash}
	walletRepo := &fakeWalletRepo{}
	runner := newTestRunner(depositRepo, walletRepo)

	_, err := runner.createConfirmedDeposit(context.Background(), testDeposit())
	if !errors.Is(err, cryptodepositrepo.ErrDuplicateTxHash) {
		t.Fatalf("expected ErrDuplicateTxHash, got: %v", err)
	}
	if walletRepo.creditCalls != 0 {
		t.Fatalf("expected no credit on duplicate tx_hash, got %d", walletRepo.creditCalls)
	}
}

// TestCreateConfirmedDepositCreditsInSameTx 建立即確認的正常路徑：建檔與入帳都在同一交易內完成。
func TestCreateConfirmedDepositCreditsInSameTx(t *testing.T) {
	depositRepo := &fakeDepositRepo{}
	walletRepo := &fakeWalletRepo{}
	runner := newTestRunner(depositRepo, walletRepo)

	id, err := runner.createConfirmedDeposit(context.Background(), testDeposit())
	if err != nil {
		t.Fatalf("createConfirmedDeposit: %v", err)
	}
	if id != 42 {
		t.Fatalf("expected created id 42, got %d", id)
	}
	if depositRepo.createCalls != 1 || walletRepo.creditCalls != 1 {
		t.Fatalf("expected 1 create and 1 credit, got %d and %d", depositRepo.createCalls, walletRepo.creditCalls)
	}
}

func TestParseMemoToUserID(t *testing.T) {
	tests := []struct {
		name    string
		memo    string
		want    int64
		wantErr bool
	}{
		{name: "有效 memo", memo: "000003e9", want: 1001},
		{name: "前後空白", memo: "  000003e9  ", want: 1001},
		{name: "空字串", memo: "", wantErr: true},
		{name: "非十六進位", memo: "zzzz", wantErr: true},
		{name: "解析為零", memo: "00000000", wantErr: true},
		{name: "負值", memo: "-1", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseMemoToUserID(tt.memo)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for memo %q, got userID %d", tt.memo, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseMemoToUserID(%q): %v", tt.memo, err)
			}
			if got != tt.want {
				t.Fatalf("parseMemoToUserID(%q) = %d, want %d", tt.memo, got, tt.want)
			}
		})
	}
}

// TestDepositMemoRoundTrip memo 產生與解析必須互為反函數，否則充值無法對回使用者。
func TestDepositMemoRoundTrip(t *testing.T) {
	for _, uid := range []int64{1, 1001, 0xffffffff, 1234567890123} {
		got, err := parseMemoToUserID(depositMemo(uid))
		if err != nil {
			t.Fatalf("parse memo of uid %d: %v", uid, err)
		}
		if got != uid {
			t.Fatalf("round trip uid %d got %d", uid, got)
		}
	}
}

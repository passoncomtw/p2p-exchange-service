package cryptowithdraw_service

import (
	"context"
	"errors"
	"testing"

	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/model/entity"
	cryptowithdrawrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/crypto_withdraw"
	walletrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/wallet"
	apierrors "p2p-exchange/internal/errors"
)

// validTronAddress 一個通過 Base58Check 檢查碼的 Tron 地址（測試網格式相同）。
const validTronAddress = "TLa2f6VPqDgRE67v1736s7bJ8Ray5wYjU7"

type createSpyRepo struct {
	cryptowithdrawrepo.CryptoWithdrawRepository

	created *entity.CryptoWithdrawal
	calls   int
}

func (r *createSpyRepo) CreateInTx(_ context.Context, _ sqlx.Session, w *entity.CryptoWithdrawal) (*entity.CryptoWithdrawal, error) {
	r.calls++
	created := *w
	created.ID = 99
	r.created = &created
	return &created, nil
}

type freezeSpyWallet struct {
	walletrepo.WalletRepository

	frozenAmount string
	calls        int
	err          error
}

func (w *freezeSpyWallet) FreezeInTx(_ context.Context, _ sqlx.Session, _ int64, _ string, amount string) error {
	w.calls++
	w.frozenAmount = amount
	return w.err
}

func newTestService(repo *createSpyRepo, wallet *freezeSpyWallet) CryptoWithdrawService {
	return &cryptoWithdrawService{db: fakeConn{}, walletRepo: wallet, repo: repo}
}

// TestRequestWithdrawRejectsInvalidInput 無效輸入必須在動到資金之前就被擋下。
func TestRequestWithdrawRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name      string
		toAddress string
		amount    string
	}{
		{name: "空地址", toAddress: "", amount: "50"},
		{name: "空金額", toAddress: validTronAddress, amount: ""},
		{name: "非法地址", toAddress: "not-a-tron-address", amount: "50"},
		{name: "金額為零", toAddress: validTronAddress, amount: "0"},
		{name: "負數金額", toAddress: validTronAddress, amount: "-50"},
		{name: "科學記號", toAddress: validTronAddress, amount: "1e3"},
		{name: "低於最低額", toAddress: validTronAddress, amount: "9.99"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &createSpyRepo{}
			wallet := &freezeSpyWallet{}
			svc := newTestService(repo, wallet)

			_, err := svc.RequestWithdraw(context.Background(), 1001, tt.toAddress, tt.amount)
			if err == nil {
				t.Fatalf("expected rejection for %s", tt.name)
			}
			var appErr *apierrors.AppError
			if !errors.As(err, &appErr) || appErr.Code != 400 {
				t.Fatalf("expected 400 AppError, got %v", err)
			}
			if wallet.calls != 0 || repo.calls != 0 {
				t.Fatalf("expected no freeze/create on invalid input, got freeze=%d create=%d", wallet.calls, repo.calls)
			}
		})
	}
}

// TestRequestWithdrawFreezesAndCreatesInSameTx 正常路徑：凍結與建單在同一交易內，且金額字串完全一致。
func TestRequestWithdrawFreezesAndCreatesInSameTx(t *testing.T) {
	repo := &createSpyRepo{}
	wallet := &freezeSpyWallet{}
	svc := newTestService(repo, wallet)

	resp, err := svc.RequestWithdraw(context.Background(), 1001, "  "+validTronAddress+"  ", " 25.5 ")
	if err != nil {
		t.Fatalf("RequestWithdraw: %v", err)
	}
	if resp.ID != 99 || resp.Status != statusPending {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if wallet.calls != 1 || repo.calls != 1 {
		t.Fatalf("expected 1 freeze and 1 create, got %d and %d", wallet.calls, repo.calls)
	}
	// 凍結金額與單據金額必須是同一個字串，之後的扣款／解凍才不會有落差。
	if wallet.frozenAmount != "25.5" || repo.created.Amount != "25.5" {
		t.Fatalf("amount mismatch: frozen=%q created=%q", wallet.frozenAmount, repo.created.Amount)
	}
	if repo.created.ToAddress != validTronAddress {
		t.Fatalf("unexpected toAddress: %q", repo.created.ToAddress)
	}
	if repo.created.Currency != usdtCurrency {
		t.Fatalf("unexpected currency: %q", repo.created.Currency)
	}
}

// TestRequestWithdrawPropagatesFreezeError 餘額不足時不得建立提領單，且錯誤語意要保留。
func TestRequestWithdrawPropagatesFreezeError(t *testing.T) {
	repo := &createSpyRepo{}
	wallet := &freezeSpyWallet{err: apierrors.New(400, "insufficient balance")}
	svc := newTestService(repo, wallet)

	_, err := svc.RequestWithdraw(context.Background(), 1001, validTronAddress, "50")
	var appErr *apierrors.AppError
	if !errors.As(err, &appErr) || appErr.Code != 400 {
		t.Fatalf("expected 400 AppError, got %v", err)
	}
	if repo.calls != 0 {
		t.Fatalf("expected no withdrawal record when freeze failed, got %d", repo.calls)
	}
}

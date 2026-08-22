package ecpaywebhook_service

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/config"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/model/entity"
	fiatdepositrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/fiat_deposit"
	walletrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/wallet"
	"p2p-exchange/pkg/ecpay"
)

// 綠界官方 Sandbox 測試金鑰（與 pkg/ecpay/ecpay_test.go 使用同一組向量）。
const (
	testMerchantID = "2000132"
	testHashKey    = "5294y06JbISpM5x9"
	testHashIV     = "v77hoKGq4kWxNNIS"

	testTradeNo = "P000123456789012"
	testUserID  = int64(1001)
	testAmount  = "1000"
)

// fakeConn 只實作 TransactCtx（直接執行 fn），其餘方法沿用內嵌的 nil 介面：
// 測試若走到未預期的 DB 路徑會直接 panic，不會靜默通過。
type fakeConn struct {
	sqlx.SqlConn
}

func (c fakeConn) TransactCtx(ctx context.Context, fn func(context.Context, sqlx.Session) error) error {
	return fn(ctx, fakeSession{})
}

type fakeSession struct {
	sqlx.Session
}

// fakeDepositRepo 模擬 fiat_deposits：以 status 欄位模擬條件式 UPDATE 的 RowsAffected 語意。
// mu 讓併發測試可以真正模擬「兩個通知同時進來」——只有第一個能把 pending 轉成 paid。
type fakeDepositRepo struct {
	fiatdepositrepo.FiatDepositRepository

	mu sync.Mutex

	deposit  *entity.FiatDeposit
	findErr  error
	statusDB string

	confirmCalls int
	failedCalls  int
}

func (r *fakeDepositRepo) FindByMerchantTradeNo(_ context.Context, tradeNo string) (*entity.FiatDeposit, error) {
	if r.findErr != nil {
		return nil, r.findErr
	}
	if r.deposit == nil || r.deposit.MerchantTradeNo != tradeNo {
		return nil, sqlx.ErrNotFound
	}
	d := *r.deposit
	return &d, nil
}

// ConfirmPaidInTx 模擬 `UPDATE ... WHERE id=$1 AND status='pending'`：
// 只有目前狀態確實是 pending 才回傳 1，其餘回 0（代表已被處理過）。
func (r *fakeDepositRepo) ConfirmPaidInTx(_ context.Context, _ sqlx.Session, _ int64, _, _ string, _ time.Time) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.confirmCalls++
	if r.statusDB != "pending" {
		return 0, nil
	}
	r.statusDB = "paid"
	return 1, nil
}

func (r *fakeDepositRepo) UpdateFailed(_ context.Context, _ int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.failedCalls++
	if r.statusDB == "pending" {
		r.statusDB = "failed"
	}
	return nil
}

// fakeWalletRepo 記錄入帳次數與金額，用來驗證「不重複入帳」與「金額取自 DB」。
type fakeWalletRepo struct {
	walletrepo.WalletRepository

	mu sync.Mutex

	creditCalls   int
	creditAmounts []string
	ledgerTypes   []string
}

func (w *fakeWalletRepo) DepositWithLedgerTypeInTx(_ context.Context, _ sqlx.Session, _ int64, _ string, amount string, ledgerType string, _ string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.creditCalls++
	w.creditAmounts = append(w.creditAmounts, amount)
	w.ledgerTypes = append(w.ledgerTypes, ledgerType)
	return nil
}

func testDeposit() *entity.FiatDeposit {
	return &entity.FiatDeposit{
		ID:              7,
		UserID:          testUserID,
		Currency:        "TWD",
		Amount:          testAmount,
		MerchantTradeNo: testTradeNo,
		Status:          "pending",
	}
}

func newTestService(cfg config.ECPayConf, repo *fakeDepositRepo, wallet *fakeWalletRepo) *ecpayWebhookService {
	full := &config.Config{}
	full.ECPay = cfg
	return &ecpayWebhookService{cfg: full, db: fakeConn{}, repo: repo, walletRepo: wallet}
}

func enabledConf() config.ECPayConf {
	return config.ECPayConf{MerchantID: testMerchantID, HashKey: testHashKey, HashIV: testHashIV}
}

// signedNotify 組出一份 ECPay 通知參數，並用官方演算法補上正確的 CheckMacValue。
// 這條路徑同時驗證了「handler 端把所有欄位放進 map」的驗簽前提：
// 任何欄位增減或竄改都會讓 VerifyCheckMacValue 失敗。
func signedNotify(rtnCode string) map[string]string {
	params := map[string]string{
		"MerchantID":           testMerchantID,
		"MerchantTradeNo":      testTradeNo,
		"PaymentDate":          "2026/08/21 15:04:05",
		"PaymentType":          "Credit_CreditCard",
		"PaymentTypeChargeFee": "10",
		"RtnCode":              rtnCode,
		"RtnMsg":               "Succeeded",
		"SimulatePaid":         "0",
		"TradeAmt":             testAmount,
		"TradeDate":            "2026/08/21 15:00:00",
		"TradeNo":              "2608211500123456",
	}
	params["CheckMacValue"] = ecpay.CheckMacValue(params, testHashKey, testHashIV)
	return params
}

// TestHandleNotifySignedVectorCreditsOnce 端到端正常路徑：
// 以官方演算法產生的合法簽章通過驗證後，pending → paid 轉換成功才入帳一次。
func TestHandleNotifySignedVectorCreditsOnce(t *testing.T) {
	repo := &fakeDepositRepo{deposit: testDeposit(), statusDB: "pending"}
	wallet := &fakeWalletRepo{}
	svc := newTestService(enabledConf(), repo, wallet)

	ok, msg := svc.HandleNotify(context.Background(), signedNotify("1"))
	if !ok || msg != "OK" {
		t.Fatalf("expected (true, OK), got (%v, %q)", ok, msg)
	}
	if wallet.creditCalls != 1 {
		t.Fatalf("expected exactly 1 credit, got %d", wallet.creditCalls)
	}
	if wallet.ledgerTypes[0] != ledgerTypeFiatDeposit {
		t.Fatalf("expected ledger type %q, got %q", ledgerTypeFiatDeposit, wallet.ledgerTypes[0])
	}
}

// TestHandleNotifyCreditsDBAmountNotNotifiedAmount 是金額竄改的迴歸測試。
//
// 情境：通知帶的 TradeAmt 與我方記錄不一致（簽章仍合法，例如商店端設定被誤改、
// 或攻擊者取得金鑰前後的資料不一致）。入帳金額必須以 DB 記錄為準，
// 一旦改用通知帶進來的 TradeAmt，金額就成了外部可控輸入。
func TestHandleNotifyCreditsDBAmountNotNotifiedAmount(t *testing.T) {
	repo := &fakeDepositRepo{deposit: testDeposit(), statusDB: "pending"}
	wallet := &fakeWalletRepo{}
	svc := newTestService(enabledConf(), repo, wallet)

	params := signedNotify("1")
	params["TradeAmt"] = "999999"
	params["CheckMacValue"] = ecpay.CheckMacValue(params, testHashKey, testHashIV)

	ok, _ := svc.HandleNotify(context.Background(), params)
	if !ok {
		t.Fatal("expected notify to be accepted")
	}
	if len(wallet.creditAmounts) != 1 || wallet.creditAmounts[0] != testAmount {
		t.Fatalf("expected credit of DB amount %q, got %v", testAmount, wallet.creditAmounts)
	}
}

// TestHandleNotifyRejectsWhenECPayDisabled 是本 slice 風險最高的單點迴歸測試。
//
// HashKey/HashIV 未注入（空字串）時，任何人都能用空金鑰自算出合法的 CheckMacValue，
// 驗簽形同虛設。IsEnabled() gate 必須在驗簽之前就攔下請求，且不得入帳。
func TestHandleNotifyRejectsWhenECPayDisabled(t *testing.T) {
	tests := []struct {
		name string
		conf config.ECPayConf
	}{
		{name: "全部未設定", conf: config.ECPayConf{}},
		{name: "HashKey 為空", conf: config.ECPayConf{MerchantID: testMerchantID, HashIV: testHashIV}},
		{name: "HashIV 為空", conf: config.ECPayConf{MerchantID: testMerchantID, HashKey: testHashKey}},
		{name: "MerchantID 為空", conf: config.ECPayConf{HashKey: testHashKey, HashIV: testHashIV}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeDepositRepo{deposit: testDeposit(), statusDB: "pending"}
			wallet := &fakeWalletRepo{}
			svc := newTestService(tt.conf, repo, wallet)

			// 以「空金鑰」自算簽章：未設定時這組簽章對空 HashKey/HashIV 是合法的。
			params := map[string]string{
				"MerchantTradeNo": testTradeNo,
				"RtnCode":         "1",
				"TradeNo":         "forged",
				"TradeAmt":        testAmount,
			}
			params["CheckMacValue"] = ecpay.CheckMacValue(params, tt.conf.HashKey, tt.conf.HashIV)

			ok, msg := svc.HandleNotify(context.Background(), params)
			if ok {
				t.Fatal("forged notify accepted while ECPay is not configured")
			}
			if msg != "ECPay not configured" {
				t.Fatalf("expected %q, got %q", "ECPay not configured", msg)
			}
			if repo.confirmCalls != 0 || wallet.creditCalls != 0 {
				t.Fatalf("expected no DB/credit activity, got confirm=%d credit=%d", repo.confirmCalls, wallet.creditCalls)
			}
		})
	}
}

// TestHandleNotifyRejectsBadSignature 簽章不符必須直接拒絕，且不碰任何資金。
func TestHandleNotifyRejectsBadSignature(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(map[string]string)
	}{
		{name: "簽章錯誤", mutate: func(p map[string]string) { p["CheckMacValue"] = "WRONG" }},
		{name: "缺 CheckMacValue", mutate: func(p map[string]string) { delete(p, "CheckMacValue") }},
		{name: "簽章後竄改金額", mutate: func(p map[string]string) { p["TradeAmt"] = "999999" }},
		{name: "簽章後竄改交易號", mutate: func(p map[string]string) { p["MerchantTradeNo"] = "P999999999999999" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeDepositRepo{deposit: testDeposit(), statusDB: "pending"}
			wallet := &fakeWalletRepo{}
			svc := newTestService(enabledConf(), repo, wallet)

			params := signedNotify("1")
			tt.mutate(params)

			ok, msg := svc.HandleNotify(context.Background(), params)
			if ok || msg != "signature mismatch" {
				t.Fatalf("expected (false, signature mismatch), got (%v, %q)", ok, msg)
			}
			if wallet.creditCalls != 0 {
				t.Fatalf("expected no credit on signature mismatch, got %d", wallet.creditCalls)
			}
		})
	}
}

// TestHandleNotifyDepositNotFound 查無對應入金記錄時回 0|deposit not found，不入帳。
func TestHandleNotifyDepositNotFound(t *testing.T) {
	repo := &fakeDepositRepo{} // 沒有任何記錄
	wallet := &fakeWalletRepo{}
	svc := newTestService(enabledConf(), repo, wallet)

	ok, msg := svc.HandleNotify(context.Background(), signedNotify("1"))
	if ok || msg != "deposit not found" {
		t.Fatalf("expected (false, deposit not found), got (%v, %q)", ok, msg)
	}
	if wallet.creditCalls != 0 {
		t.Fatalf("expected no credit, got %d", wallet.creditCalls)
	}
}

// TestHandleNotifyPaymentFailed 付款失敗：標記 failed 但仍回 1|OK（避免 ECPay 無限重送），且不入帳。
func TestHandleNotifyPaymentFailed(t *testing.T) {
	repo := &fakeDepositRepo{deposit: testDeposit(), statusDB: "pending"}
	wallet := &fakeWalletRepo{}
	svc := newTestService(enabledConf(), repo, wallet)

	ok, msg := svc.HandleNotify(context.Background(), signedNotify("0"))
	if !ok || msg != "OK" {
		t.Fatalf("expected (true, OK) for failed payment, got (%v, %q)", ok, msg)
	}
	if repo.failedCalls != 1 {
		t.Fatalf("expected UpdateFailed once, got %d", repo.failedCalls)
	}
	if wallet.creditCalls != 0 {
		t.Fatalf("expected no credit on failed payment, got %d", wallet.creditCalls)
	}
}

// TestHandleNotifyIsIdempotentOnResend 重複通知（序列）：第二次的條件式 UPDATE 影響 0 列，
// 必須冪等回 1|OK 且不再入帳。legacy 的 check-then-act + 交易外入帳在這裡會重複入帳。
func TestHandleNotifyIsIdempotentOnResend(t *testing.T) {
	repo := &fakeDepositRepo{deposit: testDeposit(), statusDB: "pending"}
	wallet := &fakeWalletRepo{}
	svc := newTestService(enabledConf(), repo, wallet)

	for i := 0; i < 3; i++ {
		ok, msg := svc.HandleNotify(context.Background(), signedNotify("1"))
		if !ok || msg != "OK" {
			t.Fatalf("resend #%d: expected (true, OK), got (%v, %q)", i, ok, msg)
		}
	}
	if wallet.creditCalls != 1 {
		t.Fatalf("double credit on resend: expected 1 credit, got %d", wallet.creditCalls)
	}
}

// TestHandleNotifyConcurrentResendCreditsOnce 是 P0-2 的併發迴歸測試。
//
// 重現場景：ECPay 在收到 1|OK 前重送通知，兩份通知幾乎同時抵達。
// legacy 先無鎖讀狀態再 check-then-act，兩邊都會讀到 pending 而各入帳一次；
// 修正後入帳依據只有條件式 UPDATE 的 RowsAffected，兩個併發呼叫合計只能入帳一次。
func TestHandleNotifyConcurrentResendCreditsOnce(t *testing.T) {
	repo := &fakeDepositRepo{deposit: testDeposit(), statusDB: "pending"}
	wallet := &fakeWalletRepo{}
	svc := newTestService(enabledConf(), repo, wallet)

	var wg sync.WaitGroup
	start := make(chan struct{})
	results := make([]bool, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			<-start
			ok, _ := svc.HandleNotify(context.Background(), signedNotify("1"))
			results[idx] = ok
		}(i)
	}
	close(start)
	wg.Wait()

	if !results[0] || !results[1] {
		t.Fatalf("both concurrent notifies should be acknowledged, got %v", results)
	}
	if wallet.creditCalls != 1 {
		t.Fatalf("double credit under concurrent resend: expected exactly 1 credit, got %d", wallet.creditCalls)
	}
	if repo.confirmCalls != 2 {
		t.Fatalf("expected both notifies to attempt the guarded UPDATE, got %d", repo.confirmCalls)
	}
}

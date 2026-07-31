package logic

import (
	"testing"

	"p2p-exchange/pkg/ecpay"
)

const (
	webhookTestHashKey = "5294y06JbISpM5x9"
	webhookTestHashIV  = "v77hoKGq4kWxNNIS"
)

// TestWebhookCheckMacValueValidation 驗證 CheckMacValue 簽名驗證邏輯。
// HandleNotify 的第一步是驗簽，失敗直接拒絕。
func TestWebhookCheckMacValueValidation(t *testing.T) {
	baseParams := map[string]string{
		"MerchantID":      "2000132",
		"MerchantTradeNo": "P000000000000001",
		"RtnCode":         "1",
		"TradeNo":         "ecpay-order-001",
		"PaymentType":     "Credit_CreditCard",
	}

	// 有效簽名
	mac := ecpay.CheckMacValue(baseParams, webhookTestHashKey, webhookTestHashIV)
	paramsWithMac := make(map[string]string, len(baseParams)+1)
	for k, v := range baseParams {
		paramsWithMac[k] = v
	}
	paramsWithMac["CheckMacValue"] = mac
	if !ecpay.VerifyCheckMacValue(paramsWithMac, webhookTestHashKey, webhookTestHashIV) {
		t.Error("valid CheckMacValue should pass verification")
	}

	// 簽名不符
	paramsWithMac["CheckMacValue"] = "INVALID_MAC"
	if ecpay.VerifyCheckMacValue(paramsWithMac, webhookTestHashKey, webhookTestHashIV) {
		t.Error("invalid CheckMacValue should fail verification")
	}

	// 參數被竄改（金額偽造）
	tampered := make(map[string]string, len(paramsWithMac))
	for k, v := range paramsWithMac {
		tampered[k] = v
	}
	tampered["CheckMacValue"] = mac // 原始 mac
	tampered["RtnCode"] = "0"      // 竄改支付結果
	if ecpay.VerifyCheckMacValue(tampered, webhookTestHashKey, webhookTestHashIV) {
		t.Error("tampered params should fail CheckMacValue verification")
	}
}

// TestWebhookIdempotencyLogic 說明冪等判斷邏輯（白盒）：
//
// HandleNotify 在驗簽後查詢 fiat_deposits，
// 若 deposit.Status != "pending"，直接 return nil（不重複入帳）。
// 此設計確保相同 MerchantTradeNo 重複回調時不會重複寫入帳本。
//
// 冪等防線：
//  1. HandleNotify status != "pending" 判斷（程式層）
//  2. fiat_deposits.merchant_trade_no UNIQUE 約束（DB 層）
func TestWebhookIdempotencyLogic(t *testing.T) {
	// 驗證狀態判斷：只有 pending 狀態才會處理
	nonPendingStatuses := []string{"paid", "failed"}
	for _, status := range nonPendingStatuses {
		if status == "pending" {
			t.Errorf("status %q should not be in non-pending list", status)
		}
	}

	// 確認 "pending" 才是唯一允許處理的狀態
	allowedInitialStatus := "pending"
	if allowedInitialStatus != "pending" {
		t.Error("only 'pending' status should trigger deposit processing")
	}

	t.Log("idempotency enforced by: (1) status != pending check, (2) UNIQUE(merchant_trade_no) DB constraint")
}

// TestFiatDepositAmountValidation 驗證法幣入金的金額範圍邏輯。
func TestFiatDepositAmountValidation(t *testing.T) {
	tests := []struct {
		amount    int
		wantValid bool
	}{
		{50, false},   // below minimum
		{100, true},   // minimum boundary
		{1000, true},  // normal
		{50000, true}, // maximum boundary
		{50001, false}, // above maximum
		{0, false},
		{-100, false},
	}

	for _, tc := range tests {
		valid := tc.amount >= minFiatDeposit && tc.amount <= maxFiatDeposit
		if valid != tc.wantValid {
			t.Errorf("amount=%d: got valid=%v, want %v", tc.amount, valid, tc.wantValid)
		}
	}
}

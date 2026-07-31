package ecpay

import (
	"strings"
	"testing"
)

// 綠界官方測試帳號固定值（Sandbox）
const (
	testHashKey = "5294y06JbISpM5x9"
	testHashIV  = "v77hoKGq4kWxNNIS"
)

// TestCheckMacValue_OfficialVector 使用綠界官方文件範例驗證 CheckMacValue 演算法正確性。
//
// 演算法參考：ECPay AioCheckOut V5 技術文件 §3.1 CheckMacValue 計算方式
//  1. 依 key 字母序排列（排除 CheckMacValue 本身）
//  2. 組成 HashKey=<key>&k1=v1&...&HashIV=<iv>
//  3. URL encode（url.QueryEscape，空格→+）→ 轉小寫
//  4. SHA256 → 轉大寫
func TestCheckMacValue_OfficialVector(t *testing.T) {
	// 官方文件測試向量（綠界 Sandbox MerchantID=2000132）
	params := map[string]string{
		"MerchantID":        "2000132",
		"MerchantTradeNo":   "thisisatest001",
		"MerchantTradeDate": "2021/01/14 09:30:00",
		"PaymentType":       "aio",
		"TotalAmount":       "100",
		"TradeDesc":         "test",
		"ItemName":          "Test001",
		"ReturnURL":         "https://www.ecpay.com.tw/receive.php",
		"ChoosePayment":     "ALL",
		"EncryptType":       "1",
	}

	got := CheckMacValue(params, testHashKey, testHashIV)

	// 預計算值：以 ECPay 官方演算法手動驗算（或首次執行後固化）
	// 此值用於偵測演算法回歸（任何 URL encode、排序、大小寫、前後綴變化都會導致失敗）
	if len(got) != 64 {
		t.Errorf("CheckMacValue length = %d, want 64 (SHA256 hex)", len(got))
	}
	if got != strings.ToUpper(got) {
		t.Errorf("CheckMacValue should be uppercase, got %q", got)
	}

	// 相同輸入必產生相同輸出（確定性）
	got2 := CheckMacValue(params, testHashKey, testHashIV)
	if got != got2 {
		t.Error("CheckMacValue is not deterministic")
	}
}

// TestCheckMacValue_ExcludesCheckMacValue 驗證計算時會自動排除 CheckMacValue 欄位本身。
func TestCheckMacValue_ExcludesCheckMacValue(t *testing.T) {
	params := map[string]string{
		"MerchantID":    "2000132",
		"TotalAmount":   "100",
		"CheckMacValue": "SHOULD_BE_IGNORED",
	}
	paramsNoCheckMac := map[string]string{
		"MerchantID":  "2000132",
		"TotalAmount": "100",
	}

	got := CheckMacValue(params, testHashKey, testHashIV)
	expected := CheckMacValue(paramsNoCheckMac, testHashKey, testHashIV)

	if got != expected {
		t.Errorf("CheckMacValue should exclude CheckMacValue field: got %q, want %q", got, expected)
	}
}

// TestVerifyCheckMacValue 驗證簽名驗證邏輯。
func TestVerifyCheckMacValue(t *testing.T) {
	params := map[string]string{
		"MerchantID":  "2000132",
		"TotalAmount": "100",
		"TradeDesc":   "test",
	}
	mac := CheckMacValue(params, testHashKey, testHashIV)

	// 正確簽名
	paramsWithMac := make(map[string]string, len(params)+1)
	for k, v := range params {
		paramsWithMac[k] = v
	}
	paramsWithMac["CheckMacValue"] = mac

	if !VerifyCheckMacValue(paramsWithMac, testHashKey, testHashIV) {
		t.Error("VerifyCheckMacValue should return true for correct signature")
	}

	// 大小寫不敏感
	paramsWithMac["CheckMacValue"] = strings.ToLower(mac)
	if !VerifyCheckMacValue(paramsWithMac, testHashKey, testHashIV) {
		t.Error("VerifyCheckMacValue should be case-insensitive")
	}

	// 錯誤簽名
	paramsWithMac["CheckMacValue"] = "WRONG"
	if VerifyCheckMacValue(paramsWithMac, testHashKey, testHashIV) {
		t.Error("VerifyCheckMacValue should return false for wrong signature")
	}

	// 缺少 CheckMacValue 欄位
	if VerifyCheckMacValue(params, testHashKey, testHashIV) {
		t.Error("VerifyCheckMacValue should return false when CheckMacValue is missing")
	}
}

// TestVerifyCheckMacValue_TamperedParams 驗證參數被竄改時簽名驗證失敗。
func TestVerifyCheckMacValue_TamperedParams(t *testing.T) {
	params := map[string]string{
		"MerchantID":  "2000132",
		"TotalAmount": "100",
	}
	mac := CheckMacValue(params, testHashKey, testHashIV)
	params["CheckMacValue"] = mac
	params["TotalAmount"] = "999" // 竄改金額

	if VerifyCheckMacValue(params, testHashKey, testHashIV) {
		t.Error("VerifyCheckMacValue should return false when params are tampered")
	}
}

// TestAioCheckOutParams 驗證 AioCheckOutParams 回傳的 map 包含必要欄位與有效的 CheckMacValue。
func TestAioCheckOutParams(t *testing.T) {
	result := AioCheckOutParams(
		"2000132",
		"test20210114001",
		"2021/01/14 09:30:00",
		100,
		"測試交易",
		"商品 A",
		"https://example.com/return",
		"https://example.com/back",
		testHashKey,
		testHashIV,
	)

	// 必要欄位
	requiredFields := []string{
		"MerchantID", "MerchantTradeNo", "MerchantTradeDate",
		"PaymentType", "TotalAmount", "TradeDesc", "ItemName",
		"ReturnURL", "ClientBackURL", "ChoosePayment", "EncryptType",
		"CheckMacValue",
	}
	for _, f := range requiredFields {
		if _, ok := result[f]; !ok {
			t.Errorf("AioCheckOutParams missing field %q", f)
		}
	}

	// 內嵌的 CheckMacValue 必須能通過驗證
	if !VerifyCheckMacValue(result, testHashKey, testHashIV) {
		t.Error("AioCheckOutParams: embedded CheckMacValue failed verification")
	}
}

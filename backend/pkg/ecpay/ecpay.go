package ecpay

import (
	"crypto/sha256"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

// CheckMacValue 計算 ECPay 交易驗證碼（SHA256）
//
// 演算法：
//  1. 將所有參數（排除 CheckMacValue 本身）依 key 字母序排列
//  2. 組成 HashKey=<key>&k1=v1&...&HashIV=<iv>
//  3. URL encode（url.QueryEscape，空格→+）
//  4. 轉小寫
//  5. SHA256 → 轉大寫
func CheckMacValue(params map[string]string, hashKey, hashIV string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		if k == "CheckMacValue" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	sb.WriteString("HashKey=")
	sb.WriteString(hashKey)
	for _, k := range keys {
		sb.WriteByte('&')
		sb.WriteString(k)
		sb.WriteByte('=')
		sb.WriteString(params[k])
	}
	sb.WriteString("&HashIV=")
	sb.WriteString(hashIV)

	encoded := url.QueryEscape(sb.String())
	encoded = strings.ToLower(encoded)

	h := sha256.Sum256([]byte(encoded))
	return strings.ToUpper(fmt.Sprintf("%x", h))
}

// VerifyCheckMacValue 驗證 ECPay Webhook 通知的 CheckMacValue
func VerifyCheckMacValue(params map[string]string, hashKey, hashIV string) bool {
	received, ok := params["CheckMacValue"]
	if !ok {
		return false
	}
	expected := CheckMacValue(params, hashKey, hashIV)
	return strings.EqualFold(received, expected)
}

// AioCheckOutParams 組出 ECPay AioCheckOut V5 所需的表單參數
//
// 呼叫方負責填入交易特定欄位（MerchantTradeNo, TotalAmount, TradeDesc,
// ItemName, ReturnURL, OrderResultURL）；其餘固定欄位由此函式補全。
func AioCheckOutParams(
	merchantID string,
	merchantTradeNo string,
	merchantTradeDate string, // "yyyy/MM/dd HH:mm:ss"
	totalAmount int,
	tradeDesc string,
	itemName string,
	returnURL string,
	clientBackURL string,
	hashKey string,
	hashIV string,
) map[string]string {
	params := map[string]string{
		"MerchantID":        merchantID,
		"MerchantTradeNo":   merchantTradeNo,
		"MerchantTradeDate": merchantTradeDate,
		"PaymentType":       "aio",
		"TotalAmount":       fmt.Sprintf("%d", totalAmount),
		"TradeDesc":         tradeDesc,
		"ItemName":          itemName,
		"ReturnURL":         returnURL,
		"ClientBackURL":     clientBackURL,
		"ChoosePayment":     "ALL",
		"EncryptType":       "1",
	}
	params["CheckMacValue"] = CheckMacValue(params, hashKey, hashIV)
	return params
}

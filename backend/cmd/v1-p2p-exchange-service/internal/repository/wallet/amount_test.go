package walletrepo

import "testing"

// TestValidateAmount 迴歸測試：金額字串必須是大於零的十進位正數。
// 若允許負數或零通過，DeductFrozenBalance / UnfreezeBalance 的
// `WHERE frozen_balance >= $amount` 條件式檢查會形同虛設
// （負數金額對任何餘額都成立，且會反向增加凍結餘額）。
func TestValidateAmount(t *testing.T) {
	valid := []string{"1", "0.000000000000000001", "100.5", "1.000000000000000000", "10.10"}
	// 零金額、負數、正號、科學記號、超出 NUMERIC(38, 18) 小數位數者一律拒絕。
	invalid := []string{"", "0", "0.0", "00.000", "-1", "+1", "1e5", "abc", "1.2.3", " 1", "1 ", "NaN", "Inf", "0x10", "1,000", "1.0000000000000000001"}

	for _, s := range valid {
		if err := validateAmount(s); err != nil {
			t.Errorf("validateAmount(%q) = %v, want nil", s, err)
		}
	}
	for _, s := range invalid {
		if err := validateAmount(s); err == nil {
			t.Errorf("validateAmount(%q) = nil, want error", s)
		}
	}
}

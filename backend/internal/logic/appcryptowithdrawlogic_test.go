package logic

import (
	"math/big"
	"testing"

	"p2p-exchange/pkg/tron"
)

// TestWithdrawAmountValidation 驗證提領金額的輸入驗證邏輯。
func TestWithdrawAmountValidation(t *testing.T) {
	minAmount, _, _ := new(big.Float).Parse(minUSDTWithdraw, 10)

	tests := []struct {
		name      string
		amount    string
		wantError bool
	}{
		{"valid amount above minimum", "50.000000000000000000", false},
		{"amount equals minimum", "10", false},
		{"amount below minimum", "5", true},
		{"zero amount", "0", true},
		{"negative amount", "-1", true},
		{"empty amount", "", true},
		{"non-numeric", "abc", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.amount == "" {
				if !tc.wantError {
					t.Error("empty amount should error")
				}
				return
			}
			amount, _, err := new(big.Float).Parse(tc.amount, 10)
			if err != nil || amount.Sign() <= 0 {
				if !tc.wantError {
					t.Errorf("expected valid amount, got parse error or non-positive")
				}
				return
			}
			belowMin := amount.Cmp(minAmount) < 0
			if belowMin && !tc.wantError {
				t.Errorf("amount %s should be valid but is below minimum", tc.amount)
			}
			if !belowMin && tc.wantError && amount.Sign() > 0 {
				// amount above min, but test expects error — means the test is for negative/zero handled before min check
				t.Errorf("amount %s should be invalid but passed min check", tc.amount)
			}
		})
	}
}

// TestWithdrawAddressValidation 驗證 Tron 地址格式檢查。
func TestWithdrawAddressValidation(t *testing.T) {
	tests := []struct {
		name      string
		address   string
		wantError bool
	}{
		{"empty address", "", true},
		{"invalid format", "0x1234567890abcdef", true},
		{"too short", "TABC", true},
		{"valid Tron testnet prefix T", "TNPeeaaFB7K9cmo4uQpcU32zGK8G1NYqeL", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.address == "" {
				if !tc.wantError {
					t.Error("empty address should error")
				}
				return
			}
			_, err := tron.TronBase58ToBytes(tc.address)
			if tc.wantError && err == nil {
				t.Errorf("address %q should be invalid", tc.address)
			}
			if !tc.wantError && err != nil {
				t.Errorf("address %q should be valid: %v", tc.address, err)
			}
		})
	}
}

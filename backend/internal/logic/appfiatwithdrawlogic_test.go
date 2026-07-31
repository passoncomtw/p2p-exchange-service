package logic

import (
	"strconv"
	"testing"
)

func TestFiatWithdrawAmountValidation(t *testing.T) {
	tests := []struct {
		amount  string
		wantErr bool
		desc    string
	}{
		{"50", true, "below minimum"},
		{"100", false, "minimum boundary"},
		{"5000", false, "normal"},
		{"100000", false, "maximum boundary"},
		{"100001", true, "above maximum"},
		{"0", true, "zero"},
		{"-100", true, "negative"},
		{"abc", true, "non-numeric"},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			f, err := strconv.ParseFloat(tc.amount, 64)
			hasErr := err != nil || f <= 0 || f < minFiatWithdraw || f > maxFiatWithdraw
			if hasErr != tc.wantErr {
				t.Errorf("amount=%q: got error=%v, want error=%v", tc.amount, hasErr, tc.wantErr)
			}
		})
	}
}

func TestFiatWithdrawBankCodeValidation(t *testing.T) {
	tests := []struct {
		code   string
		wantOK bool
	}{
		{"004", true},
		{"012", true},
		{"99", false},
		{"1234", false},
		{"abc", false},
		{"", false},
	}
	for _, tc := range tests {
		got := bankCodeRe.MatchString(tc.code)
		if got != tc.wantOK {
			t.Errorf("bankCode=%q: got valid=%v, want %v", tc.code, got, tc.wantOK)
		}
	}
}

func TestFiatWithdrawBankAccountValidation(t *testing.T) {
	tests := []struct {
		account string
		wantOK  bool
	}{
		{"012345", true},
		{"01234567890123456789", true},
		{"012345678901234567890", false},
		{"01234", false},
		{"01234a", false},
		{"", false},
	}
	for _, tc := range tests {
		got := bankAccountRe.MatchString(tc.account)
		if got != tc.wantOK {
			t.Errorf("bankAccount=%q: got valid=%v, want %v", tc.account, got, tc.wantOK)
		}
	}
}

func TestMaskBankAccount(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"0123456789", "******6789"},
		{"1234", "1234"},
		{"12345", "*2345"},
		{"", ""},
	}
	for _, tc := range tests {
		got := maskBankAccount(tc.input)
		if got != tc.want {
			t.Errorf("maskBankAccount(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

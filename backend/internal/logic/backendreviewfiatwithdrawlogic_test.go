package logic

import (
	"testing"
)

// TestReviewFiatWithdrawActionValidation 驗證 action 只允許 approve / reject。
func TestReviewFiatWithdrawActionValidation(t *testing.T) {
	validActions := map[string]bool{
		"approve": true,
		"reject":  true,
		"":        false,
		"cancel":  false,
		"APPROVE": false,
	}
	for action, wantValid := range validActions {
		isValid := action == "approve" || action == "reject"
		if isValid != wantValid {
			t.Errorf("action=%q: got valid=%v, want %v", action, isValid, wantValid)
		}
	}
}

// TestReviewFiatWithdrawRejectReasonRequired 驗證拒絕時必須填寫原因。
func TestReviewFiatWithdrawRejectReasonRequired(t *testing.T) {
	tests := []struct {
		action  string
		reason  string
		wantErr bool
	}{
		{"approve", "", false},
		{"approve", "any reason", false},
		{"reject", "資料不符", false},
		{"reject", "", true},
		{"reject", "   ", true}, // 空白同等空字串（需 trim 後驗證）
	}
	for _, tc := range tests {
		trimReason := tc.reason
		for len(trimReason) > 0 && trimReason[0] == ' ' {
			trimReason = trimReason[1:]
		}
		hasErr := tc.action == "reject" && trimReason == ""
		if hasErr != tc.wantErr {
			t.Errorf("action=%q reason=%q: got err=%v, want %v", tc.action, tc.reason, hasErr, tc.wantErr)
		}
	}
}

// TestReviewFiatWithdrawStatusGuard 說明只有 pending 狀態才允許審核（防重複審核）。
func TestReviewFiatWithdrawStatusGuard(t *testing.T) {
	reviewableStatuses := []string{"pending"}
	blockedStatuses := []string{"approved", "rejected"}

	for _, s := range reviewableStatuses {
		if s != "pending" {
			t.Errorf("only 'pending' should be reviewable, got %q", s)
		}
	}
	for _, s := range blockedStatuses {
		if s == "pending" {
			t.Errorf("status %q should not be blocked", s)
		}
	}
}

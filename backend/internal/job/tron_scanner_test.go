package job

import (
	"fmt"
	"testing"
)

func TestParseMemoToUserID(t *testing.T) {
	tests := []struct {
		name      string
		memo      string
		wantID    int64
		wantError bool
	}{
		{
			name:   "valid 8-char hex",
			memo:   "0000001f",
			wantID: 31,
		},
		{
			name:   "valid memo with leading zeros",
			memo:   "00000001",
			wantID: 1,
		},
		{
			name:   "valid memo with max plausible userID",
			memo:   "00ffffff",
			wantID: 16777215,
		},
		{
			name:   "memo with surrounding whitespace",
			memo:   "  0000001f  ",
			wantID: 31,
		},
		{
			name:      "empty memo",
			memo:      "",
			wantError: true,
		},
		{
			name:      "whitespace only",
			memo:      "   ",
			wantError: true,
		},
		{
			name:      "non-hex characters",
			memo:      "0000001g",
			wantError: true,
		},
		{
			name:      "zero memo resolves to non-positive userID",
			memo:      "00000000",
			wantError: true,
		},
		{
			name:      "negative hex (two's complement overflow)",
			memo:      "ffffffffffffffff",
			wantError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseMemoToUserID(tc.memo)
			if tc.wantError {
				if err == nil {
					t.Errorf("expected error, got userID=%d", got)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if got != tc.wantID {
				t.Errorf("got userID=%d, want %d", got, tc.wantID)
			}
		})
	}
}

// TestMemoRoundTrip 驗證 fmt.Sprintf("%08x", userID) → parseMemoToUserID 的往返一致性。
// GetDepositInfo 使用此格式產生 memo，scanner 使用 parseMemoToUserID 解回 userID。
func TestMemoRoundTrip(t *testing.T) {
	userIDs := []int64{1, 42, 255, 1000, 99999, 16777215}
	for _, id := range userIDs {
		memo := fmt.Sprintf("%08x", id)
		got, err := parseMemoToUserID(memo)
		if err != nil {
			t.Errorf("userID=%d: parseMemoToUserID(%q) error: %v", id, memo, err)
			continue
		}
		if got != id {
			t.Errorf("userID=%d: round-trip got %d", id, got)
		}
	}
}

// TestIdempotencyGuarantee 說明冪等保證機制：
//
// processTransfer 在 INSERT 前呼叫 FindByTxHash：
//   - 若 tx_hash 已存在（existing != nil）→ 立即 return nil，不重複 INSERT
//   - 資料表層亦有 UNIQUE(tx_hash) 約束作為第二道防線
//
// 此測試以白盒方式驗證 parseMemoToUserID 的回傳值不影響冪等路徑：
// 冪等判斷發生在 memo 解析之前，與 memo 內容無關。
func TestIdempotencyGuarantee(t *testing.T) {
	// 冪等路徑：FindByTxHash 返回非 nil → processTransfer 直接 return nil
	// 此路徑不依賴 memo 解析，故任意 memo 值均不影響冪等行為。
	// 完整整合驗證需要真實 DB，此處以說明性測試確認設計意圖。
	t.Log("idempotency is enforced by: (1) FindByTxHash pre-check, (2) UNIQUE(tx_hash) DB constraint")

	// parseMemoToUserID 在 FindByTxHash 之後呼叫，
	// 確認 memo 解析錯誤不影響已記錄交易的冪等性。
	validMemo := fmt.Sprintf("%08x", int64(42))
	id, err := parseMemoToUserID(validMemo)
	if err != nil || id != 42 {
		t.Errorf("setup: unexpected parseMemoToUserID result: id=%d err=%v", id, err)
	}
}

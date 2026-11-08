package orderrepo

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestBuildStatusUpdateQueryGuardsFromStatus 迴歸測試：狀態轉換的 UPDATE 必須以 fromStatus 為條件。
// 若 WHERE 只比對 id，兩個並發請求（例如同時 Confirm 與 Cancel，或後台重複裁決）在 FindByID
// 都看到舊狀態後，會各自成功改寫狀態並各自執行一次錢包操作，造成重複放幣或解凍。
func TestBuildStatusUpdateQueryGuardsFromStatus(t *testing.T) {
	cases := []struct {
		name   string
		extras map[string]interface{}
	}{
		{name: "no extras", extras: nil},
		{name: "with extras", extras: map[string]interface{}{"confirmed_at": time.Now(), "completed_at": time.Now()}},
		{name: "unknown extras are ignored", extras: map[string]interface{}{"status": "hacked"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			const id int64 = 99
			query, args := buildStatusUpdateQuery(id, "paid", "completed", tc.extras)

			// 最後兩個參數依序是 id 與 fromStatus，WHERE 子句必須同時引用兩者。
			n := len(args)
			if n < 3 {
				t.Fatalf("expected at least 3 args, got %d: %v", n, args)
			}
			wantWhere := fmt.Sprintf(" WHERE id = $%d AND status = $%d", n-1, n)
			if !strings.HasSuffix(query, wantWhere) {
				t.Fatalf("query must end with %q, got %q", wantWhere, query)
			}
			if args[0] != "completed" {
				t.Fatalf("args[0] (toStatus) = %v, want completed", args[0])
			}
			if args[n-2] != id {
				t.Fatalf("args[%d] (id) = %v, want %d", n-2, args[n-2], id)
			}
			if args[n-1] != "paid" {
				t.Fatalf("args[%d] (fromStatus) = %v, want paid", n-1, args[n-1])
			}
			if strings.Count(query, "status = $") != 2 {
				t.Fatalf("expected exactly one SET status and one WHERE status placeholder, got %q", query)
			}
		})
	}
}

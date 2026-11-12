package orderrepo

import (
	"context"
	"fmt"
	"github.com/shopspring/decimal"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/model/entity"
)

// recordingSession 記錄實際送進交易 session 的 SQL 與參數；其餘方法沿用內嵌的 nil 介面（誤用即 panic）。
type recordingSession struct {
	sqlx.Session

	queries []string
	args    [][]interface{}
}

func (s *recordingSession) QueryRowCtx(_ context.Context, _ any, query string, args ...any) error {
	s.queries = append(s.queries, query)
	s.args = append(s.args, args)
	return nil
}

func normalizeSQL(q string) string {
	return strings.Join(strings.Fields(q), " ")
}

var lockClausePattern = regexp.MustCompile(`(?i)\bFOR\s+(UPDATE|SHARE|NO\s+KEY\s+UPDATE|KEY\s+SHARE)\b`)

// TestHasActiveOrdersByListingInTxIsPlainSelect 迴歸測試：查詢必須涵蓋所有非終態訂單，
// 且為一般 SELECT（不得加 FOR UPDATE / FOR SHARE）。呼叫端已持有掛單列的 FOR UPDATE 鎖，
// 這裡再鎖訂單列會與「先鎖訂單列、再鎖掛單列」的取消與逾時路徑形成相反鎖序而死鎖。
func TestHasActiveOrdersByListingInTxIsPlainSelect(t *testing.T) {
	s := &recordingSession{}
	if _, err := New(nil).HasActiveOrdersByListingInTx(context.Background(), s, 10); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.queries) != 1 {
		t.Fatalf("expected exactly 1 query on the session, got %d", len(s.queries))
	}
	q := normalizeSQL(s.queries[0])
	if lockClausePattern.MatchString(q) {
		t.Fatalf("query must not contain a lock clause, got %q", q)
	}
	want := "WHERE listing_id = $1 AND status IN ('matched', 'paid', 'releasing', 'disputed')"
	if !strings.Contains(q, want) {
		t.Fatalf("query must contain %q, got %q", want, q)
	}
	if s.args[0][0] != int64(10) {
		t.Fatalf("unexpected args: %v", s.args[0])
	}
}

// TestCreateInTxUsesSession 驗證建單寫入走呼叫端的交易 session（而非連線池），才能與凍結、扣量一起回滾。
func TestCreateInTxUsesSession(t *testing.T) {
	s := &recordingSession{}
	o := &entity.Order{OrderNo: "P2P1", ListingID: 10, CryptoAmount: decimal.RequireFromString("1.5"), Status: "matched"}
	if _, err := New(nil).CreateInTx(context.Background(), s, o); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.queries) != 1 || !strings.HasPrefix(normalizeSQL(s.queries[0]), "INSERT INTO orders") {
		t.Fatalf("expected one INSERT INTO orders on the session, got %v", s.queries)
	}
	if got := len(s.args[0]); got != 19 {
		t.Fatalf("expected 19 args, got %d", got)
	}
	if s.args[0][0] != "P2P1" || s.args[0][1] != int64(10) || !decimal.RequireFromString("1.5").Equal(s.args[0][7].(decimal.Decimal)) || s.args[0][17] != "matched" {
		t.Fatalf("unexpected args: %v", s.args[0])
	}
}

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

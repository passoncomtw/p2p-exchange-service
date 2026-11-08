package escrowrepo

import (
	"context"
	"strings"
	"testing"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
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

// TestCreateInTxUsesSession 迴歸測試：託管紀錄必須寫在建單的交易 session 內。
// legacy 在交易外寫入且忽略錯誤，建單失敗時會留下孤兒 lock 紀錄，成功時也可能缺紀錄。
func TestCreateInTxUsesSession(t *testing.T) {
	s := &recordingSession{}
	rec := &EscrowRecord{OrderID: 20, CryptoCurrency: "USDT", Amount: 3, Action: "lock", Status: "completed"}
	if err := New(nil).CreateInTx(context.Background(), s, rec); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.queries) != 1 || !strings.Contains(s.queries[0], "INSERT INTO escrow_records") {
		t.Fatalf("expected one INSERT INTO escrow_records on the session, got %v", s.queries)
	}
	want := []interface{}{int64(20), "USDT", 3.0, "lock", "completed"}
	for i, v := range want {
		if s.args[0][i] != v {
			t.Fatalf("args[%d] = %v, want %v", i, s.args[0][i], v)
		}
	}
}

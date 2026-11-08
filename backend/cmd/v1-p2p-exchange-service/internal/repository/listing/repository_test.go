package listingrepo

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// recordingSession 記錄實際送進交易 session 的 SQL 與參數，ExecCtx 回傳指定的影響列數。
// 其餘方法沿用內嵌的 nil 介面：走到未預期的路徑會直接 panic。
type recordingSession struct {
	sqlx.Session

	rowsAffected int64
	queries      []string
	args         [][]interface{}
}

func (s *recordingSession) QueryRowCtx(_ context.Context, _ any, query string, args ...any) error {
	s.queries = append(s.queries, query)
	s.args = append(s.args, args)
	return nil
}

func (s *recordingSession) ExecCtx(_ context.Context, query string, args ...any) (sql.Result, error) {
	s.queries = append(s.queries, query)
	s.args = append(s.args, args)
	return fakeResult{rows: s.rowsAffected}, nil
}

type fakeResult struct {
	rows int64
}

func (r fakeResult) LastInsertId() (int64, error) { return 0, nil }
func (r fakeResult) RowsAffected() (int64, error) { return r.rows, nil }

// normalizeSQL 把多個空白壓成一個，讓斷言不受縮排影響。
func normalizeSQL(q string) string {
	return strings.Join(strings.Fields(q), " ")
}

// TestFindByIDForUpdateInTxLocksRow 迴歸測試：建單、釋放訂單配置與取消掛單都靠這把列鎖序列化。
// 若漏掉 FOR UPDATE，兩筆並發吃單會各自通過剩餘量檢查而超賣，掛單取消也可能漏看同時建立的訂單。
func TestFindByIDForUpdateInTxLocksRow(t *testing.T) {
	s := &recordingSession{}
	if _, err := New(nil).FindByIDForUpdateInTx(context.Background(), s, 10); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.queries) != 1 {
		t.Fatalf("expected exactly 1 query on the session, got %d", len(s.queries))
	}
	q := normalizeSQL(s.queries[0])
	if !strings.HasSuffix(q, "FROM listings WHERE id = $1 FOR UPDATE") {
		t.Fatalf("query must lock the listing row with FOR UPDATE, got %q", q)
	}
	if s.args[0][0] != int64(10) {
		t.Fatalf("unexpected args: %v", s.args[0])
	}
}

// TestDeductAmountInTxGuardsStatusAndRemaining 迴歸測試：扣減剩餘量必須同時限制 status='active'
// 與 remaining_amount >= amount，並把影響列數交給呼叫端。legacy 的扣減在交易外執行且忽略錯誤，
// 已取消的掛單或剩餘量不足時仍會留下訂單。
func TestDeductAmountInTxGuardsStatusAndRemaining(t *testing.T) {
	for _, rows := range []int64{0, 1} {
		s := &recordingSession{rowsAffected: rows}
		n, err := New(nil).DeductAmountInTx(context.Background(), s, 10, 2.5)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if n != rows {
			t.Fatalf("expected rows affected %d to be returned, got %d", rows, n)
		}
		q := normalizeSQL(s.queries[0])
		for _, want := range []string{
			"UPDATE listings SET remaining_amount = remaining_amount - $1",
			"WHERE id = $2",
			"status = 'active'",
			"remaining_amount >= $1",
		} {
			if !strings.Contains(q, want) {
				t.Fatalf("query must contain %q, got %q", want, q)
			}
		}
		if s.args[0][0] != 2.5 || s.args[0][1] != int64(10) {
			t.Fatalf("unexpected args (want amount, id): %v", s.args[0])
		}
	}
}

// TestCancelInTxIsConditional 迴歸測試：取消掛單必須是條件式 UPDATE（只允許 active/paused），
// 影響列數交給呼叫端判斷；無條件 UPDATE 會讓同一掛單被重複取消、重複解凍。
func TestCancelInTxIsConditional(t *testing.T) {
	s := &recordingSession{rowsAffected: 0}
	n, err := New(nil).CancelInTx(context.Background(), s, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 0 {
		t.Fatalf("expected rows affected 0, got %d", n)
	}
	q := normalizeSQL(s.queries[0])
	if !strings.Contains(q, "SET status = 'cancelled'") || !strings.Contains(q, "WHERE id = $1 AND status IN ('active', 'paused')") {
		t.Fatalf("cancel must be a conditional update on active/paused, got %q", q)
	}
}

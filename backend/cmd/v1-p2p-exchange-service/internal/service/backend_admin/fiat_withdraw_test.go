package backend_admin_service

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"

	backend_interface "p2p-exchange/cmd/v1-p2p-exchange-service/internal/interfaces/backend"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/model/entity"
	fiatwithdrawrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/fiat_withdraw"
	walletrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/wallet"
	apierrors "p2p-exchange/internal/errors"
)

const (
	testWithdrawID   int64 = 42
	testWithdrawUser int64 = 1001
	testReviewerID   int64 = 7
	testAmount             = "5000.00"
)

// fakeConn 只實作 TransactCtx（直接執行 fn），其餘方法沿用內嵌的 nil 介面：
// 測試若不小心走到未預期的路徑會直接 panic，不會靜默通過。
// onRollback 用來模擬「交易失敗整批回滾」：真實 DB 會連同狀態轉換一起復原。
type fakeConn struct {
	sqlx.SqlConn
	onRollback func()
}

func (c fakeConn) TransactCtx(ctx context.Context, fn func(context.Context, sqlx.Session) error) error {
	err := fn(ctx, fakeSession{})
	if err != nil && c.onRollback != nil {
		c.onRollback()
	}
	return err
}

type fakeSession struct {
	sqlx.Session
}

// fakeFiatWithdrawRepo 以互斥鎖保護的 status 欄位模擬 DB 的條件式 UPDATE 語意：
// 只有看到 'pending' 的呼叫才會「異動 1 列」，其餘一律 0 列。
type fakeFiatWithdrawRepo struct {
	fiatwithdrawrepo.FiatWithdrawRepository

	mu     sync.Mutex
	status string

	findCalls    int
	findErr      error
	approveCalls int
	rejectCalls  int

	// reviewerIDs 記錄每次狀態轉換實際寫入的審核人，用來驗證審核人來源。
	reviewerIDs []int64
	reasons     []string
}

func (r *fakeFiatWithdrawRepo) FindByID(_ context.Context, id int64) (*entity.FiatWithdrawal, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.findCalls++
	if r.findErr != nil {
		return nil, r.findErr
	}
	return &entity.FiatWithdrawal{
		ID:       id,
		UserID:   testWithdrawUser,
		Currency: "TWD",
		Amount:   testAmount,
		Status:   r.status,
	}, nil
}

func (r *fakeFiatWithdrawRepo) UpdateApprovedInTx(_ context.Context, _ sqlx.Session, _, reviewerID int64, _ time.Time) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.approveCalls++
	if r.status != "pending" {
		return 0, nil
	}
	r.status = "approved"
	r.reviewerIDs = append(r.reviewerIDs, reviewerID)
	return 1, nil
}

func (r *fakeFiatWithdrawRepo) UpdateRejectedInTx(_ context.Context, _ sqlx.Session, _, reviewerID int64, _ time.Time, reason string) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rejectCalls++
	if r.status != "pending" {
		return 0, nil
	}
	r.status = "rejected"
	r.reviewerIDs = append(r.reviewerIDs, reviewerID)
	r.reasons = append(r.reasons, reason)
	return 1, nil
}

func (r *fakeFiatWithdrawRepo) currentStatus() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.status
}

func (r *fakeFiatWithdrawRepo) restorePending() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.status = "pending"
}

// fakeWalletRepo 記錄扣款與解凍被呼叫幾次，並累積等同 wallet_ledgers 的寫入紀錄，
// 用來驗證「資金只異動一次、帳本只多一筆」。
type fakeWalletRepo struct {
	walletrepo.WalletRepository

	mu            sync.Mutex
	deductCalls   int
	unfreezeCalls int
	ledgers       []string

	deductErr error
}

func (r *fakeWalletRepo) DeductFrozenBalanceWithLedgerTypeInTx(_ context.Context, _ sqlx.Session, _ int64, _, _, ledgerType string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if ledgerType != fiatWithdrawLedgerType {
		return errors.New("unexpected ledger type: " + ledgerType)
	}
	if r.deductErr != nil {
		return r.deductErr
	}
	r.deductCalls++
	r.ledgers = append(r.ledgers, ledgerType)
	return nil
}

func (r *fakeWalletRepo) UnfreezeBalanceInTx(_ context.Context, _ sqlx.Session, _ int64, _, _ string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.unfreezeCalls++
	r.ledgers = append(r.ledgers, "unfreeze")
	return nil
}

func (r *fakeWalletRepo) ledgerCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.ledgers)
}

func newTestService(repo *fakeFiatWithdrawRepo, wallet *fakeWalletRepo, onRollback func()) *backendAdminService {
	return &backendAdminService{
		db:               fakeConn{onRollback: onRollback},
		walletRepo:       wallet,
		fiatWithdrawRepo: repo,
	}
}

func errCode(t *testing.T, err error) int {
	t.Helper()
	var appErr *apierrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected *apierrors.AppError, got %T: %v", err, err)
	}
	return appErr.Code
}

// TestReviewApproveDeductsOnce 核可路徑：狀態轉為 approved 且自凍結餘額扣款一次。
func TestReviewApproveDeductsOnce(t *testing.T) {
	repo := &fakeFiatWithdrawRepo{status: "pending"}
	wallet := &fakeWalletRepo{}
	svc := newTestService(repo, wallet, nil)

	if err := svc.ReviewFiatWithdrawal(context.Background(), testReviewerID, testWithdrawID, "approve", ""); err != nil {
		t.Fatalf("approve: %v", err)
	}
	if repo.currentStatus() != "approved" {
		t.Fatalf("expected status approved, got %s", repo.currentStatus())
	}
	if wallet.deductCalls != 1 {
		t.Fatalf("expected exactly 1 deduct, got %d", wallet.deductCalls)
	}
	if wallet.unfreezeCalls != 0 {
		t.Fatalf("approve must not unfreeze, got %d", wallet.unfreezeCalls)
	}
}

// TestReviewRejectUnfreezesOnce 駁回路徑：狀態轉為 rejected、解凍一次並記錄原因。
func TestReviewRejectUnfreezesOnce(t *testing.T) {
	repo := &fakeFiatWithdrawRepo{status: "pending"}
	wallet := &fakeWalletRepo{}
	svc := newTestService(repo, wallet, nil)

	if err := svc.ReviewFiatWithdrawal(context.Background(), testReviewerID, testWithdrawID, "reject", "帳號戶名不符"); err != nil {
		t.Fatalf("reject: %v", err)
	}
	if repo.currentStatus() != "rejected" {
		t.Fatalf("expected status rejected, got %s", repo.currentStatus())
	}
	if wallet.unfreezeCalls != 1 {
		t.Fatalf("expected exactly 1 unfreeze, got %d", wallet.unfreezeCalls)
	}
	if wallet.deductCalls != 0 {
		t.Fatalf("reject must not deduct, got %d", wallet.deductCalls)
	}
	if len(repo.reasons) != 1 || repo.reasons[0] != "帳號戶名不符" {
		t.Fatalf("reject reason not persisted: %v", repo.reasons)
	}
}

// TestConcurrentApproveApproveMovesFundsOnce 是 P0-1 的迴歸測試（情境一：兩個都 approve）。
//
// 重現場景：後台人員連點兩次、前端重送、或兩個副本同時處理同一筆 pending 提領。
// legacy 先用無鎖的 FindByID 判斷 `w.Status != "pending"`（check-then-act 非原子），
// 再在交易外扣款，最後才用沒有 WHERE status 守衛的 UPDATE 改狀態 ——
// 兩次都會通過檢查並各扣一次款，使用者的凍結餘額被憑空多扣一次。
// 修正後：狀態轉換（WHERE status='pending'）與扣款在同一交易且狀態轉換在前，
// 敗者的 RowsAffected=0 直接回 409，完全不會呼叫扣款。
func TestConcurrentApproveApproveMovesFundsOnce(t *testing.T) {
	repo := &fakeFiatWithdrawRepo{status: "pending"}
	wallet := &fakeWalletRepo{}
	svc := newTestService(repo, wallet, nil)

	errs := runConcurrentReviews(t, svc, [2]string{"approve", "approve"})

	assertSingleWinner(t, errs)
	if repo.approveCalls != 2 {
		t.Fatalf("expected 2 status transition attempts, got %d", repo.approveCalls)
	}
	if wallet.deductCalls != 1 {
		t.Fatalf("double deduct: expected exactly 1 deduct across 2 concurrent approvals, got %d", wallet.deductCalls)
	}
	if got := wallet.ledgerCount(); got != 1 {
		t.Fatalf("expected exactly 1 wallet_ledgers row, got %d", got)
	}
}

// TestConcurrentApproveRejectMovesFundsOnce 是 P0-1 的迴歸測試（情境二：一個 approve 一個 reject）。
// legacy 會同時扣款又解凍（等於使用者白拿一次金額且提領仍被視為已撥款）；
// 修正後只有一方能通過狀態轉換，資金只異動一次。
func TestConcurrentApproveRejectMovesFundsOnce(t *testing.T) {
	repo := &fakeFiatWithdrawRepo{status: "pending"}
	wallet := &fakeWalletRepo{}
	svc := newTestService(repo, wallet, nil)

	errs := runConcurrentReviews(t, svc, [2]string{"approve", "reject"})

	assertSingleWinner(t, errs)
	if total := wallet.deductCalls + wallet.unfreezeCalls; total != 1 {
		t.Fatalf("expected exactly 1 fund movement, got deduct=%d unfreeze=%d", wallet.deductCalls, wallet.unfreezeCalls)
	}
	if got := wallet.ledgerCount(); got != 1 {
		t.Fatalf("expected exactly 1 wallet_ledgers row, got %d", got)
	}
	if s := repo.currentStatus(); s != "approved" && s != "rejected" {
		t.Fatalf("expected a terminal status, got %s", s)
	}
}

// runConcurrentReviews 讓兩個 goroutine 同時審核同一筆申請，回傳各自的錯誤。
func runConcurrentReviews(t *testing.T, svc *backendAdminService, actions [2]string) [2]error {
	t.Helper()

	var errs [2]error
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			<-start
			errs[idx] = svc.ReviewFiatWithdrawal(context.Background(), testReviewerID, testWithdrawID, actions[idx], "重複審核測試")
		}(i)
	}
	close(start)
	wg.Wait()
	return errs
}

// assertSingleWinner 斷言兩次併發審核恰有一次成功，另一次是 409 衝突（而非靜默成功）。
func assertSingleWinner(t *testing.T, errs [2]error) {
	t.Helper()

	successes, conflicts := 0, 0
	for _, err := range errs {
		if err == nil {
			successes++
			continue
		}
		if code := errCode(t, err); code == 409 {
			conflicts++
			continue
		}
		t.Fatalf("unexpected error: %v", err)
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("expected exactly 1 success + 1 conflict, got %d success / %d conflict", successes, conflicts)
	}
}

// TestReviewAlreadyReviewedReturnsConflict 已審核過的申請必須回 409，
// 且完全不得動用資金；審核是人為決策，不可當成冪等成功靜默吞掉。
func TestReviewAlreadyReviewedReturnsConflict(t *testing.T) {
	for _, status := range []string{"approved", "rejected"} {
		t.Run(status, func(t *testing.T) {
			repo := &fakeFiatWithdrawRepo{status: status}
			wallet := &fakeWalletRepo{}
			svc := newTestService(repo, wallet, nil)

			err := svc.ReviewFiatWithdrawal(context.Background(), testReviewerID, testWithdrawID, "approve", "")
			if err == nil {
				t.Fatal("expected conflict error for an already-reviewed withdrawal, got nil")
			}
			if code := errCode(t, err); code != 409 {
				t.Fatalf("expected 409, got %d (%v)", code, err)
			}
			if wallet.deductCalls != 0 || wallet.unfreezeCalls != 0 {
				t.Fatalf("no fund movement allowed: deduct=%d unfreeze=%d", wallet.deductCalls, wallet.unfreezeCalls)
			}
			if got := wallet.ledgerCount(); got != 0 {
				t.Fatalf("expected no wallet_ledgers row, got %d", got)
			}
		})
	}
}

// TestReviewRollsBackStatusWhenFundMoveFails 資金異動失敗時，狀態轉換必須連同回滾，
// 不能留下「狀態已改成 approved 但錢沒扣」的中間態。
func TestReviewRollsBackStatusWhenFundMoveFails(t *testing.T) {
	repo := &fakeFiatWithdrawRepo{status: "pending"}
	wallet := &fakeWalletRepo{deductErr: apierrors.New(400, "insufficient frozen balance")}
	svc := newTestService(repo, wallet, repo.restorePending)

	err := svc.ReviewFiatWithdrawal(context.Background(), testReviewerID, testWithdrawID, "approve", "")
	if err == nil {
		t.Fatal("expected error when deduct fails, got nil")
	}
	if code := errCode(t, err); code != 400 {
		t.Fatalf("expected the 400 to surface unchanged, got %d (%v)", code, err)
	}
	if repo.currentStatus() != "pending" {
		t.Fatalf("status must roll back to pending, got %s", repo.currentStatus())
	}
	if wallet.deductCalls != 0 {
		t.Fatalf("expected no successful deduct, got %d", wallet.deductCalls)
	}
}

// TestReviewValidatesActionAndReason 非法 action 與缺原因的駁回都必須在碰 DB 之前擋下。
func TestReviewValidatesActionAndReason(t *testing.T) {
	tests := []struct {
		name   string
		action string
		reason string
	}{
		{name: "未知 action", action: "delete", reason: ""},
		{name: "空 action", action: "", reason: ""},
		{name: "駁回未填原因", action: "reject", reason: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeFiatWithdrawRepo{status: "pending"}
			wallet := &fakeWalletRepo{}
			svc := newTestService(repo, wallet, nil)

			err := svc.ReviewFiatWithdrawal(context.Background(), testReviewerID, testWithdrawID, tt.action, tt.reason)
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
			if code := errCode(t, err); code != 400 {
				t.Fatalf("expected 400, got %d (%v)", code, err)
			}
			if repo.findCalls != 0 {
				t.Fatalf("validation must fail before touching the DB, findCalls=%d", repo.findCalls)
			}
			if wallet.deductCalls != 0 || wallet.unfreezeCalls != 0 {
				t.Fatalf("no fund movement allowed on validation failure")
			}
		})
	}
}

// TestReviewNotFoundReturns404 查無此申請時回 404，且不動用資金。
func TestReviewNotFoundReturns404(t *testing.T) {
	repo := &fakeFiatWithdrawRepo{status: "pending", findErr: sqlx.ErrNotFound}
	wallet := &fakeWalletRepo{}
	svc := newTestService(repo, wallet, nil)

	err := svc.ReviewFiatWithdrawal(context.Background(), testReviewerID, testWithdrawID, "approve", "")
	if code := errCode(t, err); code != 404 {
		t.Fatalf("expected 404, got %d (%v)", code, err)
	}
	if wallet.deductCalls != 0 {
		t.Fatalf("no fund movement allowed when withdrawal is missing")
	}
}

// TestReviewerIDComesFromCaller 寫入 fiat_withdrawals.reviewed_by 的審核人
// 必定是呼叫端（handler 從 JWT context 取得）傳入的值。
func TestReviewerIDComesFromCaller(t *testing.T) {
	repo := &fakeFiatWithdrawRepo{status: "pending"}
	wallet := &fakeWalletRepo{}
	svc := newTestService(repo, wallet, nil)

	const jwtReviewerID int64 = 99
	if err := svc.ReviewFiatWithdrawal(context.Background(), jwtReviewerID, testWithdrawID, "approve", ""); err != nil {
		t.Fatalf("approve: %v", err)
	}
	if len(repo.reviewerIDs) != 1 || repo.reviewerIDs[0] != jwtReviewerID {
		t.Fatalf("expected reviewed_by=%d, got %v", jwtReviewerID, repo.reviewerIDs)
	}
}

// TestReviewRequestHasNoReviewerField 審核請求結構不得有任何可承載審核人的欄位：
// 審核人只能來自 Backend JWT context，body 一旦能帶審核人就等於稽核紀錄可被偽造。
func TestReviewRequestHasNoReviewerField(t *testing.T) {
	typ := reflect.TypeOf(backend_interface.BackendReviewFiatWithdrawalRequest{})

	want := map[string]bool{"ID": true, "Action": true, "Reason": true}
	if typ.NumField() != len(want) {
		t.Fatalf("unexpected field count %d; request must carry only id/action/reason", typ.NumField())
	}
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		if !want[f.Name] {
			t.Fatalf("unexpected field %q on BackendReviewFiatWithdrawalRequest: reviewer info must come from JWT only", f.Name)
		}
	}
}

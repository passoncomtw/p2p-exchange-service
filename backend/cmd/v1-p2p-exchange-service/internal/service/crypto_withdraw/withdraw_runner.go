package cryptowithdraw_service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/config"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/model/entity"
	cryptowithdrawrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/crypto_withdraw"
	walletrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/wallet"
	"p2p-exchange/internal/infra/rdb"
	"p2p-exchange/pkg/tron"
)

const (
	// defaultWithdrawInterval 設定檔未給有效輪詢間隔時的預設值。
	defaultWithdrawInterval = 10 * time.Second
	// withdrawBatchSize 單輪最多處理的筆數（legacy 固定 50，這裡沿用保守上限）。
	withdrawBatchSize = 50

	// withdrawLockPrefix Redis 鎖前綴（與 legacy 相同的 key，兩邊併行時可互相排擠）。
	// 這個鎖只是減少多副本間無謂的重複嘗試，不是併發防線：
	// 真正的防線是 ClaimForBroadcastInTx 的條件式 UPDATE。
	withdrawLockPrefix = "lock:crypto_withdraw:"
	withdrawLockTTL    = 60 * time.Second

	// ledgerTypeCryptoWithdraw 鏈上提領的帳本類型（與 legacy 相同）。
	ledgerTypeCryptoWithdraw = "crypto_withdraw"
)

// TronWithdrawRunner 定時處理鏈上 USDT 提領：
//  1. broadcastPending：認領 pending 記錄並廣播上鏈。
//  2. confirmBroadcasting：對已達確認區塊數的 broadcasting 記錄完成確認與扣款。
//
// 資金完整性：
//   - 廣播前一定先用 DB 條件式 UPDATE 認領（pending → broadcasting）並檢查 RowsAffected，
//     認領不到就完全不碰鏈上 API。鏈上轉帳不可回滾，重複廣播等於熱錢包真實多付一次錢（P0-5）。
//   - 確認與扣款在同一個 DB 交易內完成，並以條件式 UPDATE 的 RowsAffected 作為唯一扣款依據，
//     重複觸發不會重複扣款（P0-3）。
type TronWithdrawRunner struct {
	db           sqlx.SqlConn
	rdb          *rdb.Client
	withdrawRepo cryptowithdrawrepo.CryptoWithdrawRepository
	walletRepo   walletrepo.WalletRepository
	cfg          *config.Config

	// broadcast 實際執行鏈上廣播；獨立成欄位是為了讓測試能在不觸網的情況下
	// 驗證「認領失敗時絕不呼叫鏈上 API」這條不可逆行為的守則。
	broadcast func(ctx context.Context, client *tron.Client, w *entity.CryptoWithdrawal) (txHash string, err error)
}

func NewTronWithdrawRunner(
	db sqlx.SqlConn,
	rdbClient *rdb.Client,
	withdrawRepo cryptowithdrawrepo.CryptoWithdrawRepository,
	walletRepo walletrepo.WalletRepository,
	cfg *config.Config,
) *TronWithdrawRunner {
	r := &TronWithdrawRunner{
		db:           db,
		rdb:          rdbClient,
		withdrawRepo: withdrawRepo,
		walletRepo:   walletRepo,
		cfg:          cfg,
	}
	r.broadcast = r.broadcastOnChain
	return r
}

func (r *TronWithdrawRunner) Name() string { return "tron-withdraw" }

// Run 依設定的間隔輪詢提領佇列，直到 ctx 結束。
// Tron 未設定時直接返回，不會留下空轉的 loop。
func (r *TronWithdrawRunner) Run(ctx context.Context) {
	conf := r.cfg.Tron
	if !conf.IsEnabled() {
		logx.Info("[tron-withdraw] disabled: hot wallet not configured")
		return
	}

	client := tron.NewClient(conf.TronGridURL, conf.TronGridAPIKey)
	interval := time.Duration(conf.WithdrawIntervalSeconds) * time.Second
	if interval <= 0 {
		interval = defaultWithdrawInterval
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.broadcastPending(ctx, client)
			r.confirmBroadcasting(ctx, client)
		}
	}
}

// broadcastPending 逐筆認領並廣播待處理的提領申請。
func (r *TronWithdrawRunner) broadcastPending(ctx context.Context, client *tron.Client) {
	pending, err := r.withdrawRepo.ListPending(ctx, withdrawBatchSize)
	if err != nil {
		logx.Errorf("[tron-withdraw] ListPending: %v", err)
		return
	}
	for _, w := range pending {
		r.processWithdrawal(ctx, client, w)
	}
}

// processWithdrawal 處理單筆提領：認領 →（成功才）廣播 → 記錄 tx_hash；廣播失敗則轉 failed 並解凍。
//
// 順序不可顛倒：DB 認領必須在任何鏈上呼叫之前完成。
func (r *TronWithdrawRunner) processWithdrawal(ctx context.Context, client *tron.Client, w *entity.CryptoWithdrawal) {
	// Redis 鎖僅為效能最佳化（減少多副本同時撞同一筆的無謂嘗試），
	// 取不到就讓別人處理；取得與否都不影響下面的 DB 認領是唯一防線這件事。
	if r.rdb != nil {
		unlock, err := r.rdb.AcquireLock(ctx, fmt.Sprintf("%s%d", withdrawLockPrefix, w.ID), withdrawLockTTL)
		if err != nil {
			return
		}
		defer unlock()
	}

	claimed, err := r.claimForBroadcast(ctx, w.ID)
	if err != nil {
		logx.Errorf("[tron-withdraw] claim id=%d: %v", w.ID, err)
		return
	}
	if !claimed {
		// 已被其他流程（其他副本／legacy job／上一輪）認領：絕不重複廣播。
		logx.Infof("[tron-withdraw] id=%d already claimed, skip", w.ID)
		return
	}

	txHash, err := r.broadcast(ctx, client, w)
	if err != nil {
		logx.Errorf("[tron-withdraw] broadcast id=%d: %v", w.ID, err)
		r.markFailedAndUnfreeze(ctx, w)
		return
	}

	if err := r.recordBroadcasted(ctx, w.ID, txHash, time.Now()); err != nil {
		// 鏈上已送出但 tx_hash 沒寫進 DB：狀態仍是 broadcasting，
		// 不會被再次認領（不會重複付款），但也無法自動確認，需人工對帳補上 tx_hash。
		logx.Errorf("[tron-withdraw] MANUAL ACTION REQUIRED: broadcast succeeded but recording failed id=%d txHash=%s: %v",
			w.ID, txHash, err)
		return
	}
	if txHash == "" {
		logx.Errorf("[tron-withdraw] MANUAL ACTION REQUIRED: broadcast succeeded without txID id=%d", w.ID)
		return
	}
	logx.Infof("[tron-withdraw] broadcast id=%d txHash=%s", w.ID, txHash)
}

// claimForBroadcast 以條件式 UPDATE（WHERE status='pending'）原子認領一筆待廣播記錄。
// 回傳 true 才代表本執行緒取得了「唯一一次」廣播該筆提領的授權。
func (r *TronWithdrawRunner) claimForBroadcast(ctx context.Context, id int64) (bool, error) {
	var affected int64
	err := r.db.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		var err error
		affected, err = r.withdrawRepo.ClaimForBroadcastInTx(ctx, session, id)
		return err
	})
	if err != nil {
		return false, err
	}
	return affected == 1, nil
}

// recordBroadcasted 在廣播成功後補寫 tx_hash 與 broadcast_at。
func (r *TronWithdrawRunner) recordBroadcasted(ctx context.Context, id int64, txHash string, broadcastAt time.Time) error {
	return r.db.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		return r.withdrawRepo.UpdateBroadcastedInTx(ctx, session, id, txHash, broadcastAt)
	})
}

// markFailedAndUnfreeze 廣播失敗時把已認領的記錄轉為 failed 並解凍餘額。
//
// 轉 failed 帶 WHERE status='broadcasting' 守衛並檢查 RowsAffected：
// 只有真的完成 broadcasting → failed 的執行緒才會解凍，避免重複解凍憑空放大可用餘額。
// 解凍走交易外的 UnfreezeBalance（自帶錢包鎖與 frozen_balance 守衛）：
// 此時狀態已是 failed，不會再被其他流程推進；解凍失敗只會讓資金暫時卡在凍結餘額，
// 不會造成資金憑空產生，記錄錯誤後由人工處理。
func (r *TronWithdrawRunner) markFailedAndUnfreeze(ctx context.Context, w *entity.CryptoWithdrawal) {
	var affected int64
	if err := r.db.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		var err error
		affected, err = r.withdrawRepo.UpdateFailedInTx(ctx, session, w.ID)
		return err
	}); err != nil {
		logx.Errorf("[tron-withdraw] UpdateFailed id=%d: %v", w.ID, err)
		return
	}
	if affected == 0 {
		logx.Infof("[tron-withdraw] id=%d not in broadcasting, skip unfreeze", w.ID)
		return
	}
	if err := r.walletRepo.UnfreezeBalance(ctx, w.UserID, w.Currency, w.Amount); err != nil {
		logx.Errorf("[tron-withdraw] MANUAL ACTION REQUIRED: unfreeze failed id=%d user=%d amount=%s: %v",
			w.ID, w.UserID, w.Amount, err)
		return
	}
	logx.Infof("[tron-withdraw] failed id=%d user=%d amount=%s unfrozen", w.ID, w.UserID, w.Amount)
}

// broadcastOnChain 組出 TRC-20 轉帳交易、以熱錢包私鑰簽章並廣播，回傳鏈上 txID。
// 只在認領成功後才會被呼叫。
func (r *TronWithdrawRunner) broadcastOnChain(ctx context.Context, client *tron.Client, w *entity.CryptoWithdrawal) (string, error) {
	conf := r.cfg.Tron

	sunAmount, err := tron.USDTToSun(w.Amount)
	if err != nil {
		return "", fmt.Errorf("USDTToSun: %w", err)
	}

	trigger, rawDataHex, err := client.TriggerTRC20Transfer(ctx,
		conf.HotWalletAddress, w.ToAddress, conf.USDTContractAddress, sunAmount)
	if err != nil {
		return "", fmt.Errorf("TriggerTRC20Transfer: %w", err)
	}

	signature, err := tron.SignRawDataHex(rawDataHex, conf.HotWalletPrivateKey)
	if err != nil {
		return "", fmt.Errorf("SignRawDataHex: %w", err)
	}

	if err := client.BroadcastTransaction(ctx, trigger.Transaction, signature); err != nil {
		return "", fmt.Errorf("BroadcastTransaction: %w", err)
	}

	// 交易已送出：之後的錯誤都不能再回傳 error（會被當成廣播失敗而解凍餘額）。
	txHash, err := extractTxID(trigger.Transaction)
	if err != nil {
		logx.Errorf("[tron-withdraw] extractTxID id=%d: %v", w.ID, err)
		return "", nil
	}
	return txHash, nil
}

// confirmBroadcasting 對達到確認區塊數的 broadcasting 記錄完成確認與扣款。
func (r *TronWithdrawRunner) confirmBroadcasting(ctx context.Context, client *tron.Client) {
	broadcasting, err := r.withdrawRepo.ListBroadcasting(ctx, withdrawBatchSize)
	if err != nil {
		logx.Errorf("[tron-withdraw] ListBroadcasting: %v", err)
		return
	}
	if len(broadcasting) == 0 {
		return
	}

	currentBlock, err := client.GetCurrentBlockNumber(ctx)
	if err != nil {
		logx.Errorf("[tron-withdraw] GetCurrentBlockNumber: %v", err)
		return
	}

	for _, w := range broadcasting {
		if w.TxHash == nil {
			continue
		}
		blockNumber, _, err := client.GetTransactionDetail(ctx, *w.TxHash)
		if err != nil || blockNumber == 0 {
			continue
		}
		if (currentBlock - blockNumber) < int64(r.cfg.Tron.ConfirmationBlocks) {
			continue
		}

		if err := r.confirmAndDeduct(ctx, w, time.Now()); err != nil {
			logx.Errorf("[tron-withdraw] confirm+deduct id=%d: %v", w.ID, err)
			continue
		}
		logx.Infof("[tron-withdraw] confirmed id=%d user=%d amount=%s", w.ID, w.UserID, w.Amount)
	}
}

// confirmAndDeduct 在同一個交易內完成 broadcasting → confirmed 的狀態轉換與凍結餘額扣款。
//
// 條件式 UPDATE（WHERE status='broadcasting'）的 RowsAffected 是唯一的扣款依據：
// 確認流程被重複觸發（Job 重啟、重複輪詢、legacy/v1 併行）時，第二次的 affected 為 0，
// 直接冪等結束，不會重複扣款；狀態轉換與扣款同屬一個交易，也不會有只成功一半的中間態。
func (r *TronWithdrawRunner) confirmAndDeduct(ctx context.Context, w *entity.CryptoWithdrawal, confirmedAt time.Time) error {
	return r.db.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		affected, err := r.withdrawRepo.ConfirmInTx(ctx, session, w.ID, confirmedAt)
		if err != nil {
			return err
		}
		if affected == 0 {
			return nil
		}
		return r.walletRepo.DeductFrozenBalanceWithLedgerTypeInTx(ctx, session,
			w.UserID, w.Currency, w.Amount, ledgerTypeCryptoWithdraw)
	})
}

// extractTxID 從鏈上回傳的原始交易 JSON 取出 txID 欄位。
// 改用標準 encoding/json（legacy 為手刻字串搜尋），欄位缺漏時回傳空字串由呼叫端記錄。
func extractTxID(txJSON json.RawMessage) (string, error) {
	var tx struct {
		TxID string `json:"txID"`
	}
	if err := json.Unmarshal(txJSON, &tx); err != nil {
		return "", err
	}
	return tx.TxID, nil
}

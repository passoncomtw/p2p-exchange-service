package cryptodeposit_service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/config"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/model/entity"
	cryptodepositrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/crypto_deposit"
	userrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/user"
	walletrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/wallet"
	"p2p-exchange/internal/infra/rdb"
	"p2p-exchange/pkg/tron"
)

const (
	// tronScannerLastTsKey 掃描游標（毫秒時間戳）的 Redis key。
	// 刻意與 legacy job 使用完全相同的 key：兩邊共用同一組游標，
	// 部署上任一時刻只會有一邊在跑，切換時不會重掃或漏掃。
	tronScannerLastTsKey = "tron:scanner:last_ts"
	// tronScannerLookbackMs 游標不存在時的回溯區間（5 分鐘）。
	tronScannerLookbackMs = int64(5 * 60 * 1000)
	// tronScannerCursorTTL 游標的保存期限。
	tronScannerCursorTTL = 30 * 24 * time.Hour

	// defaultScanInterval 設定檔未給有效掃描間隔時的預設值。
	defaultScanInterval = 30 * time.Second
	// pendingBatchSize 單輪最多處理的待確認筆數（legacy 無上限，這裡補一個保守上限）。
	pendingBatchSize = 100

	usdtCurrency = "USDT"
	// ledgerTypeCryptoDeposit 鏈上充值的帳本類型（與 legacy 相同）。
	ledgerTypeCryptoDeposit = "crypto_deposit"

	statusPending   = "pending"
	statusConfirmed = "confirmed"
)

// TronScannerRunner 定時掃描平台熱錢包的 TRC-20 入帳交易：
// 1. scanDeposits：抓取新的鏈上轉入，比對 memo 找出使用者並建立充值記錄。
// 2. confirmPendingDeposits：對已達確認區塊數的 pending 記錄完成確認與入帳。
//
// 資金完整性：任何「狀態轉換 + 入帳」都在同一個 DB 交易內完成，
// 並以條件式 UPDATE 的 RowsAffected 作為唯一入帳依據，重複觸發不會重複入帳。
type TronScannerRunner struct {
	db          sqlx.SqlConn
	rdb         *rdb.Client
	depositRepo cryptodepositrepo.CryptoDepositRepository
	walletRepo  walletrepo.WalletRepository
	userRepo    userrepo.UserRepository
	cfg         *config.Config
}

func NewTronScannerRunner(
	db sqlx.SqlConn,
	rdbClient *rdb.Client,
	depositRepo cryptodepositrepo.CryptoDepositRepository,
	walletRepo walletrepo.WalletRepository,
	userRepo userrepo.UserRepository,
	cfg *config.Config,
) *TronScannerRunner {
	return &TronScannerRunner{
		db:          db,
		rdb:         rdbClient,
		depositRepo: depositRepo,
		walletRepo:  walletRepo,
		userRepo:    userRepo,
		cfg:         cfg,
	}
}

func (r *TronScannerRunner) Name() string { return "tron-scanner" }

// Run 依設定的間隔輪詢鏈上交易，直到 ctx 結束。
// Tron 未設定時直接返回，不會留下空轉的 loop。
func (r *TronScannerRunner) Run(ctx context.Context) {
	conf := r.cfg.Tron
	if !conf.IsEnabled() {
		logx.Info("[tron-scanner] disabled: hot wallet not configured")
		return
	}

	client := tron.NewClient(conf.TronGridURL, conf.TronGridAPIKey)
	interval := time.Duration(conf.ScanIntervalSeconds) * time.Second
	if interval <= 0 {
		interval = defaultScanInterval
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.scanDeposits(ctx, client)
			r.confirmPendingDeposits(ctx, client)
		}
	}
}

// scanDeposits 從游標時間起抓取熱錢包的入帳交易，逐筆建檔，最後推進游標。
func (r *TronScannerRunner) scanDeposits(ctx context.Context, client *tron.Client) {
	conf := r.cfg.Tron
	minTs := r.lastScanTsMs(ctx)

	transfers, err := client.GetIncomingTRC20Transfers(ctx, conf.HotWalletAddress, conf.USDTContractAddress, minTs)
	if err != nil {
		logx.Errorf("[tron-scanner] GetIncomingTRC20Transfers error: %v", err)
		return
	}

	maxTs := minTs
	for _, tx := range transfers {
		if tx.BlockTimestamp > maxTs {
			maxTs = tx.BlockTimestamp
		}
		if err := r.processTransfer(ctx, client, tx); err != nil {
			logx.Errorf("[tron-scanner] processTransfer %s error: %v", tx.TransactionID, err)
		}
	}

	if maxTs > minTs {
		r.setLastScanTs(ctx, maxTs+1)
	}
}

// processTransfer 處理單筆鏈上轉入：去重、解析 memo、建立充值記錄，
// 若建立當下已達確認區塊數，則在同一交易內完成入帳。
func (r *TronScannerRunner) processTransfer(ctx context.Context, client *tron.Client, tx tron.TRC20Transfer) error {
	// 主要去重路徑：寫入前先查既有記錄（DB 的 tx_hash UNIQUE 是最後一道防線）。
	existing, err := r.depositRepo.FindByTxHash(ctx, tx.TransactionID)
	if err != nil && !errors.Is(err, sqlx.ErrNotFound) {
		return fmt.Errorf("FindByTxHash: %w", err)
	}
	if existing != nil {
		return nil
	}

	amountFloat, ok := tron.USDTAmountFromSun(tx.Value)
	if !ok {
		logx.Errorf("[tron-scanner] invalid value %s in tx %s", tx.Value, tx.TransactionID)
		return nil
	}
	if amountFloat.Sign() <= 0 {
		logx.Errorf("[tron-scanner] non-positive value %s in tx %s", tx.Value, tx.TransactionID)
		return nil
	}
	amountStr := amountFloat.Text('f', 18)

	blockNumber, memo, err := client.GetTransactionDetail(ctx, tx.TransactionID)
	if err != nil {
		logx.Errorf("[tron-scanner] GetTransactionDetail %s: %v", tx.TransactionID, err)
	}

	// memo 無法對應到使用者時只記錄不入帳（人工處理），不建立無主的充值記錄。
	userID, parseErr := parseMemoToUserID(memo)
	if parseErr != nil {
		logx.Infof("[tron-scanner] tx %s: unmatched memo %q – skipping credit", tx.TransactionID, memo)
		return nil
	}
	if _, err := r.userRepo.FindByID(ctx, userID); err != nil {
		logx.Infof("[tron-scanner] tx %s: user %d not found", tx.TransactionID, userID)
		return nil
	}

	status := statusPending
	if blockNumber > 0 {
		currentBlock, err := client.GetCurrentBlockNumber(ctx)
		if err == nil && (currentBlock-blockNumber) >= int64(r.cfg.Tron.ConfirmationBlocks) {
			status = statusConfirmed
		}
	}

	memoStr := memo
	deposit := &entity.CryptoDeposit{
		UserID:      userID,
		Currency:    usdtCurrency,
		Amount:      amountStr,
		TxHash:      tx.TransactionID,
		FromAddress: tx.From,
		Memo:        &memoStr,
		Status:      status,
	}

	if status != statusConfirmed {
		created, err := r.depositRepo.Create(ctx, deposit)
		if err != nil {
			return fmt.Errorf("create deposit: %w", err)
		}
		logx.Infof("[tron-scanner] recorded deposit id=%d user=%d amount=%s status=%s",
			created.ID, created.UserID, created.Amount, created.Status)
		return nil
	}

	// 建立當下就已確認：建檔與入帳必須原子。
	confirmedAt := time.Now()
	deposit.ConfirmedAt = &confirmedAt

	createdID, err := r.createConfirmedDeposit(ctx, deposit)
	if err != nil {
		// tx_hash 撞 UNIQUE：同一筆鏈上交易已被處理過，整筆交易已回滾（沒有重複入帳）。
		if errors.Is(err, cryptodepositrepo.ErrDuplicateTxHash) {
			logx.Infof("[tron-scanner] tx %s already recorded, skip", tx.TransactionID)
			return nil
		}
		return fmt.Errorf("create and credit deposit: %w", err)
	}

	logx.Infof("[tron-scanner] recorded deposit id=%d user=%d amount=%s status=%s",
		createdID, userID, amountStr, statusConfirmed)
	return nil
}

// createConfirmedDeposit 在同一個交易內建立「已確認」的充值記錄並入帳。
// 拆成單一交易是為了避免 legacy 的缺陷模式：建檔成功後才在交易外入帳，
// 中途失敗會留下已確認卻未入帳的記錄，重跑時又會重複入帳。
func (r *TronScannerRunner) createConfirmedDeposit(ctx context.Context, deposit *entity.CryptoDeposit) (int64, error) {
	var createdID int64
	err := r.db.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		created, err := r.depositRepo.CreateInTx(ctx, session, deposit)
		if err != nil {
			return err
		}
		createdID = created.ID
		return r.walletRepo.DepositWithLedgerTypeInTx(ctx, session,
			created.UserID, created.Currency, created.Amount, ledgerTypeCryptoDeposit, "")
	})
	if err != nil {
		return 0, err
	}
	return createdID, nil
}

// confirmPendingDeposits 對達到確認區塊數的 pending 記錄完成 pending → confirmed 與入帳。
func (r *TronScannerRunner) confirmPendingDeposits(ctx context.Context, client *tron.Client) {
	pending, err := r.depositRepo.ListPending(ctx, pendingBatchSize)
	if err != nil {
		logx.Errorf("[tron-scanner] ListPending: %v", err)
		return
	}
	if len(pending) == 0 {
		return
	}

	currentBlock, err := client.GetCurrentBlockNumber(ctx)
	if err != nil {
		logx.Errorf("[tron-scanner] GetCurrentBlockNumber: %v", err)
		return
	}

	for _, d := range pending {
		blockNumber, _, err := client.GetTransactionDetail(ctx, d.TxHash)
		if err != nil || blockNumber == 0 {
			continue
		}
		if (currentBlock - blockNumber) < int64(r.cfg.Tron.ConfirmationBlocks) {
			continue
		}

		if err := r.confirmAndCredit(ctx, d, time.Now()); err != nil {
			logx.Errorf("[tron-scanner] confirm+credit deposit %d: %v", d.ID, err)
			continue
		}
		logx.Infof("[tron-scanner] confirmed deposit id=%d user=%d amount=%s", d.ID, d.UserID, d.Amount)
	}
}

// confirmAndCredit 在同一個交易內完成 pending → confirmed 的狀態轉換與入帳。
//
// 條件式 UPDATE（WHERE status='pending'）的 RowsAffected 是唯一的入帳依據：
// 確認流程被重複觸發（Job 重啟、重複輪詢、legacy/v1 併行）時，第二次的 affected 為 0，
// 直接冪等結束，不會重複入帳；狀態轉換與入帳同屬一個交易，也不會有只成功一半的中間態。
func (r *TronScannerRunner) confirmAndCredit(ctx context.Context, d *entity.CryptoDeposit, confirmedAt time.Time) error {
	return r.db.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		affected, err := r.depositRepo.ConfirmInTx(ctx, session, d.ID, confirmedAt)
		if err != nil {
			return err
		}
		if affected == 0 {
			return nil
		}
		return r.walletRepo.DepositWithLedgerTypeInTx(ctx, session,
			d.UserID, d.Currency, d.Amount, ledgerTypeCryptoDeposit, "")
	})
}

// lastScanTsMs 讀取掃描游標；Redis 未設定或游標不存在時回溯固定區間。
func (r *TronScannerRunner) lastScanTsMs(ctx context.Context) int64 {
	fallback := time.Now().UnixMilli() - tronScannerLookbackMs
	if r.rdb == nil {
		return fallback
	}
	val, found, err := r.rdb.Get(ctx, tronScannerLastTsKey)
	if err != nil || !found {
		return fallback
	}
	ts, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return fallback
	}
	return ts
}

// setLastScanTs 推進掃描游標。
func (r *TronScannerRunner) setLastScanTs(ctx context.Context, tsMs int64) {
	if r.rdb == nil {
		return
	}
	if err := r.rdb.Set(ctx, tronScannerLastTsKey, strconv.FormatInt(tsMs, 10), tronScannerCursorTTL); err != nil {
		logx.Errorf("[tron-scanner] set cursor: %v", err)
	}
}

// parseMemoToUserID 把鏈上 memo（8 位十六進位）解析回 userID，與 depositMemo 對應。
func parseMemoToUserID(memo string) (int64, error) {
	memo = strings.TrimSpace(memo)
	if memo == "" {
		return 0, fmt.Errorf("empty memo")
	}
	userID, err := strconv.ParseInt(memo, 16, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid hex memo %q: %w", memo, err)
	}
	if userID <= 0 {
		return 0, fmt.Errorf("memo %q parsed to non-positive userID %d", memo, userID)
	}
	return userID, nil
}

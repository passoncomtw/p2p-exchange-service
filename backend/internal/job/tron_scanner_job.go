package job

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"p2p-exchange/internal/config"
	"p2p-exchange/internal/infra/rdb"
	"p2p-exchange/internal/model"
	"p2p-exchange/pkg/tron"
)

const (
	tronScannerLastTsKey = "tron:scanner:last_ts"
	// lookback scans up to 5 minutes in the past on first run
	tronScannerLookbackMs = int64(5 * 60 * 1000)
)

type TronScannerDeps struct {
	DB             sqlx.SqlConn
	RDB            *rdb.Client
	CryptoDeposit  *model.CryptoDepositModel
	AppUser        *model.AppUserModel
	Wallet         *model.WalletModel
}

// StartTronScannerJob polls TronGrid for incoming USDT deposits and credits user wallets.
// Runs in a goroutine; cancelled via ctx.
func StartTronScannerJob(ctx context.Context, conf config.TronConf, deps TronScannerDeps) {
	if !conf.IsEnabled() {
		logx.Info("[tron-scanner] disabled: hot wallet not configured")
		return
	}

	client := tron.NewClient(conf.TronGridURL, conf.TronGridAPIKey)
	interval := time.Duration(conf.ScanIntervalSeconds) * time.Second

	go func() {
		logx.Infof("[tron-scanner] started (interval=%s, network=%s)", interval, conf.Network)
		for {
			select {
			case <-ctx.Done():
				logx.Info("[tron-scanner] stopped")
				return
			case <-time.After(interval):
				scanDeposits(ctx, conf, client, deps)
				confirmPendingDeposits(ctx, conf, client, deps)
			}
		}
	}()
}

func scanDeposits(ctx context.Context, conf config.TronConf, client *tron.Client, deps TronScannerDeps) {
	minTs := tronScannerLastTsMs(ctx, deps.RDB)

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
		if err := processTransfer(ctx, conf, client, tx, deps); err != nil {
			logx.Errorf("[tron-scanner] processTransfer %s error: %v", tx.TransactionID, err)
		}
	}

	if maxTs > minTs {
		setTronScannerLastTs(ctx, deps.RDB, maxTs+1)
	}
}

func processTransfer(ctx context.Context, conf config.TronConf, client *tron.Client, tx tron.TRC20Transfer, deps TronScannerDeps) error {
	// Skip if already recorded
	existing, err := deps.CryptoDeposit.FindByTxHash(ctx, tx.TransactionID)
	if err != nil && err != sqlx.ErrNotFound {
		return fmt.Errorf("FindByTxHash: %w", err)
	}
	if existing != nil {
		return nil
	}

	// Decode amount: USDT has 6 decimals
	amountFloat, ok := tron.USDTAmountFromSun(tx.Value)
	if !ok {
		logx.Errorf("[tron-scanner] invalid value %s in tx %s", tx.Value, tx.TransactionID)
		return nil
	}
	amountStr := amountFloat.Text('f', 18)

	// Get transaction detail for memo and block number
	blockNumber, memo, err := client.GetTransactionDetail(ctx, tx.TransactionID)
	if err != nil {
		logx.Errorf("[tron-scanner] GetTransactionDetail %s: %v", tx.TransactionID, err)
	}

	// Match memo to user (8-char hex of user_id)
	memo = strings.TrimSpace(memo)
	userID, parseErr := strconv.ParseInt(memo, 16, 64)
	if parseErr != nil || userID <= 0 {
		logx.Infof("[tron-scanner] tx %s: unmatched memo %q – skipping credit", tx.TransactionID, memo)
		return nil
	}

	// Validate user exists
	_, err = deps.AppUser.FindByID(ctx, userID)
	if err != nil {
		logx.Infof("[tron-scanner] tx %s: user %d not found", tx.TransactionID, userID)
		return nil
	}

	// Determine status based on block confirmations
	status := "pending"
	if blockNumber > 0 {
		currentBlock, err := client.GetCurrentBlockNumber(ctx)
		if err == nil && (currentBlock-blockNumber) >= int64(conf.ConfirmationBlocks) {
			status = "confirmed"
		}
	}

	memoStr := memo
	deposit := &model.CryptoDeposit{
		UserID:      userID,
		Currency:    "USDT",
		Amount:      amountStr,
		TxHash:      tx.TransactionID,
		FromAddress: tx.From,
		Memo:        &memoStr,
		Status:      status,
	}
	if err := deps.CryptoDeposit.Create(ctx, deposit); err != nil {
		return fmt.Errorf("Create deposit: %w", err)
	}

	if status == "confirmed" {
		if err := creditDeposit(ctx, deposit, deps); err != nil {
			logx.Errorf("[tron-scanner] credit deposit %d: %v", deposit.ID, err)
		}
	}
	logx.Infof("[tron-scanner] recorded deposit id=%d user=%d amount=%s status=%s", deposit.ID, userID, amountStr, status)
	return nil
}

func confirmPendingDeposits(ctx context.Context, conf config.TronConf, client *tron.Client, deps TronScannerDeps) {
	pending, err := deps.CryptoDeposit.ListPending(ctx)
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
		if (currentBlock - blockNumber) < int64(conf.ConfirmationBlocks) {
			continue
		}

		now := time.Now()
		if err := deps.CryptoDeposit.UpdateConfirmed(ctx, d.ID, now); err != nil {
			logx.Errorf("[tron-scanner] UpdateConfirmed %d: %v", d.ID, err)
			continue
		}
		if err := creditDeposit(ctx, d, deps); err != nil {
			logx.Errorf("[tron-scanner] credit confirmed deposit %d: %v", d.ID, err)
		} else {
			logx.Infof("[tron-scanner] confirmed deposit id=%d user=%d amount=%s", d.ID, d.UserID, d.Amount)
		}
	}
}

// creditDeposit adds USDT to user's available balance and records a crypto_deposit ledger entry.
func creditDeposit(ctx context.Context, d *model.CryptoDeposit, deps TronScannerDeps) error {
	// Validate amount is positive
	f, _, err := new(big.Float).Parse(d.Amount, 10)
	if err != nil || f.Sign() <= 0 {
		return fmt.Errorf("invalid amount: %s", d.Amount)
	}
	return deps.Wallet.DepositWithLedgerType(ctx, d.UserID, d.Currency, d.Amount, "crypto_deposit")
}

func tronScannerLastTsMs(ctx context.Context, rdbClient *rdb.Client) int64 {
	fallback := time.Now().UnixMilli() - tronScannerLookbackMs
	if rdbClient == nil {
		return fallback
	}
	val, found, err := rdbClient.Get(ctx, tronScannerLastTsKey)
	if err != nil || !found {
		return fallback
	}
	ts, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return fallback
	}
	return ts
}

func setTronScannerLastTs(ctx context.Context, rdbClient *rdb.Client, tsMs int64) {
	if rdbClient == nil {
		return
	}
	_ = rdbClient.Set(ctx, tronScannerLastTsKey, strconv.FormatInt(tsMs, 10), 30*24*time.Hour)
}

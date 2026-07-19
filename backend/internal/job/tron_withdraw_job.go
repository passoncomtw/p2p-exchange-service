package job

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"p2p-exchange/internal/config"
	"p2p-exchange/internal/infra/rdb"
	"p2p-exchange/internal/model"
	"p2p-exchange/pkg/tron"
)

type TronWithdrawDeps struct {
	RDB            *rdb.Client
	CryptoWithdraw *model.CryptoWithdrawalModel
	Wallet         *model.WalletModel
}

// StartTronWithdrawJob broadcasts pending USDT withdrawals and confirms completed ones.
func StartTronWithdrawJob(ctx context.Context, conf config.TronConf, deps TronWithdrawDeps) {
	if !conf.IsEnabled() {
		logx.Info("[tron-withdraw] disabled: hot wallet not configured")
		return
	}

	client := tron.NewClient(conf.TronGridURL, conf.TronGridAPIKey)
	interval := time.Duration(conf.WithdrawIntervalSeconds) * time.Second

	go func() {
		logx.Infof("[tron-withdraw] started (interval=%s)", interval)
		for {
			select {
			case <-ctx.Done():
				logx.Info("[tron-withdraw] stopped")
				return
			case <-time.After(interval):
				broadcastPending(ctx, conf, client, deps)
				confirmBroadcasting(ctx, conf, client, deps)
			}
		}
	}()
}

func broadcastPending(ctx context.Context, conf config.TronConf, client *tron.Client, deps TronWithdrawDeps) {
	pending, err := deps.CryptoWithdraw.ListPending(ctx)
	if err != nil {
		logx.Errorf("[tron-withdraw] ListPending: %v", err)
		return
	}

	for _, w := range pending {
		processSingleWithdrawal(ctx, conf, client, w, deps)
	}
}

func processSingleWithdrawal(ctx context.Context, conf config.TronConf, client *tron.Client, w *model.CryptoWithdrawal, deps TronWithdrawDeps) {
	lockKey := fmt.Sprintf("lock:crypto_withdraw:%d", w.ID)
	if deps.RDB != nil {
		unlock, err := deps.RDB.AcquireLock(ctx, lockKey, 60*time.Second)
		if err != nil {
			return // already being processed
		}
		defer unlock()
	}
	if err := broadcastWithdrawal(ctx, conf, client, w, deps); err != nil {
		logx.Errorf("[tron-withdraw] broadcast id=%d: %v", w.ID, err)
		_ = deps.CryptoWithdraw.UpdateFailed(ctx, w.ID)
		_ = deps.Wallet.UnfreezeBalance(ctx, w.UserID, w.Currency, w.Amount)
	}
}

func broadcastWithdrawal(ctx context.Context, conf config.TronConf, client *tron.Client, w *model.CryptoWithdrawal, deps TronWithdrawDeps) error {
	sunAmount, err := tron.USDTToSun(w.Amount)
	if err != nil {
		return err
	}

	trigger, rawDataHex, err := client.TriggerTRC20Transfer(ctx,
		conf.HotWalletAddress,
		w.ToAddress,
		conf.USDTContractAddress,
		sunAmount,
	)
	if err != nil {
		return err
	}

	signature, err := tron.SignRawDataHex(rawDataHex, conf.HotWalletPrivateKey)
	if err != nil {
		return err
	}

	if err := client.BroadcastTransaction(ctx, trigger.Transaction, signature); err != nil {
		return err
	}

	// Extract tx hash from the transaction JSON (it's in the txID field)
	txHash, err := extractTxID(trigger.Transaction)
	if err != nil {
		logx.Errorf("[tron-withdraw] extractTxID: %v", err)
		txHash = "unknown"
	}

	now := time.Now()
	if err := deps.CryptoWithdraw.UpdateBroadcasting(ctx, w.ID, txHash, now); err != nil {
		return err
	}
	logx.Infof("[tron-withdraw] broadcast id=%d txHash=%s", w.ID, txHash)
	return nil
}

func confirmBroadcasting(ctx context.Context, conf config.TronConf, client *tron.Client, deps TronWithdrawDeps) {
	broadcasting, err := deps.CryptoWithdraw.ListBroadcasting(ctx)
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
		if (currentBlock - blockNumber) < int64(conf.ConfirmationBlocks) {
			continue
		}

		now := time.Now()
		if err := deps.CryptoWithdraw.UpdateConfirmed(ctx, w.ID, now); err != nil {
			logx.Errorf("[tron-withdraw] UpdateConfirmed %d: %v", w.ID, err)
			continue
		}
		// Permanently deduct from frozen balance (ledger type: crypto_withdraw)
		if err := deps.Wallet.DeductFrozenBalance(ctx, w.UserID, w.Currency, w.Amount, "crypto_withdraw"); err != nil {
			logx.Errorf("[tron-withdraw] DeductFrozenBalance id=%d: %v", w.ID, err)
		} else {
			logx.Infof("[tron-withdraw] confirmed id=%d user=%d amount=%s", w.ID, w.UserID, w.Amount)
		}
	}
}

// extractTxID pulls the txID field from a raw transaction JSON.
func extractTxID(txJSON []byte) (string, error) {
	// Quick manual extraction to avoid importing encoding/json here
	// Look for "txID":"<value>"
	s := string(txJSON)
	key := `"txID":"`
	start := indexOf(s, key)
	if start < 0 {
		return "", nil
	}
	start += len(key)
	end := indexOf(s[start:], `"`)
	if end < 0 {
		return "", nil
	}
	return s[start : start+end], nil
}

func indexOf(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

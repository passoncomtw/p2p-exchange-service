// Package ecpaywebhook_service 處理 ECPay 付款結果通知（Webhook）。
//
// 這是全服務唯一沒有 JWT 保護的端點，呼叫方身分完全靠 ECPay 的 CheckMacValue 簽章驗證：
//   - conf.IsEnabled() 未通過就直接拒絕，避免 HashKey/HashIV 未注入時用空字串當金鑰驗簽
//     （空金鑰任何人都能自算出合法簽章，等同沒有驗證）。
//   - 入帳金額一律取自 DB 既有記錄的 deposit.Amount，不採用通知帶進來的 TradeAmt，
//     這是擋住金額竄改的唯一機制。
package ecpaywebhook_service

import (
	"context"
	"errors"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/config"
	fiatdepositrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/fiat_deposit"
	walletrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/wallet"
	"p2p-exchange/pkg/ecpay"
)

const (
	// ledgerTypeFiatDeposit TWD 入金的帳本類型（與 legacy 相同）。
	ledgerTypeFiatDeposit = "fiat_deposit"
	// rtnCodeSuccess ECPay 付款成功的回傳碼，其餘一律視為付款失敗。
	rtnCodeSuccess = "1"
)

// ECPayWebhookService 處理 ECPay 付款結果通知。
type ECPayWebhookService interface {
	// HandleNotify 驗簽後把對應的入金記錄由 pending 轉為 paid 並入帳。
	//
	// 回傳的 ok/message 直接對應 ECPay 通知協定的回應內容（"1|OK" / "0|<message>"）：
	// ok=true 代表本次通知已處理完畢、ECPay 不需再重送（含「付款失敗」與「重複通知」兩種情形）；
	// ok=false 代表本次通知未被接受，ECPay 會依其重送機制再送一次。
	HandleNotify(ctx context.Context, params map[string]string) (ok bool, message string)
}

type ecpayWebhookService struct {
	cfg        *config.Config
	db         sqlx.SqlConn
	repo       fiatdepositrepo.FiatDepositRepository
	walletRepo walletrepo.WalletRepository
}

func New(
	cfg *config.Config,
	db sqlx.SqlConn,
	repo fiatdepositrepo.FiatDepositRepository,
	walletRepo walletrepo.WalletRepository,
) ECPayWebhookService {
	return &ecpayWebhookService{cfg: cfg, db: db, repo: repo, walletRepo: walletRepo}
}

func (s *ecpayWebhookService) HandleNotify(ctx context.Context, params map[string]string) (bool, string) {
	conf := s.cfg.ECPay
	// ECPay 未設定時一律拒絕：HashKey/HashIV 為空字串時驗簽形同虛設，
	// 寧可拒收通知，也不能用空金鑰放行任何請求。
	if !conf.IsEnabled() {
		logx.WithContext(ctx).Error("[ecpay-webhook] rejected: ECPay not configured")
		return false, "ECPay not configured"
	}
	if !ecpay.VerifyCheckMacValue(params, conf.HashKey, conf.HashIV) {
		// 只記錄交易號，不把整包 params 寫進日誌（含簽章與付款資訊）。
		logx.WithContext(ctx).Errorf("[ecpay-webhook] CheckMacValue mismatch: tradeNo=%s", params["MerchantTradeNo"])
		return false, "signature mismatch"
	}

	tradeNo := params["MerchantTradeNo"]
	deposit, err := s.repo.FindByMerchantTradeNo(ctx, tradeNo)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			logx.WithContext(ctx).Errorf("[ecpay-webhook] deposit not found: %s", tradeNo)
			return false, "deposit not found"
		}
		logx.WithContext(ctx).Errorf("[ecpay-webhook] FindByMerchantTradeNo %s: %v", tradeNo, err)
		return false, "internal error"
	}

	// 付款失敗：標記 failed 後仍回 1|OK，避免 ECPay 對一筆已成定局的失敗通知無限重送。
	// UpdateFailed 自帶 WHERE status='pending'，已入帳的記錄不會被改成 failed。
	if params["RtnCode"] != rtnCodeSuccess {
		if err := s.repo.UpdateFailed(ctx, deposit.ID); err != nil {
			logx.WithContext(ctx).Errorf("[ecpay-webhook] UpdateFailed deposit %d: %v", deposit.ID, err)
		}
		return true, "OK"
	}

	if err := s.confirmAndCredit(ctx, deposit.ID, deposit.UserID, deposit.Currency, deposit.Amount,
		params["TradeNo"], params["PaymentType"], time.Now()); err != nil {
		logx.WithContext(ctx).Errorf("[ecpay-webhook] confirm+credit deposit %d: %v", deposit.ID, err)
		return false, "internal error"
	}

	logx.WithContext(ctx).Infof("[ecpay-webhook] paid deposit id=%d user=%d amount=%s", deposit.ID, deposit.UserID, deposit.Amount)
	return true, "OK"
}

// confirmAndCredit 在同一個交易內完成 pending → paid 的狀態轉換與入帳。
//
// 條件式 UPDATE（WHERE status='pending'）的 RowsAffected 是唯一的入帳依據：
// ECPay 在收到 1|OK 前會重送通知，兩個幾乎同時抵達的通知只有一個能讓 affected=1，
// 另一個直接冪等結束，不會重複入帳；狀態轉換與入帳同屬一個交易，
// 也不會出現 legacy 那種「狀態已是 paid、錢卻沒入帳」且再也無法重試的漏帳。
//
// amount 由呼叫端從 DB 記錄帶入（不是通知的 TradeAmt），確保金額不可被通知竄改。
func (s *ecpayWebhookService) confirmAndCredit(
	ctx context.Context,
	depositID, userID int64,
	currency, amount string,
	ecpayOrderNo, paymentType string,
	paidAt time.Time,
) error {
	return s.db.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		affected, err := s.repo.ConfirmPaidInTx(ctx, session, depositID, ecpayOrderNo, paymentType, paidAt)
		if err != nil {
			return err
		}
		if affected == 0 {
			// 已被其他通知處理過，冪等結束，絕對不能再入帳一次。
			return nil
		}
		return s.walletRepo.DepositWithLedgerTypeInTx(ctx, session, userID, currency, amount, ledgerTypeFiatDeposit, "")
	})
}

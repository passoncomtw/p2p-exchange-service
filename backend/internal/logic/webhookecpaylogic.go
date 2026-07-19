package logic

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	apperrors "p2p-exchange/internal/errors"
	"p2p-exchange/internal/svc"
	"p2p-exchange/pkg/ecpay"
)

type WebhookECPayLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWebhookECPayLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WebhookECPayLogic {
	return &WebhookECPayLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WebhookECPayLogic) HandleNotify(params map[string]string) error {
	conf := l.svcCtx.Config.ECPay
	if !conf.IsEnabled() {
		return apperrors.New(503, "ECPay not configured")
	}

	// Verify CheckMacValue
	if !ecpay.VerifyCheckMacValue(params, conf.HashKey, conf.HashIV) {
		l.Errorf("[ecpay-webhook] CheckMacValue mismatch: %v", params)
		return fmt.Errorf("signature mismatch")
	}

	tradeNo := params["MerchantTradeNo"]
	rtnCode := params["RtnCode"]
	ecpayOrderNo := params["TradeNo"]
	paymentType := params["PaymentType"]

	deposit, err := l.svcCtx.FiatDeposit.FindByMerchantTradeNo(l.ctx, tradeNo)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return fmt.Errorf("deposit not found: %s", tradeNo)
		}
		return err
	}

	if deposit.Status != "pending" {
		// Already processed (idempotent)
		return nil
	}

	if rtnCode != "1" {
		// Payment failed
		if err := l.svcCtx.FiatDeposit.UpdateFailed(l.ctx, deposit.ID); err != nil {
			l.Errorf("[ecpay-webhook] UpdateFailed %d: %v", deposit.ID, err)
		}
		return nil
	}

	// Payment success: update record and credit wallet
	now := time.Now()
	if err := l.svcCtx.FiatDeposit.UpdatePaid(l.ctx, deposit.ID, ecpayOrderNo, paymentType, now); err != nil {
		return err
	}

	if err := l.svcCtx.Wallet.DepositWithLedgerType(l.ctx, deposit.UserID, deposit.Currency, deposit.Amount, "fiat_deposit"); err != nil {
		l.Errorf("[ecpay-webhook] DepositWithLedgerType user=%d amount=%s: %v", deposit.UserID, deposit.Amount, err)
		return err
	}

	l.Infof("[ecpay-webhook] deposit confirmed: id=%d user=%d amount=%s", deposit.ID, deposit.UserID, deposit.Amount)
	return nil
}

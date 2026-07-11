package fee

import (
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
)

// Breakdown 記錄一筆交易的完整手續費明細。
type Breakdown struct {
	// 輸入
	FiatAmount float64
	CryptoAmount float64
	Price      float64
	Currency   string

	// 平台手續費
	PlatformFeeBase   float64
	PlatformFeeRate   float64
	PlatformFeeAmount float64 // = FiatAmount * PlatformFeeRate

	// 收款管道手續費
	PaymentFeeBase   float64
	PaymentFeeRate   float64
	PaymentFeeAmount float64 // = FiatAmount * PaymentFeeRate

	// 合計
	TotalFee    float64 // = PlatformFeeBase + PlatformFeeAmount + PaymentFeeBase + PaymentFeeAmount
	TotalAmount float64 // = FiatAmount + TotalFee（買家實付）
}

// Calculate 計算手續費明細。
// 所有費率與基礎費由 listing 快照傳入，第一階段均為 0。
func Calculate(
	cryptoAmount float64,
	price float64,
	currency string,
	platformFeeBase float64,
	platformFeeRate float64,
	paymentFeeBase float64,
	paymentFeeRate float64,
) Breakdown {
	fiatAmount := cryptoAmount * price
	platformFeeAmount := fiatAmount * platformFeeRate
	paymentFeeAmount := fiatAmount * paymentFeeRate
	totalFee := platformFeeBase + platformFeeAmount + paymentFeeBase + paymentFeeAmount
	totalAmount := fiatAmount + totalFee

	return Breakdown{
		FiatAmount:        fiatAmount,
		CryptoAmount:      cryptoAmount,
		Price:             price,
		Currency:          currency,
		PlatformFeeBase:   platformFeeBase,
		PlatformFeeRate:   platformFeeRate,
		PlatformFeeAmount: platformFeeAmount,
		PaymentFeeBase:    paymentFeeBase,
		PaymentFeeRate:    paymentFeeRate,
		PaymentFeeAmount:  paymentFeeAmount,
		TotalFee:          totalFee,
		TotalAmount:       totalAmount,
	}
}

// LogCreateOrder 在建立訂單時印出手續費計算明細。
// 節點：接單時（POST /app/orders），費率由 listing 快照。
func LogCreateOrder(orderNo string, b Breakdown) {
	logx.Infof("[FEE] create_order | order=%s | crypto=%.8f %s | price=%.4f | fiat=%.4f"+
		" | platform_base=%.4f rate=%.6f amount=%.4f"+
		" | payment_base=%.4f rate=%.6f amount=%.4f"+
		" | total_fee=%.4f | total_amount=%.4f",
		orderNo,
		b.CryptoAmount, b.Currency,
		b.Price,
		b.FiatAmount,
		b.PlatformFeeBase, b.PlatformFeeRate, b.PlatformFeeAmount,
		b.PaymentFeeBase, b.PaymentFeeRate, b.PaymentFeeAmount,
		b.TotalFee,
		b.TotalAmount,
	)
	printSummary("create_order", orderNo, b)
}

// LogConfirmOrder 在賣家確認收款、準備放行時印出手續費明細。
// 節點：PUT /app/orders/:id/confirm，此時費用實際從餘額扣除。
func LogConfirmOrder(orderNo string, b Breakdown) {
	logx.Infof("[FEE] confirm_order | order=%s | total_fee=%.4f | buyer_receives_crypto=%.8f %s | seller_deducted_fiat=%.4f",
		orderNo, b.TotalFee, b.CryptoAmount, b.Currency, b.TotalAmount,
	)
	printSummary("confirm_order", orderNo, b)
}

// LogCancelOrder 在訂單取消或超時時印出退還資訊（手續費不收取）。
// 節點：PUT /app/orders/:id/cancel 或 timeout job。
func LogCancelOrder(orderNo string, reason string, b Breakdown) {
	logx.Infof("[FEE] cancel_order | order=%s | reason=%s | refund_crypto=%.8f %s | fee_waived=%.4f",
		orderNo, reason, b.CryptoAmount, b.Currency, b.TotalFee,
	)
}

// printSummary 在 stdout 印出可讀的手續費區塊，方便開發期追蹤。
func printSummary(event string, ref string, b Breakdown) {
	fmt.Printf("\n┌─ [FEE] %s | %s\n", event, ref)
	fmt.Printf("│  crypto     : %.8f %s @ %.4f\n", b.CryptoAmount, b.Currency, b.Price)
	fmt.Printf("│  fiat_amount: %.4f\n", b.FiatAmount)
	fmt.Printf("│  platform   : base=%.4f + rate(%.4f%%)=%.4f\n",
		b.PlatformFeeBase, b.PlatformFeeRate*100, b.PlatformFeeAmount)
	fmt.Printf("│  payment    : base=%.4f + rate(%.4f%%)=%.4f\n",
		b.PaymentFeeBase, b.PaymentFeeRate*100, b.PaymentFeeAmount)
	fmt.Printf("│  total_fee  : %.4f\n", b.TotalFee)
	fmt.Printf("└─ total_amount: %.4f\n\n", b.TotalAmount)
}

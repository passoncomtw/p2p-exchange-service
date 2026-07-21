package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	apierrors "p2p-exchange/internal/errors"
	"p2p-exchange/internal/model"
	"p2p-exchange/internal/svc"
	"p2p-exchange/internal/types"
	"p2p-exchange/pkg/notification"
	pkgws "p2p-exchange/pkg/ws"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func publishOrderStatusChanged(ctx context.Context, svcCtx *svc.ServiceContext, order *model.Order, status string) {
	if svcCtx.MQ == nil {
		return
	}
	payload := pkgws.OrderStatusChangedData{
		OrderID:  order.ID,
		Status:   status,
		BuyerID:  order.BuyerID,
		SellerID: order.SellerID,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	svcCtx.MQ.PublishAsync(pkgws.EventOrderStatusChanged, data)
}

func sendNotification(ctx context.Context, svcCtx *svc.ServiceContext, recipientID, orderID int64, title, body, channel, priority string) {
	svcCtx.Notifier.Send(ctx, notification.Message{
		RecipientID: recipientID,
		Title:       title,
		Body:        body,
		OrderID:     orderID,
		Channel:     channel,
		Priority:    priority,
	})
}

func timePtr(t time.Time) *string {
	s := t.Format(time.RFC3339)
	return &s
}

func orderToItem(o *model.Order) types.OrderItem {
	item := types.OrderItem{
		ID:                o.ID,
		OrderNo:           o.OrderNo,
		ListingID:         o.ListingID,
		ListingType:       o.ListingType,
		SellerID:          o.SellerID,
		BuyerID:           o.BuyerID,
		CryptoCurrency:    o.CryptoCurrency,
		FiatCurrency:      o.FiatCurrency,
		CryptoAmount:      o.CryptoAmount,
		Price:             o.Price,
		FiatAmount:        o.FiatAmount,
		PlatformFeeBase:   o.PlatformFeeBase,
		PlatformFeeAmount: o.PlatformFeeAmount,
		PaymentFeeBase:    o.PaymentFeeBase,
		PaymentFeeAmount:  o.PaymentFeeAmount,
		TotalFee:          o.TotalFee,
		TotalAmount:       o.TotalAmount,
		PaymentMethodID:   o.PaymentMethodID,
		Status:            o.Status,
		PaymentDeadline:   o.PaymentDeadline.Format(time.RFC3339),
		CancelReason:      o.CancelReason,
		CreatedAt:         o.CreatedAt.Format(time.RFC3339),
	}
	if o.PaidAt != nil {
		item.PaidAt = timePtr(*o.PaidAt)
	}
	if o.ConfirmedAt != nil {
		item.ConfirmedAt = timePtr(*o.ConfirmedAt)
	}
	if o.CompletedAt != nil {
		item.CompletedAt = timePtr(*o.CompletedAt)
	}
	if o.CancelledAt != nil {
		item.CancelledAt = timePtr(*o.CancelledAt)
	}
	return item
}

// ── AppCreateOrderLogic ───────────────────────────────────────────────────────

type AppCreateOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppCreateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppCreateOrderLogic {
	return &AppCreateOrderLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *AppCreateOrderLogic) Create(uid int64, req *types.CreateOrderRequest) (*types.CreateOrderResponse, error) {
	// 1. Get listing
	listing, err := l.svcCtx.Listing.FindByID(l.ctx, req.ListingID)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, apierrors.ErrNotFound
		}
		l.Errorf("get listing id=%d failed: %v", req.ListingID, err)
		return nil, apierrors.ErrInternal
	}

	if listing.Status != "active" {
		return nil, apierrors.New(400, "listing is not active")
	}
	if listing.RemainingAmount < req.CryptoAmount {
		return nil, apierrors.New(400, "insufficient remaining amount in listing")
	}

	// 2. Calculate fiat amount
	fiatAmount := req.CryptoAmount * listing.Price

	// 3. Validate fiat amount range
	if fiatAmount < listing.MinOrderFiat {
		return nil, apierrors.New(400, "order amount below minimum")
	}
	if fiatAmount > listing.MaxOrderFiat {
		return nil, apierrors.New(400, "order amount above maximum")
	}

	// 4. Calculate fees (first phase: rates and bases all 0)
	// platform_fee_amount is the rate-portion; total_fee sums all four components
	platformFeeAmount := fiatAmount * listing.PlatformFeeRate
	paymentFeeAmount := fiatAmount * listing.PaymentFeeRate
	totalFee := listing.PlatformFeeBase + platformFeeAmount + listing.PaymentFeeBase + paymentFeeAmount
	totalAmount := fiatAmount + totalFee

	// 5. Determine buyer/seller
	var buyerID, sellerID int64
	if listing.Type == "sell" {
		buyerID = uid
		sellerID = listing.UserID
	} else {
		sellerID = uid
		buyerID = listing.UserID
	}

	// 6. Generate order number
	orderNo := fmt.Sprintf("P2P%s%06d", time.Now().Format("20060102"), rand.Intn(1000000))

	// 7. Payment deadline
	paymentDeadline := time.Now().Add(time.Duration(listing.PaymentTimeLimit) * time.Minute)

	// 8. Get payment method ID
	var paymentMethodID int64
	if listing.PaymentMethodID != nil {
		paymentMethodID = *listing.PaymentMethodID
	} else if listing.Type == "buy" {
		// For buy listings, seller (uid) needs an active payment method
		pms, err := l.svcCtx.PaymentMethod.FindByUserID(l.ctx, uid)
		if err != nil || len(pms) == 0 {
			return nil, apierrors.New(400, "seller has no active payment method")
		}
		paymentMethodID = pms[0].ID
	}

	// 9. Create order
	order := &model.Order{
		OrderNo:           orderNo,
		ListingID:         listing.ID,
		ListingType:       listing.Type,
		SellerID:          sellerID,
		BuyerID:           buyerID,
		CryptoCurrency:    listing.CryptoCurrency,
		FiatCurrency:      listing.FiatCurrency,
		CryptoAmount:      req.CryptoAmount,
		Price:             listing.Price,
		FiatAmount:        fiatAmount,
		PlatformFeeBase:   listing.PlatformFeeBase,
		PlatformFeeAmount: platformFeeAmount,
		PaymentFeeBase:    listing.PaymentFeeBase,
		PaymentFeeAmount:  paymentFeeAmount,
		TotalFee:          totalFee,
		TotalAmount:       totalAmount,
		PaymentMethodID:   paymentMethodID,
		Status:            "matched",
		PaymentDeadline:   paymentDeadline,
	}

	orderID, err := l.svcCtx.Order.Create(l.ctx, order)
	if err != nil {
		l.Errorf("create order failed: %v", err)
		return nil, apierrors.ErrInternal
	}

	// 10. Create escrow record (lock)
	escrow := &model.EscrowRecord{
		OrderID:        orderID,
		CryptoCurrency: listing.CryptoCurrency,
		Amount:         req.CryptoAmount,
		Action:         "lock",
		Status:         "completed",
	}
	if _, err := l.svcCtx.EscrowRecord.Create(l.ctx, escrow); err != nil {
		l.Errorf("create escrow record failed: %v", err)
		return nil, apierrors.ErrInternal
	}

	// 11. Append order status log
	toStatus := "matched"
	log := &model.OrderStatusLog{
		OrderID:      orderID,
		FromStatus:   nil,
		ToStatus:     toStatus,
		OperatorType: "system",
		OperatorID:   nil,
	}
	if err := l.svcCtx.OrderStatusLog.Append(l.ctx, log); err != nil {
		l.Errorf("append order status log failed: %v", err)
	}

	// 12. Deduct from listing
	if err := l.svcCtx.Listing.DeductAmount(l.ctx, listing.ID, req.CryptoAmount); err != nil {
		l.Errorf("deduct listing amount failed: %v", err)
		return nil, apierrors.ErrInternal
	}

	return &types.CreateOrderResponse{
		ID:      orderID,
		OrderNo: orderNo,
	}, nil
}

// ── AppListOrdersLogic ────────────────────────────────────────────────────────

type AppListOrdersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppListOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppListOrdersLogic {
	return &AppListOrdersLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *AppListOrdersLogic) List(uid int64, req *types.ListOrdersRequest) (*types.ListOrdersResponse, error) {
	rows, err := l.svcCtx.Order.List(l.ctx, uid, req.Role, req.Status, req.Limit, req.Offset)
	if err != nil {
		l.Errorf("list orders uid=%d failed: %v", uid, err)
		return nil, apierrors.ErrInternal
	}

	list := make([]types.OrderItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, orderToItem(row))
	}

	return &types.ListOrdersResponse{List: list}, nil
}

// ── AppGetOrderLogic ──────────────────────────────────────────────────────────

type AppGetOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppGetOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppGetOrderLogic {
	return &AppGetOrderLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *AppGetOrderLogic) Get(id int64) (*types.GetOrderResponse, error) {
	order, err := l.svcCtx.Order.FindByID(l.ctx, id)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, apierrors.ErrNotFound
		}
		l.Errorf("get order id=%d failed: %v", id, err)
		return nil, apierrors.ErrInternal
	}

	return &types.GetOrderResponse{OrderItem: orderToItem(order)}, nil
}

// ── AppPayOrderLogic ──────────────────────────────────────────────────────────

type AppPayOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppPayOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppPayOrderLogic {
	return &AppPayOrderLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *AppPayOrderLogic) Pay(uid int64, id int64) error {
	order, err := l.svcCtx.Order.FindByID(l.ctx, id)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return apierrors.ErrNotFound
		}
		l.Errorf("get order id=%d failed: %v", id, err)
		return apierrors.ErrInternal
	}

	if order.BuyerID != uid {
		return apierrors.ErrForbidden
	}
	if order.Status != "matched" {
		return apierrors.New(400, "order is not in matched status")
	}

	now := time.Now()
	extras := map[string]interface{}{
		"paid_at": now,
	}
	if err := l.svcCtx.Order.UpdateStatus(l.ctx, id, "paid", extras); err != nil {
		l.Errorf("update order status to paid failed: %v", err)
		return apierrors.ErrInternal
	}

	fromStatus := "matched"
	operatorID := uid
	log := &model.OrderStatusLog{
		OrderID:      id,
		FromStatus:   &fromStatus,
		ToStatus:     "paid",
		OperatorType: "buyer",
		OperatorID:   &operatorID,
	}
	if err := l.svcCtx.OrderStatusLog.Append(l.ctx, log); err != nil {
		l.Errorf("append order status log failed: %v", err)
	}

	sendNotification(l.ctx, l.svcCtx, order.SellerID, id, "買家已付款", "請確認收款並放行訂單", "orders", "high")
	publishOrderStatusChanged(l.ctx, l.svcCtx, order, "paid")

	return nil
}

// ── AppConfirmOrderLogic ──────────────────────────────────────────────────────

type AppConfirmOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppConfirmOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppConfirmOrderLogic {
	return &AppConfirmOrderLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *AppConfirmOrderLogic) Confirm(uid int64, id int64) error {
	order, err := l.svcCtx.Order.FindByID(l.ctx, id)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return apierrors.ErrNotFound
		}
		l.Errorf("get order id=%d failed: %v", id, err)
		return apierrors.ErrInternal
	}

	if order.SellerID != uid {
		return apierrors.ErrForbidden
	}
	if order.Status != "paid" {
		return apierrors.New(400, "order is not in paid status")
	}

	now := time.Now()

	// 先取得雙錢包分散式鎖，再開 DB transaction（防止 lock ordering deadlock）
	if l.svcCtx.RDB != nil {
		keys := []string{
			model.WalletLockKey(order.SellerID, order.CryptoCurrency),
			model.WalletLockKey(order.BuyerID, order.CryptoCurrency),
		}
		unlock, err := l.svcCtx.RDB.AcquireLocks(l.ctx, keys, 5*time.Second)
		if err != nil {
			return err
		}
		defer unlock()
	}

	// 原子操作：完成訂單 + 雙錢包轉帳
	if err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		result, err := session.ExecCtx(ctx,
			`UPDATE orders SET status = 'completed', confirmed_at = $1, completed_at = $2, updated_at = NOW() WHERE id = $3 AND status = 'paid'`,
			now, now, id,
		)
		if err != nil {
			return err
		}
		if n, _ := result.RowsAffected(); n == 0 {
			return apierrors.New(400, "order cannot be completed")
		}
		return l.svcCtx.Wallet.TransferInTx(ctx, session, order.SellerID, order.BuyerID, order.CryptoCurrency, order.CryptoAmount, order.OrderNo)
	}); err != nil {
		l.Errorf("confirm order %d failed: %v", id, err)
		return err
	}

	fromStatus := "paid"
	operatorID := uid
	log := &model.OrderStatusLog{
		OrderID:      id,
		FromStatus:   &fromStatus,
		ToStatus:     "completed",
		OperatorType: "seller",
		OperatorID:   &operatorID,
	}
	if err := l.svcCtx.OrderStatusLog.Append(l.ctx, log); err != nil {
		l.Errorf("append order status log failed: %v", err)
	}

	sendNotification(l.ctx, l.svcCtx, order.BuyerID, id, "訂單已完成", "賣家已放行，資產已轉入您的錢包", "orders", "high")
	publishOrderStatusChanged(l.ctx, l.svcCtx, order, "completed")

	return nil
}

// ── AppCancelOrderLogic ───────────────────────────────────────────────────────

type AppCancelOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppCancelOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppCancelOrderLogic {
	return &AppCancelOrderLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *AppCancelOrderLogic) Cancel(uid int64, req *types.CancelOrderRequest) error {
	order, err := l.svcCtx.Order.FindByID(l.ctx, req.ID)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return apierrors.ErrNotFound
		}
		l.Errorf("get order id=%d failed: %v", req.ID, err)
		return apierrors.ErrInternal
	}

	if order.BuyerID != uid && order.SellerID != uid {
		return apierrors.ErrForbidden
	}
	if order.Status != "matched" {
		return apierrors.New(400, "order can only be cancelled in matched status")
	}

	now := time.Now()
	reason := req.Reason

	// 先取得賣家錢包分散式鎖，再開 DB transaction
	if l.svcCtx.RDB != nil {
		unlock, err := l.svcCtx.RDB.AcquireLock(l.ctx, model.WalletLockKey(order.SellerID, order.CryptoCurrency), 5*time.Second)
		if err != nil {
			return err
		}
		defer unlock()
	}

	// 原子操作：取消訂單 + 解凍賣家 USDT
	if err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		result, err := session.ExecCtx(ctx,
			`UPDATE orders SET status = 'cancelled', cancelled_at = $1, cancel_reason = $2, updated_at = NOW() WHERE id = $3 AND status = 'matched'`,
			now, reason, req.ID,
		)
		if err != nil {
			return err
		}
		if n, _ := result.RowsAffected(); n == 0 {
			return apierrors.New(400, "order cannot be cancelled")
		}
		return l.svcCtx.Wallet.UnfreezeInTx(ctx, session, order.SellerID, order.CryptoCurrency, order.CryptoAmount, order.OrderNo)
	}); err != nil {
		l.Errorf("cancel order %d failed: %v", req.ID, err)
		return err
	}

	cancelOperatorType := "buyer"
	if uid == order.SellerID {
		cancelOperatorType = "seller"
	}
	fromStatus := "matched"
	operatorID := uid
	remark := reason
	log := &model.OrderStatusLog{
		OrderID:      req.ID,
		FromStatus:   &fromStatus,
		ToStatus:     "cancelled",
		OperatorType: cancelOperatorType,
		OperatorID:   &operatorID,
		Remark:       &remark,
	}
	if err := l.svcCtx.OrderStatusLog.Append(l.ctx, log); err != nil {
		l.Errorf("append order status log failed: %v", err)
	}

	// 通知對方訂單已取消
	notifyID := order.SellerID
	if uid == order.SellerID {
		notifyID = order.BuyerID
	}
	sendNotification(l.ctx, l.svcCtx, notifyID, req.ID, "訂單已取消", "對方已取消此筆訂單", "orders", "normal")
	publishOrderStatusChanged(l.ctx, l.svcCtx, order, "cancelled")

	return nil
}

// ── AppDisputeOrderLogic ──────────────────────────────────────────────────────

type AppDisputeOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppDisputeOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppDisputeOrderLogic {
	return &AppDisputeOrderLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *AppDisputeOrderLogic) Dispute(uid int64, req *types.DisputeOrderRequest) error {
	order, err := l.svcCtx.Order.FindByID(l.ctx, req.ID)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return apierrors.ErrNotFound
		}
		l.Errorf("get order id=%d failed: %v", req.ID, err)
		return apierrors.ErrInternal
	}

	if order.BuyerID != uid && order.SellerID != uid {
		return apierrors.ErrForbidden
	}
	if order.Status != "paid" {
		return apierrors.New(400, "order can only be disputed in paid status")
	}

	if err := l.svcCtx.Order.UpdateStatus(l.ctx, req.ID, "disputed", nil); err != nil {
		l.Errorf("update order status to disputed failed: %v", err)
		return apierrors.ErrInternal
	}

	disputeOperatorType := "buyer"
	if uid == order.SellerID {
		disputeOperatorType = "seller"
	}
	fromStatus := "paid"
	operatorID := uid
	remark := req.Reason
	log := &model.OrderStatusLog{
		OrderID:      req.ID,
		FromStatus:   &fromStatus,
		ToStatus:     "disputed",
		OperatorType: disputeOperatorType,
		OperatorID:   &operatorID,
		Remark:       &remark,
	}
	if err := l.svcCtx.OrderStatusLog.Append(l.ctx, log); err != nil {
		l.Errorf("append order status log failed: %v", err)
	}

	// 通知對方已提交申訴
	notifyID := order.SellerID
	if uid == order.SellerID {
		notifyID = order.BuyerID
	}
	sendNotification(l.ctx, l.svcCtx, notifyID, req.ID, "訂單申訴", "對方已對此筆訂單提交申訴，客服將介入處理", "orders", "high")
	publishOrderStatusChanged(l.ctx, l.svcCtx, order, "disputed")

	return nil
}

package order_service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"time"

	app_interface "p2p-exchange/cmd/v1-p2p-exchange-service/internal/interfaces/app"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/model/entity"
	escrowrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/escrow_record"
	listingrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/listing"
	orderrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/order"
	orderstatuslogrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/order_status_log"
	paymentrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/payment_method"
	walletrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/wallet"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/service/orderallocation"
	apierrors "p2p-exchange/internal/errors"
	"p2p-exchange/pkg/mq"
	pkgws "p2p-exchange/pkg/ws"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"go.uber.org/fx"
)

// errOrderStatusChanged 條件式狀態轉換影響 0 列：訂單在讀取後已被其他請求改變狀態，
// 本次請求不得再執行任何錢包操作。
var errOrderStatusChanged = apierrors.New(409, "order status changed")

// errInsufficientRemaining 掛單剩餘量不足以成立這筆訂單（鎖內檢查或條件式扣量影響 0 列）。
var errInsufficientRemaining = apierrors.New(400, "insufficient remaining amount in listing")

// minCryptoAmount 為 NUMERIC(38, 18) 可表示的最小正數。
const minCryptoAmount = 1e-18

type OrderService interface {
	Create(ctx context.Context, uid int64, req app_interface.CreateOrderRequest) (*app_interface.CreateOrderResponse, error)
	List(ctx context.Context, uid int64, req app_interface.ListOrdersRequest) (*app_interface.ListOrdersResponse, error)
	Get(ctx context.Context, id int64) (*app_interface.OrderItem, error)
	Pay(ctx context.Context, uid, id int64) error
	Confirm(ctx context.Context, uid, id int64) error
	Cancel(ctx context.Context, uid int64, req app_interface.CancelOrderRequest) error
	Dispute(ctx context.Context, uid int64, req app_interface.DisputeOrderRequest) error
}

type orderService struct {
	db            sqlx.SqlConn
	mqPublisher   mq.Publisher
	orderRepo     orderrepo.OrderRepository
	listingRepo   listingrepo.ListingRepository
	paymentRepo   paymentrepo.PaymentMethodRepository
	walletRepo    walletrepo.WalletRepository
	escrowRepo    escrowrepo.EscrowRepository
	statusLogRepo orderstatuslogrepo.OrderStatusLogRepository
}

type Params struct {
	fx.In

	DB            sqlx.SqlConn
	MQPublisher   mq.Publisher
	OrderRepo     orderrepo.OrderRepository
	ListingRepo   listingrepo.ListingRepository
	PaymentRepo   paymentrepo.PaymentMethodRepository
	WalletRepo    walletrepo.WalletRepository
	EscrowRepo    escrowrepo.EscrowRepository
	StatusLogRepo orderstatuslogrepo.OrderStatusLogRepository
}

func New(p Params) OrderService {
	return &orderService{
		db:            p.DB,
		mqPublisher:   p.MQPublisher,
		orderRepo:     p.OrderRepo,
		listingRepo:   p.ListingRepo,
		paymentRepo:   p.PaymentRepo,
		walletRepo:    p.WalletRepo,
		escrowRepo:    p.EscrowRepo,
		statusLogRepo: p.StatusLogRepo,
	}
}

func (s *orderService) Create(ctx context.Context, uid int64, req app_interface.CreateOrderRequest) (*app_interface.CreateOrderResponse, error) {
	if !isPositiveFinite(req.CryptoAmount) {
		return nil, apierrors.New(400, "cryptoAmount must be a positive finite number")
	}

	// 收款帳戶查詢走連線池而非交易 session，必須在交易外完成：交易已持有一條連線時再向連線池要第二條，
	// 並發量高時會耗盡連線池而互相等待。只有吃 buy 掛單（且掛單未指定收款帳戶）時才會用到結果。
	takerPaymentMethods, takerPaymentErr := s.paymentRepo.FindByUserID(ctx, uid)

	var order *entity.Order
	// 依 p2p-order「建立訂單的鎖與原子性」：以掛單列的 FOR UPDATE 鎖序列化同一掛單的建單、
	// 訂單釋放與掛單取消，並在同一交易內完成凍結、寫入訂單、扣量、託管紀錄與狀態日誌；
	// 任何一步失敗整筆回滾，錯誤一律不得忽略。
	if err := s.db.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		listing, err := s.listingRepo.FindByIDForUpdateInTx(ctx, session, req.ListingID)
		if err != nil {
			if err == sqlx.ErrNotFound {
				return apierrors.ErrNotFound
			}
			return err
		}

		order, err = buildOrder(uid, listing, req.CryptoAmount, takerPaymentMethods, takerPaymentErr)
		if err != nil {
			return err
		}

		// buy 掛單由吃單者（賣方）凍結 cryptoAmount；sell 掛單的額度已在掛單時凍結，這裡不再凍結。
		// 金額字串格式與 UnfreezeInTx / TransferInTx 內部的 %.18f 一致，凍結與釋放才會是同一個數值。
		if listing.Type == "buy" {
			if err := s.walletRepo.FreezeInTx(ctx, session, order.SellerID, order.CryptoCurrency, formatCryptoAmount(order.CryptoAmount)); err != nil {
				return err
			}
		}

		orderID, err := s.orderRepo.CreateInTx(ctx, session, order)
		if err != nil {
			return err
		}
		order.ID = orderID

		n, err := s.listingRepo.DeductAmountInTx(ctx, session, listing.ID, order.CryptoAmount)
		if err != nil {
			return err
		}
		if n == 0 {
			return errInsufficientRemaining
		}

		if err := s.escrowRepo.CreateInTx(ctx, session, &escrowrepo.EscrowRecord{
			OrderID:        orderID,
			CryptoCurrency: order.CryptoCurrency,
			Amount:         order.CryptoAmount,
			Action:         "lock",
			Status:         "completed",
		}); err != nil {
			return err
		}

		return s.statusLogRepo.AppendInTx(ctx, session, &orderstatuslogrepo.OrderStatusLog{
			OrderID:      orderID,
			ToStatus:     "matched",
			OperatorType: "system",
		})
	}); err != nil {
		return nil, toAppError(err)
	}

	s.publishOrderStatusChanged(order, "matched")
	return &app_interface.CreateOrderResponse{ID: order.ID, OrderNo: order.OrderNo}, nil
}

// buildOrder 以鎖內讀到的掛單再驗證吃單條件並組出訂單（尚未寫入）。
// takerPaymentMethods / takerPaymentErr 為交易外查到的吃單者收款帳戶，只在 buy 掛單未指定收款帳戶時使用。
func buildOrder(uid int64, listing *entity.Listing, cryptoAmount float64, takerPaymentMethods []*entity.PaymentMethod, takerPaymentErr error) (*entity.Order, error) {
	// 買賣方為同一人沒有合法的交易意義，且屬於單一帳號自行造幣的路徑，一律拒絕。
	if listing.UserID == uid {
		return nil, apierrors.New(400, "cannot take own listing")
	}
	if listing.Status != "active" {
		return nil, apierrors.New(400, "listing is not active")
	}
	if listing.RemainingAmount < cryptoAmount {
		return nil, errInsufficientRemaining
	}
	if !isStorableCryptoAmount(cryptoAmount) {
		return nil, apierrors.New(400, "cryptoAmount must be at least 1e-18 with at most 18 decimal places")
	}

	fiatAmount := cryptoAmount * listing.Price
	if fiatAmount < listing.MinOrderFiat {
		return nil, apierrors.New(400, "order amount below minimum")
	}
	if fiatAmount > listing.MaxOrderFiat {
		return nil, apierrors.New(400, "order amount above maximum")
	}

	platformFeeAmount := fiatAmount * listing.PlatformFeeRate
	paymentFeeAmount := fiatAmount * listing.PaymentFeeRate
	totalFee := listing.PlatformFeeBase + platformFeeAmount + listing.PaymentFeeBase + paymentFeeAmount
	totalAmount := fiatAmount + totalFee

	var buyerID, sellerID int64
	if listing.Type == "sell" {
		buyerID = uid
		sellerID = listing.UserID
	} else {
		sellerID = uid
		buyerID = listing.UserID
	}

	var paymentMethodID int64
	if listing.PaymentMethodID != nil {
		paymentMethodID = *listing.PaymentMethodID
	} else if listing.Type == "buy" {
		if takerPaymentErr != nil || len(takerPaymentMethods) == 0 {
			return nil, apierrors.New(400, "seller has no active payment method")
		}
		paymentMethodID = takerPaymentMethods[0].ID
	}

	return &entity.Order{
		OrderNo:           fmt.Sprintf("P2P%s%06d", time.Now().Format("20060102"), rand.Intn(1000000)),
		ListingID:         listing.ID,
		ListingType:       listing.Type,
		SellerID:          sellerID,
		BuyerID:           buyerID,
		CryptoCurrency:    listing.CryptoCurrency,
		FiatCurrency:      listing.FiatCurrency,
		CryptoAmount:      cryptoAmount,
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
		PaymentDeadline:   time.Now().Add(time.Duration(listing.PaymentTimeLimit) * time.Minute),
	}, nil
}

func (s *orderService) List(ctx context.Context, uid int64, req app_interface.ListOrdersRequest) (*app_interface.ListOrdersResponse, error) {
	rows, err := s.orderRepo.List(ctx, uid, req.Role, req.Status, req.Limit, req.Offset)
	if err != nil {
		return nil, apierrors.ErrInternal
	}
	list := make([]app_interface.OrderItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, toOrderItem(row))
	}
	return &app_interface.ListOrdersResponse{List: list}, nil
}

func (s *orderService) Get(ctx context.Context, id int64) (*app_interface.OrderItem, error) {
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, apierrors.ErrNotFound
		}
		return nil, apierrors.ErrInternal
	}
	item := toOrderItem(order)
	return &item, nil
}

func (s *orderService) Pay(ctx context.Context, uid, id int64) error {
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return apierrors.ErrNotFound
		}
		return apierrors.ErrInternal
	}
	if order.BuyerID != uid {
		return apierrors.ErrForbidden
	}
	if order.Status != "matched" {
		return apierrors.New(400, "order is not in matched status")
	}

	fromStatus := "matched"
	now := time.Now()
	n, err := s.orderRepo.UpdateStatus(ctx, id, fromStatus, "paid", map[string]interface{}{"paid_at": now})
	if err != nil {
		return apierrors.ErrInternal
	}
	if n == 0 {
		return errOrderStatusChanged
	}

	_ = s.statusLogRepo.Append(ctx, &orderstatuslogrepo.OrderStatusLog{
		OrderID:      id,
		FromStatus:   &fromStatus,
		ToStatus:     "paid",
		OperatorType: "buyer",
		OperatorID:   &uid,
	})

	s.publishOrderStatusChanged(order, "paid")
	return nil
}

func (s *orderService) Confirm(ctx context.Context, uid, id int64) error {
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return apierrors.ErrNotFound
		}
		return apierrors.ErrInternal
	}
	if order.SellerID != uid {
		return apierrors.ErrForbidden
	}
	if order.Status != "paid" {
		return apierrors.New(400, "order is not in paid status")
	}

	now := time.Now()
	unlock, err := s.walletRepo.AcquireLocks(ctx, []int64{order.SellerID, order.BuyerID}, order.CryptoCurrency)
	if err != nil {
		return err
	}
	defer unlock()

	fromStatus := "paid"
	if err := s.db.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		n, err := s.orderRepo.UpdateStatusInTx(ctx, session, id, fromStatus, "completed", map[string]interface{}{
			"confirmed_at": now,
			"completed_at": now,
		})
		if err != nil {
			return err
		}
		if n == 0 {
			return errOrderStatusChanged
		}
		return s.walletRepo.TransferInTx(ctx, session, order.SellerID, order.BuyerID, order.CryptoCurrency, order.CryptoAmount, order.OrderNo)
	}); err != nil {
		return err
	}

	_ = s.statusLogRepo.Append(ctx, &orderstatuslogrepo.OrderStatusLog{
		OrderID:      id,
		FromStatus:   &fromStatus,
		ToStatus:     "completed",
		OperatorType: "seller",
		OperatorID:   &uid,
	})

	s.publishOrderStatusChanged(order, "completed")
	return nil
}

func (s *orderService) Cancel(ctx context.Context, uid int64, req app_interface.CancelOrderRequest) error {
	order, err := s.orderRepo.FindByID(ctx, req.ID)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return apierrors.ErrNotFound
		}
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

	unlock, err := s.walletRepo.AcquireLocks(ctx, []int64{order.SellerID}, order.CryptoCurrency)
	if err != nil {
		return err
	}
	defer unlock()

	fromStatus := "matched"
	if err := s.db.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		n, err := s.orderRepo.UpdateStatusInTx(ctx, session, req.ID, fromStatus, "cancelled", map[string]interface{}{
			"cancelled_at":  now,
			"cancel_reason": reason,
		})
		if err != nil {
			return err
		}
		if n == 0 {
			return errOrderStatusChanged
		}
		return orderallocation.ReleaseInTx(ctx, session, s.listingRepo, s.walletRepo, order)
	}); err != nil {
		return err
	}

	operatorType := "buyer"
	if uid == order.SellerID {
		operatorType = "seller"
	}
	_ = s.statusLogRepo.Append(ctx, &orderstatuslogrepo.OrderStatusLog{
		OrderID:      req.ID,
		FromStatus:   &fromStatus,
		ToStatus:     "cancelled",
		OperatorType: operatorType,
		OperatorID:   &uid,
		Remark:       &reason,
	})

	s.publishOrderStatusChanged(order, "cancelled")
	return nil
}

func (s *orderService) Dispute(ctx context.Context, uid int64, req app_interface.DisputeOrderRequest) error {
	order, err := s.orderRepo.FindByID(ctx, req.ID)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return apierrors.ErrNotFound
		}
		return apierrors.ErrInternal
	}
	if order.BuyerID != uid && order.SellerID != uid {
		return apierrors.ErrForbidden
	}
	if order.Status != "paid" {
		return apierrors.New(400, "order can only be disputed in paid status")
	}

	fromStatus := "paid"
	n, err := s.orderRepo.UpdateStatus(ctx, req.ID, fromStatus, "disputed", nil)
	if err != nil {
		return apierrors.ErrInternal
	}
	if n == 0 {
		return errOrderStatusChanged
	}

	operatorType := "buyer"
	if uid == order.SellerID {
		operatorType = "seller"
	}
	remark := req.Reason
	_ = s.statusLogRepo.Append(ctx, &orderstatuslogrepo.OrderStatusLog{
		OrderID:      req.ID,
		FromStatus:   &fromStatus,
		ToStatus:     "disputed",
		OperatorType: operatorType,
		OperatorID:   &uid,
		Remark:       &remark,
	})

	s.publishOrderStatusChanged(order, "disputed")
	return nil
}

func (s *orderService) publishOrderStatusChanged(order *entity.Order, status string) {
	if s.mqPublisher == nil {
		return
	}
	data, err := json.Marshal(pkgws.OrderStatusChangedData{
		OrderID:  order.ID,
		Status:   status,
		BuyerID:  order.BuyerID,
		SellerID: order.SellerID,
	})
	if err != nil {
		return
	}
	s.mqPublisher.PublishAsync(pkgws.SubjectNotifyAdmin, data)
	s.mqPublisher.PublishAsync(pkgws.SubjectNotifyBuyerPrefix+strconv.FormatInt(order.BuyerID, 10), data)
	s.mqPublisher.PublishAsync(pkgws.SubjectNotifySellerPrefix+strconv.FormatInt(order.SellerID, 10), data)
}

// isPositiveFinite 回傳 v 是否為大於零的有限值（排除 NaN 與 ±Inf）。
func isPositiveFinite(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0) && v > 0
}

// formatCryptoAmount 把加密貨幣數量格式化為 18 位小數字串，與錢包 UnfreezeInTx / TransferInTx 內部格式一致。
func formatCryptoAmount(v float64) string {
	return fmt.Sprintf("%.18f", v)
}

// isStorableCryptoAmount 回傳 v 是否不小於 1e-18，且格式化為 18 位小數後再解析仍是同一個值。
// 不符合時，寫入 NUMERIC(38, 18) 的訂單量、凍結量與後續釋放量會各自被捨入成不同的數值，
// 或被捨入成 0 而形成「凍結 0、訂單非 0」的帳務落差。
func isStorableCryptoAmount(v float64) bool {
	if v < minCryptoAmount {
		return false
	}
	parsed, err := strconv.ParseFloat(formatCryptoAmount(v), 64)
	return err == nil && parsed == v
}

// toAppError 保留交易中回傳的 AppError；其餘（資料庫等）錯誤一律轉為 500，不把內部訊息回給使用者。
func toAppError(err error) error {
	var appErr *apierrors.AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return apierrors.ErrInternal
}

func toOrderItem(o *entity.Order) app_interface.OrderItem {
	item := app_interface.OrderItem{
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
		s := o.PaidAt.Format(time.RFC3339)
		item.PaidAt = &s
	}
	if o.ConfirmedAt != nil {
		s := o.ConfirmedAt.Format(time.RFC3339)
		item.ConfirmedAt = &s
	}
	if o.CompletedAt != nil {
		s := o.CompletedAt.Format(time.RFC3339)
		item.CompletedAt = &s
	}
	if o.CancelledAt != nil {
		s := o.CancelledAt.Format(time.RFC3339)
		item.CancelledAt = &s
	}
	return item
}

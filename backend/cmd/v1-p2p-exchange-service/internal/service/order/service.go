package order_service

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"go.uber.org/fx"
	app_interface "p2p-exchange/cmd/v1-p2p-exchange-service/internal/interfaces/app"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/model/entity"
	escrowrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/escrow_record"
	listingrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/listing"
	orderrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/order"
	orderstatuslogrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/order_status_log"
	paymentrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/payment_method"
	walletrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/wallet"
	apierrors "p2p-exchange/internal/errors"
	"p2p-exchange/pkg/mq"
	pkgws "p2p-exchange/pkg/ws"
)

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
	db             sqlx.SqlConn
	mqClient       *mq.Client
	orderRepo      orderrepo.OrderRepository
	listingRepo    listingrepo.ListingRepository
	paymentRepo    paymentrepo.PaymentMethodRepository
	walletRepo     walletrepo.WalletRepository
	escrowRepo     escrowrepo.EscrowRepository
	statusLogRepo  orderstatuslogrepo.OrderStatusLogRepository
}

type Params struct {
	fx.In

	DB            sqlx.SqlConn
	MQClient      *mq.Client
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
		mqClient:      p.MQClient,
		orderRepo:     p.OrderRepo,
		listingRepo:   p.ListingRepo,
		paymentRepo:   p.PaymentRepo,
		walletRepo:    p.WalletRepo,
		escrowRepo:    p.EscrowRepo,
		statusLogRepo: p.StatusLogRepo,
	}
}

func (s *orderService) Create(ctx context.Context, uid int64, req app_interface.CreateOrderRequest) (*app_interface.CreateOrderResponse, error) {
	listing, err := s.listingRepo.FindByID(ctx, req.ListingID)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, apierrors.ErrNotFound
		}
		return nil, apierrors.ErrInternal
	}

	if listing.Status != "active" {
		return nil, apierrors.New(400, "listing is not active")
	}
	if listing.RemainingAmount < req.CryptoAmount {
		return nil, apierrors.New(400, "insufficient remaining amount in listing")
	}

	fiatAmount := req.CryptoAmount * listing.Price
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

	orderNo := fmt.Sprintf("P2P%s%06d", time.Now().Format("20060102"), rand.Intn(1000000))
	paymentDeadline := time.Now().Add(time.Duration(listing.PaymentTimeLimit) * time.Minute)

	var paymentMethodID int64
	if listing.PaymentMethodID != nil {
		paymentMethodID = *listing.PaymentMethodID
	} else if listing.Type == "buy" {
		pms, err := s.paymentRepo.FindByUserID(ctx, uid)
		if err != nil || len(pms) == 0 {
			return nil, apierrors.New(400, "seller has no active payment method")
		}
		paymentMethodID = pms[0].ID
	}

	order := &entity.Order{
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

	orderID, err := s.orderRepo.Create(ctx, order)
	if err != nil {
		return nil, apierrors.ErrInternal
	}

	_ = s.escrowRepo.Create(ctx, &escrowrepo.EscrowRecord{
		OrderID:        orderID,
		CryptoCurrency: listing.CryptoCurrency,
		Amount:         req.CryptoAmount,
		Action:         "lock",
		Status:         "completed",
	})

	toStatus := "matched"
	_ = s.statusLogRepo.Append(ctx, &orderstatuslogrepo.OrderStatusLog{
		OrderID:      orderID,
		ToStatus:     toStatus,
		OperatorType: "system",
	})

	_ = s.orderRepo.DeductListingAmount(ctx, listing.ID, req.CryptoAmount)

	return &app_interface.CreateOrderResponse{ID: orderID, OrderNo: orderNo}, nil
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

	now := time.Now()
	if err := s.orderRepo.UpdateStatus(ctx, id, "paid", map[string]interface{}{"paid_at": now}); err != nil {
		return apierrors.ErrInternal
	}

	fromStatus := "matched"
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

	if err := s.db.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		n, err := s.orderRepo.UpdateStatusInTx(ctx, session, id, "completed", map[string]interface{}{
			"confirmed_at": now,
			"completed_at": now,
		})
		if err != nil {
			return err
		}
		if n == 0 {
			return apierrors.New(400, "order cannot be completed")
		}
		return s.walletRepo.TransferInTx(ctx, session, order.SellerID, order.BuyerID, order.CryptoCurrency, order.CryptoAmount, order.OrderNo)
	}); err != nil {
		return err
	}

	fromStatus := "paid"
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

	if err := s.db.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		n, err := s.orderRepo.UpdateStatusInTx(ctx, session, req.ID, "cancelled", map[string]interface{}{
			"cancelled_at": now,
			"cancel_reason": reason,
		})
		if err != nil {
			return err
		}
		if n == 0 {
			return apierrors.New(400, "order cannot be cancelled")
		}
		return s.walletRepo.UnfreezeInTx(ctx, session, order.SellerID, order.CryptoCurrency, order.CryptoAmount)
	}); err != nil {
		return err
	}

	operatorType := "buyer"
	if uid == order.SellerID {
		operatorType = "seller"
	}
	fromStatus := "matched"
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

	if err := s.orderRepo.UpdateStatus(ctx, req.ID, "disputed", nil); err != nil {
		return apierrors.ErrInternal
	}

	operatorType := "buyer"
	if uid == order.SellerID {
		operatorType = "seller"
	}
	fromStatus := "paid"
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
	if s.mqClient == nil {
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
	s.mqClient.PublishAsync(pkgws.SubjectNotifyAdmin, data)
	s.mqClient.PublishAsync(pkgws.SubjectNotifyBuyerPrefix+strconv.FormatInt(order.BuyerID, 10), data)
	s.mqClient.PublishAsync(pkgws.SubjectNotifySellerPrefix+strconv.FormatInt(order.SellerID, 10), data)
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

package backend_admin_service

import (
	"context"
	"errors"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	app_interface "p2p-exchange/cmd/v1-p2p-exchange-service/internal/interfaces/app"
	backend_interface "p2p-exchange/cmd/v1-p2p-exchange-service/internal/interfaces/backend"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/model/entity"
	listingrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/listing"
	orderrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/order"
	walletrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/wallet"
	apierrors "p2p-exchange/internal/errors"
)

type BackendAdminService interface {
	ListListings(ctx context.Context, req backend_interface.BackendListListingsRequest) (*app_interface.ListListingsResponse, error)
	ListOrders(ctx context.Context, req backend_interface.BackendListOrdersRequest) (*app_interface.ListOrdersResponse, error)
	ResolveOrder(ctx context.Context, req backend_interface.BackendResolveOrderRequest) error
}

type backendAdminService struct {
	db          sqlx.SqlConn
	listingRepo listingrepo.ListingRepository
	orderRepo   orderrepo.OrderRepository
	walletRepo  walletrepo.WalletRepository
}

func New(
	db sqlx.SqlConn,
	listingRepo listingrepo.ListingRepository,
	orderRepo orderrepo.OrderRepository,
	walletRepo walletrepo.WalletRepository,
) BackendAdminService {
	return &backendAdminService{
		db:          db,
		listingRepo: listingRepo,
		orderRepo:   orderRepo,
		walletRepo:  walletRepo,
	}
}

func (s *backendAdminService) ListListings(ctx context.Context, req backend_interface.BackendListListingsRequest) (*app_interface.ListListingsResponse, error) {
	listings, err := s.listingRepo.List(ctx, req.Type, req.Status, req.Limit, req.Offset)
	if err != nil {
		return nil, apierrors.ErrInternal
	}

	items := make([]app_interface.ListingItem, 0, len(listings))
	for _, l := range listings {
		items = append(items, toListingItem(l))
	}
	return &app_interface.ListListingsResponse{List: items}, nil
}

func (s *backendAdminService) ListOrders(ctx context.Context, req backend_interface.BackendListOrdersRequest) (*app_interface.ListOrdersResponse, error) {
	orders, err := s.orderRepo.BackendList(ctx, req.Status, req.Limit, req.Offset)
	if err != nil {
		return nil, apierrors.ErrInternal
	}

	items := make([]app_interface.OrderItem, 0, len(orders))
	for _, o := range orders {
		items = append(items, toOrderItem(o))
	}
	return &app_interface.ListOrdersResponse{List: items}, nil
}

func (s *backendAdminService) ResolveOrder(ctx context.Context, req backend_interface.BackendResolveOrderRequest) error {
	order, err := s.orderRepo.FindByID(ctx, req.ID)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return apierrors.ErrNotFound
		}
		return apierrors.ErrInternal
	}

	if order.Status != "disputed" {
		return apierrors.New(400, "order is not in disputed status")
	}

	switch req.Action {
	case "complete":
		unlock, err := s.walletRepo.AcquireLocks(ctx, []int64{order.SellerID, order.BuyerID}, order.CryptoCurrency)
		if err != nil {
			return apierrors.ErrInternal
		}
		defer unlock()

		return s.db.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
			if _, err := s.orderRepo.UpdateStatusInTx(ctx, session, order.ID, "completed", map[string]interface{}{
				"completed_at": "NOW()",
			}); err != nil {
				return err
			}
			return s.walletRepo.TransferInTx(ctx, session, order.SellerID, order.BuyerID, order.CryptoCurrency, order.CryptoAmount, order.OrderNo)
		})

	case "refund":
		unlock, err := s.walletRepo.AcquireLocks(ctx, []int64{order.SellerID}, order.CryptoCurrency)
		if err != nil {
			return apierrors.ErrInternal
		}
		defer unlock()

		return s.db.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
			if _, err := s.orderRepo.UpdateStatusInTx(ctx, session, order.ID, "refunded", map[string]interface{}{
				"cancel_reason": req.Reason,
			}); err != nil {
				return err
			}
			if err := s.walletRepo.UnfreezeInTx(ctx, session, order.SellerID, order.CryptoCurrency, order.CryptoAmount); err != nil {
				return err
			}
			return s.listingRepo.RestoreAmountInTx(ctx, session, order.ListingID, order.CryptoAmount)
		})

	default:
		return errors.New("invalid action")
	}
}

func toListingItem(l *entity.Listing) app_interface.ListingItem {
	return app_interface.ListingItem{
		ID:               l.ID,
		UserID:           l.UserID,
		Type:             l.Type,
		CryptoCurrency:   l.CryptoCurrency,
		FiatCurrency:     l.FiatCurrency,
		TotalAmount:      l.TotalAmount,
		RemainingAmount:  l.RemainingAmount,
		Price:            l.Price,
		MinOrderFiat:     l.MinOrderFiat,
		MaxOrderFiat:     l.MaxOrderFiat,
		PlatformFeeBase:  l.PlatformFeeBase,
		PlatformFeeRate:  l.PlatformFeeRate,
		PaymentFeeBase:   l.PaymentFeeBase,
		PaymentFeeRate:   l.PaymentFeeRate,
		PaymentTimeLimit: l.PaymentTimeLimit,
		PaymentMethodID:  l.PaymentMethodID,
		Status:           l.Status,
		CreatedAt:        l.CreatedAt.Format(time.RFC3339),
	}
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

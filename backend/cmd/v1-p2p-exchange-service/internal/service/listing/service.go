package listing_service

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	app_interface "p2p-exchange/cmd/v1-p2p-exchange-service/internal/interfaces/app"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/model/entity"
	listingrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/listing"
	walletrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/wallet"
	apierrors "p2p-exchange/internal/errors"
	"p2p-exchange/internal/infra/rdb"
)

type ListingService interface {
	Create(ctx context.Context, uid int64, req app_interface.CreateListingRequest) (*app_interface.CreateListingResponse, error)
	List(ctx context.Context, req app_interface.ListListingsRequest) (*app_interface.ListListingsResponse, error)
	ListMine(ctx context.Context, uid int64, req app_interface.ListListingsRequest) (*app_interface.ListListingsResponse, error)
	Get(ctx context.Context, id int64) (*app_interface.ListingItem, error)
	Cancel(ctx context.Context, uid, id int64) error
}

type listingService struct {
	db          sqlx.SqlConn
	rdb         *rdb.Client
	listingRepo listingrepo.ListingRepository
	walletRepo  walletrepo.WalletRepository
}

func New(db sqlx.SqlConn, rdb *rdb.Client, listingRepo listingrepo.ListingRepository, walletRepo walletrepo.WalletRepository) ListingService {
	return &listingService{db: db, rdb: rdb, listingRepo: listingRepo, walletRepo: walletRepo}
}

func (s *listingService) Create(ctx context.Context, uid int64, req app_interface.CreateListingRequest) (*app_interface.CreateListingResponse, error) {
	if req.Type == "sell" && req.PaymentMethodID == nil {
		return nil, apierrors.New(400, "sell listing requires a payment method")
	}

	if req.Type == "sell" {
		if err := s.walletRepo.Freeze(ctx, uid, req.CryptoCurrency, req.TotalAmount); err != nil {
			return nil, err
		}
	}

	l := &entity.Listing{
		UserID:           uid,
		Type:             req.Type,
		CryptoCurrency:   req.CryptoCurrency,
		FiatCurrency:     req.FiatCurrency,
		TotalAmount:      req.TotalAmount,
		RemainingAmount:  req.TotalAmount,
		Price:            req.Price,
		MinOrderFiat:     req.MinOrderFiat,
		MaxOrderFiat:     req.MaxOrderFiat,
		PlatformFeeBase:  0,
		PlatformFeeRate:  0,
		PaymentFeeBase:   0,
		PaymentFeeRate:   0,
		PaymentTimeLimit: req.PaymentTimeLimit,
		PaymentMethodID:  req.PaymentMethodID,
		Status:           "active",
	}

	id, err := s.listingRepo.Create(ctx, l)
	if err != nil {
		return nil, apierrors.ErrInternal
	}

	return &app_interface.CreateListingResponse{ID: id}, nil
}

func (s *listingService) List(ctx context.Context, req app_interface.ListListingsRequest) (*app_interface.ListListingsResponse, error) {
	rows, err := s.listingRepo.List(ctx, req.Type, req.Status, req.Limit, req.Offset)
	if err != nil {
		return nil, apierrors.ErrInternal
	}

	list := make([]app_interface.ListingItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, toListingItem(row))
	}
	return &app_interface.ListListingsResponse{List: list}, nil
}

func (s *listingService) ListMine(ctx context.Context, uid int64, req app_interface.ListListingsRequest) (*app_interface.ListListingsResponse, error) {
	rows, err := s.listingRepo.ListByUser(ctx, uid, req.Type, req.Status, req.Limit, req.Offset)
	if err != nil {
		return nil, apierrors.ErrInternal
	}

	list := make([]app_interface.ListingItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, toListingItem(row))
	}
	return &app_interface.ListListingsResponse{List: list}, nil
}

func (s *listingService) Get(ctx context.Context, id int64) (*app_interface.ListingItem, error) {
	row, err := s.listingRepo.FindByID(ctx, id)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, apierrors.ErrNotFound
		}
		return nil, apierrors.ErrInternal
	}
	item := toListingItem(row)
	return &item, nil
}

func (s *listingService) Cancel(ctx context.Context, uid, id int64) error {
	listing, err := s.listingRepo.FindByID(ctx, id)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return apierrors.ErrNotFound
		}
		return apierrors.ErrInternal
	}

	if listing.UserID != uid {
		return apierrors.ErrForbidden
	}

	if listing.Status != "active" && listing.Status != "paused" {
		return apierrors.New(400, "only active or paused listings can be cancelled")
	}

	if listing.Type == "sell" && s.rdb != nil {
		unlock, err := s.rdb.AcquireLock(ctx, walletLockKey(uid, listing.CryptoCurrency), 10*time.Second)
		if err != nil {
			return apierrors.ErrInternal
		}
		defer unlock()
	}

	return s.db.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		if err := s.listingRepo.UpdateStatusInTx(ctx, session, id, "cancelled"); err != nil {
			return apierrors.ErrInternal
		}

		if listing.Type == "sell" && listing.RemainingAmount > 0 {
			if err := s.walletRepo.UnfreezeInTx(ctx, session, uid, listing.CryptoCurrency, listing.RemainingAmount); err != nil {
				return err
			}
		}

		return nil
	})
}

func walletLockKey(userID int64, currency string) string {
	return fmt.Sprintf("lock:wallet:%d:%s", userID, currency)
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

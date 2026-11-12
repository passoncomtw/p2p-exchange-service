package listing_service

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	app_interface "p2p-exchange/cmd/v1-p2p-exchange-service/internal/interfaces/app"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/model/entity"
	listingrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/listing"
	orderrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/order"
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

// errListingNotCancellable 掛單已是終態（cancelled / completed），不可再取消。
var errListingNotCancellable = apierrors.New(400, "only active or paused listings can be cancelled")

type listingService struct {
	db          sqlx.SqlConn
	rdb         *rdb.Client
	listingRepo listingrepo.ListingRepository
	orderRepo   orderrepo.OrderRepository
	walletRepo  walletrepo.WalletRepository
}

func New(db sqlx.SqlConn, rdb *rdb.Client, listingRepo listingrepo.ListingRepository, orderRepo orderrepo.OrderRepository, walletRepo walletrepo.WalletRepository) ListingService {
	return &listingService{db: db, rdb: rdb, listingRepo: listingRepo, orderRepo: orderRepo, walletRepo: walletRepo}
}

func (s *listingService) Create(ctx context.Context, uid int64, req app_interface.CreateListingRequest) (*app_interface.CreateListingResponse, error) {
	if err := validateCreateListingRequest(req); err != nil {
		return nil, err
	}

	if req.Type == "sell" && req.PaymentMethodID == nil {
		return nil, apierrors.New(400, "sell listing requires a payment method")
	}

	if req.Type == "sell" {
		if err := s.walletRepo.Freeze(ctx, uid, req.CryptoCurrency, req.TotalAmount.Decimal); err != nil {
			return nil, err
		}
	}

	l := &entity.Listing{
		UserID:           uid,
		Type:             req.Type,
		CryptoCurrency:   req.CryptoCurrency,
		FiatCurrency:     req.FiatCurrency,
		TotalAmount:      req.TotalAmount.Decimal,
		RemainingAmount:  req.TotalAmount.Decimal,
		Price:            req.Price.Decimal,
		MinOrderFiat:     req.MinOrderFiat.Decimal,
		MaxOrderFiat:     req.MaxOrderFiat.Decimal,
		PlatformFeeBase:  decimal.Zero,
		PlatformFeeRate:  decimal.Zero,
		PaymentFeeBase:   decimal.Zero,
		PaymentFeeRate:   decimal.Zero,
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
		return errListingNotCancellable
	}

	if listing.Type == "sell" && s.rdb != nil {
		unlock, err := s.rdb.AcquireLock(ctx, walletLockKey(uid, listing.CryptoCurrency), 10*time.Second)
		if err != nil {
			return apierrors.ErrInternal
		}
		defer unlock()
	}

	// 以掛單列的 FOR UPDATE 鎖與建單、訂單釋放序列化：鎖內讀到的狀態與 remaining 才是可解凍的額度，
	// 鎖內查到的進行中訂單也不會漏掉並發建立的訂單（建單同樣先鎖這一列）。
	return s.db.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		locked, err := s.listingRepo.FindByIDForUpdateInTx(ctx, session, id)
		if err != nil {
			if err == sqlx.ErrNotFound {
				return apierrors.ErrNotFound
			}
			return apierrors.ErrInternal
		}
		if locked.Status != "active" && locked.Status != "paused" {
			return errListingNotCancellable
		}

		hasActiveOrders, err := s.orderRepo.HasActiveOrdersByListingInTx(ctx, session, id)
		if err != nil {
			return apierrors.ErrInternal
		}
		if hasActiveOrders {
			return apierrors.New(400, "listing has active orders")
		}

		n, err := s.listingRepo.CancelInTx(ctx, session, id)
		if err != nil {
			return apierrors.ErrInternal
		}
		if n == 0 {
			return apierrors.New(409, "listing status changed")
		}

		if locked.Type == "sell" && locked.RemainingAmount.IsPositive() {
			if err := s.walletRepo.UnfreezeInTx(ctx, session, uid, locked.CryptoCurrency, locked.RemainingAmount); err != nil {
				return err
			}
		}

		return nil
	})
}

// validateCreateListingRequest 驗證掛單的數值欄位，必須在凍結餘額之前執行：
// 負數的 totalAmount 會讓 Freeze 反向增加可用餘額，等同憑空造幣；
// 金額字串的格式、小數位數與範圍已由 money.Amount 在反序列化時驗證。
func validateCreateListingRequest(req app_interface.CreateListingRequest) error {
	if !req.TotalAmount.IsPositive() {
		return apierrors.New(400, "totalAmount must be greater than 0")
	}
	if !req.Price.IsPositive() {
		return apierrors.New(400, "price must be greater than 0")
	}
	if req.MinOrderFiat.IsNegative() {
		return apierrors.New(400, "minOrderFiat must not be negative")
	}
	if req.MaxOrderFiat.LessThan(req.MinOrderFiat.Decimal) {
		return apierrors.New(400, "maxOrderFiat must not be less than minOrderFiat")
	}
	if req.PaymentTimeLimit <= 0 {
		return apierrors.New(400, "paymentTimeLimit must be greater than 0")
	}
	return nil
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

package backend_admin_service

import (
	"context"
	"errors"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/config"
	app_interface "p2p-exchange/cmd/v1-p2p-exchange-service/internal/interfaces/app"
	backend_interface "p2p-exchange/cmd/v1-p2p-exchange-service/internal/interfaces/backend"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/model/entity"
	fiatwithdrawrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/fiat_withdraw"
	listingrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/listing"
	orderrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/order"
	userrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/user"
	walletrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/wallet"
	apierrors "p2p-exchange/internal/errors"
)

// errOrderStatusChanged 條件式狀態轉換影響 0 列：爭議訂單在讀取後已被其他請求處理，
// 本次裁決不得再執行任何錢包操作。
var errOrderStatusChanged = apierrors.New(409, "order status changed")

type BackendAdminService interface {
	ListListings(ctx context.Context, req backend_interface.BackendListListingsRequest) (*app_interface.ListListingsResponse, error)
	ListOrders(ctx context.Context, req backend_interface.BackendListOrdersRequest) (*app_interface.ListOrdersResponse, error)
	ResolveOrder(ctx context.Context, req backend_interface.ResolveOrderRequest) error
	// ListMembers 分頁查詢 App 會員。
	ListMembers(ctx context.Context, keyword string, limit, offset int64) (*backend_interface.BackendListMembersResponse, error)
	// DepositToMember 後台手動為會員入金。
	// adminUID 為執行此操作的後台帳號 ID，必須由 handler 從 JWT context 取得後傳入，
	// 用於稽核記錄；絕不可來自 request body。
	DepositToMember(ctx context.Context, adminUID int64, id int64, currency string, amount string) (*backend_interface.BackendDepositResponse, error)
	// GetPlatformWalletInfo 聚合平台 Tron 熱錢包與 ECPay 設定資訊（唯讀，永遠成功回應）。
	GetPlatformWalletInfo(ctx context.Context) (*backend_interface.PlatformWalletInfoResponse, error)
	// ListFiatWithdrawals 分頁查詢法幣提領申請（含完整銀行帳號，供後台審核核對）。
	ListFiatWithdrawals(ctx context.Context, status string, limit, offset int64) (*backend_interface.BackendListFiatWithdrawalsResponse, error)
	// ReviewFiatWithdrawal 審核法幣提領申請（approve 扣款 / reject 解凍）。
	// reviewerID 為審核人（後台帳號 ID），必須由 handler 從 JWT context 取得後傳入；
	// 絕不可來自 request body。已審核過的申請回傳 409。
	ReviewFiatWithdrawal(ctx context.Context, reviewerID, id int64, action, reason string) error
}

type backendAdminService struct {
	db               sqlx.SqlConn
	cfg              *config.Config
	listingRepo      listingrepo.ListingRepository
	orderRepo        orderrepo.OrderRepository
	walletRepo       walletrepo.WalletRepository
	userRepo         userrepo.UserRepository
	fiatWithdrawRepo fiatwithdrawrepo.FiatWithdrawRepository
}

func New(
	db sqlx.SqlConn,
	cfg *config.Config,
	listingRepo listingrepo.ListingRepository,
	orderRepo orderrepo.OrderRepository,
	walletRepo walletrepo.WalletRepository,
	userRepo userrepo.UserRepository,
	fiatWithdrawRepo fiatwithdrawrepo.FiatWithdrawRepository,
) BackendAdminService {
	return &backendAdminService{
		db:               db,
		cfg:              cfg,
		listingRepo:      listingRepo,
		orderRepo:        orderRepo,
		walletRepo:       walletRepo,
		userRepo:         userRepo,
		fiatWithdrawRepo: fiatWithdrawRepo,
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

func (s *backendAdminService) ResolveOrder(ctx context.Context, req backend_interface.ResolveOrderRequest) error {
	order, err := s.orderRepo.FindByID(ctx, req.ID)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return apierrors.ErrNotFound
		}
		return apierrors.ErrInternal
	}

	const fromStatus = "disputed"
	if order.Status != fromStatus {
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
			n, err := s.orderRepo.UpdateStatusInTx(ctx, session, order.ID, fromStatus, "completed", map[string]interface{}{
				"completed_at": "NOW()",
			})
			if err != nil {
				return err
			}
			if n == 0 {
				return errOrderStatusChanged
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
			n, err := s.orderRepo.UpdateStatusInTx(ctx, session, order.ID, fromStatus, "refunded", map[string]interface{}{
				"cancel_reason": req.Reason,
			})
			if err != nil {
				return err
			}
			if n == 0 {
				return errOrderStatusChanged
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

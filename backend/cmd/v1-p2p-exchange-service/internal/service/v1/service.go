package v1_service

import (
	"context"
	"strconv"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	v1_interface "p2p-exchange/cmd/v1-p2p-exchange-service/internal/interfaces/v1"
	userrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/user"
	v1orderrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/v1order"
	apierrors "p2p-exchange/internal/errors"
)

const demoUsername = "demo_user"

const (
	defaultAsset = "USDT"
	defaultFiat  = "TWD"
)

var validPaymentMethods = map[string]bool{
	"bank_transfer":     true,
	"convenience_store": true,
}

type V1Service interface {
	Create(ctx context.Context, req v1_interface.V1CreateOrderRequest) (*v1_interface.V1Order, error)
	ListMine(ctx context.Context) (*v1_interface.V1OrderListResponse, error)
	Cancel(ctx context.Context, id int64) error
	AdminList(ctx context.Context, apiStatus string) (*v1_interface.V1OrderListResponse, error)
	AdminGet(ctx context.Context, id int64) (*v1_interface.V1Order, error)
	AdminComplete(ctx context.Context, id int64) error
}

type v1Service struct {
	userRepo     userrepo.UserRepository
	v1OrderRepo  v1orderrepo.V1OrderRepository
}

func New(userRepo userrepo.UserRepository, v1OrderRepo v1orderrepo.V1OrderRepository) V1Service {
	return &v1Service{
		userRepo:    userRepo,
		v1OrderRepo: v1OrderRepo,
	}
}

func (s *v1Service) demoUserID(ctx context.Context) (int64, error) {
	user, err := s.userRepo.FindByUsername(ctx, demoUsername)
	if err != nil {
		return 0, apierrors.New(500, "demo_user not seeded")
	}
	return user.ID, nil
}

func (s *v1Service) Create(ctx context.Context, req v1_interface.V1CreateOrderRequest) (*v1_interface.V1Order, error) {
	if err := validateCreate(req); err != nil {
		return nil, err
	}

	uid, err := s.demoUserID(ctx)
	if err != nil {
		return nil, err
	}

	id, err := s.v1OrderRepo.Create(ctx, &v1orderrepo.V1CreateParams{
		UserID:             uid,
		Type:               req.Type,
		CryptoCurrency:     req.Asset,
		FiatCurrency:       req.Fiat,
		Price:              req.Price,
		Quantity:           req.Quantity,
		PaymentMethodLabel: req.PaymentMethod,
	})
	if err != nil {
		return nil, apierrors.ErrInternal
	}

	row, err := s.v1OrderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, apierrors.ErrInternal
	}
	order := rowToV1Order(row)
	return &order, nil
}

func (s *v1Service) ListMine(ctx context.Context) (*v1_interface.V1OrderListResponse, error) {
	uid, err := s.demoUserID(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := s.v1OrderRepo.ListByUserID(ctx, uid)
	if err != nil {
		return nil, apierrors.ErrInternal
	}
	return &v1_interface.V1OrderListResponse{List: mapRows(rows)}, nil
}

func (s *v1Service) Cancel(ctx context.Context, id int64) error {
	row, err := s.findRow(ctx, id)
	if err != nil {
		return err
	}

	if row.Username != demoUsername {
		return apierrors.ErrForbidden
	}
	if dbStatusToAPI(row.Status) != "open" {
		return apierrors.New(400, "only open orders can be cancelled")
	}

	if err := s.v1OrderRepo.UpdateStatus(ctx, id, apiStatusToDB("cancelled")); err != nil {
		return apierrors.ErrInternal
	}
	return nil
}

func (s *v1Service) AdminList(ctx context.Context, apiStatus string) (*v1_interface.V1OrderListResponse, error) {
	dbStatus := ""
	if apiStatus != "" {
		dbStatus = apiStatusToDB(apiStatus)
	}
	rows, err := s.v1OrderRepo.ListAll(ctx, dbStatus)
	if err != nil {
		return nil, apierrors.ErrInternal
	}
	return &v1_interface.V1OrderListResponse{List: mapRows(rows)}, nil
}

func (s *v1Service) AdminGet(ctx context.Context, id int64) (*v1_interface.V1Order, error) {
	row, err := s.findRow(ctx, id)
	if err != nil {
		return nil, err
	}
	order := rowToV1Order(row)
	return &order, nil
}

func (s *v1Service) AdminComplete(ctx context.Context, id int64) error {
	row, err := s.findRow(ctx, id)
	if err != nil {
		return err
	}
	if dbStatusToAPI(row.Status) != "open" {
		return apierrors.New(400, "only open orders can be completed")
	}
	if err := s.v1OrderRepo.UpdateStatus(ctx, id, apiStatusToDB("completed")); err != nil {
		return apierrors.ErrInternal
	}
	return nil
}

func (s *v1Service) findRow(ctx context.Context, id int64) (*v1orderrepo.V1OrderRow, error) {
	row, err := s.v1OrderRepo.FindByID(ctx, id)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, apierrors.ErrNotFound
		}
		return nil, apierrors.ErrInternal
	}
	return row, nil
}

func rowToV1Order(r *v1orderrepo.V1OrderRow) v1_interface.V1Order {
	paymentMethod := "bank_transfer"
	if r.PaymentMethodLabel != nil && *r.PaymentMethodLabel != "" {
		paymentMethod = *r.PaymentMethodLabel
	}
	return v1_interface.V1Order{
		ID:            strconv.FormatInt(r.ID, 10),
		Type:          r.Type,
		Asset:         r.CryptoCurrency,
		Fiat:          r.FiatCurrency,
		Price:         r.Price,
		Quantity:      r.Quantity,
		TotalAmount:   r.Price * r.Quantity,
		PaymentMethod: paymentMethod,
		Status:        dbStatusToAPI(r.Status),
		CreatedBy:     r.Username,
		CreatedAt:     r.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     r.UpdatedAt.Format(time.RFC3339),
	}
}

func mapRows(rows []*v1orderrepo.V1OrderRow) []v1_interface.V1Order {
	list := make([]v1_interface.V1Order, 0, len(rows))
	for _, r := range rows {
		list = append(list, rowToV1Order(r))
	}
	return list
}

func apiStatusToDB(api string) string {
	switch api {
	case "open":
		return "active"
	default:
		return api
	}
}

func dbStatusToAPI(db string) string {
	switch db {
	case "active", "paused":
		return "open"
	default:
		return db
	}
}

func validateCreate(req v1_interface.V1CreateOrderRequest) error {
	if req.Type != "buy" && req.Type != "sell" {
		return apierrors.New(400, "invalid type")
	}
	if req.Asset != defaultAsset {
		return apierrors.New(400, "invalid asset")
	}
	if req.Fiat != defaultFiat {
		return apierrors.New(400, "invalid fiat")
	}
	if req.Price <= 0 {
		return apierrors.New(400, "price must be greater than 0")
	}
	if req.Quantity <= 0 {
		return apierrors.New(400, "quantity must be greater than 0")
	}
	if !validPaymentMethods[req.PaymentMethod] {
		return apierrors.New(400, "invalid payment method")
	}
	return nil
}

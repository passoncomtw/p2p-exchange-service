package payment_service

import (
	"context"

	app_interface "p2p-exchange/cmd/v1-p2p-exchange-service/internal/interfaces/app"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/model/entity"
	paymentrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/payment_method"
)

type PaymentMethodService interface {
	Create(ctx context.Context, uid int64, req app_interface.CreatePaymentMethodRequest) (*app_interface.CreatePaymentMethodResponse, error)
	List(ctx context.Context, uid int64) (*app_interface.ListPaymentMethodsResponse, error)
	Delete(ctx context.Context, uid, id int64) error
}

type paymentMethodService struct {
	repo paymentrepo.PaymentMethodRepository
}

func New(repo paymentrepo.PaymentMethodRepository) PaymentMethodService {
	return &paymentMethodService{repo: repo}
}

func (s *paymentMethodService) Create(ctx context.Context, uid int64, req app_interface.CreatePaymentMethodRequest) (*app_interface.CreatePaymentMethodResponse, error) {
	pm := &entity.PaymentMethod{
		UserID:        uid,
		Type:          req.Type,
		BankName:      req.BankName,
		AccountName:   req.AccountName,
		AccountNumber: req.AccountNumber,
		IsActive:      true,
	}
	id, err := s.repo.Create(ctx, pm)
	if err != nil {
		return nil, err
	}
	return &app_interface.CreatePaymentMethodResponse{ID: id}, nil
}

func (s *paymentMethodService) List(ctx context.Context, uid int64) (*app_interface.ListPaymentMethodsResponse, error) {
	rows, err := s.repo.FindByUserID(ctx, uid)
	if err != nil {
		return nil, err
	}
	list := make([]app_interface.PaymentMethodItem, 0, len(rows))
	for _, pm := range rows {
		list = append(list, app_interface.PaymentMethodItem{
			ID:            pm.ID,
			Type:          pm.Type,
			BankName:      pm.BankName,
			AccountName:   pm.AccountName,
			AccountNumber: pm.AccountNumber,
			IsActive:      pm.IsActive,
		})
	}
	return &app_interface.ListPaymentMethodsResponse{List: list}, nil
}

func (s *paymentMethodService) Delete(ctx context.Context, uid, id int64) error {
	return s.repo.SoftDelete(ctx, id, uid)
}

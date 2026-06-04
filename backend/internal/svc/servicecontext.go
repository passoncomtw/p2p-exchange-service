package svc

import (
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"p2p-exchange/internal/config"
	"p2p-exchange/internal/model"
)

type ServiceContext struct {
	Config          config.Config
	AppUser         *model.AppUserModel
	BackendUser     *model.BackendUserModel
	PaymentMethod   *model.PaymentMethodModel
	Listing         *model.ListingModel
	Order           *model.OrderModel
	EscrowRecord    *model.EscrowRecordModel
	OrderStatusLog  *model.OrderStatusLogModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewSqlConn("pgx", c.Database.DSN)
	return &ServiceContext{
		Config:         c,
		AppUser:        model.NewAppUserModel(conn),
		BackendUser:    model.NewBackendUserModel(conn),
		PaymentMethod:  model.NewPaymentMethodModel(conn),
		Listing:        model.NewListingModel(conn),
		Order:          model.NewOrderModel(conn),
		EscrowRecord:   model.NewEscrowRecordModel(conn),
		OrderStatusLog: model.NewOrderStatusLogModel(conn),
	}
}

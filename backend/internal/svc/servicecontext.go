// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"p2p-exchange/internal/config"
	"p2p-exchange/internal/model"
)

type ServiceContext struct {
	Config      config.Config
	AppUser     *model.AppUserModel
	BackendUser *model.BackendUserModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewSqlConn("pgx", c.Database.DSN)
	return &ServiceContext{
		Config:      c,
		AppUser:     model.NewAppUserModel(conn),
		BackendUser: model.NewBackendUserModel(conn),
	}
}

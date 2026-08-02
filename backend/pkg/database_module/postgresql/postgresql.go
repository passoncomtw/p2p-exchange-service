package postgresql

import (
	"fmt"
	"p2p-exchange/internal/config"

	_ "github.com/jackc/pgx/v5/stdlib" // ← 必須，註冊 pgx driver
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"go.uber.org/fx"
)

var Module = fx.Module("postgresql",
	fx.Provide(NewPostgreSQL),
)

func NewPostgreSQL(config config.Config) sqlx.SqlConn {
	conn := sqlx.NewSqlConn("pgx", config.Database.DSN)
	fmt.Println("postgresql is connected")
	return conn
}

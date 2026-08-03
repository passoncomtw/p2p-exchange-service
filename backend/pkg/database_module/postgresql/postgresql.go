package postgresql

import (
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"go.uber.org/fx"
)

var Module = fx.Module("postgresql",
	fx.Provide(NewPostgreSQL),
)

type Config struct {
	DSN string
}

func NewPostgreSQL(cfg Config) sqlx.SqlConn {
	conn := sqlx.NewSqlConn("pgx", cfg.DSN)
	fmt.Println("postgresql is connected")
	return conn
}

package model

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type OrderStatusLog struct {
	ID           int64     `db:"id"`
	OrderID      int64     `db:"order_id"`
	FromStatus   *string   `db:"from_status"`
	ToStatus     string    `db:"to_status"`
	OperatorType string    `db:"operator_type"`
	OperatorID   *int64    `db:"operator_id"`
	Remark       *string   `db:"remark"`
	CreatedAt    time.Time `db:"created_at"`
}

type OrderStatusLogModel struct {
	conn sqlx.SqlConn
}

func NewOrderStatusLogModel(conn sqlx.SqlConn) *OrderStatusLogModel {
	return &OrderStatusLogModel{conn: conn}
}

func (m *OrderStatusLogModel) Append(ctx context.Context, log *OrderStatusLog) error {
	_, err := m.conn.ExecCtx(ctx,
		`INSERT INTO order_status_logs (order_id, from_status, to_status, operator_type, operator_id, remark, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, NOW())`,
		log.OrderID, log.FromStatus, log.ToStatus, log.OperatorType, log.OperatorID, log.Remark,
	)
	return err
}

func (m *OrderStatusLogModel) AppendInTx(ctx context.Context, session sqlx.Session, log *OrderStatusLog) error {
	_, err := session.ExecCtx(ctx,
		`INSERT INTO order_status_logs (order_id, from_status, to_status, operator_type, operator_id, remark, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, NOW())`,
		log.OrderID, log.FromStatus, log.ToStatus, log.OperatorType, log.OperatorID, log.Remark,
	)
	return err
}

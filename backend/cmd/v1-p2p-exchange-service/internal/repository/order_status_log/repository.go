package orderstatuslogrepo

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type OrderStatusLog struct {
	OrderID      int64
	FromStatus   *string
	ToStatus     string
	OperatorType string
	OperatorID   *int64
	Remark       *string
}

type OrderStatusLogRepository interface {
	Append(ctx context.Context, log *OrderStatusLog) error
}

type orderStatusLogRepository struct {
	db sqlx.SqlConn
}

func New(db sqlx.SqlConn) OrderStatusLogRepository {
	return &orderStatusLogRepository{db: db}
}

func (r *orderStatusLogRepository) Append(ctx context.Context, log *OrderStatusLog) error {
	_, err := r.db.ExecCtx(ctx,
		`INSERT INTO order_status_logs (order_id, from_status, to_status, operator_type, operator_id, remark, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, NOW())`,
		log.OrderID, log.FromStatus, log.ToStatus, log.OperatorType, log.OperatorID, log.Remark,
	)
	return err
}

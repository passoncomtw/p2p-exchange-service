package paymentrepo

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/model/entity"
)

type PaymentMethodRepository interface {
	Create(ctx context.Context, pm *entity.PaymentMethod) (int64, error)
	FindByUserID(ctx context.Context, userID int64) ([]*entity.PaymentMethod, error)
	SoftDelete(ctx context.Context, id, userID int64) error
}

type paymentMethodRepository struct {
	db sqlx.SqlConn
}

func New(db sqlx.SqlConn) PaymentMethodRepository {
	return &paymentMethodRepository{db: db}
}

func (r *paymentMethodRepository) Create(ctx context.Context, pm *entity.PaymentMethod) (int64, error) {
	var id int64
	err := r.db.QueryRowCtx(ctx, &id,
		`INSERT INTO payment_methods (user_id, type, bank_name, account_name, account_number, is_active, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		 RETURNING id`,
		pm.UserID, pm.Type, pm.BankName, pm.AccountName, pm.AccountNumber, pm.IsActive,
	)
	return id, err
}

func (r *paymentMethodRepository) FindByUserID(ctx context.Context, userID int64) ([]*entity.PaymentMethod, error) {
	var rows []*entity.PaymentMethod
	err := r.db.QueryRowsCtx(ctx, &rows,
		`SELECT id, user_id, type, bank_name, account_name, account_number, is_active, created_at, updated_at
		 FROM payment_methods
		 WHERE user_id = $1 AND is_active = true
		 ORDER BY created_at DESC`,
		userID,
	)
	return rows, err
}

func (r *paymentMethodRepository) SoftDelete(ctx context.Context, id, userID int64) error {
	_, err := r.db.ExecCtx(ctx,
		`UPDATE payment_methods SET is_active = false, updated_at = NOW() WHERE id = $1 AND user_id = $2`,
		id, userID,
	)
	return err
}

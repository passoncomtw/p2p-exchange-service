package logic

import (
	"context"
	"strconv"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	apierrors "p2p-exchange/internal/errors"
	"p2p-exchange/internal/model"
	"p2p-exchange/internal/svc"
	"p2p-exchange/internal/types"
)

// v1 目前固定使用者（無登入）。seed 時須建立此 app_user。
const demoUsername = "demo_user"

// v1 預設幣別與法幣。
const (
	defaultAsset = "USDT"
	defaultFiat  = "TWD"
)

// 允許的付款方式。
var validPaymentMethods = map[string]bool{
	"bank_transfer":     true,
	"convenience_store": true,
}

// ── 狀態映射：API（open/completed/cancelled）↔ DB（active/completed/cancelled） ──

func apiStatusToDB(api string) string {
	switch api {
	case "open":
		return "active"
	default:
		return api // completed / cancelled 兩端一致
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

func rowToV1Order(r *model.V1OrderRow) types.V1Order {
	paymentMethod := "bank_transfer"
	if r.PaymentMethodLabel != nil && *r.PaymentMethodLabel != "" {
		paymentMethod = *r.PaymentMethodLabel
	}
	return types.V1Order{
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

// ── V1OrderLogic ──────────────────────────────────────────────────────────────

type V1OrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewV1OrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *V1OrderLogic {
	return &V1OrderLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// demoUserID 取得 demo_user 的 id；未 seed 時回傳錯誤提示。
func (l *V1OrderLogic) demoUserID() (int64, error) {
	user, err := l.svcCtx.AppUser.FindByUsername(l.ctx, demoUsername)
	if err != nil {
		l.Errorf("demo_user not found, run seed first: %v", err)
		return 0, apierrors.New(500, "demo_user not seeded")
	}
	return user.ID, nil
}

// Create 建立一筆 open 掛單（createdBy 固定 demo_user）。
func (l *V1OrderLogic) Create(req *types.V1CreateOrderRequest) (*types.V1Order, error) {
	if err := validateCreate(req); err != nil {
		return nil, err
	}

	uid, err := l.demoUserID()
	if err != nil {
		return nil, err
	}

	id, err := l.svcCtx.V1Order.Create(l.ctx, &model.V1CreateParams{
		UserID:             uid,
		Type:               req.Type,
		CryptoCurrency:     req.Asset,
		FiatCurrency:       req.Fiat,
		Price:              req.Price,
		Quantity:           req.Quantity,
		PaymentMethodLabel: req.PaymentMethod,
	})
	if err != nil {
		l.Errorf("create v1 order failed: %v", err)
		return nil, apierrors.ErrInternal
	}

	row, err := l.svcCtx.V1Order.FindByID(l.ctx, id)
	if err != nil {
		l.Errorf("reload v1 order id=%d failed: %v", id, err)
		return nil, apierrors.ErrInternal
	}
	order := rowToV1Order(row)
	return &order, nil
}

// ListMine 取得 demo_user 自己的掛單。
func (l *V1OrderLogic) ListMine() (*types.V1OrderListResponse, error) {
	uid, err := l.demoUserID()
	if err != nil {
		return nil, err
	}

	rows, err := l.svcCtx.V1Order.ListByUserID(l.ctx, uid)
	if err != nil {
		l.Errorf("list my v1 orders failed: %v", err)
		return nil, apierrors.ErrInternal
	}
	return &types.V1OrderListResponse{List: mapRows(rows)}, nil
}

// Cancel 取消 open 掛單（僅 demo_user 自己的訂單）。
func (l *V1OrderLogic) Cancel(id int64) error {
	row, err := l.findRow(id)
	if err != nil {
		return err
	}

	if row.Username != demoUsername {
		return apierrors.ErrForbidden
	}
	if dbStatusToAPI(row.Status) != "open" {
		return apierrors.New(400, "only open orders can be cancelled")
	}

	if err := l.svcCtx.V1Order.UpdateStatus(l.ctx, id, apiStatusToDB("cancelled")); err != nil {
		l.Errorf("cancel v1 order id=%d failed: %v", id, err)
		return apierrors.ErrInternal
	}
	return nil
}

// AdminList 取得全部掛單，可依 API 狀態篩選。
func (l *V1OrderLogic) AdminList(apiStatus string) (*types.V1OrderListResponse, error) {
	dbStatus := ""
	if apiStatus != "" {
		dbStatus = apiStatusToDB(apiStatus)
	}
	rows, err := l.svcCtx.V1Order.ListAll(l.ctx, dbStatus)
	if err != nil {
		l.Errorf("admin list v1 orders failed: %v", err)
		return nil, apierrors.ErrInternal
	}
	return &types.V1OrderListResponse{List: mapRows(rows)}, nil
}

// AdminGet 取得單筆掛單詳情。
func (l *V1OrderLogic) AdminGet(id int64) (*types.V1Order, error) {
	row, err := l.findRow(id)
	if err != nil {
		return nil, err
	}
	order := rowToV1Order(row)
	return &order, nil
}

// AdminComplete 將 open 掛單標記為 completed。
func (l *V1OrderLogic) AdminComplete(id int64) error {
	row, err := l.findRow(id)
	if err != nil {
		return err
	}
	if dbStatusToAPI(row.Status) != "open" {
		return apierrors.New(400, "only open orders can be completed")
	}
	if err := l.svcCtx.V1Order.UpdateStatus(l.ctx, id, apiStatusToDB("completed")); err != nil {
		l.Errorf("complete v1 order id=%d failed: %v", id, err)
		return apierrors.ErrInternal
	}
	return nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

func (l *V1OrderLogic) findRow(id int64) (*model.V1OrderRow, error) {
	row, err := l.svcCtx.V1Order.FindByID(l.ctx, id)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, apierrors.ErrNotFound
		}
		l.Errorf("find v1 order id=%d failed: %v", id, err)
		return nil, apierrors.ErrInternal
	}
	return row, nil
}

func mapRows(rows []*model.V1OrderRow) []types.V1Order {
	list := make([]types.V1Order, 0, len(rows))
	for _, r := range rows {
		list = append(list, rowToV1Order(r))
	}
	return list
}

// validateCreate 伺服器端輸入驗證（與 shared/validation 對齊）。
func validateCreate(req *types.V1CreateOrderRequest) error {
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

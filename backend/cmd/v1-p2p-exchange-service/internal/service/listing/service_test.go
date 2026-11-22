package listing_service

import (
	"context"
	"errors"
	"github.com/shopspring/decimal"
	"p2p-exchange/internal/money"
	"testing"

	app_interface "p2p-exchange/cmd/v1-p2p-exchange-service/internal/interfaces/app"
	"p2p-exchange/cmd/v1-p2p-exchange-service/internal/model/entity"
	listingrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/listing"
	walletrepo "p2p-exchange/cmd/v1-p2p-exchange-service/internal/repository/wallet"
	apierrors "p2p-exchange/internal/errors"
)

const testUID int64 = 1001

// fakeWalletRepo 只實作 Freeze 並記錄呼叫次數，其餘方法沿用內嵌的 nil 介面（誤用即 panic）。
type fakeWalletRepo struct {
	walletrepo.WalletRepository
	freezeCalls int
}

func (r *fakeWalletRepo) Freeze(_ context.Context, _ int64, _ string, _ decimal.Decimal) error {
	r.freezeCalls++
	return nil
}

// fakeListingRepo 只實作 Create 並記錄呼叫次數。
type fakeListingRepo struct {
	listingrepo.ListingRepository
	createCalls int
}

func (r *fakeListingRepo) Create(_ context.Context, _ *entity.Listing) (int64, error) {
	r.createCalls++
	return 1, nil
}

// amt 以測試用的字串建立 API 金額型別。
func amt(s string) money.Amount {
	return money.Amount{Decimal: decimal.RequireFromString(s)}
}

func validSellListingRequest() app_interface.CreateListingRequest {
	paymentMethodID := int64(5)
	return app_interface.CreateListingRequest{
		Type:             "sell",
		CryptoCurrency:   "USDT",
		FiatCurrency:     "TWD",
		TotalAmount:      amt("100"),
		Price:            amt("32.5"),
		MinOrderFiat:     amt("100"),
		MaxOrderFiat:     amt("3250"),
		PaymentTimeLimit: 15,
		PaymentMethodID:  &paymentMethodID,
	}
}

// TestCreateListingValidatesAmounts 迴歸測試：掛單的數值欄位必須在 Freeze 之前驗證。
// 重現場景：已登入使用者送出 totalAmount=-1000 的賣幣掛單，legacy 直接以負數呼叫 Freeze，
// `available_balance - (-1000)` 讓可用餘額憑空增加 1000。
// 修正後所有非法輸入回 400，且完全不會呼叫 Freeze 或寫入掛單。
func TestCreateListingValidatesAmounts(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*app_interface.CreateListingRequest)
	}{
		{name: "negative totalAmount", mutate: func(r *app_interface.CreateListingRequest) { r.TotalAmount = amt("-1000") }},
		{name: "zero totalAmount", mutate: func(r *app_interface.CreateListingRequest) { r.TotalAmount = amt("0") }},
		{name: "negative price", mutate: func(r *app_interface.CreateListingRequest) { r.Price = amt("-1") }},
		{name: "zero price", mutate: func(r *app_interface.CreateListingRequest) { r.Price = amt("0") }},
		{name: "negative minOrderFiat", mutate: func(r *app_interface.CreateListingRequest) { r.MinOrderFiat = amt("-1") }},
		{name: "min greater than max", mutate: func(r *app_interface.CreateListingRequest) { r.MinOrderFiat, r.MaxOrderFiat = amt("500"), amt("100") }},
		{name: "zero paymentTimeLimit", mutate: func(r *app_interface.CreateListingRequest) { r.PaymentTimeLimit = 0 }},
		{name: "negative paymentTimeLimit", mutate: func(r *app_interface.CreateListingRequest) { r.PaymentTimeLimit = -5 }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wallet := &fakeWalletRepo{}
			listings := &fakeListingRepo{}
			svc := New(nil, nil, listings, nil, wallet)

			req := validSellListingRequest()
			tc.mutate(&req)

			_, err := svc.Create(context.Background(), testUID, req)
			if code := errCode(t, err); code != 400 {
				t.Fatalf("expected 400, got %d (%v)", code, err)
			}
			if wallet.freezeCalls != 0 {
				t.Fatalf("Freeze must not be called for invalid input, got %d calls", wallet.freezeCalls)
			}
			if listings.createCalls != 0 {
				t.Fatalf("listing must not be created for invalid input, got %d calls", listings.createCalls)
			}
		})
	}
}

// TestCreateListingAcceptsValidInput 合法輸入：賣幣掛單凍結一次並建立掛單；min == max 與 min == 0 屬合法邊界。
func TestCreateListingAcceptsValidInput(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*app_interface.CreateListingRequest)
	}{
		{name: "typical", mutate: func(*app_interface.CreateListingRequest) {}},
		{name: "min equals max", mutate: func(r *app_interface.CreateListingRequest) { r.MinOrderFiat, r.MaxOrderFiat = amt("1000"), amt("1000") }},
		{name: "zero minOrderFiat", mutate: func(r *app_interface.CreateListingRequest) { r.MinOrderFiat = amt("0") }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wallet := &fakeWalletRepo{}
			listings := &fakeListingRepo{}
			svc := New(nil, nil, listings, nil, wallet)

			req := validSellListingRequest()
			tc.mutate(&req)

			resp, err := svc.Create(context.Background(), testUID, req)
			if err != nil {
				t.Fatalf("expected success, got %v", err)
			}
			if resp == nil || resp.ID != 1 {
				t.Fatalf("unexpected response: %+v", resp)
			}
			if wallet.freezeCalls != 1 {
				t.Fatalf("expected exactly 1 Freeze, got %d", wallet.freezeCalls)
			}
			if listings.createCalls != 1 {
				t.Fatalf("expected exactly 1 listing create, got %d", listings.createCalls)
			}
		})
	}
}

func errCode(t *testing.T, err error) int {
	t.Helper()
	var appErr *apierrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected *apierrors.AppError, got %T: %v", err, err)
	}
	return appErr.Code
}

package app_interface

import "github.com/shopspring/decimal"

type CreateOrderResponse struct {
	ID      int64  `json:"id"`
	OrderNo string `json:"orderNo"`
}

type OrderItem struct {
	ID                int64           `json:"id"`
	OrderNo           string          `json:"orderNo"`
	ListingID         int64           `json:"listingId"`
	ListingType       string          `json:"listingType"`
	SellerID          int64           `json:"sellerId"`
	BuyerID           int64           `json:"buyerId"`
	CryptoCurrency    string          `json:"cryptoCurrency"`
	FiatCurrency      string          `json:"fiatCurrency"`
	CryptoAmount      decimal.Decimal `json:"cryptoAmount"`
	Price             decimal.Decimal `json:"price"`
	FiatAmount        decimal.Decimal `json:"fiatAmount"`
	PlatformFeeBase   decimal.Decimal `json:"platformFeeBase"`
	PlatformFeeAmount decimal.Decimal `json:"platformFeeAmount"`
	PaymentFeeBase    decimal.Decimal `json:"paymentFeeBase"`
	PaymentFeeAmount  decimal.Decimal `json:"paymentFeeAmount"`
	TotalFee          decimal.Decimal `json:"totalFee"`
	TotalAmount       decimal.Decimal `json:"totalAmount"`
	PaymentMethodID   int64           `json:"paymentMethodId"`
	Status            string          `json:"status"`
	PaymentDeadline   string          `json:"paymentDeadline"`
	PaidAt            *string         `json:"paidAt,omitempty"`
	ConfirmedAt       *string         `json:"confirmedAt,omitempty"`
	CompletedAt       *string         `json:"completedAt,omitempty"`
	CancelledAt       *string         `json:"cancelledAt,omitempty"`
	CancelReason      *string         `json:"cancelReason,omitempty"`
	CreatedAt         string          `json:"createdAt"`
}

type ListOrdersResponse struct {
	List []OrderItem `json:"list"`
}

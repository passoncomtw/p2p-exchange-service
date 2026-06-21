package types

// ── health check (goctl scaffold) ────────────────────────────────────────────

type Request struct {
	Name string `path:"name,options=you|me"`
}

type Response struct {
	Message string `json:"message"`
}

// ── auth ─────────────────────────────────────────────────────────────────────

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresIn int64  `json:"expiresIn"`
}

// AppLoginResponse — App 端登入回應，包含 user info 供 App 初始化 store 使用
type AppLoginUserInfo struct {
	ID      int64  `json:"id"`
	Account string `json:"account"` // = username
	Name    string `json:"name"`    // 第一階段同 username，未來可獨立
}

type AppLoginResponse struct {
	AccessToken string           `json:"access_token"`
	ExpireIn    int64            `json:"expireIn"`
	User        AppLoginUserInfo `json:"user"`
}

// ── app profile ───────────────────────────────────────────────────────────────

type ProfileRequest struct{}

type ProfileResponse struct {
	Username string `json:"username"`
}

// ── backend dashboard ─────────────────────────────────────────────────────────

type DashboardResponse struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}

// ── payment methods ───────────────────────────────────────────────────────────

type CreatePaymentMethodRequest struct {
	Type          string `json:"type"`
	BankName      string `json:"bankName"`
	AccountName   string `json:"accountName"`
	AccountNumber string `json:"accountNumber"`
}

type CreatePaymentMethodResponse struct {
	ID int64 `json:"id"`
}

type PaymentMethodItem struct {
	ID            int64  `json:"id"`
	Type          string `json:"type"`
	BankName      string `json:"bankName"`
	AccountName   string `json:"accountName"`
	AccountNumber string `json:"accountNumber"`
	IsActive      bool   `json:"isActive"`
}

type ListPaymentMethodsResponse struct {
	List []PaymentMethodItem `json:"list"`
}

// ── listings ──────────────────────────────────────────────────────────────────

type CreateListingRequest struct {
	Type             string  `json:"type"`
	CryptoCurrency   string  `json:"cryptoCurrency"`
	FiatCurrency     string  `json:"fiatCurrency"`
	TotalAmount      float64 `json:"totalAmount"`
	Price            float64 `json:"price"`
	MinOrderFiat     float64 `json:"minOrderFiat"`
	MaxOrderFiat     float64 `json:"maxOrderFiat"`
	PaymentTimeLimit int64   `json:"paymentTimeLimit"`
	PaymentMethodID  *int64  `json:"paymentMethodId,optional"`
}

type CreateListingResponse struct {
	ID int64 `json:"id"`
}

type ListingItem struct {
	ID               int64   `json:"id"`
	UserID           int64   `json:"userId"`
	Type             string  `json:"type"`
	CryptoCurrency   string  `json:"cryptoCurrency"`
	FiatCurrency     string  `json:"fiatCurrency"`
	TotalAmount      float64 `json:"totalAmount"`
	RemainingAmount  float64 `json:"remainingAmount"`
	Price            float64 `json:"price"`
	MinOrderFiat     float64 `json:"minOrderFiat"`
	MaxOrderFiat     float64 `json:"maxOrderFiat"`
	PlatformFeeBase  float64 `json:"platformFeeBase"`
	PlatformFeeRate  float64 `json:"platformFeeRate"`
	PaymentFeeBase   float64 `json:"paymentFeeBase"`
	PaymentFeeRate   float64 `json:"paymentFeeRate"`
	PaymentTimeLimit int64   `json:"paymentTimeLimit"`
	PaymentMethodID  *int64  `json:"paymentMethodId"`
	Status           string  `json:"status"`
	CreatedAt        string  `json:"createdAt"`
}

type ListListingsRequest struct {
	Type   string `form:"type,optional"`
	Status string `form:"status,optional"`
	Limit  int64  `form:"limit,optional,default=20"`
	Offset int64  `form:"offset,optional,default=0"`
}

type ListListingsResponse struct {
	List []ListingItem `json:"list"`
}

type CancelListingRequest struct {
	ID int64 `path:"id"`
}

// ── orders ────────────────────────────────────────────────────────────────────

type CreateOrderRequest struct {
	ListingID    int64   `json:"listingId"`
	CryptoAmount float64 `json:"cryptoAmount"`
}

type CreateOrderResponse struct {
	ID      int64  `json:"id"`
	OrderNo string `json:"orderNo"`
}

type OrderItem struct {
	ID                int64   `json:"id"`
	OrderNo           string  `json:"orderNo"`
	ListingID         int64   `json:"listingId"`
	ListingType       string  `json:"listingType"`
	SellerID          int64   `json:"sellerId"`
	BuyerID           int64   `json:"buyerId"`
	CryptoCurrency    string  `json:"cryptoCurrency"`
	FiatCurrency      string  `json:"fiatCurrency"`
	CryptoAmount      float64 `json:"cryptoAmount"`
	Price             float64 `json:"price"`
	FiatAmount        float64 `json:"fiatAmount"`
	PlatformFeeBase   float64 `json:"platformFeeBase"`
	PlatformFeeAmount float64 `json:"platformFeeAmount"`
	PaymentFeeBase    float64 `json:"paymentFeeBase"`
	PaymentFeeAmount  float64 `json:"paymentFeeAmount"`
	TotalFee          float64 `json:"totalFee"`
	TotalAmount       float64 `json:"totalAmount"`
	PaymentMethodID   int64   `json:"paymentMethodId"`
	Status            string  `json:"status"`
	PaymentDeadline   string  `json:"paymentDeadline"`
	PaidAt            *string `json:"paidAt"`
	ConfirmedAt       *string `json:"confirmedAt"`
	CompletedAt       *string `json:"completedAt"`
	CancelledAt       *string `json:"cancelledAt"`
	CancelReason      *string `json:"cancelReason"`
	CreatedAt         string  `json:"createdAt"`
}

type GetOrderRequest struct {
	ID int64 `path:"id"`
}

type GetOrderResponse struct {
	OrderItem
}

type ListOrdersRequest struct {
	Role   string `form:"role,optional"`
	Status string `form:"status,optional"`
	Limit  int64  `form:"limit,optional,default=20"`
	Offset int64  `form:"offset,optional,default=0"`
}

type ListOrdersResponse struct {
	List []OrderItem `json:"list"`
}

type UpdateOrderPathRequest struct {
	ID int64 `path:"id"`
}

type CancelOrderRequest struct {
	ID     int64  `path:"id"`
	Reason string `json:"reason"`
}

type DisputeOrderRequest struct {
	ID     int64  `path:"id"`
	Reason string `json:"reason"`
}

// ── backend admin ─────────────────────────────────────────────────────────────

type BackendListListingsRequest struct {
	Type   string `form:"type,optional"`
	Status string `form:"status,optional"`
	Limit  int64  `form:"limit,optional,default=20"`
	Offset int64  `form:"offset,optional,default=0"`
}

type BackendListOrdersRequest struct {
	Status string `form:"status,optional"`
	Limit  int64  `form:"limit,optional,default=20"`
	Offset int64  `form:"offset,optional,default=0"`
}

type ResolveOrderRequest struct {
	ID     int64  `path:"id"`
	Action string `json:"action"` // complete | refund
	Reason string `json:"reason"`
}

// ── v1 掛單（免登入，沿用 listings；createdBy 固定 demo_user） ──────────────────

type V1CreateOrderRequest struct {
	Type          string  `json:"type"`          // buy | sell
	Asset         string  `json:"asset"`         // v1: USDT
	Fiat          string  `json:"fiat"`          // v1: TWD
	Price         float64 `json:"price"`         // 單價
	Quantity      float64 `json:"quantity"`      // 數量
	PaymentMethod string  `json:"paymentMethod"` // bank_transfer | convenience_store
}

type V1Order struct {
	ID            string  `json:"id"`
	Type          string  `json:"type"`
	Asset         string  `json:"asset"`
	Fiat          string  `json:"fiat"`
	Price         float64 `json:"price"`
	Quantity      float64 `json:"quantity"`
	TotalAmount   float64 `json:"totalAmount"`
	PaymentMethod string  `json:"paymentMethod"`
	Status        string  `json:"status"` // open | completed | cancelled
	CreatedBy     string  `json:"createdBy"`
	CreatedAt     string  `json:"createdAt"`
	UpdatedAt     string  `json:"updatedAt"`
}

type V1OrderIDRequest struct {
	ID int64 `path:"id"`
}

type V1AdminListOrdersRequest struct {
	Status string `form:"status,optional"` // open | completed | cancelled
}

type V1OrderListResponse struct {
	List []V1Order `json:"list"`
}

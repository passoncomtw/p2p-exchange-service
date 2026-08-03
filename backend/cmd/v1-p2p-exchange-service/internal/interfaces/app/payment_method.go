package app_interface

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

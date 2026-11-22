package app_interface

type CreatePaymentMethodRequest struct {
	Type          string `json:"type" example:"bank_transfer" enums:"bank_transfer"` // 目前僅支援 bank_transfer
	BankName      string `json:"bankName" example:"台灣銀行"`
	AccountName   string `json:"accountName" example:"王小明"`
	AccountNumber string `json:"accountNumber" example:"123456789012"`
}

type CreatePaymentMethodResponse struct {
	ID int64 `json:"id" example:"1"`
}

type PaymentMethodItem struct {
	ID            int64  `json:"id" example:"1"`
	Type          string `json:"type" example:"bank_transfer" enums:"bank_transfer"`
	BankName      string `json:"bankName" example:"台灣銀行"`
	AccountName   string `json:"accountName" example:"王小明"`
	AccountNumber string `json:"accountNumber" example:"123456789012"`
	IsActive      bool   `json:"isActive" example:"true"`
}

type ListPaymentMethodsResponse struct {
	List []PaymentMethodItem `json:"list"`
}

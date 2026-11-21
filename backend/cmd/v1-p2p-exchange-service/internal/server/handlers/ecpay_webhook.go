package handlers

import (
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	ecpaywebhook_service "p2p-exchange/cmd/v1-p2p-exchange-service/internal/service/ecpay_webhook"
)

// ECPayWebhookHandler 接收 ECPay 的付款結果通知（public route，無 JWT）。
//
// 回應格式是 ECPay 通知協定規定的純文字，不是本服務標準的 JSON envelope：
// 成功回 "1|OK"，失敗回 "0|<message>"，且兩者都是 HTTP 200。
// 改成 4xx/5xx 或 JSON 會讓 ECPay 判定通知未送達而無限重送。
type ECPayWebhookHandler struct {
	svc ecpaywebhook_service.ECPayWebhookService
}

func NewECPayWebhookHandler(svc ecpaywebhook_service.ECPayWebhookService) *ECPayWebhookHandler {
	return &ECPayWebhookHandler{svc: svc}
}

// Notify 接收 ECPay 付款結果通知。
//
//	@Summary		ECPay 付款結果通知
//	@Description	由 ECPay 以 form-urlencoded 呼叫，不含 JWT。所有 form 欄位皆納入 CheckMacValue 驗簽。回應為純文字：成功 1|OK，失敗 0|錯誤訊息，HTTP 狀態碼一律 200。
//	@Tags			ecpay-webhook
//	@Accept			x-www-form-urlencoded
//	@Produce		plain
//	@Param			MerchantTradeNo	formData	string	true	"對應 fiat_deposits.merchant_trade_no"	Example(P178495360000000)
//	@Param			RtnCode			formData	string	true	"1 為付款成功，其餘視為失敗"	Example(1)
//	@Param			TradeNo			formData	string	false	"ECPay 交易編號"	Example(2611211230001234)
//	@Param			PaymentType		formData	string	false	"付款方式"	Example(Credit_CreditCard)
//	@Param			CheckMacValue	formData	string	true	"ECPay 簽章"
//	@Success		200				{string}	string	"1|OK"
//	@Router			/webhook/ecpay/notify [post]
func (h *ECPayWebhookHandler) Notify(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		logx.WithContext(r.Context()).Errorf("[ecpay-webhook] ParseForm: %v", err)
		writeECPayResponse(w, "0|parse error")
		return
	}

	// 所有 form 欄位都要納入驗簽，不能只擷取已知欄位：
	// CheckMacValue 是對整組參數計算的，ECPay 未來新增欄位時白名單會讓驗簽全面失敗。
	params := make(map[string]string, len(r.Form))
	for k, vs := range r.Form {
		if len(vs) > 0 {
			params[k] = vs[0]
		}
	}

	ok, message := h.svc.HandleNotify(r.Context(), params)
	if ok {
		writeECPayResponse(w, "1|OK")
		return
	}
	writeECPayResponse(w, "0|"+message)
}

// writeECPayResponse 以 ECPay 要求的純文字格式回應，狀態碼一律 200。
func writeECPayResponse(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(body)); err != nil {
		logx.Errorf("[ecpay-webhook] write response: %v", err)
	}
}

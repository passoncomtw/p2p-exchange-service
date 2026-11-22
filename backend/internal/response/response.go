package response

// Body 為所有 API 的統一 JSON 回傳格式。
type Body struct {
	Code    int    `json:"code" example:"0"`
	Message string `json:"message" example:"success"`
	Data    any    `json:"data" swaggertype:"object"`
}

// ErrorBody 為失敗回應的 JSON 格式（僅供 Swagger 文件使用），HTTP 狀態碼與 Code 相同。
type ErrorBody struct {
	Code    int    `json:"code" example:"400"`
	Message string `json:"message" example:"error message"`
	Data    any    `json:"data" swaggertype:"object"`
}

func Success(data any) Body {
	return Body{
		Code:    0,
		Message: "success",
		Data:    data,
	}
}

func Fail(code int, message string) Body {
	return Body{
		Code:    code,
		Message: message,
		Data:    nil,
	}
}

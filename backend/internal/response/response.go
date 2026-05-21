package response

// Body 為所有 API 的統一 JSON 回傳格式。
type Body struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
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

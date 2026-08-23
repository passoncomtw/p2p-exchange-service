package app_interface

// RegisterPushTokenRequest 更新 Expo 推播 token 的請求。
// 使用者身分一律取自 JWT context，不接受由 body 指定。
type RegisterPushTokenRequest struct {
	Token string `json:"token"`
}

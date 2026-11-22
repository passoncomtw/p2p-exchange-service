package app_interface

type LoginRequest struct {
	Username string `json:"username" example:"testdemo001"` // App 登入帳號；後台登入請用 admin001
	Password string `json:"password" example:"a12345678"`
}

// RegisterRequest 註冊請求。
// 欄位雖與 LoginRequest 相同，但語意不同（註冊的驗證規則未來可能與登入分岔），
// 因此刻意不共用同一個型別。
type RegisterRequest struct {
	Username string `json:"username" example:"newuser001"` // 3~50 字元
	Password string `json:"password" example:"a12345678"`  // 至少 8 字元
}

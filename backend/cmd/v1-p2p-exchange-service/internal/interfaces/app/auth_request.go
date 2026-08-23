package app_interface

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RegisterRequest 註冊請求。
// 欄位雖與 LoginRequest 相同，但語意不同（註冊的驗證規則未來可能與登入分岔），
// 因此刻意不共用同一個型別。
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

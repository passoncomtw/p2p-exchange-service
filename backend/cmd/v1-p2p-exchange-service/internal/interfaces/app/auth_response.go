package app_interface

type AppLoginUserInfo struct {
	ID      int64  `json:"id"`
	Account string `json:"account"`
	Name    string `json:"name"`
}

type AppLoginResponse struct {
	AccessToken string           `json:"access_token"`
	ExpireIn    int64            `json:"expireIn"`
	User        AppLoginUserInfo `json:"user"`
}

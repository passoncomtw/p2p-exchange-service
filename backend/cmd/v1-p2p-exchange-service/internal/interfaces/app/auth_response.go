package app_interface

type AppLoginUserInfo struct {
	ID      int64  `json:"id" example:"1"`
	Account string `json:"account" example:"testdemo001"`
	Name    string `json:"name" example:"testdemo001"`
}

type AppLoginResponse struct {
	AccessToken string           `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOjF9.abc123"`
	ExpireIn    int64            `json:"expireIn" example:"86400"`
	User        AppLoginUserInfo `json:"user"`
}

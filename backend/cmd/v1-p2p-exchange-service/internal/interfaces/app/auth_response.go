package app_interface

type LoginUser struct {
	ID      int64  `json:"id"`
	Account string `json:"account"`
	Name    string `json:"name"`
}

type LoginResponse struct {
	AccessToken string    `json:"access_token"`
	ExpireIn    int64     `json:"expireIn"`
	User        LoginUser `json:"user"`
}

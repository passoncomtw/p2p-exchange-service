package app_interface

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

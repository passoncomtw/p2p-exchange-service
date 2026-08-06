package backend_interface

type BackendLoginResponse struct {
	Token     string `json:"token"`
	ExpiresIn int64  `json:"expiresIn"`
}

type BackendDashboardResponse struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}

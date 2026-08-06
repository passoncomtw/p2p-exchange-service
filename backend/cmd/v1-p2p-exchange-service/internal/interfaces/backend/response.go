package backend_interface

type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresIn int64  `json:"expiresIn"`
}

type DashboardResponse struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}

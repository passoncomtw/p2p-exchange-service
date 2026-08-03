package app_interface

import (
	"net/http"
)

type LoginHandler interface {
	Handle(w http.ResponseWriter, r *http.Request)
}

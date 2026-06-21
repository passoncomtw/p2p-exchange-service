package handler

import (
	"encoding/json"
	"net/http"
)

func ctxUID(r *http.Request) int64 {
	v := r.Context().Value("uid")
	switch val := v.(type) {
	case json.Number:
		uid, _ := val.Int64()
		return uid
	case float64:
		return int64(val)
	case int64:
		return val
	default:
		return 0
	}
}

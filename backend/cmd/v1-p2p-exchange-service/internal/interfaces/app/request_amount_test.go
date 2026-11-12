package app_interface

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func parseCreateOrder(t *testing.T, body string) (CreateOrderRequest, error) {
	t.Helper()
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	var req CreateOrderRequest
	err := httpx.Parse(r, &req)
	return req, err
}

func TestCreateOrderRequestAmountIsDecimalString(t *testing.T) {
	req, err := parseCreateOrder(t, `{"listingId":1,"cryptoAmount":"1.234567890123456789"}`)
	if err != nil {
		t.Fatalf("合法字串應被接受: %v", err)
	}
	if req.CryptoAmount.String() != "1.234567890123456789" {
		t.Fatalf("got %s", req.CryptoAmount.String())
	}
}

func TestCreateOrderRequestRejectsNonStringAmount(t *testing.T) {
	for _, body := range []string{
		`{"listingId":1,"cryptoAmount":1.5}`,
		`{"listingId":1,"cryptoAmount":"1e3"}`,
		`{"listingId":1,"cryptoAmount":"0.0000000000000000001"}`,
		`{"listingId":1,"cryptoAmount":""}`,
	} {
		if _, err := parseCreateOrder(t, body); err == nil {
			t.Errorf("%s 應被拒絕", body)
		}
	}
}

package money

import (
	"encoding/json"
	"testing"

	"github.com/shopspring/decimal"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{"integer", "100", "100", nil},
		{"decimal", "0.000000000000000001", "0.000000000000000001", nil},
		{"negative", "-1.5", "-1.5", nil},
		{"max integer digits", "12345678901234567890.5", "12345678901234567890.5", nil},
		{"empty", "", "", ErrInvalidFormat},
		{"scientific", "1e3", "", ErrInvalidFormat},
		{"leading space", " 1", "", ErrInvalidFormat},
		{"trailing space", "1 ", "", ErrInvalidFormat},
		{"plus sign", "+1", "", ErrInvalidFormat},
		{"no integer part", ".5", "", ErrInvalidFormat},
		{"no fraction digits", "1.", "", ErrInvalidFormat},
		{"NaN", "NaN", "", ErrInvalidFormat},
		{"Inf", "Inf", "", ErrInvalidFormat},
		{"scale 19", "0.0000000000000000001", "", ErrScaleExceeded},
		{"integer digits 21", "123456789012345678901", "", ErrOutOfRange},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if err != tt.wantErr {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if err == nil && got.String() != tt.want {
				t.Fatalf("got %s, want %s", got.String(), tt.want)
			}
		})
	}
}

func TestRoundHalfUp(t *testing.T) {
	tests := []struct{ input, want string }{
		{"0.0000000000000000005", "0.000000000000000001"},
		{"0.0000000000000000004", "0"},
		{"-0.0000000000000000005", "-0.000000000000000001"},
		{"1.2345678901234567895", "1.234567890123456790"},
	}
	for _, tt := range tests {
		d, err := decimal.NewFromString(tt.input)
		if err != nil {
			t.Fatal(err)
		}
		want, _ := decimal.NewFromString(tt.want)
		if got := Round(d); !got.Equal(want) {
			t.Errorf("Round(%s) = %s, want %s", tt.input, got, want)
		}
	}
}

func TestIsStorable(t *testing.T) {
	storable, _ := decimal.NewFromString("123.123456789012345678")
	tooPrecise, _ := decimal.NewFromString("0.0000000000000000001")
	tooLarge, _ := decimal.NewFromString("123456789012345678901")
	if !IsStorable(storable) {
		t.Error("expected storable")
	}
	if IsStorable(tooPrecise) {
		t.Error("19 位小數不應可儲存")
	}
	if IsStorable(tooLarge) {
		t.Error("21 位整數不應可儲存")
	}
}

func TestAmountJSON(t *testing.T) {
	var ok struct {
		A Amount `json:"a"`
	}
	if err := json.Unmarshal([]byte(`{"a":"1.50"}`), &ok); err != nil {
		t.Fatal(err)
	}
	if ok.A.String() != "1.5" {
		t.Fatalf("got %s", ok.A.String())
	}
	out, _ := json.Marshal(ok)
	if string(out) != `{"a":"1.5"}` {
		t.Fatalf("marshal = %s", out)
	}
	for _, bad := range []string{`{"a":1.5}`, `{"a":null}`, `{"a":"1e3"}`, `{"a":""}`, `{"a":true}`} {
		var v struct {
			A Amount `json:"a"`
		}
		if err := json.Unmarshal([]byte(bad), &v); err == nil {
			t.Errorf("%s 應被拒絕", bad)
		}
	}
}

// Package money 提供全專案統一的金額型別與解析規則。
//
// 金額一律使用 decimal.Decimal，資料庫為 NUMERIC(38, 18)，API 以十進位字串傳遞；
// 不使用 float64 或 big.Float。
package money

import (
	"encoding/json"
	"errors"
	"regexp"

	"github.com/shopspring/decimal"
)

const (
	// Scale 金額小數位數上限，對應 NUMERIC(38, 18)。
	Scale = 18
	// maxIntegerDigits 整數位數上限，對應 NUMERIC(38, 18) 的 38 - 18。
	maxIntegerDigits = 20
)

var (
	// ErrInvalidFormat 字串不是合法的十進位金額（不接受科學記號、空白、正負號以外的符號）。
	ErrInvalidFormat = errors.New("invalid decimal format")
	// ErrScaleExceeded 小數位數超過 Scale。
	ErrScaleExceeded = errors.New("decimal scale exceeds 18")
	// ErrOutOfRange 整數位數超過 NUMERIC(38, 18) 可儲存的範圍。
	ErrOutOfRange = errors.New("decimal out of range")

	decimalPattern = regexp.MustCompile(`^-?(\d+)(\.(\d+))?$`)
)

// Parse 嚴格解析十進位字串：僅接受 -?整數(.小數)，小數位數最多 18、整數位數最多 20。
func Parse(s string) (decimal.Decimal, error) {
	matches := decimalPattern.FindStringSubmatch(s)
	if matches == nil {
		return decimal.Decimal{}, ErrInvalidFormat
	}
	if len(matches[1]) > maxIntegerDigits {
		return decimal.Decimal{}, ErrOutOfRange
	}
	if len(matches[3]) > Scale {
		return decimal.Decimal{}, ErrScaleExceeded
	}
	return decimal.NewFromString(s)
}

// Round 將運算結果取至 Scale 位小數，採四捨五入（half-up，負數為遠離零）。
func Round(d decimal.Decimal) decimal.Decimal {
	return d.Round(Scale)
}

// IsStorable 判斷數值是否能存入 NUMERIC(38, 18) 而不被資料庫捨入或溢位。
func IsStorable(d decimal.Decimal) bool {
	if !d.Equal(Round(d)) {
		return false
	}
	return len(d.Abs().Truncate(0).String()) <= maxIntegerDigits
}

// Amount 是 API request 專用的金額型別：JSON 必須是十進位字串，並套用 Parse 的嚴格規則。
type Amount struct {
	decimal.Decimal
}

// UnmarshalJSON 只接受 JSON 字串，拒絕 number、null 與科學記號。
func (a *Amount) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return ErrInvalidFormat
	}
	parsed, err := Parse(s)
	if err != nil {
		return err
	}
	a.Decimal = parsed
	return nil
}

// UnmarshalText 與 UnmarshalJSON 套用同一組嚴格規則；
// go-zero 的 httpx 優先使用 TextUnmarshaler，不覆寫的話會落到 decimal 寬鬆的解析（接受科學記號與超過 18 位的小數）。
func (a *Amount) UnmarshalText(text []byte) error {
	parsed, err := Parse(string(text))
	if err != nil {
		return err
	}
	a.Decimal = parsed
	return nil
}

// MarshalJSON 輸出去除尾端零的十進位字串。
func (a Amount) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.Decimal.String())
}

package weeklyvaluelane

import (
	"errors"
	"math/big"
	"strings"
)

const (
	maxArithmeticBits = 256
	maxDecimalDigits  = 78
	maxIdentityBytes  = 256
)

var errArithmetic = errors.New("weekly value arithmetic invalid")

func parseUnsigned(raw string) (*big.Int, bool) {
	if raw == "" || len(raw) > maxDecimalDigits || strings.HasPrefix(raw, "+") || strings.HasPrefix(raw, "-") || (len(raw) > 1 && raw[0] == '0') {
		return nil, false
	}
	for _, r := range raw {
		if r < '0' || r > '9' {
			return nil, false
		}
	}
	value, ok := new(big.Int).SetString(raw, 10)
	return value, ok && value.Sign() >= 0 && value.BitLen() <= maxArithmeticBits
}

func parseSigned(raw string) (*big.Int, bool) {
	if raw == "" || len(raw) > maxDecimalDigits+1 || strings.HasPrefix(raw, "+") || raw == "-0" {
		return nil, false
	}
	digits := strings.TrimPrefix(raw, "-")
	if digits == "" || (len(digits) > 1 && digits[0] == '0') {
		return nil, false
	}
	for _, r := range digits {
		if r < '0' || r > '9' {
			return nil, false
		}
	}
	value, ok := new(big.Int).SetString(raw, 10)
	return value, ok && value.BitLen() <= maxArithmeticBits
}

func checkedAdd(values ...*big.Int) (*big.Int, bool) {
	total := new(big.Int)
	for _, value := range values {
		if value == nil {
			return nil, false
		}
		total.Add(total, value)
		if total.BitLen() > maxArithmeticBits {
			return nil, false
		}
	}
	return total, true
}

func checkedMul(a, b *big.Int) (*big.Int, bool) {
	if a == nil || b == nil {
		return nil, false
	}
	value := new(big.Int).Mul(a, b)
	return value, value.BitLen() <= maxArithmeticBits
}

func parsePositiveDecimal(raw string) (*big.Rat, bool) {
	if raw == "" || len(raw) > 128 || strings.ContainsAny(raw, "+-/eE") || strings.Count(raw, ".") > 1 {
		return nil, false
	}
	parts := strings.Split(raw, ".")
	for _, part := range parts {
		if part == "" {
			return nil, false
		}
		for _, r := range part {
			if r < '0' || r > '9' {
				return nil, false
			}
		}
	}
	value, ok := new(big.Rat).SetString(raw)
	return value, ok && value.Sign() > 0 && value.Num().BitLen() <= maxArithmeticBits && value.Denom().BitLen() <= maxArithmeticBits
}

func ceilRat(value *big.Rat) (*big.Int, bool) {
	if value == nil || value.Sign() < 0 || value.Num().BitLen() > maxArithmeticBits || value.Denom().BitLen() > maxArithmeticBits {
		return nil, false
	}
	q, rem := new(big.Int), new(big.Int)
	q.QuoRem(value.Num(), value.Denom(), rem)
	if rem.Sign() > 0 {
		q.Add(q, big.NewInt(1))
	}
	return q, q.BitLen() <= maxArithmeticBits
}

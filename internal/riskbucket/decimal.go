package riskbucket

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

const defaultMaxDecimalBits uint = 256

var errDecimalOverflow = errors.New("decimal exceeds supported precision")

func parseDecimal(raw string, allowZero bool, maxBits uint) (*big.Rat, error) {
	if raw == "" || strings.HasPrefix(raw, "+") || strings.HasPrefix(raw, "-") || strings.Count(raw, ".") > 1 {
		return nil, fmt.Errorf("invalid exact decimal %q", raw)
	}
	parts := strings.Split(raw, ".")
	if len(parts[0]) == 0 {
		return nil, fmt.Errorf("invalid exact decimal %q", raw)
	}
	for _, part := range parts {
		if part == "" {
			return nil, fmt.Errorf("invalid exact decimal %q", raw)
		}
		for _, r := range part {
			if r < '0' || r > '9' {
				return nil, fmt.Errorf("invalid exact decimal %q", raw)
			}
		}
	}
	digits := strings.Join(parts, "")
	numerator, ok := new(big.Int).SetString(digits, 10)
	if !ok {
		return nil, fmt.Errorf("invalid exact decimal %q", raw)
	}
	denominator := big.NewInt(1)
	if len(parts) == 2 {
		denominator.Exp(big.NewInt(10), big.NewInt(int64(len(parts[1]))), nil)
	}
	if !allowZero && numerator.Sign() == 0 {
		return nil, fmt.Errorf("decimal must be positive")
	}
	if exceedsBits(maxBits, numerator, denominator) {
		return nil, errDecimalOverflow
	}
	return new(big.Rat).SetFrac(numerator, denominator), nil
}

func parseMinor(raw string, maxBits uint) (*big.Int, error) {
	if raw == "" || strings.HasPrefix(raw, "+") || strings.HasPrefix(raw, "-") || strings.Contains(raw, ".") {
		return nil, fmt.Errorf("invalid non-negative minor amount %q", raw)
	}
	for _, r := range raw {
		if r < '0' || r > '9' {
			return nil, fmt.Errorf("invalid non-negative minor amount %q", raw)
		}
	}
	value, ok := new(big.Int).SetString(raw, 10)
	if !ok || exceedsBits(maxBits, value) {
		return nil, fmt.Errorf("invalid non-negative minor amount %q", raw)
	}
	return value, nil
}

func effectiveMaxBits(maxBits uint) uint {
	if maxBits == 0 {
		return defaultMaxDecimalBits
	}
	return maxBits
}

func exceedsBits(maxBits uint, values ...*big.Int) bool {
	limit := effectiveMaxBits(maxBits)
	for _, value := range values {
		if value != nil && uint(value.BitLen()) > limit {
			return true
		}
	}
	return false
}

func checkRatBits(maxBits uint, value *big.Rat) error {
	if value == nil || exceedsBits(maxBits, value.Num(), value.Denom()) {
		return fmt.Errorf("decimal arithmetic exceeds supported precision")
	}
	return nil
}

func ceilRat(value *big.Rat, maxBits uint) (*big.Int, error) {
	if value == nil || value.Sign() < 0 {
		return nil, fmt.Errorf("cannot ceil negative or nil value")
	}
	quotient, remainder := new(big.Int), new(big.Int)
	quotient.QuoRem(value.Num(), value.Denom(), remainder)
	if remainder.Sign() > 0 {
		quotient.Add(quotient, big.NewInt(1))
	}
	if exceedsBits(maxBits, quotient) {
		return nil, fmt.Errorf("minor amount exceeds supported precision")
	}
	return quotient, nil
}

func uintString(value uint64) string { return strconv.FormatUint(value, 10) }

func compareMinor(a, b string) int {
	ai, aok := new(big.Int).SetString(a, 10)
	bi, bok := new(big.Int).SetString(b, 10)
	if !aok || !bok {
		return strings.Compare(a, b)
	}
	return ai.Cmp(bi)
}

func addMinor(a, b *big.Int, maxBits uint) (*big.Int, error) {
	result := new(big.Int).Add(a, b)
	if exceedsBits(maxBits, result) {
		return nil, fmt.Errorf("minor addition exceeds supported precision")
	}
	return result, nil
}

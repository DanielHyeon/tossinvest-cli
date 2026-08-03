package continuationlane

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

const (
	maxArithmeticBits     = 256
	maxMinorDecimalDigits = 78
)

var errOverflow = errors.New("continuation lane arithmetic exceeds 256 bits")

func parseUnsigned(raw string) (*big.Int, error) {
	if raw == "" || strings.HasPrefix(raw, "+") || strings.HasPrefix(raw, "-") {
		return nil, fmt.Errorf("invalid unsigned integer %q", raw)
	}
	if len(raw) > maxMinorDecimalDigits {
		return nil, errOverflow
	}
	for _, r := range raw {
		if r < '0' || r > '9' {
			return nil, fmt.Errorf("invalid unsigned integer %q", raw)
		}
	}
	if len(raw) > 1 && raw[0] == '0' {
		return nil, fmt.Errorf("non-canonical unsigned integer %q", raw)
	}
	value, ok := new(big.Int).SetString(raw, 10)
	if !ok {
		return nil, fmt.Errorf("invalid unsigned integer %q", raw)
	}
	if value.BitLen() > maxArithmeticBits {
		return nil, errOverflow
	}
	return value, nil
}

func parseSigned(raw string) (*big.Int, error) {
	if raw == "" || strings.HasPrefix(raw, "+") {
		return nil, fmt.Errorf("invalid signed integer %q", raw)
	}
	digits := raw
	if strings.HasPrefix(digits, "-") {
		digits = digits[1:]
	}
	if digits == "" {
		return nil, fmt.Errorf("invalid signed integer %q", raw)
	}
	if len(digits) > maxMinorDecimalDigits {
		return nil, errOverflow
	}
	for _, r := range digits {
		if r < '0' || r > '9' {
			return nil, fmt.Errorf("invalid signed integer %q", raw)
		}
	}
	if (len(digits) > 1 && digits[0] == '0') || raw == "-0" {
		return nil, fmt.Errorf("non-canonical signed integer %q", raw)
	}
	value, ok := new(big.Int).SetString(raw, 10)
	if !ok {
		return nil, fmt.Errorf("invalid signed integer %q", raw)
	}
	if value.BitLen() > maxArithmeticBits {
		return nil, errOverflow
	}
	return value, nil
}

func checkedAdd(values ...*big.Int) (*big.Int, error) {
	total := new(big.Int)
	for _, value := range values {
		if value == nil {
			return nil, fmt.Errorf("nil integer")
		}
		total.Add(total, value)
		if total.BitLen() > maxArithmeticBits {
			return nil, errOverflow
		}
	}
	return total, nil
}

func checkedMul(a, b *big.Int) (*big.Int, error) {
	if a == nil || b == nil {
		return nil, fmt.Errorf("nil integer")
	}
	value := new(big.Int).Mul(a, b)
	if value.BitLen() > maxArithmeticBits {
		return nil, errOverflow
	}
	return value, nil
}

func signedPPM(numerator, denominator string) (int64, error) {
	n, err := parseSigned(numerator)
	if err != nil {
		return 0, err
	}
	d, err := parseUnsigned(denominator)
	if err != nil || d.Sign() <= 0 {
		return 0, fmt.Errorf("denominator must be positive: %w", err)
	}
	scaled, err := checkedMul(n, big.NewInt(1_000_000))
	if err != nil {
		return 0, err
	}
	result := new(big.Int).Quo(scaled, d)
	if !result.IsInt64() {
		return 0, errOverflow
	}
	return result.Int64(), nil
}

func unsignedPPM(numerator, denominator string) (int64, error) {
	n, err := parseUnsigned(numerator)
	if err != nil {
		return 0, err
	}
	d, err := parseUnsigned(denominator)
	if err != nil || d.Sign() <= 0 {
		return 0, fmt.Errorf("denominator must be positive: %w", err)
	}
	scaled, err := checkedMul(n, big.NewInt(1_000_000))
	if err != nil {
		return 0, err
	}
	result := new(big.Int).Quo(scaled, d)
	if !result.IsInt64() {
		return 0, errOverflow
	}
	return result.Int64(), nil
}

func parsePositiveDecimal(raw string) (*big.Rat, error) {
	if raw == "" || strings.HasPrefix(raw, "+") || strings.HasPrefix(raw, "-") || strings.Count(raw, ".") > 1 {
		return nil, fmt.Errorf("invalid decimal %q", raw)
	}
	parts := strings.Split(raw, ".")
	for _, part := range parts {
		if part == "" {
			return nil, fmt.Errorf("invalid decimal %q", raw)
		}
		for _, r := range part {
			if r < '0' || r > '9' {
				return nil, fmt.Errorf("invalid decimal %q", raw)
			}
		}
	}
	digits := strings.Join(parts, "")
	if len(digits) > maxMinorDecimalDigits {
		return nil, errOverflow
	}
	numerator, ok := new(big.Int).SetString(digits, 10)
	if !ok || numerator.Sign() <= 0 || numerator.BitLen() > maxArithmeticBits {
		return nil, fmt.Errorf("decimal must be positive")
	}
	denominator := big.NewInt(1)
	if len(parts) == 2 {
		denominator.Exp(big.NewInt(10), big.NewInt(int64(len(parts[1]))), nil)
	}
	if denominator.BitLen() > maxArithmeticBits {
		return nil, errOverflow
	}
	return new(big.Rat).SetFrac(numerator, denominator), nil
}

func ceilRat(value *big.Rat) (*big.Int, error) {
	if value == nil || value.Sign() < 0 || value.Num().BitLen() > maxArithmeticBits || value.Denom().BitLen() > maxArithmeticBits {
		return nil, errOverflow
	}
	q, rem := new(big.Int), new(big.Int)
	q.QuoRem(value.Num(), value.Denom(), rem)
	if rem.Sign() > 0 {
		q.Add(q, big.NewInt(1))
	}
	if q.BitLen() > maxArithmeticBits {
		return nil, errOverflow
	}
	return q, nil
}

func canonicalUint(value uint64) string { return strconv.FormatUint(value, 10) }

func codeForArithmetic(err error) RefusalCode {
	if errors.Is(err, errOverflow) {
		return RefusalArithmeticOverflow
	}
	return RefusalArithmeticInvalid
}

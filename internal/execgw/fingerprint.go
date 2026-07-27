package execgw

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// FingerprintInput is the identity of a mutation as the broker could later show it
// back to us.
//
// The order-execution spec names the matching key for IN_DOUBT resolution as
// "계좌·심볼·방향·수량·가격·제출 시각 창". The first five live here; the submission
// time window deliberately does not, because it is not stable: the window is
// derived at resolution time from the attempt's recorded dispatch timestamp, and
// baking it into the hash would mean a fingerprint that can never be recomputed
// from a broker order.
//
// The trading day *is* included, which is the coarse-grained half of the time
// dimension and is stable — an order from yesterday is a different order.
type FingerprintInput struct {
	AccountRef string
	Market     string
	Symbol     string
	Side       string // "BUY" | "SELL"
	Quantity   float64
	Price      float64
	Currency   string
	TradingDay string // market-local YYYY-MM-DD
}

// Fingerprint hashes the identity fields into the value stored on the intent and
// the attempt.
//
// It is a hash rather than the raw tuple because it lands in an indexed column and
// in operator-visible logs; the resolution procedure matches on the *fields*
// (which it reads back from the immutable intent row), not on this string. What
// the hash gives us is a cheap equality check and a stable id to talk about.
func Fingerprint(in FingerprintInput) string {
	canonical := strings.Join([]string{
		"v1",
		strings.TrimSpace(in.AccountRef),
		strings.ToLower(strings.TrimSpace(in.Market)),
		strings.ToUpper(strings.TrimSpace(in.Symbol)),
		strings.ToUpper(strings.TrimSpace(in.Side)),
		decimalString(in.Quantity),
		decimalString(in.Price),
		strings.ToUpper(strings.TrimSpace(in.Currency)),
		strings.TrimSpace(in.TradingDay),
	}, "|")
	sum := sha256.Sum256([]byte(canonical))
	return hex.EncodeToString(sum[:])
}

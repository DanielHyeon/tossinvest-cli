package costs

// fingerprint.go is change add-net-rr-measurement task 4.4.
//
// # Why a measurement needs the model's identity stamped on it
//
// Every rate in DefaultModel is tagged `[미검증 — 2b 실측 대체 대상]`: a deliberate
// over-estimate standing in until somebody measures the real ones. Any net ratio
// computed today is therefore a statement about the placeholders, not about what
// the account actually pays.
//
// That is fine as long as it stays visible. What is not fine is aggregating
// observations taken before the measurement with observations taken after it — the
// mean of two different cost models is a number describing nothing. So each
// observation carries the fingerprint of the rate set it was computed under, and
// mixing them becomes a query somebody has to write deliberately rather than the
// default.
//
// # Why the digest and not a version string
//
// A hand-maintained version would need updating whenever a rate changed, and the
// failure mode of forgetting is silent: two different rate sets sharing one label,
// which is precisely the confusion the field exists to prevent. Deriving it from
// the rates makes "the rates changed and the fingerprint did not" unrepresentable.

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
)

// FingerprintUnconfigured is the fingerprint of the zero Model.
//
// It is a named sentinel rather than a digest because the zero value is not a rate
// set anybody configured — it is the absence of one, and every consumer that
// judges anything on cost refuses it. An observation carrying this has no cost
// basis at all, which is a different fact from "computed under rates that happened
// to be zero" (a configured all-zero model gets a real digest).
const FingerprintUnconfigured = "costs/unconfigured"

// Fingerprint identifies the rate set, deterministically.
//
// Derived from every key in OverrideKeys() paired with its rate, in the registry's
// stable order, hashed and truncated. Truncated because this is stored on every
// observation row and 16 hex characters is far more collision resistance than a
// handful of rate sets needs; the full digest would be storage spent on nothing.
//
// The registry is the single source of truth for which rates exist, so a rate
// added to Model without being wired into overrideKeys would be missing here — and
// costs_test.go's sweep over that same registry is what catches it, which is the
// arrangement that keeps this honest rather than merely tidy.
func (m Model) Fingerprint() string {
	if !m.configured {
		return FingerprintUnconfigured
	}
	var b strings.Builder
	b.WriteString("costs/v1")
	for _, key := range overrideKeys {
		b.WriteByte('\n')
		b.WriteString(key)
		b.WriteByte('=')
		// 'g' with −1 precision is the shortest representation that round-trips,
		// so two models with equal rates always produce identical text and two
		// with different rates never do.
		b.WriteString(strconv.FormatFloat(m.Rate(key), 'g', -1, 64))
	}
	sum := sha256.Sum256([]byte(b.String()))
	return "costs/" + hex.EncodeToString(sum[:8])
}

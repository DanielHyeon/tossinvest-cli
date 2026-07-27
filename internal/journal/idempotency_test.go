package journal_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// The broker's own constraint, transcribed from
// docs/migration/openapi.latest.json (OrderCreateRequest.clientOrderId):
// "최대 36자, 영숫자 및 `-`, `_` 허용", pattern ^[a-zA-Z0-9\-_]+$.
var openapiKeyPattern = regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`)

// TestDeriveClientOrderIDIsDeterministic is the property the whole replay path
// rests on: the key is a pure function of values the issuer owns, so the same
// decision always addresses the same broker order (design D2).
func TestDeriveClientOrderIDIsDeterministic(t *testing.T) {
	const id = "d-7f3c9a12"
	first := journal.DeriveClientOrderID(id, 0)
	for i := 0; i < 100; i++ {
		if got := journal.DeriveClientOrderID(id, 0); got != first {
			t.Fatalf("derivation is not stable: %q then %q", first, got)
		}
	}
}

// TestDeriveClientOrderIDGoldenVector pins the algorithm itself. Changing it
// silently would make a replay send a *new* key for an attempt that is already
// in flight — a duplicate order rather than a recovered one — so the constant
// below is a contract, not a fixture.
func TestDeriveClientOrderIDGoldenVector(t *testing.T) {
	cases := []struct {
		id         string
		generation int
		want       string
	}{
		{"d-7f3c9a12", 0, "tos-wxBCiXOpDqyffCRmqHv-IDOBUS2wlr97"},
		{"d-7f3c9a12", 1, "tos-kRc2obADo1bwggvxowF2XraaGu8dYKFX"},
		{"", 0, "tos-BdjATK9xnUrqJ-v7bO_lAWV1Qh3oWPUQ"},
	}
	for _, tc := range cases {
		if got := journal.DeriveClientOrderID(tc.id, tc.generation); got != tc.want {
			t.Errorf("DeriveClientOrderID(%q, %d) = %q, want %q", tc.id, tc.generation, got, tc.want)
		}
	}
}

// TestDeriveClientOrderIDSatisfiesTheBrokerPattern checks every key this build
// can produce against the documented constraint. A key the broker rejects is a
// place that cannot be replayed.
func TestDeriveClientOrderIDSatisfiesTheBrokerPattern(t *testing.T) {
	for i := 0; i < 2000; i++ {
		key := journal.DeriveClientOrderID(fmt.Sprintf("decision-%d-%s", i, "ÄÖ 空白/포함"), i%7)
		if len(key) > 36 {
			t.Fatalf("key %q is %d chars, the broker allows 36", key, len(key))
		}
		if !openapiKeyPattern.MatchString(key) {
			t.Fatalf("key %q is outside ^[a-zA-Z0-9\\-_]+$", key)
		}
	}
}

// TestDeriveClientOrderIDSeparatesDecisionsAndGenerations: the key must
// distinguish both inputs, or a reissue would land on the broker's cached result
// for the previous generation instead of placing the new order.
func TestDeriveClientOrderIDSeparatesDecisionsAndGenerations(t *testing.T) {
	seen := make(map[string]string)
	for i := 0; i < 5000; i++ {
		id := fmt.Sprintf("d-%d", i)
		for gen := 0; gen < 3; gen++ {
			key := journal.DeriveClientOrderID(id, gen)
			label := fmt.Sprintf("%s#%d", id, gen)
			if prev, dup := seen[key]; dup {
				t.Fatalf("key collision: %s and %s both derive %q", prev, label, key)
			}
			seen[key] = label
		}
	}
}

// TestDeriveClientOrderIDIsNotAConcatenation guards the one shortcut that would
// look simplest and be wrong: an id that already contains the delimiter must not
// be able to impersonate another (id, generation) pair.
func TestDeriveClientOrderIDIsNotAConcatenation(t *testing.T) {
	if journal.DeriveClientOrderID("a", 12) == journal.DeriveClientOrderID("a1", 2) {
		t.Fatal("the derivation is ambiguous across its two inputs")
	}
}

// TestValidClientOrderID is the guard used wherever a key crosses a boundary
// (Prepare, decision recording, gateway verification).
func TestValidClientOrderID(t *testing.T) {
	valid := []string{"a", "tos-abc_DEF-123", journal.DeriveClientOrderID("d", 0),
		"012345678901234567890123456789012345"}
	for _, k := range valid {
		if !journal.ValidClientOrderID(k) {
			t.Errorf("ValidClientOrderID(%q) = false, want true", k)
		}
	}
	invalid := []string{"", " ", "has space", "slash/", "0123456789012345678901234567890123456", "키"}
	for _, k := range invalid {
		if journal.ValidClientOrderID(k) {
			t.Errorf("ValidClientOrderID(%q) = true, want false", k)
		}
	}
}

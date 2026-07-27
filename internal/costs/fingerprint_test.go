package costs

import (
	"strconv"
	"strings"
	"testing"
)

// fingerprint_test.go is change add-net-rr-measurement task 4.4: the fingerprint
// is derived from the rates and cannot fail to move when they do.

// TestFingerprintIsDeterministic: the same rates always produce the same string,
// or an observation could not be grouped by the model that produced it.
func TestFingerprintIsDeterministic(t *testing.T) {
	first := DefaultModel().Fingerprint()
	second := DefaultModel().Fingerprint()
	if first != second {
		t.Fatalf("two identical models fingerprinted differently: %s and %s", first, second)
	}
	if !strings.HasPrefix(first, "costs/") {
		t.Errorf("fingerprint %q should be namespaced so a bare digest is not mistaken "+
			"for some other identifier", first)
	}

	// Rebuilt through NewModel with no overrides — same rates, so same identity.
	rebuilt, err := NewModel(nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := rebuilt.Fingerprint(); got != first {
		t.Errorf("NewModel(nil) = %s, want DefaultModel's %s", got, first)
	}
}

// TestEveryRateMovesTheFingerprint is the property the whole field rests on. It
// sweeps the override registry rather than listing rates, so a rate added later is
// covered without anybody remembering to add it here.
func TestEveryRateMovesTheFingerprint(t *testing.T) {
	base := DefaultModel()
	baseline := base.Fingerprint()

	seen := map[string]string{baseline: "the default model"}
	for _, key := range OverrideKeys() {
		t.Run(key, func(t *testing.T) {
			// A rate that differs from the default by a hair. If a hair does not
			// move the fingerprint, neither does a rounding difference between two
			// deployments, and two rate sets would share an identity.
			nudged := formatRateForTest(base.Rate(key) + 0.000001)
			model, err := NewModel(map[string]string{key: nudged})
			if err != nil {
				t.Fatalf("NewModel(%s=%s): %v", key, nudged, err)
			}
			got := model.Fingerprint()
			if got == baseline {
				t.Fatalf("changing %s did not move the fingerprint; observations computed "+
					"under two different rate sets would aggregate as one", key)
			}
			if other, clash := seen[got]; clash {
				t.Fatalf("changing %s produced the same fingerprint as %s", key, other)
			}
			seen[got] = "the model with " + key + " nudged"
		})
	}
}

// TestUnconfiguredIsNamedNotDigested: the zero model has no rate set, and saying
// so is different from reporting a digest of seven zeros. A configured all-zero
// model is a decision somebody made and gets a real fingerprint.
func TestUnconfiguredIsNamedNotDigested(t *testing.T) {
	var absent Model
	if got := absent.Fingerprint(); got != FingerprintUnconfigured {
		t.Errorf("the zero model fingerprints as %q, want %q", got, FingerprintUnconfigured)
	}

	zeroed := map[string]string{}
	for _, key := range OverrideKeys() {
		zeroed[key] = "0"
	}
	configured, err := NewModel(zeroed)
	if err != nil {
		t.Fatal(err)
	}
	got := configured.Fingerprint()
	if got == FingerprintUnconfigured {
		t.Error("a model an operator configured to charge nothing is not an absent model; " +
			"the two must be distinguishable in the observation record")
	}
	if got == DefaultModel().Fingerprint() {
		t.Error("an all-zero model must not share the default's fingerprint")
	}
}

// formatRateForTest renders a rate the way an override map would carry it: plain
// decimal, no exponent, since parseRate refuses scientific notation.
func formatRateForTest(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

package riskbucket

import (
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/officialfx"
)

func TestBindFXAuthorityRejectsPublicEvidenceWithoutOpaqueCapability(t *testing.T) {
	now := testNow
	policy := identityPolicy()
	policy.EvaluatedAt = now
	policy.FX = FXEvidence{
		RateQuoteToBase: "1", Haircut: "1",
		Evidence: Evidence{Source: officialfx.IdentitySource, Version: officialfx.IdentityVersion, Digest: "caller",
			Official: true, Frozen: true, ObservedAt: now.Add(-time.Minute), FreshUntil: now.Add(time.Minute)},
	}
	if _, err := BindFXAuthority(policy, officialfx.Evidence{}, now); !IsRefusal(err, RefusalCurrencyUnresolved) {
		t.Fatalf("caller-created FX evidence accepted without opaque authority: %v", err)
	}
}

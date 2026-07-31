package exitpolicy_test

import (
	"errors"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
)

func TestApprovedPolicyTablesMatchPinnedSemanticDigests(t *testing.T) {
	t.Parallel()

	want := map[string]string{
		"default_v1":                    "sha256:e320c21c67852d98aad3ee7b9daa7afa7d5fd619f91f9dfaed93f5857744ea25",
		exitpolicy.CommonLadderBalanced: "sha256:81efea35eb31f02ef9736ceac920a4b895af4bcfad96b0884e41ad4190585a38",
		exitpolicy.CommonLadderRunner:   "sha256:4ff1db694f777c3aa58310566593f26edb537fcfc380115c5ce6ee3da06bce1a",
		exitpolicy.CommonLadderHybrid50: "sha256:a4e2df8f7971abfdf00beb8be840bd0370bccd5d6ef1ba56a757340013514e4a",
		"COMMON_LADDER_RUNNER@adopted":  "sha256:97ae29e33c5c9530b43ecc2c2a830defc16a805904f5e28d7a96960623f9dd3f",
		exitpolicy.RatchetPolicyID:      "sha256:e466b80c94435f8b737eeaf8458d2ffb70ec82d6407f0e6ed93c8dafa2a3f032",
	}

	defaultIdentity, err := exitpolicy.DefaultLadderPolicy().Identity()
	if err != nil {
		t.Fatal(err)
	}
	if defaultIdentity.Digest != want["default_v1"] {
		t.Fatalf("default digest = %s", defaultIdentity.Digest)
	}
	for _, policy := range exitpolicy.RegisteredCommonPolicies() {
		identity, err := policy.Ladder.Identity()
		if err != nil {
			t.Fatalf("%s identity: %v", policy.ID, err)
		}
		if identity.Digest != want[policy.ID] {
			t.Fatalf("%s digest = %s, want pinned %s", policy.ID, identity.Digest, want[policy.ID])
		}
	}
	adopted, err := exitpolicy.CommonLadderForPosition(exitpolicy.CommonLadderRunner, true)
	if err != nil {
		t.Fatal(err)
	}
	adoptedIdentity, err := adopted.Identity()
	if err != nil {
		t.Fatal(err)
	}
	if adoptedIdentity.Digest != want["COMMON_LADDER_RUNNER@adopted"] {
		t.Fatalf("adopted RUNNER digest = %s", adoptedIdentity.Digest)
	}
	ratchet, err := exitpolicy.RatchetPolicyIdentity(exitpolicy.DefaultRatchetConfig())
	if err != nil {
		t.Fatal(err)
	}
	if ratchet.Digest != want[exitpolicy.RatchetPolicyID] {
		t.Fatalf("ratchet digest = %s", ratchet.Digest)
	}
}

func TestPinnedPolicyDigestRejectsTableChangeWithoutIdentityBump(t *testing.T) {
	t.Parallel()

	profile, ok := exitpolicy.CommonPolicyByID(exitpolicy.CommonLadderBalanced)
	if !ok {
		t.Fatal("balanced common policy is not registered")
	}
	policy := profile.Ladder
	policy.Rungs[0].TargetPct = "1.4"
	if _, err := policy.Identity(); !errors.Is(err, exitpolicy.ErrPolicyIdentityConflict) {
		t.Fatalf("mutated table error = %v, want identity conflict", err)
	}
}

func TestLegacyPolicyIdentityIsExactAndUnknownIsFailClosed(t *testing.T) {
	t.Parallel()

	identity, err := exitpolicy.LegacyLadderPolicyIdentity("default_v1", false)
	if err != nil {
		t.Fatal(err)
	}
	if identity.Digest != "sha256:e320c21c67852d98aad3ee7b9daa7afa7d5fd619f91f9dfaed93f5857744ea25" {
		t.Fatalf("legacy identity = %+v", identity)
	}
	if _, err := exitpolicy.LegacyLadderPolicyIdentity("FUTURE_POLICY", false); !errors.Is(err, exitpolicy.ErrPolicyIdentityConflict) {
		t.Fatalf("unknown legacy identity error = %v", err)
	}
}

package soak_test

// supervised_test.go covers the second evidence source.
//
// The soak proves reads by running unattended for days and contains no mutation
// transport, so it can never prove that placing or cancelling an order works —
// and those two calls are in the gate's required set. They come from the
// supervised live check, where a person approved every request against a real
// account.
//
// Every test here is about the ways that second source could claim more than it
// proved. Widening any of them widens what a live-trading gate will start on.

import (
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
	"github.com/JungHoonGhae/tossinvest-cli/internal/soak"
)

// supervisedProofs is a clean pair of mutation proofs for the fixture account,
// stamped just before the attestation is issued.
func supervisedProofs(at time.Time) []attest.Proof {
	var out []attest.Proof
	for _, e := range soak.LiveOnlyEndpoints() {
		out = append(out, attest.Proof{
			Endpoint:   e,
			At:         at,
			AccountRef: attest.Mask("123-45-678901"),
			Source:     "capability-verify.jsonl",
			Market:     "KR",
		})
	}
	return out
}

func TestSupervisedEvidenceCompletesTheEnginesRequiredSet(t *testing.T) {
	s := soak.Summarize(threeCleanDays())
	now := soakStart.AddDate(0, 0, 3)

	a, err := soak.BuildAttestation(s, criteria(), now, "tester", "", supervisedProofs(now.Add(-time.Hour)))
	if err != nil {
		t.Fatalf("BuildAttestation: %v", err)
	}

	for _, want := range soak.LiveOnlyEndpoints() {
		if missing := a.MissingEndpoints([]string{want}); len(missing) != 0 {
			t.Errorf("%s was proven by the supervised check but is not in the attestation", want)
		}
	}
	if len(a.SupervisedBy) != len(soak.LiveOnlyEndpoints()) {
		t.Errorf("SupervisedBy = %+v, want one entry per mutation endpoint so an auditor can see what proved it",
			a.SupervisedBy)
	}
}

// TestSupervisedEvidenceCannotStandInForTheSoak is the reason the borrowed set is
// closed: one supervised success is not what days of unattended operation prove,
// and letting a read in this way would launder a soak failure.
func TestSupervisedEvidenceCannotStandInForTheSoak(t *testing.T) {
	now := soakStart.AddDate(0, 0, 3)
	proofs := append(supervisedProofs(now.Add(-time.Hour)), attest.Proof{
		Endpoint:   "GET /api/v1/holdings",
		At:         now.Add(-time.Hour),
		AccountRef: attest.Mask("123-45-678901"),
	})

	_, err := soak.BuildAttestation(soak.Summarize(threeCleanDays()), criteria(), now, "tester", "", proofs)
	if err == nil {
		t.Fatal("a read endpoint was accepted from the supervised record; the soak is what proves reads")
	}
	if !strings.Contains(err.Error(), "GET /api/v1/holdings") {
		t.Errorf("err = %v, want it to name the endpoint it refused", err)
	}
}

// TestSupervisedEvidenceIsClosedToEndpointsTheEngineDoesNotRequire. The
// verification record also exercises conditional orders; those are not in the
// gate's set and attesting them would be claiming something nobody asked for.
func TestSupervisedEvidenceIsClosedToOtherMutations(t *testing.T) {
	now := soakStart.AddDate(0, 0, 3)
	proofs := append(supervisedProofs(now.Add(-time.Hour)), attest.Proof{
		Endpoint:   "POST /api/v1/conditional-orders",
		At:         now.Add(-time.Hour),
		AccountRef: attest.Mask("123-45-678901"),
	})

	_, err := soak.BuildAttestation(soak.Summarize(threeCleanDays()), criteria(), now, "tester", "", proofs)
	if err == nil {
		t.Fatal("a conditional-order endpoint was attested; it is not in the set the gate requires")
	}
}

// TestSupervisedEvidenceFromAnotherAccountRefusesTheIssue — refuse, not skip. A
// record for a different account sitting on the expected path is a
// misconfiguration, and skipping it makes that indistinguishable from having no
// evidence at all.
func TestSupervisedEvidenceFromAnotherAccountRefusesTheIssue(t *testing.T) {
	now := soakStart.AddDate(0, 0, 3)
	proofs := supervisedProofs(now.Add(-time.Hour))
	proofs[0].AccountRef = attest.Mask("999-99-999999")

	_, err := soak.BuildAttestation(soak.Summarize(threeCleanDays()), criteria(), now, "tester", "", proofs)
	if err == nil {
		t.Fatal("evidence from another account was folded into this account's attestation")
	}
}

// TestSupervisedEvidenceOlderThanTheValidityIsNotEvidence. Skipped rather than
// refused: stale evidence is an ordinary state (the check was run months ago),
// and the gate's own MissingEndpoints reports it.
func TestSupervisedEvidenceOlderThanTheValidityIsNotEvidence(t *testing.T) {
	now := soakStart.AddDate(0, 0, 3)
	stale := now.Add(-criteria().Validity - time.Hour)

	a, err := soak.BuildAttestation(soak.Summarize(threeCleanDays()), criteria(), now, "tester", "",
		supervisedProofs(stale))
	if err != nil {
		t.Fatalf("BuildAttestation: %v", err)
	}
	if missing := a.MissingEndpoints(soak.LiveOnlyEndpoints()); len(missing) != len(soak.LiveOnlyEndpoints()) {
		t.Fatalf("stale evidence was attested: missing = %v", missing)
	}
	if len(a.SupervisedBy) != 0 {
		t.Errorf("SupervisedBy = %+v on stale evidence", a.SupervisedBy)
	}
}

// TestSupervisedEvidenceFromTheFutureIsNotEvidence.
func TestSupervisedEvidenceFromTheFutureIsNotEvidence(t *testing.T) {
	now := soakStart.AddDate(0, 0, 3)

	a, err := soak.BuildAttestation(soak.Summarize(threeCleanDays()), criteria(), now, "tester", "",
		supervisedProofs(now.Add(time.Hour)))
	if err != nil {
		t.Fatalf("BuildAttestation: %v", err)
	}
	if len(a.SupervisedBy) != 0 {
		t.Errorf("evidence stamped after the issue instant was accepted: %+v", a.SupervisedBy)
	}
}

// TestWithoutSupervisedEvidenceTheGateStillRefuses is the regression guard on the
// existing behaviour: no supervised check yet means an attestation covering reads
// only, and the interlock refusing on the two it does not carry.
func TestWithoutSupervisedEvidenceTheGateStillRefuses(t *testing.T) {
	now := soakStart.AddDate(0, 0, 3)

	a, err := soak.BuildAttestation(soak.Summarize(threeCleanDays()), criteria(), now, "tester", "", nil)
	if err != nil {
		t.Fatalf("BuildAttestation: %v", err)
	}
	missing := a.MissingEndpoints(soak.LiveOnlyEndpoints())
	if len(missing) != len(soak.LiveOnlyEndpoints()) {
		t.Fatalf("missing = %v, want both mutation endpoints still unproven", missing)
	}
	if len(a.SupervisedBy) != 0 {
		t.Errorf("SupervisedBy = %+v with no supervised evidence", a.SupervisedBy)
	}
}

// TestTheNoteSaysWhichMutationsAreStillUnproven — the operator reads this line
// before deciding whether the gate can be turned on.
func TestTheNoteSaysWhichMutationsAreStillUnproven(t *testing.T) {
	now := soakStart.AddDate(0, 0, 3)

	covered, err := soak.BuildAttestation(soak.Summarize(threeCleanDays()), criteria(), now, "tester", "",
		supervisedProofs(now.Add(-time.Hour)))
	if err != nil {
		t.Fatalf("BuildAttestation: %v", err)
	}
	if strings.Contains(covered.Notes, "NOT covered") {
		t.Errorf("the note still says the mutations are uncovered after they were proven:\n%s", covered.Notes)
	}

	bare, err := soak.BuildAttestation(soak.Summarize(threeCleanDays()), criteria(), now, "tester", "", nil)
	if err != nil {
		t.Fatalf("BuildAttestation: %v", err)
	}
	if !strings.Contains(bare.Notes, "NOT covered") {
		t.Errorf("the note does not say the mutations are uncovered:\n%s", bare.Notes)
	}
}

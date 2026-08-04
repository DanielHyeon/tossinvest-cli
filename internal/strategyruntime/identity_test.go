package strategyruntime

import (
	"strings"
	"testing"
	"time"
)

func TestIdentityConstructorsRejectWhitespaceControlAndOversize(t *testing.T) {
	fixture := newLeaseFixture(t)

	lineage := lineageInput(fixture.lineage)
	lineage.seal = [32]byte{}
	lineage.CandidateID = "candidate with space"
	if _, err := newLineage(lineage); err == nil {
		t.Fatal("lineage accepted whitespace identity")
	}

	authority := fixture.authorityInput
	authority.AccountID = "acct\n2"
	if _, err := newAuthoritySnapshot(authority); err == nil {
		t.Fatal("authority accepted control identity")
	}
	if _, err := newOwnerFence(2, strings.Repeat("x", 257)); err == nil {
		t.Fatal("owner accepted oversized identity")
	}
	if _, err := newLease(leaseInput{ID: " lease", OperationID: "operation", Lineage: fixture.lineage, Authority: fixture.authority,
		Owner: fixture.owner, IssuedAt: runtimeNow, ExpiresAt: runtimeNow.Add(time.Minute)}); err == nil {
		t.Fatal("lease accepted noncanonical identity")
	}
	workerInput := workerStateInput{Market: MarketKR, Desired: EntryOn, Effective: EntryOn, Runtime: RuntimeObserved,
		CalendarGeneration: 1, CalendarDigest: digest("calendar"), ActivationGeneration: 1, ActivationDigest: digest("activation"),
		EvidenceCursor: "cursor\tbad", EvidenceDigest: digest("evidence"), BudgetKey: "quotes:KR"}
	if _, err := newWorkerState(workerInput); err == nil {
		t.Fatal("worker accepted control identity")
	}
	outcome := validOutcomeInput("operation 1", OutcomeAccepted, "order-1", runtimeNow.Add(time.Second), nil)
	if _, err := newOutcomeEvidence(outcome); err == nil {
		t.Fatal("outcome accepted whitespace operation identity")
	}
}

func TestPostMintSealRecalculationCannotMakeNoncanonicalIdentityValid(t *testing.T) {
	fixture := newLeaseFixture(t)

	lineage := fixture.lineage
	lineage.AttemptID = "attempt\n2"
	lineage.seal = lineageSeal(lineage)
	if validLineage(lineage) {
		t.Fatal("re-sealed noncanonical lineage became valid")
	}

	authority := fixture.authority
	authority.symbol = " 005930"
	authority.seal = authoritySeal(authority)
	if validAuthority(authority) {
		t.Fatal("re-sealed noncanonical authority became valid")
	}

	lease := fixture.lease
	lease.id = "lease\x00two"
	lease.seal = leaseSeal(lease)
	if validLease(lease) {
		t.Fatal("re-sealed noncanonical lease became valid")
	}

	worker := readyCoordinator(t).Worker(MarketKR)
	worker.BudgetKey = "budget key"
	worker.seal = workerSeal(worker)
	if validWorker(worker) {
		t.Fatal("re-sealed noncanonical worker became valid")
	}

	evidence := mustOutcome(t, fixture.lease.OperationID(), OutcomeAccepted, "order-1")
	evidence.outcomeCode = "accepted response"
	evidence.seal = outcomeEvidenceSeal(evidence)
	if validOutcomeEvidence(evidence, fixture.lease.OperationID()) {
		t.Fatal("re-sealed noncanonical outcome became valid")
	}
}

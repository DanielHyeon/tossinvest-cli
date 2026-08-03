package strategyruntime

import (
	"crypto/sha256"
	"testing"
	"time"
)

type leaseFixture struct {
	lease          Lease
	lineage        Lineage
	authority      authoritySnapshot
	authorityInput authoritySnapshotInput
	owner          ownerFence
}

func newLeaseFixture(t *testing.T) leaseFixture {
	t.Helper()
	lineage, err := newLineage(lineageInput{CandidateID: "candidate-1", CandidateDigest: digest("candidate"), EvidenceDigest: digest("evidence"),
		RouterDecisionID: "router-1", LaneID: "kr-continuation", LaneVersion: "v1", CampaignID: "campaign-1", LegID: "leg-1",
		RiskPolicyDigest: digest("risk-policy"), ReservationID: "reservation-1", GuardianDecisionID: "guardian-1", AttemptID: "attempt-1"})
	if err != nil {
		t.Fatal(err)
	}
	authorityInput := authoritySnapshotInput{AccountID: "acct", Market: MarketKR, Symbol: "005930",
		Activation: generationBinding{Generation: 1, Digest: digest("activation-A")}, Calendar: generationBinding{Generation: 1, Digest: digest("calendar")},
		Protection:     protectionBinding{Generation: 1, AttestationSerial: 10, Digest: digest("protection"), State: ProtectionWired},
		Reconciliation: generationBinding{Generation: 1, Digest: digest("reconciliation")}, Risk: generationBinding{Generation: 1, Digest: digest("risk")},
		Guardian: generationBinding{Generation: 1, Digest: digest("guardian-A")}, Build: generationBinding{Generation: 1, Digest: digest("build")}}
	authority, err := newAuthoritySnapshot(authorityInput)
	if err != nil {
		t.Fatal(err)
	}
	owner, err := newOwnerFence(1, "owner-token-1")
	if err != nil {
		t.Fatal(err)
	}
	lease, err := newLease(leaseInput{ID: "lease-1", OperationID: "operation-1", Lineage: lineage, Authority: authority, Owner: owner,
		IssuedAt: runtimeNow, ExpiresAt: runtimeNow.Add(time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	return leaseFixture{lease: lease, lineage: lineage, authority: authority, authorityInput: authorityInput, owner: owner}
}

func mustOutcome(t *testing.T, operationID string, outcome BrokerOutcome, brokerOrderID string) outcomeEvidence {
	t.Helper()
	if outcome != OutcomeAccepted {
		brokerOrderID = ""
	}
	evidence, err := newOutcomeEvidence(validOutcomeInput(operationID, outcome, brokerOrderID, runtimeNow.Add(5*time.Second), nil))
	if err != nil {
		t.Fatal(err)
	}
	return evidence
}

func validOutcomeInput(operationID string, outcome BrokerOutcome, brokerOrderID string, observedAt time.Time, edit func(*outcomeEvidenceInput)) outcomeEvidenceInput {
	input := outcomeEvidenceInput{Outcome: outcome, OperationID: operationID, BrokerOrderID: brokerOrderID,
		OutcomeCode: string(outcome), QueryDigest: digest("query-" + string(outcome)), LookupDigest: digest("lookup-" + string(outcome)),
		ResponseDigest: digest("response-" + string(outcome)), ObservedAt: observedAt}
	switch outcome {
	case OutcomeAccepted:
		input.Authoritative, input.LookupComplete, input.AcceptanceKnown, input.Accepted, input.TerminalOrderFound = true, true, true, true, true
	case OutcomeDefinitiveRejected:
		input.Authoritative, input.LookupComplete, input.AcceptanceKnown, input.DefinitiveRejection = true, true, true, true
	case OutcomeAuthoritativeNotSubmitted:
		input.Authoritative, input.LookupComplete, input.AcceptanceKnown = true, true, true
	case OutcomeTransportUncertain:
	}
	if edit != nil {
		edit(&input)
	}
	return input
}

func mustRetryReservation(t *testing.T, id string, disposition ReservationDisposition) retryReservation {
	t.Helper()
	reservation, err := newRetryReservation(id, disposition)
	if err != nil {
		t.Fatal(err)
	}
	return reservation
}

func mustRecoveryCapability(t *testing.T, fixture leaseFixture, generation, maxAttempts uint64) recoveryCapability {
	t.Helper()
	capability, err := newRecoveryCapability(recoveryCapabilityInput{Market: fixture.authority.Market(), OperationID: fixture.lease.OperationID(),
		Generation: generation, AttestationSerial: fixture.authority.protection.AttestationSerial, Digest: fixture.authority.protection.Digest, ExpiresAt: runtimeNow.Add(time.Hour), ClientKeyForwarded: true, ClientKeyEchoed: true,
		ExactLookup: true, UniqueIdentity: true, PendingQuery: true, TerminalQuery: true, CancelResultQuery: true, Dedup: true, IdempotentSameKey: true, MaximumAttempts: maxAttempts})
	if err != nil {
		t.Fatal(err)
	}
	return capability
}

func mustCentralFault(t *testing.T, kind CentralFaultKind, detectedAt time.Time, owner ownerFence) centralFault {
	t.Helper()
	fault, err := newCentralFault(centralFaultInput{Kind: kind, DetectedAt: detectedAt, EvidenceDigest: digest("central-fault"), CurrentOwner: owner})
	if err != nil {
		t.Fatal(err)
	}
	return fault
}

func digest(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hexBytes(sum[:])
}

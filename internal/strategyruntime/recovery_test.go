package strategyruntime

import (
	"reflect"
	"testing"
	"time"
)

func TestClaimedCrashIsRefusedReleasedNeverAmbiguous(t *testing.T) {
	fixture := newLeaseFixture(t)
	claimed := Claim(fixture.lease, claimInput{CurrentAuthority: fixture.authority, CurrentOwner: fixture.owner, Time: newTrustedTime(runtimeNow.Add(time.Second))}).Lease
	result := RecoverCrash(claimed, recoveryInput{})
	assertAtomic(t, result, LeaseRefused, ReservationReleased, RefusalPretransportCrash, 0)
}

func TestSubmittingOutcomeClassificationMatrix(t *testing.T) {
	tests := []struct {
		outcome         BrokerOutcome
		wantState       LeaseState
		wantDisposition ReservationDisposition
		wantCode        RefusalCode
	}{
		{OutcomeAccepted, LeaseSubmitted, ReservationTransferred, RefusalNone},
		{OutcomeDefinitiveRejected, LeaseRefused, ReservationReleased, RefusalBrokerRejected},
		{OutcomeAuthoritativeNotSubmitted, LeaseRefused, ReservationReleased, RefusalNotSubmitted},
		{OutcomeTransportUncertain, LeaseAmbiguous, ReservationHeld, RefusalTransportUncertain},
	}
	for _, test := range tests {
		t.Run(string(test.outcome), func(t *testing.T) {
			fixture := newLeaseFixture(t)
			submitting := submittingLease(t, fixture)
			result := RecoverCrash(submitting, recoveryInput{Outcome: mustOutcome(t, fixture.lease.OperationID(), test.outcome, "order-1")})
			assertAtomic(t, result, test.wantState, test.wantDisposition, test.wantCode, 0)
		})
	}
}

func TestSubmittingClassificationCannotSynthesizeMissingOutcomeEvidence(t *testing.T) {
	fixture := newLeaseFixture(t)
	submitting := submittingLease(t, fixture)
	result := ClassifySubmitting(submitting, outcomeEvidence{})
	if result.Code != RefusalInvalid || result.StateTransitions != 0 || result.Lease.State() != LeaseSubmitting || result.Lease.Disposition() != ReservationReserved {
		t.Fatalf("missing evidence synthesized terminal outcome=%+v", result)
	}
}

func TestResubmitRequiresCompleteCurrentAttestationAndKeepsSameIdentity(t *testing.T) {
	fixture := newLeaseFixture(t)
	ambiguous := ClassifySubmitting(submittingLease(t, fixture), mustOutcome(t, fixture.lease.OperationID(), OutcomeTransportUncertain, "")).Lease
	if got := AssessResubmit(ambiguous, recoveryCapability{}, fixture.authority, fixture.owner, newTrustedTime(runtimeNow.Add(3*time.Second)), 0); got.Allowed || got.Code != RefusalCapabilityUnattested || got.BrokerRequests != 0 {
		t.Fatalf("unattested resubmit allowed=%+v", got)
	}
	capability := mustRecoveryCapability(t, fixture, 1, 2)
	got := AssessResubmit(ambiguous, capability, fixture.authority, fixture.owner, newTrustedTime(runtimeNow.Add(3*time.Second)), 0)
	if !got.Allowed || got.OperationID != fixture.lease.OperationID() || got.NewLease || got.Code != RefusalNone || !reflect.DeepEqual(got.Lease, ambiguous) {
		t.Fatalf("attested same-key recovery refused/changed identity=%+v", got)
	}
	if exhausted := AssessResubmit(ambiguous, capability, fixture.authority, fixture.owner, newTrustedTime(runtimeNow.Add(3*time.Second)), 2); exhausted.Allowed || exhausted.Code != RefusalRetryExhausted {
		t.Fatalf("bounded retry exceeded=%+v", exhausted)
	}
	drifted := mustRecoveryCapability(t, fixture, 2, 2)
	if got := AssessResubmit(ambiguous, drifted, fixture.authority, fixture.owner, newTrustedTime(runtimeNow.Add(3*time.Second)), 0); got.Allowed || got.Code != RefusalCapabilityUnattested {
		t.Fatalf("drifted attestation generation allowed resubmit=%+v", got)
	}
}

func TestResubmitRejectsSubmittingReleasedAndCurrentAuthorityOrOwnerDrift(t *testing.T) {
	fixture := newLeaseFixture(t)
	submitting := submittingLease(t, fixture)
	capability := mustRecoveryCapability(t, fixture, 1, 2)
	if got := AssessResubmit(submitting, capability, fixture.authority, fixture.owner, newTrustedTime(runtimeNow.Add(3*time.Second)), 0); got.Allowed || got.BrokerRequests != 0 || got.Code != RefusalInvalidTransition {
		t.Fatalf("SUBMITTING resubmit authorized=%+v", got)
	}

	ambiguous := ClassifySubmitting(submitting, mustOutcome(t, fixture.lease.OperationID(), OutcomeTransportUncertain, "")).Lease
	input := fixture.authorityInput
	input.Guardian.Generation++
	driftedAuthority, err := newAuthoritySnapshot(input)
	if err != nil {
		t.Fatal(err)
	}
	if got := AssessResubmit(ambiguous, capability, driftedAuthority, fixture.owner, newTrustedTime(runtimeNow.Add(3*time.Second)), 0); got.Allowed || got.BrokerRequests != 0 || got.Code != RefusalAuthorityDrift {
		t.Fatalf("authority-drift resubmit authorized=%+v", got)
	}
	restarted, err := restartOwner(fixture.owner, "owner-token-2")
	if err != nil {
		t.Fatal(err)
	}
	if got := AssessResubmit(ambiguous, capability, fixture.authority, restarted, newTrustedTime(runtimeNow.Add(3*time.Second)), 0); got.Allowed || got.BrokerRequests != 0 || got.Code != RefusalStaleOwner {
		t.Fatalf("stale-owner resubmit authorized=%+v", got)
	}

	released := ReconcileReservation(ambiguous, mustOutcome(t, fixture.lease.OperationID(), OutcomeAuthoritativeNotSubmitted, "")).Lease
	if got := AssessResubmit(released, capability, fixture.authority, fixture.owner, newTrustedTime(runtimeNow.Add(6*time.Second)), 0); got.Allowed || got.Code != RefusalInvalidTransition {
		t.Fatalf("released ambiguous resubmit authorized=%+v", got)
	}
}

func TestOutcomeEvidenceRejectsContradictionsAndPredatesLease(t *testing.T) {
	fixture := newLeaseFixture(t)
	tests := []struct {
		name  string
		input outcomeEvidenceInput
	}{
		{"accepted without authoritative acceptance", validOutcomeInput(fixture.lease.OperationID(), OutcomeAccepted, "order-1", runtimeNow.Add(time.Second), func(input *outcomeEvidenceInput) { input.AcceptanceKnown = false })},
		{"rejected with broker order", validOutcomeInput(fixture.lease.OperationID(), OutcomeDefinitiveRejected, "", runtimeNow.Add(time.Second), func(input *outcomeEvidenceInput) { input.BrokerOrderID = "contradictory" })},
		{"not submitted with fill", validOutcomeInput(fixture.lease.OperationID(), OutcomeAuthoritativeNotSubmitted, "", runtimeNow.Add(time.Second), func(input *outcomeEvidenceInput) { input.FillQuantity = 1 })},
		{"uncertain claiming authoritative lookup", validOutcomeInput(fixture.lease.OperationID(), OutcomeTransportUncertain, "", runtimeNow.Add(time.Second), func(input *outcomeEvidenceInput) { input.Authoritative = true; input.LookupComplete = true })},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := newOutcomeEvidence(test.input); err == nil {
				t.Fatal("contradictory outcome evidence accepted")
			}
		})
	}

	submitting := submittingLease(t, fixture)
	beforeIssue := validOutcomeInput(fixture.lease.OperationID(), OutcomeDefinitiveRejected, "", runtimeNow.Add(-time.Second), nil)
	evidence, err := newOutcomeEvidence(beforeIssue)
	if err != nil {
		t.Fatal(err)
	}
	result := ClassifySubmitting(submitting, evidence)
	if result.Code != RefusalInvalid || result.StateTransitions != 0 || result.Lease.State() != LeaseSubmitting {
		t.Fatalf("pre-issue outcome evidence classified=%+v", result)
	}
}

func TestAmbiguousReconciliationChangesDispositionNeverLeaseState(t *testing.T) {
	fixture := newLeaseFixture(t)
	ambiguous := ClassifySubmitting(submittingLease(t, fixture), mustOutcome(t, fixture.lease.OperationID(), OutcomeTransportUncertain, "")).Lease
	for _, test := range []struct {
		outcome BrokerOutcome
		want    ReservationDisposition
	}{{OutcomeAuthoritativeNotSubmitted, ReservationReleased}, {OutcomeAccepted, ReservationTransferred}} {
		result := ReconcileReservation(ambiguous, mustOutcome(t, fixture.lease.OperationID(), test.outcome, "order-1"))
		if result.Lease.State() != LeaseAmbiguous || result.Lease.Disposition() != test.want || result.Lease.OperationID() != ambiguous.OperationID() || result.BrokerRequests != 0 {
			t.Fatalf("reconciliation revived/rewrote lease=%+v", result)
		}
	}
}

func submittingLease(t *testing.T, fixture leaseFixture) Lease {
	t.Helper()
	claimed := Claim(fixture.lease, claimInput{CurrentAuthority: fixture.authority, CurrentOwner: fixture.owner, Time: newTrustedTime(runtimeNow.Add(time.Second))}).Lease
	return BeginSubmitting(claimed, submittingInput{CurrentAuthority: fixture.authority, CurrentOwner: fixture.owner, Time: newTrustedTime(runtimeNow.Add(2 * time.Second))}).Lease
}

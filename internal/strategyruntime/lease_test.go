package strategyruntime

import (
	"reflect"
	"testing"
	"time"
)

func TestCompleteLineageIsRequiredBeforeLeaseIssuance(t *testing.T) {
	fixture := newLeaseFixture(t)
	lineage := fixture.lineage
	lineage.LegID = ""
	lineage.seal = lineageSeal(lineage)
	if _, err := newLease(leaseInput{ID: "lease-bad", OperationID: "operation-bad", Lineage: lineage, Authority: fixture.authority, Owner: fixture.owner, IssuedAt: runtimeNow, ExpiresAt: runtimeNow.Add(time.Minute)}); err == nil {
		t.Fatal("lease issued with incomplete lineage")
	}
}

func TestLeaseHappyPathIsIrreversibleAndAtomic(t *testing.T) {
	fixture := newLeaseFixture(t)
	claimed := Claim(fixture.lease, claimInput{CurrentAuthority: fixture.authority, CurrentOwner: fixture.owner, Time: newTrustedTime(runtimeNow.Add(time.Second))})
	assertAtomic(t, claimed, LeaseClaimed, ReservationReserved, RefusalNone, 0)
	if !claimed.CommitAllowed || claimed.TransportAuthorized {
		t.Fatalf("claim boundary invalid=%+v", claimed)
	}

	submitting := BeginSubmitting(claimed.Lease, submittingInput{CurrentAuthority: fixture.authority, CurrentOwner: fixture.owner, Time: newTrustedTime(runtimeNow.Add(2 * time.Second))})
	assertAtomic(t, submitting, LeaseSubmitting, ReservationReserved, RefusalNone, 0)
	if !submitting.TransportAuthorized || !submitting.Lease.TransportStarted() {
		t.Fatalf("transport marker/authorization missing=%+v", submitting)
	}

	accepted := mustOutcome(t, fixture.lease.OperationID(), OutcomeAccepted, "order-1")
	submitted := ClassifySubmitting(submitting.Lease, accepted)
	assertAtomic(t, submitted, LeaseSubmitted, ReservationTransferred, RefusalNone, 0)

	replay := ClassifySubmitting(submitted.Lease, accepted)
	if replay.Code != RefusalTerminalReplay || !reflect.DeepEqual(replay.Lease, submitted.Lease) || replay.BrokerRequests != 0 || replay.StateTransitions != 0 {
		t.Fatalf("terminal lease revived=%+v", replay)
	}
}

func TestOutOfOrderNonterminalCallsConsumeLeaseBeforeTransport(t *testing.T) {
	fixture := newLeaseFixture(t)
	begin := BeginSubmitting(fixture.lease, submittingInput{CurrentAuthority: fixture.authority, CurrentOwner: fixture.owner, Time: newTrustedTime(runtimeNow.Add(time.Second))})
	assertAtomic(t, begin, LeaseRefused, ReservationReleased, RefusalInvalidTransition, 0)
	if begin.TransportAuthorized || begin.Lease.TransportStarted() {
		t.Fatalf("out-of-order begin authorized transport=%+v", begin)
	}

	fixture = newLeaseFixture(t)
	classifyIssued := ClassifySubmitting(fixture.lease, mustOutcome(t, fixture.lease.OperationID(), OutcomeAccepted, "order-1"))
	assertAtomic(t, classifyIssued, LeaseRefused, ReservationReleased, RefusalInvalidTransition, 0)

	fixture = newLeaseFixture(t)
	claimed := Claim(fixture.lease, claimInput{CurrentAuthority: fixture.authority, CurrentOwner: fixture.owner, Time: newTrustedTime(runtimeNow.Add(time.Second))}).Lease
	classifyClaimed := ClassifySubmitting(claimed, mustOutcome(t, fixture.lease.OperationID(), OutcomeAccepted, "order-1"))
	assertAtomic(t, classifyClaimed, LeaseRefused, ReservationReleased, RefusalInvalidTransition, 0)
}

func TestEveryPretransportAuthorityMismatchConsumesLeaseRefusedReleased(t *testing.T) {
	tests := []struct {
		name string
		edit func(*authoritySnapshotInput)
		code RefusalCode
	}{
		{"activation generation ABA", func(input *authoritySnapshotInput) { input.Activation.Generation = 3 }, RefusalAuthorityDrift},
		{"guardian generation", func(input *authoritySnapshotInput) { input.Guardian.Generation++ }, RefusalAuthorityDrift},
		{"protection generation", func(input *authoritySnapshotInput) { input.Protection.Generation++ }, RefusalAuthorityDrift},
		{"reconciliation generation", func(input *authoritySnapshotInput) { input.Reconciliation.Generation++ }, RefusalAuthorityDrift},
		{"cross market", func(input *authoritySnapshotInput) { input.Market = MarketUS }, RefusalScopeMismatch},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newLeaseFixture(t)
			input := fixture.authorityInput
			test.edit(&input)
			current, err := newAuthoritySnapshot(input)
			if err != nil {
				t.Fatal(err)
			}
			result := Claim(fixture.lease, claimInput{CurrentAuthority: current, CurrentOwner: fixture.owner, Time: newTrustedTime(runtimeNow.Add(time.Second))})
			assertAtomic(t, result, LeaseRefused, ReservationReleased, test.code, 0)
			if result.TransportAuthorized {
				t.Fatal("pretransport mismatch authorized transport")
			}
		})
	}
}

func TestExpiryAndStaleOwnerFenceRefuseBeforeTransport(t *testing.T) {
	fixture := newLeaseFixture(t)
	expired := Claim(fixture.lease, claimInput{CurrentAuthority: fixture.authority, CurrentOwner: fixture.owner, Time: newTrustedTime(fixture.lease.ExpiresAt())})
	assertAtomic(t, expired, LeaseRefused, ReservationReleased, RefusalExpired, 0)

	newOwner, err := newOwnerFence(2, "owner-token-2")
	if err != nil {
		t.Fatal(err)
	}
	stale := Claim(fixture.lease, claimInput{CurrentAuthority: fixture.authority, CurrentOwner: newOwner, Time: newTrustedTime(runtimeNow.Add(time.Second))})
	assertAtomic(t, stale, LeaseRefused, ReservationReleased, RefusalStaleOwner, 0)
}

func TestOwnerRestartStrictlyIncrementsEpochAndPermanentlyFencesOldToken(t *testing.T) {
	fixture := newLeaseFixture(t)
	restarted, err := restartOwner(fixture.owner, "owner-token-2")
	if err != nil || restarted.epoch != fixture.owner.epoch+1 || restarted.token == fixture.owner.token {
		t.Fatalf("owner restart=%+v err=%v", restarted, err)
	}
	result := Claim(fixture.lease, claimInput{CurrentAuthority: fixture.authority, CurrentOwner: restarted, Time: newTrustedTime(runtimeNow.Add(time.Second))})
	assertAtomic(t, result, LeaseRefused, ReservationReleased, RefusalStaleOwner, 0)
	if _, err := restartOwner(restarted, restarted.token); err == nil {
		t.Fatal("owner token reused across epoch")
	}
}

func TestConsumedReplayPreservesOriginalAndReleasesOnlyRetryHeldReservation(t *testing.T) {
	fixture := newLeaseFixture(t)
	terminal := Claim(fixture.lease, claimInput{CurrentAuthority: fixture.authority, CurrentOwner: fixture.owner, Time: newTrustedTime(fixture.lease.ExpiresAt())}).Lease
	retry := mustRetryReservation(t, "retry-reservation", ReservationHeld)
	result := Claim(terminal, claimInput{CurrentAuthority: fixture.authority, CurrentOwner: fixture.owner, Time: newTrustedTime(runtimeNow.Add(time.Second)), RetryReservation: retry})
	if result.Code != RefusalTerminalReplay || !result.OriginalUnchanged || !reflect.DeepEqual(result.Lease, terminal) || result.RetryDisposition != ReservationReleased || result.RetryReservationID != "retry-reservation" || result.ReservationTransitions != 1 || result.StateTransitions != 0 {
		t.Fatalf("terminal replay altered original/retry disposition=%+v", result)
	}
}

func assertAtomic(t *testing.T, result AtomicResult, state LeaseState, disposition ReservationDisposition, code RefusalCode, brokerRequests uint64) {
	t.Helper()
	if result.Lease.State() != state || result.Lease.Disposition() != disposition || result.Code != code || result.BrokerRequests != brokerRequests || !result.CommitAllowed || result.NextRevision != result.ExpectedRevision+1 || result.Lease.Revision() != result.NextRevision {
		t.Fatalf("atomic result=%+v state=%s disposition=%s code=%s", result, state, disposition, code)
	}
}

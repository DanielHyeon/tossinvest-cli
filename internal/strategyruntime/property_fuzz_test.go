package strategyruntime

import (
	"testing"
	"time"
)

func FuzzTerminalLeaseNeverHasOutgoingTransition(f *testing.F) {
	f.Add(uint8(0))
	f.Add(uint8(1))
	f.Add(uint8(2))
	f.Fuzz(func(t *testing.T, selector uint8) {
		fixture := newLeaseFixture(t)
		submitting := submittingLease(t, fixture)
		outcome := []BrokerOutcome{OutcomeAccepted, OutcomeDefinitiveRejected, OutcomeTransportUncertain}[selector%3]
		terminal := ClassifySubmitting(submitting, mustOutcome(t, fixture.lease.OperationID(), outcome, "order-1")).Lease
		before := terminal
		result := BeginSubmitting(terminal, submittingInput{CurrentAuthority: fixture.authority, CurrentOwner: fixture.owner, Time: newTrustedTime(runtimeNow.Add(10 * time.Second))})
		if result.StateTransitions != 0 || result.Code != RefusalTerminalReplay || result.Lease.State() != before.State() || result.Lease.Disposition() != before.Disposition() {
			t.Fatalf("terminal transition selector=%d result=%+v", selector, result)
		}
	})
}

func FuzzAuthorityGenerationNeverRevivesLease(f *testing.F) {
	f.Add(uint64(1))
	f.Add(uint64(2))
	f.Add(^uint64(0))
	f.Fuzz(func(t *testing.T, generation uint64) {
		if generation == 0 {
			return
		}
		fixture := newLeaseFixture(t)
		input := fixture.authorityInput
		input.Activation.Generation = generation
		input.Activation.Digest = fixture.authorityInput.Activation.Digest
		current, err := newAuthoritySnapshot(input)
		if err != nil {
			t.Fatal(err)
		}
		result := Claim(fixture.lease, claimInput{CurrentAuthority: current, CurrentOwner: fixture.owner, Time: newTrustedTime(runtimeNow.Add(time.Second))})
		if (result.Lease.State() == LeaseClaimed) != (generation == fixture.authorityInput.Activation.Generation) {
			t.Fatalf("generation=%d result=%+v", generation, result)
		}
	})
}

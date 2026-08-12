package journal

// a099 R1: the outbox decides who sends.
//
// Today it does not decide. It answers a different question — "is a send owed
// for this row?" — and that answer is a property of the row, so every caller
// that asks gets the same one. claimOwed's PENDING arm returns (true, false)
// unconditionally (outbox.go:276-278), and ClaimAlertForDelivery hands that
// straight back. Nothing in the transaction records that a sender is already
// working on the row, so nothing can refuse the second sender.
//
// a096 stopped one *observer* from sending twice by holding a mutex across the
// claim and the send. A mutex excludes goroutines inside one Notifier value.
// It says nothing about a second Notifier, a second process, or the delivery
// executor a098 adds. Exclusion that lives in memory is not exclusion.
//
// This test is RED today: both senders are granted the right to send.
//
// §4.3 replaces the boolean with a claim result (design C1), so the call below
// is re-typed there. What it asserts does not change — of two senders reaching
// one PENDING row, exactly one leaves with the right to send.

import (
	"context"
	"sync"
	"testing"
	"time"
)

const a099Remind = time.Hour

func a099Alert() Alert {
	return Alert{
		EventKey: "exit.proposal_refused|pos-1|LADDER_PARTIAL|2",
		Type:     "exit.proposal_refused",
		Severity: "critical",
	}
}

// TestTwoSendersReachingOnePendingRowLeaveWithOneRightToSend is R1.
func TestTwoSendersReachingOnePendingRowLeaveWithOneRightToSend(t *testing.T) {
	j, _ := outboxJournal(t)
	ctx := context.Background()

	// The row the two senders will race for: written, PENDING, never sent, and
	// crucially not claimed. EnqueueAlert is the record-only path and takes no
	// lease, which is what makes it the honest arrangement here — claiming it
	// first would leave the row held by the test itself and both racers would
	// lose to the setup rather than to each other.
	if _, err := j.EnqueueAlert(ctx, a099Alert()); err != nil {
		t.Fatalf("arranging the pending row: %v", err)
	}
	if got := alertState(t, j, 1); got != AlertPending {
		t.Fatalf("state = %q, want %q — the arrangement is wrong", got, AlertPending)
	}

	const senders = 2
	var (
		mu      sync.Mutex
		granted int
		failed  []error
		wg      sync.WaitGroup
	)
	start := make(chan struct{})
	for i := 0; i < senders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			claim, err := j.ClaimAlertForDelivery(ctx, a099Alert(), a099Remind, testClaimant)
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err != nil:
				failed = append(failed, err)
			case claim.Disposition == ClaimAcquired:
				granted++
			}
		}()
	}
	close(start)
	wg.Wait()

	for _, err := range failed {
		t.Fatalf("ClaimAlertForDelivery: %v", err)
	}
	if granted != 1 {
		t.Errorf("senders granted the right to send = %d, want 1 — "+
			"two senders each believing the send is theirs is the double delivery "+
			"this change exists to stop", granted)
	}
}

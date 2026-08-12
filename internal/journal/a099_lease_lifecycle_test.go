package journal

// a099, the ledger half of the lease: expiry, the backwards clock, and the
// settlements a sender no longer owns.
//
// The registry in tasks.md calls these RED — timed. Each one is true today only
// because the concept it talks about does not exist yet; each becomes false at a
// named point in §4 and has to be seen failing there. The moment is written on
// each test.

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// leaseJournal opens a journal with an explicit lease so a test can put the
// lease and the reminder window in a known order. Both matter: a lease shorter
// than the window and a lease longer than it exercise different rows of C5.
func leaseJournal(t *testing.T, lease time.Duration) (*Journal, *clock.Fake) {
	t.Helper()
	clk := clock.NewFake(time.Date(2026, 7, 26, 8, 0, 0, 0, time.UTC))
	j, err := Open(context.Background(), Options{
		Path:       filepath.Join(t.TempDir(), "journal.db"),
		Clock:      clk,
		FSProber:   FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}),
		AlertLease: lease,
	})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = j.Close() })
	return j, clk
}

// TestOpenFillsTheDefaultLease is R20's ledger half, and it exists because a
// constant that is right while the wiring that carries it is missing buys
// nothing: a Journal opened with no lease would claim with a zero one, every
// claim would expire the instant it was issued, and the exclusion would be
// theatre.
func TestOpenFillsTheDefaultLease(t *testing.T) {
	j, _ := outboxJournal(t)
	if j.alertLease != DefaultAlertLease {
		t.Errorf("alertLease = %v, want %v — a zero lease expires on issue", j.alertLease, DefaultAlertLease)
	}

	explicit, _ := leaseJournal(t, 7*time.Minute)
	if explicit.alertLease != 7*time.Minute {
		t.Errorf("alertLease = %v, want the injected 7m", explicit.alertLease)
	}

	// And the value reaches the row, which is the part a constant cannot prove.
	claim, err := explicit.ClaimAlertForDelivery(context.Background(), a099Alert(), claimRemind, testClaimant)
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery: %v", err)
	}
	if got := claim.ExpiresAt.Sub(claim.ClaimedAt); got != 7*time.Minute {
		t.Errorf("stored lease span = %v, want 7m — the injected lease never reached the UPDATE", got)
	}
}

// TestAnExpiredClaimIsPickedUpByAnotherSender is R9.
//
// RED between §4.3 and §4.3d: with acquisition but no expiry predicate, a row
// whose holder died is claimable by nobody, ever. The alert that reports a
// failed stop would sit in the outbox for the life of the process.
func TestAnExpiredClaimIsPickedUpByAnotherSender(t *testing.T) {
	lease := 90 * time.Second
	j, clk := leaseJournal(t, lease)
	ctx := context.Background()

	first, err := j.ClaimAlertForDelivery(ctx, a099Alert(), claimRemind, "sender-one")
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery: %v", err)
	}
	if first.Disposition != ClaimAcquired {
		t.Fatalf("disposition = %v, want acquired", first.Disposition)
	}

	// One second before expiry the row is still the first sender's.
	clk.Advance(lease - time.Second)
	if held, err := j.ClaimAlertByID(ctx, first.ID, "sender-two"); err != nil {
		t.Fatalf("ClaimAlertByID (inside the lease): %v", err)
	} else if held.Disposition != ClaimHeldElsewhere {
		t.Fatalf("disposition = %v one second inside the lease, want held-elsewhere", held.Disposition)
	} else if held.ClaimedBy != "sender-one" {
		t.Errorf("claimed_by = %q, want sender-one — the operator has to see who is holding it", held.ClaimedBy)
	}

	// Past it, the sender that never came back loses it.
	clk.Advance(time.Second)
	second, err := j.ClaimAlertByID(ctx, first.ID, "sender-two")
	if err != nil {
		t.Fatalf("ClaimAlertByID (past the lease): %v", err)
	}
	if second.Disposition != ClaimAcquired {
		t.Fatalf("disposition = %v past the lease — a dead sender would hold this alert forever",
			second.Disposition)
	}
	if !second.Stole || second.StoleBy != "sender-one" {
		t.Errorf("stole=%v by=%q, want a recorded takeover from sender-one — "+
			"a steal means somebody stopped, and that is never silent", second.Stole, second.StoleBy)
	}
	if second.Token == first.Token {
		t.Error("the new holder got the old token — then the dead sender could still settle")
	}
}

// TestAClaimIssuedInTheFutureIsReopened is R10.
//
// RED between §4.3d and §4.3e: with only the expiry predicate, a clock that
// steps backwards leaves every stored expiry in the future, so the row stays
// locked for the whole step *plus* the lease. This repository has already had a
// backwards clock once (a096b), which is why the case is here and not in a
// comment.
func TestAClaimIssuedInTheFutureIsReopened(t *testing.T) {
	lease := 90 * time.Second
	j, clk := leaseJournal(t, lease)
	ctx := context.Background()

	first, err := j.ClaimAlertForDelivery(ctx, a099Alert(), claimRemind, "sender-one")
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery: %v", err)
	}

	// The clock steps back an hour: from here the claim was issued in the
	// future, and so is its expiry.
	clk.Advance(-time.Hour)
	reopened, err := j.ClaimAlertByID(ctx, first.ID, "sender-two")
	if err != nil {
		t.Fatalf("ClaimAlertByID (after the step back): %v", err)
	}
	if reopened.Disposition != ClaimAcquired {
		t.Errorf("disposition = %v after the clock stepped back — the row would be locked "+
			"for the whole step plus the lease", reopened.Disposition)
	}
}

// TestAClaimJustIssuedIsNotStolenAtTheSkewBoundary is a regression pin, not a
// RED. It is not red at any point in §4; it excludes a design — writing the
// backwards-clock boundary as `>=` — rather than describing behaviour that is
// missing.
//
// The boundary value is the whole of it. Exactly skew ahead is the largest gap
// the truncation of RFC3339 plus a second of headroom is allowed to produce, and
// it must still read as "the clock is fine".
func TestAClaimJustIssuedIsNotStolenAtTheSkewBoundary(t *testing.T) {
	lease := 10 * time.Minute
	j, clk := leaseJournal(t, lease)
	ctx := context.Background()

	first, err := j.ClaimAlertForDelivery(ctx, a099Alert(), claimRemind, "sender-one")
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery: %v", err)
	}

	// Step back by exactly the skew: claimed_at is now skew ahead of "now".
	clk.Advance(-alertClaimSkew)
	if held, err := j.ClaimAlertByID(ctx, first.ID, "sender-two"); err != nil {
		t.Fatalf("ClaimAlertByID: %v", err)
	} else if held.Disposition != ClaimHeldElsewhere {
		t.Errorf("disposition = %v exactly at the skew boundary, want held-elsewhere — "+
			"the boundary is strict, and a `>=` here steals live claims", held.Disposition)
	}
}

// TestALeaseHolderWithAShorterLeaseCannotStealALiveClaim is a regression pin. It
// excludes deriving expiry at judgement time from the judge's own lease, which
// is what the third draft of this design did.
//
// Two instances of the same ledger with different leases is not a contrived
// setup: C4 has obs computing the lease from live configuration and injecting
// it, so two processes with different configuration is exactly what production
// looks like while a deployment is half done.
func TestALeaseHolderWithAShorterLeaseCannotStealALiveClaim(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	clk := clock.NewFake(time.Date(2026, 7, 26, 8, 0, 0, 0, time.UTC))
	open := func(lease time.Duration) *Journal {
		j, err := Open(context.Background(), Options{
			Path: path, Clock: clk,
			FSProber:   FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}),
			AlertLease: lease,
		})
		if err != nil {
			t.Fatalf("Open with lease %v: %v", lease, err)
		}
		t.Cleanup(func() { _ = j.Close() })
		return j
	}
	generous := open(10 * time.Minute)
	ctx := context.Background()

	first, err := generous.ClaimAlertForDelivery(ctx, a099Alert(), claimRemind, "sender-one")
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery: %v", err)
	}

	impatient := open(time.Minute)
	clk.Advance(5 * time.Minute) // past the impatient one's lease, inside the issuer's
	if held, err := impatient.ClaimAlertByID(ctx, first.ID, "sender-two"); err != nil {
		t.Fatalf("ClaimAlertByID: %v", err)
	} else if held.Disposition != ClaimHeldElsewhere {
		t.Errorf("disposition = %v — a claimant judged somebody else's live claim by its own "+
			"lease, which is the row nobody abandoned being stolen", held.Disposition)
	}
}

// TestASettlementFromALostLeaseLeavesTheNewHolderAlone is R13.
//
// RED before §4.5: without the token in the settlement predicate, id and state
// are enough, so the sender that stalled long enough to lose its row settles the
// send the new holder is in the middle of making.
func TestASettlementFromALostLeaseLeavesTheNewHolderAlone(t *testing.T) {
	lease := 60 * time.Second
	j, clk := leaseJournal(t, lease)
	ctx := context.Background()

	stalled, err := j.ClaimAlertForDelivery(ctx, a099Alert(), claimRemind, "sender-one")
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery: %v", err)
	}
	clk.Advance(lease + time.Second)
	holder, err := j.ClaimAlertByID(ctx, stalled.ID, "sender-two")
	if err != nil {
		t.Fatalf("ClaimAlertByID: %v", err)
	}
	if holder.Disposition != ClaimAcquired {
		t.Fatalf("disposition = %v, want acquired — the arrangement is wrong", holder.Disposition)
	}

	// The stalled sender wakes up and tries to finish.
	res, err := j.MarkAlertDelivered(ctx, stalled.ID, stalled.Token)
	if err != nil {
		t.Fatalf("MarkAlertDelivered from the stalled sender: %v", err)
	}
	if res.Outcome != SettleLeaseLost {
		t.Errorf("outcome = %v, want lease-lost", res.Outcome)
	}
	if res.ClaimedBy != "sender-two" {
		t.Errorf("claimed_by = %q, want sender-two — the sender has to be told who it lost to",
			res.ClaimedBy)
	}
	if got := alertState(t, j, stalled.ID); got != AlertPending {
		t.Errorf("state = %q, want %q — the stalled sender settled somebody else's row", got, AlertPending)
	}

	// And the new holder's own settlement still works, which is the half that
	// says the refusal was targeted rather than a blanket lock.
	if applied, err := j.MarkAlertDelivered(ctx, holder.ID, holder.Token); err != nil {
		t.Fatalf("MarkAlertDelivered from the holder: %v", err)
	} else if applied.Outcome != SettleApplied {
		t.Errorf("outcome = %v for the holder, want applied", applied.Outcome)
	}
}

// TestTheSameSenderNameCannotSettleAnEarlierLease is R14 — the ABA case.
//
// RED at the same point as R13, and it is the reason the predicate is the token
// and not the holder's name. A delivery loop that restarts keeps its name. If
// the name were the proof, the loop that died mid-send could come back, re-take
// the row, and its *previous* incarnation's settlement would be accepted against
// the new lease.
func TestTheSameSenderNameCannotSettleAnEarlierLease(t *testing.T) {
	lease := 60 * time.Second
	j, clk := leaseJournal(t, lease)
	ctx := context.Background()

	before, err := j.ClaimAlertForDelivery(ctx, a099Alert(), claimRemind, "delivery-loop")
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery: %v", err)
	}
	clk.Advance(lease + time.Second)
	after, err := j.ClaimAlertByID(ctx, before.ID, "delivery-loop") // same name, new life
	if err != nil {
		t.Fatalf("ClaimAlertByID: %v", err)
	}
	if after.Disposition != ClaimAcquired {
		t.Fatalf("disposition = %v, want acquired", after.Disposition)
	}
	if after.Token == before.Token {
		t.Fatal("the re-acquired claim reused the token — then the two lives are indistinguishable")
	}

	res, err := j.MarkAlertAttemptFailed(ctx, before.ID, before.Token, "the previous life")
	if err != nil {
		t.Fatalf("MarkAlertAttemptFailed: %v", err)
	}
	if res.Outcome != SettleLeaseLost {
		t.Errorf("outcome = %v, want lease-lost — the name matched and the lease did not",
			res.Outcome)
	}
}

// TestARecordedAlertCanBeClaimedImmediately is R16.
//
// RED between §4.3 and §4.4b: if the recording path went through the claiming
// one, every row replay.go wrote would come out already leased to a caller that
// never delivers and never settles, and would be unsendable until it expired.
//
// Compiling is not the assertion. A version that satisfied the new signature by
// passing a claimant string through would compile and would be exactly the bug.
func TestARecordedAlertCanBeClaimedImmediately(t *testing.T) {
	j, _ := outboxJournal(t)
	ctx := context.Background()

	id, err := j.EnqueueAlert(ctx, a099Alert())
	if err != nil {
		t.Fatalf("EnqueueAlert: %v", err)
	}

	var token string
	var at any
	if err := j.db.QueryRowContext(ctx,
		`SELECT claim_token, claimed_at FROM alert_outbox WHERE id = ?`, id).Scan(&token, &at); err != nil {
		t.Fatalf("reading the lease: %v", err)
	}
	if token != "" || at != nil {
		t.Errorf("recorded row holds a lease (token=%q at=%v) — the recorder never settles, "+
			"so nothing would release it", token, at)
	}

	claim, err := j.ClaimAlertByID(ctx, id, testClaimant)
	if err != nil {
		t.Fatalf("ClaimAlertByID: %v", err)
	}
	if claim.Disposition != ClaimAcquired {
		t.Errorf("disposition = %v on a recorded row, want acquired", claim.Disposition)
	}
}

// TestReArmingClearsTheLeaseOfThePreviousEpisode is R17.
//
// RED before §4.6. The condition matters and is arranged deliberately: the lease
// has to outlast the reminder window, or the previous episode's claim has
// already expired by the time the re-arm happens and the test passes without
// testing anything.
func TestReArmingClearsTheLeaseOfThePreviousEpisode(t *testing.T) {
	remind := time.Hour
	j, clk := leaseJournal(t, 3*remind) // lease > remindAfter, or this is vacuous
	ctx := context.Background()

	first, err := j.ClaimAlertForDelivery(ctx, a099Alert(), remind, "sender-one")
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery: %v", err)
	}
	if _, err := j.MarkAlertDelivered(ctx, first.ID, first.Token); err != nil {
		t.Fatalf("MarkAlertDelivered: %v", err)
	}
	// Put the lease back on by hand: this is the ACKNOWLEDGED-with-a-lease row
	// of C5, the state the operator's release leaves behind when it lands on a
	// row somebody is sending.
	if _, err := j.db.ExecContext(ctx,
		`UPDATE alert_outbox SET claim_token = 'stale', claimed_by = 'sender-one',
		        claimed_at = ?, claim_expires_at = ? WHERE id = ?`,
		RFC3339(clk.Now()), RFC3339(clk.Now().Add(3*remind)), first.ID); err != nil {
		t.Fatalf("arranging a surviving lease: %v", err)
	}

	clk.Advance(remind)
	again, err := j.ClaimAlertForDelivery(ctx, a099Alert(), remind, "sender-two")
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery (reminder): %v", err)
	}
	if again.Disposition != ClaimAcquired {
		t.Fatalf("disposition = %v on the re-armed row — the new episode inherited the old "+
			"episode's lease, and its holder is not sending this", again.Disposition)
	}
	if again.Stole {
		t.Error("stole = true on a re-armed row — the re-arm cleared the lease, so there was " +
			"no live claim to displace, and reporting one would send a false alarm about a dead sender")
	}
	if res, err := j.MarkAlertDelivered(ctx, again.ID, "stale"); err != nil {
		t.Fatalf("MarkAlertDelivered with the previous episode's token: %v", err)
	} else if res.Outcome == SettleApplied {
		t.Error("the previous episode's token settled the new episode")
	}
}

// TestPendingAlertsProjectTheLeaseButNeverTheToken is R25.
//
// RED before §4.10. Two undelivered rows look identical to an operator without
// this — one nobody is sending and one a sender is holding — and the triage
// decision is different for each.
//
// The token's absence is half the assertion and not a stylistic preference: it
// is the proof of ownership, and a listing that showed it would let anyone who
// read the screen settle somebody else's send.
func TestPendingAlertsProjectTheLeaseButNeverTheToken(t *testing.T) {
	j, _ := leaseJournal(t, 5*time.Minute)
	ctx := context.Background()

	claim, err := j.ClaimAlertForDelivery(ctx, a099Alert(), claimRemind, "sender-one")
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery: %v", err)
	}
	idle, err := j.EnqueueAlert(ctx, Alert{
		EventKey: "order.unresolved_in_doubt|att-77", Type: "order.unresolved_in_doubt",
		Severity: "critical"})
	if err != nil {
		t.Fatalf("EnqueueAlert: %v", err)
	}

	rows, err := j.PendingAlerts(ctx, 0)
	if err != nil {
		t.Fatalf("PendingAlerts: %v", err)
	}
	byID := map[int64]Alert{}
	for _, r := range rows {
		byID[r.ID] = r
	}
	held, ok := byID[claim.ID]
	if !ok {
		t.Fatalf("the held row is missing from the backlog — a row being sent is still undelivered")
	}
	if held.ClaimedBy != "sender-one" {
		t.Errorf("claimed_by = %q, want sender-one — the operator cannot tell stuck from in-flight",
			held.ClaimedBy)
	}
	if held.ClaimedAt == nil || held.ClaimExpiresAt == nil {
		t.Errorf("claimed_at=%v claim_expires_at=%v — the operator needs to know since when and until when",
			held.ClaimedAt, held.ClaimExpiresAt)
	}
	free, ok := byID[idle]
	if !ok {
		t.Fatalf("the unclaimed row is missing from the backlog")
	}
	if free.ClaimedBy != "" || free.ClaimedAt != nil || free.ClaimExpiresAt != nil {
		t.Errorf("an unclaimed row projects a lease: %+v", free)
	}

	// Nothing on the projected row is the token. Alert has no field for it, so
	// the check is that the value cannot be found anywhere the projection
	// reaches.
	if claim.Token == "" {
		t.Fatal("the claim carries no token, so this check would pass vacuously")
	}
	for _, r := range rows {
		for _, s := range []string{r.ClaimedBy, r.Title, r.Body, r.Payload, r.LastError, r.EventKey} {
			if s == claim.Token {
				t.Errorf("the claim token appears in the operator's projection — "+
					"whoever reads it can settle somebody else's send (row %d)", r.ID)
			}
		}
	}
}

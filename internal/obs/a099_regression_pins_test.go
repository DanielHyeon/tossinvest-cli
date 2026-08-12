package obs_test

// a099 §3.3 — the sender's regression pins.
//
// Neither of these is RED at any point in §4. Each excludes a design that was
// written down and rejected, and the rejected design is named on it.

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// closesTheJournal publishes successfully and then closes the sender's handle on
// the journal, so the settlement that follows fails for a reason that has
// nothing to do with the lease.
type closesTheJournal struct {
	j     *journal.Journal
	calls int
}

func (p *closesTheJournal) Publish(context.Context, obs.Notification) error {
	p.calls++
	if p.calls == 1 && p.j != nil {
		_ = p.j.Close()
	}
	return nil
}

// TestAPublishedButUnsettledRowKeepsItsLease is R8.
//
// It excludes releasing the claim at the "published, could not record it" exit.
// That release reads like tidiness and is the storm: the push has gone out, the
// row is still PENDING, and an unleased PENDING row is one the next observation
// picks up and sends again — and again, because the settlement is still broken.
// The lease left in place is the only thing suppressing that, and expiry is what
// eventually lets go.
//
// Two handles on one file: the sender's is closed underneath it, and the test
// keeps its own to read what the row looks like afterwards.
func TestAPublishedButUnsettledRowKeepsItsLease(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.db")
	clk := clock.NewFake(obsNow)
	open := func() *journal.Journal {
		j, err := journal.Open(context.Background(), journal.Options{
			Path:     path,
			Clock:    clk,
			FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
		})
		if err != nil {
			t.Fatalf("journal.Open: %v", err)
		}
		return j
	}
	senders := open()
	observer := open()
	t.Cleanup(func() { _ = observer.Close() })

	pub := &closesTheJournal{j: senders}
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})
	n := &obs.Notifier{
		Log:         obs.NewLogger(obs.LogOptions{Writer: &bytes.Buffer{}, JSON: true, Clock: clk}),
		Publisher:   pub,
		Journal:     senders,
		Gate:        gate,
		Clock:       clock.System(),
		Attempts:    1,
		RetryDelay:  time.Millisecond,
		RemindAfter: a099Remind,
	}

	if err := n.Notify(context.Background(), a099Event()); err != nil {
		t.Fatalf("Notify: %v", err)
	}

	pending, err := observer.PendingAlerts(context.Background(), 0)
	if err != nil {
		t.Fatalf("PendingAlerts: %v", err)
	}
	if len(pending) != 1 {
		t.Fatalf("pending rows = %d, want 1 — the push went out and the record did not", len(pending))
	}
	if pending[0].ClaimedBy == "" || pending[0].ClaimExpiresAt == nil {
		t.Errorf("the published-but-unsettled row lost its lease (%+v) — the next observation "+
			"would find it unheld and send the same alert again, which is the storm through "+
			"the success path", pending[0])
	}
	// And the block is the behaviour that already existed: an unaccountable send
	// stops new entries.
	if gate.CheckEntry() == nil {
		t.Error("new entries are still allowed after a send that could not be recorded")
	}
}

// TestContentionNeitherLocksNorUnlocksTheEntryGate is R23.
//
// It excludes two designs at once, one from each direction.
//
// Latching on contention was the second draft. EntryGate.Block is a per-reason
// latch in memory with no automatic release, so a routine race between two
// healthy senders would leave a lock only a human can open — after a delivery
// that succeeded.
//
// Deriving the release from the outbox was the third draft, and it is worse: it
// unlatches on a machine's judgement a norm requires a human to make.
//
// What it does not cover: a Block followed by a Clear of that same new reason
// inside one call. Nothing in these paths has that shape — Clear has exactly one
// caller and it is Acknowledge — and §5.3b diffs the five gate sites to say so
// structurally. Written down because an uncovered case that nobody names reads
// like a covered one.
func TestContentionNeitherLocksNorUnlocksTheEntryGate(t *testing.T) {
	pub := newA099Publisher(true)
	defer pub.releaseAll()
	holder, latecomer, _ := a099Pair(t, pub)
	ctx := context.Background()

	// A pre-existing block from an unrelated reason: if anything on the
	// contention path cleared latches, this is what it would take with it.
	const sentinel = execgw.ReasonBrokerAuthRejected
	latecomer.Gate.Block(sentinel, "an earlier, unrelated refusal")
	before := latecomer.Gate.Blocks()

	done := make(chan error, 1)
	go func() { done <- holder.Notify(ctx, a099Event()) }()
	pub.waitInside(t)
	if err := latecomer.Notify(ctx, a099Event()); err != nil {
		t.Fatalf("the second sender's Notify: %v", err)
	}
	pub.releaseAll()
	if err := <-done; err != nil {
		t.Fatalf("the holder's Notify: %v", err)
	}

	after := latecomer.Gate.Blocks()
	if len(after) != len(before) {
		t.Fatalf("blocks = %v, want %v — losing a claim is a normal race and must move the "+
			"gate in neither direction", after, before)
	}
	for reason, detail := range before {
		if after[reason] != detail {
			t.Errorf("block %v changed from %q to %q", reason, detail, after[reason])
		}
	}
	if _, ok := after[execgw.ReasonAlertUndelivered]; ok {
		t.Error("contention latched the alert-undelivered block — a successful delivery would " +
			"leave behind a lock only a human can open")
	}
}

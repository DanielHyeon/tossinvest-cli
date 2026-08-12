package obs_test

// a097 R2: a critical alert that could not even be *recorded* must not be
// silent.
//
// a074 latches the entry gate when a critical alert cannot be delivered, and
// a096b extended that to a send this system cannot account for. Both of those
// are downstream of a row existing. When the claim transaction itself fails
// there is no row, no send, no gate latch, no escalation and no log line — and
// internal/flatten/flatten.go:694 discards the returned error, so for that
// caller the failure is invisible end to end.
//
// The measured RED coverage says the same thing more plainly: claimAndDeliver's
// error branch (B1@227) has count=0. Nothing in this repository has ever
// executed it.

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// a097BrokenClaim builds a notifier whose journal is closed, so every outbox
// write fails at BeginTx. That is the general shape of the failure — the store
// is unavailable — and it is the shape that matters most, because it is also
// the shape in which the durable response cannot be written either.
func a097BrokenClaim(t *testing.T) (*obs.Notifier, *execgw.EntryGate, *bytes.Buffer) {
	t.Helper()
	clk := clock.NewFake(obsNow)
	j := openJournal(t, clk)
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})
	buf := &bytes.Buffer{}
	n := &obs.Notifier{
		Log:        obs.NewLogger(obs.LogOptions{Writer: buf, JSON: true, Clock: clk}),
		Publisher:  &failingPublisher{},
		Journal:    j,
		Gate:       gate,
		AccountRef: "acct-a097",
		Clock:      clock.System(),
		Attempts:   3,
		RetryDelay: time.Millisecond,
	}
	if err := j.Close(); err != nil {
		t.Fatalf("closing the journal to break the claim: %v", err)
	}
	return n, gate, buf
}

// TestAClaimThatFailsBlocksNewEntries: the operator was not told and cannot be
// told later, because nothing was written down. Continuing to open positions in
// that state is the one thing the alert existed to prevent.
//
// The block is entries only. Exits are untouched — §0.3 is not negotiable and
// no alert failure may slow a stop.
func TestAClaimThatFailsBlocksNewEntries(t *testing.T) {
	n, gate, buf := a097BrokenClaim(t)
	ctx := context.Background()

	if err := n.Notify(ctx, a096Event()); err == nil {
		t.Fatal("Notify returned nil with a closed journal — the outbox write cannot have succeeded")
	}

	if gate.CheckEntry() == nil {
		t.Error("new entries are still allowed after a critical alert could not be recorded — " +
			"nothing was sent, nothing was written, and the engine would open a position on it")
	}
	if !strings.Contains(buf.String(), string(obs.EventAlertUndelivered)) {
		t.Errorf("no %s line for a failed claim — this branch has no log at all today; got:\n%s",
			obs.EventAlertUndelivered, buf.String())
	}
}

// TestAClaimThatFailsAttemptsTheDurableBlock: the gate latch lives in a map in
// memory (execgw/retry.go:498). When the claim fails there is also no outbox
// row. So a restart erases the latch *and* the evidence, and entries reopen
// with nobody having been told anything.
//
// The first draft of a097 argued the opposite — latch cheaply, do not escalate,
// because escalation is durable and expensive to undo. That reasoning had the
// facts backwards: the durable part is exactly the part that is missing.
// escalate's own comment names it, "a restart would lift the block".
//
// Here the journal is closed, so the escalation fails too. It is still made,
// and its failure is logged at error level. A recorded failure to make the block
// durable is not the same as never having tried: the first tells an operator
// reading the log that a restart will reopen entries.
func TestAClaimThatFailsAttemptsTheDurableBlock(t *testing.T) {
	n, _, buf := a097BrokenClaim(t)
	ctx := context.Background()

	if err := n.Notify(ctx, a096Event()); err == nil {
		t.Fatal("Notify returned nil with a closed journal")
	}

	if !strings.Contains(buf.String(), string(obs.EventOperatingMode)) {
		t.Errorf("no %s line for a failed claim — the durable block was never attempted, "+
			"so a restart lifts the entry block and nothing records that; got:\n%s",
			obs.EventOperatingMode, buf.String())
	}
}

// TestAFailedClaimStillReturnsItsError is a guard, not a pin: it asserts a
// behaviour a097 did NOT change, and it passes with the a097 production diff
// reverted. That is the point. Latching the gate is an addition, and a change
// that quietly turned the error into a handled-internally outcome would trade
// one silence for another without any test noticing.
func TestAFailedClaimStillReturnsItsError(t *testing.T) {
	n, _, _ := a097BrokenClaim(t)

	err := n.Notify(context.Background(), a096Event())
	if err == nil {
		t.Fatal("Notify returned nil with a closed journal")
	}
	if !strings.Contains(err.Error(), "recording a critical alert") {
		t.Errorf("error = %q, want it to still name the outbox write — "+
			"handling the failure must not swallow the report of it", err)
	}
}

// TestAFailedClaimWithNothingWiredStillReports covers the assembly the other
// tests here do not: Gate, Log and AccountRef are all optional fields, and with
// all three empty every consequence a097 adds is skipped.
//
// That configuration is not a violation of the contract, it is outside its
// reach — a notifier with no gate cannot block anything, and one with no account
// reference has nowhere to write a durable transition. The engine's own assembly
// supplies all three (internal/app/engine/exitwiring.go:73-78). What must hold
// in every assembly is the last channel that does not depend on wiring: the
// error reaches the caller, and the nil guards do not panic on the way.
func TestAFailedClaimWithNothingWiredStillReports(t *testing.T) {
	clk := clock.NewFake(obsNow)
	j := openJournal(t, clk)
	n := &obs.Notifier{
		Publisher:  &failingPublisher{},
		Journal:    j,
		Clock:      clock.System(),
		Attempts:   3,
		RetryDelay: time.Millisecond,
	}
	if err := j.Close(); err != nil {
		t.Fatalf("closing the journal to break the claim: %v", err)
	}

	err := n.Notify(context.Background(), a096Event())
	if err == nil {
		t.Fatal("Notify returned nil with a closed journal and nothing wired")
	}
	if !strings.Contains(err.Error(), "recording a critical alert") {
		t.Errorf("error = %q, want it to name the outbox write", err)
	}
}

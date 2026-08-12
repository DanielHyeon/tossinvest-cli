package journal

// a099 §3.3 — the ledger's regression pins.
//
// None of these is RED at any point in §4. They are not descriptions of missing
// behaviour; each one excludes a design that was considered and rejected, and
// the design it excludes is named on it. Written down because a rejected design
// that leaves no test behind comes back: three of the four below were in a draft
// of this change.

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// TestRecordingAFailedAttemptKeepsTheLease is R6.
//
// This is not a RED. Before §4.5 the settlement functions do not know leases
// exist, so nothing here can fail. What it excludes is the first draft — "a
// recorded failure releases the claim" — which reads reasonably and is wrong: a
// sender records a failure between two attempts it is about to make, and a
// second sender arriving in that gap would publish the same alert.
func TestRecordingAFailedAttemptKeepsTheLease(t *testing.T) {
	j, _ := leaseJournal(t, 10*time.Minute)
	ctx := context.Background()

	claim, err := j.ClaimAlertForDelivery(ctx, a099Alert(), claimRemind, "sender-one")
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery: %v", err)
	}
	if _, err := j.MarkAlertAttemptFailed(ctx, claim.ID, claim.Token, "transport is down"); err != nil {
		t.Fatalf("MarkAlertAttemptFailed: %v", err)
	}

	var token, by string
	if err := j.db.QueryRowContext(ctx,
		`SELECT claim_token, claimed_by FROM alert_outbox WHERE id = ?`, claim.ID).
		Scan(&token, &by); err != nil {
		t.Fatalf("reading the lease: %v", err)
	}
	if token != claim.Token || by != "sender-one" {
		t.Errorf("lease after a recorded failure = (%q, %q), want the sender's own — "+
			"a released lease lets a second sender in between two attempts of the first",
			token, by)
	}
	if held, err := j.ClaimAlertByID(ctx, claim.ID, "sender-two"); err != nil {
		t.Fatalf("ClaimAlertByID: %v", err)
	} else if held.Disposition != ClaimHeldElsewhere {
		t.Errorf("disposition = %v after a recorded failure, want held-elsewhere", held.Disposition)
	}
}

// TestARowBeingSentIsStillUndelivered is R21.
//
// It excludes the design this one was chosen over: a SENDING state. A third
// state would have had to be added to the predicates the entry gate reads, and
// a sender that died mid-send would have left the row in a state no reader
// counts — an undelivered critical alert that the gate cannot see.
//
// The lease is orthogonal to the state, and these two queries are where that
// pays off: neither of them mentions it.
func TestARowBeingSentIsStillUndelivered(t *testing.T) {
	j, _ := leaseJournal(t, 10*time.Minute)
	ctx := context.Background()

	claim, err := j.ClaimAlertForDelivery(ctx, a099Alert(), claimRemind, "sender-one")
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery: %v", err)
	}

	count, err := j.UndeliveredCount(ctx)
	if err != nil {
		t.Fatalf("UndeliveredCount: %v", err)
	}
	if count != 1 {
		t.Errorf("undelivered = %d while a sender holds the row, want 1 — the entry gate "+
			"reads this number and a row in flight is not a row delivered", count)
	}
	rows, err := j.PendingAlerts(ctx, 0)
	if err != nil {
		t.Fatalf("PendingAlerts: %v", err)
	}
	if len(rows) != 1 || rows[0].ID != claim.ID {
		t.Errorf("pending rows = %+v, want the held row — an operator's backlog that hides "+
			"rows in flight hides the ones a dead sender is holding", rows)
	}
}

// TestAcknowledgementIgnoresTheLease is R22.
//
// It excludes putting the lease into the acknowledgement predicate. A person
// asserting they have seen an alert must not be blocked by a machine being
// midway through telling them — and if a sender has died, the operator would be
// blocked for the length of the lease, which is precisely when they are most
// likely to be acknowledging things by hand.
func TestAcknowledgementIgnoresTheLease(t *testing.T) {
	j, _ := leaseJournal(t, 10*time.Minute)
	ctx := context.Background()

	claim, err := j.ClaimAlertForDelivery(ctx, a099Alert(), claimRemind, "sender-one")
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery: %v", err)
	}
	if err := j.AcknowledgeAlert(ctx, claim.ID, "daniel"); err != nil {
		t.Errorf("AcknowledgeAlert on a row under a live lease: %v — the operator outranks "+
			"the sender", err)
	}
	if got := alertState(t, j, claim.ID); got != AlertAcknowledged {
		t.Errorf("state = %q, want %q", got, AlertAcknowledged)
	}
}

// TestAnAcknowledgedRowRefusesTheSendersSettlement is R26.
//
// It excludes loosening the state predicate while adding the token one. The
// state test is what makes an operator's acknowledgement final; a settlement
// that matched on the token alone would move an ACKNOWLEDGED row to DELIVERED
// and the ledger would record that a human saw something they never saw.
func TestAnAcknowledgedRowRefusesTheSendersSettlement(t *testing.T) {
	j, _ := leaseJournal(t, 10*time.Minute)
	ctx := context.Background()

	claim, err := j.ClaimAlertForDelivery(ctx, a099Alert(), claimRemind, "sender-one")
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery: %v", err)
	}
	if err := j.AcknowledgeAlert(ctx, claim.ID, "daniel"); err != nil {
		t.Fatalf("AcknowledgeAlert: %v", err)
	}

	res, err := j.MarkAlertDelivered(ctx, claim.ID, claim.Token)
	if err != nil {
		t.Fatalf("MarkAlertDelivered: %v", err)
	}
	if res.Outcome != SettleAlreadySettled {
		t.Errorf("outcome = %v with a matching token on an acknowledged row, want already-settled",
			res.Outcome)
	}
	if got := alertState(t, j, claim.ID); got != AlertAcknowledged {
		t.Errorf("state = %q, want %q — the sender overwrote the operator's acknowledgement",
			got, AlertAcknowledged)
	}
}

// TestOneSenderAloneDecidesExactlyAsBefore is R24.
//
// It excludes the shape of failure this change most plausibly has: reading "the
// claim was not acquired" as "do not send" in a case where the old code would
// have sent. With one sender and nobody dead, every row of the decision table
// has to come out the way it did before the lease existed.
//
// One case would be weak, so this walks the table: never sent, sent and failed,
// delivered inside the window, delivered past it, acknowledged inside the
// window, and a state this build does not recognise.
func TestOneSenderAloneDecidesExactlyAsBefore(t *testing.T) {
	type step struct {
		name    string
		arrange func(t *testing.T, j *Journal, id int64, clk *clock.Fake)
		want    ClaimDisposition
	}
	// Each case arranges the row and then makes ONE claim, which is what a single
	// sender does: it holds its delivery mutex from the claim through the
	// settlement, so it never asks twice while still holding.
	claimIt := func(t *testing.T, j *Journal, id int64) ClaimResult {
		t.Helper()
		c, err := j.ClaimAlertByID(context.Background(), id, "the-only-sender")
		if err != nil {
			t.Fatalf("arranging: %v", err)
		}
		if c.Disposition != ClaimAcquired {
			t.Fatalf("arranging: disposition = %v", c.Disposition)
		}
		return c
	}
	steps := []step{
		{"never sent", func(*testing.T, *Journal, int64, *clock.Fake) {}, ClaimAcquired},
		{"sent and failed", func(t *testing.T, j *Journal, id int64, _ *clock.Fake) {
			c := claimIt(t, j, id)
			if _, err := j.MarkAlertAttemptFailed(context.Background(), id, c.Token, "down"); err != nil {
				t.Fatal(err)
			}
			if _, err := j.ReleaseAlertClaim(context.Background(), id, c.Token); err != nil {
				t.Fatal(err)
			}
		}, ClaimAcquired},
		{"delivered, inside the window", func(t *testing.T, j *Journal, id int64, clk *clock.Fake) {
			c := claimIt(t, j, id)
			if _, err := j.MarkAlertDelivered(context.Background(), id, c.Token); err != nil {
				t.Fatal(err)
			}
			clk.Advance(claimRemind - time.Second)
		}, ClaimSettled},
		{"delivered, past the window", func(t *testing.T, j *Journal, id int64, clk *clock.Fake) {
			c := claimIt(t, j, id)
			if _, err := j.MarkAlertDelivered(context.Background(), id, c.Token); err != nil {
				t.Fatal(err)
			}
			clk.Advance(claimRemind)
		}, ClaimAcquired},
		{"acknowledged, inside the window", func(t *testing.T, j *Journal, id int64, clk *clock.Fake) {
			if err := j.AcknowledgeAlert(context.Background(), id, "daniel"); err != nil {
				t.Fatal(err)
			}
			clk.Advance(claimRemind - time.Second)
		}, ClaimSettled},
		{"acknowledged, past the window", func(t *testing.T, j *Journal, id int64, clk *clock.Fake) {
			if err := j.AcknowledgeAlert(context.Background(), id, "daniel"); err != nil {
				t.Fatal(err)
			}
			clk.Advance(claimRemind)
		}, ClaimAcquired},
		{"a state this build does not know", func(t *testing.T, j *Journal, id int64, _ *clock.Fake) {
			if _, err := j.db.ExecContext(context.Background(),
				`UPDATE alert_outbox SET state = 'QUARANTINED' WHERE id = ?`, id); err != nil {
				t.Fatal(err)
			}
		}, ClaimAcquired},
	}

	for _, s := range steps {
		t.Run(s.name, func(t *testing.T) {
			j, clk := leaseJournal(t, 10*time.Minute)
			ctx := context.Background()
			id, err := j.EnqueueAlert(ctx, a099Alert())
			if err != nil {
				t.Fatalf("arranging: %v", err)
			}
			s.arrange(t, j, id, clk)

			got, err := j.ClaimAlertForDelivery(ctx, a099Alert(), claimRemind, "the-only-sender")
			if err != nil {
				t.Fatalf("ClaimAlertForDelivery: %v", err)
			}
			if got.Disposition != s.want {
				t.Errorf("disposition = %v, want %v — one sender alone must decide exactly as "+
					"it did before the lease existed", got.Disposition, s.want)
			}
			if got.ID != id {
				t.Errorf("id = %d, want %d — the same condition is the same row", got.ID, id)
			}
		})
	}
}

// TestAFailedV31MigrationLeavesTheJournalAtV30 is R19.
//
// It is not RED at any point: the general property — a step that dies partway
// rolls back with its user_version — is already pinned by
// TestFailedMigrationLeavesTheJournalRestorable, which does it for v4→v5. What
// that test cannot do is cover a case that did not exist when it was written, so
// the v31 case is added here rather than assumed.
func TestAFailedV31MigrationLeavesTheJournalAtV30(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.db")
	ctx := context.Background()

	old := openJournalAtSchema(t, path, 30)
	if _, err := old.db.Exec(`INSERT INTO alert_outbox
		(event_key, event_type, severity, state, attempts, created_at)
		VALUES ('k','order.unresolved_in_doubt','critical','PENDING',1,'2026-08-11T04:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}

	// Three of the four columns land and the fourth does not.
	broken := append(migrationsThrough(30), migration{
		Version: 31,
		SQL: `ALTER TABLE alert_outbox ADD COLUMN claim_token TEXT NOT NULL DEFAULT '';
		      ALTER TABLE alert_outbox ADD COLUMN claimed_by  TEXT NOT NULL DEFAULT '';
		      ALTER TABLE alert_outbox ADD COLUMN claimed_at  TEXT;
		      ALTER TABLE table_that_does_not_exist ADD COLUMN claim_expires_at TEXT;`,
	})
	_, err := Open(ctx, Options{
		Path:              path,
		Clock:             clock.NewFake(migrationTestInstant),
		FSProber:          FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}),
		migrationOverride: &migrationPlan{steps: broken, target: 31},
	})
	if err == nil {
		t.Fatal("a v31 step that fails partway must fail the open")
	}

	survivor := openJournalAtSchema(t, path, 30)
	defer survivor.Close()
	if version, verr := survivor.SchemaVersion(ctx); verr != nil || version != 30 {
		t.Fatalf("version = %d err = %v, want 30 — the failed step took its version with it",
			version, verr)
	}
	// Not one of the four columns survived: a journal claiming v30 with three of
	// them present is a journal a v30 binary reads as v30 and a v31 binary would
	// migrate on top of.
	rows, qerr := survivor.db.Query(`SELECT name FROM pragma_table_info('alert_outbox')`)
	if qerr != nil {
		t.Fatal(qerr)
	}
	defer rows.Close()
	var leftovers []string
	for rows.Next() {
		var name string
		if serr := rows.Scan(&name); serr != nil {
			t.Fatal(serr)
		}
		if strings.HasPrefix(name, "claim") {
			leftovers = append(leftovers, name)
		}
	}
	if len(leftovers) != 0 {
		t.Errorf("lease columns survived a failed migration: %v", leftovers)
	}
}

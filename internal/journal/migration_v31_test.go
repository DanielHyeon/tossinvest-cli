package journal

// a099 §4.2: a journal written before the delivery lease existed migrates
// forward, and every row it was already holding comes out claimable.
//
// This is the migration's whole risk. The four columns are additive, so nothing
// can be lost; what could be lost is the *meaning* of an existing row. A row
// that came out of the migration looking claimed would be a critical alert
// nobody can send until a lease that was never issued expires — and the rows in
// this table are the ones that report a failed stop.

import (
	"context"
	"path/filepath"
	"testing"
)

func TestMigrationV30ToV31LeavesExistingAlertsClaimable(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	old := openJournalAtSchema(t, path, 30)
	if _, err := old.db.Exec(`INSERT INTO alert_outbox
		(event_key, event_type, severity, title, body, payload, state, attempts, created_at)
		VALUES ('exit.proposal_refused|pos-1|LADDER_PARTIAL|2','exit.proposal_refused','critical',
		        'protective exit refused','the broker refused the exit order','{}','PENDING',2,
		        '2026-08-11T04:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}

	j := openJournalAtSchema(t, path, 31)
	defer j.Close()
	if version, err := j.SchemaVersion(context.Background()); err != nil || version != 31 {
		t.Fatalf("version=%d err=%v", version, err)
	}

	var id int64
	var token, by string
	var at, expires any
	var attempts int
	if err := j.db.QueryRow(`SELECT id, claim_token, claimed_by, claimed_at, claim_expires_at, attempts
		FROM alert_outbox WHERE event_key='exit.proposal_refused|pos-1|LADDER_PARTIAL|2'`).
		Scan(&id, &token, &by, &at, &expires, &attempts); err != nil {
		t.Fatal(err)
	}
	if token != "" || by != "" {
		t.Errorf("claim_token=%q claimed_by=%q, want both empty — the migration issues no leases",
			token, by)
	}
	if at != nil || expires != nil {
		t.Errorf("claimed_at=%v claim_expires_at=%v, want both NULL", at, expires)
	}
	if attempts != 2 {
		t.Errorf("attempts = %d, want 2 — the migration must not rewrite a historical row", attempts)
	}

	// The point of all of the above: this row is still sendable. A sender
	// reaching it after the upgrade takes the lease, exactly as it would for a
	// row written after it.
	claim, err := j.ClaimAlertByID(context.Background(), id, testClaimant)
	if err != nil {
		t.Fatalf("ClaimAlertByID on a migrated row: %v", err)
	}
	if claim.Disposition != ClaimAcquired {
		t.Fatalf("disposition = %v on a migrated row, want acquired — the alert would be "+
			"stuck behind a lease nobody issued", claim.Disposition)
	}
	if claim.Stole {
		t.Error("stole = true on a migrated row — there was no previous holder to displace")
	}
}

package journal

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

// DefaultAlertLease is how long a delivery claim stays valid when Open is given
// no other value.
//
// It is not a number somebody liked. The lease has to outlast one sender's whole
// cycle, not one attempt: a sender claims once and then spends its entire retry
// budget under that claim. With today's defaults that cycle is bounded by
//
//	3 attempts × (10s publish timeout + 5s write wait) + 2 × 2s retry delay + 5s release wait = 54s
//
// and the lease is that bound times a 1.5 margin, rounded up: 81s. The bound is
// already the worst path — every attempt timing out, every write waiting the
// full busy timeout — so more margin buys little and costs directly: a longer
// lease is a longer time a *dead* sender's row stays locked, and the rows here
// are critical alerts, where late is worse than twice.
//
// journal cannot import obs (the dependency runs the other way), so this is a
// literal. The thing that stops it silently falling below the bound when the
// delivery configuration changes is the test that asserts
// DefaultAlertLease > obs.DefaultAlertDeliveryBound(). Copying 54 or 81 into
// that test would defeat it.
const DefaultAlertLease = 81 * time.Second

// alertClaimSkew is how far a stored issue time may sit in the future before the
// row is treated as written by a clock that has since moved backwards.
//
// The floor is two seconds: RFC3339 truncates to whole seconds, which alone can
// put a freshly stored claimed_at up to one second ahead of a later reading, and
// one second of headroom on top of that.
//
// Raising it makes the ledger *less* willing to reopen a row, not more: this is
// the "how far can I distrust the clock" knob, not a tolerance for stealing.
// Today both senders are the same process and therefore share one clock, so the
// real skew is zero and only the truncation matters.
const alertClaimSkew = 2 * time.Second

// ClaimDisposition is why a claim attempt ended the way it did. The three are not
// shades of failure: the caller does something different for each.
type ClaimDisposition int

const (
	// ClaimAcquired means send it. The result carries the token every
	// settlement of this send must present.
	ClaimAcquired ClaimDisposition = iota
	// ClaimSettled means there is nothing to send — delivered, or acknowledged
	// by an operator, and still inside the reminder window.
	ClaimSettled
	// ClaimHeldElsewhere means another sender holds a live lease on this row.
	// It is not an error. It is the whole point of this change.
	ClaimHeldElsewhere
)

func (d ClaimDisposition) String() string {
	switch d {
	case ClaimAcquired:
		return "acquired"
	case ClaimSettled:
		return "settled"
	case ClaimHeldElsewhere:
		return "held-elsewhere"
	default:
		return fmt.Sprintf("claim-disposition(%d)", int(d))
	}
}

// ClaimResult is the answer to a claim attempt, with the facts an observer needs
// to turn it into an event. The ledger has no logger and no event sink, so it
// returns facts and obs decides what to say about them.
type ClaimResult struct {
	Disposition ClaimDisposition
	ID          int64
	// Token is non-empty only for ClaimAcquired. It is the row's proof of
	// ownership and must never be logged: whoever reads it can settle somebody
	// else's send.
	Token string
	// ClaimedBy, ClaimedAt and ExpiresAt describe the lease that exists now —
	// mine when acquired, the other sender's when held elsewhere.
	ClaimedBy string
	ClaimedAt time.Time
	ExpiresAt time.Time

	// Stole and its two companions are filled only when acquiring took a lease
	// away from an earlier holder whose claim had expired. A steal is a normal,
	// designed recovery, but it is never silent: somebody died or stalled.
	Stole   bool
	StoleBy string
	StoleAt time.Time
}

// SettleOutcome is why a settlement affected the rows it did. Zero rows has four
// causes and the sender behaves differently for each, so the ledger reads the row
// once more, inside the same transaction, rather than returning a bare count.
type SettleOutcome int

const (
	// SettleApplied means the lease was mine and the write landed.
	SettleApplied SettleOutcome = iota
	// SettleLeaseLost means the row is still PENDING but carries a different
	// token. Somebody else owns this send now; stop sending immediately.
	SettleLeaseLost
	// SettleAlreadySettled means the row is no longer PENDING — delivered, or
	// acknowledged by an operator while this sender was publishing. It is not an
	// error: the operator's acknowledgement deliberately ignores leases.
	SettleAlreadySettled
	// SettleNotFound means there is no such row. That one *is* an error: the
	// ledger lost a row a sender was holding.
	SettleNotFound
)

func (o SettleOutcome) String() string {
	switch o {
	case SettleApplied:
		return "applied"
	case SettleLeaseLost:
		return "lease-lost"
	case SettleAlreadySettled:
		return "already-settled"
	case SettleNotFound:
		return "not-found"
	default:
		return fmt.Sprintf("settle-outcome(%d)", int(o))
	}
}

// SettleResult carries the outcome and, when the lease was lost, who holds it
// now. An enum alone cannot describe "you lost it to this holder, whose claim is
// this old", and that is the event the operator needs.
type SettleResult struct {
	Outcome   SettleOutcome
	ClaimedBy string
	ClaimedAt time.Time
	ExpiresAt time.Time
}

// alertClaimCleared is the one spelling of "this row holds no lease". It is
// written in exactly one place because the same four-column reset appears in
// three statements, and a copy of a value is how this change went wrong before.
const alertClaimCleared = `claim_token = '', claimed_by = '', claimed_at = NULL, claim_expires_at = NULL`

// alertClaimable is the predicate that decides whether a PENDING row can be
// claimed. It takes two bound values, both made by the ledger: the current time
// and that time plus the skew. The claimant's own lease is deliberately absent —
// expiry is read from what the issuer stored, so a claimant with a shorter lease
// cannot reinterpret a live claim and steal a row nobody abandoned.
//
// The two boundaries point in opposite directions on purpose. Expiry uses <=,
// which opens a shade early, and early here means "send again". The backwards
// clock test is strict, so that a claim issued this instant does not match its
// own rule.
const alertClaimable = `(claim_token = ''
	     OR claim_expires_at IS NULL
	     OR claim_expires_at <= ?
	     OR claimed_at > ?)`

// mintAlertClaimToken produces the value every settlement of this send will be
// checked against. It is random rather than derived from the holder's name
// because a loop that restarts under the same name would otherwise be able to
// settle the lease it held before it died.
func mintAlertClaimToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("journal: minting an alert claim token: %w", err)
	}
	return hex.EncodeToString(raw), nil
}

// alertLeaseRow is the lease as it stands on disk right now.
type alertLeaseRow struct {
	state     string
	token     string
	by        string
	at        sql.NullString
	expiresAt sql.NullString
}

func readAlertLeaseTx(ctx context.Context, tx *sql.Tx, id int64) (alertLeaseRow, error) {
	var row alertLeaseRow
	err := tx.QueryRowContext(ctx,
		`SELECT state, claim_token, claimed_by, claimed_at, claim_expires_at
		   FROM alert_outbox WHERE id = ?`, id).
		Scan(&row.state, &row.token, &row.by, &row.at, &row.expiresAt)
	return row, err
}

// claimStamp reads one of the two nullable lease timestamps. Absent, blank and
// unparsable all come back as the zero time, which every caller here reads as
// "there is no lease to speak of".
func claimStamp(v sql.NullString) time.Time {
	if ts := parseOptionalStamp(v); ts != nil {
		return *ts
	}
	return time.Time{}
}

// acquireAlertClaimTx takes the lease on one PENDING row, inside a transaction
// the caller already holds.
//
// The decision of whether the row is claimable is the UPDATE's alone. The row is
// read first only so that a steal can name the holder it displaced, and read
// again on zero rows only to explain a refusal. Deciding in Go as well would be a
// second copy of the predicate, and this change has already paid for copies.
func acquireAlertClaimTx(
	ctx context.Context, tx *sql.Tx, id int64, claimant string, now time.Time, lease time.Duration,
) (ClaimResult, error) {
	before, err := readAlertLeaseTx(ctx, tx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ClaimResult{ID: id}, fmt.Errorf("%w: %d", ErrAlertNotFound, id)
		}
		return ClaimResult{ID: id}, fmt.Errorf("journal: reading the claim on alert %d: %w", id, err)
	}

	token, err := mintAlertClaimToken()
	if err != nil {
		return ClaimResult{ID: id}, err
	}
	nowStamp := RFC3339(now)
	expires := now.Add(lease)
	res, err := tx.ExecContext(ctx,
		`UPDATE alert_outbox
		    SET claim_token = ?, claimed_by = ?, claimed_at = ?, claim_expires_at = ?
		  WHERE id = ? AND state = ? AND `+alertClaimable,
		token, strings.TrimSpace(claimant), nowStamp, RFC3339(expires),
		id, AlertPending, nowStamp, RFC3339(now.Add(alertClaimSkew)))
	if err != nil {
		return ClaimResult{ID: id}, fmt.Errorf("journal: claiming alert %d: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return ClaimResult{ID: id}, fmt.Errorf("journal: claiming alert %d: %w", id, err)
	}
	if n == 1 {
		out := ClaimResult{
			Disposition: ClaimAcquired,
			ID:          id,
			Token:       token,
			ClaimedBy:   strings.TrimSpace(claimant),
			ClaimedAt:   now.UTC().Truncate(time.Second),
			ExpiresAt:   expires.UTC().Truncate(time.Second),
		}
		if before.token != "" {
			out.Stole = true
			out.StoleBy = before.by
			out.StoleAt = claimStamp(before.at)
		}
		return out, nil
	}

	// Zero rows. The row was read before the UPDATE, in this transaction, so
	// that read is what it looked like when the UPDATE ran.
	if before.state != AlertPending {
		return ClaimResult{Disposition: ClaimSettled, ID: id}, nil
	}
	return ClaimResult{
		Disposition: ClaimHeldElsewhere,
		ID:          id,
		ClaimedBy:   before.by,
		ClaimedAt:   claimStamp(before.at),
		ExpiresAt:   claimStamp(before.expiresAt),
	}, nil
}

// ClaimAlertByID takes the delivery lease on a row whose id the caller already
// has.
//
// It exists because a lease with one bypass around it is not a lease. Flush read
// PendingAlerts and published straight from the list; every row it sent was a row
// no claim covered.
//
// It takes no lease duration. The duration belongs to the ledger instance that
// issues the claim, and letting a caller pass one is how a second sender ends up
// judging somebody else's claim by its own clock.
func (j *Journal) ClaimAlertByID(ctx context.Context, id int64, claimant string) (ClaimResult, error) {
	if strings.TrimSpace(claimant) == "" {
		return ClaimResult{ID: id}, errors.New("journal: claiming an alert requires the claimant's name")
	}
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return ClaimResult{ID: id}, fmt.Errorf("journal: claiming alert %d: %w", id, err)
	}
	defer tx.Rollback()

	out, err := acquireAlertClaimTx(ctx, tx, id, claimant, j.clk.Now(), j.alertLease)
	if err != nil {
		return out, err
	}
	if err := tx.Commit(); err != nil {
		return ClaimResult{ID: id}, fmt.Errorf("journal: committing the claim on alert %d: %w", id, err)
	}
	return out, nil
}

// settleUnderClaim runs one settlement statement under the caller's token and
// says why it affected no rows when it affected none.
//
// stmt must end in the CAS the four outcomes depend on — id, PENDING, token —
// and args must match it.
func (j *Journal) settleUnderClaim(
	ctx context.Context, id int64, token, what, stmt string, args []any,
) (SettleResult, error) {
	if strings.TrimSpace(token) == "" {
		return SettleResult{}, fmt.Errorf("journal: %s alert %d requires the claim token the send was made under", what, id)
	}
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return SettleResult{}, fmt.Errorf("journal: %s alert %d: %w", what, id, err)
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, stmt, args...)
	if err != nil {
		return SettleResult{}, fmt.Errorf("journal: %s alert %d: %w", what, id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return SettleResult{}, fmt.Errorf("journal: %s alert %d: %w", what, id, err)
	}
	if n == 1 {
		if err := tx.Commit(); err != nil {
			return SettleResult{}, fmt.Errorf("journal: committing %s of alert %d: %w", what, id, err)
		}
		return SettleResult{Outcome: SettleApplied}, nil
	}

	out, err := explainSettleTx(ctx, tx, id)
	if err != nil {
		return SettleResult{}, fmt.Errorf("journal: %s alert %d: %w", what, id, err)
	}
	if err := tx.Commit(); err != nil {
		return SettleResult{}, fmt.Errorf("journal: committing %s of alert %d: %w", what, id, err)
	}
	return out, nil
}

// explainSettleTx reads the row once more, in the transaction that just wrote
// nothing, and turns "zero rows" into the reason the sender has to act on.
func explainSettleTx(ctx context.Context, tx *sql.Tx, id int64) (SettleResult, error) {
	row, err := readAlertLeaseTx(ctx, tx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return SettleResult{Outcome: SettleNotFound}, nil
		}
		return SettleResult{}, err
	}
	if row.state != AlertPending {
		return SettleResult{Outcome: SettleAlreadySettled}, nil
	}
	return SettleResult{
		Outcome:   SettleLeaseLost,
		ClaimedBy: row.by,
		ClaimedAt: claimStamp(row.at),
		ExpiresAt: claimStamp(row.expiresAt),
	}, nil
}

// ReleaseAlertClaim hands the lease back without recording an outcome.
//
// It is for the sender that gives up: the retry budget is spent, or Flush has
// finished with the row. It is deliberately *not* called after a publish that
// succeeded but whose settlement failed — that lease is the only thing left
// suppressing a re-send of an alert the operator already has, and letting go of
// it brings back the storm this family of changes exists to stop.
func (j *Journal) ReleaseAlertClaim(ctx context.Context, id int64, token string) (SettleResult, error) {
	return j.settleUnderClaim(ctx, id, token, "releasing the claim on",
		`UPDATE alert_outbox SET `+alertClaimCleared+`
		  WHERE id = ? AND state = ? AND claim_token = ?`,
		[]any{id, AlertPending, token})
}

// AlertLease is the delivery-claim duration this instance issues.
//
// It is readable because the safety property it carries is not this package's to
// check: the lease has to outlast the *sender's* whole retry budget, and that
// budget lives in obs, which this package cannot import. Exposing the value lets
// the side that knows the budget compare the two and say so when they disagree.
func (j *Journal) AlertLease() time.Duration { return j.alertLease }

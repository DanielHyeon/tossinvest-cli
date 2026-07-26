package journal

// adoption.go is schema v7 and the whole of the write path that uses it (change
// adopt-external-positions, design A1 — that DDL is normative and this file is
// its transcription).
//
// # What is being recorded, and why it is not a decision
//
// A position the engine did not open has no entry decision, and until this
// change that was the end of the story: no decision meant no stop, no stop meant
// no t0 baseline, and the exit policy left the shares alone (position-ledger,
// exit-policy). The user's 2026-07-26 decision reverses the conclusion and not
// the premise — the position still has no decision, so what is supplied instead
// is a *record of the adoption itself*: the price observed at the instant of
// adoption, the stop synthesised from it, the quantity and cost basis as the
// account reported them, and when that observation was taken.
//
// It is deliberately not a row in `decisions`. `decisions.safety_class` and
// `preimage_kind` are SQL CHECK enums, the additive rules forbid rewriting a
// table, and — more importantly — an adoption authorises no mutation. The
// three-class contract engine-safety and risk-management are written against
// stays exactly as it was (design A1).
//
// # set-once, and why it is enforced twice
//
// `positions.adoption_id` is written by [Journal.AdoptPosition] and by nothing
// else. Two mechanisms say so:
//
//	the statement  the UPDATE carries `WHERE adoption_id IS NULL`, so a second
//	               write affects no row and is reported as ErrPositionAlreadyAdopted
//	               rather than silently repointing a live position at a new
//	               baseline;
//	the layout     adoption_static_test.go fails if any other production file in
//	               this package writes the column at all — the same arrangement
//	               apply_hook.go's guarded four use, for the same reason: a second
//	               writer must not be addable without someone editing a list and
//	               justifying it in review.
//
// `entry_decision_id` is read here and never written. That column's immutability
// is a landed SHALL NOT (position-ledger) and this change does not become its
// first mutator.
//
// # Broker-behaviour claims
//
// One, tagged. `cost_basis` is the broker's `averagePurchasePrice` verbatim —
// whether it is inclusive of fees is `[미측정 — 2b 실측 대상]`, which is exactly
// why it is stored as the raw string and excluded from every R formula (design
// A7). `observed_price` comes through the engine's ordinary price path and is
// therefore `[기존 제약 — 엔진 가격 경로 전체가 float64(Quote.Last)]`.

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

// schemaV7 is the adoption record. Additive per schema.go's rules: one new
// table and one nullable ADD COLUMN with no DEFAULT, no released column changed
// and no historical row rewritten. There is no down-migration; an older binary
// refuses the newer user_version (ErrSchemaTooNew) and recovery is the
// pre-migration backup taken by backup.go.
//
// The reference is one-directional. `position_adoptions` carries no position id:
// a mutual pair of foreign keys cannot be inserted in either order without
// deferring one of them, and the projection is the authority for symbol, market
// and quantity anyway — the columns here are the snapshot as of the adoption
// instant, kept because the synthetic t0 has to remain explicable after the
// position has moved.
const schemaV7 = `
CREATE TABLE position_adoptions (
	id              TEXT PRIMARY KEY,
	symbol          TEXT NOT NULL,             -- snapshot at adoption; authority is positions
	market          TEXT NOT NULL,             -- 〃
	quantity        TEXT NOT NULL,             -- 〃 decimal string
	-- The broker's averagePurchasePrice, stored verbatim. Nullable because a
	-- broker that reports none is a fact, not a zero.
	cost_basis      TEXT,
	cost_basis_src  TEXT NOT NULL,             -- 'BROKER_AVG' | 'ABSENT'
	observed_price  TEXT NOT NULL,             -- t0 EntryPrice
	synthetic_stop  TEXT NOT NULL,             -- observed_price × (1 − default_stop_pct)
	observed_at     TEXT NOT NULL,
	preimage_digest TEXT NOT NULL
) STRICT;

ALTER TABLE positions ADD COLUMN adoption_id TEXT REFERENCES position_adoptions(id);

CREATE INDEX idx_positions_adoption ON positions(adoption_id);
`

// Cost-basis sources. Two, because "the broker said 70000" and "the broker said
// nothing" are different facts and a stored empty string cannot be told from a
// price nobody looked for.
const (
	CostBasisBrokerAvg = "BROKER_AVG"
	CostBasisAbsent    = "ABSENT"
)

// ErrPositionAlreadyAdopted means the position already carries an adoption. The
// reference is set once for the life of an instance: repointing it would move a
// live position's synthetic t0, which is the baseline its stop is measured from.
var ErrPositionAlreadyAdopted = errors.New("journal: the position already carries an adoption record")

// ErrPositionNotAdoptable means the position carries an entry decision, so it is
// already an exit-policy target and adopting it would give one position two
// justifications for its baseline.
var ErrPositionNotAdoptable = errors.New(
	"journal: the position was opened by an entry decision, so there is nothing to adopt")

// ErrAdoptionNotFound means no position_adoptions row matches.
var ErrAdoptionNotFound = errors.New("journal: no such position adoption")

// AdoptionRequest is one adoption, as the caller computed it.
//
// Every price in it was decided outside this package: the observation is the
// engine's ordinary price read taken immediately before this call, and the
// synthetic stop is internal/exitpolicy's arithmetic over it. The journal
// records the pair and refuses an unusable one; it does not invent either.
type AdoptionRequest struct {
	// PositionID is the instance being adopted. It must exist, carry no entry
	// decision and carry no adoption yet.
	PositionID string

	Symbol string
	Market string
	// Quantity is the account's own, as of the observation.
	Quantity string

	// CostBasis is the broker's average purchase price, as the *raw decimal
	// string it reported* — never a re-rendered float. Empty means the broker
	// reported none, which is recorded as ABSENT.
	CostBasis string

	// ObservedPrice is t0: the price observed immediately before this call.
	// SyntheticStop is `ObservedPrice × (1 − pct)`.
	ObservedPrice string
	SyntheticStop string
	// ObservedAt is when the observation was taken, RFC3339 UTC.
	ObservedAt string
}

// PositionAdoption is one stored adoption record.
type PositionAdoption struct {
	ID              string
	Symbol          string
	Market          string
	Quantity        string
	CostBasis       string
	CostBasisSource string
	ObservedPrice   string
	SyntheticStop   string
	ObservedAt      string
	PreimageDigest  string
}

// AdoptPosition records an adoption and points the position at it, in one
// transaction.
//
// The order inside the transaction is the one the one-directional reference
// forces: the adoption row first, then the position's pointer at it. Both or
// neither — a position pointing at a record that does not exist would fail the
// foreign key, and a record nothing points at would be an adoption that never
// happened.
//
// It issues no sell proposal and reads no price. That is the SHALL NOT design A2
// puts on the adoption transaction: the act of adopting protects a position, and
// a transaction that could also decide to liquidate it would make "adoption
// never sells" a property of the caller's discipline.
func (j *Journal) AdoptPosition(ctx context.Context, req AdoptionRequest) (PositionAdoption, error) {
	adoption, err := req.record()
	if err != nil {
		return PositionAdoption{}, err
	}

	tx, err := j.db.BeginTx(ctx, nil) // BEGIN IMMEDIATE
	if err != nil {
		return PositionAdoption{}, fmt.Errorf(
			"journal: adopting position %s: %w", req.PositionID, err)
	}
	defer tx.Rollback()

	positionID := strings.TrimSpace(req.PositionID)
	var (
		decisionID sql.NullString
		existing   sql.NullString
	)
	err = tx.QueryRowContext(ctx,
		`SELECT entry_decision_id, adoption_id FROM positions WHERE id = ?`, positionID).
		Scan(&decisionID, &existing)
	if errors.Is(err, sql.ErrNoRows) {
		return PositionAdoption{}, fmt.Errorf("%w: %s", ErrPositionNotFound, positionID)
	}
	if err != nil {
		return PositionAdoption{}, fmt.Errorf("journal: reading position %s: %w", positionID, err)
	}
	if strings.TrimSpace(decisionID.String) != "" {
		return PositionAdoption{}, fmt.Errorf("%w: %s", ErrPositionNotAdoptable, positionID)
	}
	if stored := strings.TrimSpace(existing.String); stored != "" {
		// Idempotent recovery: a crash between the commit and whatever the caller
		// did next re-runs this call with the same derived id, and the honest
		// answer is the record that is already on disk. A *different* id is a
		// second adoption of one instance, which is what set-once forbids.
		if stored == adoption.ID {
			return readAdoptionTx(ctx, tx, stored)
		}
		return PositionAdoption{}, fmt.Errorf("%w: %s already carries %s",
			ErrPositionAlreadyAdopted, positionID, stored)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO position_adoptions
		  (id, symbol, market, quantity, cost_basis, cost_basis_src, observed_price,
		   synthetic_stop, observed_at, preimage_digest)
		VALUES (?,?,?,?,?,?,?,?,?,?)`,
		adoption.ID, adoption.Symbol, adoption.Market, adoption.Quantity,
		nullableString(adoption.CostBasis), adoption.CostBasisSource, adoption.ObservedPrice,
		adoption.SyntheticStop, adoption.ObservedAt, adoption.PreimageDigest); err != nil {
		if isUniqueViolation(err) {
			return PositionAdoption{}, fmt.Errorf("%w: %s", ErrPositionAlreadyAdopted, adoption.ID)
		}
		return PositionAdoption{}, fmt.Errorf(
			"journal: recording the adoption of %s: %w", positionID, err)
	}

	// The one write of `positions.adoption_id` in the whole build. The predicate
	// is the set-once guarantee expressed as a statement rather than as a check
	// the caller performs: even a caller that raced past the read above cannot
	// repoint a position that already carries one.
	result, err := tx.ExecContext(ctx,
		`UPDATE positions SET adoption_id = ? WHERE id = ? AND adoption_id IS NULL`,
		adoption.ID, positionID)
	if err != nil {
		return PositionAdoption{}, fmt.Errorf(
			"journal: pointing position %s at its adoption: %w", positionID, err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return PositionAdoption{}, fmt.Errorf(
			"journal: pointing position %s at its adoption: %w", positionID, err)
	}
	if affected != 1 {
		return PositionAdoption{}, fmt.Errorf("%w: %s", ErrPositionAlreadyAdopted, positionID)
	}

	if err := tx.Commit(); err != nil {
		return PositionAdoption{}, fmt.Errorf(
			"journal: committing the adoption of %s: %w", positionID, err)
	}
	return adoption, nil
}

// OpenAdoptedExitState opens the protection state of an adopted position from
// its own record.
//
// It is the second arm of "the exit state opening path accepts both sources"
// (exit-policy: exit_state 개설 경로는 두 출처를 모두 수용해야 한다 SHALL). An
// adopted position has no entry decision, so the engine's ordinary path — look
// the decision up, parse its RiskIntent, take its limit and stop — cannot run at
// all: there is nothing to look up. What replaces it is not a fallback but a
// different, equally explicit source:
//
//	EntryPrice  = observed_price   the observation the adoption was judged on
//	InitialStop = synthetic_stop   observed_price × (1 − pct), frozen at adoption
//
// The high-water mark takes no argument here and never did: OpenRatchetState
// seeds it from the entry price, so an adopted position opens at R = 0 with its
// watermark exactly at t0, the same as an engine-entered one.
func (j *Journal) OpenAdoptedExitState(ctx context.Context, positionID string) (ExitState, error) {
	adoption, err := j.AdoptionOf(ctx, positionID)
	if err != nil {
		return ExitState{}, err
	}
	return j.OpenExitState(ctx, ExitStateSeed{
		PositionID:  strings.TrimSpace(positionID),
		PolicyKind:  ExitPolicyRatchet,
		EntryPrice:  adoption.ObservedPrice,
		InitialStop: adoption.SyntheticStop,
	})
}

// AdoptionOf returns the adoption record a position carries.
func (j *Journal) AdoptionOf(ctx context.Context, positionID string) (PositionAdoption, error) {
	rows, err := j.db.QueryContext(ctx, adoptionSelect+`
		  JOIN positions p ON p.adoption_id = a.id
		 WHERE p.id = ?`, strings.TrimSpace(positionID))
	if err != nil {
		return PositionAdoption{}, fmt.Errorf(
			"journal: reading the adoption of %s: %w", positionID, err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return PositionAdoption{}, fmt.Errorf(
				"journal: reading the adoption of %s: %w", positionID, err)
		}
		return PositionAdoption{}, fmt.Errorf("%w: position %s", ErrAdoptionNotFound, positionID)
	}
	return scanAdoption(rows)
}

// Adoption returns one adoption record by id.
func (j *Journal) Adoption(ctx context.Context, id string) (PositionAdoption, error) {
	rows, err := j.db.QueryContext(ctx, adoptionSelect+" WHERE a.id = ?", strings.TrimSpace(id))
	if err != nil {
		return PositionAdoption{}, fmt.Errorf("journal: reading adoption %s: %w", id, err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return PositionAdoption{}, fmt.Errorf("journal: reading adoption %s: %w", id, err)
		}
		return PositionAdoption{}, fmt.Errorf("%w: %s", ErrAdoptionNotFound, id)
	}
	return scanAdoption(rows)
}

const adoptionSelect = `SELECT a.id, a.symbol, a.market, a.quantity, coalesce(a.cost_basis, ''),
	a.cost_basis_src, a.observed_price, a.synthetic_stop, a.observed_at, a.preimage_digest
	FROM position_adoptions a`

func readAdoptionTx(ctx context.Context, tx *sql.Tx, id string) (PositionAdoption, error) {
	rows, err := tx.QueryContext(ctx, adoptionSelect+" WHERE a.id = ?", id)
	if err != nil {
		return PositionAdoption{}, fmt.Errorf("journal: reading adoption %s: %w", id, err)
	}
	defer rows.Close()
	if !rows.Next() {
		return PositionAdoption{}, fmt.Errorf("%w: %s", ErrAdoptionNotFound, id)
	}
	return scanAdoption(rows)
}

func scanAdoption(rows *sql.Rows) (PositionAdoption, error) {
	var a PositionAdoption
	if err := rows.Scan(&a.ID, &a.Symbol, &a.Market, &a.Quantity, &a.CostBasis,
		&a.CostBasisSource, &a.ObservedPrice, &a.SyntheticStop, &a.ObservedAt,
		&a.PreimageDigest); err != nil {
		return PositionAdoption{}, fmt.Errorf("journal: reading an adoption record: %w", err)
	}
	return a, nil
}

// record validates the request and derives the stored row.
func (r AdoptionRequest) record() (PositionAdoption, error) {
	a := PositionAdoption{
		Symbol:        normaliseSymbol(r.Symbol),
		Market:        normaliseMarket(r.Market),
		CostBasis:     strings.TrimSpace(r.CostBasis),
		ObservedPrice: strings.TrimSpace(r.ObservedPrice),
		SyntheticStop: strings.TrimSpace(r.SyntheticStop),
		ObservedAt:    strings.TrimSpace(r.ObservedAt),
	}
	switch {
	case strings.TrimSpace(r.PositionID) == "":
		return PositionAdoption{}, fmt.Errorf(
			"%w: an adoption is about one position instance; none was named", ErrInvalidRequest)
	case a.Symbol == "" || a.Market == "":
		return PositionAdoption{}, fmt.Errorf(
			"%w: an adoption record carries the symbol and market it was taken on", ErrInvalidRequest)
	case a.ObservedAt == "":
		return PositionAdoption{}, fmt.Errorf(
			"%w: an adoption records when its observation was taken; a synthetic stop with no "+
				"observation instant cannot be audited", ErrInvalidRequest)
	}

	var err error
	if a.Quantity, err = canonicalQuantity("adopted quantity", orZero(r.Quantity)); err != nil {
		return PositionAdoption{}, err
	}
	if isZeroDecimal(a.Quantity) {
		return PositionAdoption{}, fmt.Errorf(
			"%w: adopting a holding of zero would open a protection state over nothing", ErrInvalidRequest)
	}
	if a.ObservedPrice, err = positivePrice("observed price", a.ObservedPrice); err != nil {
		return PositionAdoption{}, err
	}
	if a.SyntheticStop, err = positivePrice("synthetic stop", a.SyntheticStop); err != nil {
		return PositionAdoption{}, err
	}
	// The pair has to be a usable t0 before it is stored, not when the exit state
	// is opened: a stored stop at or above the observation is a row whose only
	// possible reading is "liquidate now", and it would be discovered by the
	// first observation rather than by the write that made it.
	cmp, cerr := riskcalc.CompareDecimal(a.SyntheticStop, a.ObservedPrice)
	if cerr != nil {
		return PositionAdoption{}, fmt.Errorf("%w: comparing the adoption's prices: %v",
			ErrInvalidRequest, cerr)
	}
	if cmp >= 0 {
		return PositionAdoption{}, fmt.Errorf(
			"%w: the synthetic stop %s is not below the observed price %s, so the adopted position "+
				"would have no risk per unit", ErrInvalidRequest, a.SyntheticStop, a.ObservedPrice)
	}

	a.CostBasisSource = CostBasisAbsent
	if a.CostBasis != "" {
		a.CostBasisSource = CostBasisBrokerAvg
	}

	a.PreimageDigest = adoptionDigest(strings.TrimSpace(r.PositionID), a)
	// Derived rather than minted, the same rule adjustmentID follows: a retry
	// after a crash carries the same id, which is what makes the set-once check
	// able to recognise its own previous commit instead of refusing it.
	a.ID = "adopt-" + a.PreimageDigest[:24]
	return a, nil
}

// positivePrice canonicalises a price and refuses one that is not positive.
//
// A zero or negative price is not a cheap position: it is an unusable
// observation, and storing one would freeze a t0 whose R denominator is
// meaningless for the life of the instance.
func positivePrice(field, value string) (string, error) {
	canonical, err := canonicalQuantity(field, value)
	if err != nil {
		return "", err
	}
	cmp, err := riskcalc.CompareDecimal(canonical, "0")
	if err != nil {
		return "", fmt.Errorf("%w: %s %q: %v", ErrInvalidRequest, field, value, err)
	}
	if cmp <= 0 {
		return "", fmt.Errorf("%w: %s is %s; an adoption needs a positive price",
			ErrInvalidRequest, field, canonical)
	}
	return canonical, nil
}

// adoptionDigest is the canonical hash of what justified the adoption.
//
// It is the same shape as the decision preimage's: a length-prefixed rendering
// so ("ab","c") and ("a","bc") cannot collide, over exactly the fields that make
// the synthetic t0 what it is. It is stored so that an operator — or a later
// audit — can prove the baseline in `exit_states` came from this observation and
// not from a value edited afterwards.
func adoptionDigest(positionID string, a PositionAdoption) string {
	h := sha256.New()
	for _, part := range []string{
		positionID, a.Symbol, a.Market, a.Quantity, a.CostBasis, a.CostBasisSource,
		a.ObservedPrice, a.SyntheticStop, a.ObservedAt,
	} {
		fmt.Fprintf(h, "%d:%s|", len(part), part)
	}
	return hex.EncodeToString(h.Sum(nil))
}

package journal

// entryobservation.go is schema v8 and the write path that uses it (change
// add-net-rr-measurement, design D1 — that DDL is normative and this file is its
// transcription).
//
// # What is being recorded, and why it is not a decision
//
// Today half of every Guardian entry verdict leaves no trace. `decisions` holds
// the ones that were *issued*; a refusal is an IssueRefusal built in memory and
// returned, and nothing writes it down. That makes the one question a threshold
// has to answer — "what would this gate have refused?" — unanswerable from disk,
// which is the gap this table closes.
//
// It is deliberately not a column on `decisions`. That table is the contract the
// gateway re-verifies against: its preimage is hashed, its shape is what a
// tampered submission is checked against, and letting an analysis requirement
// grow its surface is how a contract stops being one. So the observation is its
// own table, and the two are joined by nothing the database enforces.
//
// # Self-contained, and why the duplication is the point
//
// The three prices live here as well as in the decision preimage. They are
// written from the same input at the same instant, so they cannot disagree, and
// the copy is what makes two things possible that a join would not: an
// observation for a *refusal*, which has no decision to join to, and an analysis
// query that never touches a contract table (trade-analytics: 분석 경로의 격리).
//
// # decision_id is a reference, not a constraint
//
// No foreign key, for the reason spent_nonces.decision_id has none: a decision is
// pruned when it expires, and that pruning must not be blocked by — or cascade
// into — a row that exists to be read long afterwards. The uniqueness is enforced
// (one observation per issued decision) and the existence is not.
//
// # Broker-behaviour claims
//
// None. Every number here is computed from values the caller already had.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// schemaV8 is the entry observation table. Additive per schema.go's rules: one
// new table and three new indexes, no released column changed, no historical row
// rewritten, and nothing added to `decisions`. There is no down-migration; an
// older binary refuses the newer user_version (ErrSchemaTooNew) and recovery is
// the pre-migration backup taken by backup.go.
const schemaV8 = `
CREATE TABLE entry_decision_observations (
	id                     TEXT PRIMARY KEY,
	account_ref            TEXT NOT NULL,
	market                 TEXT NOT NULL,
	symbol                 TEXT NOT NULL,

	-- The geometry, as decimal strings like everywhere else in the journal.
	-- stop and target are nullable because an intent refused for not carrying
	-- one is exactly the observation worth keeping, and a stored '' could not be
	-- told from a price nobody supplied.
	entry_price            TEXT NOT NULL,
	stop_price             TEXT,
	target_price           TEXT,

	-- 실질본전 and the two ratios. All three are nullable: the chain can stop
	-- before they exist, and a NULL says "not computed" where a 0 would say
	-- "computed, and it was nothing" (risk-management: 0 대체 금지).
	break_even_price       TEXT,
	gross_reward_risk      TEXT,
	net_reward_risk        TEXT,

	-- What the two ratios are ratios *of*. Without these a number computed from
	-- [미검증] placeholder rates reads exactly like a measured one.
	cost_scope             TEXT NOT NULL,
	cost_model_fingerprint TEXT NOT NULL,

	-- Which rung stopped it, reported by the verdict rather than inferred from
	-- the reason: one reason code is raised at 42 sites in internal/risk, so the
	-- inverse map is many-to-one and would be quietly wrong.
	stopped_step           TEXT NOT NULL DEFAULT '',
	reason_code            TEXT NOT NULL DEFAULT '',

	-- Chain ALLOW and issuance success are different facts. A null decision_id
	-- cannot tell the three cases apart, so the case is stored.
	outcome                TEXT NOT NULL,
	issuance_reason_code   TEXT NOT NULL DEFAULT '',

	reconstructed          INTEGER NOT NULL DEFAULT 0,
	observed_at            TEXT NOT NULL,
	-- Set on issued rows, from decisions.issued_at. A reconstructed row carries
	-- both this and reconstructed_at: one row must not mix one clock's instant
	-- with another row's fingerprint.
	issued_at              TEXT,
	reconstructed_at       TEXT,

	-- Nullable, and deliberately not a foreign key. See the file header.
	decision_id            TEXT,

	CHECK (outcome IN ('REFUSED_CHAIN','ALLOWED_ISSUED','ALLOWED_ISSUANCE_REFUSED')),
	CHECK (cost_scope IN ('FEE_TAX_ONLY')),
	CHECK (reconstructed IN (0,1)),
	CHECK (reconstructed = 0 OR reconstructed_at IS NOT NULL)
) STRICT;

-- Unique, not foreign. It is what makes a late-landing write and a
-- reconstruction of the same decision collapse into one row rather than two:
-- a double-counted entry would bias the very distribution a live threshold is
-- going to be drawn from.
CREATE UNIQUE INDEX idx_entry_observations_decision
	ON entry_decision_observations(decision_id) WHERE decision_id IS NOT NULL;

CREATE INDEX idx_entry_observations_observed ON entry_decision_observations(observed_at);
CREATE INDEX idx_entry_observations_outcome  ON entry_decision_observations(outcome);
`

// Outcome classes. Three, because "the chain refused", "the chain allowed and the
// ledger issued" and "the chain allowed and the issuance was refused" are three
// different facts and only the middle one leaves a decision row behind.
const (
	// OutcomeRefusedChain: a rung of the entry chain refused. There is no
	// decision, and there never was one.
	OutcomeRefusedChain = "REFUSED_CHAIN"
	// OutcomeAllowedIssued: the chain allowed and the atomic issuance committed.
	OutcomeAllowedIssued = "ALLOWED_ISSUED"
	// OutcomeAllowedIssuanceRefused: the chain allowed and the issuance
	// transaction refused (LIMIT_REACHED, DECISION_EXPIRED, VERSION_CONFLICT,
	// SNAPSHOT_RECOLLECTION_EXHAUSTED). Recording this as a chain refusal would
	// attribute a ledger answer to the intent's geometry.
	OutcomeAllowedIssuanceRefused = "ALLOWED_ISSUANCE_REFUSED"
)

// CostScopeFeeTaxOnly is the only cost scope this change measures: commission and
// tax on both legs, no slippage. The metric's name follows from it — 수수료·세금
// 차감 후 RR, not "cost-adjusted".
const CostScopeFeeTaxOnly = "FEE_TAX_ONLY"

// ErrObservationExists means an observation is already recorded for that
// decision. It is the unique index answering, and it is not an error the trading
// path reacts to: the reconstruction job uses it to skip a decision whose real
// write landed late.
var ErrObservationExists = errors.New("journal: an observation is already recorded for that decision")

// EntryObservation is one entry verdict, as measured.
//
// Every ratio field is a decimal string and empty means "not computed". That is
// the same convention the geometry uses and it carries the same rule: an empty
// value is unknown, never zero.
type EntryObservation struct {
	// ID is the caller's own identifier, minted the way attempt ids are.
	ID string
	// AccountRef, Market, Symbol identify what was judged.
	AccountRef, Market, Symbol string
	// EntryPrice is required. StopPrice and TargetPrice are empty when the
	// refused intent carried none.
	EntryPrice, StopPrice, TargetPrice string
	// BreakEvenPrice is 실질본전 — the same BreakEvenSellPrice(entry, "1", market)
	// the stop-contract rung compared the target against. Empty when the chain
	// stopped before it, or when computing it failed.
	BreakEvenPrice string
	// GrossRewardRisk is (target − entry) / (entry − stop): the ratio the gate
	// judges. NetRewardRisk is (target − B) / (B − stop): the observation this
	// change adds. Neither is a gate input in the other's place.
	GrossRewardRisk, NetRewardRisk string
	// CostScope is CostScopeFeeTaxOnly. CostModelFingerprint identifies the rate
	// set the two ratios were computed under.
	CostScope, CostModelFingerprint string
	// StoppedStep is the entryChain rung name that refused, empty when allowed.
	// ReasonCode is that rung's code.
	StoppedStep, ReasonCode string
	// Outcome is one of the three classes above.
	Outcome string
	// IssuanceReasonCode is the stable issuance code, set only for
	// OutcomeAllowedIssuanceRefused.
	IssuanceReasonCode string
	// Reconstructed marks a row rebuilt from a decision preimage rather than
	// written by the verdict. ReconstructedAt is then required.
	Reconstructed bool
	// ObservedAt is when the verdict happened. IssuedAt mirrors
	// decisions.issued_at on issued rows.
	ObservedAt, IssuedAt, ReconstructedAt time.Time
	// DecisionID references the decision, for issued rows only. Not a foreign
	// key — see the file header.
	DecisionID string
}

// RecordEntryObservation persists one observation.
//
// It is a single statement outside every other transaction, and that placement is
// the requirement rather than an implementation detail (risk-management: 관측
// 기록은 결정·예약의 원자 트랜잭션 밖에서 수행된다). Inside, a failed insert would
// roll the decision back with it and a measurement defect would have refused a
// trade the chain allowed.
//
// The error it returns is for counting, not for deciding: no caller may turn it
// into a refusal.
func (j *Journal) RecordEntryObservation(ctx context.Context, obs EntryObservation) error {
	row, err := obs.validate()
	if err != nil {
		return err
	}
	_, err = j.db.ExecContext(ctx,
		`INSERT INTO entry_decision_observations
		   (id, account_ref, market, symbol, entry_price, stop_price, target_price,
		    break_even_price, gross_reward_risk, net_reward_risk, cost_scope,
		    cost_model_fingerprint, stopped_step, reason_code, outcome,
		    issuance_reason_code, reconstructed, observed_at, issued_at,
		    reconstructed_at, decision_id)
		 VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		row.ID, row.AccountRef, row.Market, row.Symbol, row.EntryPrice,
		nullableString(row.StopPrice), nullableString(row.TargetPrice),
		nullableString(row.BreakEvenPrice), nullableString(row.GrossRewardRisk),
		nullableString(row.NetRewardRisk), row.CostScope, row.CostModelFingerprint,
		row.StoppedStep, row.ReasonCode, row.Outcome, row.IssuanceReasonCode,
		boolToInt(row.Reconstructed), formatJournalTime(row.ObservedAt),
		nullableTime(row.IssuedAt), nullableTime(row.ReconstructedAt),
		nullableString(row.DecisionID))
	if err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("%w: decision %s", ErrObservationExists, row.DecisionID)
		}
		return fmt.Errorf("journal: recording entry observation %s: %w", row.ID, err)
	}
	return nil
}

// validate normalises the row and refuses one that could not be read back as the
// fact it claims to be.
func (o EntryObservation) validate() (EntryObservation, error) {
	invalid := func(format string, args ...any) (EntryObservation, error) {
		return EntryObservation{}, fmt.Errorf("%w: %s", ErrInvalidRequest, fmt.Sprintf(format, args...))
	}
	out := o
	out.ID = strings.TrimSpace(o.ID)
	out.AccountRef = strings.TrimSpace(o.AccountRef)
	out.Market = strings.ToLower(strings.TrimSpace(o.Market))
	out.Symbol = strings.ToUpper(strings.TrimSpace(o.Symbol))
	out.EntryPrice = strings.TrimSpace(o.EntryPrice)
	out.StopPrice = strings.TrimSpace(o.StopPrice)
	out.TargetPrice = strings.TrimSpace(o.TargetPrice)
	out.BreakEvenPrice = strings.TrimSpace(o.BreakEvenPrice)
	out.GrossRewardRisk = strings.TrimSpace(o.GrossRewardRisk)
	out.NetRewardRisk = strings.TrimSpace(o.NetRewardRisk)
	out.CostScope = strings.TrimSpace(o.CostScope)
	out.CostModelFingerprint = strings.TrimSpace(o.CostModelFingerprint)
	out.StoppedStep = strings.TrimSpace(o.StoppedStep)
	out.ReasonCode = strings.TrimSpace(o.ReasonCode)
	out.Outcome = strings.TrimSpace(o.Outcome)
	out.IssuanceReasonCode = strings.TrimSpace(o.IssuanceReasonCode)
	out.DecisionID = strings.TrimSpace(o.DecisionID)

	switch {
	case out.ID == "":
		return invalid("an observation id is required")
	case out.AccountRef == "":
		return invalid("an observation names the account it measured")
	case out.Market == "":
		return invalid("an observation names the market it measured")
	case out.Symbol == "":
		return invalid("an observation names the symbol it measured")
	case out.EntryPrice == "":
		return invalid("an observation carries the entry price; there is nothing to measure without one")
	case out.CostScope != CostScopeFeeTaxOnly:
		return invalid("cost scope %q is not one this build measures", o.CostScope)
	case out.CostModelFingerprint == "":
		return invalid("an observation names the cost model it was computed under — " +
			"an unfingerprinted ratio cannot be told from a measured one")
	case out.ObservedAt.IsZero():
		return invalid("an observation is stamped with when it was taken")
	}

	switch out.Outcome {
	case OutcomeRefusedChain:
		if out.DecisionID != "" {
			return invalid("a chain refusal writes no decision, so it references none")
		}
		if out.StoppedStep == "" || out.ReasonCode == "" {
			return invalid("a chain refusal records which rung refused and why")
		}
		if out.IssuanceReasonCode != "" {
			return invalid("a chain refusal never reached the issuance, so it carries no issuance reason")
		}
	case OutcomeAllowedIssued:
		if out.DecisionID == "" {
			return invalid("an issued observation references the decision that was written")
		}
		if out.StoppedStep != "" || out.ReasonCode != "" {
			return invalid("an allowed verdict stopped at no rung")
		}
		if out.IssuanceReasonCode != "" {
			return invalid("an issued observation carries no issuance refusal reason")
		}
		if out.IssuedAt.IsZero() {
			return invalid("an issued observation carries the decision's issued_at")
		}
	case OutcomeAllowedIssuanceRefused:
		if out.DecisionID != "" {
			return invalid("a refused issuance rolled its decision back, so there is none to reference")
		}
		if out.IssuanceReasonCode == "" {
			return invalid("a refused issuance records the stable reason the ledger gave")
		}
		if out.StoppedStep != "" || out.ReasonCode != "" {
			return invalid("the chain allowed, so it stopped at no rung")
		}
	default:
		return invalid("outcome %q is not one of the three this build records", o.Outcome)
	}

	if out.Reconstructed {
		if out.ReconstructedAt.IsZero() {
			return invalid("a reconstructed observation is stamped with when it was rebuilt")
		}
		if out.Outcome != OutcomeAllowedIssued {
			return invalid("only an issued decision can be reconstructed; a refusal has no preimage to rebuild from")
		}
	} else if !out.ReconstructedAt.IsZero() {
		return invalid("a row that was not reconstructed carries no reconstruction instant")
	}
	return out, nil
}

// --- gap detection -----------------------------------------------------------

// ErrGapScanUnusable means the scan was asked for with a schedule that cannot do
// what it claims. It is refused rather than run, for the reason PruneSpentNonces
// refuses a too-short retention: a scan on the wrong schedule silently produces
// either double rows or permanent loss, and neither announces itself.
var ErrGapScanUnusable = errors.New("journal: the observation gap scan schedule is unusable")

// MissingObservation is one issued entry decision with no observation.
//
// Preimage is the parsed RiskIntent, which is what makes the row rebuildable:
// three prices, market, symbol, quantity and policy version are all in it. What
// is *not* in it is the rate set the original break-even was computed under, so
// the net ratio a reconstruction writes is a new measurement wearing the old
// decision's timestamp — which is why a reconstructed row is marked.
type MissingObservation struct {
	DecisionID string
	IssuedAt   time.Time
	Preimage   RiskIntent
}

// GapScanOptions is the schedule the scan is running on.
type GapScanOptions struct {
	// WriteDeadline is how long after issuance an observation write is still
	// treated as in flight. A decision younger than this is skipped: mistaking a
	// write still on its way for a gap produces a reconstruction *and* the real
	// row a moment later, and one entry counted twice biases the distribution a
	// live threshold gets drawn from (R3, task 2.5c).
	WriteDeadline time.Duration
	// Cycle is how often this scan runs.
	Cycle time.Duration
	// PruningHorizon is how long a decision row survives. The cycle must be
	// shorter than it: once the decision is gone the gap can be neither detected
	// nor rebuilt, so a scan slower than the horizon loses rows by design.
	PruningHorizon time.Duration
}

func (o GapScanOptions) validate() error {
	switch {
	case o.WriteDeadline <= 0:
		return fmt.Errorf("%w: the write deadline must be positive", ErrGapScanUnusable)
	case o.Cycle <= 0:
		return fmt.Errorf("%w: the cycle must be positive", ErrGapScanUnusable)
	case o.PruningHorizon <= 0:
		return fmt.Errorf("%w: the pruning horizon must be positive", ErrGapScanUnusable)
	case o.Cycle >= o.PruningHorizon:
		return fmt.Errorf("%w: cycle %s is not shorter than the decision pruning horizon %s, "+
			"so a gap can be pruned before it is ever seen", ErrGapScanUnusable, o.Cycle, o.PruningHorizon)
	case o.WriteDeadline >= o.PruningHorizon:
		return fmt.Errorf("%w: write deadline %s is not shorter than the decision pruning horizon %s",
			ErrGapScanUnusable, o.WriteDeadline, o.PruningHorizon)
	}
	return nil
}

// EntryObservationGap is one scan's answer.
type EntryObservationGap struct {
	// Missing is the rebuildable set: past the write deadline, still inside the
	// pruning horizon.
	Missing []MissingObservation
	// LapsedBeyondHorizon counts unobserved issued decisions older than the
	// pruning horizon — the ones a decision pruner would already have taken, and
	// therefore the count of measurements this schedule is losing (SHALL — 영구
	// 손실이 된 건수는 계수되어야 한다).
	//
	// TossOS prunes no decisions today (nonce.go and trade_outcomes.go are the
	// only sweeps), so these rows are still on disk and still rebuildable. The
	// count is therefore an alarm that the scan is running slower than the
	// horizon it was configured against, which is the actionable form of the
	// requirement: once a pruner lands, exactly this set is what disappears.
	LapsedBeyondHorizon int
}

// DetectMissingEntryObservations finds issued entry decisions that carry no
// observation, by anti-join against the contract table.
//
// The anti-join is read-only and takes no lock the order path competes for. It
// reads `decisions` because that is the only place the evidence is; the *analysis*
// queries (EntryObservations) still touch nothing but the observation table,
// which is what trade-analytics' isolation requirement is about.
func (j *Journal) DetectMissingEntryObservations(
	ctx context.Context, now time.Time, opts GapScanOptions,
) (EntryObservationGap, error) {
	if err := opts.validate(); err != nil {
		return EntryObservationGap{}, err
	}
	now = now.UTC()
	deadline := formatJournalTime(now.Add(-opts.WriteDeadline))
	horizon := formatJournalTime(now.Add(-opts.PruningHorizon))

	rows, err := j.db.QueryContext(ctx,
		`SELECT d.id, d.issued_at, d.risk_preimage
		   FROM decisions d
		  WHERE d.safety_class = ?
		    AND d.preimage_kind = ?
		    AND d.issued_at < ?
		    AND d.issued_at >= ?
		    AND NOT EXISTS (SELECT 1 FROM entry_decision_observations o
		                     WHERE o.decision_id = d.id)
		  ORDER BY d.issued_at, d.id`,
		SafetyClassExposureRaising, PreimageKindRiskIntent, deadline, horizon)
	if err != nil {
		return EntryObservationGap{}, fmt.Errorf("journal: detecting missing entry observations: %w", err)
	}
	defer rows.Close()

	var gap EntryObservationGap
	for rows.Next() {
		var id, issuedAt, canonical string
		if err := rows.Scan(&id, &issuedAt, &canonical); err != nil {
			return EntryObservationGap{}, fmt.Errorf("journal: detecting missing entry observations: %w", err)
		}
		parsed, err := ParsePreimage(PreimageKindRiskIntent, canonical)
		if err != nil {
			return EntryObservationGap{}, fmt.Errorf("journal: decision %s preimage: %w", id, err)
		}
		intent, ok := parsed.(RiskIntent)
		if !ok {
			return EntryObservationGap{}, fmt.Errorf("journal: decision %s carries a %T, not a RiskIntent", id, parsed)
		}
		at, err := parseJournalTime(issuedAt)
		if err != nil {
			return EntryObservationGap{}, fmt.Errorf("journal: decision %s issued_at: %w", id, err)
		}
		gap.Missing = append(gap.Missing, MissingObservation{DecisionID: id, IssuedAt: at, Preimage: intent})
	}
	if err := rows.Err(); err != nil {
		return EntryObservationGap{}, fmt.Errorf("journal: detecting missing entry observations: %w", err)
	}

	if err := j.db.QueryRowContext(ctx,
		`SELECT count(*) FROM decisions d
		  WHERE d.safety_class = ?
		    AND d.preimage_kind = ?
		    AND d.issued_at < ?
		    AND NOT EXISTS (SELECT 1 FROM entry_decision_observations o
		                     WHERE o.decision_id = d.id)`,
		SafetyClassExposureRaising, PreimageKindRiskIntent, horizon).
		Scan(&gap.LapsedBeyondHorizon); err != nil {
		return EntryObservationGap{}, fmt.Errorf("journal: counting lapsed entry observations: %w", err)
	}
	return gap, nil
}

// --- retention ---------------------------------------------------------------

// EntryObservationRetention is how long an observation is kept. It is
// TradeOutcomeRetention's 180 days, and deliberately the same number: the two
// tables are read together by the analysis this change exists to enable, and a
// shorter window on one of them would silently truncate the join.
const EntryObservationRetention = 180 * 24 * time.Hour

// PruneEntryObservations deletes observations taken before the cutoff and reports
// how many rows went.
//
// A plain delete on an indexed column, taking no other lock — the same shape as
// PruneTradeOutcomes, and asynchronous for the same reason: an observation is
// written on every entry verdict, so an unbounded table would grow with trading
// volume, and a sweep that contended with the order path would make the measuring
// slow down the thing it measures.
func (j *Journal) PruneEntryObservations(ctx context.Context, before time.Time) (int64, error) {
	res, err := j.db.ExecContext(ctx,
		`DELETE FROM entry_decision_observations WHERE observed_at < ?`,
		formatJournalTime(before.UTC()))
	if err != nil {
		return 0, fmt.Errorf("journal: pruning entry observations: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("journal: pruning entry observations: %w", err)
	}
	return n, nil
}

// --- reading back ------------------------------------------------------------

const entryObservationSelect = `SELECT id, account_ref, market, symbol, entry_price,
	stop_price, target_price, break_even_price, gross_reward_risk, net_reward_risk,
	cost_scope, cost_model_fingerprint, stopped_step, reason_code, outcome,
	issuance_reason_code, reconstructed, observed_at, issued_at, reconstructed_at,
	decision_id
	FROM entry_decision_observations`

// EntryObservations reads observations back, oldest first. It is the analysis
// path's only entry point and it touches no contract table.
func (j *Journal) EntryObservations(ctx context.Context) ([]EntryObservation, error) {
	rows, err := j.db.QueryContext(ctx, entryObservationSelect+" ORDER BY observed_at, id")
	if err != nil {
		return nil, fmt.Errorf("journal: reading entry observations: %w", err)
	}
	defer rows.Close()

	var out []EntryObservation
	for rows.Next() {
		obs, err := scanEntryObservation(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, obs)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: reading entry observations: %w", err)
	}
	return out, nil
}

func scanEntryObservation(row rowScanner) (EntryObservation, error) {
	var (
		obs                                   EntryObservation
		stop, target, breakEven, gross, net   sql.NullString
		issuedAt, reconstructedAt, decisionID sql.NullString
		reconstructed                         int
		observedAt                            string
	)
	err := row.Scan(&obs.ID, &obs.AccountRef, &obs.Market, &obs.Symbol, &obs.EntryPrice,
		&stop, &target, &breakEven, &gross, &net, &obs.CostScope,
		&obs.CostModelFingerprint, &obs.StoppedStep, &obs.ReasonCode, &obs.Outcome,
		&obs.IssuanceReasonCode, &reconstructed, &observedAt, &issuedAt,
		&reconstructedAt, &decisionID)
	if err != nil {
		return EntryObservation{}, fmt.Errorf("journal: reading an entry observation: %w", err)
	}
	obs.StopPrice = stop.String
	obs.TargetPrice = target.String
	obs.BreakEvenPrice = breakEven.String
	obs.GrossRewardRisk = gross.String
	obs.NetRewardRisk = net.String
	obs.DecisionID = decisionID.String
	obs.Reconstructed = reconstructed == 1
	if obs.ObservedAt, err = parseJournalTime(observedAt); err != nil {
		return EntryObservation{}, fmt.Errorf("journal: observation %s observed_at: %w", obs.ID, err)
	}
	if issuedAt.Valid {
		if obs.IssuedAt, err = parseJournalTime(issuedAt.String); err != nil {
			return EntryObservation{}, fmt.Errorf("journal: observation %s issued_at: %w", obs.ID, err)
		}
	}
	if reconstructedAt.Valid {
		if obs.ReconstructedAt, err = parseJournalTime(reconstructedAt.String); err != nil {
			return EntryObservation{}, fmt.Errorf("journal: observation %s reconstructed_at: %w", obs.ID, err)
		}
	}
	return obs, nil
}

// --- small helpers -----------------------------------------------------------
//
// boolToInt (fills.go) and isUniqueViolation (exit_state.go) are the package's
// existing ones. A second spelling of either would be a second answer.

// nullableTime stores a zero instant as NULL. A zero time formatted as text would
// read as the year 1, which sorts before every real row and is exactly the kind
// of value a range scan silently includes.
func nullableTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return formatJournalTime(t)
}

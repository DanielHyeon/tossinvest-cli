package journal

// provenance.go answers "이 포지션은 왜 존재하는가" with one query
// (change add-core-domain task 6.4; position-ledger "Provenance Lineage").
//
// # The requirement, and the part of it that is a prohibition
//
// 자동 거래의 전 과정은 결정 근거와 함께 추적 가능해야 한다(SHALL) …
// "이 포지션은 왜 존재하는가"를 단일 질의로 재구성할 수 있어야 하며(SHALL),
// 그 조인 경로는 스키마의 명시적 참조 컬럼이다(SHALL — 시간창 휴리스틱 매칭
// 금지).
//
// The prohibition is the harder half. An engine that placed one order per second
// leaves a journal in which almost any two rows within a second of each other
// look related, and a chain built by "the decision closest in time to when the
// position opened" is a chain that is right until the day it matters. So every
// edge below is a declared column, and a link that has no column is a link this
// file does not draw:
//
//	positions.entry_decision_id      → decisions.id
//	positions.adoption_id            → position_adoptions.id
//	mutation_attempts.decision_id    → decisions.id
//	mutation_attempts.intent_id      → intents.id
//	mutation_attempts.broker_order_id → fill_events.order_id
//	position_adjustments.position_id → positions.id
//	exit_events.position_id          → positions.id
//	exit_events.proposed_intent_id   → intents.id  (and its attempts, and their fills)
//
// An externally acquired position has `entry_decision_id` NULL, and the chain
// then carries no decision at all rather than the nearest one — which is the
// same fact NULL already records, said out loud.
//
// # The adopted shape
//
// Since adopt-external-positions there is a second shape, and position-ledger
// fixes it: `ADOPTION → POSITION → EXIT_EVENT …`. The entry-side arms — DECISION,
// INTENT, ATTEMPT, FILL — are empty and that is not a gap: nobody decided to buy
// the shares, no intent was recorded, and no local order filled into them. What
// replaces the decision is the adoption record, which carries the observation and
// the synthetic stop the baseline was built from. The exit-side arms fill in
// normally from the moment the engine liquidates any of it, because those *are*
// the engine's own orders.
//
// # Why the exit half joins through the intent
//
// `fill_events` has no position reference (design D7's table does not carry one,
// and adding a column would be a second schema version inside a change whose
// migration rule is one atomic v6 — issues.md, task 6.1). That is not a gap on
// the exit side, because `exit_events` already carries the reference the exit
// half needs: the judgement names the intent it proposed, the intent's attempts
// name their broker orders, and those orders' fills are the liquidation. The
// whole path is declared columns.
//
// What it does not reach is a reduction nobody proposed through an exit
// judgement — a manual flatten, an operator's own sell. Those have their own
// decision and their own intent, and attaching them to a position by timing is
// the heuristic the spec forbids. issues.md records the options.
//
// # One query
//
// A UNION ALL over the join path rather than one round trip per hop. Not for
// speed: a chain assembled from several reads is a chain assembled from several
// *instants*, and a fill landing between two of them produces an answer that was
// never true. The rank column keeps the causal order inside one timestamp, which
// matters because journal stamps are second-resolution.
//
// # Broker-behaviour claims
//
// None. Every value is a column another part of the journal wrote.

import (
	"context"
	"fmt"
	"strings"
)

// Provenance step kinds. The strings are part of the answer — an operator or an
// export reads them — so they are appended to rather than renamed.
const (
	// ProvenanceDecision is the entry decision the position was opened under.
	ProvenanceDecision = "DECISION"
	// ProvenanceAdoption is the adoption record an externally acquired position
	// was taken into exit management under. It is the alternative to DECISION and
	// never appears beside one.
	ProvenanceAdoption = "ADOPTION"
	// ProvenanceIntent is an intent an attempt under that decision belongs to.
	ProvenanceIntent = "INTENT"
	// ProvenanceAttempt is one PLACE/CANCEL/AMEND against the broker.
	ProvenanceAttempt = "ATTEMPT"
	// ProvenanceFill is one positive delta on an order those attempts named.
	ProvenanceFill = "FILL"
	// ProvenancePosition is the instance opening.
	ProvenancePosition = "POSITION"
	// ProvenanceAdjustment is a compare-and-append correction against it.
	ProvenanceAdjustment = "ADJUSTMENT"
	// ProvenanceExitEvent is one exit judgement recorded against the position.
	ProvenanceExitEvent = "EXIT_EVENT"
	// ProvenanceExitIntent is an intent an exit judgement proposed.
	ProvenanceExitIntent = "EXIT_INTENT"
	// ProvenanceExitAttempt is an attempt of that intent.
	ProvenanceExitAttempt = "EXIT_ATTEMPT"
	// ProvenanceExitFill is a fill of the liquidation those attempts placed.
	ProvenanceExitFill = "EXIT_FILL"
	// ProvenanceClose is the instance reaching CLOSED.
	ProvenanceClose = "CLOSE"
)

// ProvenanceStep is one link of the chain.
type ProvenanceStep struct {
	// Kind is one of the Provenance* constants.
	Kind string
	// At is when it happened, RFC3339 UTC.
	At string
	// Ref is the row's identifier: a decision id, an intent id, a broker order
	// number, an adjustment id, an exit event's sequence number.
	Ref string
	// Detail is what that row said, rendered for a person.
	Detail string
}

// PositionProvenance is the whole chain, oldest first.
type PositionProvenance struct {
	PositionID string
	Steps      []ProvenanceStep
}

// PositionProvenance reconstructs why a position exists, in time order.
//
// It returns ErrPositionNotFound when no instance carries that id. Every chain
// of a real position contains at least its POSITION step, so no rows means no
// position — which is a different answer from "a position with no reason", and
// the two must not be confused by a caller deciding whether to trust it.
func (j *Journal) PositionProvenance(ctx context.Context, positionID string) (PositionProvenance, error) {
	id := strings.TrimSpace(positionID)
	chain := PositionProvenance{PositionID: id}

	rows, err := j.db.QueryContext(ctx, provenanceQuery, id)
	if err != nil {
		return PositionProvenance{}, fmt.Errorf("journal: reading the provenance of %s: %w", id, err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			rank            int
			step            ProvenanceStep
			one, two, three string
		)
		if err := rows.Scan(&rank, &step.Kind, &step.At, &step.Ref, &one, &two, &three); err != nil {
			return PositionProvenance{}, fmt.Errorf(
				"journal: reading the provenance of %s: %w", id, err)
		}
		step.Detail = provenanceDetail(step.Kind, one, two, three)
		chain.Steps = append(chain.Steps, step)
	}
	if err := rows.Err(); err != nil {
		return PositionProvenance{}, fmt.Errorf("journal: reading the provenance of %s: %w", id, err)
	}
	if len(chain.Steps) == 0 {
		return PositionProvenance{}, fmt.Errorf("%w: %s", ErrPositionNotFound, id)
	}
	return chain, nil
}

// provenanceDetail renders a step's three carried values.
//
// The rendering is here rather than in SQL so the columns stay columns: a
// concatenation in the query would make the values unreadable to anything that
// wanted them apart, and this file is the only reader that wants them together.
func provenanceDetail(kind, one, two, three string) string {
	switch kind {
	case ProvenanceDecision:
		return fmt.Sprintf("%s decision, preimage %s (hash %s)", one, two, three)
	case ProvenanceAdoption:
		return fmt.Sprintf("adopted at %s with a synthetic stop of %s (cost basis %s)", one, two, three)
	case ProvenanceIntent, ProvenanceExitIntent:
		return fmt.Sprintf("%s %s of %s", one, two, three)
	case ProvenanceAttempt, ProvenanceExitAttempt:
		return fmt.Sprintf("%s attempt, %s, broker order %s", one, two, orNone(three))
	case ProvenanceFill, ProvenanceExitFill:
		return fmt.Sprintf("filled %s (cumulative %s) at %s", one, two, orNone(three))
	case ProvenancePosition:
		return fmt.Sprintf("instance opened %s holding %s at %s", one, two, orNone(three))
	case ProvenanceAdjustment:
		return fmt.Sprintf("%s adjustment to %s, from the account as of %s", one, two, three)
	case ProvenanceExitEvent:
		return fmt.Sprintf("judged %s at level %s, proposing intent %s",
			orNone(one), orNone(two), orNone(three))
	case ProvenanceClose:
		return fmt.Sprintf("instance %s holding %s", one, two)
	default:
		return strings.TrimSpace(strings.Join([]string{one, two, three}, " "))
	}
}

func orNone(s string) string {
	if strings.TrimSpace(s) == "" {
		return "(none)"
	}
	return s
}

// provenanceQuery is the single query. The CTEs are the join path, written once
// each so the UNION arms cannot drift from one another; DISTINCT appears where a
// retry can name one intent or one broker order twice, because a chain that
// lists the same order three times is a chain that reads like three orders.
const provenanceQuery = `
WITH pos AS (
	SELECT id, entry_decision_id, adoption_id, state, quantity, avg_price,
	       coalesce(opened_at, '') AS opened_at, coalesce(closed_at, '') AS closed_at
	  FROM positions WHERE id = ?
),
entry_attempt AS (
	SELECT a.id, a.intent_id, a.kind, a.state, a.broker_order_id, a.recorded_at
	  FROM mutation_attempts a
	  JOIN pos ON a.decision_id = pos.entry_decision_id
),
entry_intent AS (SELECT DISTINCT intent_id FROM entry_attempt WHERE intent_id <> ''),
entry_order AS (
	SELECT DISTINCT a.broker_order_id, TRIM(i.account_ref) account_ref,
	       LOWER(TRIM(i.market)) market, TRIM(i.trading_day) trading_day,
	       UPPER(TRIM(i.symbol)) symbol, UPPER(TRIM(i.side)) side
	  FROM entry_attempt a JOIN intents i ON i.id = a.intent_id
	 WHERE a.broker_order_id <> ''
),
exit_intent AS (
	SELECT DISTINCT e.proposed_intent_id AS intent_id
	  FROM exit_events e JOIN pos ON e.position_id = pos.id
	 WHERE e.proposed_intent_id IS NOT NULL AND e.proposed_intent_id <> ''
),
exit_attempt AS (
	SELECT a.id, a.intent_id, a.kind, a.state, a.broker_order_id, a.recorded_at
	  FROM mutation_attempts a JOIN exit_intent x ON a.intent_id = x.intent_id
),
exit_order AS (
	SELECT DISTINCT a.broker_order_id, TRIM(i.account_ref) account_ref,
	       LOWER(TRIM(i.market)) market, TRIM(i.trading_day) trading_day,
	       UPPER(TRIM(i.symbol)) symbol, UPPER(TRIM(i.side)) side
	  FROM exit_attempt a JOIN intents i ON i.id = a.intent_id
	 WHERE a.broker_order_id <> ''
),
order_scope_count AS (
	SELECT broker_order_id, count(*) scope_count FROM (
		SELECT DISTINCT a.broker_order_id, TRIM(i.account_ref), LOWER(TRIM(i.market)),
		       TRIM(i.trading_day), UPPER(TRIM(i.symbol)), UPPER(TRIM(i.side))
		  FROM mutation_attempts a JOIN intents i ON i.id = a.intent_id
		 WHERE a.broker_order_id <> ''
	) GROUP BY broker_order_id
)

SELECT 0, 'ADOPTION', ad.observed_at, ad.id, ad.observed_price, ad.synthetic_stop,
       coalesce(ad.cost_basis, '')
  FROM pos JOIN position_adoptions ad ON ad.id = pos.adoption_id
UNION ALL
SELECT 1, 'DECISION', d.issued_at, d.id, d.safety_class, d.preimage_kind, d.risk_hash
  FROM pos JOIN decisions d ON d.id = pos.entry_decision_id
UNION ALL
SELECT 2, 'INTENT', i.created_at, i.id, i.side, i.quantity, i.symbol
  FROM entry_intent x JOIN intents i ON i.id = x.intent_id
UNION ALL
SELECT 3, 'ATTEMPT', a.recorded_at, a.id, a.kind, a.state, a.broker_order_id
  FROM entry_attempt a
UNION ALL
SELECT 4, 'FILL', f.committed_at, f.order_id, f.delta_quantity, f.cumulative_quantity, f.average_price
  FROM entry_order o
  JOIN order_scope_count c ON c.broker_order_id = o.broker_order_id
  JOIN fill_events f ON f.order_id = o.broker_order_id AND (
	(TRIM(f.account_ref) = o.account_ref AND LOWER(TRIM(f.market)) = o.market
	 AND TRIM(f.trading_day) = o.trading_day AND UPPER(TRIM(f.symbol)) = o.symbol
	 AND UPPER(TRIM(f.side)) = o.side)
	OR (TRIM(f.account_ref) = '' AND TRIM(f.trading_day) = '' AND TRIM(f.side) = ''
	    AND c.scope_count = 1 AND (TRIM(f.market) = '' OR LOWER(TRIM(f.market)) = o.market)
	    AND UPPER(TRIM(f.symbol)) = o.symbol)
  )
UNION ALL
SELECT 5, 'POSITION', pos.opened_at, pos.id, pos.state, pos.quantity, pos.avg_price
  FROM pos WHERE pos.opened_at <> ''
UNION ALL
SELECT 6, 'ADJUSTMENT', adj.created_at, adj.id, adj.kind, adj.new_quantity, adj.broker_as_of
  FROM position_adjustments adj JOIN pos ON adj.position_id = pos.id
UNION ALL
SELECT 7, 'EXIT_EVENT', e.created_at, cast(e.id AS TEXT), coalesce(e.action, ''),
       coalesce(e.level_after, ''), coalesce(e.proposed_intent_id, '')
  FROM exit_events e JOIN pos ON e.position_id = pos.id
UNION ALL
SELECT 8, 'EXIT_INTENT', i.created_at, i.id, i.side, i.quantity, i.symbol
  FROM exit_intent x JOIN intents i ON i.id = x.intent_id
UNION ALL
SELECT 9, 'EXIT_ATTEMPT', a.recorded_at, a.id, a.kind, a.state, a.broker_order_id
  FROM exit_attempt a
UNION ALL
SELECT 10, 'EXIT_FILL', f.committed_at, f.order_id, f.delta_quantity, f.cumulative_quantity,
       f.average_price
  FROM exit_order o
  JOIN order_scope_count c ON c.broker_order_id = o.broker_order_id
  JOIN fill_events f ON f.order_id = o.broker_order_id AND (
	(TRIM(f.account_ref) = o.account_ref AND LOWER(TRIM(f.market)) = o.market
	 AND TRIM(f.trading_day) = o.trading_day AND UPPER(TRIM(f.symbol)) = o.symbol
	 AND UPPER(TRIM(f.side)) = o.side)
	OR (TRIM(f.account_ref) = '' AND TRIM(f.trading_day) = '' AND TRIM(f.side) = ''
	    AND c.scope_count = 1 AND (TRIM(f.market) = '' OR LOWER(TRIM(f.market)) = o.market)
	    AND UPPER(TRIM(f.symbol)) = o.symbol)
  )
UNION ALL
SELECT 11, 'CLOSE', pos.closed_at, pos.id, pos.state, pos.quantity, ''
  FROM pos WHERE pos.closed_at <> ''

ORDER BY 3, 1, 4
`

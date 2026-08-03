package journal

// account_views.go holds the three account-scoped reads the operator console
// needs (change add-operator-dashboard, task 1.2).
//
// # Why they are new rather than reused
//
// Every existing reader in this package answers the question the engine asks, and
// the console asks different ones:
//
//	OpenExitStates   the observation loop's working set — completed = 0 only. A
//	                 dashboard built on it would show the managed positions and
//	                 silently omit the unmanaged ones, which is the exact set an
//	                 operator opens the screen to find.
//	ExitEvents       keyed by position id, which the console does not have until
//	                 it has done the join below.
//	TradeOutcomes    the frozen rows without the symbol they belong to, and with
//	                 held_seconds coalesced to zero — a NULL hold time and a
//	                 zero-second one are different facts and the screen renders
//	                 them differently.
//
// # Nothing here computes anything
//
// Every value is a column. The joins add the two facts the frozen outcome row
// does not carry (positions.symbol, exit_states.entry_price) and stop there: the
// exit price has no column anywhere in the schema, and deriving one from the
// fills would be the recomputation the freeze exists to prevent (trade-analytics,
// and trade_outcomes.go's header).

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	marketclock "github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
)

// PositionExit is one projected position with its protection state, if it has
// one.
//
// HasExit is separate from a zero-valued Exit because "no exit state" and "an
// exit state whose numbers happen to be empty" are different things, and the
// screen says different words about them.
type PositionExit struct {
	Position Position
	Exit     ExitState
	HasExit  bool
	// Quarantine is the unreleased exit-snapshot quarantine on *this* position's
	// current generation, when there is one. It travels with the row because the
	// engine skips a quarantined position entirely — the exit loop refuses it
	// with ErrExitSnapshotQuarantined and judges nothing — so a reader that shows
	// this row's protection line without it would be asserting protection the
	// engine is not providing.
	//
	// Correlating it here rather than in each reader is the point: a second place
	// that decides which quarantine belongs to which generation is a second place
	// that can get it wrong.
	Quarantine *ExitSnapshotQuarantine
}

// AccountExitEvent is one exit judgement with the symbol it was made about.
type AccountExitEvent struct {
	Event  ExitEvent
	Symbol string
	Market string
}

// BrokerOrderExitLink is the exact read-only lineage from a broker order to the
// exit judgement that proposed its intent. AttemptID/AttemptNo are carried so a
// transport can disclose the join rather than merely claim that it happened.
// Ambiguous links deliberately carry no Event.
type BrokerOrderExitLink struct {
	BrokerOrderID string
	AccountRef    string
	Market        string
	TradingDay    string
	Engine        bool
	AttemptID     string
	AttemptNo     int
	IntentID      string
	Event         ExitEvent
	Ambiguous     bool
	UnknownReason string
}

// BrokerOrderScope is the minimum canonical identity needed to ask whether a
// visible broker order belongs to one journal exit. Broker order numbers alone
// are not assumed globally unique across accounts or trading days.
type BrokerOrderScope struct {
	BrokerOrderID string `json:"broker_order_id"`
	AccountRef    string `json:"account_ref"`
	Market        string `json:"market"`
	TradingDay    string `json:"trading_day"`
}

const (
	MaxBrokerOrderEvidenceScopes = 256
	maxBrokerOrderLineageDepth   = 32
	maxEvidenceRowsPerNode       = 8
)

var ErrBrokerOrderEvidenceScope = errors.New("journal: broker-order evidence scope is invalid or too large")

// schemaV11 is intentionally index-only. It changes no historical row and
// supports the console's bounded event-to-intent evidence join.
const schemaV11 = `
CREATE INDEX idx_exit_events_proposed_intent
ON exit_events(proposed_intent_id)
WHERE proposed_intent_id IS NOT NULL;
`

// TradeTrip is one frozen round trip plus the two joined facts the frozen row
// does not hold.
//
// EntryPrice is exit_states.entry_price — the price frozen at t0 — and is empty
// when the position never had an exit state. HeldSecondsKnown distinguishes a
// NULL hold time from a zero one; Outcome.HeldSeconds is 0 in both cases, which
// is why the flag exists.
type TradeTrip struct {
	Outcome          TradeOutcome
	Symbol           string
	Market           string
	EntryPrice       string
	HeldSecondsKnown bool
}

// AccountRefs lists every account reference the projection holds, sorted.
//
// The console has no configured account: it renders whatever the engine has been
// trading, and this is how it finds out what that is.
func (r *ReadOnly) AccountRefs(ctx context.Context) ([]string, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT DISTINCT account_ref FROM positions ORDER BY account_ref`)
	if err != nil {
		return nil, fmt.Errorf("journal: listing the accounts in %s: %w", r.path, err)
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var ref string
		if err := rows.Scan(&ref); err != nil {
			return nil, fmt.Errorf("journal: reading an account reference: %w", err)
		}
		out = append(out, ref)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: listing the accounts in %s: %w", r.path, err)
	}
	return out, nil
}

// accountExitStatesFrom completes exitStateSelect (apply_hook.go) into an
// account-wide read.
//
// The projection is joined only to scope the account; every selected column
// comes from exit_states, so scanExitState reads it. The select list is not
// spelled here on purpose: four of those columns have exactly one writer and
// apply_hook.go's layout guard fails if any other file in this package so much as
// names them, which is a rule worth keeping intact for the sake of a dashboard
// query.
const accountExitStatesFrom = ` e JOIN positions p ON p.id = e.position_id
	 WHERE p.account_ref = ?`

// LivePositionExits returns every non-CLOSED position of one account with its
// exit state, oldest symbol first.
//
// "Live" is state <> CLOSED rather than "has quantity": a position in OPENING or
// CLOSING is one the operator has to be able to see, and CLOSED instances belong
// to the history screen where their frozen outcome is.
//
// It is two statements rather than one LEFT JOIN — see accountExitStatesFrom.
// The merge is in Go, and a position with no exit state keeps HasExit false,
// which is the row the screen marks 관리 외(미편입) or "아직 exit 상태 없음".
func (r *ReadOnly) LivePositionExits(ctx context.Context, accountRef string) ([]PositionExit, error) {
	account := strings.TrimSpace(accountRef)

	states, err := r.accountExitStates(ctx, account)
	if err != nil {
		return nil, err
	}
	quarantines, err := r.accountActiveQuarantines(ctx, account)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, positionSelect+`
		 WHERE account_ref = ? AND state <> ?
		 ORDER BY market, symbol, instance_seq`, account, PositionClosed)
	if err != nil {
		return nil, fmt.Errorf("journal: listing the live positions of %s: %w", account, err)
	}
	defer rows.Close()

	var out []PositionExit
	for rows.Next() {
		p, err := scanPosition(rows)
		if err != nil {
			return nil, err
		}
		pe := PositionExit{Position: p, Exit: ExitState{ActiveRung: exitpolicy.NoRung}}
		if q, ok := quarantines[p.ID]; ok {
			held := q
			pe.Quarantine = &held
		}
		if state, ok := states[p.ID]; ok {
			if state.SnapshotStatus == "" && state.PolicyIdentity.ID == "" {
				identity, identityErr := legacyPolicyIdentity(state, p.Adopted())
				if identityErr == nil {
					state.PolicyIdentity = identity
				} else {
					state.Snapshot.UnknownReason = "legacy_policy_identity_unknown"
				}
			}
			pe.Exit, pe.HasExit = state, true
		}
		out = append(out, pe)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: listing the live positions of %s: %w", account, err)
	}
	return out, nil
}

// accountActiveQuarantines reads the unreleased exit-snapshot quarantines of one
// account, keyed by position id.
//
// The generation predicate is the whole correctness of this read. The active
// index is on (position_id, position_generation), so a position that was
// released and re-adopted keeps its old instance's quarantine row on disk
// forever — unreleased, because nothing releases the quarantine of a generation
// that no longer exists. The engine asks about the current generation only
// (exitloop.go: ActiveExitSnapshotQuarantine(ctx, p.ID, p.InstanceSeq)) and so
// does this: joining on p.instance_seq is what keeps a dead generation from
// closing a live position's protection line.
func (r *ReadOnly) accountActiveQuarantines(ctx context.Context, accountRef string) (map[string]ExitSnapshotQuarantine, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT q.position_id,q.position_generation,q.quarantine_version,
		 q.reason,q.evidence,q.quarantined_at
		 FROM exit_snapshot_quarantines q JOIN positions p ON p.id = q.position_id
		 WHERE p.account_ref = ? AND q.released_at IS NULL AND q.position_generation = p.instance_seq`,
		accountRef)
	if err != nil {
		return nil, fmt.Errorf("journal: listing the exit snapshot quarantines of %s: %w", accountRef, err)
	}
	defer rows.Close()

	out := map[string]ExitSnapshotQuarantine{}
	for rows.Next() {
		var q ExitSnapshotQuarantine
		if err := rows.Scan(&q.PositionID, &q.PositionGeneration, &q.Version, &q.Reason,
			&q.Evidence, &q.QuarantinedAt); err != nil {
			return nil, fmt.Errorf("journal: reading an exit snapshot quarantine: %w", err)
		}
		out[q.PositionID] = q
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: listing the exit snapshot quarantines of %s: %w", accountRef, err)
	}
	return out, nil
}

// accountExitStates reads every exit state of one account, completed or not,
// keyed by position id.
//
// Completed rows are included, unlike OpenExitStates: the observation loop wants
// its working set, and the screen wants the truth about a position whose policy
// has finished but whose instance is not CLOSED yet.
func (r *ReadOnly) accountExitStates(ctx context.Context, accountRef string) (map[string]ExitState, error) {
	rows, err := r.db.QueryContext(ctx, exitStateSelect+accountExitStatesFrom, accountRef)
	if err != nil {
		return nil, fmt.Errorf("journal: listing the exit states of %s: %w", accountRef, err)
	}
	defer rows.Close()

	out := map[string]ExitState{}
	for rows.Next() {
		result, err := scanExitStateResult(rows)
		if err != nil {
			return nil, err
		}
		// A semantic defect belongs to this position generation. Preserve the
		// typed unknown reason for the operator instead of hiding every other
		// healthy row behind a global query error.
		out[result.State.PositionID] = result.State
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: listing the exit states of %s: %w", accountRef, err)
	}
	return out, nil
}

// AccountExitEvents returns the newest limit judgements of one account, oldest
// first within that window.
//
// The window keeps the newest end. A screen that truncated the other way would
// show an operator last month while this morning's ratchet is off the bottom of
// the page.
func (r *ReadOnly) AccountExitEvents(ctx context.Context, accountRef string, limit int) ([]AccountExitEvent, error) {
	account := strings.TrimSpace(accountRef)
	if limit <= 0 {
		return nil, nil
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT * FROM (
			SELECT v.id, v.position_id, coalesce(v.observed_price,''), coalesce(v.high_water,''),
			       coalesce(v.baseline_after,''), coalesce(v.level_after,''), coalesce(v.action,''),
			       coalesce(v.proposed_intent_id,''), v.created_at, v.position_generation, v.policy_id,
			       v.policy_version, v.policy_digest, v.snapshot_id, v.decision_id, v.observation_id,
			       v.next_target, v.next_protection, v.observation_source, v.observed_at,
			       v.projected_quantity, v.proposal_ratio, v.state_only, v.suppressed_reason,
			       v.saved_snapshot_json, v.recomputed_snapshot_json, v.effective_snapshot_json,
			       v.effective_source, v.arm_suppressed_reason,
			       p.symbol, p.market
			  FROM exit_events v
			  JOIN positions p ON p.id = v.position_id
			 WHERE p.account_ref = ?
			 ORDER BY v.created_at DESC, v.id DESC
			 LIMIT ?
		) ORDER BY created_at, id`, account, limit)
	if err != nil {
		return nil, fmt.Errorf("journal: listing the exit history of %s: %w", account, err)
	}
	defer rows.Close()

	var out []AccountExitEvent
	for rows.Next() {
		var e AccountExitEvent
		var evidence exitEventEvidence
		if err := rows.Scan(&e.Event.ID, &e.Event.PositionID, &e.Event.ObservedPrice,
			&e.Event.HighWater, &e.Event.BaselineAfter, &e.Event.LevelAfter, &e.Event.Action,
			&e.Event.ProposedIntentID, &e.Event.CreatedAt, &evidence.Generation, &evidence.PolicyID,
			&evidence.PolicyVersion, &evidence.PolicyDigest, &evidence.SnapshotID, &evidence.DecisionID,
			&evidence.ObservationID, &evidence.NextTarget, &evidence.NextProtection,
			&evidence.ObservationSource, &evidence.ObservedAt, &evidence.ProjectedQuantity,
			&evidence.ProposalRatio, &evidence.StateOnly, &evidence.SuppressedReason,
			&evidence.SavedJSON, &evidence.RecomputedJSON, &evidence.EffectiveJSON,
			&evidence.EffectiveSource, &evidence.ArmSuppressedReason,
			&e.Symbol, &e.Market); err != nil {
			return nil, fmt.Errorf("journal: reading an exit event of %s: %w", account, err)
		}
		// Event corruption is represented on the individual view. The account
		// history remains readable and never recomputes a replacement value.
		if err := hydrateExitEventEvidence(&e.Event, evidence); err != nil {
			e.Event.Evaluation.Effective.Snapshot = nil
			e.Event.Evaluation.Effective.UnknownReason = "invalid_event_evidence"
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: listing the exit history of %s: %w", account, err)
	}
	return out, nil
}

// BrokerOrderExitLinks connects only the requested visible broker orders to
// persisted exit snapshots through explicit journal references:
//
//	broker order id -> mutation attempt -> intent id -> exit event
//
// Account/day are mandatory query predicates. A validated AMEND edge may carry
// the evidence from an ancestor order to its replacement descendant. Cycle,
// branching, excessive depth, duplicate event, or row-bound evidence is never
// guessed through. The one recursive set query contains every event column, so
// it neither materializes all exit_events nor performs an event-per-row N+1.
func (r *ReadOnly) BrokerOrderExitLinks(ctx context.Context,
	scopes []BrokerOrderScope,
) ([]BrokerOrderExitLink, error) {
	if len(scopes) == 0 {
		return nil, nil
	}
	if len(scopes) > MaxBrokerOrderEvidenceScopes {
		return nil, fmt.Errorf("%w: %d scopes exceeds %d", ErrBrokerOrderEvidenceScope,
			len(scopes), MaxBrokerOrderEvidenceScopes)
	}
	clean := make([]BrokerOrderScope, len(scopes))
	for i, scope := range scopes {
		market, marketErr := marketclock.ParseMarket(scope.Market)
		if marketErr != nil {
			return nil, fmt.Errorf("%w: scope %d market: %v", ErrBrokerOrderEvidenceScope, i, marketErr)
		}
		clean[i] = BrokerOrderScope{BrokerOrderID: scope.BrokerOrderID,
			AccountRef: strings.TrimSpace(scope.AccountRef), Market: string(market),
			TradingDay: strings.TrimSpace(scope.TradingDay)}
		if len(clean[i].BrokerOrderID) == 0 || clean[i].AccountRef == "" {
			return nil, fmt.Errorf("%w: scope %d has no broker order or account", ErrBrokerOrderEvidenceScope, i)
		}
		if _, err := time.Parse("2006-01-02", clean[i].TradingDay); err != nil {
			return nil, fmt.Errorf("%w: scope %d trading day: %v", ErrBrokerOrderEvidenceScope, i, err)
		}
	}
	rawScopes, err := json.Marshal(clean)
	if err != nil {
		return nil, fmt.Errorf("journal: encoding broker-order evidence scope: %w", err)
	}
	rowLimit := len(clean)*(maxBrokerOrderLineageDepth+1)*maxEvidenceRowsPerNode + 1
	rows, err := r.db.QueryContext(ctx, `
		WITH RECURSIVE
		requested(request_index, order_id, account_ref, market, trading_day) AS (
			SELECT CAST(key AS INTEGER), json_extract(value,'$.broker_order_id'),
			       trim(json_extract(value,'$.account_ref')), trim(json_extract(value,'$.market')),
			       trim(json_extract(value,'$.trading_day'))
			  FROM json_each(?)
		),
		ancestors(request_index, requested_id, account_ref, market, trading_day,
		          current_id, depth, path, cycle) AS (
			SELECT request_index, order_id, account_ref, market, trading_day, order_id, 0,
			       json_array(order_id), 0 FROM requested
			UNION ALL
			SELECT x.request_index, x.requested_id, x.account_ref, x.market, x.trading_day,
			       l.parent_order_id, x.depth+1, json_insert(x.path,'$[#]',l.parent_order_id),
			       EXISTS(SELECT 1 FROM json_each(x.path) seen
			               WHERE seen.value=l.parent_order_id)
			  FROM ancestors x
			  JOIN lineage_edges l ON l.id=(
			       SELECT min(sl.id) FROM lineage_edges sl
			       JOIN mutation_attempts sa ON sa.id=sl.attempt_id AND sa.intent_id=sl.intent_id
			            AND sa.kind='AMEND' AND sa.state='CONFIRMED'
			            AND sa.target_order_id=sl.parent_order_id AND sa.broker_order_id=sl.child_order_id
			       JOIN intents si ON si.id=sl.intent_id AND si.account_ref=x.account_ref
			            AND si.market=x.market AND si.trading_day=x.trading_day
			       WHERE sl.child_order_id=x.current_id AND sl.relation='replaces'
			       HAVING count(*)=1)
			 WHERE x.cycle=0 AND x.depth < ?
		)
		SELECT x.request_index, x.requested_id, x.account_ref, x.market, x.trading_day,
		       x.current_id, x.depth, x.cycle,
		       (SELECT count(*) FROM lineage_edges bl
		         JOIN mutation_attempts ba ON ba.id=bl.attempt_id AND ba.intent_id=bl.intent_id
		              AND ba.kind='AMEND' AND ba.state='CONFIRMED' AND ba.target_order_id=bl.parent_order_id
		              AND ba.broker_order_id=bl.child_order_id
		         JOIN intents bi ON bi.id=bl.intent_id AND bi.account_ref=x.account_ref
		              AND bi.market=x.market AND bi.trading_day=x.trading_day
		        WHERE bl.child_order_id=x.current_id AND bl.relation='replaces') AS parent_count,
		       (SELECT count(*) FROM lineage_edges tl
		         WHERE tl.child_order_id=x.current_id AND tl.relation='replaces') AS total_parent_count,
		       EXISTS(SELECT 1 FROM mutation_attempts oa JOIN intents oi ON oi.id=oa.intent_id
		               WHERE oa.broker_order_id=x.current_id AND oa.kind='PLACE' AND oa.state='CONFIRMED'
		                 AND oi.account_ref=x.account_ref AND oi.market=x.market
		                 AND oi.trading_day=x.trading_day) AS engine,
		       (SELECT count(*) FROM mutation_attempts pa JOIN intents pi ON pi.id=pa.intent_id
		         WHERE pa.broker_order_id=x.current_id AND pa.kind='PLACE' AND pa.state='CONFIRMED'
		           AND pi.account_ref=x.account_ref AND pi.market=x.market
		           AND pi.trading_day=x.trading_day) AS place_count,
		       a.id, a.attempt_no, i.id,
		       (SELECT count(*) FROM exit_events ce WHERE ce.proposed_intent_id=i.id) AS event_count,
		       e.id, e.position_id, coalesce(e.observed_price,''), coalesce(e.high_water,''),
		       coalesce(e.baseline_after,''), coalesce(e.level_after,''), coalesce(e.action,''),
		       coalesce(e.proposed_intent_id,''), e.created_at, e.position_generation, e.policy_id,
		       e.policy_version, e.policy_digest, e.snapshot_id, e.decision_id, e.observation_id,
		       e.next_target, e.next_protection, e.observation_source, e.observed_at,
		       e.projected_quantity, e.proposal_ratio, e.state_only, e.suppressed_reason,
		       e.saved_snapshot_json, e.recomputed_snapshot_json, e.effective_snapshot_json,
		       e.effective_source, e.arm_suppressed_reason
		  FROM ancestors x
		  LEFT JOIN mutation_attempts a ON a.id=(
		       SELECT min(pa.id) FROM mutation_attempts pa JOIN intents pi ON pi.id=pa.intent_id
		       WHERE pa.broker_order_id=x.current_id AND pa.kind='PLACE' AND pa.state='CONFIRMED'
		         AND pi.account_ref=x.account_ref AND pi.market=x.market AND pi.trading_day=x.trading_day
		       HAVING count(*)=1)
		  LEFT JOIN intents i ON i.id=a.intent_id AND i.account_ref=x.account_ref
		       AND i.market=x.market AND i.trading_day=x.trading_day
		  LEFT JOIN exit_events e ON e.id=(
		       SELECT min(ee.id) FROM exit_events ee WHERE ee.proposed_intent_id=i.id HAVING count(*)=1)
		 ORDER BY x.request_index, x.depth
		 LIMIT ?`, string(rawScopes), maxBrokerOrderLineageDepth, rowLimit)
	if err != nil {
		return nil, fmt.Errorf("journal: linking scoped broker orders to exit evidence in %s: %w", r.path, err)
	}
	defer rows.Close()

	type aggregate struct {
		unsafeReason string
		events       map[int64]ExitEvent
		attemptSet   map[string]bool
		attempts     map[int64]struct {
			id string
			no int
			in string
		}
	}
	out := make([]BrokerOrderExitLink, len(clean))
	aggregates := make([]aggregate, len(clean))
	for i, scope := range clean {
		out[i] = BrokerOrderExitLink{BrokerOrderID: scope.BrokerOrderID, AccountRef: scope.AccountRef,
			Market: scope.Market, TradingDay: scope.TradingDay, UnknownReason: "exit_evidence_unlinked"}
		aggregates[i].events = map[int64]ExitEvent{}
		aggregates[i].attemptSet = map[string]bool{}
		aggregates[i].attempts = map[int64]struct {
			id string
			no int
			in string
		}{}
	}
	rowCount := 0
	for rows.Next() {
		rowCount++
		if rowCount >= rowLimit {
			return nil, fmt.Errorf("%w: evidence row bound %d reached", ErrBrokerOrderEvidenceScope, rowLimit-1)
		}
		var requestIndex, depth, cycle, parentCount, totalParentCount, engine, placeCount, eventCount int
		var requestedID, accountRef, market, tradingDay, currentID string
		var attemptID, intentID sql.NullString
		var attemptNo sql.NullInt64
		var eventID sql.NullInt64
		var eventPosition, observedPrice, highWater, baselineAfter, levelAfter sql.NullString
		var action, proposedIntent, createdAt sql.NullString
		var evidence exitEventEvidence
		if err := rows.Scan(&requestIndex, &requestedID, &accountRef, &market, &tradingDay,
			&currentID, &depth, &cycle, &parentCount, &totalParentCount, &engine, &placeCount,
			&attemptID, &attemptNo, &intentID, &eventCount, &eventID, &eventPosition, &observedPrice,
			&highWater, &baselineAfter, &levelAfter, &action, &proposedIntent, &createdAt,
			&evidence.Generation, &evidence.PolicyID, &evidence.PolicyVersion, &evidence.PolicyDigest,
			&evidence.SnapshotID, &evidence.DecisionID, &evidence.ObservationID,
			&evidence.NextTarget, &evidence.NextProtection, &evidence.ObservationSource,
			&evidence.ObservedAt, &evidence.ProjectedQuantity, &evidence.ProposalRatio,
			&evidence.StateOnly, &evidence.SuppressedReason, &evidence.SavedJSON,
			&evidence.RecomputedJSON, &evidence.EffectiveJSON, &evidence.EffectiveSource,
			&evidence.ArmSuppressedReason); err != nil {
			return nil, fmt.Errorf("journal: reading scoped broker-order exit evidence: %w", err)
		}
		if requestIndex < 0 || requestIndex >= len(out) || requestedID != out[requestIndex].BrokerOrderID ||
			accountRef != out[requestIndex].AccountRef || market != out[requestIndex].Market ||
			tradingDay != out[requestIndex].TradingDay {
			return nil, fmt.Errorf("journal: scoped broker-order evidence returned an invalid request identity")
		}
		if engine != 0 {
			out[requestIndex].Engine = true
		}
		switch {
		case cycle != 0:
			aggregates[requestIndex].unsafeReason = "lineage_cycle"
		case parentCount > 1 && aggregates[requestIndex].unsafeReason == "":
			aggregates[requestIndex].unsafeReason = "lineage_ambiguous"
		case totalParentCount > parentCount && aggregates[requestIndex].unsafeReason == "":
			aggregates[requestIndex].unsafeReason = "lineage_scope_mismatch"
		case depth >= maxBrokerOrderLineageDepth && totalParentCount > 0 && aggregates[requestIndex].unsafeReason == "":
			aggregates[requestIndex].unsafeReason = "lineage_depth_exceeded"
		case placeCount > 1 && aggregates[requestIndex].unsafeReason == "":
			aggregates[requestIndex].unsafeReason = "ambiguous_exit_evidence"
		case eventCount > 1 && aggregates[requestIndex].unsafeReason == "":
			aggregates[requestIndex].unsafeReason = "ambiguous_exit_evidence"
		}
		if !eventID.Valid {
			continue
		}
		event := ExitEvent{ID: eventID.Int64, PositionID: eventPosition.String,
			ObservedPrice: observedPrice.String, HighWater: highWater.String,
			BaselineAfter: baselineAfter.String, LevelAfter: levelAfter.String, Action: action.String,
			ProposedIntentID: proposedIntent.String, CreatedAt: createdAt.String}
		if err := hydrateExitEventEvidence(&event, evidence); err != nil {
			event.Evaluation = ExitEvaluation{}
			event.Evaluation.Effective.UnknownReason = "invalid_event_evidence"
		}
		aggregates[requestIndex].events[event.ID] = event
		aggregates[requestIndex].attemptSet[attemptID.String+"\x00"+intentID.String] = true
		aggregates[requestIndex].attempts[event.ID] = struct {
			id string
			no int
			in string
		}{id: attemptID.String, no: int(attemptNo.Int64), in: intentID.String}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: reading scoped broker-order exit evidence: %w", err)
	}
	for i := range out {
		agg := aggregates[i]
		if agg.unsafeReason != "" || len(agg.events) > 1 || len(agg.attemptSet) > 1 {
			out[i].Ambiguous = true
			out[i].UnknownReason = agg.unsafeReason
			if out[i].UnknownReason == "" {
				out[i].UnknownReason = "ambiguous_exit_evidence"
			}
			continue
		}
		for id, event := range agg.events {
			attempt := agg.attempts[id]
			out[i].Event, out[i].AttemptID, out[i].AttemptNo, out[i].IntentID =
				event, attempt.id, attempt.no, attempt.in
			out[i].UnknownReason = event.Evaluation.Effective.UnknownReason
		}
	}
	return out, nil
}

// AccountTradeTrips returns the frozen round trips of one account, oldest close
// first, with the symbol and the t0 entry price joined in.
//
// held_seconds is read as a nullable rather than coalesced, which is the one
// difference from TradeOutcomes that matters here: the screen renders an unknown
// hold time as "—" and must not be able to print it as zero seconds.
func (r *ReadOnly) AccountTradeTrips(ctx context.Context, accountRef string) ([]TradeTrip, error) {
	account := strings.TrimSpace(accountRef)
	rows, err := r.db.QueryContext(ctx, `
		SELECT o.position_id, o.realized_pnl_after_costs, o.realized_r, o.initial_risk,
		       o.initial_quantity, o.held_seconds, coalesce(o.exit_ratchet_level, ''),
		       o.exit_rung, o.closed_at, p.symbol, p.market, coalesce(e.entry_price, '')
		  FROM trade_outcomes o
		  JOIN positions p ON p.id = o.position_id
		  LEFT JOIN exit_states e ON e.position_id = o.position_id
		 WHERE p.account_ref = ?
		 ORDER BY o.closed_at, o.rowid`, account)
	if err != nil {
		return nil, fmt.Errorf("journal: listing the round trips of %s: %w", account, err)
	}
	defer rows.Close()

	var out []TradeTrip
	for rows.Next() {
		var (
			trip TradeTrip
			held sql.NullInt64
			rung sql.NullInt64
		)
		if err := rows.Scan(&trip.Outcome.PositionID, &trip.Outcome.RealizedPnLAfterCosts,
			&trip.Outcome.RealizedR, &trip.Outcome.InitialRisk, &trip.Outcome.InitialQuantity,
			&held, &trip.Outcome.ExitRatchetLevel, &rung, &trip.Outcome.ClosedAt,
			&trip.Symbol, &trip.Market, &trip.EntryPrice); err != nil {
			return nil, fmt.Errorf("journal: reading a round trip of %s: %w", account, err)
		}
		trip.HeldSecondsKnown = held.Valid
		if held.Valid {
			trip.Outcome.HeldSeconds = held.Int64
		}
		trip.Outcome.ExitRung = -1
		if rung.Valid {
			trip.Outcome.ExitRung = int(rung.Int64)
		}
		out = append(out, trip)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: listing the round trips of %s: %w", account, err)
	}
	return out, nil
}

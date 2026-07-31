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
	"fmt"
	"strings"

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
}

// AccountExitEvent is one exit judgement with the symbol it was made about.
type AccountExitEvent struct {
	Event  ExitEvent
	Symbol string
	Market string
}

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
			       coalesce(v.proposed_intent_id,''), v.created_at, v.saved_snapshot_json,
			       v.recomputed_snapshot_json, v.effective_snapshot_json,
			       coalesce(v.effective_source,''), coalesce(v.arm_suppressed_reason,''),
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
		var saved, recomputed, effective sql.NullString
		if err := rows.Scan(&e.Event.ID, &e.Event.PositionID, &e.Event.ObservedPrice,
			&e.Event.HighWater, &e.Event.BaselineAfter, &e.Event.LevelAfter, &e.Event.Action,
			&e.Event.ProposedIntentID, &e.Event.CreatedAt, &saved, &recomputed, &effective,
			&e.Event.Evaluation.EffectiveSource, &e.Event.ArmSuppressedReason,
			&e.Symbol, &e.Market); err != nil {
			return nil, fmt.Errorf("journal: reading an exit event of %s: %w", account, err)
		}
		// Event corruption is represented on the individual view. The account
		// history remains readable and never recomputes a replacement value.
		_ = decodeExitEventSnapshot(saved, &e.Event.Evaluation.Saved, "no_saved_evaluation")
		_ = decodeExitEventSnapshot(recomputed, &e.Event.Evaluation.Recomputed, "legacy_event")
		_ = decodeExitEventSnapshot(effective, &e.Event.Evaluation.Effective, "legacy_event")
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: listing the exit history of %s: %w", account, err)
	}
	return out, nil
}

// BrokerOrderIDs returns every broker order id this engine's own mutation
// attempts recorded, sorted and de-duplicated (change console-orders-screen,
// task 2.1).
//
// # What it answers and what it deliberately does not
//
// The orders screen asks one question of the ledger: was this order number one
// the engine caused to exist? mutation_attempts.broker_order_id is the whole
// answer — it is written when the broker acks a PLACE, a CANCEL or an AMEND, so
// an id in this set is an id some attempt of this engine's produced. An id that
// is not in it is either somebody's manual order or an attempt the broker never
// acked, and the screen says "그 밖" rather than naming a person.
//
// It is NOT scoped by account, unlike every other read in this file. The other
// four project an account's positions and their history, where mixing two
// accounts would be a wrong screen; this one compares against order numbers the
// broker issued, which are the broker's own identifiers. Scoping would mean
// joining through intents.account_ref, and a spelling difference between that
// column and the account the console is looking at would silently drop rows —
// which on this screen renders as "the engine placed none of these".
//
// The empty id is excluded rather than returned. broker_order_id defaults to ”
// for an attempt that never got an ack, and an empty entry would match every
// order whose id the payload omitted.
func (r *ReadOnly) BrokerOrderIDs(ctx context.Context) ([]string, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT DISTINCT broker_order_id FROM mutation_attempts
		  WHERE broker_order_id <> ''
		  ORDER BY broker_order_id`)
	if err != nil {
		return nil, fmt.Errorf("journal: listing the broker order ids in %s: %w", r.path, err)
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("journal: reading a broker order id: %w", err)
		}
		out = append(out, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: listing the broker order ids in %s: %w", r.path, err)
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

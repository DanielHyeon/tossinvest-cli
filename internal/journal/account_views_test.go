package journal

// account_views_test.go pins the three account-scoped reads the operator console
// needs (change add-operator-dashboard, task 1.2).
//
// They are new rather than reused because the existing readers answer different
// questions. OpenExitStates returns the observation loop's working set — only
// positions whose policy is still running — and a dashboard that used it would
// silently drop every position the engine is not managing, which is precisely the
// set the operator opened the screen to see. ExitEvents is keyed by position, and
// the console has no position id until it has done this join.

import (
	"context"
	"path/filepath"
	"testing"
)

// seedAccountViews builds one account with the four shapes the console has to
// render: an engine position with an exit state, an engine position without one,
// an externally acquired position (no entry decision), and a closed round trip.
func seedAccountViews(t *testing.T, j *Journal) {
	t.Helper()
	ctx := context.Background()

	insertDecision(t, j, "d-1", "nonce-1")

	exec := func(what, query string, args ...any) {
		t.Helper()
		if _, err := j.db.ExecContext(ctx, query, args...); err != nil {
			t.Fatalf("%s: %v", what, err)
		}
	}

	// Managed: entry decision + exit state.
	exec("position managed", `INSERT INTO positions
		  (id, account_ref, market, symbol, instance_seq, entry_decision_id, state,
		   quantity, avg_price, opened_at)
		VALUES ('p-managed','acct-1','kr','005930',1,'d-1',?,'10','70000','2026-03-30T00:30:00Z')`,
		PositionOpen)
	exec("exit state", `INSERT INTO exit_states
		  (position_id, policy_kind, entry_price, initial_stop, initial_risk,
		   baseline_price, high_water, ratchet_level, active_rung, taken_ratio_total,
		   pending_action, pending_level, pending_intent_id, completed, updated_at)
		VALUES ('p-managed',?,'70000','68000','2000','69000','74000',?,NULL,'0.25',
		        'TAKE_PROFIT','HALF_RISK','int-9',0,'2026-03-30T01:00:00Z')`,
		ExitPolicyRatchet, RatchetHalfRisk)

	// Managed by decision, but no exit state row yet.
	exec("position unarmed", `INSERT INTO positions
		  (id, account_ref, market, symbol, instance_seq, entry_decision_id, state,
		   quantity, avg_price, opened_at)
		VALUES ('p-unarmed','acct-1','kr','000660',1,'d-1',?,'3','120000','2026-03-30T00:35:00Z')`,
		PositionOpen)

	// External: no entry decision, therefore not an exit-policy target.
	exec("position external", `INSERT INTO positions
		  (id, account_ref, market, symbol, instance_seq, entry_decision_id, state,
		   quantity, avg_price, opened_at)
		VALUES ('p-external','acct-1','kr','035420',1,NULL,?,'7','200000','2026-03-29T05:00:00Z')`,
		PositionOpen)

	// Closed, with a frozen outcome and an exit state that survives it.
	exec("position closed", `INSERT INTO positions
		  (id, account_ref, market, symbol, instance_seq, entry_decision_id, state,
		   quantity, avg_price, opened_at, closed_at)
		VALUES ('p-closed','acct-1','kr','005380',1,'d-1',?,'0','0',
		        '2026-03-28T00:30:00Z','2026-03-28T06:30:00Z')`,
		PositionClosed)
	exec("exit state closed", `INSERT INTO exit_states
		  (position_id, policy_kind, entry_price, initial_stop, initial_risk,
		   baseline_price, high_water, ratchet_level, taken_ratio_total, completed, updated_at)
		VALUES ('p-closed',?,'55000','53000','2000','56000','60000',?,'1',1,'2026-03-28T06:30:00Z')`,
		ExitPolicyRatchet, RatchetProfitLock)
	exec("outcome closed", `INSERT INTO trade_outcomes
		  (position_id, realized_pnl_after_costs, realized_r, initial_risk,
		   initial_quantity, held_seconds, exit_ratchet_level, exit_rung, closed_at)
		VALUES ('p-closed','48000','2.4','2000','10',21600,?,NULL,'2026-03-28T06:30:00Z')`,
		RatchetProfitLock)

	// A round trip whose hold time is unknown: held_seconds is nullable and the
	// screen has to say so rather than print a zero.
	exec("position untimed", `INSERT INTO positions
		  (id, account_ref, market, symbol, instance_seq, entry_decision_id, state,
		   quantity, avg_price, closed_at)
		VALUES ('p-untimed','acct-1','kr','051910',1,'d-1',?,'0','0','2026-03-27T06:30:00Z')`,
		PositionClosed)
	exec("outcome untimed", `INSERT INTO trade_outcomes
		  (position_id, realized_pnl_after_costs, realized_r, initial_risk,
		   initial_quantity, held_seconds, exit_ratchet_level, exit_rung, closed_at)
		VALUES ('p-untimed','-9000','-0.9','2000','5',NULL,NULL,2,'2026-03-27T06:30:00Z')`)

	// Another account, to prove the scoping.
	exec("position other account", `INSERT INTO positions
		  (id, account_ref, market, symbol, instance_seq, state, quantity, avg_price)
		VALUES ('p-other','acct-2','kr','005930',1,?,'1','70000')`, PositionOpen)

	// Exit history, deliberately interleaved across two positions.
	exec("event 1", `INSERT INTO exit_events
		  (position_id, observed_price, high_water, baseline_after, level_after, action,
		   proposed_intent_id, created_at)
		VALUES ('p-managed','71000','71000','68000',?,'','','2026-03-30T00:40:00Z')`, RatchetNone)
	exec("event 2", `INSERT INTO exit_events
		  (position_id, observed_price, high_water, baseline_after, level_after, action,
		   proposed_intent_id, created_at)
		VALUES ('p-closed','58000','60000','56000',?,'TAKE_PROFIT','int-3','2026-03-28T04:00:00Z')`,
		RatchetProfitLock)
	exec("event 3", `INSERT INTO exit_events
		  (position_id, observed_price, high_water, baseline_after, level_after, action,
		   proposed_intent_id, created_at)
		VALUES ('p-managed','74000','74000','69000',?,'RATCHET','int-9','2026-03-30T01:00:00Z')`,
		RatchetHalfRisk)
	exec("event other", `INSERT INTO exit_events
		  (position_id, observed_price, high_water, baseline_after, level_after, action,
		   proposed_intent_id, created_at)
		VALUES ('p-other','70000','70000','69000',?,'','','2026-03-30T01:00:00Z')`, RatchetNone)
}

func seededReadOnly(t *testing.T) *ReadOnly {
	t.Helper()
	path := filepath.Join(t.TempDir(), DBFileName)
	seedAccountViews(t, openTestJournalAt(t, path))
	return openTestReadOnly(t, path)
}

// TestLivePositionExitsCarriesEveryLivePositionAndItsEligibility.
//
// The set is "what the account is holding according to the projection", not "what
// the exit policy is managing". A position with no entry decision is exactly the
// row the screen has to mark 관리 외(미편입), so dropping it would remove the one
// thing the operator most needs to see.
func TestLivePositionExitsCarriesEveryLivePositionAndItsEligibility(t *testing.T) {
	ro := seededReadOnly(t)

	rows, err := ro.LivePositionExits(context.Background(), "acct-1")
	if err != nil {
		t.Fatalf("LivePositionExits: %v", err)
	}

	byID := map[string]PositionExit{}
	for _, r := range rows {
		byID[r.Position.ID] = r
	}
	if len(rows) != 3 {
		t.Fatalf("got %d live position(s), want 3 (managed, unarmed, external): %v", len(rows), byID)
	}
	if _, closed := byID["p-closed"]; closed {
		t.Error("a CLOSED instance is on the live list; the history screen owns those")
	}
	if _, other := byID["p-other"]; other {
		t.Error("another account's position leaked into the list")
	}

	managed := byID["p-managed"]
	if !managed.HasExit {
		t.Fatal("the managed position has no exit state attached")
	}
	if !managed.Position.ExitEligible() {
		t.Error("a position with an entry decision must be exit-eligible")
	}
	if managed.Exit.Baseline != "69000" || managed.Exit.HighWater != "74000" {
		t.Errorf("exit line = baseline %q / high water %q, want 69000 / 74000",
			managed.Exit.Baseline, managed.Exit.HighWater)
	}
	if managed.Exit.RatchetLevel != RatchetHalfRisk {
		t.Errorf("ratchet level = %q, want %q", managed.Exit.RatchetLevel, RatchetHalfRisk)
	}
	if managed.Exit.ActiveRung != -1 {
		t.Errorf("active rung = %d, want -1 for a RATCHET position", managed.Exit.ActiveRung)
	}
	if managed.Exit.TakenRatioTotal != "0.25" {
		t.Errorf("taken ratio = %q, want 0.25", managed.Exit.TakenRatioTotal)
	}
	if !managed.Exit.Pending() || managed.Exit.PendingIntentID != "int-9" {
		t.Errorf("pending proposal = %+v, want the armed one", managed.Exit)
	}

	unarmed := byID["p-unarmed"]
	if unarmed.HasExit {
		t.Error("a position with no exit_states row must not report one")
	}
	if !unarmed.Position.ExitEligible() {
		t.Error("an entry decision makes a position eligible even before its exit state exists")
	}

	external := byID["p-external"]
	if external.HasExit {
		t.Error("an external position must not report an exit state")
	}
	if external.Position.ExitEligible() {
		t.Error("a position with no entry decision must not be exit-eligible")
	}
}

// TestAccountTradeTripsJoinsOnlyWhatTheSchemaHolds.
//
// The frozen row has no symbol and no entry price; both are joined in from the
// tables that do hold them. Nothing else is: the exit price has no column
// anywhere, and computing one from the fills would be the recomputation the
// freeze exists to prevent.
func TestAccountTradeTripsJoinsOnlyWhatTheSchemaHolds(t *testing.T) {
	ro := seededReadOnly(t)

	trips, err := ro.AccountTradeTrips(context.Background(), "acct-1")
	if err != nil {
		t.Fatalf("AccountTradeTrips: %v", err)
	}
	if len(trips) != 2 {
		t.Fatalf("got %d trip(s), want 2: %+v", len(trips), trips)
	}
	// Oldest close first, the order TradeOutcomes already uses.
	if trips[0].Outcome.PositionID != "p-untimed" || trips[1].Outcome.PositionID != "p-closed" {
		t.Fatalf("trips are not ordered by close time: %+v", trips)
	}

	closed := trips[1]
	if closed.Symbol != "005380" || closed.Market != "kr" {
		t.Errorf("symbol/market = %q/%q, want 005380/kr", closed.Symbol, closed.Market)
	}
	if closed.EntryPrice != "55000" {
		t.Errorf("entry price = %q, want the exit state's frozen 55000", closed.EntryPrice)
	}
	if !closed.HeldSecondsKnown || closed.Outcome.HeldSeconds != 21600 {
		t.Errorf("held seconds = %d (known %v), want 21600 known",
			closed.Outcome.HeldSeconds, closed.HeldSecondsKnown)
	}
	if closed.Outcome.RealizedR != "2.4" || closed.Outcome.RealizedPnLAfterCosts != "48000" {
		t.Errorf("the frozen numbers were not passed through: %+v", closed.Outcome)
	}
	if closed.Outcome.ExitRatchetLevel != RatchetProfitLock || closed.Outcome.ExitRung != -1 {
		t.Errorf("exit stage = %q / rung %d, want %s / -1",
			closed.Outcome.ExitRatchetLevel, closed.Outcome.ExitRung, RatchetProfitLock)
	}

	untimed := trips[0]
	if untimed.HeldSecondsKnown {
		t.Error("a NULL held_seconds must be reported as unknown, not as zero seconds")
	}
	if untimed.Symbol != "051910" {
		t.Errorf("symbol = %q, want 051910", untimed.Symbol)
	}
	// No exit state was ever opened for it, so there is no entry price to show.
	if untimed.EntryPrice != "" {
		t.Errorf("entry price = %q, want empty when no exit state froze one", untimed.EntryPrice)
	}
	if untimed.Outcome.ExitRung != 2 {
		t.Errorf("exit rung = %d, want 2", untimed.Outcome.ExitRung)
	}
}

// TestAccountExitEventsAreAccountWideAndOrdered.
//
// The existing reader takes a position id. The history screen has an account and
// wants the judgement stream across all of it, oldest first, with the symbol each
// row belongs to — otherwise a closure with no performance row (an external sell)
// shows up nowhere at all.
func TestAccountExitEventsAreAccountWideAndOrdered(t *testing.T) {
	ro := seededReadOnly(t)

	events, err := ro.AccountExitEvents(context.Background(), "acct-1", 100)
	if err != nil {
		t.Fatalf("AccountExitEvents: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("got %d event(s), want 3 (acct-2's is excluded): %+v", len(events), events)
	}
	wantOrder := []string{"2026-03-28T04:00:00Z", "2026-03-30T00:40:00Z", "2026-03-30T01:00:00Z"}
	for i, want := range wantOrder {
		if events[i].Event.CreatedAt != want {
			t.Errorf("event %d at %q, want %q", i, events[i].Event.CreatedAt, want)
		}
	}
	if events[0].Symbol != "005380" || events[1].Symbol != "005930" {
		t.Errorf("symbols = %q, %q; want 005380, 005930", events[0].Symbol, events[1].Symbol)
	}
	if events[2].Event.Action != "RATCHET" || events[2].Event.ProposedIntentID != "int-9" {
		t.Errorf("the newest event lost its proposal: %+v", events[2].Event)
	}

	// The limit keeps the newest, because a screen that truncates the oldest end
	// of a stream is showing the operator last month instead of this morning.
	newest, err := ro.AccountExitEvents(context.Background(), "acct-1", 2)
	if err != nil {
		t.Fatalf("AccountExitEvents(limit 2): %v", err)
	}
	if len(newest) != 2 {
		t.Fatalf("got %d event(s), want 2", len(newest))
	}
	if newest[0].Event.CreatedAt != wantOrder[1] || newest[1].Event.CreatedAt != wantOrder[2] {
		t.Errorf("the limit dropped the wrong end: %+v", newest)
	}
}

// TestAccountRefsNamesEveryAccountTheJournalHolds.
//
// The console has no configured account: it renders what the journal says the
// engine has been trading, and the reference is what every other read is scoped
// by.
func TestAccountRefsNamesEveryAccountTheJournalHolds(t *testing.T) {
	ro := seededReadOnly(t)

	refs, err := ro.AccountRefs(context.Background())
	if err != nil {
		t.Fatalf("AccountRefs: %v", err)
	}
	if len(refs) != 2 || refs[0] != "acct-1" || refs[1] != "acct-2" {
		t.Fatalf("AccountRefs = %v, want [acct-1 acct-2]", refs)
	}
}

// TestTheAccountViewsAreEmptyRatherThanFailingOnAnEmptyJournal.
//
// "The engine has not traded yet" is a state the dashboard reports, not an error
// it refuses to render.
func TestTheAccountViewsAreEmptyRatherThanFailingOnAnEmptyJournal(t *testing.T) {
	path := filepath.Join(t.TempDir(), DBFileName)
	openTestJournalAt(t, path)
	ro := openTestReadOnly(t, path)
	ctx := context.Background()

	refs, err := ro.AccountRefs(ctx)
	if err != nil || len(refs) != 0 {
		t.Errorf("AccountRefs = %v, %v; want no refs and no error", refs, err)
	}
	positions, err := ro.LivePositionExits(ctx, "acct-1")
	if err != nil || len(positions) != 0 {
		t.Errorf("LivePositionExits = %v, %v; want empty and no error", positions, err)
	}
	trips, err := ro.AccountTradeTrips(ctx, "acct-1")
	if err != nil || len(trips) != 0 {
		t.Errorf("AccountTradeTrips = %v, %v; want empty and no error", trips, err)
	}
	events, err := ro.AccountExitEvents(ctx, "acct-1", 10)
	if err != nil || len(events) != 0 {
		t.Errorf("AccountExitEvents = %v, %v; want empty and no error", events, err)
	}
}

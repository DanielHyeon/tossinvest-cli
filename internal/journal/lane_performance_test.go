package journal

import (
	"context"
	"strings"
	"testing"
	"time"
)

type strategySourceFixture struct {
	positionID string
	account    string
	suffix     string
	sequence   int
	closedAt   time.Time
	cost       any
	links      []struct{ kind, ref string }
}

func exactStrategySourceLinks(positionID, suffix string) []struct{ kind, ref string } {
	return []struct{ kind, ref string }{
		{kind: "MUTATION_ATTEMPT", ref: "mutation-" + suffix},
		{kind: "BROKER_ORDER", ref: "order-" + suffix},
		{kind: "FILL", ref: "fill-" + suffix},
		{kind: "POSITION", ref: positionID},
		{kind: "CLOSE_OUTCOME", ref: positionID},
	}
}

func seedStrategySource(t *testing.T, j *Journal, fixture strategySourceFixture) StrategyAtomicPlan {
	t.Helper()
	ctx := context.Background()
	plan := strategyPlanFixture(t, fixture.suffix, fixture.account)
	if _, err := j.planStrategyEntryForTest(ctx, plan); err != nil {
		t.Fatalf("plan strategy source: %v", err)
	}
	opened := fixture.closedAt.Add(-time.Hour).UTC().Format(time.RFC3339)
	closed := fixture.closedAt.UTC().Format(time.RFC3339)
	if _, err := j.db.ExecContext(ctx, `INSERT INTO positions
		(id,account_ref,market,symbol,instance_seq,entry_decision_id,state,quantity,avg_price,opened_at,closed_at)
		VALUES (?,?,?,?,?,?,?,'0','0',?,?)`,
		fixture.positionID, fixture.account, "kr", "005930", fixture.sequence,
		plan.RiskDecision.ID, PositionClosed, opened, closed); err != nil {
		t.Fatalf("insert source position: %v", err)
	}
	if _, err := j.db.ExecContext(ctx, `INSERT INTO exit_states
		(position_id,policy_kind,policy_id,entry_price,initial_stop,initial_risk,
		 baseline_price,high_water,ratchet_level,taken_ratio_total,completed,updated_at,
		 snapshot_status,policy_version)
		VALUES (?,?,'ratchet-v1','100.2','99.3993','0.8007','99.3993','103',?,'1',1,?,'SEED','ratchet/v1')`,
		fixture.positionID, ExitPolicyRatchet, RatchetProfitLock, closed); err != nil {
		t.Fatalf("insert source exit state: %v", err)
	}
	if _, err := j.db.ExecContext(ctx, `INSERT INTO trade_outcomes
		(position_id,realized_pnl_after_costs,realized_r,initial_risk,initial_quantity,
		 held_seconds,exit_ratchet_level,exit_rung,closed_at,cost_total)
		VALUES (?,'1.5','1.87336','0.8007','1',3600,?,NULL,?,?)`,
		fixture.positionID, RatchetProfitLock, closed, fixture.cost); err != nil {
		t.Fatalf("insert source outcome: %v", err)
	}
	for _, link := range fixture.links {
		if err := j.AppendStrategyExecutionLink(ctx, fixture.account, plan.AttemptID, link.kind, link.ref); err != nil {
			t.Fatalf("append %s source link: %v", link.kind, err)
		}
	}
	return plan
}

func TestClosedStrategyTradeSourcesReturnsExactIdentifiersFactsAndNullableCost(t *testing.T) {
	j := openTestJournal(t)
	closed := time.Date(2026, 3, 30, 2, 0, 0, 0, time.UTC)
	links := exactStrategySourceLinks("position-exact", "exact")
	plan := seedStrategySource(t, j, strategySourceFixture{
		positionID: "position-exact", account: "acct-1", suffix: "exact", sequence: 1,
		closedAt: closed, cost: "0.25", links: links,
	})
	seedStrategySource(t, j, strategySourceFixture{
		positionID: "position-other", account: "acct-2", suffix: "other", sequence: 1,
		closedAt: closed, cost: nil, links: []struct{ kind, ref string }{
			{kind: "MUTATION_ATTEMPT", ref: "mutation-exact"},
			{kind: "BROKER_ORDER", ref: "order-exact"},
			{kind: "FILL", ref: "fill-exact"},
			{kind: "POSITION", ref: "position-other"},
			{kind: "CLOSE_OUTCOME", ref: "position-other"},
		},
	})

	rows, err := openTestReadOnly(t, j.Path()).ClosedStrategyTradeSources(
		context.Background(), "acct-1", closed.Add(-time.Second), closed)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("rows=%+v, want exact account-scoped closed trade", rows)
	}
	got := rows[0]
	if got.TradeID != "position-exact" || got.PositionID != "position-exact" || got.CloseID != "position-exact" ||
		got.Market != "kr" || got.Side != "BUY" || got.DecisionPrice != "100.1" || got.EntryPrice != "100.2" ||
		got.Quantity != "1" || got.RealizedPnLAfterCosts != "1.5" || got.RealizedR != "1.87336" ||
		got.PolicyID != "ratchet-v1" || got.PolicyVersion != "ratchet/v1" || !got.ClosedAt.Equal(closed) {
		t.Fatalf("source facts=%+v", got)
	}
	if got.CostTotal == nil || *got.CostTotal != "0.25" {
		t.Fatalf("cost_total=%v", got.CostTotal)
	}
	if got.Lineage == nil {
		t.Fatal("exact persisted strategy chain was marked link_missing")
	}
	want := got.Lineage
	if want.StrategyDecisionIdentity != plan.Lineage.DecisionIdentity || want.RiskIntentID != plan.RiskDecision.ID ||
		want.StrategyAttemptID != plan.AttemptID || want.MutationAttemptID != "mutation-exact" ||
		want.BrokerOrderID != "order-exact" || want.FillID != "fill-exact" ||
		want.PositionID != "position-exact" || want.CloseOutcomeID != "position-exact" {
		t.Fatalf("exact chain=%+v", want)
	}

	// The inclusive upper bound did not turn SQL text into a string-built query;
	// an account value containing syntax is simply another account with no rows.
	injected, err := openTestReadOnly(t, j.Path()).ClosedStrategyTradeSources(
		context.Background(), `acct-1' OR 1=1 --`, closed.Add(-time.Hour), closed)
	if err != nil || len(injected) != 0 {
		t.Fatalf("bound account query rows=%+v err=%v", injected, err)
	}
}

func TestClosedStrategyTradeSourcesUsesExclusiveLowerInclusiveUpperAndStableOrder(t *testing.T) {
	j := openTestJournal(t)
	base := time.Date(2026, 3, 30, 1, 0, 0, 0, time.UTC)
	for index, fixture := range []struct {
		id     string
		suffix string
		closed time.Time
	}{
		{id: "lower", suffix: "lower", closed: base},
		{id: "same-z", suffix: "same-z", closed: base.Add(time.Hour)},
		{id: "same-a", suffix: "same-a", closed: base.Add(time.Hour)},
		{id: "after", suffix: "after", closed: base.Add(2 * time.Hour)},
	} {
		seedStrategySource(t, j, strategySourceFixture{
			positionID: fixture.id, account: "acct-1", suffix: fixture.suffix, sequence: index + 1,
			closedAt: fixture.closed, cost: nil, links: exactStrategySourceLinks(fixture.id, fixture.suffix),
		})
	}
	rows, err := openTestReadOnly(t, j.Path()).ClosedStrategyTradeSources(
		context.Background(), "acct-1", base, base.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[0].PositionID != "same-a" || rows[1].PositionID != "same-z" {
		t.Fatalf("window/order rows=%+v", rows)
	}
	for _, row := range rows {
		if row.CostTotal != nil {
			t.Fatalf("legacy/not-measured cost was coerced to %v", row.CostTotal)
		}
	}
}

func TestClosedStrategyTradeSourcesNeverChoosesAZeroOrMultiFillOrDistractor(t *testing.T) {
	j := openTestJournal(t)
	closed := time.Date(2026, 3, 30, 2, 0, 0, 0, time.UTC)
	noFill := exactStrategySourceLinks("no-fill", "no-fill")
	noFill = noFill[:2+copy(noFill[2:], noFill[3:])]
	seedStrategySource(t, j, strategySourceFixture{
		positionID: "no-fill", account: "acct-1", suffix: "no-fill", sequence: 1,
		closedAt: closed, cost: nil, links: noFill,
	})
	multiFill := exactStrategySourceLinks("multi-fill", "multi-fill")
	multiFill = append(multiFill, struct{ kind, ref string }{kind: "FILL", ref: "fill-distractor"})
	seedStrategySource(t, j, strategySourceFixture{
		positionID: "multi-fill", account: "acct-1", suffix: "multi-fill", sequence: 2,
		closedAt: closed.Add(time.Second), cost: nil, links: multiFill,
	})
	seedStrategySource(t, j, strategySourceFixture{
		positionID: "complete", account: "acct-1", suffix: "complete", sequence: 3,
		closedAt: closed.Add(2 * time.Second), cost: nil, links: exactStrategySourceLinks("complete", "complete"),
	})

	rows, err := openTestReadOnly(t, j.Path()).ClosedStrategyTradeSources(
		context.Background(), "acct-1", closed.Add(-time.Second), closed.Add(3*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 3 {
		t.Fatalf("rows=%+v", rows)
	}
	byID := map[string]ClosedStrategyTradeSource{}
	for _, row := range rows {
		byID[row.PositionID] = row
	}
	if byID["no-fill"].Lineage != nil || byID["multi-fill"].Lineage != nil {
		t.Fatalf("ambiguous cardinality was guessed: no-fill=%+v multi-fill=%+v",
			byID["no-fill"].Lineage, byID["multi-fill"].Lineage)
	}
	if byID["complete"].Lineage == nil || byID["complete"].Lineage.FillID != "fill-complete" {
		t.Fatalf("complete lineage=%+v", byID["complete"].Lineage)
	}
}

func TestClosedStrategyTradeSourcesRejectsInvalidWindowAndPersistedData(t *testing.T) {
	j := openTestJournal(t)
	closed := time.Date(2026, 3, 30, 2, 0, 0, 0, time.UTC)
	seedStrategySource(t, j, strategySourceFixture{
		positionID: "invalid-time", account: "acct-1", suffix: "invalid-time", sequence: 1,
		closedAt: closed, cost: nil, links: exactStrategySourceLinks("invalid-time", "invalid-time"),
	})
	ro := openTestReadOnly(t, j.Path())
	for _, window := range [][2]time.Time{
		{{}, closed}, {closed, closed}, {closed, closed.Add(-time.Second)},
	} {
		if _, err := ro.ClosedStrategyTradeSources(context.Background(), "acct-1", window[0], window[1]); err == nil {
			t.Fatalf("invalid window %+v was accepted", window)
		}
	}
	if _, err := j.db.Exec(`UPDATE positions SET opened_at='not-a-time' WHERE id='invalid-time'`); err != nil {
		t.Fatal(err)
	}
	_, err := ro.ClosedStrategyTradeSources(context.Background(), "acct-1", closed.Add(-time.Hour), closed)
	if err == nil || !strings.Contains(err.Error(), "opened_at") {
		t.Fatalf("invalid persisted time err=%v", err)
	}
	if _, err := j.db.Exec(`UPDATE positions SET opened_at='2026-03-30T01:00:00Z', entry_decision_id=NULL WHERE id='invalid-time'`); err != nil {
		t.Fatal(err)
	}
	if _, err := ro.ClosedStrategyTradeSources(context.Background(), "acct-1", closed.Add(-time.Hour), closed); err == nil {
		t.Fatal("an outcome with no exact risk decision was silently omitted instead of rejected")
	}
}

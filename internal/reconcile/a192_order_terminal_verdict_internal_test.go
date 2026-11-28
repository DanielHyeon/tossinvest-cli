package reconcile

// a192_order_terminal_verdict_internal_test.go 는 종결 증명 해제의 판정 함수 하나를 조건(C1~C13)마다
// 직접 찌름(a192 design D3). 픽스처는 합성 값만 씀.

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

const (
	a192Account = "acct-7"
	a192Symbol  = "SYN001"
	a192Row     = "rec-a192-row"
)

func a192Order(id string) boundOrder {
	return boundOrder{
		identity: OrderIdentity{AccountRef: a192Account, Market: "kr", TradingDay: "2026-03-30",
			Symbol: a192Symbol, Side: "SELL", OrderID: id},
		original: id,
	}
}

// a192Binding 은 통과하는 결속 하나(무장 안 함).
func a192Binding(orders ...boundOrder) terminalBinding {
	b := terminalBinding{rowID: a192Row, symbol: a192Symbol, orders: map[boundOrder]struct{}{}, fillWatermark: 5}
	for _, order := range orders {
		b.orders[order] = struct{}{}
	}
	return b
}

func a192Proof(b terminalBinding) journal.OrderTerminalProof {
	proof := journal.OrderTerminalProof{
		StateFound: true,
		State: journal.ReconcileState{ID: a192Row, AccountRef: a192Account, Symbol: a192Symbol,
			Cause: journal.ReconcileCauseQuantityMismatch, Evidence: "order on SYN001 is not in the account's open list"},
	}
	for _, order := range sortedBoundOrders(b) {
		proof.Snapshots = append(proof.Snapshots, journal.OrderTerminalSnapshot{
			Scope: journal.FillSnapshotScope{OrderID: order.original}, Found: true,
			Record: journal.FillSnapshotRecord{OrderID: order.original, State: "CANCELLED", Terminal: true,
				FilledQuantity: "0", CommittedAt: "2026-03-30T01:31:00Z"},
		})
	}
	return proof
}

func a192Comparison(asOf string) terminalComparison {
	return terminalComparison{trackerAccount: a192Account, diffAccount: a192Account, asOf: asOf,
		externalIDs: map[string]bool{}}
}

// TestA192VerdictArmsThenReleases: 모두 통과하면 처음엔 무장, 같은 증명으로 더 나중 비교면 해제.
func TestA192VerdictArmsThenReleases(t *testing.T) {
	b := a192Binding(a192Order("o-a192-1"))
	proof := a192Proof(b)
	first := orderTerminalReleaseVerdict(b, a192Comparison("2026-03-30T01:32:00Z"), proof)
	if first.outcome != terminalArm || first.reason != "armed" || first.digest == "" {
		t.Fatalf("first passing verdict = %+v, want armed with a digest", first)
	}
	b.armed, b.armAsOf, b.armDigest = true, first.asOf, first.digest
	same := orderTerminalReleaseVerdict(b, a192Comparison("2026-03-30T01:32:00Z"), proof)
	if same.outcome != terminalHold || same.reason != "arm_not_later" {
		t.Fatalf("same-instant verdict = %+v, want arm_not_later", same)
	}
	later := orderTerminalReleaseVerdict(b, a192Comparison("2026-03-30T01:33:00Z"), proof)
	if later.outcome != terminalReleaseOutcome || later.boundOrders != 1 || later.latestCommit != "2026-03-30T01:31:00Z" {
		t.Fatalf("later verdict = %+v, want release of one bound order", later)
	}
}

// TestA192VerdictRearmsWhenTheProofChanges: 무장 뒤 증명(스냅숏·결속)이 바뀌면 해제하지 않고 다시 무장함.
func TestA192VerdictRearmsWhenTheProofChanges(t *testing.T) {
	b := a192Binding(a192Order("o-a192-1"))
	proof := a192Proof(b)
	first := orderTerminalReleaseVerdict(b, a192Comparison("2026-03-30T01:32:00Z"), proof)
	b.armed, b.armAsOf, b.armDigest = true, first.asOf, first.digest

	changed := a192Proof(b)
	changed.Snapshots[0].Record.CommittedAt = "2026-03-30T01:31:30Z"
	again := orderTerminalReleaseVerdict(b, a192Comparison("2026-03-30T01:33:00Z"), changed)
	if again.outcome != terminalArm || again.digest == first.digest {
		t.Fatalf("verdict after a rewritten snapshot = %+v, want a fresh arm", again)
	}

	grown := a192Binding(a192Order("o-a192-1"), a192Order("o-a192-2"))
	grown.armed, grown.armAsOf, grown.armDigest = true, first.asOf, first.digest
	added := orderTerminalReleaseVerdict(grown, a192Comparison("2026-03-30T01:33:00Z"), a192Proof(grown))
	if added.outcome != terminalArm {
		t.Fatalf("verdict after a new bound order = %+v, want a fresh arm", added)
	}
}

// TestA192VerdictHoldsForEveryFailedCondition 은 조건마다 처음 실패한 사유를 고정함.
func TestA192VerdictHoldsForEveryFailedCondition(t *testing.T) {
	base := func() (terminalBinding, terminalComparison, journal.OrderTerminalProof) {
		b := a192Binding(a192Order("o-a192-1"))
		return b, a192Comparison("2026-03-30T01:32:00Z"), a192Proof(b)
	}
	cases := []struct {
		name   string
		mutate func(*terminalBinding, *terminalComparison, *journal.OrderTerminalProof)
		reason string
	}{
		{"no insertion proof", func(b *terminalBinding, _ *terminalComparison, _ *journal.OrderTerminalProof) {
			b.rowID = ""
		}, "not_inserted_by_this_process"},
		{"read failure", func(_ *terminalBinding, _ *terminalComparison, p *journal.OrderTerminalProof) {
			p.ReadErr = errors.New("disk")
		}, "terminal_read_failed"},
		{"row gone", func(_ *terminalBinding, _ *terminalComparison, p *journal.OrderTerminalProof) {
			p.StateFound = false
		}, "row_changed"},
		{"row released", func(_ *terminalBinding, _ *terminalComparison, p *journal.OrderTerminalProof) {
			p.State.ReleaseCause = journal.ReconcileReleaseOperator
		}, "row_changed"},
		{"another row id", func(_ *terminalBinding, _ *terminalComparison, p *journal.OrderTerminalProof) {
			p.State.ID = "rec-other"
		}, "row_changed"},
		{"another cause", func(_ *terminalBinding, _ *terminalComparison, p *journal.OrderTerminalProof) {
			p.State.Cause = journal.ReconcileCauseIdentifierConflict
		}, "row_changed"},
		{"market scoped row", func(_ *terminalBinding, _ *terminalComparison, p *journal.OrderTerminalProof) {
			p.State.ScopeMarket = "KR"
		}, "row_changed"},
		{"row of another account", func(_ *terminalBinding, _ *terminalComparison, p *journal.OrderTerminalProof) {
			p.State.AccountRef = "acct-other"
		}, "row_changed"},
		{"quantity dispute seen", func(b *terminalBinding, _ *terminalComparison, _ *journal.OrderTerminalProof) {
			b.tainted = true
		}, "quantity_dispute_seen"},
		{"no bound order", func(b *terminalBinding, _ *terminalComparison, p *journal.OrderTerminalProof) {
			b.orders = map[boundOrder]struct{}{}
			p.Snapshots = nil
		}, "identity_incomplete"},
		{"blank side", func(b *terminalBinding, _ *terminalComparison, _ *journal.OrderTerminalProof) {
			order := a192Order("o-a192-1")
			order.identity.Side = ""
			b.orders = map[boundOrder]struct{}{order: {}}
		}, "identity_incomplete"},
		{"foreign order account", func(b *terminalBinding, _ *terminalComparison, _ *journal.OrderTerminalProof) {
			order := a192Order("o-a192-1")
			order.identity.AccountRef = "acct-other"
			b.orders = map[boundOrder]struct{}{order: {}}
		}, "account_mismatch"},
		{"foreign comparison account", func(_ *terminalBinding, c *terminalComparison, _ *journal.OrderTerminalProof) {
			c.diffAccount = ""
		}, "account_mismatch"},
		{"order of another symbol", func(b *terminalBinding, _ *terminalComparison, _ *journal.OrderTerminalProof) {
			order := a192Order("o-a192-1")
			order.identity.Symbol = "SYNX"
			b.orders = map[boundOrder]struct{}{order: {}}
		}, "account_mismatch"},
		{"proof out of step", func(_ *terminalBinding, _ *terminalComparison, p *journal.OrderTerminalProof) {
			p.Snapshots[0].Scope.OrderID = "o-a192-x"
		}, "proof_mismatch"},
		{"lineage moved in memory", func(b *terminalBinding, _ *terminalComparison, _ *journal.OrderTerminalProof) {
			order := a192Order("o-a192-1")
			order.identity.OrderID = "o-a192-child"
			b.orders = map[boundOrder]struct{}{order: {}}
		}, "lineage_moved"},
		{"replacement trace", func(_ *terminalBinding, _ *terminalComparison, p *journal.OrderTerminalProof) {
			p.Snapshots[0].Amended = true
		}, "lineage_moved"},
		{"no snapshot", func(_ *terminalBinding, _ *terminalComparison, p *journal.OrderTerminalProof) {
			p.Snapshots[0].Found = false
		}, "terminal_not_found"},
		{"not terminal", func(_ *terminalBinding, _ *terminalComparison, p *journal.OrderTerminalProof) {
			p.Snapshots[0].Record.Terminal = false
		}, "not_clean_terminal"},
		{"terminal but fail closed", func(_ *terminalBinding, _ *terminalComparison, p *journal.OrderTerminalProof) {
			p.Snapshots[0].Record.FailClosed = true
		}, "not_clean_terminal"},
		{"unreadable commit", func(_ *terminalBinding, _ *terminalComparison, p *journal.OrderTerminalProof) {
			p.Snapshots[0].Record.CommittedAt = "soon"
		}, "comparison_not_after_terminal"},
		{"comparison at the commit", func(_ *terminalBinding, c *terminalComparison, _ *journal.OrderTerminalProof) {
			c.asOf = "2026-03-30T01:31:00Z"
		}, "comparison_not_after_terminal"},
		{"comparison before the commit", func(_ *terminalBinding, c *terminalComparison, _ *journal.OrderTerminalProof) {
			c.asOf = "2026-03-30T01:30:00Z"
		}, "comparison_not_after_terminal"},
		{"undated comparison", func(_ *terminalBinding, c *terminalComparison, _ *journal.OrderTerminalProof) {
			c.asOf = ""
		}, "comparison_not_after_terminal"},
		{"symbol disputed", func(_ *terminalBinding, c *terminalComparison, _ *journal.OrderTerminalProof) {
			c.disputed = true
		}, "symbol_disputed"},
		{"same id still open", func(_ *terminalBinding, c *terminalComparison, _ *journal.OrderTerminalProof) {
			c.externalIDs["o-a192-1"] = true
		}, "terminal_order_still_open"},
		{"comparison blocks entry", func(_ *terminalBinding, c *terminalComparison, _ *journal.OrderTerminalProof) {
			c.blocksEntry = true
		}, "comparison_blocks_entry"},
		{"fill after the block", func(_ *terminalBinding, _ *terminalComparison, p *journal.OrderTerminalProof) {
			p.FillAfterWatermark = true
		}, "fill_after_block"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, c, p := base()
			tc.mutate(&b, &c, &p)
			got := orderTerminalReleaseVerdict(b, c, p)
			if got.outcome != terminalHold || got.reason != tc.reason {
				t.Fatalf("verdict = %+v, want hold %q", got, tc.reason)
			}
		})
	}
	// 대조군: 손대지 않은 입력은 통과(무장)해야 위의 사유들이 의미가 있음.
	b, c, p := base()
	if got := orderTerminalReleaseVerdict(b, c, p); got.outcome != terminalArm {
		t.Fatalf("unmutated verdict = %+v, want armed", got)
	}
}

// TestA192VerdictIDsAreComparedByteForByte: 공백이 다른 외부 주문 번호는 같은 주문이 아님(불투명 식별자).
func TestA192VerdictIDsAreComparedByteForByte(t *testing.T) {
	b := a192Binding(a192Order("o-a192-1"))
	c := a192Comparison("2026-03-30T01:32:00Z")
	c.externalIDs[" o-a192-1"] = true
	if got := orderTerminalReleaseVerdict(b, c, a192Proof(b)); got.outcome != terminalArm {
		t.Fatalf("verdict with a differently spelled open id = %+v, want armed", got)
	}
}

// TestA192VerdictRequiresEveryBoundOrder: 결속 주문 둘 중 하나만 종결이면 유지.
func TestA192VerdictRequiresEveryBoundOrder(t *testing.T) {
	b := a192Binding(a192Order("o-a192-1"), a192Order("o-a192-2"))
	proof := a192Proof(b)
	if len(proof.Snapshots) != 2 {
		t.Fatalf("snapshots = %d, want two bound orders", len(proof.Snapshots))
	}
	proof.Snapshots[1].Record.Terminal = false
	if got := orderTerminalReleaseVerdict(b, a192Comparison("2026-03-30T01:32:00Z"), proof); got.reason != "not_clean_terminal" {
		t.Fatalf("verdict with one open bound order = %+v, want not_clean_terminal", got)
	}
}

// TestA192VerdictUsesTheLatestTerminalCommit: as-of 는 가장 늦은 종결 기록보다 나중이어야 함.
func TestA192VerdictUsesTheLatestTerminalCommit(t *testing.T) {
	b := a192Binding(a192Order("o-a192-1"), a192Order("o-a192-2"))
	proof := a192Proof(b)
	proof.Snapshots[1].Record.CommittedAt = "2026-03-30T01:35:00Z"
	got := orderTerminalReleaseVerdict(b, a192Comparison("2026-03-30T01:32:00Z"), proof)
	if got.reason != "comparison_not_after_terminal" {
		t.Fatalf("verdict before the later commit = %+v, want comparison_not_after_terminal", got)
	}
	if _, err := time.Parse(time.RFC3339, proof.Snapshots[1].Record.CommittedAt); err != nil {
		t.Fatal(err)
	}
}

// a192FakeTerminal 은 늘 통과하는 증명을 돌려주는 장부 대역 — 호출자 배관만 재려는 것.
type a192FakeTerminal struct{ proof journal.OrderTerminalProof }

func (f a192FakeTerminal) MaxFillEventID(context.Context) (int64, error) { return 0, nil }
func (f a192FakeTerminal) ReadOrderTerminalProof(context.Context, journal.OrderTerminalProofRequest) journal.OrderTerminalProof {
	return f.proof
}
func (f a192FakeTerminal) ReleaseReconcileOnOrderTerminalProof(context.Context, journal.OrderTerminalProofRequest,
	string, func(journal.OrderTerminalProof) bool) (journal.ReconcileState, bool, error) {
	return journal.ReconcileState{}, false, nil
}

// a192NoopStore 는 Journal 이 nil 이 아님만 만족시키는 대역.
type a192NoopStore struct{ ReconcileStore }

// TestA192CandidateRefusesWithoutTerminalOrJournal 은 후보 판정의 첫 가드를 못 박음: 결속이 기억에 있어도 Terminal 이나
// Journal 이 없으면 판정하지 않음. (운영에서는 결속 생성이 둘 다를 요구해 이 가드가 겹쳐 있음 — 그 우연에 기대지 않게 고정.)
func TestA192CandidateRefusesWithoutTerminalOrJournal(t *testing.T) {
	b := a192Binding(a192Order("o-a192-1"))
	for name, tr := range map[string]*Tracker{
		"no terminal": {AccountRef: a192Account, Journal: a192NoopStore{}},
		"no journal":  {AccountRef: a192Account, Terminal: a192FakeTerminal{proof: a192Proof(b)}},
	} {
		t.Run(name, func(t *testing.T) {
			copied := b
			tr.terminal.bindings = map[string]*terminalBinding{a192Symbol: &copied}
			block := Block{Symbol: a192Symbol}
			diff := Diff{AsOf: "2026-03-30T01:32:00Z", AccountRef: a192Account}
			if tr.orderTerminalCandidate(context.Background(), &block, diff, map[string]bool{}) {
				t.Fatal("the candidate judged without its journal")
			}
			if copied.armed {
				t.Fatal("the candidate armed without its journal")
			}
		})
	}
}

// TestA192CandidatePassesTheDisputedSymbolThrough 는 호출자 배관을 못 박음: Observe 가 넘긴 분쟁 집합이 판정에 그대로
// 감 — 그 종목이 분쟁 중이면 무장하지 않음.
func TestA192CandidatePassesTheDisputedSymbolThrough(t *testing.T) {
	b := a192Binding(a192Order("o-a192-1"))
	tr := &Tracker{AccountRef: a192Account, Journal: a192NoopStore{}, Terminal: a192FakeTerminal{proof: a192Proof(b)}}
	copied := b
	tr.terminal.bindings = map[string]*terminalBinding{a192Symbol: &copied}
	block := Block{Symbol: a192Symbol}
	diff := Diff{AsOf: "2026-03-30T01:32:00Z", AccountRef: a192Account}
	tr.orderTerminalCandidate(context.Background(), &block, diff, map[string]bool{a192Symbol: true})
	if copied.armed {
		t.Fatal("a disputed symbol was armed — the disputed set did not reach the verdict")
	}
	// 대조: 분쟁이 없으면 같은 입력이 무장함(판정 경로가 산다).
	tr.orderTerminalCandidate(context.Background(), &block, diff, map[string]bool{})
	if !copied.armed {
		t.Fatal("the undisputed control did not arm")
	}
}

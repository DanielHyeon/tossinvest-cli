package positioncampaign

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAggregateIdentityAndImmutableLegLineage(t *testing.T) {
	aggregate := PositionCampaign{
		ID: "campaign", AccountRef: "acct", Market: "us", Symbol: "AAPL",
		LaneID: "us-swing", LaneVersion: "v1", DecisionID: "decision", EvidenceDigest: "sha256:evidence",
		ProspectiveToken: "token", ExpectedPositionGeneration: 3,
		Legs: []CampaignLeg{{CampaignID: "campaign", Sequence: 1, PlanID: "plan-1", IntentID: "intent-1", AttemptID: "attempt-1"}},
	}
	if err := aggregate.Validate(); err != nil {
		t.Fatal(err)
	}
	bad := aggregate
	bad.Legs = []CampaignLeg{{CampaignID: "other", Sequence: 2, PlanID: "plan-1", IntentID: "intent-1", AttemptID: "attempt-1"}}
	if err := bad.Validate(); !errors.Is(err, ErrInvalidIdentity) {
		t.Fatalf("bad aggregate err=%v, want invalid identity", err)
	}
	if !aggregate.Legs[0].SameIdentity(CampaignLeg{CampaignID: "campaign", Sequence: 1, PlanID: "plan-1", IntentID: "intent-1", AttemptID: "attempt-1"}) {
		t.Fatal("same immutable leg lineage was not equal")
	}
	if aggregate.Legs[0].SameIdentity(CampaignLeg{CampaignID: "campaign", Sequence: 1, PlanID: "plan-1", IntentID: "other", AttemptID: "attempt-1"}) {
		t.Fatal("changed intent lineage was accepted as the same leg")
	}
}

func TestReplayRejectsClosedReopenRebindAndIncompleteHistory(t *testing.T) {
	base := []Event{
		{Sequence: 1, CampaignVersion: 1, EventKind: "CREATED", CommandKind: "CREATE", CommandKey: "create", RequestDigest: "d1", CampaignState: CampaignPlanned, ProspectiveToken: "token", ExpectedPositionGeneration: 1},
		{Sequence: 2, CampaignVersion: 2, EventKind: "LEG_PLANNED", CommandKind: "PLAN_LEG", CommandKey: "plan", RequestDigest: "d2", CampaignState: CampaignPlanned, LegSequence: 1, PlanID: "plan-1", LegState: LegPlanned, LegRequestedQuantity: "3", LegFilledQuantity: "0", LegResidualQuantity: "3", ProspectiveToken: "token"},
		{Sequence: 3, CampaignVersion: 3, EventKind: "ORDER_LINKED", CommandKind: "LINK_ORDER", CommandKey: "link", RequestDigest: "d3", CampaignState: CampaignActive, LegSequence: 1, PlanID: "plan-1", LegState: LegSubmitted, LegRequestedQuantity: "3", LegFilledQuantity: "0", LegResidualQuantity: "3", OrderID: "order-1", RequestedCap: "3", CumulativeQuantity: "0", OrderRemainingQuantity: "3", ProspectiveToken: "token"},
		{Sequence: 4, CampaignVersion: 4, EventKind: "ORDER_WATERMARK_ADVANCED", CommandKind: "APPLY_FILL", CommandKey: "fill-1", RequestDigest: "d4", CampaignState: CampaignActive, PositionGeneration: 2, LegSequence: 1, PlanID: "plan-1", LegState: LegPartial, LegRequestedQuantity: "3", LegFilledQuantity: "1", LegResidualQuantity: "2", OrderID: "order-1", RequestedCap: "3", CumulativeQuantity: "1", DeltaQuantity: "1", OrderRemainingQuantity: "2", ProspectiveToken: "token"},
		{Sequence: 5, CampaignVersion: 5, EventKind: "ORDER_WATERMARK_ADVANCED", CommandKind: "APPLY_FILL", CommandKey: "close", RequestDigest: "d5", CampaignState: CampaignClosed, EntryBlocked: true, PositionGeneration: 2, LegSequence: 1, PlanID: "plan-1", LegState: LegFilled, LegRequestedQuantity: "3", LegFilledQuantity: "3", LegResidualQuantity: "0", OrderID: "order-1", RequestedCap: "3", CumulativeQuantity: "3", DeltaQuantity: "2", OrderRemainingQuantity: "0", OrderTerminal: true, ProspectiveToken: "token"},
	}
	reopened := appendCopy(base, Event{Sequence: 6, CampaignVersion: 6, EventKind: "AMBIGUOUS_ORDER_FILL", CommandKind: "RECORD_EVIDENCE", CommandKey: "recover", RequestDigest: "d6", CampaignState: CampaignActive, PositionGeneration: 2, ProspectiveToken: "token"})
	if got := Replay(reopened, Snapshot{}); got.Valid || got.Reason != ReplayClosedReopened {
		t.Fatalf("closed reopen replay=%+v", got)
	}
	reboundEvent := base[4]
	reboundEvent.CampaignState, reboundEvent.EntryBlocked, reboundEvent.PositionGeneration = CampaignActive, false, 3
	reboundEvent.LegState, reboundEvent.LegFilledQuantity, reboundEvent.LegResidualQuantity = LegPartial, "2", "1"
	reboundEvent.CumulativeQuantity, reboundEvent.DeltaQuantity, reboundEvent.OrderRemainingQuantity, reboundEvent.OrderTerminal = "2", "1", "1", false
	rebound := appendCopy(base[:4], reboundEvent)
	if got := Replay(rebound, Snapshot{}); got.Valid || got.Reason != ReplayGenerationRebound {
		t.Fatalf("generation rebound replay=%+v", got)
	}
	gap := []Event{
		base[0],
		{Sequence: 2, CampaignVersion: 2, EventKind: "LEG_PLANNED", CommandKind: "PLAN_LEG", CommandKey: "plan", RequestDigest: "d2", CampaignState: CampaignPlanned, LegSequence: 2, PlanID: "plan-2", LegState: LegPlanned, LegRequestedQuantity: "1", LegFilledQuantity: "0", LegResidualQuantity: "1", ProspectiveToken: "token"},
	}
	if got := Replay(gap, Snapshot{}); got.Valid || got.Reason != ReplayLegSequenceGap {
		t.Fatalf("leg gap replay=%+v", got)
	}
	orphan := appendCopy(base[:2], Event{Sequence: 3, CampaignVersion: 3, EventKind: "ORDER_LINKED", CommandKind: "LINK_ORDER", CommandKey: "link-orphan", RequestDigest: "d3x", CampaignState: CampaignActive, LegSequence: 1, PlanID: "plan-1", LegState: LegSubmitted, LegRequestedQuantity: "3", LegFilledQuantity: "0", LegResidualQuantity: "3", OrderID: "child", PredecessorOrderID: "missing", RequestedCap: "1", CumulativeQuantity: "0", OrderRemainingQuantity: "1", ProspectiveToken: "token"})
	if got := Replay(orphan, Snapshot{}); got.Valid || got.Reason != ReplayOrphanOrderLineage {
		t.Fatalf("orphan replay=%+v", got)
	}
	retreat := appendCopy(base[:4], Event{Sequence: 5, CampaignVersion: 5, EventKind: "ORDER_WATERMARK_ADVANCED", CommandKind: "APPLY_FILL", CommandKey: "fill-low", RequestDigest: "d5x", CampaignState: CampaignActive, PositionGeneration: 2, LegSequence: 1, PlanID: "plan-1", LegState: LegPartial, LegRequestedQuantity: "3", LegFilledQuantity: "0.5", LegResidualQuantity: "2.5", OrderID: "order-1", RequestedCap: "3", CumulativeQuantity: "0.5", DeltaQuantity: "0", OrderRemainingQuantity: "2.5", ProspectiveToken: "token"})
	if got := Replay(retreat, Snapshot{}); got.Valid || got.Reason != ReplayWatermarkRetreat {
		t.Fatalf("watermark retreat replay=%+v", got)
	}
}

func TestReplayRejectsStopRetreatAndProspectiveTokenRebind(t *testing.T) {
	events := []Event{
		{Sequence: 1, CampaignVersion: 1, EventKind: "CREATED", CommandKind: "CREATE", CommandKey: "create", RequestDigest: "d1", CampaignState: CampaignPlanned, ProspectiveToken: "token-a"},
		{Sequence: 2, CampaignVersion: 2, EventKind: "STOP_COMPOSED", CommandKind: "UPDATE_STOP", CommandKey: "stop-1", RequestDigest: "d2", CampaignState: CampaignPlanned, ProspectiveToken: "token-a", EffectiveStop: "100", StopSource: "risk", StopPolicy: "v1", StopObservedAt: "t1"},
		{Sequence: 3, CampaignVersion: 3, EventKind: "STOP_COMPOSED", CommandKind: "UPDATE_STOP", CommandKey: "stop-2", RequestDigest: "d3", CampaignState: CampaignPlanned, ProspectiveToken: "token-a", EffectiveStop: "90", StopSource: "risk", StopPolicy: "v2", StopObservedAt: "t2"},
	}
	if got := Replay(events, Snapshot{}); got.Valid || got.Reason != ReplayStopRetreat {
		t.Fatalf("stop retreat replay=%+v", got)
	}
	events[2] = Event{Sequence: 3, CampaignVersion: 3, EventKind: "STOP_COMPOSED", CommandKind: "UPDATE_STOP", CommandKey: "token-change", RequestDigest: "d3x", CampaignState: CampaignPlanned, ProspectiveToken: "token-b"}
	if got := Replay(events, Snapshot{}); got.Valid || got.Reason != ReplayProspectiveRebound {
		t.Fatalf("token rebound replay=%+v", got)
	}
}

func TestLegLedgerAggregateCapAndCalculationFailureAreFailClosedAtomic(t *testing.T) {
	ledger, _ := NewLegLedger("7")
	_ = ledger.LinkOrder("a", "", "4")
	_ = ledger.LinkOrder("b", "", "4")
	_, _ = ledger.Observe(OrderObservation{OrderID: "a", Cumulative: "4"})
	out, err := ledger.Observe(OrderObservation{OrderID: "b", Cumulative: "4"})
	if err != nil || out.Delta != "4" || !out.Reconcile || ledger.Filled != "8" {
		t.Fatalf("aggregate cap out=%+v ledger=%+v err=%v", out, ledger, err)
	}

	broken, _ := NewLegLedger("5")
	_ = broken.LinkOrder("order", "", "5")
	broken.Requested = "not-a-decimal"
	before := *broken.Orders["order"]
	if _, err := broken.Observe(OrderObservation{OrderID: "order", Cumulative: "2"}); err == nil {
		t.Fatal("corrupt aggregate must fail")
	}
	if after := *broken.Orders["order"]; after != before || broken.Filled != "0" || broken.Residual != "5" {
		t.Fatalf("failed observation mutated state: before=%+v after=%+v ledger=%+v", before, after, broken)
	}
}

func TestStopRequiresPositivePriceAndCompleteProvenance(t *testing.T) {
	saved := &EffectiveStop{Price: "100", Source: "risk", Policy: "v1", ObservedAt: "t0"}
	for _, candidate := range []StopCandidate{
		{Price: "0", Valid: true, Source: "risk", Policy: "v2", ObservedAt: "t1"},
		{Price: "110", Valid: true, Source: "", Policy: "v2", ObservedAt: "t1"},
		{Price: "110", Valid: true, Source: "risk", Policy: "", ObservedAt: "t1"},
		{Price: "110", Valid: true, Source: "risk", Policy: "v2", ObservedAt: ""},
	} {
		got, blocked, err := ComposeLongStop(saved, candidate)
		if err != nil || !blocked || got.Price != "100" || got.Candidate != candidate {
			t.Fatalf("candidate=%+v got=%+v blocked=%v err=%v", candidate, got, blocked, err)
		}
	}
}

func TestCommandIdentityRejectsUnknownWhitespaceAndNULWithoutConcatenationCollision(t *testing.T) {
	for _, tc := range []struct{ kind, key string }{
		{"ARBITRARY", "key"},
		{"CREATE", "   "},
		{"CREATE\x00PLAN_LEG", "key"},
		{"CREATE", "a\x00b"},
	} {
		if err := ValidateCommand(tc.kind, tc.key); !errors.Is(err, ErrInvalidCommand) {
			t.Fatalf("ValidateCommand(%q,%q) err=%v", tc.kind, tc.key, err)
		}
	}
	a, err := TypedCommandIdentity("CREATE", "a:b")
	if err != nil {
		t.Fatal(err)
	}
	b, err := TypedCommandIdentity("CREATE", "a_b")
	if err != nil {
		t.Fatal(err)
	}
	if a == b {
		t.Fatalf("typed command identities collided: %q", a)
	}
}

func TestProductionCoreContainsNoLaneRatiosOrBrokerDependency(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".go" || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		body, err := os.ReadFile(entry.Name())
		if err != nil {
			t.Fatal(err)
		}
		text := string(body)
		for _, forbidden := range []string{"8:4:2", "2:4:8", "net/http", "internal/broker"} {
			if strings.Contains(text, forbidden) {
				t.Errorf("%s contains forbidden lane/broker token %q", entry.Name(), forbidden)
			}
		}
	}
}

func appendCopy(events []Event, event Event) []Event {
	out := append([]Event(nil), events...)
	return append(out, event)
}

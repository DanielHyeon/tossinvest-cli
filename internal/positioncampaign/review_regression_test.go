package positioncampaign

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
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

// 이 두 시험은 "금지 토큰이 **있는지**" 를 보지 않는다.
//
// 예전 판본은 네 부분문자열("8:4:2","2:4:8","net/http","internal/broker")을 찾았다.
// 그중 하나는 존재하지 않는 디렉터리를 이름했고, 무엇보다 **철자를 고르면 통과한다** —
// 적대 리뷰가 그 리터럴을 하나도 쓰지 않은 진짜 7-leg 비율표와 시장별 cap 을 이
// 패키지에 붙였는데 스위트가 초록이었다. 이름을 보는 가드는 반드시 뚫린다.
//
// 그래서 판정을 **역할** 로 바꾼다: 무엇을 의존하는가(import 폐포)와 어떤 상수를
// 담는가(숫자 열거표). 새 비율·cadence·cap 상수나 새 의존은 철자와 무관하게 걸린다.

const positionCampaignModule = "github.com/JungHoonGhae/tossinvest-cli"

// productionGoFiles 는 한 패키지 디렉터리의 비테스트 Go 파일을 파싱한다.
func productionGoFiles(t *testing.T, dir string) map[string]*ast.File {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]*ast.File{}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".go" || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		out[path] = file
	}
	return out
}

// TestProductionCoreContainsNoLaneRatiosOrBrokerDependency 는 import **폐포** 전체를
// 본다. 직접 import 만 보면 riskcalc 가 broker 를 끌어오는 순간 눈이 먼다.
func TestProductionCoreContainsNoLaneRatiosOrBrokerDependency(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	allowed := map[string]bool{
		"internal/positioncampaign": true,
		"internal/riskcalc":         true,
	}
	visited := map[string]bool{}
	queue := []string{"internal/positioncampaign"}
	for len(queue) > 0 {
		pkg := queue[0]
		queue = queue[1:]
		if visited[pkg] {
			continue
		}
		visited[pkg] = true
		if !allowed[pkg] {
			t.Errorf("전략 중립 코어의 import 폐포에 %s 가 들어왔다", pkg)
			continue
		}
		for path, file := range productionGoFiles(t, filepath.Join(root, pkg)) {
			for _, spec := range file.Imports {
				imported := strings.Trim(spec.Path.Value, `"`)
				if strings.HasPrefix(imported, positionCampaignModule+"/") {
					queue = append(queue, strings.TrimPrefix(imported, positionCampaignModule+"/"))
					continue
				}
				// 표준 라이브러리 경로에는 점이 있는 첫 구간이 없다. 점이 있으면
				// 제3자 모듈이다 — broker SDK 든 HTTP 클라이언트든 여기서 걸린다.
				if strings.Contains(strings.SplitN(imported, "/", 2)[0], ".") {
					t.Errorf("%s: 전략 중립 코어가 제3자 모듈 %s 를 의존한다", path, imported)
				}
			}
		}
	}
	for pkg := range allowed {
		if !visited[pkg] {
			t.Errorf("허용 목록의 %s 가 실제 폐포에 없다 — 목록이 낡았다", pkg)
		}
	}
}

// TestProductionCoreNumericLiteralsAreFrozen 은 이 패키지가 담은 **모든** 숫자
// 리터럴(10진 문자열 형태 포함)을 센다. 열거표는 측정으로 만들었다.
//
// spec: "코어는 특정 lane 의 비율, 최대 leg 수 또는 cadence 상수를 포함해서는 안 된다".
// 비율표는 숫자 없이 쓸 수 없다. 그래서 값을 얼린다 — 새 숫자가 들어오면 그것이
// 무엇이든 이 시험이 먼저 실패하고, 저자는 그 숫자가 왜 중립인지 적어야 한다.
func TestProductionCoreNumericLiteralsAreFrozen(t *testing.T) {
	frozen := map[string]string{
		"0":     "빈 값·경계 비교",
		"1":     "길이·인덱스 경계",
		"128":   "command key 최대 길이 (identity 계약)",
		`str:0`: "10진 수량 0",
	}
	numeric := regexp.MustCompile(`^[+-]?[0-9]+(\.[0-9]+)?$`)
	found := map[string][]string{}
	for path, file := range productionGoFiles(t, ".") {
		ast.Inspect(file, func(node ast.Node) bool {
			lit, ok := node.(*ast.BasicLit)
			if !ok {
				return true
			}
			value := lit.Value
			switch lit.Kind {
			case token.INT, token.FLOAT:
			case token.STRING:
				unquoted := strings.Trim(value, "`\"")
				if !numeric.MatchString(unquoted) {
					return true
				}
				value = "str:" + unquoted
			default:
				return true
			}
			found[value] = append(found[value], filepath.Base(path))
			return true
		})
	}
	for value, files := range found {
		if _, ok := frozen[value]; !ok {
			sort.Strings(files)
			t.Errorf("전략 중립 코어에 새 숫자 상수 %s 가 들어왔다 (%v) — 비율/cadence/cap 인지 밝힐 것", value, files)
		}
	}
	for value, why := range frozen {
		if len(found[value]) == 0 {
			t.Errorf("얼린 숫자 %s (%s) 가 사라졌다 — 열거표를 다시 재고 근거를 적을 것", value, why)
		}
	}
}

func appendCopy(events []Event, event Event) []Event {
	out := append([]Event(nil), events...)
	return append(out, event)
}

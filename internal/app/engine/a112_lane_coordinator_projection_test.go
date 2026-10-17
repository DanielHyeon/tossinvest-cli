//go:build tossos_testseams

package engine

// a112 태스크 7.3 — 읽기 전용 투영의 additive 자식 `lanes[8]` · `coordinators[2]`(Manager 판정 Q1~Q3, 2026-10-01).
//
//   - lanes: Context.Read 가 프로세스의 여덟 레인 런타임을 **읽기만** 해서 덧씌운다(감독자 잠금 overlay 와 같은 자리). cycleGeneration 은
//     시장별 evaluate 물결 번호 중 그 레인이 마지막으로 관측된 번호(0 = 이 프로세스에서 미관측) — Q1=(B).
//   - coordinators: 조립 발행 때 시장별 제안 조정 스냅숏과, 주문 경로와 **같은** handoff 목록에서 승인된 소유자 범위 전부(selected[] — R4).
//
// 읽기 전용 불변(조건 ②): 투영을 몇 번 읽어도 레인 상태(투입 · 버림 · 건강 · 잠금 · 기한)와 관측 기록이 그대로여야 한다.

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyarbiter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojection"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyworker"
)

// a112LaneProjectionContext 는 투영 저장소(dormant)와 생산 레인 런타임(잠금 없이 태어난 여덟)을 가진 Context 다. 감독자는 돌지 않는다 —
// 레인 상태를 바꾸는 것은 시험이 부르는 evaluate 뿐이다.
func a112LaneProjectionContext(t *testing.T) (*Context, *strategyLaneRuntime, *clock.Fake) {
	t.Helper()
	cycle, _, j, _ := pairedStrategyDispatchCycleFixture(t)
	loader := cycle.firstLeg.loader.(*productionStrategyFirstLegAuthorityLoader)
	fake := clock.NewFake(loader.clk.Now())
	store, err := strategyprojection.NewStore(strategyprojection.DormantSnapshot(fake.Now()))
	if err != nil {
		t.Fatal(err)
	}
	c := &Context{Journal: j, AccountRef: "acct-risk-loader", strategyProjection: store}
	lanes, err := c.productionStrategyLanes(context.Background(), fake)
	if err != nil || lanes == nil {
		t.Fatalf("arrangement: production lane runtime: %v", err)
	}
	return c, lanes, fake
}

func a112Read(t *testing.T, c *Context) strategyprojection.Snapshot {
	t.Helper()
	snapshot, err := c.Read(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := strategyprojection.Validate(snapshot); err != nil {
		t.Fatalf("the projected snapshot breaks its own contract: %v", err)
	}
	return snapshot
}

// 레인 런타임이 없으면(첫 생산 주기 전) 투영은 골든 기본값 그대로다 — 미관측(건강 null).
func TestAProcessWithoutLanesProjectsTheEightUnobservedDefaults(t *testing.T) {
	store, err := strategyprojection.NewStore(strategyprojection.DormantSnapshot(time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)))
	if err != nil {
		t.Fatal(err)
	}
	snapshot := a112Read(t, &Context{strategyProjection: store})
	if len(snapshot.Lanes) != 8 {
		t.Fatalf("lanes=%d, want 8", len(snapshot.Lanes))
	}
	for _, lane := range snapshot.Lanes {
		if lane.Health != nil || lane.CycleGeneration != 0 {
			t.Fatalf("lane %s observed without a lane runtime: %+v", lane.LaneID, lane)
		}
	}
}

// 시장별 물결: KR 을 두 번 · US 를 한 번 돌리면 KR 레인 넷은 2, US 넷은 1 이다. 한 번도 안 돈 시장은 0(묵은 레인 식별).
func TestTheCycleGenerationIsTheMarketWaveInWhichTheLaneWasLastObserved(t *testing.T) {
	c, lanes, fake := a112LaneProjectionContext(t)
	policy := strategyworker.ProductionRuntimePolicy()
	ctx := context.Background()
	if err := lanes.evaluate(ctx, StrategyMarketKR, 0, strategyrouter.FamilyActivation{}, nil, strategyShadowBatch{}); err != nil {
		t.Fatal(err)
	}
	before := a112Read(t, c)
	for _, lane := range before.Lanes {
		want := uint64(0)
		if lane.Market == strategyprojection.MarketKR {
			want = 1
		}
		if lane.CycleGeneration != want {
			t.Fatalf("after one KR wave: lane %s generation=%d, want %d", lane.LaneID, lane.CycleGeneration, want)
		}
	}
	// 카덴스가 지나야 둘째 물결이 사이클을 연다(안 지나면 TOO_SOON — 그것도 관측이지만 이 시험은 열린 사이클의 모양을 잰다).
	fake.Advance(policy.Cadence())
	if err := lanes.evaluate(ctx, StrategyMarketKR, 0, strategyrouter.FamilyActivation{}, nil, strategyShadowBatch{}); err != nil {
		t.Fatal(err)
	}
	if err := lanes.evaluate(ctx, StrategyMarketUS, 0, strategyrouter.FamilyActivation{}, nil, strategyShadowBatch{}); err != nil {
		t.Fatal(err)
	}
	for _, lane := range a112Read(t, c).Lanes {
		want := uint64(1)
		if lane.Market == strategyprojection.MarketKR {
			want = 2
		}
		if lane.CycleGeneration != want || lane.Health == nil || *lane.Health != strategyprojection.LaneHealthy ||
			lane.Trigger == nil || *lane.Trigger != strategyprojection.LaneTriggerEnqueued ||
			lane.Start == nil || *lane.Start != strategyprojection.LaneStartAdmitted ||
			lane.Outcome == nil || *lane.Outcome != strategyprojection.LaneOutcomeDormant ||
			lane.Desired != strategyprojection.StateOff || lane.Effective != strategyprojection.StateOff ||
			lane.PolicyVersion == nil || *lane.PolicyVersion != policy.Version() ||
			lane.CycleDeadlineMS == nil || *lane.CycleDeadlineMS != policy.CycleDeadline().Milliseconds() ||
			lane.NextDueAt == nil || lane.SnapshotDigest != nil || lane.EvidenceDigest != nil {
			t.Fatalf("lane %s projection=%+v, want wave %d · HEALTHY · ENQUEUED/ADMITTED/DORMANT · OFF/OFF · server policy", lane.LaneID, lane, want)
		}
	}
}

// 서명 활성화가 KR 네 가족을 켠 물결은 KR 레인 넷을 ON/ON 으로 보인다(관측 시점의 활성화 — 레인 자기 열쇠로 물음). US 는 OFF.
func TestTheLaneDesiredAndEffectiveAreTheActivationTheWaveRanWith(t *testing.T) {
	c, lanes, _ := a112LaneProjectionContext(t)
	activation := strategyrouter.FamilyActivationForTest(strategyrouter.MarketKR, 1, strategyrouter.AllFourFamiliesForTest(strategyrouter.MarketKR))
	if err := lanes.evaluate(context.Background(), StrategyMarketKR, 1, activation, nil, strategyShadowBatch{}); err != nil {
		t.Fatal(err)
	}
	for _, lane := range a112Read(t, c).Lanes {
		want := strategyprojection.StateOff
		if lane.Market == strategyprojection.MarketKR {
			want = strategyprojection.StateOn
		}
		if lane.Desired != want || lane.Effective != want {
			t.Fatalf("lane %s desired=%s effective=%s, want %s", lane.LaneID, lane.Desired, lane.Effective, want)
		}
		// 판정 (A) — first refusal: 켜진 KR 레인은 자기 제안 없이 돌아 REFUSED 와 골든 코드(봉인 불일치 — 「이 레인 제안 아님」)를 싣고,
		// 꺼진 US 레인은 DORMANT 라 거절 코드가 null 이다.
		if lane.Market == strategyprojection.MarketKR {
			if lane.Outcome == nil || *lane.Outcome != strategyprojection.LaneOutcomeRefused || lane.Refusal == nil ||
				*lane.Refusal != string(strategyarbiter.RefusalSealMismatch) {
				t.Fatalf("activated KR lane %s outcome=%v refusal=%v, want REFUSED with %s", lane.LaneID, lane.Outcome, lane.Refusal,
					strategyarbiter.RefusalSealMismatch)
			}
		} else if lane.Refusal != nil {
			t.Fatalf("dormant US lane %s carries refusal %q", lane.LaneID, *lane.Refusal)
		}
	}
}

// 5.6.2.2 하네스 재사용: 여덟 레인이 동시에 잠기고 두 시장 주기가 돈 뒤, 투영은 여덟을 생산 순서로 LATCHED 로 보이고 첫 원인을 싣는다.
func TestEightLatchedLanesAreProjectedInProductionOrderWithTheirFirstFailure(t *testing.T) {
	run := newA112EightLaneRun(t, nil)
	run.waitBothMarketsObserved(t)
	store, err := strategyprojection.NewStore(strategyprojection.DormantSnapshot(time.Now().UTC()))
	if err != nil {
		t.Fatal(err)
	}
	run.context.strategyProjectionMu.Lock()
	run.context.strategyProjection = store
	run.context.strategyProjectionMu.Unlock()
	snapshot := a112Read(t, run.context)
	for index, lane := range run.lanes.lanes {
		got := snapshot.Lanes[index]
		key := lane.Key()
		if string(got.Market) != string(key.Market) || got.Family != string(key.Family) || got.LaneID != key.LaneID ||
			got.Health == nil || *got.Health != strategyprojection.LaneLatched || got.LatchRevision != 1 ||
			got.FirstFailure == nil || *got.FirstFailure != "injected lane fault (5.6.2.2)" ||
			got.Trigger == nil || *got.Trigger != strategyprojection.LaneTriggerDisabled || got.Start != nil || got.CycleGeneration == 0 {
			t.Fatalf("lane %d (%v) projection=%+v, want LATCHED revision 1 with its first failure, DISABLED, observed", index, key, got)
		}
	}
}

// 레인 잠금 사유는 자유 문장이다(패닉 값 · 원장에서 되살린 사유). 줄바꿈 · NUL 이 섞여도 투영은 정규화해 싣고 스냅숏 전체가
// 거절되지 않는다 — 한 레인의 사유 문장이 legacy 시장 레코드까지 끌고 내려가면 안 된다.
func TestALatchReasonWithControlCharactersIsNormalizedNotRefused(t *testing.T) {
	c, lanes, _ := a112LaneProjectionContext(t)
	lanes.lanes[3].Fail("panic: bad\nreason\x00 tail", true)
	got := a112Read(t, c).Lanes[3]
	if got.Health == nil || *got.Health != strategyprojection.LaneLatched || got.FirstFailure == nil || *got.FirstFailure != "panic: bad reason  tail" {
		t.Fatalf("lane 3 projection=%+v first=%v, want LATCHED with the normalized reason", got, got.FirstFailure)
	}
}

type a112LaneState struct {
	pending                      int
	dropped, abandoned, failures uint64
	revision                     uint64
	health                       strategyworker.LaneHealth
	first                        string
	nextDue, restartNotBefore    time.Time
}

func a112LaneStates(runtime *strategyLaneRuntime) []a112LaneState {
	runtime.mu.RLock()
	defer runtime.mu.RUnlock()
	out := make([]a112LaneState, 0, len(runtime.lanes))
	for _, lane := range runtime.lanes {
		out = append(out, a112LaneState{pending: lane.Pending(), dropped: lane.Dropped(), abandoned: lane.Abandoned(),
			failures: lane.ConsecutiveFailures(), revision: lane.LatchRevision(), health: lane.Health(), first: lane.FirstFailure(),
			nextDue: lane.NextDue(), restartNotBefore: lane.RestartNotBefore()})
	}
	return out
}

// 조건 ②: 투영은 읽기만 한다. 투입(Offer) · 실패(Fail) · 관측 기록 · 물결 번호 · 원장 잠금 어느 것도 Read 로 움직이지 않는다.
func TestReadingTheLaneProjectionNeverChangesALane(t *testing.T) {
	c, lanes, fake := a112LaneProjectionContext(t)
	if err := lanes.evaluate(context.Background(), StrategyMarketKR, 0, strategyrouter.FamilyActivation{}, nil, strategyShadowBatch{}); err != nil {
		t.Fatal(err)
	}
	fake.Advance(time.Hour) // 카덴스가 지나도 Read 는 사이클을 열지 않는다.
	states, observations := a112LaneStates(lanes), lanes.observations()
	open, err := c.Journal.OpenStrategyLaneLatches(context.Background(), "acct-risk-loader")
	if err != nil {
		t.Fatal(err)
	}
	first := a112Read(t, c)
	for range 16 {
		a112Read(t, c)
	}
	last := a112Read(t, c)
	if got := a112LaneStates(lanes); !reflect.DeepEqual(got, states) {
		t.Fatalf("reading the projection changed lane state:\nbefore %+v\nafter  %+v", states, got)
	}
	if got := lanes.observations(); !reflect.DeepEqual(got, observations) {
		t.Fatalf("reading the projection changed the recorded observations")
	}
	if after, err := c.Journal.OpenStrategyLaneLatches(context.Background(), "acct-risk-loader"); err != nil || len(after) != len(open) {
		t.Fatalf("reading the projection changed durable lane latches: %d → %d (%v)", len(open), len(after), err)
	}
	if !reflect.DeepEqual(first.Lanes, last.Lanes) {
		t.Fatal("eighteen reads of an unchanged runtime projected different lanes")
	}
}

// R4: 서명 활성화된 두 범위 시장은 조정자 자식에 **두 범위 다** 보인다(조정자 순서). 시장 레코드는 오늘처럼 첫 범위 그대로다.
// 활성화 없는 두 범위 시장은 시장 단위 handoff 가 상한으로 거절되므로 selected 는 비어 있다 — 주문 경로와 같은 목록을 읽는다.
func TestTheCoordinatorChildShowsEveryAdmittedOwnerScope(t *testing.T) {
	_, proposals, _, _ := pairedStrategyDispatchCycleFixture(t)
	two := a112TwoScopeKR(t, proposals.kr, proposals.observedAt)
	unactivated := proposals
	unactivated.kr = two
	activated := unactivated
	activated.kr.activation = strategyrouter.FamilyActivationForTest(strategyrouter.MarketKR, 1, strategyrouter.AllFourFamiliesForTest(strategyrouter.MarketKR))
	// 범위마다 **다른** 중재 보정을 경로 권한에 붙인다 — 투영이 범위의 자기 항목에서 읽는지(첫 항목 · 남의 항목이 아니라) 가를 수 있게.
	calibrations := []strategyrouter.ProductionRouteCalibration{
		{ScoreVersion: "arbitration-score:v1", CalibrationDigest: "sha256:calibration-scope-1"},
		{ScoreVersion: "arbitration-score:v2", CalibrationDigest: "sha256:calibration-scope-2"},
	}
	entries := append([]strategyProposalEntryAuthority(nil), activated.kr.entries...)
	for index := range entries {
		entries[index].route.route = strategyrouter.WithArbitrationScoresForTest(entries[index].route.route, calibrations[index], nil)
	}
	activated.kr.entries = entries
	supervisor, err := NewStrategyEntrySupervisor(StrategyEntrySupervisorOptions{
		Workers: []StrategyMarketWorker{{Market: StrategyMarketKR}, {Market: StrategyMarketUS}}, Clock: clock.NewFake(proposals.observedAt)})
	if err != nil {
		t.Fatal(err)
	}

	schedule := PairedStrategyScheduleSnapshot{ObservedAt: proposals.observedAt}
	projected := strategyProjectionFromAssembly(StrategyEntryProductionAssembly{Supervisor: supervisor, Schedule: schedule, proposals: activated})
	if err := strategyprojection.Validate(projected); err != nil {
		t.Fatal(err)
	}
	kr, us := projected.Coordinators[0], projected.Coordinators[1]
	if kr.Market != strategyprojection.MarketKR || us.Market != strategyprojection.MarketUS {
		t.Fatalf("coordinator order=%s,%s, want KR,US", kr.Market, us.Market)
	}
	if len(kr.Selected) != 2 {
		t.Fatalf("activated two-scope KR selected=%+v, want both owner scopes", kr.Selected)
	}
	for index, entry := range activated.kr.entries {
		lineage := entry.authority.Proposal().Lineage
		got := kr.Selected[index]
		if got.Symbol != lineage.Symbol || got.LaneID != lineage.LaneID || got.CampaignID != lineage.CampaignID || got.LineageIdentity != lineage.Identity {
			t.Fatalf("selected[%d]=%+v, want the coordinator-order scope %s/%s", index, got, lineage.Symbol, lineage.Identity)
		}
		// 판정 (A) — config · calibration 계보: 설정 digest 는 계보에서, 보정은 그 범위 **자기** 항목의 경로 권한에서.
		if got.ConfigDigest == nil || *got.ConfigDigest != lineage.ConfigDigest || got.ScoreVersion == nil ||
			*got.ScoreVersion != calibrations[index].ScoreVersion || got.CalibrationDigest == nil ||
			*got.CalibrationDigest != calibrations[index].CalibrationDigest {
			t.Fatalf("selected[%d] lineage config=%v score=%v calibration=%v, want %q · %+v", index, got.ConfigDigest, got.ScoreVersion,
				got.CalibrationDigest, lineage.ConfigDigest, calibrations[index])
		}
	}
	snapshot := activated.kr.snapshot
	if kr.Reason == nil || *kr.Reason != string(snapshot.Reason) || kr.Ready != snapshot.Ready || kr.ProposedCount != snapshot.ProposedCount ||
		kr.ProposalSetDigest == nil || *kr.ProposalSetDigest != projectionDigest(snapshot.ProposalSetDigest) {
		t.Fatalf("KR coordinator=%+v, want the market's proposal coordination snapshot %+v", kr, snapshot)
	}
	if len(us.Selected) != 1 || us.Selected[0].Symbol != activated.us.entries[0].authority.Proposal().Lineage.Symbol {
		t.Fatalf("US selected=%+v, want its one market handoff", us.Selected)
	}
	// US 항목의 경로 권한에는 보정이 없다 — 지어내지 않고 null.
	if us.Selected[0].ScoreVersion != nil || us.Selected[0].CalibrationDigest != nil {
		t.Fatalf("US selected calibration=%v/%v, want null when the route carries none", us.Selected[0].ScoreVersion, us.Selected[0].CalibrationDigest)
	}

	closed := strategyProjectionFromAssembly(StrategyEntryProductionAssembly{Supervisor: supervisor, Schedule: schedule, proposals: unactivated})
	if got := closed.Coordinators[0].Selected; got == nil || len(got) != 0 {
		t.Fatalf("unactivated two-scope KR selected=%+v, want empty — its one market handoff is refused by the capacity", got)
	}
}

// 조정자 사유 어휘는 엔진의 StrategyProposalReason 상수 전부다(AST census) — 엔진이 사유를 하나 더하면 투영이 그것을 알아야
// Validate 가 스냅숏 전체를 거절하지 않는다. 지어낸 사유도, 빠진 사유도 실패다.
func TestTheCoordinatorReasonVocabularyIsEveryEngineProposalReason(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "strategy_proposal_authority.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var engine []string
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		for _, spec := range gen.Specs {
			value := spec.(*ast.ValueSpec)
			if ident, ok := value.Type.(*ast.Ident); !ok || ident.Name != "StrategyProposalReason" {
				continue
			}
			for _, literal := range value.Values {
				text, err := strconv.Unquote(literal.(*ast.BasicLit).Value)
				if err != nil {
					t.Fatal(err)
				}
				engine = append(engine, text)
			}
		}
	}
	sort.Strings(engine)
	projection := append([]string(nil), strategyprojection.CoordinatorReasons()...)
	sort.Strings(projection)
	if len(engine) == 0 || !reflect.DeepEqual(engine, projection) {
		t.Fatalf("engine proposal reasons=%v, projection vocabulary=%v", engine, projection)
	}
}

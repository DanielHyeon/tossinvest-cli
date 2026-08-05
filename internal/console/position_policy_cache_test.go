package console

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
)

// The intervals a081 fixes, written here as literals so the assertions below say
// what they mean rather than restating whatever the implementation happens to
// hold. They are deliberately different from each other: only List contends for
// the engine's write connection, and only Runtime carries live reconcile state.
const (
	wantLifecycleInterval = 30 * time.Second
	wantRuntimeInterval   = 5 * time.Second
)

// What a row says about a running engine's adoption settings, and what it says
// instead when that engine could not be read (operatorview.BuildExitLineReference).
const (
	engineStopPercentMark = "기준선 미생성 · 엔진 보호 미적용"
	engineUnknownMark     = "기준선·정책 폭 알 수 없음"
)

// What a row says about its own lifecycle (positionpolicy.ProjectManagement).
const (
	managedByEngineMark   = "엔진 관리"
	managementUnknownMark = "관리 여부 불명"
	reconcileBlockedMark  = "관리 편입 · 대사 차단으로 대기"
	adoptionPendingMark   = "관리 편입 · 편입 예약됨"
)

// countingPolicyCommander counts what reaches the engine.
//
// a080's budget test could not see the engine at all: it used newDashboardHarness,
// which leaves Options.PositionPolicies nil, so it never executed the branch that
// reads the engine (a080 review.md F1). Several other tests in this package do
// wire the seam — a052, a053, a077, a079 — but none of them counts anything, and
// none renders a screen twice. Nothing in the package could observe how often the
// console reaches the engine, which is what this file adds.
//
// It embeds the ordinary fake so Preview and Apply behave exactly as the rest of
// the package expects: the invalidation tests below drive the real capability
// flow rather than a shortcut past it.
type countingPolicyCommander struct {
	fakePositionPolicyCommander

	mu           sync.Mutex
	runtimeValue positionpolicy.ManagementRuntime
	runtimeErr   error
	states       []positionpolicy.State
	statesErr    error
	runtimeCalls int
	listCalls    int
	// block, when set, is served by the next Runtime read. It stands for the
	// reconcile tracker moving under a screen that is already open.
	block *positionpolicy.ReconcileBlock
	// duringRead runs inside a read, outside this fake's own mutex. It widens the
	// window a real engine round-trip occupies, which is the only way a test can
	// tell "one reading is taken and shared" from "the reads happened to be too
	// fast to overlap".
	duringRead func()
}

// Both reads honour the context the way the shipped clients do — positionpolicyrpc
// returns ctx.Err() verbatim — because whether a cancelled request can reach the
// cache is one of the things this file is here to decide.
func (c *countingPolicyCommander) Runtime(ctx context.Context) (positionpolicy.ManagementRuntime, error) {
	c.mu.Lock()
	c.runtimeCalls++
	runtime, failure, during := c.runtimeValue, c.runtimeErr, c.duringRead
	if c.block != nil {
		runtime.Blocks = []positionpolicy.ReconcileBlock{*c.block}
	}
	c.mu.Unlock()

	if during != nil {
		during()
	}
	if err := ctx.Err(); err != nil {
		return positionpolicy.ManagementRuntime{}, err
	}
	if failure != nil {
		return positionpolicy.ManagementRuntime{}, failure
	}
	return runtime, nil
}

func (c *countingPolicyCommander) List(ctx context.Context) ([]positionpolicy.State, error) {
	c.mu.Lock()
	c.listCalls++
	states, failure, during := append([]positionpolicy.State(nil), c.states...), c.statesErr, c.duringRead
	c.mu.Unlock()

	if during != nil {
		during()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if failure != nil {
		return nil, failure
	}
	return states, nil
}

func (c *countingPolicyCommander) reads() (runtime, list int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.runtimeCalls, c.listCalls
}

func (c *countingPolicyCommander) failRuntime(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.runtimeErr = err
}

func (c *countingPolicyCommander) failList(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.statesErr = err
}

func (c *countingPolicyCommander) forget() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.states = nil
}

func (c *countingPolicyCommander) blockSymbol(market, symbol string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	block := positionpolicy.NewReconcileBlock(positionpolicy.ScopeSymbol, market, symbol,
		"QUANTITY_MISMATCH", "브로커 수량과 원장 수량이 다르다",
		time.Date(2026, 7, 27, 1, 0, 0, 0, time.UTC), false)
	c.block = &block
}

// newCountingPolicyCommander seeds a lifecycle row for the journal's managed
// position.
//
// The PositionID has to be the journal fixture's `pos-managed`: decoratePositionRows
// joins policyByID on row.PositionID, so a state keyed on anything else is never
// looked up and the whole lifecycle half of the reading goes untested — which is
// exactly how the first version of this file passed while a mutation that revived
// a stale lifecycle listing broke nothing.
func newCountingPolicyCommander() *countingPolicyCommander {
	state := managedPolicyState()
	state.PositionID = "pos-managed"
	state.Symbol = "005930"
	commander := &countingPolicyCommander{
		runtimeValue: positionpolicy.ManagementRuntime{
			Effective:      positionpolicy.NewAdoptionSettings(true, .05, nil, nil, ""),
			EffectiveKnown: true,
		},
		states: []positionpolicy.State{state},
	}
	commander.fakePositionPolicyCommander.state = state
	return commander
}

// countingPolicyHarness wires the commander and gets the session out of the way.
//
// authenticate renders the overview, which is itself one of the two screens that
// read the engine, so every test measures from a baseline taken after it. The
// clock is then advanced past the longer interval so the first render each test
// makes is one that reads.
func countingPolicyHarness(t *testing.T, commander *countingPolicyCommander) (*dashboardHarness, func() (int, int)) {
	t.Helper()
	h := newDashboardHarness(t, func(o *Options) { o.PositionPolicies = commander })
	// A readable journal is what lets a row reach a management verdict at all;
	// without one every row is UNKNOWN for a reason that has nothing to do with
	// the engine reading this change is about.
	seedJournal(t, h.journal)
	h.authenticate(t)
	h.clock.advance(wantLifecycleInterval)
	baseRuntime, baseList := commander.reads()
	return h, func() (int, int) {
		runtime, list := commander.reads()
		return runtime - baseRuntime, list - baseList
	}
}

// managedRow is the journal's managed position as the positions screen renders it.
func managedRow(t *testing.T, page string) string {
	t.Helper()
	return positionHTMLRow(t, page, "005930")
}

// candidateRow is the holding the journal has never heard of — a candidate for
// adoption, and therefore the row a reconcile block covers.
func candidateRow(t *testing.T, page string) string {
	t.Helper()
	return positionHTMLRow(t, page, "000660")
}

// TestRedrawingTheLineDoesNotAskTheEngineAgainWithinTheInterval is task 2.2 and
// the whole of a081's safety claim.
//
// The engine's journal handle has one connection and the exit loop's judgement
// transaction writes on it, so a display path that reaches the engine once per
// render puts the number of open screens and the reload period directly into the
// interval between stop-loss judgements.
func TestRedrawingTheLineDoesNotAskTheEngineAgainWithinTheInterval(t *testing.T) {
	commander := newCountingPolicyCommander()
	h, since := countingPolicyHarness(t, commander)

	// Task 2.3: the redraws have to outnumber the reads the interval allows, or
	// the assertion below is satisfiable without any caching at all. Derived from
	// the intervals rather than written as a literal, so changing an interval
	// cannot quietly turn this test into a tautology.
	redraws := int(wantLifecycleInterval/wantRuntimeInterval) + 2
	step := wantLifecycleInterval / time.Duration(redraws+1)
	if redraws < 3 || step*time.Duration(redraws) >= wantLifecycleInterval {
		t.Fatalf("this test only means something when %d redraws fit inside one lifecycle interval "+
			"with room to spare; step=%s", redraws, step)
	}
	for i := 0; i < redraws; i++ {
		h.page(t, "/positions")
		h.clock.advance(step)
	}

	_, list := since()
	if list != 1 {
		t.Errorf("%d redraws inside one lifecycle interval reached the engine's lifecycle listing %d "+
			"time(s); the bound is one. Every read past the first is time taken from the connection "+
			"the exit loop judges on", redraws, list)
	}
}

// TestBothLineScreensShareOneEngineReading is task 2.4.
//
// The two screens share decoratePositionRows precisely so one holding cannot get
// two different protection answers. They must share the reading behind it too, or
// having both tabs open doubles what the console costs the engine.
func TestBothLineScreensShareOneEngineReading(t *testing.T) {
	commander := newCountingPolicyCommander()
	h, since := countingPolicyHarness(t, commander)

	positions := h.page(t, "/positions")
	h.clock.advance(wantRuntimeInterval / 2)
	overview := h.page(t, "/dashboard")

	// Both screens must actually be rendering the engine's verdict; a /dashboard
	// that stopped calling decoratePositionRows would otherwise satisfy the count
	// below by doing nothing.
	for name, page := range map[string]string{"/positions": positions, "/dashboard": overview} {
		if !strings.Contains(page, managedByEngineMark) {
			t.Fatalf("%s does not render the engine's management verdict, so it cannot be sharing "+
				"a reading with the other screen", name)
		}
	}

	runtime, list := since()
	if runtime != 1 || list != 1 {
		t.Errorf("two renders across both screens inside one interval read the engine runtime %d time(s) "+
			"and the list %d time(s); one reading belongs to the interval, not to a screen", runtime, list)
	}
}

// TestEachHalfIsReadAgainOnItsOwnInterval keeps the bounds from becoming a
// freeze, and pins that the two halves are not tied to one number.
func TestEachHalfIsReadAgainOnItsOwnInterval(t *testing.T) {
	commander := newCountingPolicyCommander()
	h, since := countingPolicyHarness(t, commander)

	h.page(t, "/positions")
	h.clock.advance(wantRuntimeInterval)
	h.page(t, "/positions")

	runtime, list := since()
	if runtime != 2 {
		t.Errorf("after a full runtime interval the runtime was read %d time(s); want 2", runtime)
	}
	if list != 1 {
		t.Errorf("the lifecycle listing was read %d time(s) inside its own interval; want 1 — the two "+
			"halves are bounded separately because only one of them touches the write connection", list)
	}

	h.clock.advance(wantLifecycleInterval)
	h.page(t, "/positions")
	if _, list = since(); list != 2 {
		t.Errorf("after a full lifecycle interval the listing was read %d time(s); want 2", list)
	}
}

// TestAnEngineThatStopsAnsweringIsNotAskedOnEveryRedraw is task 3.3's second
// half.
//
// A failed attempt counts against the interval exactly like a successful one.
// The moment the engine cannot answer is the worst moment to open a fresh socket
// to it on every render, and holdings.go already made that call about a broker
// returning 429.
func TestAnEngineThatStopsAnsweringIsNotAskedOnEveryRedraw(t *testing.T) {
	commander := newCountingPolicyCommander()
	h, since := countingPolicyHarness(t, commander)

	h.page(t, "/positions")
	commander.failRuntime(errors.New("engine socket unavailable"))
	commander.failList(errors.New("engine socket unavailable"))
	h.clock.advance(wantLifecycleInterval)
	h.page(t, "/positions") // the one attempt that observes the failure

	runtimeAfterFailure, listAfterFailure := since()
	// Four more redraws, all of them still inside the shorter of the two
	// intervals, so neither half is due.
	for i := 0; i < 4; i++ {
		h.clock.advance(wantRuntimeInterval / 8)
		h.page(t, "/positions")
	}

	runtime, list := since()
	if runtime != runtimeAfterFailure || list != listAfterFailure {
		t.Errorf("four redraws after a failed reading re-dialled the engine (runtime %d→%d, list %d→%d); "+
			"a failure is a result and is served for the interval like any other",
			runtimeAfterFailure, runtime, listAfterFailure, list)
	}
}

// TestAFailedRuntimeReadingIsNotMaskedByThePreviousSuccess is task 3.3's first
// half, and the reason this cache does not copy holdingsCache wholesale.
//
// holdingsCache keeps its last good reading when a refresh fails, because a
// position size an hour old is still information about somebody's money. An
// adoption snapshot is not: it is a claim about what is protecting the account
// right now, and serving a dead engine's last success would have the screen
// assert an effective policy nobody is maintaining. The operator-console spec
// forbids exactly that — "runtime unavailable인 non-managed 행은 desired를
// effective로 위장하지 않고 UNKNOWN".
func TestAFailedRuntimeReadingIsNotMaskedByThePreviousSuccess(t *testing.T) {
	commander := newCountingPolicyCommander()
	h, _ := countingPolicyHarness(t, commander)

	withEngine := candidateRow(t, h.page(t, "/positions"))
	if !strings.Contains(withEngine, engineStopPercentMark) {
		t.Fatalf("the fixture never rendered the running engine's policy width, so this test cannot tell "+
			"a stale reading from a fresh one; looked for %q in:\n%s", engineStopPercentMark, withEngine)
	}

	commander.failRuntime(errors.New("engine socket unavailable"))
	h.clock.advance(wantRuntimeInterval)
	withoutEngine := candidateRow(t, h.page(t, "/positions"))

	if strings.Contains(withoutEngine, engineStopPercentMark) {
		t.Error("the row still shows the policy width of an engine that has stopped answering; " +
			"a cached success must not outlive the attempt that failed")
	}
	if !strings.Contains(withoutEngine, engineUnknownMark) {
		t.Errorf("a row that could not read the engine must say so; %q is missing from:\n%s",
			engineUnknownMark, withoutEngine)
	}
}

// TestAFailedLifecycleReadingIsNotMaskedByThePreviousSuccess is the other half of
// the same rule, and the one the first version of this file left unproven.
//
// It is the dangerous half. positionRow.Managed reads only the cached lifecycle
// state, so a revived stale listing makes a row claim 엔진 관리 — protected by the
// engine — on the evidence of an engine that has stopped answering. The Runtime
// half being honest does not save it: ProjectManagement checks Managed before it
// ever looks at EffectiveKnown.
func TestAFailedLifecycleReadingIsNotMaskedByThePreviousSuccess(t *testing.T) {
	commander := newCountingPolicyCommander()
	h, _ := countingPolicyHarness(t, commander)

	withEngine := managedRow(t, h.page(t, "/positions"))
	if !strings.Contains(withEngine, managedByEngineMark) {
		t.Fatalf("the fixture never rendered a row managed by the engine, so this test cannot tell a "+
			"stale lifecycle listing from a fresh one; looked for %q in:\n%s", managedByEngineMark, withEngine)
	}

	commander.failList(errors.New("engine socket unavailable"))
	h.clock.advance(wantLifecycleInterval)
	withoutEngine := managedRow(t, h.page(t, "/positions"))

	if strings.Contains(withoutEngine, managedByEngineMark) {
		t.Error("the row still claims the engine is managing it, on a lifecycle listing the engine " +
			"could not serve. A cached success must not outlive the attempt that failed — this is the " +
			"direction that puts a false protection claim on the screen")
	}
	if !strings.Contains(withoutEngine, managementUnknownMark) {
		t.Errorf("a row whose lifecycle could not be read must say so; %q is missing from:\n%s",
			managementUnknownMark, withoutEngine)
	}
}

// TestAReconcileBlockIsNotHeldForTheLifecycleInterval is why Runtime has an
// interval of its own.
//
// The blocks Runtime carries are live reconcile-tracker state. Holding them for
// the lifecycle interval would have a candidate holding render 편입 예약됨 — queued
// for protection — while the engine has in fact stopped adopting it. That is
// optimistic-wrong, which is the direction that matters on this screen.
func TestAReconcileBlockIsNotHeldForTheLifecycleInterval(t *testing.T) {
	commander := newCountingPolicyCommander()
	h, since := countingPolicyHarness(t, commander)

	before := candidateRow(t, h.page(t, "/positions"))
	if !strings.Contains(before, adoptionPendingMark) {
		t.Fatalf("the candidate row does not start out queued for adoption, so a block cannot be "+
			"observed to change it; looked for %q in:\n%s", adoptionPendingMark, before)
	}

	commander.blockSymbol("kr", "000660")
	h.clock.advance(wantRuntimeInterval)
	after := candidateRow(t, h.page(t, "/positions"))

	if !strings.Contains(after, reconcileBlockedMark) {
		t.Errorf("a reconcile block that the engine already has is still invisible one runtime interval "+
			"later; the screen is telling the operator this holding is queued for protection when "+
			"adoption has stopped. Looked for %q in:\n%s", reconcileBlockedMark, after)
	}
	if _, list := since(); list != 1 {
		t.Errorf("picking the block up cost %d lifecycle listing(s); want 1 — the runtime half is meant "+
			"to move without dragging the half that contends for the write connection", list)
	}
}

// TestAnAbandonedRenderDoesNotPoisonTheSharedReading is the review's reproduced
// defect.
//
// A reading is shared; the request that happens to take it is not. If a browser
// abandons a render — a reload during load, a closed tab, a speculative
// navigation — recording that cancellation would blank the protection line on
// both screens, on a healthy engine, for a whole interval, and reloading could
// not fix it because the interval is the point.
func TestAnAbandonedRenderDoesNotPoisonTheSharedReading(t *testing.T) {
	commander := newCountingPolicyCommander()
	cache := newPositionPolicyCache(commander, wantLifecycleInterval, wantRuntimeInterval)
	clk := newFakeClock()

	abandoned, cancel := context.WithCancel(context.Background())
	cancel()
	reading := cache.read(abandoned, clk.Now())

	if reading.RuntimeErr != nil || reading.StatesErr != nil {
		t.Fatalf("the abandoned render's own reading carries the cancellation (runtime=%v states=%v); "+
			"the refresh must run on a context of its own", reading.RuntimeErr, reading.StatesErr)
	}

	// Well inside both intervals: a healthy render must be served the same good
	// reading rather than the abandoned request's failure.
	clk.advance(wantRuntimeInterval / 2)
	healthy := cache.read(context.Background(), clk.Now())
	if healthy.RuntimeErr != nil || healthy.StatesErr != nil || !healthy.Runtime.EffectiveKnown ||
		len(healthy.States) == 0 {
		t.Errorf("a healthy render was served a poisoned reading: runtimeErr=%v statesErr=%v "+
			"effectiveKnown=%v states=%d", healthy.RuntimeErr, healthy.StatesErr,
			healthy.Runtime.EffectiveKnown, len(healthy.States))
	}
}

// TestConcurrentRendersCostOneReading pins task 3.4 and the SHALL that says so.
//
// The mutex is held across the reads on purpose (position_policy_cache.go). The
// obvious future refactor — release it during the RPCs so a slow engine does not
// block the second page — would break the bound with every other test still
// green, which is the defect class this change exists to fix.
func TestConcurrentRendersCostOneReading(t *testing.T) {
	commander := newCountingPolicyCommander()
	cache := newPositionPolicyCache(commander, wantLifecycleInterval, wantRuntimeInterval)
	now := newFakeClock().Now()

	// A read that returns instantly can be shared by accident: eight goroutines
	// queued on a mutex serialize, and each one finds the reading already taken
	// even if the implementation stamps it late. Widening the read is what makes
	// the difference observable.
	commander.duringRead = func() { time.Sleep(20 * time.Millisecond) }

	const renders = 8
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < renders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			cache.read(context.Background(), now)
		}()
	}
	close(start)
	wg.Wait()

	runtime, list := commander.reads()
	if runtime != 1 || list != 1 {
		t.Errorf("%d renders arriving together cost runtime=%d list=%d engine read(s); want one of each",
			renders, runtime, list)
	}
}

// TestAStaleLifecycleListCanOnlyWithholdAVerdictNeverInventOne is the property
// that replaces the withdrawn "one moment" claim.
//
// c.positions() reads the console's own journal handle on every render, so the
// row a reading is joined against is always fresher than the reading — there was
// never one moment to preserve. What matters is the direction of the
// disagreement: a listing older than the journal must be unable to assert
// protection that is not there.
func TestAStaleLifecycleListCanOnlyWithholdAVerdictNeverInventOne(t *testing.T) {
	commander := newCountingPolicyCommander()
	h, _ := countingPolicyHarness(t, commander)

	// A listing that has never heard of the journal's managed position — which is
	// what a listing taken before the engine adopted it looks like.
	commander.forget()

	row := managedRow(t, h.page(t, "/positions"))
	if strings.Contains(row, managedByEngineMark) {
		t.Error("a lifecycle listing that does not know this position produced a managed verdict")
	}
	if !strings.Contains(row, managementUnknownMark) {
		t.Errorf("a row the listing does not cover must render as unknown, not as anything else:\n%s", row)
	}
}

// releaseAPolicy drives the real capability flow — select, preview, confirm,
// apply — and fails the test unless the engine actually applied something.
func releaseAPolicy(t *testing.T, h *dashboardHarness) {
	t.Helper()
	selection := actionToken(t, h.page(t, "/position-management"), "자동관리 해제")
	previewPage := body(t, h.post(t, "/position-management/preview", url.Values{
		"csrf": {h.csrf}, "action_token": {selection},
	}))
	// The harness client follows the redirect, so the applied notice on the screen
	// it lands on is what says the engine accepted it.
	applied := body(t, h.post(t, "/position-management/apply", url.Values{
		"csrf": {h.csrf}, "capability": {applyToken(t, previewPage)}, "confirm": {"yes"},
	}))
	if !strings.Contains(applied, "적용됨") {
		t.Fatalf("apply was refused; the invalidation this test is about never ran:\n%s", applied)
	}
}

// TestASuccessfulPolicyChangeDropsTheCachedReading is task 4.1.
//
// An operator who has just released a position and cannot see it released is
// looking at a bug, not at a cache.
func TestASuccessfulPolicyChangeDropsTheCachedReading(t *testing.T) {
	commander := newCountingPolicyCommander()
	h, _ := countingPolicyHarness(t, commander)

	h.page(t, "/positions")
	beforeRuntime, beforeList := commander.reads()
	h.page(t, "/positions")
	if runtime, list := commander.reads(); runtime != beforeRuntime || list != beforeList {
		t.Fatalf("the reading was not being shared to begin with (runtime %d→%d, list %d→%d), so this "+
			"test cannot show that a mutation drops it", beforeRuntime, runtime, beforeList, list)
	}

	releaseAPolicy(t, h)

	// The command screen reads the engine directly, so measure from after the flow.
	afterFlowRuntime, afterFlowList := commander.reads()
	h.page(t, "/positions")
	runtime, list := commander.reads()
	if runtime == afterFlowRuntime || list == afterFlowList {
		t.Errorf("the render after a successful release reused the cached reading (runtime %d→%d, "+
			"list %d→%d); the lifecycle it draws has just moved",
			afterFlowRuntime, runtime, afterFlowList, list)
	}
}

// TestARefusedPolicyChangeKeepsTheCachedReading is task 4.3.
//
// A refused apply changed nothing, so there is nothing to re-read. Dropping the
// reading on refusal would let a repeated bad request defeat the bound the cache
// exists to hold.
func TestARefusedPolicyChangeKeepsTheCachedReading(t *testing.T) {
	commander := newCountingPolicyCommander()
	h, _ := countingPolicyHarness(t, commander)

	h.page(t, "/positions")
	refused := body(t, h.post(t, "/position-management/apply", url.Values{
		"csrf": {h.csrf}, "capability": {"not-a-capability-this-engine-issued"},
	}))
	// Assert the refusal positively: a 404, a renamed route or any other failure
	// to reach the handler would satisfy a "does not say 적용됨" check while
	// proving nothing.
	if !strings.Contains(refused, "정책 capability 거부") {
		t.Fatalf("this test needs the apply handler to refuse the capability, and it did not:\n%s", refused)
	}

	beforeRuntime, beforeList := commander.reads()
	h.page(t, "/positions")
	if runtime, list := commander.reads(); runtime != beforeRuntime || list != beforeList {
		t.Errorf("a refused apply dropped the shared reading (runtime %d→%d, list %d→%d); nothing "+
			"changed, so nothing needed re-reading", beforeRuntime, runtime, beforeList, list)
	}
}

// TestTheCacheHandsOutItsOwnCopies keeps one render from editing what the next
// one reads. The rows built from a reading outlive the lock.
func TestTheCacheHandsOutItsOwnCopies(t *testing.T) {
	commander := newCountingPolicyCommander()
	commander.blockSymbol("kr", "000660")
	cache := newPositionPolicyCache(commander, wantLifecycleInterval, wantRuntimeInterval)
	now := newFakeClock().Now()

	first := cache.read(context.Background(), now)
	if len(first.States) == 0 || len(first.Runtime.Blocks) == 0 {
		t.Fatal("the fixture handed out no states or no blocks, so aliasing cannot be observed")
	}
	first.States[0].Status = positionpolicy.StatusReleased
	first.Runtime.Blocks[0].Symbol = "TAMPERED"

	second := cache.read(context.Background(), now)
	if second.States[0].Status == positionpolicy.StatusReleased {
		t.Error("one render's rows reached into the shared lifecycle listing; the states must be copied out")
	}
	if second.Runtime.Blocks[0].Symbol == "TAMPERED" {
		t.Error("one render reached into the shared reconcile blocks; they must be copied out")
	}
}

// TestTheCommandScreenReadsTheEngineDirectly is task 5.3 and design D6.
//
// /position-management issues capabilities bound to an exact before-state. The
// list an operator acts on has to be a reading taken for that action, not one
// shared with whatever a trading tab rendered up to an interval ago. That screen
// has no meta refresh, so it costs only what a person opening it costs.
func TestTheCommandScreenReadsTheEngineDirectly(t *testing.T) {
	commander := newCountingPolicyCommander()
	h, _ := countingPolicyHarness(t, commander)

	h.page(t, "/positions") // fills the shared reading
	beforeRuntime, beforeList := commander.reads()
	h.page(t, "/position-management")

	runtime, list := commander.reads()
	if runtime == beforeRuntime || list == beforeList {
		t.Errorf("the command screen served itself from the trading screens' cache (runtime %d→%d, "+
			"list %d→%d); an operator's action must be decided on a reading taken for it",
			beforeRuntime, runtime, beforeList, list)
	}
}

// TestTheCacheIntervalsAreTheirOwnConstants pins design D2.
//
// They must not be derived from holdingsTTL: that is the broker budget, and a
// change to it would silently change how hard the console leans on the engine's
// write connection, and the reverse. They must not be derived from each other
// either — only one of the two reads contends for that connection, and only the
// other carries live reconcile state.
func TestTheCacheIntervalsAreTheirOwnConstants(t *testing.T) {
	if engineLifecycleInterval != wantLifecycleInterval {
		t.Errorf("engineLifecycleInterval = %s, want %s", engineLifecycleInterval, wantLifecycleInterval)
	}
	if engineRuntimeInterval != wantRuntimeInterval {
		t.Errorf("engineRuntimeInterval = %s, want %s", engineRuntimeInterval, wantRuntimeInterval)
	}
	if engineLifecycleInterval == engineRuntimeInterval {
		t.Error("the two halves are back on one number; each interval is set by its own argument")
	}
	source := packageFiles(t)["position_policy_cache.go"]
	for _, literal := range []string{
		"const engineLifecycleInterval = 30 * time.Second",
		"const engineRuntimeInterval = 5 * time.Second",
	} {
		if !strings.Contains(source, literal) {
			t.Errorf("%q is no longer a literal in position_policy_cache.go — if an interval is now "+
				"derived from holdingsTTL or from the other one, two decisions are sharing a constant "+
				"again", literal)
		}
	}
}

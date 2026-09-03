//go:build tossos_testseams

package strategyworker

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategycoordinator"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

// 이 파일은 태스크 5.7 이다 — 여덟 worker 와 두 조정자를 **함께** 세우고
// race·goroutine 누수·큐 압력·가짜 시계·고장 주입을 잰다.
//
// **이것은 생산 배선이 아니다.** `design.md:255` 가 정한 순서가 그렇다: 여덟과
// 둘을 dormant/shadow 로 세우고, 공유 dispatch 로 가는 길은 선결 조건이 닫힐
// 때까지 spy 로 막은 채 고장 주입·race·큐 압력·결정적 재생을 먼저 검증한다.
// 여기 서는 레인들은 시험이 켠 사본이고(생산 진입점은 여덟을 언제나 OFF 로 준다),
// 생산 경로를 이 타입들 위로 옮기는 일은 5.1.2/5.2 다.
//
// **한 리허설이 증명하지 못하는 것도 적는다.** 이 파일은 타입들이 동시에 돌 때
// 서로를 밟지 않는다는 것을 증명한다. 오늘의 엔진이 그렇다는 것은 증명하지
// 않는다 — 엔진 쪽 빈칸은 두 FLM 번들의 branch-test-map 에 좌표로 적혀 있다.

// dispatchSpy 는 조정자가 고른 것이 공유 경계로 건너가려는 자리에 선 감시자다.
//
// **왜 진짜 handoff 가 아닌가.** `design.md:255` 는 "공유 dispatch handoff 를
// spy 로 막으라"고 하지만, 그 spy 를 이 패키지에 둘 수는 없다:
// `strategyhandoff/dependency_closure_test.go` 의 importer census 가 그 seam 을
// 들여올 수 있는 패키지를 엔진 하나로 못 박고 있다(들여오는 것이 `Admit` 을
// 부를 수 있는 유일한 길이라서). 실제로 이 파일의 첫 판본이 그것을 들여왔고,
// 그 census 가 잡았다. 그러니 **이 리허설은 seam 에 닿을 수조차 없고**, 그것이
// 곧 "아무것도 건너가지 않았다"의 증거다 — 이 spy 가 아니라 census 가 지킨다.
// 진짜 handoff 를 spy 로 막는 배선은 엔진 쪽 일이고 태스크 5.1.2/5.2 다.
//
// 세는 것이 두 가지다: 경계를 **본** 횟수와 경계를 **지나온** 값의 수. 둘을
// 따로 세지 않으면 "거절이라 아무것도 안 나갔다"와 "아무도 부르지 않았다"가
// 같은 0 이 된다.
type dispatchSpy struct {
	mu        sync.Mutex
	observed  int
	delivered []string
}

func (spy *dispatchSpy) take(outcome strategycoordinator.Outcome) {
	spy.mu.Lock()
	defer spy.mu.Unlock()
	spy.observed++
	if outcome.Closed() {
		return
	}
	for _, selection := range outcome.Selections {
		spy.delivered = append(spy.delivered, selection.LineageIdentity)
	}
}

func (spy *dispatchSpy) counts() (int, int) {
	spy.mu.Lock()
	defer spy.mu.Unlock()
	return spy.observed, len(spy.delivered)
}

// rehearsal 은 여덟 레인 · 두 조정자 · 하나의 spy 를 함께 든 무대다.
type rehearsal struct {
	clk    *clock.Fake
	lanes  []*Lane
	coords map[strategyrouter.Market]*strategycoordinator.MarketCoordinator
	spy    *dispatchSpy
}

// rehearse 는 생산 목록과 **같은 순서**로 여덟 레인을 세운다.
//
// 순서를 생산과 맞추는 이유는 아래 시험들이 위치로 레인을 가리키기 때문이다.
// 순서가 갈리면 "레인 3 이 잠겼다"가 다른 레인을 뜻하게 된다.
func rehearse(t *testing.T, policy RuntimePolicy) *rehearsal {
	t.Helper()
	fake := clock.NewFake(laneNow)
	lanes := make([]*Lane, 0, len(ProductionWorkers()))
	for _, worker := range ProductionWorkers() {
		key := worker.Key()
		lanes = append(lanes, newLane(workerFor(t, key.Market, key.Family), policy, fake))
	}
	if len(lanes) != 8 {
		t.Fatalf("the golden fixes eight workers and this rehearsal built %d", len(lanes))
	}
	coords := make(map[strategyrouter.Market]*strategycoordinator.MarketCoordinator, 2)
	for _, market := range allMarkets() {
		coords[market] = strategycoordinator.NewMarketCoordinator(market, laneNow)
	}
	return &rehearsal{clk: fake, lanes: lanes, coords: coords, spy: &dispatchSpy{}}
}

// drive 는 여덟 레인을 여덟 goroutine 으로 **동시에** 한 사이클씩 돌린다.
//
// goroutine 하나가 여덟을 차례로 도는 판본은 동시성을 전혀 증명하지 못한다 —
// 이 저장소가 이미 한 번 그 함정을 밟았다(a099 R1 은 핸들 하나를 공유해 세
// 라운드 동안 순차만 증명했다). 여기서는 레인마다 goroutine 이 하나이고,
// 모두 같은 출발 신호를 기다렸다가 함께 출발한다.
func (stage *rehearsal) drive(t *testing.T, step func(index int) Step) []BoundedCycle {
	t.Helper()
	results := make([]BoundedCycle, len(stage.lanes))
	starts := make([]Start, len(stage.lanes))
	gate := make(chan struct{})
	var wg sync.WaitGroup
	for index, lane := range stage.lanes {
		if got := lane.Offer(); got != TriggerEnqueued {
			t.Fatalf("lane %d refused its trigger: %s", index, got)
		}
		wg.Add(1)
		go func(index int, lane *Lane) {
			defer wg.Done()
			<-gate
			results[index], starts[index] = lane.RunBounded(context.Background(), Input{}, step(index))
		}(index, lane)
	}
	close(gate)
	wg.Wait()
	for index, start := range starts {
		if start != StartAdmitted {
			t.Fatalf("lane %d was not admitted: %s", index, start)
		}
	}
	return results
}

// 여덟 레인이 동시에 돌고, 한 레인의 고장은 그 레인 안에만 남는다.
//
// 이 시험의 이빨은 **이유 문자열**에 있다. "이웃이 건강한가"만 보면 상태를
// 공유하는 구현도 통과할 수 있다(고장 난 레인이 하나면 나머지는 어차피
// 건강하다). 여기서는 잠긴 레인마다 **자기 번호가 박힌 이유**를 요구하므로,
// 상태가 섞이면 이유가 서로 건너간다.
func TestEightLanesRunConcurrentlyAndEachFaultStaysInsideItsOwnLane(t *testing.T) {
	stage := rehearse(t, ProductionRuntimePolicy())
	reason := func(index int) string { return fmt.Sprintf("lane %d failed on its own evidence", index) }
	kind := func(index int) string {
		switch index % 3 {
		case 1:
			return "error"
		case 2:
			return "panic"
		}
		return "ok"
	}

	results := stage.drive(t, func(index int) Step {
		return func(context.Context, Input) (Cycle, error) {
			switch kind(index) {
			case "error":
				return Cycle{}, errors.New(reason(index))
			case "panic":
				panic(reason(index))
			}
			return Cycle{Outcome: OutcomeEmitted}, nil
		}
	})

	for index, lane := range stage.lanes {
		switch kind(index) {
		case "ok":
			if lane.Health() != LaneHealthy || lane.Latched() || lane.ConsecutiveFailures() != 0 {
				t.Errorf("lane %d succeeded and is %s (latched=%v failures=%d)",
					index, lane.Health(), lane.Latched(), lane.ConsecutiveFailures())
			}
			if results[index].Err != nil {
				t.Errorf("lane %d succeeded and reported %v", index, results[index].Err)
			}
		case "error":
			if !lane.Latched() || lane.FirstFailure() != reason(index) {
				t.Errorf("lane %d kept %q, want %q (latched=%v)",
					index, lane.FirstFailure(), reason(index), lane.Latched())
			}
			if results[index].Abnormal {
				t.Errorf("lane %d returned an ordinary error and was recorded as abnormal", index)
			}
		case "panic":
			if !lane.Latched() || !results[index].Abnormal {
				t.Errorf("lane %d panicked and is latched=%v abnormal=%v",
					index, lane.Latched(), results[index].Abnormal)
			}
			if got := lane.FirstFailure(); got == "" || !strings.Contains(got, reason(index)) {
				t.Errorf("lane %d kept %q, which does not carry its own panic %q", index, got, reason(index))
			}
		}
	}
}

// 사이클이 끝나면 그 사이클이 만든 goroutine 도 끝난다.
//
// 감시견은 goroutine 을 둘 만든다(사이클 하나, 감시견 하나). 감시견 쪽은
// `defer cancelWatchdog()` 가 거두지만, 그것을 **재어 보지 않으면** 여덟 레인이
// 주기마다 두 개씩 흘리는 것을 아무도 모른다.
func TestAFinishedCycleLeavesNoGoroutineBehind(t *testing.T) {
	stage := rehearse(t, ProductionRuntimePolicy())
	before := stableGoroutines(t)

	for round := 0; round < 4; round++ {
		stage.drive(t, func(int) Step {
			return func(context.Context, Input) (Cycle, error) { return Cycle{Outcome: OutcomeEmitted}, nil }
		})
		stage.clk.Advance(stage.lanes[0].Policy().Cadence())
	}

	if after := settledGoroutines(t, before); after > before {
		t.Fatalf("32 cycles left %d goroutines behind (%d → %d)", after-before, before, after)
	}
}

// 버려진 사이클의 goroutine 은 **일부러** 남는다. 그 사실을 재어서 못 박는다.
//
// 엔진도 같다: 마감 시한을 넘긴 사이클을 취소하지 않고 결과만 안 읽는다.
// 그러니 "누수 0" 이라고 쓰면 거짓이다. 남는 것은 정확히 **일**이 안 끝난
// goroutine 하나이고, 일이 끝나면 사라진다 — 감시견 쪽은 남지 않는다.
func TestAnAbandonedCycleKeepsItsOwnGoroutineUntilTheStepReturns(t *testing.T) {
	stage := rehearse(t, ProductionRuntimePolicy())
	lane := stage.lanes[0]
	before := stableGoroutines(t)

	release := make(chan struct{})
	admit(t, lane)
	done := make(chan struct{})
	go func() {
		lane.RunBounded(context.Background(), Input{}, func(context.Context, Input) (Cycle, error) {
			<-release
			return Cycle{}, nil
		})
		close(done)
	}()
	if !stage.clk.WaitForSleepers(1, 2*time.Second) {
		t.Fatal("no watchdog registered a sleep")
	}
	stage.clk.Advance(lane.Policy().CycleDeadline())
	<-done

	stranded := settledGoroutines(t, before+1)
	if stranded != before+1 {
		t.Fatalf("an abandoned cycle should strand exactly one goroutine; %d → %d", before, stranded)
	}
	close(release)
	if after := settledGoroutines(t, before); after > before {
		t.Fatalf("the stranded goroutine did not end when its step returned: %d → %d", before, after)
	}
}

// stableGoroutines 는 개수가 **멈출 때까지** 기다린 뒤 그 값을 기준선으로 준다.
//
// 그냥 `runtime.NumGoroutine()` 을 부르면 앞 시험이 끝내는 중인 goroutine 이
// 섞여 기준선이 부풀고, 그러면 하나가 늘어난 것을 못 본다 — 2026-09-02 에 이
// 시험이 정확히 그렇게 깜빡였다(4 → 4). 값이 연속으로 같은지를 요구해서
// "가라앉았다"를 재고, 그래야 아래 ±1 비교가 뜻을 갖는다.
func stableGoroutines(t *testing.T) int {
	t.Helper()
	const settled = 5
	deadline := time.Now().Add(2 * time.Second)
	previous, same := -1, 0
	for {
		count := runtime.NumGoroutine()
		if count == previous {
			same++
			if same >= settled {
				return count
			}
		} else {
			previous, same = count, 1
		}
		if time.Now().After(deadline) {
			return count
		}
		time.Sleep(2 * time.Millisecond)
	}
}

// settledGoroutines 는 개수가 want 이하로 가라앉기를 잠깐 기다린 뒤 센다.
//
// 기다림이 필요한 이유는 끝난 goroutine 이 즉시 사라지지 않기 때문이다. 한 번만
// 세면 깜빡이고, 깜빡이는 시험은 다음 사람에게 무시당한다. 가라앉으면 **곧바로**
// 돌려주므로, 통과하는 경우에 시간을 쓰지 않는다.
func settledGoroutines(t *testing.T, want int) int {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for {
		count := runtime.NumGoroutine()
		if count <= want || time.Now().After(deadline) {
			return count
		}
		time.Sleep(2 * time.Millisecond)
	}
}

// 큐 압력: 투입이 아무리 몰려도 막히지 않고, 들어간 것과 버린 것의 합이 정확하다.
//
// 합을 보는 것이 요점이다. "깊이를 안 넘는다"만 보면 투입을 **잃는** 구현도
// 통과한다 — 세지 않고 버리면 깊이는 언제나 만족된다.
func TestQueuePressureNeverBlocksAndNeverLosesCount(t *testing.T) {
	stage := rehearse(t, ProductionRuntimePolicy())
	lane := stage.lanes[0]
	const producers, each = 16, 100

	var mu sync.Mutex
	tally := map[Trigger]int{}
	var wg sync.WaitGroup
	started := time.Now()
	for producer := 0; producer < producers; producer++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for offer := 0; offer < each; offer++ {
				got := lane.Offer()
				mu.Lock()
				tally[got]++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	elapsed := time.Since(started)

	total := 0
	for _, count := range tally {
		total += count
	}
	if total != producers*each {
		t.Fatalf("%d offers produced %d results", producers*each, total)
	}
	if tally[TriggerDisabled] != 0 {
		t.Fatalf("an unlatched lane refused %d offers as DISABLED", tally[TriggerDisabled])
	}
	if got := uint64(tally[TriggerFull]); got != lane.Dropped() {
		t.Fatalf("%d offers came back FULL and the drop counter says %d", got, lane.Dropped())
	}
	if tally[TriggerEnqueued]+tally[TriggerFull] != producers*each {
		t.Fatalf("enqueued %d + full %d does not account for %d offers",
			tally[TriggerEnqueued], tally[TriggerFull], producers*each)
	}
	if lane.Pending() > lane.Policy().QueueDepth() {
		t.Fatalf("the slot holds %d, over its depth %d", lane.Pending(), lane.Policy().QueueDepth())
	}
	// 막지 않는다는 것은 시간으로만 볼 수 있다. 1600 번의 비차단 offer 가
	// 초 단위로 걸린다면 어딘가에서 기다린 것이다.
	if elapsed > 5*time.Second {
		t.Fatalf("1600 non-blocking offers took %v", elapsed)
	}
}

// 결정적 재생: 같은 제안 넷을 어떤 순서로 넣어도 조정 결과가 같다.
//
// 조정자는 도착 순서가 아니라 봉인된 레인 신원으로 정렬한다. 그 성질이 없으면
// 같은 주기가 매번 다른 가족을 고를 수 있고, 그것은 재생이 불가능하다는 뜻이다.
func TestTheSameFourProposalsArbitrateIdenticallyInEveryArrivalOrder(t *testing.T) {
	f := newFixture(allKRFamilies()...)
	envelopes := make([]strategycoordinator.Envelope, 0, 4)
	for _, laneID := range allKRLanes() {
		input := f.input(t, laneID)
		envelopes = append(envelopes, strategycoordinator.Envelope{
			Scope: input.Scope, SnapshotDigest: input.SnapshotDigest, Proposal: input.Proposal})
	}

	var want string
	shuffler := rand.New(rand.NewSource(20260902))
	for attempt := 0; attempt < 32; attempt++ {
		order := shuffler.Perm(len(envelopes))
		coordinator := strategycoordinator.NewMarketCoordinator(strategyrouter.MarketKR, f.now)
		for _, index := range order {
			if admission := coordinator.Submit(envelopes[index]); !admission.Admitted {
				t.Fatalf("attempt %d: envelope %d was refused: %s", attempt, index, admission.Detail)
			}
		}
		got := renderOutcome(coordinator.Arbitrate())
		if attempt == 0 {
			want = got
			continue
		}
		if got != want {
			t.Fatalf("arrival order %v produced a different outcome.\n got: %s\nwant: %s", order, got, want)
		}
	}
	if want == "" {
		t.Fatal("the comparison never ran, so it proves nothing")
	}
}

func renderOutcome(outcome strategycoordinator.Outcome) string {
	rendered := fmt.Sprintf("closed=%v overflow=%v refusal=%s drops=%d selections=%d",
		outcome.Closed(), outcome.Overflow, outcome.Refusal, outcome.Drops, len(outcome.Selections))
	for _, selection := range outcome.Selections {
		rendered += fmt.Sprintf(" | %s/%s/%s ppm=%d id=%s owner=%v",
			selection.Scope.Market, selection.Scope.Symbol, selection.Family,
			selection.ScorePPM, selection.LineageIdentity, selection.ExistingOwner)
	}
	return rendered
}

// 같은 종목에 네 가족이 **동시에** 제안해도 그 범위가 고르는 것은 하나뿐이고,
// 그 하나는 순차로 넣었을 때와 같다.
func TestFourFamiliesProposingOneSymbolAtOnceStillSelectExactlyOne(t *testing.T) {
	f := newFixture(allKRFamilies()...)
	inputs := make([]Input, 0, 4)
	for _, laneID := range allKRLanes() {
		inputs = append(inputs, f.input(t, laneID))
	}

	sequential := strategycoordinator.NewMarketCoordinator(strategyrouter.MarketKR, f.now)
	for _, input := range inputs {
		sequential.Submit(strategycoordinator.Envelope{Scope: input.Scope,
			SnapshotDigest: input.SnapshotDigest, Proposal: input.Proposal})
	}
	want := renderOutcome(sequential.Arbitrate())

	concurrent := strategycoordinator.NewMarketCoordinator(strategyrouter.MarketKR, f.now)
	gate := make(chan struct{})
	var wg sync.WaitGroup
	for _, input := range inputs {
		wg.Add(1)
		go func(input Input) {
			defer wg.Done()
			<-gate
			concurrent.Submit(strategycoordinator.Envelope{Scope: input.Scope,
				SnapshotDigest: input.SnapshotDigest, Proposal: input.Proposal})
		}(input)
	}
	close(gate)
	wg.Wait()

	outcome := concurrent.Arbitrate()
	if got := renderOutcome(outcome); got != want {
		t.Fatalf("concurrent submission chose differently.\n got: %s\nwant: %s", got, want)
	}
	if len(outcome.Selections) != 1 {
		t.Fatalf("one owner scope produced %d selections", len(outcome.Selections))
	}
}

// 한 소유자 범위가 경계로 보내는 것은 최대 하나다.
//
// 골든 `queue.selected_limit` 가 그것이고, 그 상한이 지켜지지 않으면 아래쪽
// `strategyhandoff.Capacity` 는 거절만 늘어난다. 상한 자체는 저 패키지가 자기
// 시험으로 지키고, 여기서는 **조정자가 그보다 많이 만들지 않는지**를 본다 —
// 위 dispatchSpy 주석대로 이 패키지는 그 경계를 들여올 수 없다.
func TestOneOwnerScopeHandsAtMostOneThingToTheSeam(t *testing.T) {
	f := newFixture(allKRFamilies()...)
	stage := rehearse(t, ProductionRuntimePolicy())
	coordinator := stage.coords[strategyrouter.MarketKR]
	identities := make(map[string]bool, 4)
	for _, laneID := range allKRLanes() {
		input := f.input(t, laneID)
		identities[input.Proposal.Result.Lineage.Identity] = true
		if admission := coordinator.Submit(strategycoordinator.Envelope{Scope: input.Scope,
			SnapshotDigest: input.SnapshotDigest, Proposal: input.Proposal}); !admission.Admitted {
			t.Fatalf("envelope for %s was refused: %s", laneID, admission.Detail)
		}
	}

	outcome := coordinator.Arbitrate()
	stage.spy.take(outcome)
	observed, delivered := stage.spy.counts()
	if observed != 1 {
		t.Fatalf("the seam was observed %d times", observed)
	}
	if delivered != 1 {
		t.Fatalf("four families in one owner scope handed %d things onward, want 1", delivered)
	}
	if len(outcome.Selections) != 1 {
		t.Fatalf("one owner scope produced %d selections", len(outcome.Selections))
	}
	if !identities[outcome.Selections[0].LineageIdentity] {
		t.Fatalf("the coordinator selected an identity this rehearsal never submitted: %s",
			outcome.Selections[0].LineageIdentity)
	}
}

// 두 조정자는 서로를 닫지 않는다.
//
// KR 을 닫아 놓고 US 가 계속 고르는지 본다. 하나로 뭉친 구현이라면 KR 의 결함이
// US 의 진입까지 멈추고, 그것이 이 change 가 없애려는 시장 전체 결합이다.
func TestAFaultInOneMarketCoordinatorDoesNotCloseTheOther(t *testing.T) {
	stage := rehearse(t, ProductionRuntimePolicy())
	us := newMarketFixture(strategyrouter.MarketUS, allKRFamilies()...)

	// KR 조정자에 남의 시장 범위를 넣어 닫는다.
	foreign := us.input(t, allLanes(strategyrouter.MarketUS)[0])
	if admission := stage.coords[strategyrouter.MarketKR].Submit(strategycoordinator.Envelope{
		Scope: foreign.Scope, SnapshotDigest: foreign.SnapshotDigest, Proposal: foreign.Proposal,
	}); admission.Admitted {
		t.Fatal("the KR coordinator admitted a US scope")
	}
	if !stage.coords[strategyrouter.MarketKR].Arbitrate().Closed() {
		t.Fatal("the KR coordinator did not close on a foreign scope")
	}

	good := us.input(t, allLanes(strategyrouter.MarketUS)[0])
	if admission := stage.coords[strategyrouter.MarketUS].Submit(strategycoordinator.Envelope{
		Scope: good.Scope, SnapshotDigest: good.SnapshotDigest, Proposal: good.Proposal,
	}); !admission.Admitted {
		t.Fatalf("the US coordinator refused its own scope: %s", admission.Detail)
	}
	outcome := stage.coords[strategyrouter.MarketUS].Arbitrate()
	if outcome.Closed() || len(outcome.Selections) != 1 {
		t.Fatalf("KR closing took US with it: closed=%v selections=%d",
			outcome.Closed(), len(outcome.Selections))
	}
	if outcome.Selections[0].Scope.Market != strategyrouter.MarketUS {
		t.Fatalf("the US coordinator selected a %s scope", outcome.Selections[0].Scope.Market)
	}
}

// 종료: 도는 사이클을 취소하면 실패로 세지 않고, 어떤 레인도 잠기지 않는다.
func TestShutdownCancelsEveryLaneWithoutLatchingAny(t *testing.T) {
	stage := rehearse(t, ProductionRuntimePolicy())
	ctx, cancel := context.WithCancel(context.Background())
	blocked := make(chan struct{})
	entered := make(chan struct{}, len(stage.lanes))

	var wg sync.WaitGroup
	for index, lane := range stage.lanes {
		if got := lane.Offer(); got != TriggerEnqueued {
			t.Fatalf("lane %d refused its trigger: %s", index, got)
		}
		wg.Add(1)
		go func(lane *Lane) {
			defer wg.Done()
			lane.RunBounded(ctx, Input{}, func(context.Context, Input) (Cycle, error) {
				entered <- struct{}{}
				<-blocked
				return Cycle{}, nil
			})
		}(lane)
	}
	for range stage.lanes {
		<-entered
	}
	cancel()
	wg.Wait()
	close(blocked)

	for index, lane := range stage.lanes {
		if lane.Latched() || lane.ConsecutiveFailures() != 0 {
			t.Errorf("lane %d treated shutdown as a failure: latched=%v failures=%d",
				index, lane.Latched(), lane.ConsecutiveFailures())
		}
	}
}

// 재기동: **기록 없이** 세운 레인 여덟은 모두 건강하다.
//
// 이 시험은 5.7 이 쓸 때 "재기동이 잠금을 지운다"는 **구멍**을 못 박는 것이었고,
// 5.3.3 이 그 구멍을 닫은 뒤에도 단언은 그대로 옳다 — 다만 뜻이 바뀌었다.
// 잠금은 이제 이 패키지 밖의 durable 기록에 살고, 이 패키지는 기록을 받지 않으면
// 잠금을 **지어낼 수 없다**. 즉 아래가 재는 것은 결함이 아니라 계약이다:
// 기록 없이 태어난 레인은 열려 있다.
//
// 기록에서 태어나는 쪽은 `restore_test.go` 의
// `TestAProductionLaneBornFromADurableRecordIsLatched` 가 잰다. 둘이 함께 있어야
// "잠금의 출처는 기록 하나뿐"이라는 문장이 된다.
func TestALaneBuiltWithoutADurableRecordIsBornUnlatched(t *testing.T) {
	stage := rehearse(t, ProductionRuntimePolicy())
	lane := stage.lanes[0]
	if _, latched := lane.Fail("evidence replay failed", false); !latched {
		t.Fatal("the production threshold is 1 so the first failure must latch")
	}
	if lane.Health() != LaneLatched {
		t.Fatalf("the lane is %s after latching", lane.Health())
	}

	restarted := rehearse(t, ProductionRuntimePolicy())
	for index, fresh := range restarted.lanes {
		if fresh.Health() != LaneHealthy || fresh.Latched() {
			t.Errorf("restarted lane %d is %s (latched=%v)", index, fresh.Health(), fresh.Latched())
		}
	}
	// 그리고 원래 레인은 **여전히** 잠겨 있다. 두 단언이 함께 있어야 이 시험이
	// "재기동이 잠금을 지운다"를 말한다 — 새 레인만 보면 "잠근 적이 없다"와
	// 구별되지 않는다.
	if !lane.Latched() {
		t.Fatal("the original lane unlatched itself, which no code in this package can do")
	}
	if restarted.lanes[0].Key() != lane.Key() {
		t.Fatalf("the restarted lane 0 is %v, not the same key as %v",
			restarted.lanes[0].Key().Parts(), lane.Key().Parts())
	}
}

// 고장 주입표: 설계 고장표(`design.md:198`) 중 이 폐포가 표현할 수 있는 세 줄.
func TestTheLaneFollowsTheDesignFaultTableForEveryKindItCanExpress(t *testing.T) {
	for _, testCase := range []struct {
		name     string
		step     Step
		abnormal bool
		abandon  uint64
	}{
		{"ordinary error", func(context.Context, Input) (Cycle, error) {
			return Cycle{}, errors.New("evidence replay failed")
		}, false, 0},
		{"panic", func(context.Context, Input) (Cycle, error) { panic("evidence replay exploded") }, true, 0},
		{"deadline", nil, true, 1},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			stage := rehearse(t, ProductionRuntimePolicy())
			lane := stage.lanes[3]
			admit(t, lane)
			bounded, start := runFaultCycle(t, stage, lane, testCase.step)
			if start != StartAdmitted {
				t.Fatalf("the cycle was refused: %s", start)
			}
			if bounded.Abnormal != testCase.abnormal {
				t.Errorf("abnormal=%v, want %v", bounded.Abnormal, testCase.abnormal)
			}
			if lane.Abandoned() != testCase.abandon {
				t.Errorf("abandoned=%d, want %d", lane.Abandoned(), testCase.abandon)
			}
			if !lane.Latched() {
				t.Error("the production threshold is 1 so any failure latches")
			}
			for index, peer := range stage.lanes {
				if index == 3 {
					continue
				}
				if peer.Health() != LaneHealthy {
					t.Errorf("peer lane %d is %s", index, peer.Health())
				}
			}
			for market, coordinator := range stage.coords {
				if coordinator.Arbitrate().Closed() {
					t.Errorf("a lane fault closed the %s coordinator", market)
				}
			}
			observed, delivered := stage.spy.counts()
			if observed != 0 || delivered != 0 {
				t.Errorf("a lane fault reached the dispatch seam: observed=%d delivered=%d",
					observed, delivered)
			}
		})
	}
}

// runFaultCycle 은 사이클 하나를 돌린다. step 이 nil 이면 마감 시한 줄이라는
// 뜻이고, 그때만 가짜 시계를 민다 — 다른 줄에서 밀면 잠자는 감시견이 없어
// 시계를 미는 것 자체가 무의미하다.
func runFaultCycle(t *testing.T, stage *rehearsal, lane *Lane, step Step) (BoundedCycle, Start) {
	t.Helper()
	if step != nil {
		return lane.RunBounded(context.Background(), Input{}, step)
	}
	release := make(chan struct{})
	t.Cleanup(func() { close(release) })
	type outcome struct {
		bounded BoundedCycle
		start   Start
	}
	done := make(chan outcome, 1)
	go func() {
		bounded, start := lane.RunBounded(context.Background(), Input{},
			func(context.Context, Input) (Cycle, error) {
				<-release
				return Cycle{}, nil
			})
		done <- outcome{bounded, start}
	}()
	if !stage.clk.WaitForSleepers(1, 2*time.Second) {
		t.Fatal("no watchdog registered a sleep")
	}
	stage.clk.Advance(lane.Policy().CycleDeadline())
	got := <-done
	return got.bounded, got.start
}

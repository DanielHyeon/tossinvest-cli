package main

// a102_engine_ready_test.go is written before the code it tests (openspec change
// a102, tasks 3.4, 3.4b, 3.5 and 3.6).
//
// 형태는 a101이 정한 것과 같다: `runConsole`은 0.0%로 측정되므로(이 change의
// `cmd-tossctl--runconsole/branch-test-map.md`) 판정을 그 안에 쓰면 어떤 테스트도 닿지
// 못한다. 그래서 판정은 전부 인자를 받는 함수에 있고, `runConsole`에는 호출과 출력만 남는다.
//
// 시계는 T1의 `virtualClock`(internal/reconcile/a102_recovery_rate_limit_test.go) 선례를
// 따른다 — 요청된 대기를 **기록**하고 가상 시간을 그만큼 민다. 그래야 "상한 120초를 2초
// 간격으로 관측한다"가 벽시계 없이 산술로 확인된다.

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/enginelock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

// --- the test clock -----------------------------------------------------------

type a102Clock struct {
	now      time.Time
	waits    []time.Duration
	cancelAt int
	cancel   context.CancelFunc
}

func newA102Clock() *a102Clock {
	return &a102Clock{now: time.Date(2026, 8, 13, 2, 3, 29, 0, time.UTC)}
}

// a102ClockRunaway는 이 시계가 참아 주는 대기 횟수다.
//
// 상한이 제대로 걸리면 요청되는 대기는 상한÷간격(60회)이다. 그 열 배를 넘어가면 대기에
// 상한이 없다는 뜻이고, 그때 이 시계는 **멈추는 대신 오류를 돌려준다** — 그러지 않으면
// 상한을 지운 뮤테이션이 테스트를 10분 timeout까지 매달아 두고, 그 결말은 "어느 테스트가
// 죽었는가"를 말해 주지 않는다.
const a102ClockRunaway = 10 * int(engineReadyCap/engineReadyPoll)

var errA102ClockRunaway = errors.New("a102: 대기가 상한을 훨씬 넘겼다 — 상한이 없는 것이다")

func (c *a102Clock) Now() time.Time                  { return c.now }
func (c *a102Clock) Since(t time.Time) time.Duration { return c.now.Sub(t) }

func (c *a102Clock) Sleep(ctx context.Context, d time.Duration) error {
	c.waits = append(c.waits, d)
	if len(c.waits) > a102ClockRunaway {
		return errA102ClockRunaway
	}
	if c.cancelAt > 0 && len(c.waits) == c.cancelAt && c.cancel != nil {
		c.cancel()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	c.now = c.now.Add(d)
	return nil
}

// observations replays a scripted sequence of marker readings, repeating the last
// one forever. It counts how many times it was asked.
type observations struct {
	script []enginelock.Status
	calls  int
}

func (o *observations) observe() enginelock.Status {
	o.calls++
	if len(o.script) == 0 {
		return enginelock.Status{}
	}
	if o.calls-1 < len(o.script) {
		return o.script[o.calls-1]
	}
	return o.script[len(o.script)-1]
}

func runningMarker() enginelock.Status {
	return enginelock.Status{Running: true}
}

func readyMarker(at time.Time) enginelock.Status {
	stamp := at.UTC()
	return enginelock.Status{Running: true, Marker: enginelock.Marker{ReadyAt: &stamp}}
}

// --- D5: the ready signal belongs to a recovery that finished --------------------

// TestRecoveryPublishesReadyOnlyWhenItFinished. 마커는 복구보다 먼저 생긴다
// (engine.go 6단계 → 7단계). 신호가 복구의 성공에 묶이지 않으면 그것은 "엔진이 떴다"의
// 다른 이름일 뿐이고, spec이 금지한 바로 그 신호다.
func TestRecoveryPublishesReadyOnlyWhenItFinished(t *testing.T) {
	ready := 0
	recover := recoverThenReady(
		func(context.Context) (reconcile.Report, error) { return reconcile.Report{}, nil },
		func() { ready++ },
		nil,
	)
	if err := recover(context.Background()); err != nil {
		t.Fatalf("a successful recovery returned %v", err)
	}
	if ready != 1 {
		t.Fatalf("ready called %d times, want exactly 1", ready)
	}
}

// TestAFailedRecoveryNeverPublishesReady is design 검증계약 (e).
func TestAFailedRecoveryNeverPublishesReady(t *testing.T) {
	failure := errors.New("복구가 rate limit 예산을 다 썼다")
	ready := 0
	recover := recoverThenReady(
		func(context.Context) (reconcile.Report, error) { return reconcile.Report{}, failure },
		func() { ready++ },
		nil,
	)
	err := recover(context.Background())
	if !errors.Is(err, failure) {
		t.Fatalf("err = %v, want the recovery's own failure", err)
	}
	if ready != 0 {
		t.Fatalf("ready called %d times after a failed recovery, want 0 — "+
			"the survey would start against an engine with no protection", ready)
	}
}

// TestACancelledRecoveryNeverPublishesReady. 정지 신호가 온 프로세스는 곧 내려간다.
// "준비됐다"를 남기고 죽으면 다음 콘솔이 그 마커를 stale로 인정할 때까지 5분 동안 거짓말이다.
func TestACancelledRecoveryNeverPublishesReady(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ready := 0
	recover := recoverThenReady(
		func(context.Context) (reconcile.Report, error) {
			cancel()
			return reconcile.Report{}, nil
		},
		func() { ready++ },
		nil,
	)
	err := recover(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
	if ready != 0 {
		t.Fatalf("ready called %d times on a cancelled recovery, want 0", ready)
	}
}

// --- D5b (A1 F1): the wait is not silent ----------------------------------------

// TestTheRateLimitedWaitIsReportedOnBothPaths. §1은 최대 5분을 기다릴 수 있게 만들었고,
// engine.go:402는 그 Report를 버리고 있었다. 성공했든 실패했든 운영자는 "보호가 없던
// 시간"을 사후에 셀 수 있어야 한다.
func TestTheRateLimitedWaitIsReportedOnBothPaths(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
	}{
		{name: "성공", err: nil},
		{name: "실패", err: errors.New("recovery incomplete")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			seen := []reconcile.Report{}
			recover := recoverThenReady(
				func(context.Context) (reconcile.Report, error) {
					return reconcile.Report{RateLimitWaits: 3, RateLimitWaited: 45 * time.Second}, tc.err
				},
				func() {},
				func(r reconcile.Report) { seen = append(seen, r) },
			)
			_ = recover(context.Background())
			if len(seen) != 1 {
				t.Fatalf("the report was observed %d times, want exactly 1", len(seen))
			}
			if seen[0].RateLimitWaits != 3 || seen[0].RateLimitWaited != 45*time.Second {
				t.Fatalf("observed report = %d/%s, want 3/45s",
					seen[0].RateLimitWaits, seen[0].RateLimitWaited)
			}
		})
	}
}

// TestARateLimitedRecoveryLeavesOneCountableLine. 문장이 아니라 **필드**여야 한다
// (obs/log.go: "셀 수 있는 이벤트 규약"). 몇 번과 총 얼마나가 각각 필드로 나온다.
func TestARateLimitedRecoveryLeavesOneCountableLine(t *testing.T) {
	buffer := &strings.Builder{}
	logger := obs.NewLogger(obs.LogOptions{Writer: buffer, JSON: true})

	engineRecoveryObserver(logger)(reconcile.Report{
		RateLimitWaits: 4, RateLimitWaited: 60 * time.Second,
	})

	lines := nonEmptyLines(buffer.String())
	if len(lines) != 1 {
		t.Fatalf("wrote %d lines, want exactly 1:\n%s", len(lines), buffer.String())
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &got); err != nil {
		t.Fatalf("the line is not JSON: %v\n%s", err, lines[0])
	}
	if got[obs.FieldEvent] != string(obs.EventRecoveryRateLimited) {
		t.Errorf("event = %v, want %s", got[obs.FieldEvent], obs.EventRecoveryRateLimited)
	}
	if got[obs.FieldCount] != float64(4) {
		t.Errorf("count = %v, want 4", got[obs.FieldCount])
	}
	if got[obs.FieldDurationMS] != float64(60000) {
		t.Errorf("duration_ms = %v, want 60000", got[obs.FieldDurationMS])
	}
}

// TestARecoveryThatWasNeverThrottledSaysNothing. 조용한 것이 기본이다 — 이 줄은 이상
// 상태의 표지이고, 정상 복구마다 나오면 셀 수 없게 된다.
func TestARecoveryThatWasNeverThrottledSaysNothing(t *testing.T) {
	buffer := &strings.Builder{}
	logger := obs.NewLogger(obs.LogOptions{Writer: buffer, JSON: true})
	engineRecoveryObserver(logger)(reconcile.Report{})
	if strings.TrimSpace(buffer.String()) != "" {
		t.Fatalf("a recovery with no rate-limit wait still logged:\n%s", buffer.String())
	}
}

// TestTheRecoveryObserverSurvivesANilLogger. `engineRuntime`은 테스트에서 nil logger로
// 불린다(engine_runtime_branch_test.go). 관측 한 줄이 조립을 깨면 안 된다.
func TestTheRecoveryObserverSurvivesANilLogger(t *testing.T) {
	engineRecoveryObserver(nil)(reconcile.Report{RateLimitWaits: 1, RateLimitWaited: time.Second})
}

// --- D6: the bounded wait --------------------------------------------------------

// TestAwaitReturnsAsSoonAsTheSignalIsThere — spec 시나리오 ①.
func TestAwaitReturnsAsSoonAsTheSignalIsThere(t *testing.T) {
	clk := newA102Clock()
	marks := &observations{script: []enginelock.Status{
		runningMarker(), runningMarker(), readyMarker(clk.now.Add(4 * time.Second)),
	}}

	verdict := awaitEngineReady(context.Background(), marks.observe, clk,
		engineReadyCap, engineReadyPoll)

	if verdict != engineReadyConfirmed {
		t.Fatalf("verdict = %v, want confirmed", verdict)
	}
	if marks.calls != 3 {
		t.Errorf("observed %d times, want 3", marks.calls)
	}
	if len(clk.waits) != 2 {
		t.Fatalf("waited %v, want exactly two polls", clk.waits)
	}
	for i, d := range clk.waits {
		if d != engineReadyPoll {
			t.Errorf("wait %d = %s, want the poll interval %s", i, d, engineReadyPoll)
		}
	}
}

// TestAwaitDoesNotWaitForAnEngineThatIsNotThere — spec 시나리오 ②. 오늘 02:03의 엔진은
// 1초 만에 죽었다. 죽은 엔진을 120초 기다리는 것은 서베이를 이유 없이 미루는 것이다.
func TestAwaitDoesNotWaitForAnEngineThatIsNotThere(t *testing.T) {
	clk := newA102Clock()
	marks := &observations{script: []enginelock.Status{{}}}

	verdict := awaitEngineReady(context.Background(), marks.observe, clk,
		engineReadyCap, engineReadyPoll)

	if verdict != engineReadyNoEngine {
		t.Fatalf("verdict = %v, want no-engine", verdict)
	}
	if marks.calls != 1 || len(clk.waits) != 0 {
		t.Fatalf("observed %d times and waited %v, want one look and no wait", marks.calls, clk.waits)
	}
}

// TestAnEngineThatDiesMidWaitStopsTheWait. 마커가 stale로 넘어가면 Running이 꺼진다 —
// 그 순간부터는 기다릴 대상이 없다.
func TestAnEngineThatDiesMidWaitStopsTheWait(t *testing.T) {
	clk := newA102Clock()
	marks := &observations{script: []enginelock.Status{
		runningMarker(), runningMarker(), {},
	}}

	verdict := awaitEngineReady(context.Background(), marks.observe, clk,
		engineReadyCap, engineReadyPoll)

	if verdict != engineReadyNoEngine {
		t.Fatalf("verdict = %v, want no-engine", verdict)
	}
	if len(clk.waits) != 2 {
		t.Errorf("waited %v, want two polls before the engine disappeared", clk.waits)
	}
}

// TestTheWaitStopsAtTheCap is design 검증계약 (f).
//
// 상한이 없으면 죽지 않은 채 준비도 못 하는 엔진 옆에서 서베이가 영원히 서지 않고,
// attestation 시계가 조용히 멈춘다 — a101이 고친 바로 그 결말이다.
func TestTheWaitStopsAtTheCap(t *testing.T) {
	clk := newA102Clock()
	marks := &observations{script: []enginelock.Status{runningMarker()}}

	verdict := awaitEngineReady(context.Background(), marks.observe, clk,
		engineReadyCap, engineReadyPoll)

	if verdict != engineReadyCapExceeded {
		t.Fatalf("verdict = %v after %d waits, want cap-exceeded (%q)",
			verdict, len(clk.waits), engineReadyNote(engineReadyCapExceeded))
	}
	wantPolls := int(engineReadyCap / engineReadyPoll)
	if len(clk.waits) != wantPolls {
		t.Fatalf("waited %d times, want %d (%s ÷ %s)",
			len(clk.waits), wantPolls, engineReadyCap, engineReadyPoll)
	}
	var total time.Duration
	for _, d := range clk.waits {
		total += d
	}
	if total != engineReadyCap {
		t.Errorf("total wait = %s, want exactly the cap %s", total, engineReadyCap)
	}
	if marks.calls != wantPolls+1 {
		t.Errorf("observed %d times, want %d", marks.calls, wantPolls+1)
	}
}

// TestTheWaitIsAbandonedWhenTheConsoleCloses. ctx는 콘솔의 것이다. 콘솔이 내려가는데
// 서베이를 새로 띄우면 그 서베이는 아무도 보지 않는 자식이 된다.
func TestTheWaitIsAbandonedWhenTheConsoleCloses(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	clk := newA102Clock()
	clk.cancelAt, clk.cancel = 3, cancel
	marks := &observations{script: []enginelock.Status{runningMarker()}}

	verdict := awaitEngineReady(ctx, marks.observe, clk, engineReadyCap, engineReadyPoll)

	if verdict != engineReadyAbandoned {
		t.Fatalf("verdict = %v, want abandoned", verdict)
	}
	if len(clk.waits) != 3 {
		t.Errorf("waited %v, want the wait to stop at the cancellation", clk.waits)
	}
}

// TestAnAlreadyClosedConsoleStartsNothing. 취소된 ctx로 들어오면 관측조차 하지 않는다.
func TestAnAlreadyClosedConsoleStartsNothing(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	clk := newA102Clock()
	marks := &observations{script: []enginelock.Status{readyMarker(clk.now)}}

	if verdict := awaitEngineReady(ctx, marks.observe, clk,
		engineReadyCap, engineReadyPoll); verdict != engineReadyAbandoned {
		t.Fatalf("verdict = %v, want abandoned", verdict)
	}
	if marks.calls != 0 {
		t.Errorf("observed %d times on a closed console, want 0", marks.calls)
	}
}

// TestWithoutAMarkerReaderThereIsNothingToWaitFor. 콘솔이 엔진 디렉터리를 못 풀면
// `engineMarkerPath`가 빈 문자열이다(console.go:282 else). 그 경우는 "엔진 없음"이다.
func TestWithoutAMarkerReaderThereIsNothingToWaitFor(t *testing.T) {
	clk := newA102Clock()
	if verdict := awaitEngineReady(context.Background(), nil, clk,
		engineReadyCap, engineReadyPoll); verdict != engineReadyNoEngine {
		t.Fatalf("verdict = %v, want no-engine", verdict)
	}
}

// TestTheCapAndPollAreTheMeasuredNumbers. 120s는 실측 복구 ~50s
// (2026-08-13 01:35:02→01:35:52)에 429 여유를 더한 값이고, 겹1의 5분 예산보다 **짧아야**
// 한다 — 상한은 조용한 대기의 끝일 뿐, 그 뒤의 경합은 겹1이 생존시킨다.
func TestTheCapAndPollAreTheMeasuredNumbers(t *testing.T) {
	if engineReadyCap != 2*time.Minute {
		t.Errorf("cap = %s, want 120s", engineReadyCap)
	}
	if engineReadyPoll != 2*time.Second {
		t.Errorf("poll = %s, want 2s", engineReadyPoll)
	}
	if engineReadyCap >= reconcile.DefaultMaxRateLimitWait {
		t.Errorf("cap %s is not shorter than the recovery's rate-limit budget %s",
			engineReadyCap, reconcile.DefaultMaxRateLimitWait)
	}
	if engineReadyPoll >= enginelock.RefreshEvery {
		t.Errorf("poll %s is not finer than the marker refresh %s",
			engineReadyPoll, enginelock.RefreshEvery)
	}
}

// --- D7: the note says which one it was -----------------------------------------

// TestEveryVerdictSaysWhichOneItWas — spec: "어느 경우였는지를 기동 노트가 말해야 한다.
// 조용한 상한 초과는 금지다."
func TestEveryVerdictSaysWhichOneItWas(t *testing.T) {
	seen := map[string]bool{}
	for _, verdict := range []engineReadyVerdict{
		engineReadyConfirmed, engineReadyNoEngine, engineReadyCapExceeded, engineReadyAbandoned,
	} {
		note := engineReadyNote(verdict)
		if strings.TrimSpace(note) == "" {
			t.Fatalf("verdict %v produced no note", verdict)
		}
		if seen[note] {
			t.Fatalf("verdict %v repeats another verdict's note %q", verdict, note)
		}
		seen[note] = true
	}
	if !strings.Contains(engineReadyNote(engineReadyCapExceeded), engineReadyCap.String()) {
		t.Errorf("the cap-exceeded note does not say how long it waited: %q",
			engineReadyNote(engineReadyCapExceeded))
	}
	// 새 verdict가 생겼는데 문장을 안 붙이면 그것이 조용한 대기가 된다.
	if strings.TrimSpace(engineReadyNote(engineReadyVerdict(99))) == "" {
		t.Error("an unknown verdict says nothing at all")
	}
}

// TestASilentStartStillLeavesTheVerdict. `bootSurvey`가 `restartSoak`의 빈 note를
// 그대로 올릴 수 있다(a101 `TestConfiguredSoakAutostartSaysSomethingWhenTheSeamIsSilent`).
// 그때도 어느 쪽으로 끝난 대기였는지는 남아야 한다 — 조용한 상한 초과 금지는 start의
// 말수와 무관하다.
func TestASilentStartStillLeavesTheVerdict(t *testing.T) {
	clk := newA102Clock()
	marks := &observations{script: []enginelock.Status{runningMarker()}}
	note, err := soakStartAfterEngineReady(context.Background(), marks.observe, clk,
		func() (string, error) { return "   ", nil })()
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	if note != engineReadyNote(engineReadyCapExceeded) {
		t.Fatalf("note = %q, want the bare verdict %q", note, engineReadyNote(engineReadyCapExceeded))
	}
}

// TestAWaitWithNothingToStartSaysSoRatherThanPanicking. goroutine 안의 panic은
// 콘솔 프로세스를 통째로 죽인다 — 서베이의 사정이 운영자 화면을 없애는 유일한 방법이다.
func TestAWaitWithNothingToStartSaysSoRatherThanPanicking(t *testing.T) {
	clk := newA102Clock()
	marks := &observations{script: []enginelock.Status{{}}}
	_, err := soakStartAfterEngineReady(context.Background(), marks.observe, clk, nil)()
	if !errors.Is(err, errSoakBootNoStartWiring) {
		t.Fatalf("err = %v, want errSoakBootNoStartWiring", err)
	}
}

// TestTheSurveyStartsAfterTheSignalAndSaysSo — spec 시나리오 ①의 배선.
func TestTheSurveyStartsAfterTheSignalAndSaysSo(t *testing.T) {
	clk := newA102Clock()
	marks := &observations{script: []enginelock.Status{
		runningMarker(), readyMarker(clk.now.Add(2 * time.Second)),
	}}
	started := 0
	start := soakStartAfterEngineReady(context.Background(), marks.observe, clk,
		func() (string, error) { started++; return "새로 시작했다", nil })

	note, err := start()
	if err != nil || started != 1 {
		t.Fatalf("started=%d err=%v, want exactly one start", started, err)
	}
	for _, want := range []string{engineReadyNote(engineReadyConfirmed), "새로 시작했다"} {
		if !strings.Contains(note, want) {
			t.Errorf("note=%q, missing %q", note, want)
		}
	}
}

// TestTheSurveyStartsAnywayAfterTheCapAndSaysSo — spec 시나리오 ③.
func TestTheSurveyStartsAnywayAfterTheCapAndSaysSo(t *testing.T) {
	clk := newA102Clock()
	marks := &observations{script: []enginelock.Status{runningMarker()}}
	started := 0
	start := soakStartAfterEngineReady(context.Background(), marks.observe, clk,
		func() (string, error) { started++; return "새로 시작했다", nil })

	note, err := start()
	if err != nil {
		t.Fatalf("err=%v, want nil — the survey is optional machinery and must still start", err)
	}
	if started != 1 {
		t.Fatalf("started=%d, want 1 — the attestation clock has to keep running", started)
	}
	if !strings.Contains(note, engineReadyNote(engineReadyCapExceeded)) {
		t.Errorf("note=%q, want the cap named — a silent cap is forbidden", note)
	}
}

// TestNoEngineMeansTodaysBehaviour — spec 시나리오 ②.
func TestNoEngineMeansTodaysBehaviour(t *testing.T) {
	clk := newA102Clock()
	marks := &observations{script: []enginelock.Status{{}}}
	started := 0
	start := soakStartAfterEngineReady(context.Background(), marks.observe, clk,
		func() (string, error) { started++; return "새로 시작했다", nil })

	note, err := start()
	if err != nil || started != 1 {
		t.Fatalf("started=%d err=%v, want an immediate start", started, err)
	}
	if len(clk.waits) != 0 {
		t.Errorf("waited %v before starting, want none", clk.waits)
	}
	if !strings.Contains(note, engineReadyNote(engineReadyNoEngine)) {
		t.Errorf("note=%q, want the reason named", note)
	}
}

// TestAClosingConsoleStartsNoSurvey. runConfiguredSoakAutostart는 이 오류를
// "soak 자동 시작 실패: …" 한 줄로 강등한다 — 콘솔이 내려가는 중이므로 맞는 문장이다.
func TestAClosingConsoleStartsNoSurvey(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	clk := newA102Clock()
	marks := &observations{script: []enginelock.Status{runningMarker()}}
	started := 0
	start := soakStartAfterEngineReady(ctx, marks.observe, clk,
		func() (string, error) { started++; return "새로 시작했다", nil })

	_, err := start()
	if !errors.Is(err, errSoakBootConsoleClosed) {
		t.Fatalf("err=%v, want errSoakBootConsoleClosed", err)
	}
	if started != 0 {
		t.Fatalf("started=%d, want 0 — nobody is left to watch the survey", started)
	}
	// 그리고 a101의 불변식대로, 그 실패는 문자열 한 줄이 된다.
	note := runConfiguredSoakAutostart(func() (bool, error) { return true, nil }, start)
	if !strings.Contains(note, "soak 자동 시작 실패") {
		t.Errorf("note=%q, want the failure reported as one line", note)
	}
}

// TestAFailedStartKeepsItsOwnWords. 대기가 붙어도 a101의 계약은 그대로다 — 실패한 기동의
// 자기 설명이 노트에 남는다.
func TestAFailedStartKeepsItsOwnWords(t *testing.T) {
	clk := newA102Clock()
	marks := &observations{script: []enginelock.Status{{}}}
	start := soakStartAfterEngineReady(context.Background(), marks.observe, clk,
		func() (string, error) {
			return "pid 42이(가) 30s 안에 종료되지 않았다", errors.New("stop timeout")
		})

	note, err := start()
	if err == nil {
		t.Fatal("a failed start came back as success")
	}
	if !strings.Contains(note, "종료되지 않았다") {
		t.Errorf("note=%q, want the seam's own diagnosis kept", note)
	}
}

// --- D7: the wiring runConsole cannot prove about itself -------------------------

// TestTheSoakAutostartWaitsOffTheConsolePath is a source-shape regression, and it
// is that on purpose: `runConsole` is measured at 0.0% (this change's
// cmd-tossctl--runconsole/branch-test-map.md), so the three things below cannot be
// observed by running it. `TestTheRestartRecoveryRunsBeforeTheLoops`
// (engine_test.go) already uses this technique for the same reason.
//
// spec 시나리오 ④ "대기가 콘솔을 막지 않는다"가 여기 걸린다.
func TestTheSoakAutostartWaitsOffTheConsolePath(t *testing.T) {
	source := readSource(t, "console.go")

	waitAt := strings.Index(source, "soakStartAfterEngineReady(")
	if waitAt < 0 {
		t.Fatal("runConsole no longer wraps the boot start seam with the engine-ready wait")
	}
	goAt := strings.LastIndex(source[:waitAt], "go func() {")
	if goAt < 0 {
		t.Fatal("the wait is not inside a goroutine; a 120s wait in front of the console is spec-forbidden")
	}
	serveAt := strings.Index(source, "console.ListenAndServe(ctx, console.Options{")
	if serveAt < 0 || serveAt < waitAt {
		t.Fatal("the operator screen is no longer assembled after the autostart block")
	}
	// 그리고 그 goroutine이 콘솔의 ctx를 받는다 — 콘솔 종료가 대기를 끊는 유일한 경로다.
	block := source[goAt:serveAt]
	if !strings.Contains(block, "soakStartAfterEngineReady(ctx,") {
		t.Error("the wait does not take the console's context, so closing the console cannot stop it")
	}
	// 버튼은 감싸지 않는다: 운영자가 방금 누른 것을 120초 기다리게 하면 고장과 구별되지 않는다.
	buttonAt := strings.Index(source, "RestartSoak: func() (string, error) {")
	if buttonAt < 0 {
		t.Fatal("the restart button seam moved; re-audit whether it is now behind the wait")
	}
	if strings.Contains(source[buttonAt:], "soakStartAfterEngineReady(") {
		t.Error("the restart button is behind the engine-ready wait; a pressed button must not wait")
	}
}

func nonEmptyLines(value string) []string {
	out := []string{}
	for _, line := range strings.Split(value, "\n") {
		if strings.TrimSpace(line) != "" {
			out = append(out, line)
		}
	}
	return out
}

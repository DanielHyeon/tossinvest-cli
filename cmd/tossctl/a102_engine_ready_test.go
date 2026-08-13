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
	"os"
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

// observations replays a scripted sequence of looks, repeating the last one
// forever. It counts how many times it was asked.
type observations struct {
	script []engineLook
	calls  int
}

func (o *observations) observe() engineLook {
	o.calls++
	if len(o.script) == 0 {
		return engineLookAbsent
	}
	if o.calls-1 < len(o.script) {
		return o.script[o.calls-1]
	}
	return o.script[len(o.script)-1]
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
	marks := &observations{script: []engineLook{
		engineLookNotYet, engineLookNotYet, engineLookReady,
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
	marks := &observations{script: []engineLook{engineLookAbsent}}

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
	marks := &observations{script: []engineLook{
		engineLookNotYet, engineLookNotYet, engineLookAbsent,
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
	marks := &observations{script: []engineLook{engineLookNotYet}}

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
	marks := &observations{script: []engineLook{engineLookNotYet}}

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
	marks := &observations{script: []engineLook{engineLookReady}}

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
	marks := &observations{script: []engineLook{engineLookNotYet}}
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
	marks := &observations{script: []engineLook{engineLookAbsent}}
	_, err := soakStartAfterEngineReady(context.Background(), marks.observe, clk, nil)()
	if !errors.Is(err, errSoakBootNoStartWiring) {
		t.Fatalf("err = %v, want errSoakBootNoStartWiring", err)
	}
}

// TestTheSurveyStartsAfterTheSignalAndSaysSo — spec 시나리오 ①의 배선.
func TestTheSurveyStartsAfterTheSignalAndSaysSo(t *testing.T) {
	clk := newA102Clock()
	marks := &observations{script: []engineLook{
		engineLookNotYet, engineLookReady,
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
	marks := &observations{script: []engineLook{engineLookNotYet}}
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
	marks := &observations{script: []engineLook{engineLookAbsent}}
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

// TestAClosingConsoleStartsNoSurveyAndSaysSoWithoutCryingFailure is A2 F8.
//
// 콘솔이 내려가면서 서베이를 시작하지 않은 것은 **정상 종료**다. 첫 판은 그것을
// sentinel 오류로 돌려줬고, a101의 계약대로 "soak 자동 시작 실패: …"로 찍혔다 —
// 운영자가 쫓아가야 할 고장처럼 읽힌다. 노트가 스스로 "시작하지 않았다"를 말하므로
// 오류일 이유가 없다.
func TestAClosingConsoleStartsNoSurveyAndSaysSoWithoutCryingFailure(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	clk := newA102Clock()
	marks := &observations{script: []engineLook{engineLookNotYet}}
	started := 0
	start := soakStartAfterEngineReady(ctx, marks.observe, clk,
		func() (string, error) { started++; return "새로 시작했다", nil })

	note, err := start()
	if err != nil {
		t.Fatalf("err=%v, want nil — a console shutting down is not a fault", err)
	}
	if started != 0 {
		t.Fatalf("started=%d, want 0 — nobody is left to watch the survey", started)
	}
	if !strings.Contains(note, "시작하지 않았다") {
		t.Fatalf("note=%q, want it to say nothing was started", note)
	}
	line := runConfiguredSoakAutostart(func() (bool, error) { return true, nil }, start)
	if strings.Contains(line, "실패") {
		t.Errorf("line=%q, want a normal shutdown not reported as a failure", line)
	}
	if !strings.Contains(line, "시작하지 않았다") {
		t.Errorf("line=%q, want the operator told nothing was started", line)
	}
}

// TestAFailedStartKeepsItsOwnWords. 대기가 붙어도 a101의 계약은 그대로다 — 실패한 기동의
// 자기 설명이 노트에 남는다.
func TestAFailedStartKeepsItsOwnWords(t *testing.T) {
	clk := newA102Clock()
	marks := &observations{script: []engineLook{engineLookAbsent}}
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

// --- D4b: the signal has to belong to a process that is still alive --------------

// TestACorpsesReadyMarkerIsNotAReadyEngine is A2 F1, the finding that made this
// round FIX-FIRST.
//
// 크래시는 `Release`를 부르지 않는다. ready_at이 든 마커는 StaleAfter(5m) 동안 신선한
// 채로 남고, 신선도만 보는 판정은 그것을 "계좌가 보호되고 있다"로 읽는다. 02:03 사고의
// 바로 다음 장면이 그것이다 — 콘솔이 대체 엔진을 띄우고, 그 엔진이 아직 조립 중인 동안
// 전임자의 ready_at이 '준비 확인'으로 읽혀 서베이가 새 엔진의 복구 예산을 때린다.
func TestACorpsesReadyMarkerIsNotAReadyEngine(t *testing.T) {
	stamp := time.Date(2026, 8, 13, 2, 3, 29, 0, time.UTC)
	corpse := enginelock.Status{
		Running: true,
		Marker:  enginelock.Marker{PID: 4242, ReadyAt: &stamp},
	}

	// 대체 엔진(pid 5150)이 살아 있고, 디스크의 ready_at은 죽은 4242의 것이다.
	if got := readEngineSignal(corpse, []int{5150}, nil); got != engineLookNotYet {
		t.Fatalf("look = %v, want not-yet — a dead engine's ready_at was read as protection", got)
	}
	// 같은 마커라도 그 pid가 살아 있으면 준비다.
	if got := readEngineSignal(corpse, []int{4242}, nil); got != engineLookReady {
		t.Fatalf("look = %v, want ready — the live engine's own signal was ignored", got)
	}
	// 살아 있는 엔진이 하나도 없으면 유령 마커는 기다릴 이유가 아니다 (오늘의 동작).
	if got := readEngineSignal(corpse, nil, nil); got != engineLookAbsent {
		t.Fatalf("look = %v, want absent — a ghost marker made the survey wait", got)
	}
}

// TestAnEngineWithNoSignalYetIsWorthWaitingFor. 부팅 창 그 자체다 — 프로세스는 떴고
// 마커는 아직 없거나 ready_at이 없다.
func TestAnEngineWithNoSignalYetIsWorthWaitingFor(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status enginelock.Status
	}{
		{"마커 없음", enginelock.Status{}},
		{"마커는 있으나 신호 없음", enginelock.Status{Running: true, Marker: enginelock.Marker{PID: 5150}}},
	} {
		if got := readEngineSignal(tc.status, []int{5150}, nil); got != engineLookNotYet {
			t.Errorf("%s: look = %v, want not-yet", tc.name, got)
		}
	}
}

// TestAFailedEnumerationIsNeitherAbsenceNorReadiness. `bootSurvey`가 이미 그은 선과
// 같다 — 못 셌다는 없다가 아니다. 그리고 시체가 아니라는 증거도 아니다. 상한까지
// 기다린 뒤 시작하므로 대가는 유한하다.
func TestAFailedEnumerationIsNeitherAbsenceNorReadiness(t *testing.T) {
	stamp := time.Date(2026, 8, 13, 2, 3, 29, 0, time.UTC)
	live := enginelock.Status{Running: true, Marker: enginelock.Marker{PID: 4242, ReadyAt: &stamp}}
	broken := errors.New("pgrep unavailable")

	if got := readEngineSignal(live, []int{4242}, broken); got != engineLookNotYet {
		t.Errorf("look = %v, want not-yet when the process list could not be read", got)
	}
	if got := readEngineSignal(enginelock.Status{}, nil, broken); got != engineLookNotYet {
		t.Errorf("look = %v, want not-yet; an unreadable process table is not an absence", got)
	}
}

// TestTheConsoleReadinessSeamCombinesTheMarkerAndTheProcessList is A2's surviving
// mutation N2 — the composition itself, which used to live in runConsole where
// nothing could reach it.
func TestTheConsoleReadinessSeamCombinesTheMarkerAndTheProcessList(t *testing.T) {
	dir := t.TempDir()
	markerPath := enginelock.MarkerPath(dir)
	at := time.Date(2026, 8, 13, 2, 3, 29, 0, time.UTC)

	held, err := enginelock.Hold(context.Background(), markerPath, at)
	if err != nil {
		t.Fatalf("Hold: %v", err)
	}
	t.Cleanup(held.Release)

	var pids []int
	var findErr error
	previous := engineFindProcesses
	engineFindProcesses = func(string) ([]int, error) { return pids, findErr }
	t.Cleanup(func() { engineFindProcesses = previous })

	observe := consoleEngineReadiness(dir, markerPath, func() time.Time { return at.Add(time.Second) })

	pids = []int{os.Getpid()}
	if got := observe(); got != engineLookNotYet {
		t.Fatalf("look = %v, want not-yet before the engine says it is ready", got)
	}
	held.Ready(at.Add(50 * time.Second))
	if got := observe(); got != engineLookReady {
		t.Fatalf("look = %v, want ready once the live engine published the signal", got)
	}
	pids = []int{os.Getpid() + 1}
	if got := observe(); got != engineLookNotYet {
		t.Fatalf("look = %v, want not-yet — the marker belongs to a pid that is gone", got)
	}
	pids = nil
	if got := observe(); got != engineLookAbsent {
		t.Fatalf("look = %v, want absent", got)
	}
	pids, findErr = nil, errors.New("pgrep unavailable")
	if got := observe(); got != engineLookNotYet {
		t.Fatalf("look = %v, want not-yet on a failed enumeration", got)
	}
}

// TestWithoutAnEngineDirectoryThereIsNothingToObserve. `engineJournalDir`가 실패하면
// 콘솔의 마커 경로는 빈 문자열이다(console.go의 else 갈래).
func TestWithoutAnEngineDirectoryThereIsNothingToObserve(t *testing.T) {
	if got := consoleEngineReadiness("", "", nil)(); got != engineLookAbsent {
		t.Fatalf("look = %v, want absent", got)
	}
	if got := consoleEngineReadiness("/tmp", "  ", nil)(); got != engineLookAbsent {
		t.Fatalf("look = %v, want absent", got)
	}
}

// --- D5c: the wait's own numbers are the module's ---------------------------------

// TestTheWrapperSpendsTheCapAtThePollInterval is A2's surviving mutation N3.
//
// `awaitEngineReady`의 단위 테스트는 상한과 간격을 **인자로** 받으므로, 감싸는 쪽이
// 그 둘을 맞바꿔 넘겨도 죽지 않는다 — verdict는 여전히 cap 초과이기 때문이다.
// 죽이려면 감싸는 경로에서 **몇 번을 얼마씩 기다렸는지**를 재야 한다.
func TestTheWrapperSpendsTheCapAtThePollInterval(t *testing.T) {
	clk := newA102Clock()
	marks := &observations{script: []engineLook{engineLookNotYet}}
	start := soakStartAfterEngineReady(context.Background(), marks.observe, clk,
		func() (string, error) { return "새로 시작했다", nil })

	if _, err := start(); err != nil {
		t.Fatalf("err = %v", err)
	}
	wantPolls := int(engineReadyCap / engineReadyPoll)
	if len(clk.waits) != wantPolls {
		t.Fatalf("waited %d times, want %d — the cap and the poll interval were swapped",
			len(clk.waits), wantPolls)
	}
	var total time.Duration
	for i, d := range clk.waits {
		if d != engineReadyPoll {
			t.Fatalf("wait %d = %s, want the poll interval %s", i, d, engineReadyPoll)
		}
		total += d
	}
	if total != engineReadyCap {
		t.Errorf("total wait = %s, want the cap %s", total, engineReadyCap)
	}
}

// --- D7b: one console, one spawner at a time --------------------------------------

// TestTheBootPathAndTheButtonCannotSpawnAtTheSameTime is A2 F5.
//
// 편집 전 자동 시작 블록은 리스너보다 앞에서 동기로 끝났으므로 버튼과 겹칠 수 없었다.
// goroutine화가 최대 120초의 동시 창을 새로 만들었고, 두 경로 모두 같은 record 위의
// check-then-act다(열거 → spawn). 통과만으로는 증거가 아니므로, **첫 경로를 붙잡아 둔 채**
// 두 번째가 들어오지 못하는 것을 관측한다.
func TestTheBootPathAndTheButtonCannotSpawnAtTheSameTime(t *testing.T) {
	inside := make(chan struct{})
	release := make(chan struct{})
	entered := make(chan string, 2)

	go func() {
		_, _ = spawnOneSurvey(func() (string, error) {
			entered <- "boot"
			close(inside)
			<-release
			return "boot started", nil
		})
	}()
	<-inside

	buttonDone := make(chan struct{})
	go func() {
		defer close(buttonDone)
		_, _ = guardedSoakRestart(func() (string, error) {
			entered <- "button"
			return "button started", nil
		}, nil)
	}()

	select {
	case who := <-entered:
		if who != "boot" {
			t.Fatalf("first entrant was %q, want boot", who)
		}
	default:
		t.Fatal("the boot path never reported entering")
	}
	select {
	case who := <-entered:
		t.Fatalf("%q entered while the boot path was still inside the gate", who)
	case <-time.After(50 * time.Millisecond):
	}

	close(release)
	select {
	case <-buttonDone:
	case <-time.After(2 * time.Second):
		t.Fatal("the button never got in after the boot path left the gate")
	}
	if who := <-entered; who != "button" {
		t.Fatalf("second entrant was %q, want button", who)
	}
}

// TestTheButtonStillRecordsTheApproval. 게이트를 끼우면서 a101의 계약을 잃으면 안 된다.
func TestTheButtonStillRecordsTheApproval(t *testing.T) {
	saved := []bool{}
	note, err := guardedSoakRestart(
		func() (string, error) { return "새로 시작했다", nil },
		func(on bool) error { saved = append(saved, on); return nil },
	)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if len(saved) != 1 || !saved[0] {
		t.Fatalf("saved = %v, want exactly one true", saved)
	}
	if !strings.Contains(note, "새로 시작했다") {
		t.Errorf("note = %q, want the restart's own words", note)
	}
}

// --- D7c: the console never joins the wait ----------------------------------------

// TestTheBootSurveyNeverBlocksTheConsole is A2's surviving mutation N4, and it is
// the reason this assertion is an execution and not a source shape.
//
// A2 wrote a version that kept `go func() {`, kept the ctx argument, kept the call
// order relative to ListenAndServe — every string the shape test looked for — and
// joined the goroutine before returning. The console blocked for the full cap and
// all five checks passed. 영원히 돌아오지 않는 start를 주고 **바깥 호출이 돌아오는지**를
// 잰다.
//
// spec 시나리오 ④ "대기가 콘솔을 막지 않는다"가 여기 걸린다.
func TestTheBootSurveyNeverBlocksTheConsole(t *testing.T) {
	dir := t.TempDir()
	previous := engineFindProcesses
	engineFindProcesses = func(string) ([]int, error) { return nil, nil } // 살아 있는 엔진 없음
	t.Cleanup(func() { engineFindProcesses = previous })

	forever := make(chan struct{})
	t.Cleanup(func() { close(forever) })
	entered := make(chan struct{})

	returned := make(chan struct{})
	go func() {
		defer close(returned)
		startSoakAutostartAsync(
			context.Background(), dir, enginelock.MarkerPath(dir),
			func() (bool, error) { return true, nil },
			func() (string, error) { close(entered); <-forever; return "", nil },
			func(string) {},
		)
	}()

	select {
	case <-returned:
	case <-time.After(2 * time.Second):
		t.Fatal("startSoakAutostartAsync joined the survey; the operator screen waits behind it")
	}
	select {
	case <-entered:
	case <-time.After(2 * time.Second):
		t.Fatal("the survey was never started at all")
	}
}

// TestTheBootSurveyReportsItsNote. 비동기가 되었다고 노트가 사라지면 조용한 상한 초과가
// 되돌아온다.
func TestTheBootSurveyReportsItsNote(t *testing.T) {
	dir := t.TempDir()
	previous := engineFindProcesses
	engineFindProcesses = func(string) ([]int, error) { return nil, nil }
	t.Cleanup(func() { engineFindProcesses = previous })

	notes := make(chan string, 1)
	startSoakAutostartAsync(
		context.Background(), dir, enginelock.MarkerPath(dir),
		func() (bool, error) { return true, nil },
		func() (string, error) { return "새로 시작했다", nil },
		func(note string) { notes <- note },
	)

	select {
	case note := <-notes:
		for _, want := range []string{engineReadyNote(engineReadyNoEngine), "새로 시작했다"} {
			if !strings.Contains(note, want) {
				t.Errorf("note = %q, missing %q", note, want)
			}
		}
	case <-time.After(2 * time.Second):
		t.Fatal("the boot survey never reported anything")
	}
}

// TestAnAutostartThatIsOffStartsNothingAndSaysNothing. a101의 계약이 비동기 뒤에서도
// 유지되는지 — OFF면 start도 report도 없다.
func TestAnAutostartThatIsOffStartsNothingAndSaysNothing(t *testing.T) {
	dir := t.TempDir()
	previous := engineFindProcesses
	engineFindProcesses = func(string) ([]int, error) { return nil, nil }
	t.Cleanup(func() { engineFindProcesses = previous })

	starts := make(chan struct{}, 1)
	notes := make(chan string, 1)
	startSoakAutostartAsync(
		context.Background(), dir, enginelock.MarkerPath(dir),
		func() (bool, error) { return false, nil },
		func() (string, error) { starts <- struct{}{}; return "", nil },
		func(note string) { notes <- note },
	)

	select {
	case <-starts:
		t.Fatal("a survey was started with autostart off")
	case note := <-notes:
		t.Fatalf("note = %q, want silence when autostart is off", note)
	case <-time.After(200 * time.Millisecond):
	}
}

// --- what runConsole still cannot prove about itself -------------------------------

// TestTheConsoleDelegatesTheWholeBootSurveyBlock is what is left of the source
// shape check after D7c: one call, not five strings.
//
// `runConsole` is measured at 0.0% (this change's
// cmd-tossctl--runconsole/branch-test-map.md), so *that it calls this at all*
// cannot be observed by running it. Everything the call does is asserted by
// execution above.
func TestTheConsoleDelegatesTheWholeBootSurveyBlock(t *testing.T) {
	source := readSource(t, "console.go")
	if !strings.Contains(source, "startSoakAutostartAsync(ctx, engineDir, engineMarkerPath,") {
		t.Fatal("runConsole no longer delegates the boot survey block to startSoakAutostartAsync")
	}
	if !strings.Contains(source, "guardedSoakRestart(startSurvey, save)") {
		t.Fatal("the restart button no longer goes through the console's spawn gate")
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

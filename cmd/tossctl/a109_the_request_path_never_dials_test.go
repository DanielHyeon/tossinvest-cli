//go:build unix

package main

// a109_the_request_path_never_dials_test.go — T2 §2.3 의 기전 핀 (design D4).
//
// 이웃 파일(`a109_the_daemon_reattaches_when_the_engine_returns_test.go`)이 재는 것은
// **결과**다: 엔진이 돌아오면 화면이 돌아온다. 결과만 재면 그 결과를 만드는 세 가지
// 방법 중 **틀린 두 개**도 통과한다.
//
//	① 요청 goroutine 에서 동기로 dial 한다 → 결과는 같지만 요청 경로가 200ms probe 를
//	   문다(`strategyprojectionrpc.Dial` 본문). freeze P0-2 가 SHALL NOT 으로 못박았다.
//	② rate limit 없이 매 요청 재해결한다 → 엔진이 내려간 동안 read 마다 probe 비용이다.
//
// 그래서 여기서는 wrapper 를 **직접** 세워 시도의 성질을 잰다: 언제 부르는가, 몇 번
// 부르는가, 부르는 동안 요청이 기다리는가. `resolve` 를 테스트가 쥐고 있으므로 그
// 셋을 전부 관측할 수 있다.

import (
	"context"
	"errors"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/httpapi"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojection"
)

// a109FakeReader 는 붙은 reader 하나다. err 를 채우면 「붙었는데 못 읽는다」가 된다.
type a109FakeReader struct {
	mu    sync.Mutex
	err   error
	reads int
}

func (r *a109FakeReader) Read(context.Context) (strategyprojection.Snapshot, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.reads++
	if r.err != nil {
		return strategyprojection.Snapshot{}, r.err
	}
	return strategyprojection.DormantSnapshot(time.Unix(0, 0).UTC()), nil
}

func (r *a109FakeReader) fail(err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.err = err
}

// a109SyncWriter 는 시도 goroutine 과 테스트 goroutine 이 함께 쓰는 로그다.
// strings.Builder 는 잠금이 없어 -race 가 그것을 그대로 잡는다.
type a109SyncWriter struct {
	mu   sync.Mutex
	sink strings.Builder
}

func (w *a109SyncWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.sink.Write(p)
}

func (w *a109SyncWriter) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.sink.String()
}

// a109Clock 은 테스트가 손으로 미는 시계다. rate limit 은 **시간의 함수**이므로
// 진짜 시계로 재면 느린 호스트에서 흔들린다.
type a109Clock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *a109Clock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *a109Clock) advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

// a109Attachment 는 wrapper 하나를 테스트가 쥔 resolve·시계와 함께 세운다.
func a109Attachment(t *testing.T, clock *a109Clock, log io.Writer,
	resolve func(context.Context) (httpapi.StrategyRuntimeReader, bool)) *strategyRuntimeAttachment {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	return &strategyRuntimeAttachment{
		resolve: resolve, ctx: ctx, now: clock.Now, interval: time.Minute, log: log,
	}
}

// a109WaitFor 는 백그라운드 시도가 끝나기를 기다린다.
func a109WaitFor(t *testing.T, what string, ok func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if ok() {
			return
		}
		time.Sleep(2 * time.Millisecond)
	}
	t.Fatalf("%s: 2s 안에 일어나지 않았다", what)
}

// TestTheRequestPathNeverWaitsForADial 은 freeze P0-2 의 직접 측정이다.
//
// resolve 를 놓아 줄 때까지 붙잡아 두고, 그 동안 요청이 **즉시** 지금 값으로 답하는지
// 본다. 동기 dial 구현에서는 이 창 안에 아무 응답도 나오지 않는다.
func TestTheRequestPathNeverWaitsForADial(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	clock := &a109Clock{now: time.Unix(1_760_000_000, 0).UTC()}
	attachment := a109Attachment(t, clock, io.Discard,
		func(context.Context) (httpapi.StrategyRuntimeReader, bool) {
			once.Do(func() { close(entered) })
			<-release
			return nil, false
		})
	defer close(release)
	// 부재로 출발한다 — 부재는 시도 대상이다.
	attachment.attach(nil, false)

	answered := make(chan bool, 1)
	go func() { answered <- attachment.StrategyRuntimeConfigured() }()
	select {
	case configured := <-answered:
		if configured {
			t.Fatal("부재 상태가 「설정돼 있다」로 답했다")
		}
	case <-time.After(time.Second):
		t.Fatal("요청 경로가 dial 을 기다렸다 — Dial 본문에는 200ms connect probe 가 있고, " +
			"엔진이 내려가 있으면 모든 화면 요청이 그 비용을 문다")
	}
	select {
	case <-entered:
	case <-time.After(2 * time.Second):
		t.Fatal("시도가 아예 일어나지 않았다 — 비차단으로 만들면서 재부착을 지웠다면 " +
			"그것은 고친 것이 아니라 없앤 것이다")
	}
}

// TestTheAttemptIsSingleFlightAndRateLimited 는 두 성질을 한 번에 잰다.
//
// 둘을 나눠 재면 각각을 통과시키는 반쪽 구현이 있다: single-flight 만 있으면 시도가
// 끝나자마자 다음 시도가 붙어 사실상 무제한이고, rate limit 만 있으면 창 안의 동시
// 요청 수만큼 dial 이 겹친다.
func TestTheAttemptIsSingleFlightAndRateLimited(t *testing.T) {
	entered, release := make(chan struct{}, 64), make(chan struct{})
	var calls int
	var mu sync.Mutex
	clock := &a109Clock{now: time.Unix(1_760_000_000, 0).UTC()}
	attachment := a109Attachment(t, clock, io.Discard,
		func(context.Context) (httpapi.StrategyRuntimeReader, bool) {
			mu.Lock()
			calls++
			mu.Unlock()
			entered <- struct{}{}
			<-release
			return nil, false
		})
	attachment.attach(nil, false)

	// 창 안에서 20번 두드린다. 시도는 하나뿐이어야 한다 (single-flight).
	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			attachment.StrategyRuntimeConfigured()
		}()
	}
	wg.Wait()
	<-entered
	countCalls := func() int {
		mu.Lock()
		defer mu.Unlock()
		return calls
	}
	if got := countCalls(); got != 1 {
		t.Fatalf("동시 요청 20건이 dial 을 %d번 불렀다, want 1 — single-flight 가 없다", got)
	}
	close(release)
	a109WaitFor(t, "첫 시도 종료", func() bool { return !attachment.inFlight() })

	// 시계를 안 밀고 또 두드린다. rate limit 이 막아야 한다.
	for range 20 {
		attachment.StrategyRuntimeConfigured()
	}
	time.Sleep(50 * time.Millisecond)
	if got := countCalls(); got != 1 {
		t.Fatalf("창 안 재시도가 dial 을 %d번 불렀다, want 1 — rate limit 이 없다. "+
			"엔진이 내려간 동안 read 마다 200ms probe 비용을 낸다", got)
	}

	// 창을 지나면 다시 시도한다.
	clock.advance(2 * time.Minute)
	attachment.StrategyRuntimeConfigured()
	a109WaitFor(t, "창 이후 재시도", func() bool { return countCalls() == 2 })
}

// TestAFailedAttemptDoesNotClobberTheCurrentScreen 은 뒷문으로 들어오는 접힘을 막는다.
//
// 붙어 있다가 잠깐 못 읽는 상태(unavailable)에서 시도가 실패했다고 reader 를 nil 로
// 갈아끼우면, 화면이 「엔진이 없다」에서 「이 배포는 전략 화면을 안 쓴다」로 바뀐다.
// a108 D4-2 가 금지한 접힘이 재부착 경로로 다시 들어오는 모양이다.
func TestAFailedAttemptDoesNotClobberTheCurrentScreen(t *testing.T) {
	clock := &a109Clock{now: time.Unix(1_760_000_000, 0).UTC()}
	var attempts atomic.Int64
	attachment := a109Attachment(t, clock, io.Discard,
		func(context.Context) (httpapi.StrategyRuntimeReader, bool) {
			attempts.Add(1)
			return nil, false // 시도 실패: 부재로 보이더라도 갈아끼우면 안 된다
		})
	live := &a109FakeReader{}
	attachment.attach(live, true)
	live.fail(errors.New("strategy projection runtime: socket has no listener"))

	if _, err := attachment.Read(context.Background()); err == nil {
		t.Fatal("죽은 reader 가 성공을 보고했다")
	}
	a109WaitFor(t, "실패 뒤 시도", func() bool {
		return attempts.Load() >= 1 && !attachment.inFlight()
	})

	if httpapi.StrategyRuntimeAbsent(attachment) {
		t.Fatal("실패한 시도가 화면을 「부재」로 바꿨다 — 운영자는 「엔진이 죽었다」와 " +
			"「기능을 안 켰다」를 구별할 수 없게 된다")
	}
}

// TestTheAttachmentReportsOnlyTransitions 는 freeze P2-4 다.
//
// 실패는 침묵이고 전이만 말한다. 30초마다 같은 두 줄을 찍으면 엔진이 하루 내려간
// 배포에서 그 로그는 2880줄이고, 그 안에서 진짜 사건은 보이지 않는다.
func TestTheAttachmentReportsOnlyTransitions(t *testing.T) {
	clock := &a109Clock{now: time.Unix(1_760_000_000, 0).UTC()}
	log := &a109SyncWriter{}
	recovered := &a109FakeReader{}
	var succeed atomic.Bool
	attachment := a109Attachment(t, clock, log,
		func(context.Context) (httpapi.StrategyRuntimeReader, bool) {
			if succeed.Load() {
				return recovered, true
			}
			return nil, false
		})
	live := &a109FakeReader{}
	attachment.attach(live, true)

	// ① 탈착: 읽기가 실패하기 시작한다. 여러 번 실패해도 한 줄이다.
	live.fail(errors.New("strategy projection runtime: socket is invalid"))
	for range 5 {
		_, _ = attachment.Read(context.Background())
		clock.advance(2 * time.Minute)
	}
	a109WaitFor(t, "실패 시도 종료", func() bool { return !attachment.inFlight() })
	if got := strings.Count(log.String(), "읽기가 실패했다"); got != 1 {
		t.Errorf("탈착 보고 %d줄, want 1 — 실패 반복은 침묵이다\n%s", got, log.String())
	}

	// ② 부착: 엔진이 돌아온다. 이것도 한 줄이다.
	succeed.Store(true)
	clock.advance(2 * time.Minute)
	attachment.StrategyRuntimeConfigured()
	a109WaitFor(t, "재부착", func() bool {
		return strings.Contains(log.String(), "다시 붙었다")
	})
	for range 5 {
		_, _ = attachment.Read(context.Background())
		clock.advance(2 * time.Minute)
	}
	if got := strings.Count(log.String(), "다시 붙었다"); got != 1 {
		t.Errorf("부착 보고 %d줄, want 1\n%s", got, log.String())
	}
}

// TestTheAbsenceSignalIsOneJudgementForEveryConsumer 는 P1-4 의 판정 한 벌 요구다.
//
// 세 소비자(집계 스냅샷·REST·SSE)가 같은 값을 같은 뜻으로 읽는지 본다. 판정이 갈리면
// 같은 디스크 상태가 화면마다 다른 값이 된다.
func TestTheAbsenceSignalIsOneJudgementForEveryConsumer(t *testing.T) {
	clock := &a109Clock{now: time.Unix(1_760_000_000, 0).UTC()}
	attachment := a109Attachment(t, clock, io.Discard,
		func(context.Context) (httpapi.StrategyRuntimeReader, bool) { return nil, false })

	attachment.attach(nil, false)
	if !httpapi.StrategyRuntimeAbsent(attachment) {
		t.Error("부재로 출발한 wrapper 가 「있다」로 답한다 — 화면이 dormant 에서 unavailable 로 회귀한다")
	}
	attachment.attach(unavailableStrategyRuntime{cause: errors.New("dial")}, false)
	if httpapi.StrategyRuntimeAbsent(attachment) {
		t.Error("sentinel(붙지 못함)이 「없다」로 답한다 — 「엔진이 죽었다」가 「기능을 안 켰다」로 접힌다")
	}
	attachment.attach(&a109FakeReader{}, true)
	if httpapi.StrategyRuntimeAbsent(attachment) {
		t.Error("live 부착이 「없다」로 답한다")
	}
	// 상태를 말하지 않는 reader 와 진짜 nil 은 nil 검사 시절의 뜻 그대로다.
	if httpapi.StrategyRuntimeAbsent(&a109FakeReader{}) {
		t.Error("평범한 reader 가 「없다」로 답한다")
	}
	if !httpapi.StrategyRuntimeAbsent(nil) {
		t.Error("nil 이 「있다」로 답한다")
	}
}

package reconcile_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
	"github.com/JungHoonGhae/tossinvest-cli/internal/soak"
)

// a102 겹1 — 재시작 복구는 rate limit을 영구 실패와 구분한다.
//
// 관측된 사고: 2026-08-13 02:03:30.545Z, 부팅 직후 계좌 읽기가 429로 거부되자 복구가
// 즉시 미완료로 끝났다. 같은 429를 주기 대사는 다음 주기로 넘기고(2026-08-12T12:52:37Z),
// 조회만 하는 서베이는 15초 백오프로 버틴다. **보호를 든 쪽이 조회만 하는 쪽보다 먼저
// 포기하는 비대칭**이 여기서 지워진다.

// --- 시계와 브로커 fake ------------------------------------------------------

// virtualClock은 잠들지 않고 시간을 앞으로 옮기며, 무엇을 얼마나 기다렸는지 적는다.
//
// 벽시계 sleep이 없으므로 15초·2초 같은 실제 값을 그대로 쓰면서도 테스트는 즉시 끝난다.
// 그리고 대기 시간이 관측 대상이므로(Report.RateLimitWaited) 그 값을 fake가 직접 세는
// 것이 아니라 **코드가 시계에 요청한 값**을 적는다.
type virtualClock struct {
	mu    sync.Mutex
	now   time.Time
	waits []time.Duration
}

func newVirtualClock(at time.Time) *virtualClock { return &virtualClock{now: at.UTC()} }

func (c *virtualClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *virtualClock) Since(t time.Time) time.Duration { return c.Now().Sub(t) }

func (c *virtualClock) Sleep(ctx context.Context, d time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if d <= 0 {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.waits = append(c.waits, d)
	c.now = c.now.Add(d)
	return nil
}

func (c *virtualClock) recorded() []time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]time.Duration(nil), c.waits...)
}

// throttledOrders는 처음 refusals번의 목록 읽기를 지정한 오류로 거부하고, 그 뒤로는
// 빈 목록을 돌려준다. 브로커의 penalty window가 걷히는 모양이다.
//
// forever가 참이면 영원히 거부한다 — 브로커가 돌아오지 않는 경우다. 유한한 거부는
// "언젠가는 읽힌다"를 가정하므로, 예산이 실제로 무엇을 막는지는 그것으로 잴 수 없다.
//
// **ctx를 일부러 보지 않는다.** 실제 HTTP 클라이언트는 취소된 ctx에서 즉시 오류를
// 내지만, 그러면 "종료 뒤에도 계속 조회했다"는 결함이 fake에 가려진다. 이 fake는
// 부르면 답하므로 호출 수가 그대로 증거가 된다.
type throttledOrders struct {
	rec      *recorder
	refusals int
	forever  bool
	err      error
	calls    int
}

func (o *throttledOrders) OrdersPageRaw(_ context.Context, _ execgw.OrderQuery, _ string) (execgw.OrderPage, error) {
	o.rec.log("orders")
	o.calls++
	if o.forever {
		return execgw.OrderPage{}, o.err
	}
	if o.refusals > 0 {
		o.refusals--
		return execgw.OrderPage{}, o.err
	}
	return execgw.OrderPage{}, nil
}

// a102Collector는 주문 목록만 거부하는 수집기다. 세 읽기 중 첫 번째가 거부되면
// 스냅샷은 통째로 폐기되므로(Collect의 "or none"), 한 갈래만 흔들어도 429 경로 전체가
// 재현된다.
func a102Collector(clk clock.Clock, orders *throttledOrders) *reconcile.Collector {
	rec := &recorder{}
	orders.rec = rec
	return &reconcile.Collector{
		Orders:     orders,
		Positions:  &fakeHoldings{rec: rec},
		Balance:    &fakeBalance{rec: rec, cash: map[string]float64{"USD": 1000}},
		Currencies: []string{"USD"},
		AccountRef: "acct-7",
		Clock:      clk,
	}
}

func a102Options(
	j *journal.Journal,
	gate *execgw.EntryGate,
	c *reconcile.Collector,
	clk clock.Clock,
	stab reconcile.Stabilisation,
) reconcile.Options {
	return reconcile.Options{
		Journal:    j,
		Resolver:   newResolver(j, gate),
		Collector:  c,
		Gate:       gate,
		Clock:      clk,
		AccountRef: "acct-7",
		Stabilise:  stab,
	}
}

// a102Stabilisation은 이 파일의 기준 설정이다. 안정화 간격 2초는 기본값과 같고
// (DefaultStabilisationInterval), MaxAttempts 2는 **429가 시도를 소모하면 즉시 소진되는**
// 가장 작은 예산이다 — 그것이 이 파일의 핵심 주장을 반증 가능하게 만든다.
func a102Stabilisation() reconcile.Stabilisation {
	return reconcile.Stabilisation{
		Interval:         2 * time.Second,
		Required:         2,
		MaxAttempts:      2,
		RateLimitBackoff: 15 * time.Second,
		MaxRateLimitWait: 5 * time.Minute,
	}
}

// --- 시나리오: 429 두 번 뒤 안정 스냅샷 --------------------------------------

// TestRateLimitedCollectDoesNotEndRecovery: 브로커가 두 번 거부한 뒤 읽히면 복구는
// 완주하고 진입 게이트가 열린다.
func TestRateLimitedCollectDoesNotEndRecovery(t *testing.T) {
	j := openJournal(t)
	gate := execgw.NewEntryGate(clock.NewFake(asOf), map[execgw.RequiredQuery]time.Duration{})
	clk := newVirtualClock(asOf)
	orders := &throttledOrders{refusals: 2, err: official.ErrRateLimited}

	r, err := reconcile.New(a102Options(j, gate, a102Collector(clk, orders), clk, a102Stabilisation()))
	if err != nil {
		t.Fatal(err)
	}
	report, err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v — a rate limit is not evidence that the account is unstable", err)
	}
	if !report.Complete || !r.Complete() {
		t.Fatalf("report = %+v, want complete", report)
	}
	if rejected := gate.CheckEntry(); rejected != nil {
		t.Fatalf("a recovery that survived the rate limit must release the latch, got %v", rejected)
	}
	if orders.calls != 4 {
		t.Fatalf("order-list reads = %d, want 4 (2 refused + 2 stabilising)", orders.calls)
	}
}

// TestRateLimitDoesNotConsumeAStabilisationAttempt: 거부는 안정화 시도가 아니다.
//
// MaxAttempts는 2다. 429가 시도를 소모하면 두 거부만으로 예산이 끝나 복구가 실패한다.
// 그러므로 이 테스트가 초록이라는 것은 "거부는 읽기가 아니었다"가 코드에 실제로
// 있다는 뜻이다. taken(SnapshotsTaken)도 같은 규율을 진다 — 읽지 못한 것은 세지 않는다.
func TestRateLimitDoesNotConsumeAStabilisationAttempt(t *testing.T) {
	j := openJournal(t)
	gate := execgw.NewEntryGate(clock.NewFake(asOf), map[execgw.RequiredQuery]time.Duration{})
	clk := newVirtualClock(asOf)
	orders := &throttledOrders{refusals: 2, err: official.ErrRateLimited}

	r, err := reconcile.New(a102Options(j, gate, a102Collector(clk, orders), clk, a102Stabilisation()))
	if err != nil {
		t.Fatal(err)
	}
	report, err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if report.SnapshotsTaken != 2 {
		t.Errorf("SnapshotsTaken = %d, want 2 — a refused read is not a snapshot", report.SnapshotsTaken)
	}
	if report.RateLimitWaits != 2 {
		t.Errorf("RateLimitWaits = %d, want 2", report.RateLimitWaits)
	}
	if report.RateLimitWaited != 30*time.Second {
		t.Errorf("RateLimitWaited = %s, want 30s", report.RateLimitWaited)
	}

	// 실제로 요청된 대기: 429 백오프 둘 + 안정화 간격 하나.
	want := []time.Duration{15 * time.Second, 15 * time.Second, 2 * time.Second}
	got := clk.recorded()
	if len(got) != len(want) {
		t.Fatalf("waits = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("waits = %v, want %v", got, want)
		}
	}
}

// TestRateLimitDefaultsMatchTheSurveyDiscipline: 노브를 비워 두면 서베이와 같은 규율이
// 나온다.
//
// 15초는 임의의 숫자가 아니라 **같은 계좌 예산을 쓰는 read-only 서베이의 첫 재시도
// 간격**이다(soak.ReadRetryBackoff(0)). 보호를 든 쪽이 조회만 하는 쪽보다 먼저 포기하지
// 않는다는 하한을 상수로 고정한다.
func TestRateLimitDefaultsMatchTheSurveyDiscipline(t *testing.T) {
	if reconcile.DefaultRateLimitBackoff != soak.ReadRetryBackoff(0) {
		t.Errorf("DefaultRateLimitBackoff = %s, want the survey's first backoff %s",
			reconcile.DefaultRateLimitBackoff, soak.ReadRetryBackoff(0))
	}
	if reconcile.DefaultRateLimitBackoff != 15*time.Second {
		t.Errorf("DefaultRateLimitBackoff = %s, want 15s", reconcile.DefaultRateLimitBackoff)
	}
	if reconcile.DefaultMaxRateLimitWait != 5*time.Minute {
		t.Errorf("DefaultMaxRateLimitWait = %s, want 5m", reconcile.DefaultMaxRateLimitWait)
	}

	// 그리고 zero 값이 실제로 그 기본을 쓴다 — 상수만 맞고 배선이 없으면 소용없다.
	j := openJournal(t)
	gate := execgw.NewEntryGate(clock.NewFake(asOf), map[execgw.RequiredQuery]time.Duration{})
	clk := newVirtualClock(asOf)
	orders := &throttledOrders{refusals: 1, err: official.ErrRateLimited}
	stab := reconcile.Stabilisation{Interval: 2 * time.Second, Required: 2, MaxAttempts: 2}

	r, err := reconcile.New(a102Options(j, gate, a102Collector(clk, orders), clk, stab))
	if err != nil {
		t.Fatal(err)
	}
	report, err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if report.RateLimitWaited != reconcile.DefaultRateLimitBackoff {
		t.Fatalf("RateLimitWaited = %s, want the default backoff %s",
			report.RateLimitWaited, reconcile.DefaultRateLimitBackoff)
	}
	if got := clk.recorded(); len(got) == 0 || got[0] != reconcile.DefaultRateLimitBackoff {
		t.Fatalf("waits = %v, want the first one to be the default backoff %s",
			got, reconcile.DefaultRateLimitBackoff)
	}
}

// --- 경계 ---------------------------------------------------------------------

// TestNonRateLimitedCollectStillFailsImmediately: 429가 아닌 오류의 처리는 오늘 그대로다.
//
// 이 change는 복구를 더 끈질기게 만들 뿐, 못 읽는 계좌를 기다려 주지 않는다.
func TestNonRateLimitedCollectStillFailsImmediately(t *testing.T) {
	j := openJournal(t)
	gate := execgw.NewEntryGate(clock.NewFake(asOf), map[execgw.RequiredQuery]time.Duration{})
	clk := newVirtualClock(asOf)
	orders := &throttledOrders{refusals: 1, err: errors.New("broker said no")}

	r, err := reconcile.New(a102Options(j, gate, a102Collector(clk, orders), clk, a102Stabilisation()))
	if err != nil {
		t.Fatal(err)
	}
	report, err := r.Run(context.Background())
	if !errors.Is(err, reconcile.ErrRecoveryIncomplete) {
		t.Fatalf("err = %v, want ErrRecoveryIncomplete", err)
	}
	if report.Complete || r.Complete() {
		t.Fatal("an unreadable account must not be reported as recovered")
	}
	if orders.calls != 1 {
		t.Errorf("order-list reads = %d, want 1 — a 400 answered twice is the same 400", orders.calls)
	}
	if waits := clk.recorded(); len(waits) != 0 {
		t.Errorf("waits = %v, want none — only a rate limit is waited out", waits)
	}
	if report.RateLimitWaits != 0 || report.RateLimitWaited != 0 {
		t.Errorf("rate-limit report = %d/%s, want zero", report.RateLimitWaits, report.RateLimitWaited)
	}
	rejected := gate.CheckEntry()
	if rejected == nil || rejected.Reason != execgw.ReasonRecoveryIncomplete {
		t.Fatalf("the latch must stay shut, got %v", rejected)
	}
}

// TestRateLimitWaitBudgetExhaustionFailsClosed: 기다림은 유한하다.
//
// 예산은 40초, 백오프는 15초다. 두 번(30초)까지는 기다리고, 세 번째를 더하면 45초로
// 예산을 넘으므로 **넘겨서 기다리지 않고** 미완료로 실패한다. 브로커는 아직 거부를
// 백 번 남겨 두고 있다 — 죽은 브로커를 조용히 영원히 기다리지 않는다는 상한이다.
func TestRateLimitWaitBudgetExhaustionFailsClosed(t *testing.T) {
	j := openJournal(t)
	gate := execgw.NewEntryGate(clock.NewFake(asOf), map[execgw.RequiredQuery]time.Duration{})
	clk := newVirtualClock(asOf)
	orders := &throttledOrders{refusals: 100, err: official.ErrRateLimited}
	stab := a102Stabilisation()
	stab.MaxRateLimitWait = 40 * time.Second

	r, err := reconcile.New(a102Options(j, gate, a102Collector(clk, orders), clk, stab))
	if err != nil {
		t.Fatal(err)
	}
	report, err := r.Run(context.Background())
	if !errors.Is(err, reconcile.ErrRecoveryIncomplete) {
		t.Fatalf("err = %v, want ErrRecoveryIncomplete", err)
	}
	if report.Complete || r.Complete() {
		t.Fatal("a recovery that never read the account must not be reported as complete")
	}
	if !strings.Contains(err.Error(), "rate limit") {
		t.Errorf("err = %q, want the reason to name the rate limit", err)
	}
	if !strings.Contains(err.Error(), "30s") {
		t.Errorf("err = %q, want the reason to carry the total wait (30s)", err)
	}
	if report.RateLimitWaits != 2 || report.RateLimitWaited != 30*time.Second {
		t.Errorf("rate-limit report = %d/%s, want 2/30s",
			report.RateLimitWaits, report.RateLimitWaited)
	}
	if waited := clk.recorded(); len(waited) != 2 {
		t.Errorf("waits = %v, want exactly 2 — the budget must not be overspent", waited)
	}
	rejected := gate.CheckEntry()
	if rejected == nil || rejected.Reason != execgw.ReasonRecoveryIncomplete {
		t.Fatalf("the gate must stay shut while and after the wait, got %v", rejected)
	}
}

// TestRateLimitBackoffStopsOnContextCancel: 백오프 대기는 종료 신호를 즉시 통과시킨다
// (안전 불변식 4 — 손절·비상 청산의 즉시성을 약화하지 않는다).
//
// 진짜 blocking seam으로 잰다: clock.Fake의 Sleep은 시간이 흐르거나 ctx가 끝날 때까지
// 실제로 막힌다. 시계를 **한 번도 앞으로 옮기지 않고** 취소만으로 빠져나오는지를 본다.
func TestRateLimitBackoffStopsOnContextCancel(t *testing.T) {
	j := openJournal(t)
	gate := execgw.NewEntryGate(clock.NewFake(asOf), map[execgw.RequiredQuery]time.Duration{})
	clk := clock.NewFake(asOf)
	orders := &throttledOrders{refusals: 100, err: official.ErrRateLimited}

	r, err := reconcile.New(a102Options(j, gate, a102Collector(clk, orders), clk, a102Stabilisation()))
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		_, runErr := r.Run(ctx)
		done <- runErr
	}()

	if !clk.WaitForSleepers(1, 2*time.Second) {
		t.Fatal("the backoff never reached the clock seam")
	}

	// F7 — 대기 **중**에도 게이트는 닫혀 있다. 이 change는 복구를 더 끈질기게 만들 뿐,
	// 읽지 못한 계좌 위에서 진입을 허용하지 않는다.
	if rejected := gate.CheckEntry(); rejected == nil || rejected.Reason != execgw.ReasonRecoveryIncomplete {
		t.Fatalf("entries must stay blocked while the backoff waits, got %v", rejected)
	}

	cancel()

	select {
	case runErr := <-done:
		if !errors.Is(runErr, reconcile.ErrRecoveryIncomplete) {
			t.Fatalf("err = %v, want ErrRecoveryIncomplete", runErr)
		}
		// F3 — 취소의 정체가 오류를 타고 나온다. 이것이 없으면 호출자는 "운영자가
		// 멈췄다"와 "브로커가 멈췄다"를 구분할 수 없고, 그 구분 불가가 이 change가
		// 지우려는 결함 그 자체다.
		if !errors.Is(runErr, context.Canceled) {
			t.Errorf("errors.Is(err, context.Canceled) = false; err = %v", runErr)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("the backoff did not release on cancellation — a shutdown would be delayed")
	}

	// F2 — 종료 뒤에는 브로커를 **한 번도 더 부르지 않는다.**
	//
	// 오류를 삼키고 루프를 계속 돌아도 결과 오류는 같아진다(취소된 ctx에서 Sleep이
	// 즉시 돌아오므로 예산이 순식간에 탄다). 운영에서 그것은 SIGTERM 순간 이미
	// 스로틀 중인 브로커에 스무 번을 연달아 다시 묻는 것이다. 호출 수만이 그 차이를
	// 본다.
	if orders.calls != 1 {
		t.Errorf("order-list reads = %d, want 1 — nothing may be asked of the broker after cancellation",
			orders.calls)
	}
	if now := clk.Now(); !now.Equal(asOf) {
		t.Fatalf("the fake clock moved to %s; the wait must have been cut short, not waited out", now)
	}
	if rejected := gate.CheckEntry(); rejected == nil {
		t.Fatal("an interrupted recovery must leave the gate shut")
	}
}

// TestRateLimitBudgetStopsExactlyAtTheBoundary: 예산이 백오프의 정확히 2배일 때
// 두 번 기다리고 멈춘다.
//
// 경계를 지나는 값이어야 반증이 된다. 40s/15s 같은 값은 두 번째 대기(30s)와 세 번째
// 시도(45s) 사이 어디서 끊기든 결과가 같아서, `>`를 `>=`로 뒤집어도 아무 테스트가
// 죽지 않는다 — A1이 실제로 그 뮤테이션을 살려 보였다(N1). 30s/15s는 두 번째 대기가
// 정확히 예산에 닿으므로 그 한 글자를 가른다.
func TestRateLimitBudgetStopsExactlyAtTheBoundary(t *testing.T) {
	j := openJournal(t)
	gate := execgw.NewEntryGate(clock.NewFake(asOf), map[execgw.RequiredQuery]time.Duration{})
	clk := newVirtualClock(asOf)
	orders := &throttledOrders{forever: true, err: official.ErrRateLimited}
	stab := a102Stabilisation()
	stab.MaxRateLimitWait = 30 * time.Second // = 2 × RateLimitBackoff

	r, err := reconcile.New(a102Options(j, gate, a102Collector(clk, orders), clk, stab))
	if err != nil {
		t.Fatal(err)
	}
	report, err := r.Run(context.Background())
	if !errors.Is(err, reconcile.ErrRecoveryIncomplete) {
		t.Fatalf("err = %v, want ErrRecoveryIncomplete", err)
	}
	if report.RateLimitWaits != 2 {
		t.Errorf("RateLimitWaits = %d, want 2 — the wait that lands exactly on the budget is allowed",
			report.RateLimitWaits)
	}
	if report.RateLimitWaited != 30*time.Second {
		t.Errorf("RateLimitWaited = %s, want the full 30s budget", report.RateLimitWaited)
	}
	if waits := clk.recorded(); len(waits) != 2 {
		t.Errorf("waits = %v, want exactly 2", waits)
	}
}

// TestPermanentlyRefusingBrokerExhaustsTheBudget: 돌아오지 않는 브로커를 영원히
// 기다리지 않는다.
//
// 앞의 테스트들은 거부 횟수가 유한하다 — "언젠가는 읽힌다"를 이미 가정한다. 예산이
// 진짜로 막는 것은 그 가정이 틀린 경우이고, 그 경우는 무한 거부로만 잰다.
// 기본값(15s 백오프 · 5m 예산)이면 스무 번 기다리고 스물한 번째 거부에서 멈춘다.
func TestPermanentlyRefusingBrokerExhaustsTheBudget(t *testing.T) {
	j := openJournal(t)
	gate := execgw.NewEntryGate(clock.NewFake(asOf), map[execgw.RequiredQuery]time.Duration{})
	clk := newVirtualClock(asOf)
	orders := &throttledOrders{forever: true, err: official.ErrRateLimited}
	stab := reconcile.Stabilisation{Interval: 2 * time.Second, Required: 2, MaxAttempts: 2}

	r, err := reconcile.New(a102Options(j, gate, a102Collector(clk, orders), clk, stab))
	if err != nil {
		t.Fatal(err)
	}
	report, err := r.Run(context.Background())
	if !errors.Is(err, reconcile.ErrRecoveryIncomplete) {
		t.Fatalf("err = %v, want ErrRecoveryIncomplete", err)
	}
	if report.RateLimitWaits != 20 || report.RateLimitWaited != 5*time.Minute {
		t.Errorf("rate-limit report = %d/%s, want 20/5m (300s ÷ 15s)",
			report.RateLimitWaits, report.RateLimitWaited)
	}
	if orders.calls != 21 {
		t.Errorf("order-list reads = %d, want 21 — twenty waits and the refusal that spent the budget",
			orders.calls)
	}
	if !strings.Contains(err.Error(), "rate limit") || !strings.Contains(err.Error(), "5m0s") {
		t.Errorf("err = %q, want the rate limit and the total wait", err)
	}
	rejected := gate.CheckEntry()
	if rejected == nil || rejected.Reason != execgw.ReasonRecoveryIncomplete {
		t.Fatalf("a broker that never answers must leave the gate shut, got %v", rejected)
	}
}

// TestBudgetTooSmallForOneBackoffSaysSo: 한 번도 기다리지 않았으면 그렇게 말한다.
//
// 예산이 백오프 한 번보다 작으면 기다림은 일어나지 않는다. 그때 "kept being refused"와
// "waited 0s of the budget"이라고 적으면 운영자는 있지도 않은 브로커 장애를 찾으러 간다.
func TestBudgetTooSmallForOneBackoffSaysSo(t *testing.T) {
	j := openJournal(t)
	gate := execgw.NewEntryGate(clock.NewFake(asOf), map[execgw.RequiredQuery]time.Duration{})
	clk := newVirtualClock(asOf)
	orders := &throttledOrders{forever: true, err: official.ErrRateLimited}
	stab := a102Stabilisation()
	stab.MaxRateLimitWait = 10 * time.Second // < RateLimitBackoff(15s)

	r, err := reconcile.New(a102Options(j, gate, a102Collector(clk, orders), clk, stab))
	if err != nil {
		t.Fatal(err)
	}
	report, err := r.Run(context.Background())
	if !errors.Is(err, reconcile.ErrRecoveryIncomplete) {
		t.Fatalf("err = %v, want ErrRecoveryIncomplete", err)
	}
	if report.RateLimitWaits != 0 || report.RateLimitWaited != 0 {
		t.Errorf("rate-limit report = %d/%s, want 0/0s", report.RateLimitWaits, report.RateLimitWaited)
	}
	if waits := clk.recorded(); len(waits) != 0 {
		t.Errorf("waits = %v, want none", waits)
	}
	if orders.calls != 1 {
		t.Errorf("order-list reads = %d, want 1", orders.calls)
	}
	if !strings.Contains(err.Error(), "rate limit") {
		t.Errorf("err = %q, want the reason to name the rate limit", err)
	}
	if !strings.Contains(err.Error(), "waited 0s") {
		t.Errorf("err = %q, want it to say that nothing was waited out", err)
	}
	if strings.Contains(err.Error(), "kept being refused") {
		t.Errorf("err = %q, must not describe one refusal as a run of them", err)
	}
}

// TestRateLimitBackoffNeverGoesBelowTheSurveyInterval: 백오프의 하한은 설정이 아니라
// 규율이다.
//
// reconciliation 델타: 백오프는 "같은 계좌 예산을 쓰는 read-only 서베이의 재시도
// 간격보다 짧아서는 안 된다(SHALL NOT)". 노브가 더 짧은 값을 들고 오면 그 요구가
// 노브를 채운 사람에게 달리게 되므로, 채우는 자리에서 올린다.
func TestRateLimitBackoffNeverGoesBelowTheSurveyInterval(t *testing.T) {
	j := openJournal(t)
	gate := execgw.NewEntryGate(clock.NewFake(asOf), map[execgw.RequiredQuery]time.Duration{})
	clk := newVirtualClock(asOf)
	orders := &throttledOrders{refusals: 1, err: official.ErrRateLimited}
	stab := a102Stabilisation()
	stab.RateLimitBackoff = 5 * time.Second // 서베이 간격보다 짧다

	r, err := reconcile.New(a102Options(j, gate, a102Collector(clk, orders), clk, stab))
	if err != nil {
		t.Fatal(err)
	}
	report, err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if report.RateLimitWaited != reconcile.DefaultRateLimitBackoff {
		t.Errorf("RateLimitWaited = %s, want the floor %s — a shorter knob must be raised",
			report.RateLimitWaited, reconcile.DefaultRateLimitBackoff)
	}
	if got := clk.recorded(); len(got) == 0 || got[0] != reconcile.DefaultRateLimitBackoff {
		t.Fatalf("waits = %v, want the first one raised to %s", got, reconcile.DefaultRateLimitBackoff)
	}
}

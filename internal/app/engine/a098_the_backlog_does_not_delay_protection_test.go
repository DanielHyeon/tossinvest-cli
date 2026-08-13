package engine_test

// a098 §5.3 · R4 · R4b — 밀린 것을 보내는 일이 **보호를 늦추지 않는다.**
//
// a098 은 엔진 안에 발송자를 하나 더 세운다. 그 발송자는 네트워크를 기다리고,
// 네트워크는 죽는다. 그래서 이 change 가 스스로에게 물어야 하는 질문은
// *"밀린 알림이 나가는가"* 만이 아니라 **"그것 때문에 손절이 늦어지는가"** 다.
// 늦어지면 a098 은 a092 가 없앤 결함을 다른 자리에 다시 세운 것이 된다.
//
// # 두 피해자를 따로 잰다
//
//	R4   exit 관측 사이클    — 손절을 **판정하는** 쪽
//	R4b  동기 정지 알림      — 손절을 **알리는** 쪽
//
// 둘은 실행자와 다른 것을 공유한다. 사이클은 원장을 공유하고(핸들 하나에
// `SetMaxOpenConns(1)` — `journal.go:174`), 정지 알림은 거기에 더해 `Notifier` 의
// 뮤텍스를 공유한다. 그래서 한 테스트로 둘 다 재면 어느 통로가 살아 있는지 못 가른다.
//
// # 계측기가 왜 알림 종류로 갈리는가
//
// *"전송 수단이 죽는다"* 를 **모두에게** 죽은 transport 로 흉내 내면 **기준선도 같이
// 멈춘다** — 동기 경로도 같은 transport 를 쓰기 때문이다. 그러면 이 측정은 a098 이
// 무엇을 더했는지가 아니라 **a092 의 주제**를 재게 된다.
//
// 그래서 계측기는 **실행자가 보내는 행에서만** 멈춘다(`a098BacklogEvent`). 격리하는
// 변수는 하나다 — **「발송자가 transport 안에 갇혀 있다」**. 이것이 a098 이 새로
// 들여온 상태이고, 이 파일이 재는 것이다.
//
// # 왜 뮤테이션이 하나인가 (⛔ 침묵하지 않고 적는다)
//
// 안 C 아래에서 실행자는 **publish 를 건너 아무것도 쥐지 않는다.** 그러니 사이클을
// 늘릴 수 있는 결함 모양은 **「쥔 채 기다린다」 하나뿐**이고, 그 하나가 design D1.1 이
// 물리친 안 B — `Notifier.Flush` 로 지은 실행자다(`notifier.go:734-735` 가 `n.mu` 를
// 쥐고 `PendingAlerts(ctx, 0)` 로 **전부**를 돈다). 뮤테이션 **NN** 이 그것이다.
//
// 배치를 `all` 로 바꾸는 뮤테이션은 **일부러 안 쓴다** — 안 C 에서는 그래도 R4b 가 안
// 깨지기 때문이다. 배치는 *주기 길이*를 묶지 *정지 알림*을 묶지 않는다. 안 깨지는 것을
// 근거로 쓰면 그 표는 거짓말이 된다.

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// a098BacklogEvent 는 계측기가 「실행자가 보내는 행」을 알아보는 표식이다.
// critical 표에 없는 이름이라 동기 경로가 이 종류를 만들 일이 없다 — 즉 이 표식으로
// 갈리는 것은 **누가 보내는가**뿐이다.
const a098BacklogEvent = "a098.backlog"

// a098ExitCycleDwellMargin 은 §5.3 기준선 위에 얹는 **고정 여유**다.
//
// *"안 늘어난다"* 를 「유한하다」로 쓰면 나쁜 유한 지연이 통과한다(등록부 R4, B-P4).
// 그래서 상한은 §5.3 이 잰 값 **더하기 이 상수**이지, 측정값 자체가 아니다.
const a098ExitCycleDwellMargin = 250 * time.Millisecond

// a098StopAlertBound 는 정지 알림 하나의 **절대 상한**이다. 원장 왕복 몇 번과
// 즉답하는 publish 하나가 전부여야 한다 (3.2 실측: claim→settle 왕복 5.584 ms).
const a098StopAlertBound = 500 * time.Millisecond

// a098StopAlertNIndependence 는 N 이 백 배가 돼도 정지 알림 체류가 움직여도 되는 폭이다.
const a098StopAlertNIndependence = 250 * time.Millisecond

// a098SickTransport 는 **실행자가 보내는 행에서만** 아픈 transport 다.
//
// hold 가 nil 이 아니면 그 행에서 영원히 멈추고(R4), delay 가 0 이 아니면 그만큼
// 늦게 답한다(R4b). 나머지 종류는 즉답한다 — 위 헤더의 「왜 종류로 갈리는가」.
type a098SickTransport struct {
	hold  chan struct{}
	delay time.Duration

	mu       sync.Mutex
	entered  chan struct{} // 첫 backlog 행이 publish 에 들어간 순간
	openedUp bool
	backlog  int
	other    int
}

func (p *a098SickTransport) Publish(_ context.Context, n obs.Notification) error {
	if string(n.Type) != a098BacklogEvent {
		p.mu.Lock()
		p.other++
		p.mu.Unlock()
		return nil
	}
	p.mu.Lock()
	p.backlog++
	if !p.openedUp {
		p.openedUp = true
		close(p.entered)
	}
	p.mu.Unlock()
	if p.hold != nil {
		<-p.hold
	}
	if p.delay > 0 {
		time.Sleep(p.delay)
	}
	return nil
}

func (p *a098SickTransport) counts() (backlog, other int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.backlog, p.other
}

// a098SeedBacklog 는 실행자가 집어 갈 PENDING critical 행 n 개를 남긴다.
func a098SeedBacklog(t *testing.T, j *journal.Journal, n int) {
	t.Helper()
	ctx := context.Background()
	for i := range n {
		if _, err := j.EnqueueAlert(ctx, journal.Alert{
			EventKey: fmt.Sprintf("%s|%d", a098BacklogEvent, i),
			Type:     a098BacklogEvent,
			Severity: string(obs.SeverityCritical),
			Title:    "밀린 알림",
			Body:     "실행자가 보낼 행",
		}); err != nil {
			t.Fatalf("EnqueueAlert(%d): %v", i, err)
		}
	}
}

// a098StartDeliverer 는 프로덕션 조립부가 쓰는 그 생성자로 실행자를 세운다.
//
// 여기서 `alertDeliverer` 를 직접 짓지 않는 것이 요점이다 — 뮤테이션 NN 은
// `Context.AlertDeliverer` 를 안 B 로 바꾸고, 테스트가 그 자리를 안 지나가면
// **뮤테이션이 테스트에 안 보인다.**
func a098StartDeliverer(t *testing.T, j *journal.Journal, gate *execgw.EntryGate,
	n *obs.Notifier, clk clock.Clock) {
	t.Helper()
	aux, err := (&engine.Context{Journal: j, Entry: gate, Notifier: n}).AlertDeliverer(clk)
	if err != nil {
		t.Fatalf("AlertDeliverer: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { defer close(done); _ = aux.Run(ctx) }()
	t.Cleanup(func() {
		cancel()
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			t.Error("배달 실행자가 취소 뒤에도 안 멈췄다")
		}
	})
}

// a098Notifier 는 exit 사이클과 실행자가 **같이** 쓰는 하나의 Notifier 다.
//
// 하나여야 한다. 둘로 나누면 뮤텍스도 둘이 되어 안 B 의 결함이 이 테스트에서
// 사라진다 — 즉 통과하는 이유가 「안 C 라서」가 아니라 「안 재서」가 된다.
func a098Notifier(j *journal.Journal, gate *execgw.EntryGate, pub obs.Publisher,
	clk clock.Clock) *obs.Notifier {
	return &obs.Notifier{
		Publisher: pub, Journal: j, Gate: gate,
		AccountRef: exitAccount, Clock: clk, Attempts: 1,
	}
}

// TestTheExitCycleDoesNotLengthenWhileTheSenderIsStuckInTheTransport 는
// §5.3(기준선)과 R4(대조)를 한 테스트에서 잰다 — 둘은 같은 저울이어야 비교가 된다.
func TestTheExitCycleDoesNotLengthenWhileTheSenderIsStuckInTheTransport(t *testing.T) {
	// §5.3 — 실행자가 **없는** 사이클. R4 의 GREEN 값은 이 수에서 나온다.
	baseline, baseCycle := a098MeasureRefusingCycle(t, false)
	if baseCycle.Judged == 0 {
		t.Fatalf("기준선 사이클이 아무것도 판정하지 않았다 (%+v) — 잰 것이 없다", baseCycle)
	}
	t.Logf("§5.3 기준선: 실행자 없는 exit 사이클 체류 = %v", baseline)

	// R4 — 실행자가 transport 안에 **갇힌 채로** 도는 사이클.
	stuck, stuckCycle := a098MeasureRefusingCycle(t, true)
	t.Logf("R4: 발송자가 transport 에 갇힌 동안 exit 사이클 체류 = %v", stuck)

	if stuckCycle.Judged != baseCycle.Judged {
		t.Fatalf("두 사이클이 같은 일을 안 했다: 기준선 judged=%d, 대조 judged=%d — 비교가 성립 안 한다",
			baseCycle.Judged, stuckCycle.Judged)
	}
	if bound := baseline + a098ExitCycleDwellMargin; stuck > bound {
		t.Errorf("갇힌 발송자 옆에서 exit 사이클이 %v 걸렸다; 상한 %v (= §5.3 기준선 %v + 고정 여유 %v)",
			stuck, bound, baseline, a098ExitCycleDwellMargin)
	}
}

// a098MeasureRefusingCycle 은 **critical 알림을 한 번 내는** exit 사이클 하나를 재고
// 돌려준다. withStuckSender 면 그 사이클 옆에서 배달 실행자가 publish 에 갇혀 있다.
//
// 알림을 내는 사이클을 고른 것이 핵심이다. 조용한 사이클은 `Notifier` 를 안 지나가고,
// 안 지나가면 안 B 의 뮤텍스가 이 측정에 **보이지 않는다** — 공허하게 통과한다.
func a098MeasureRefusingCycle(t *testing.T, withStuckSender bool) (time.Duration, engine.ExitCycle) {
	t.Helper()

	notifier := &obs.Notifier{} // 원장이 아직 없다. 하네스가 연 뒤에 채운다.
	ladder := exitpolicy.DefaultLadderPolicy()
	ladder.Rungs[0] = exitpolicy.Rung{TargetPct: "0.5", StopPct: "0", PartialRatio: "1"}
	ladder.PolicyDigest = "" // 재해석된 상태 → 판정 거부 → critical 알림
	h := newExitHarness(t, func(opts *engine.ExitObserverOptions) {
		opts.Ladder = &ladder
		opts.Alerts = notifier
	})

	transport := &a098SickTransport{entered: make(chan struct{})}
	if withStuckSender {
		transport.hold = make(chan struct{})
	}
	*notifier = *a098Notifier(h.journal, h.gate, transport, h.clk)

	p := h.entry("005930", "10", "10000", "9800", "10000")
	if _, err := h.journal.OpenExitState(context.Background(), journal.ExitStateSeed{
		PositionID: p.ID, PolicyKind: journal.ExitPolicyLadder, PolicyID: "default_v1",
		EntryPrice: "10000", InitialStop: "9800",
	}); err != nil {
		t.Fatalf("OpenExitState: %v", err)
	}
	h.quote("005930", 10100)

	if withStuckSender {
		a098SeedBacklog(t, h.journal, 5)
		a098StartDeliverer(t, h.journal, h.gate, notifier, h.clk)
		// 실행자를 **띄운 뒤에** 등록한다. Cleanup 은 LIFO 라서 이것이 먼저 돌고,
		// 그래야 실행자가 나갈 길이 생긴다 — 취소만으로는 못 나온다. publish 가
		// 채널에서 자고 있고, 그 잠은 ctx 를 안 본다.
		var once sync.Once
		t.Cleanup(func() { once.Do(func() { close(transport.hold) }) })
		select {
		case <-transport.entered:
		case <-time.After(10 * time.Second):
			t.Fatal("실행자가 밀린 행의 publish 에 못 들어갔다 — 갇힌 상태를 못 만들었다")
		}
	}

	type measured struct {
		cycle engine.ExitCycle
		dwell time.Duration
	}
	done := make(chan measured, 1)
	go func() {
		start := time.Now()
		cycle := h.observer.ObserveOnce(context.Background())
		done <- measured{cycle, time.Since(start)}
	}()

	var got measured
	select {
	case got = <-done:
	case <-time.After(15 * time.Second):
		// 안 B 는 여기서 걸린다: 실행자가 `n.mu` 를 쥔 채 안 돌아오는 publish 를
		// 기다리고, 알림을 내려는 사이클이 그 뮤텍스 뒤에 선다. 영원히.
		t.Fatalf("exit 사이클이 15초 안에 안 끝났다 (갇힌 발송자=%v) — "+
			"발송자가 무언가를 쥔 채 네트워크를 기다리고 있다", withStuckSender)
	}
	if got.cycle.Err != nil {
		t.Fatalf("exit 사이클이 실패했다: %v", got.cycle.Err)
	}

	// 이 사이클이 정말 Notifier 를 지나갔는가. 안 지나갔으면 위 측정은 공허하다.
	if _, other := transport.counts(); other != 1 {
		t.Fatalf("사이클이 낸 critical 알림 publish = %d, 1 이어야 한다 — "+
			"이 사이클은 Notifier 를 안 지나갔고, 그러면 뮤텍스 통로를 안 잰 것이다", other)
	}
	return got.dwell, got.cycle
}

// TestAStopAlertDoesNotWaitBehindTheBacklog 는 R4b 다 — **절대 상한과 N-독립성 둘 다.**
//
// 하나만으로는 부족하다. 절대 상한만 보면 N 에 비례하되 작은 상수의 지연이 통과하고,
// N-독립성만 보면 **모든 N 에서 똑같이 나쁜** 지연이 통과한다.
func TestAStopAlertDoesNotWaitBehindTheBacklog(t *testing.T) {
	dwell := map[int]time.Duration{}
	for _, n := range []int{10, 1000} {
		dwell[n] = a098MeasureStopAlert(t, n)
		t.Logf("R4b: 밀린 행 %d 개 · 실행자가 도는 동안 정지 알림 체류 = %v", n, dwell[n])
		if dwell[n] > a098StopAlertBound {
			t.Errorf("밀린 행 %d 개 뒤에서 정지 알림이 %v 걸렸다; 절대 상한 %v",
				n, dwell[n], a098StopAlertBound)
		}
	}
	spread := dwell[1000] - dwell[10]
	if spread < 0 {
		spread = -spread
	}
	if spread > a098StopAlertNIndependence {
		t.Errorf("밀린 행이 10 → 1000 으로 늘 때 정지 알림 체류가 %v 움직였다 (허용 %v) — "+
			"백로그 길이가 정지 알림을 밀고 있다: 10개=%v, 1000개=%v",
			spread, a098StopAlertNIndependence, dwell[10], dwell[1000])
	}
}

// a098MeasureStopAlert 는 밀린 행 n 개를 두고 실행자를 돌리며 **동기** 정지 알림
// 하나의 체류를 잰다.
//
// 실행자가 publish 안에 있는 것을 **본 뒤에** 잰다. 자고 있는 실행자 옆에서 재면
// 무엇과도 경합하지 않은 수가 나오고, 그 수는 안 B 에서도 빠르다.
func a098MeasureStopAlert(t *testing.T, n int) time.Duration {
	t.Helper()
	clk := clock.NewFake(exitNow)
	j, err := journal.Open(context.Background(), journal.Options{
		Path:     filepath.Join(t.TempDir(), journal.DBFileName),
		Clock:    clk,
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
	})
	if err != nil {
		t.Fatalf("journal.Open: %v", err)
	}
	t.Cleanup(func() { _ = j.Close() })
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{
		execgw.QueryPrice: 15 * time.Second,
	})

	// 멈추지 않고 **느리게** 답한다. 멈춰 버리면 안 B 도 한 행에서 서 버려서
	// N-독립성이 공짜로 참이 된다 — 재지 않고 통과하는 그 모양이다.
	transport := &a098SickTransport{entered: make(chan struct{}), delay: 5 * time.Millisecond}
	notifier := a098Notifier(j, gate, transport, clk)

	a098SeedBacklog(t, j, n)
	a098StartDeliverer(t, j, gate, notifier, clk)
	select {
	case <-transport.entered:
	case <-time.After(20 * time.Second):
		t.Fatal("실행자가 밀린 행의 publish 에 못 들어갔다 — 경합 상태를 못 만들었다")
	}

	type measured struct {
		err   error
		dwell time.Duration
	}
	done := make(chan measured, 1)
	go func() {
		start := time.Now()
		err := notifier.Notify(context.Background(), obs.Event{
			Type:  obs.EventExitProposalRefused,
			Key:   fmt.Sprintf("a098-stop-alert|%d", n),
			Title: "청산 주문이 제출되지 않았다",
			Body:  "손절이 나가지 않았다 — 사람이 봐야 한다",
		})
		done <- measured{err, time.Since(start)}
	}()

	var got measured
	select {
	case got = <-done:
	case <-time.After(60 * time.Second):
		t.Fatalf("밀린 행 %d 개 뒤에서 정지 알림이 60초 안에 안 나갔다 — "+
			"발송자가 백로그를 도는 동안 알림 경로를 잠그고 있다", n)
	}
	if got.err != nil {
		t.Fatalf("정지 알림 Notify: %v", got.err)
	}
	if _, other := transport.counts(); other != 1 {
		t.Fatalf("정지 알림 publish = %d, 1 이어야 한다 — 잰 것이 그 알림이 아니다", other)
	}
	return got.dwell
}

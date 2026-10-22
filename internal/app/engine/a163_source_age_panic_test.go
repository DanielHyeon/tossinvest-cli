package engine_test

// a163 tasks 2.10 — 보고 경로의 패닉은 청산 관측 루프를 끝내지 않는다(freeze 리뷰 P1-4, design D4 「패닉 격리」 · D6).
//
// 감독자(Runtime.Run · ExitObserver.Run)에는 recover 가 없어 보고 코드의 패닉은 곧 프로세스 종료 = 손절 없음임. 보고 진입 함수의
// recover 경계가 거두는지를 두 주입으로 잼: 전용 로거의 쓰기 대상이 패닉(줄 쓰기 · 거둔 뒤의 고정 문구 줄 둘 다), 알림기가 패닉.
// 래치 맵은 ObserveOnce 의 goroutine 만 쓰므로(design D4 「단일 goroutine 소유」) 이 파일은 Makefile RACE_ENGINE_FILES 에 있음 —
// 시험이 Run 을 별도 goroutine 에서 돌려 그 소유 전제를 -race 아래에서 지나게 함.

import (
	"context"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

type a163PanickingWriter struct{}

func (a163PanickingWriter) Write([]byte) (int, error) { panic("a163: injected log sink panic") }

type a163PanickingAlerts struct{ inner engine.ExitAlerter }

func (p a163PanickingAlerts) Notify(ctx context.Context, e obs.Event) error {
	if e.Type == obs.EventExitSourceMissing {
		panic("a163: injected alert panic")
	}
	return p.inner.Notify(ctx, e)
}

// a163PanicRun 은 같은 입력의 두 주기(첫 주기는 판정만, 둘째 주기에 손절선 아래)를 Run goroutine 으로 돌리고 제출 수와 a090 줄 수를 돌려줌.
func a163PanicRun(t *testing.T, inject func(*engine.ExitObserverOptions)) (places int, unobservedLines int) {
	t.Helper()
	unobserved := &a090Buffer{}
	f := a163Setup(t, true, func(o *engine.ExitObserverOptions) {
		o.UnobservedLog = obs.NewLogger(obs.LogOptions{Writer: unobserved, JSON: true, Clock: o.Clock})
		if inject != nil {
			inject(o)
		}
	})
	f.entry("005930", "10", "70000", "68000", "70000")
	f.entry("000660", "4", "120000", "118000", "120000")
	_ = f.entry("035420", "5", "200000", "190000", "200000") // 시세 없음 — settleUnobserved 가 돌았다는 표지 줄
	f.quote("005930", 70100)
	f.quote("000660", 120100)
	f.stamps["005930"], f.stamps["000660"] = "", "garbage"

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- f.observer.Run(ctx) }()
	waitSleep := func() {
		t.Helper()
		if !f.clk.WaitForSleepers(1, 5*time.Second) {
			t.Fatal("the observation loop did not reach its interval sleep — a report panic ended the goroutine")
		}
	}
	waitSleep() // 첫 주기(보고가 패닉하는 주기) 끝
	select {
	case err := <-done:
		t.Fatalf("Run returned after the first cycle: %v", err)
	default:
	}
	f.quote("005930", 67000) // 둘째 주기: 손절선 아래
	f.clk.Advance(f.observer.Interval())
	waitSleep()
	cancel()
	f.clk.Advance(f.observer.Interval())
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after cancel")
	}
	return len(f.submit.places), len(a163Lines(t, unobserved.String(), obs.EventExitPositionUnobserved))
}

func TestA163APanickingReportNeitherStopsTheLoopNorChangesTheStops(t *testing.T) {
	wantPlaces, wantUnobserved := a163PanicRun(t, nil)
	if wantPlaces != 1 || wantUnobserved == 0 {
		t.Fatalf("control: places=%d a090 lines=%d, want the liquidation and the settlement's line", wantPlaces, wantUnobserved)
	}
	for name, inject := range map[string]func(*engine.ExitObserverOptions){
		"log sink panics": func(o *engine.ExitObserverOptions) {
			o.SourceAgeLog = obs.NewLogger(obs.LogOptions{Writer: a163PanickingWriter{}, JSON: true, Clock: o.Clock})
		},
		"alerter panics": func(o *engine.ExitObserverOptions) { o.Alerts = a163PanickingAlerts{inner: o.Alerts} },
	} {
		t.Run(name, func(t *testing.T) {
			places, unobserved := a163PanicRun(t, inject)
			if places != wantPlaces {
				t.Fatalf("places = %d with a panicking report, want %d — the next cycle's stop must still go out (safety 4)", places, wantPlaces)
			}
			if unobserved != wantUnobserved {
				t.Fatalf("a090 lines = %d, want %d — settleUnobserved must still run after a recovered report panic", unobserved, wantUnobserved)
			}
		})
	}
}

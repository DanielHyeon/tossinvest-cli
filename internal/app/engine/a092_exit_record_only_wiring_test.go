package engine_test

// a092 22.3 C1: exit 관측 goroutine 에서 알림 경로에 닿는 입구 셋(알림 · 관측 두절 강화 통지 · Retrier 401 강화 통지 — 가격 조회와
// 청산 상한 조회)은 주입 지점별로 기록 전용 인스턴스를 받아야 함. 공유 Retrier 는 그대로여야 함(대사 루프 등 범위 밖 소비자 — Q1 문자 해석).

import (
	"reflect"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

func TestA092ExitObserverGetsRecordOnlyAlertPaths(t *testing.T) {
	dir := isolate(t)
	writeGateConfig(t, dir, smallLiveGate())
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	writeAttestation(t, dir, nil)
	srv, _ := interlockServer(t, "123-45")

	eng, err := openProtectedGateEngine(t, dir, srv, nil)
	if err != nil {
		t.Fatalf("production assembly: %v", err)
	}
	if eng.Notifier == nil || eng.Retrier == nil {
		t.Fatal("assembly without a notifier or retrier — this test measures nothing")
	}
	shared := a092RetrierShape(*eng.Retrier)

	observer, err := eng.ExitObserver(engine.ExitObserverOptions{Costs: costs.DefaultModel()})
	if err != nil {
		t.Fatalf("ExitObserver: %v", err)
	}
	opts := observer.OptionsForTest()
	recordOnly := obs.RecordOnly{N: eng.Notifier, Relay: eng.NormalAlertRelay()}
	if recordOnly.Relay == nil {
		t.Fatal("the engine has no normal-grade relay — exit normal alerts would be dropped (a092 C8)")
	}

	// 알림 입구.
	if got, ok := opts.Alerts.(obs.RecordOnly); !ok || got != recordOnly {
		t.Errorf("Alerts = %T, want obs.RecordOnly on the engine notifier", opts.Alerts)
	}
	// 관측 두절 강화 통지(checkOutage).
	if got, ok := opts.Announcer.(obs.RecordOnly); !ok || got != recordOnly {
		t.Errorf("Announcer = %T, want obs.RecordOnly on the engine notifier", opts.Announcer)
	}
	// 가격 조회 401 강화 통지: 값 복사본이고 Announcer 만 다름.
	if opts.Retrier == eng.Retrier {
		t.Fatal("the exit loop shares the engine Retrier — a record-only announcer on it would change every other consumer")
	}
	if got, ok := opts.Retrier.Announcer.(obs.RecordOnly); !ok || got != recordOnly {
		t.Errorf("exit Retrier.Announcer = %T, want obs.RecordOnly", opts.Retrier.Announcer)
	}
	exitShape := a092RetrierShape(*opts.Retrier)
	exitShape.announcer = shared.announcer
	if exitShape != shared {
		t.Errorf("the exit Retrier differs from the shared one in more than its announcer:\n exit   %+v\n shared %+v", exitShape, shared)
	}
	// 공유 Retrier 는 그대로 — 범위 밖 소비자는 동기 통지를 유지(k4).
	if a092RetrierShape(*eng.Retrier) != shared {
		t.Error("ExitObserver mutated the shared Retrier")
	}
	if n, ok := eng.Retrier.Announcer.(*obs.Notifier); !ok || n != eng.Notifier {
		t.Errorf("shared Retrier.Announcer = %T, want the engine *obs.Notifier", eng.Retrier.Announcer)
	}
	// 청산 수량 상한 조회: exit 복사본 Retrier 를 쓰고 나머지는 엔진 공급자와 같음.
	floorRetrier, ok := engine.ExitFloorRetrierForTest(opts.Floor)
	if !ok {
		t.Fatalf("Floor = %T, want the engine reconcile floor", opts.Floor)
	}
	if floorRetrier != opts.Retrier {
		t.Errorf("the exit floor queries through %p, want the exit Retrier %p", floorRetrier, opts.Retrier)
	}
	if !engine.FloorsShareAllButRetrierForTest(opts.Floor, eng.ContextFloorForTest()) {
		t.Error("the exit floor differs from the engine floor in more than its retrier")
	}
	if shared, _ := engine.ExitFloorRetrierForTest(eng.ContextFloorForTest()); shared != eng.Retrier {
		t.Error("ExitObserver mutated the engine floor's retrier")
	}
}

// retrierShape 는 Retrier 의 비교 가능한 모양 — RetryPolicy.Rand 가 함수라 구조체 == 를 못 씀(함수는 포인터로 비교).
type retrierShape struct {
	attempts                                int
	budget, base, maxBackoff, maxRetryAfter time.Duration
	jitter                                  float64
	rand                                    uintptr
	clock, retryAfter, escalate, announcer  any
	gate                                    *execgw.EntryGate
	account                                 string
}

func a092RetrierShape(r execgw.Retrier) retrierShape {
	var rp uintptr
	if r.Policy.Rand != nil {
		rp = reflect.ValueOf(r.Policy.Rand).Pointer()
	}
	p := r.Policy
	return retrierShape{attempts: p.MaxAttempts, budget: p.Budget, base: p.BaseBackoff, maxBackoff: p.MaxBackoff,
		maxRetryAfter: p.MaxRetryAfter, jitter: p.JitterFraction, rand: rp,
		clock: r.Clock, retryAfter: r.RetryAfter, escalate: r.Escalate,
		announcer: r.Announcer, gate: r.Gate, account: r.AccountRef}
}

// 보이스 B #4: 호출자가 동기 알림기를 넘겨도 exit 관측기는 기록 전용을 받음(역할 핀 — 기본값이 아니라 덮어쓰기).
func TestA092TheExitObserverOverridesACallersSyncAlertPath(t *testing.T) {
	dir := isolate(t)
	writeGateConfig(t, dir, smallLiveGate())
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	writeAttestation(t, dir, nil)
	srv, _ := interlockServer(t, "123-45")
	eng, err := openProtectedGateEngine(t, dir, srv, nil)
	if err != nil {
		t.Fatalf("production assembly: %v", err)
	}
	observer, err := eng.ExitObserver(engine.ExitObserverOptions{Costs: costs.DefaultModel(),
		Alerts: eng.Notifier, Announcer: eng.Notifier})
	if err != nil {
		t.Fatal(err)
	}
	opts := observer.OptionsForTest()
	if _, ok := opts.Alerts.(obs.RecordOnly); !ok {
		t.Errorf("Alerts = %T — a caller-supplied sync notifier reached the exit goroutine", opts.Alerts)
	}
	if _, ok := opts.Announcer.(obs.RecordOnly); !ok {
		t.Errorf("Announcer = %T — a caller-supplied sync notifier reached the exit goroutine", opts.Announcer)
	}
}

//go:build unix

package main

// a108_the_engine_outlives_its_read_endpoint_test.go — 겹2 (tasks 3.1, 3.2).
//
// 2026-08-13 23:35 KST, 호스트가 US 정규장 중 재부팅했다. 재부팅 잔재 때문에
// `strategyprojectionrpc.Start` 가 실패했고, `runEngineRun` 이 그 오류를 그대로
// 돌려주어 엔진이 exit 1 했다. 콘솔 autostart 가 1분마다 다시 세웠고 디스크 상태는
// 그대로였으므로 **영구 기동 루프**가 됐다. 보호 루프 셋이 14:33Z 부터 전부 정지 —
// 관리 포지션의 손절 감시가 장중에 사라졌다.
//
// 죽인 것은 **조회 전용 export 표면**이다. projection 은 콘솔·httpapi 가 전략 화면을
// 그리려고 읽는 socket 하나이고, 루프의 입력이 아니다. 이 파일이 고정하는 성질은
// 하나다: **화면을 잃는 것이 판정을 잃는 이유가 되면 안 된다.**
//
// # 왜 실패를 seam 으로 주입하는가
//
// 잔재 디스크 상태로 실패를 만들면 이 테스트가 `internal/strategyprojectionrpc` 의
// 회수 규칙에 묶인다 — 그 규칙은 a108 겹1 이 **지금 바꾸고 있는 것**이다. 재는 것은
// "잔재가 어떻게 생겼는가" 가 아니라 "실패했을 때 엔진이 죽는가" 이므로 주입이 맞다.
//
// # 왜 새 파일인가
//
// 기존 파일에 덧붙이면 tools/logic-map/check_analysis.py 가 이웃 함수까지 「수정됨」
// 으로 보고, 아무도 편집하지 않은 함수의 증거 묶음을 요구한다 (a102 가 배운 것).

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/enginelock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojection"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojectionrpc"
	"github.com/JungHoonGhae/tossinvest-cli/internal/testenv"
)

var a108Now = time.Date(2026, 8, 13, 23, 35, 19, 0, time.UTC)

// a108Incident 는 사고 당일 엔진이 죽은 그 문자열이다.
var a108Incident = errors.New("strategy projection runtime: stale endpoint is incomplete")

// errA108LoopsReached 는 「부팅이 7단계 루프까지 갔다」는 표식이다.
//
// 스텁 runtime 의 Recover 가 돌려주므로 `Runtime.Run` 이 루프를 하나도 시작하지 않고
// 이 오류를 그대로 돌려준다 (runtime.go:294). 즉 이 오류가 나왔다는 것은 부팅이
// **projection 단계를 지나 rt.Run 까지 도달했다**는 뜻이고, 그것이 재려는 것이다.
var errA108LoopsReached = errors.New("a108: 부팅이 루프까지 도달했다")

// a108EngineDir 는 엔진 디렉터리 하나를 만든다.
//
// ⛔ `t.TempDir()` 이 아니라 짧은 경로다. 이 부팅은 unix socket 을 세 개 연다
// (policy runtime · alert control · projection) 이고 sun_path 에는 길이 상한이 있다.
func a108EngineDir(t *testing.T) string {
	t.Helper()
	// 실계좌 키·XDG 경로 격리. 반환값(격리된 config dir)은 쓰지 않는다 —
	// 소켓 길이 때문에 짧은 /tmp 경로를 --config-dir 로 넘긴다.
	testenv.Isolate(t)
	dir, err := os.MkdirTemp("/tmp", "a108-engine-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	return dir
}

// a108Engine 은 부팅이 7단계까지 가는 데 필요한 최소 engine.Context 다.
//
// 원장은 **진짜**다. 강등이 남기는 알림이 원장 행이므로, 가짜 handle 로는 이 change 가
// 약속한 「1회 유실형이 아니다」를 잴 수 없다.
func a108Engine(t *testing.T, dir string) (*engine.Context, *journal.Journal) {
	t.Helper()
	j, err := journal.Open(context.Background(), journal.Options{
		Path:     filepath.Join(dir, journal.DBFileName),
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
	})
	if err != nil {
		t.Fatalf("journal.Open: %v", err)
	}
	// 닫는 것은 runEngineRun 의 `defer ectx.Close()` 다. 명령이 거절로 일찍 돌아오는
	// 경로를 위해 한 번 더 닫아 두되 오류는 무시한다 (두 번째 Close 는 무해하다).
	t.Cleanup(func() { _ = j.Close() })

	gate := execgw.NewEntryGate(clock.NewFake(a108Now), map[execgw.RequiredQuery]time.Duration{})
	notifier := &obs.Notifier{
		Log:        obs.NewLogger(obs.LogOptions{Writer: new(bytes.Buffer), JSON: true, Clock: clock.System()}),
		Journal:    j,
		Gate:       gate,
		Clock:      clock.System(),
		Attempts:   1,
		RetryDelay: time.Millisecond,
	}
	return &engine.Context{
		Journal: j, Notifier: notifier, Entry: gate,
		Automation: engine.AutomationStatus{Enabled: true, Verified: true, AccountRef: "123-45"},
	}, j
}

// a108StubProjection 은 projection endpoint 의 기동 결과를 정한다.
func a108StubProjection(t *testing.T, err error) func() int {
	t.Helper()
	previous := engineStrategyProjectionStart
	calls := 0
	engineStrategyProjectionStart = func(string, strategyprojection.Reader) (*strategyprojectionrpc.Server, error) {
		calls++
		return nil, err
	}
	t.Cleanup(func() { engineStrategyProjectionStart = previous })
	return func() int { return calls }
}

// a108Loops 는 「루프까지 왔다」만 보고하고 즉시 돌아오는 runtime 이다.
//
// during 은 Recover 안에서 돈다 — 부팅이 **끝난 뒤이면서 명령이 돌아오기 전**인
// 유일한 순간이고, 원장·마커·flock 을 살아 있는 상태로 볼 수 있는 자리다.
func a108Loops(t *testing.T, during func()) *engine.Runtime {
	t.Helper()
	rt, err := engine.NewRuntime(engine.RuntimeOptions{
		Loops: []engine.SupervisedLoop{{
			Name: "a108-probe",
			Run:  func(ctx context.Context) error { return ctx.Err() },
		}},
		Recover: func(context.Context) error {
			if during != nil {
				during()
			}
			return errA108LoopsReached
		},
	})
	if err != nil {
		t.Fatalf("engine.NewRuntime: %v", err)
	}
	return rt
}

// --- 3.1 강등 -------------------------------------------------------------------

// TestAFailedStrategyProjectionDoesNotStopTheEngine 은 사고 그 자체다.
//
// 죽은 줄(engine.go B17)에 사고 당일의 오류를 그대로 넣는다. 부팅은 루프까지 가야 한다.
func TestAFailedStrategyProjectionDoesNotStopTheEngine(t *testing.T) {
	dir := a108EngineDir(t)
	ectx, _ := a108Engine(t, dir)
	stubAssembly(t, ectx, nil)
	a108StubProjection(t, a108Incident)
	stubRuntimeWithReady(t, func(*engine.Context, func()) (*engine.Runtime, error) {
		return a108Loops(t, nil), nil
	})

	_, errOut, err := runCLI(t, "--config-dir", dir, "engine", "run")
	if !errors.Is(err, errA108LoopsReached) {
		t.Fatalf("engine run = %v, want the loops to have started — "+
			"조회 전용 endpoint 하나가 손절을 든 프로세스를 죽였다", err)
	}
	if !strings.Contains(errOut, a108Incident.Error()) {
		t.Errorf("강등이 사유를 말하지 않았다; 운영자는 화면이 왜 비었는지 알 수 없다:\n%s", errOut)
	}
}

// TestTheDegradedBootWritesNoUndeliveredOutboxRow 은 D3-2 의 첫 핀이다.
//
// # 이 테스트는 원래 정반대를 요구했다
//
// 원 D3 은 강등이 **durable critical outbox 행**을 남길 것을 요구했고
// (`TestTheDegradedBootLeavesADurableCriticalAlert`), A2 적대 리뷰가 그것을 뒤집었다:
// `Journal.UndeliveredCount` 는 Type 무필터로 PENDING 행을 세고
// (`internal/journal/outbox.go:532-540`), 다음 부팅의 `restoreAlertEntryLatch` 가
// 그 수가 0 보다 크면 진입 게이트를 `ReasonAlertUndelivered` 로 latch 한다
// (`internal/app/engine/gateway.go:153-168`). 해제는 운영자 ack 뿐이다.
// publisher 가 설정되지 않은 배포에서 그 행은 영원히 PENDING 이므로,
// **화면 하나를 잃었다는 보고가 실계좌의 신규 진입을 영구 차단**한다.
// obs 교리가 정확히 그것을 금지한다(`internal/obs/event.go:266-278`).
//
// # 이 harness 가 이 계약에 맞는 이유
//
// `a108Engine` 의 Notifier 는 **Publisher 가 nil 이고 전달 루프도 없다.** 그러니
// 어떤 행이든 enqueue 되면 PENDING 으로 **남는다** — 0 이 나왔다면 그것은 「이미
// 전달돼서」가 아니라 「애초에 쓰지 않아서」다. (A2 F2 가 지적한 함정의 반대편:
// 전달 루프 없는 harness 에서 PENDING 을 보고 「미해소 유지」라 부른 것이 원죄였다.)
func TestTheDegradedBootWritesNoUndeliveredOutboxRow(t *testing.T) {
	dir := a108EngineDir(t)
	ectx, j := a108Engine(t, dir)
	stubAssembly(t, ectx, nil)
	a108StubProjection(t, a108Incident)

	var undelivered int
	var countErr error
	var blocks map[execgw.ReasonCode]string
	stubRuntimeWithReady(t, func(*engine.Context, func()) (*engine.Runtime, error) {
		return a108Loops(t, func() {
			undelivered, countErr = j.UndeliveredCount(context.Background())
			blocks = ectx.Entry.Blocks()
		}), nil
	})

	if _, errOut, err := runCLI(t, "--config-dir", dir, "engine", "run"); !errors.Is(err, errA108LoopsReached) {
		t.Fatalf("engine run = %v\n%s", err, errOut)
	}
	if countErr != nil {
		t.Fatalf("UndeliveredCount: %v", countErr)
	}
	if undelivered != 0 {
		t.Errorf("강등 기동이 미전달 outbox 행 %d개를 남겼다, want 0 — "+
			"다음 부팅의 restoreAlertEntryLatch 가 그것으로 진입을 잠근다", undelivered)
	}
	if detail, latched := blocks[execgw.ReasonAlertUndelivered]; latched {
		t.Errorf("강등이 이 프로세스의 진입 게이트를 잠갔다 (%q) — "+
			"보고가 critical rail 을 탔다는 뜻이다", detail)
	}
}

// TestASecondDegradedBootLeavesTheNextBootsEntryGateUnlatched 는 D3-2 의 둘째 핀이고,
// A2 가 이름을 붙여 요구한 **다음-부팅 측정**이다.
//
// 사고 당일의 실제 모양은 한 번의 강등이 아니라 autostart 가 1분마다 세우는 **반복**
// 기동이었다. 그래서 재는 것도 반복이다: 같은 journal 로 두 번 강등 기동한 뒤,
// 세 번째 부팅이 읽을 원장 상태가 여전히 「잠글 것이 없다」여야 한다.
//
// # 무엇을 재고 무엇을 재지 않는지 (정직하게)
//
// `restoreAlertEntryLatch` 자체는 `internal/app/engine` 소유이고 그 동작
// (0보다 크면 latch, 0이면 통과)은 a098 테스트가 진다. 여기서 재는 것은 **그 함수에
// 들어가는 값**이다 — 그 함수의 입력은 `Journal.UndeliveredCount` 하나뿐이므로
// (gateway.go:154-160), 그 값이 0 이면 다음 부팅은 이 사유로 잠길 수 없다.
// 원장은 **다시 열어서** 읽는다: 그것이 다음 부팅이 실제로 하는 일이다.
func TestASecondDegradedBootLeavesTheNextBootsEntryGateUnlatched(t *testing.T) {
	dir := a108EngineDir(t)

	degradedBoot := func() {
		t.Helper()
		ectx, _ := a108Engine(t, dir)
		stubAssembly(t, ectx, nil)
		a108StubProjection(t, a108Incident)
		stubRuntimeWithReady(t, func(*engine.Context, func()) (*engine.Runtime, error) {
			return a108Loops(t, nil), nil
		})
		if _, errOut, err := runCLI(t, "--config-dir", dir, "engine", "run"); !errors.Is(err, errA108LoopsReached) {
			t.Fatalf("engine run = %v\n%s", err, errOut)
		}
	}
	degradedBoot()
	degradedBoot()

	// 다음 부팅의 2단계: 같은 경로의 원장을 연다.
	next, err := journal.Open(context.Background(), journal.Options{
		Path:     filepath.Join(dir, journal.DBFileName),
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
	})
	if err != nil {
		t.Fatalf("다음 부팅의 journal.Open: %v", err)
	}
	defer func() { _ = next.Close() }()

	undelivered, err := next.UndeliveredCount(context.Background())
	if err != nil {
		t.Fatalf("UndeliveredCount: %v", err)
	}
	if undelivered != 0 {
		t.Fatalf("강등 기동 두 번 뒤 미전달 알림 %d건 — 다음 부팅의 진입 게이트가 "+
			"%s 로 잠긴다(해제는 운영자 ack 뿐)", undelivered, execgw.ReasonAlertUndelivered)
	}
	// 대조: 이 harness 에서 outbox 행은 실제로 남는다. 위 0 이 「원장이 안 세는 것」이
	// 아니라 「강등이 안 쓰는 것」임을 같은 handle 로 보인다.
	if _, err := next.EnqueueAlert(context.Background(), journal.Alert{
		EventKey: "a108-control", Type: "a108.control", Severity: string(obs.SeverityCritical),
		Title: "CONTROL", Body: "이 행은 이 테스트가 직접 넣었다",
	}); err != nil {
		t.Fatalf("대조군 EnqueueAlert: %v", err)
	}
	if got, err := next.UndeliveredCount(context.Background()); err != nil || got != 1 {
		t.Fatalf("대조군 뒤 미전달 = %d (err=%v), want 1 — "+
			"이 harness 는 PENDING 행을 실제로 센다", got, err)
	}
}

// TestASucceedingProjectionIsStillServedAndClosed 는 대조군이다.
//
// 대조군이 없으면 「항상 강등한다」는 뮤테이션이 위 두 테스트를 통과한다. 여기서는
// **진짜** Start 를 쓴다 — descriptor 가 도는 동안 있고 끝나면 없어야 한다.
func TestASucceedingProjectionIsStillServedAndClosed(t *testing.T) {
	dir := a108EngineDir(t)
	ectx, j := a108Engine(t, dir)
	stubAssembly(t, ectx, nil)

	var duringRun bool
	var undelivered int
	stubRuntimeWithReady(t, func(*engine.Context, func()) (*engine.Runtime, error) {
		return a108Loops(t, func() {
			_, statErr := os.Stat(strategyprojectionrpc.DescriptorPath(dir))
			duringRun = statErr == nil
			undelivered, _ = j.UndeliveredCount(context.Background())
		}), nil
	})

	if _, _, err := runCLI(t, "--config-dir", dir, "engine", "run"); !errors.Is(err, errA108LoopsReached) {
		t.Fatalf("engine run = %v", err)
	}
	if !duringRun {
		t.Error("정상 경로에서 projection descriptor 가 없다 — 강등이 성공까지 삼켰다")
	}
	if undelivered != 0 {
		t.Errorf("정상 기동이 미전달 알림 %d건을 남겼다", undelivered)
	}
	if _, err := os.Stat(strategyprojectionrpc.ControlDirectory(dir)); !os.IsNotExist(err) {
		t.Errorf("명령이 돌아온 뒤에도 control 디렉터리가 남았다 (stat err = %v) — "+
			"defer Close 가 강등 처리에서 떨어져 나갔다", err)
	}
}

// --- 3.2 안전 핀 ----------------------------------------------------------------

// TestTheDegradedBootStillHoldsTheJournalFlock 은 3.2 ① 이다.
//
// design D3 의 이중 writer 논증은 「싱글턴 권위는 부팅 1단계의 journal flock 이다」에
// 전적으로 기댄다. 강등이 그 권위를 건드리지 않았음을 **강등된 부팅이 도는 동안**
// 둘째 엔진을 거절시켜서 본다.
func TestTheDegradedBootStillHoldsTheJournalFlock(t *testing.T) {
	dir := a108EngineDir(t)
	ectx, _ := a108Engine(t, dir)
	stubAssembly(t, ectx, nil)
	a108StubProjection(t, a108Incident)

	var secondErr error
	var secondAcquired bool
	stubRuntimeWithReady(t, func(*engine.Context, func()) (*engine.Runtime, error) {
		return a108Loops(t, func() {
			lock, err := enginelock.Acquire(dir)
			secondErr = err
			if err == nil {
				secondAcquired = true
				lock.Release()
			}
		}), nil
	})

	if _, _, err := runCLI(t, "--config-dir", dir, "engine", "run"); !errors.Is(err, errA108LoopsReached) {
		t.Fatalf("engine run = %v", err)
	}
	if secondAcquired {
		t.Fatal("강등 기동 중에 둘째 엔진이 lock 을 얻었다 — 단일 writer 원장에 둘이 붙는다")
	}
	if !errors.Is(secondErr, enginelock.ErrAlreadyRunning) {
		t.Errorf("둘째 기동의 거절 = %v, want ErrAlreadyRunning", secondErr)
	}
}

// TestTheDegradedBootLeavesReadyToTheRuntimeSeam 은 3.2 ② 이고, **이름이 재는 것보다
// 컸던 테스트**다 (A2 F6 — 정직화).
//
// 옛 이름은 `TestTheDegradedBootPublishesReadyOnlyAfterRecovery` 였다. 그런데 이
// harness 는 `engineRuntimeFactory` 를 스텁으로 갈아끼우므로 진짜 `recoverThenReady`
// 를 실행하지 않는다 — `ready()` 를 부르는 것은 테스트 자신이다. 즉 여기서 실제로
// 측정되는 것은 두 가지뿐이다:
//
//  1. **`runEngineRun` 자신은 ready 를 발행하지 않는다** — 강등 처리가 마커를 건드리면
//     Recover 안에서 읽은 `beforeRecovery` 가 이미 ready 로 나온다.
//  2. **강등 기동도 ready seam 을 runtime 에 넘긴다** (nil 이 아니다) — 강등이 seam 을
//     끊으면 진짜 runtime 은 영원히 ready 를 못 쓴다.
//
// 「Recover 가 끝난 뒤에만 발행된다」는 순서 자체는 `recoverThenReady` 의 성질이고
// `TestRecoveryPublishesReadyOnlyWhenItFinished` (a102_engine_ready_test.go:94) 와
// 그 이웃들이 소유한다. 이 파일은 그것을 **다시 증명하지 않는다.**
func TestTheDegradedBootLeavesReadyToTheRuntimeSeam(t *testing.T) {
	dir := a108EngineDir(t)
	ectx, _ := a108Engine(t, dir)
	stubAssembly(t, ectx, nil)
	a108StubProjection(t, a108Incident)

	markerPath := enginelock.MarkerPath(dir)
	var readyWasNil bool
	var beforeRecovery, afterRecovery enginelock.Status
	stubRuntimeWithReady(t, func(_ *engine.Context, ready func()) (*engine.Runtime, error) {
		readyWasNil = ready == nil
		return a108Loops(t, func() {
			beforeRecovery = enginelock.Read(markerPath, time.Now())
			if ready != nil {
				ready()
			}
			afterRecovery = enginelock.Read(markerPath, time.Now())
		}), nil
	})

	if _, _, err := runCLI(t, "--config-dir", dir, "engine", "run"); !errors.Is(err, errA108LoopsReached) {
		t.Fatalf("engine run = %v", err)
	}
	if readyWasNil {
		t.Fatal("강등 기동이 ready seam 없이 runtime 을 세웠다")
	}
	if !beforeRecovery.Running {
		t.Error("강등 기동이 활성 마커를 안 남겼다 — 콘솔의 엔진 상태가 빈다")
	}
	if beforeRecovery.Ready() {
		t.Error("runEngineRun 이 스스로 ready 를 발행했다 — seam 을 우회한 경로가 생겼다")
	}
	if !afterRecovery.Ready() {
		t.Fatal("테스트가 부른 seam 이 마커에 닿지 않았다 — 강등이 seam 을 끊었다")
	}
}

// TestTheGateIsRefusedBeforeTheProjectionEndpointIsTouched 는 3.2 ③ 이다.
//
// 게이트 OFF 와 인터록 미충족은 projection **보다 먼저** 거절되어야 한다. 강등 코드가
// 그 앞으로 새면 게이트가 꺼진 계좌에서 socket 이 생긴다.
func TestTheGateIsRefusedBeforeTheProjectionEndpointIsTouched(t *testing.T) {
	for name, tc := range map[string]struct {
		ectx *engine.Context
		err  error
		want error
	}{
		"게이트 OFF": {ectx: gateOffEngine(), want: errEngineGateOff},
		"인터록 미충족": {err: engine.ErrAutomationGateRefused, want: errEngineInterlockUnmet},
	} {
		t.Run(name, func(t *testing.T) {
			dir := a108EngineDir(t)
			stubAssembly(t, tc.ectx, tc.err)
			projectionCalls := a108StubProjection(t, a108Incident)
			stubRuntimeWithReady(t, func(*engine.Context, func()) (*engine.Runtime, error) {
				return a108Loops(t, nil), nil
			})

			if _, _, err := runCLI(t, "--config-dir", dir, "engine", "run"); !errors.Is(err, tc.want) {
				t.Fatalf("engine run = %v, want %v", err, tc.want)
			}
			if projectionCalls() != 0 {
				t.Errorf("거절된 기동이 projection endpoint 를 %d번 건드렸다 — 평가 순서가 바뀌었다",
					projectionCalls())
			}
			if _, err := os.Stat(strategyprojectionrpc.ControlDirectory(dir)); !os.IsNotExist(err) {
				t.Errorf("거절된 기동이 control 디렉터리를 남겼다 (stat err = %v)", err)
			}
		})
	}
}

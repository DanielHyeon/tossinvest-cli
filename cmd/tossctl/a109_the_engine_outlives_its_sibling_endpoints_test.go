//go:build unix

package main

// a109_the_engine_outlives_its_sibling_endpoints_test.go — T2 §2.1·§2.2.
//
// a108 은 사고가 난 endpoint 하나(strategy projection)만 고쳤다. A1 적대 리뷰가 나머지
// 셋을 실측한 결과 **같은 병이 표면만 다르게** 있었다: position policy command ·
// position policy runtime · alert control 의 기동 실패는 오늘도 `runEngineRun` 에서
// `return err`(engine.go:274 · :279 · :315) 로 엔진을 죽인다. autostart 앞에서 그것은
// 영구 기동 루프가 되고, 결과는 2026-08-13 사고 그대로다 — **장중 손절 없음.**
//
// 이 파일이 고정하는 성질은 a108 과 같다: **표면을 잃는 것이 판정을 잃는 이유가 되면
// 안 된다.** 다만 형제들은 조회 전용이 아니므로 잃는 것을 정직하게 적는다(design D3):
// alert control 은 운영자 ack 표면, policy command 는 Preview/Apply 와 **격리 해제**
// 표면이다. 격리 해제를 잃으면 격리된 포지션의 무보호가 유지된다 — 그래도 fatal(전
// 포지션 무보호)보다 엄격히 낫다는 비교로만 이 강등이 성립한다.
//
// # 왜 실패를 seam 으로 주입하는가
//
// 잔재 디스크 상태로 실패를 만들면 이 테스트가 `internal/positionpolicyrpc` ·
// `internal/app/engine` 의 회수 규칙에 묶인다 — 그 규칙은 T1 이 **지금 바꾸고 있는
// 것**이다. 재는 것은 "잔재가 어떻게 생겼는가"가 아니라 "실패했을 때 엔진이 죽는가"다.
//
// harness(엔진 디렉터리·최소 Context·루프 도달 표식)는 a108 파일의 것을 그대로 쓴다.
// 두 벌을 두면 두 벌이 갈라진다.

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/enginelock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicyrpc"
)

// a109Incident 는 형제 endpoint 가 죽는 실제 모양이다 — A1 이 3/3 재현한 pre-chmod
// socket 잔재의 거부 문구를 그대로 쓴다.
var a109Incident = errors.New("position policy runtime: stale endpoint is not an exact 0600 Unix socket")

// a109Sibling 은 강등을 재는 데 필요한 endpoint 하나의 좌표다.
type a109Sibling struct {
	// stub 은 그 endpoint 의 기동만 실패시킨다. 나머지 둘은 진짜로 선다.
	stub func(t *testing.T, err error) func() int
	// control 은 강등 보고가 반드시 말해야 하는 디렉터리다. 운영자가 "어느 표면이
	// 없는가"를 여기서 읽는다.
	control func(engineDir string) string
}

func a109StubPolicyCommandStart(t *testing.T, err error) func() int {
	t.Helper()
	previous := enginePositionPolicyCommandStart
	calls := 0
	enginePositionPolicyCommandStart = func(string, *engine.PositionPolicyCommandService) (
		*engine.PositionPolicyCommandServer, error) {
		calls++
		return nil, err
	}
	t.Cleanup(func() { enginePositionPolicyCommandStart = previous })
	return func() int { return calls }
}

func a109StubPolicyRuntimeStart(t *testing.T, err error) func() int {
	t.Helper()
	previous := enginePositionPolicyRuntimeStart
	calls := 0
	enginePositionPolicyRuntimeStart = func(string, positionpolicy.RuntimeReader) (
		*engine.PositionPolicyRuntimeServer, error) {
		calls++
		return nil, err
	}
	t.Cleanup(func() { enginePositionPolicyRuntimeStart = previous })
	return func() int { return calls }
}

func a109StubAlertControlStart(t *testing.T, err error) func() int {
	t.Helper()
	previous := engineAlertControlStart
	calls := 0
	engineAlertControlStart = func(string, *engine.AlertOperations) (*engine.AlertControlServer, error) {
		calls++
		return nil, err
	}
	t.Cleanup(func() { engineAlertControlStart = previous })
	return func() int { return calls }
}

// a109Siblings 는 design D3 표의 세 endpoint다. map 이 아니라 slice 인 이유는 순서가
// 곧 부팅 순서이고, 실패 보고를 읽는 사람에게 그 순서가 의미를 갖기 때문이다.
func a109Siblings() []struct {
	name string
	a109Sibling
} {
	return []struct {
		name string
		a109Sibling
	}{
		{"position policy command", a109Sibling{
			stub: a109StubPolicyCommandStart, control: positionpolicyrpc.ControlDirectory,
		}},
		{"position policy runtime", a109Sibling{
			stub: a109StubPolicyRuntimeStart, control: positionpolicyrpc.RuntimeControlDirectory,
		}},
		{"alert control", a109Sibling{
			stub: a109StubAlertControlStart, control: engine.AlertControlDirectory,
		}},
	}
}

// a109OtherControls 는 실패하지 않은 나머지 두 endpoint 의 control 디렉터리다.
func a109OtherControls(dir, mine string) []string {
	var others []string
	for _, control := range []func(string) string{
		positionpolicyrpc.ControlDirectory, positionpolicyrpc.RuntimeControlDirectory,
		engine.AlertControlDirectory,
	} {
		if path := control(dir); path != mine {
			others = append(others, path)
		}
	}
	return others
}

// --- §2.1/§2.2 강등 ---------------------------------------------------------------

// TestAFailedSiblingEndpointDoesNotStopTheEngine 은 사고의 일반형이다.
//
// 세 endpoint 각각을 따로 죽인다. 셋을 한 번에 죽이면 "어느 하나라도 죽으면 산다"가
// 아니라 "전부 죽으면 산다"만 재게 되고, 그 사이의 어떤 경로가 아직 fatal 인지 모른다.
func TestAFailedSiblingEndpointDoesNotStopTheEngine(t *testing.T) {
	for _, sibling := range a109Siblings() {
		t.Run(sibling.name, func(t *testing.T) {
			dir := a108EngineDir(t)
			ectx, _ := a108Engine(t, dir)
			stubAssembly(t, ectx, nil)
			calls := sibling.stub(t, a109Incident)
			stubRuntimeWithReady(t, func(*engine.Context, func()) (*engine.Runtime, error) {
				return a108Loops(t, nil), nil
			})

			_, errOut, err := runCLI(t, "--config-dir", dir, "engine", "run")
			if !errors.Is(err, errA108LoopsReached) {
				t.Fatalf("engine run = %v, want the loops to have started — "+
					"%s endpoint 하나가 손절을 든 프로세스를 죽였다\n%s", err, sibling.name, errOut)
			}
			if calls() != 1 {
				t.Errorf("%s 기동 호출 %d회, want 1", sibling.name, calls())
			}
			if !strings.Contains(errOut, a109Incident.Error()) {
				t.Errorf("강등이 사유를 말하지 않았다; 운영자는 무엇을 제거해야 하는지 알 수 없다:\n%s",
					errOut)
			}
		})
	}
}

// TestADegradedSiblingSaysWhichSurfaceIsMissing 은 D3a 일반화의 핵심 요구다.
//
// 강등이 셋으로 늘면 "강등했다"만으로는 쓸모가 없다 — 운영자가 알아야 하는 것은
// **어느 표면이 없는가**이고, 그것을 말하지 않는 보고는 재시작밖에 안내하지 못한다.
// 그래서 자기 control 디렉터리를 말하고 남의 것은 말하지 않는지를 함께 잰다: 앞만
// 재면 "세 경로가 전부 projection 이라고 찍는다"가 통과한다.
func TestADegradedSiblingSaysWhichSurfaceIsMissing(t *testing.T) {
	for _, sibling := range a109Siblings() {
		t.Run(sibling.name, func(t *testing.T) {
			dir := a108EngineDir(t)
			ectx, _ := a108Engine(t, dir)
			stubAssembly(t, ectx, nil)
			sibling.stub(t, a109Incident)
			stubRuntimeWithReady(t, func(*engine.Context, func()) (*engine.Runtime, error) {
				return a108Loops(t, nil), nil
			})

			_, errOut, err := runCLI(t, "--config-dir", dir, "engine", "run")
			if !errors.Is(err, errA108LoopsReached) {
				t.Fatalf("engine run = %v\n%s", err, errOut)
			}
			mine := sibling.control(dir)
			if !strings.Contains(errOut, mine) {
				t.Errorf("강등 보고가 자기 표면(%s)을 말하지 않았다:\n%s", mine, errOut)
			}
			for _, other := range a109OtherControls(dir, mine) {
				if strings.Contains(errOut, other) {
					t.Errorf("강등 보고가 멀쩡한 표면(%s)까지 없다고 말했다 — "+
						"운영자가 지울 대상을 잘못 찾는다:\n%s", other, errOut)
				}
			}
		})
	}
}

// TestADegradedSiblingBootWritesNoUndeliveredOutboxRow 는 a108 D3-2 금지 3종을
// 형제들에게 이식한 것이다.
//
// 이 harness 의 Notifier 는 Publisher 가 nil 이고 전달 루프도 없다 — 어떤 행이든
// enqueue 되면 PENDING 으로 **남는다.** 0 이 나왔다면 "이미 전달돼서"가 아니라
// "애초에 쓰지 않아서"다. 미전달 PENDING 행은 다음 부팅의 진입 게이트를 잠근다.
func TestADegradedSiblingBootWritesNoUndeliveredOutboxRow(t *testing.T) {
	for _, sibling := range a109Siblings() {
		t.Run(sibling.name, func(t *testing.T) {
			dir := a108EngineDir(t)
			ectx, j := a108Engine(t, dir)
			stubAssembly(t, ectx, nil)
			sibling.stub(t, a109Incident)

			var undelivered int
			var countErr error
			var blocks map[execgw.ReasonCode]string
			stubRuntimeWithReady(t, func(*engine.Context, func()) (*engine.Runtime, error) {
				return a108Loops(t, func() {
					undelivered, countErr = j.UndeliveredCount(context.Background())
					blocks = ectx.Entry.Blocks()
				}), nil
			})

			if _, errOut, err := runCLI(t, "--config-dir", dir, "engine", "run"); !errors.Is(err,
				errA108LoopsReached) {
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
		})
	}
}

// TestEveryEndpointCanFailAtOnceAndTheLoopsStillRun 은 최악의 부팅이다.
//
// 잔재는 한 디렉터리에만 생기지 않는다 — 호스트가 꺼지면 넷이 동시에 남는다. 하나씩
// 강등하는 코드는 "첫 강등 뒤에 두 번째 강등이 도달하는가"를 재지 않으면 반쪽이다.
func TestEveryEndpointCanFailAtOnceAndTheLoopsStillRun(t *testing.T) {
	dir := a108EngineDir(t)
	ectx, j := a108Engine(t, dir)
	stubAssembly(t, ectx, nil)
	a109StubPolicyCommandStart(t, a109Incident)
	a109StubPolicyRuntimeStart(t, a109Incident)
	a109StubAlertControlStart(t, a109Incident)
	a108StubProjection(t, a108Incident)

	var undelivered int
	stubRuntimeWithReady(t, func(*engine.Context, func()) (*engine.Runtime, error) {
		return a108Loops(t, func() {
			undelivered, _ = j.UndeliveredCount(context.Background())
		}), nil
	})

	_, errOut, err := runCLI(t, "--config-dir", dir, "engine", "run")
	if !errors.Is(err, errA108LoopsReached) {
		t.Fatalf("engine run = %v, want the loops to have started — "+
			"네 표면이 전부 없는 부팅에서도 손절은 돌아야 한다\n%s", err, errOut)
	}
	for _, control := range []string{
		positionpolicyrpc.ControlDirectory(dir), positionpolicyrpc.RuntimeControlDirectory(dir),
		engine.AlertControlDirectory(dir),
	} {
		if !strings.Contains(errOut, control) {
			t.Errorf("네 강등 중 %s 의 보고가 없다 — 첫 강등이 나머지를 삼켰다:\n%s", control, errOut)
		}
	}
	if undelivered != 0 {
		t.Errorf("네 번의 강등이 미전달 알림 %d건을 남겼다, want 0", undelivered)
	}
}

// TestTheDegradationEventsAreNotOnTheCriticalRail 은 금지 3종의 **기계적** 핀이다.
//
// 위 테스트들은 「원장에 행이 안 생긴다」를 harness 로 관측한다. 이것은 그 결과를 만드는
// 원인 — 이벤트 타입이 obs 등급표에 없다 — 을 직접 잰다. 둘 다 필요하다: 결과만 재면
// 어느 날 등급표에 이름이 오를 때 "왜 갑자기 진입이 잠겼는가"를 아무도 설명하지 못한다.
func TestTheDegradationEventsAreNotOnTheCriticalRail(t *testing.T) {
	for _, event := range []obs.EventType{
		engineStrategyProjectionDegradedEvent, engineControlEndpointDegradedEvent,
	} {
		if got := obs.SeverityOf(event); got != obs.SeverityNormal {
			t.Errorf("%s 등급 = %q, want %q — critical 은 durable outbox 행을 만들고 "+
				"미전달 행은 다음 부팅의 진입 게이트를 잠근다", event, got, obs.SeverityNormal)
		}
		for _, critical := range obs.CriticalEvents() {
			if critical == event {
				t.Errorf("%s 가 obs 등급표에 등재됐다 — 표면 하나를 잃었다는 보고가 "+
					"실계좌의 신규 진입을 영구 차단한다", event)
			}
		}
	}
	// a108 이 세운 이벤트 타입 **값**은 바뀌지 않았다. 일반화가 기존 배포의 로그 필터를
	// 조용히 무효로 만들면 그것은 고친 것이 아니라 옮긴 것이다.
	if engineStrategyProjectionDegradedEvent != "engine.strategy_projection_unavailable" {
		t.Errorf("a108 이벤트 타입 값이 바뀌었다: %q", engineStrategyProjectionDegradedEvent)
	}
	if engineControlEndpointDegradedEvent == engineStrategyProjectionDegradedEvent {
		t.Error("형제 강등이 projection 과 같은 이벤트 타입을 쓴다 — 로그에서 종류를 가를 수 없다")
	}
}

// TestEveryDegradedEndpointNamesWhatItLost 는 D3 표의 정직성 요구다.
//
// 강등 안내가 「아무것도 느슨해지지 않는다」로 적히면 그것은 거짓이다. policy command 는
// **격리 해제**를 잃고 그것은 격리된 포지션의 무보호 유지이며, alert control 은 ack 를
// 잃고 그것은 latch 유지다. 잃는 것을 적지 않는 보고는 운영자에게 우선순위를 못 준다.
func TestEveryDegradedEndpointNamesWhatItLost(t *testing.T) {
	dir := "/tmp/a109-endpoint-copy"
	for _, endpoint := range []engineEndpoint{
		engineProjectionEndpoint(dir), enginePolicyCommandEndpoint(dir),
		enginePolicyRuntimeEndpoint(dir), engineAlertControlEndpoint(dir),
	} {
		if strings.TrimSpace(endpoint.surface) == "" || strings.TrimSpace(endpoint.lost) == "" ||
			strings.TrimSpace(endpoint.control) == "" || strings.TrimSpace(endpoint.title) == "" ||
			strings.TrimSpace(string(endpoint.event)) == "" {
			t.Errorf("endpoint 좌표가 비었다: %+v", endpoint)
		}
		if !strings.HasPrefix(endpoint.control, dir) {
			t.Errorf("%s 의 control 이 엔진 디렉터리 밖이다: %s", endpoint.surface, endpoint.control)
		}
	}
	if lost := enginePolicyCommandEndpoint(dir).lost; !strings.Contains(lost, "격리") {
		t.Errorf("policy command 강등 안내가 격리 해제를 말하지 않는다 (%q) — "+
			"격리된 포지션의 무보호가 유지된다는 사실이 사라지면 강등 논증이 거짓이 된다", lost)
	}
}

// TestTheDegradedSiblingBootStillHoldsTheJournalFlock 은 강등 논증의 전제를 잰다.
//
// D3 의 "엔진 싱글턴은 journal flock 이 강제한다" 가 참이 아니면 강등은 두 엔진을
// 허용하는 변경이 된다. 강등된 부팅이 도는 동안 둘째 엔진을 거절시켜서 본다.
func TestTheDegradedSiblingBootStillHoldsTheJournalFlock(t *testing.T) {
	dir := a108EngineDir(t)
	ectx, _ := a108Engine(t, dir)
	stubAssembly(t, ectx, nil)
	a109StubPolicyCommandStart(t, a109Incident)
	a109StubPolicyRuntimeStart(t, a109Incident)
	a109StubAlertControlStart(t, a109Incident)

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

	if _, errOut, err := runCLI(t, "--config-dir", dir, "engine", "run"); !errors.Is(err,
		errA108LoopsReached) {
		t.Fatalf("engine run = %v\n%s", err, errOut)
	}
	if secondAcquired {
		t.Fatal("강등 기동 중에 둘째 엔진이 lock 을 얻었다 — 단일 writer 원장에 둘이 붙는다")
	}
	if secondErr == nil {
		t.Error("둘째 기동이 거절되지 않았다")
	}
}

// TestSucceedingSiblingEndpointsAreStillServedAndClosed 는 **대조군**이다.
//
// 대조군이 없으면 "언제나 강등한다"는 뮤테이션이 위 테스트를 전부 통과한다. 여기서는
// 진짜 Start 셋을 쓴다 — 도는 동안 세 descriptor 가 있고, 끝나면 셋 다 없어야 한다.
func TestSucceedingSiblingEndpointsAreStillServedAndClosed(t *testing.T) {
	dir := a108EngineDir(t)
	ectx, j := a108Engine(t, dir)
	stubAssembly(t, ectx, nil)

	present := map[string]bool{}
	var undelivered, afterControlRow int
	var controlErr error
	stubRuntimeWithReady(t, func(*engine.Context, func()) (*engine.Runtime, error) {
		return a108Loops(t, func() {
			for _, path := range []string{
				positionpolicyrpc.DescriptorPath(dir), positionpolicyrpc.RuntimeDescriptorPath(dir),
				engine.AlertControlDescriptorPath(dir),
			} {
				_, statErr := os.Stat(path)
				present[path] = statErr == nil
			}
			undelivered, _ = j.UndeliveredCount(context.Background())
			// 원장이 실제로 세는지 **여기서** 확인한다 — 명령이 돌아오면
			// `defer ectx.Close()` 가 handle 을 닫으므로 밖에서는 물어볼 수 없다.
			// 위 0 이 "안 세는 것"이 아니라 "안 쓰는 것"임의 근거다.
			if _, err := j.EnqueueAlert(context.Background(), journal.Alert{
				EventKey: "a109-control", Type: "a109.control", Severity: "critical",
				Title: "CONTROL", Body: "이 행은 이 테스트가 직접 넣었다",
			}); err != nil {
				controlErr = err
				return
			}
			afterControlRow, controlErr = j.UndeliveredCount(context.Background())
		}), nil
	})

	if _, errOut, err := runCLI(t, "--config-dir", dir, "engine", "run"); !errors.Is(err,
		errA108LoopsReached) {
		t.Fatalf("engine run = %v\n%s", err, errOut)
	}
	for path, ok := range present {
		if !ok {
			t.Errorf("정상 경로에서 %s 가 없다 — 강등이 성공까지 삼켰다", path)
		}
	}
	if len(present) != 3 {
		t.Fatalf("대조군이 세 descriptor 를 다 보지 못했다: %v", present)
	}
	for _, control := range []string{
		positionpolicyrpc.ControlDirectory(dir), positionpolicyrpc.RuntimeControlDirectory(dir),
		engine.AlertControlDirectory(dir),
	} {
		if _, err := os.Stat(control); !os.IsNotExist(err) {
			t.Errorf("명령이 돌아온 뒤에도 %s 가 남았다 (stat err = %v) — "+
				"defer Close 가 강등 처리에서 떨어져 나갔다", control, err)
		}
	}
	if undelivered != 0 {
		t.Errorf("정상 기동이 미전달 알림 %d건을 남겼다", undelivered)
	}
	if controlErr != nil {
		t.Fatalf("대조군 원장 조작: %v", controlErr)
	}
	if afterControlRow != 1 {
		t.Fatalf("대조군 뒤 미전달 = %d, want 1 — 이 harness 는 PENDING 행을 실제로 센다",
			afterControlRow)
	}
}

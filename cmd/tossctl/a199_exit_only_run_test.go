package main

// a199_exit_only_run_test.go — `engine run` 이 청산 전용 Context 를 게이트 OFF 로 읽지 않고 런타임 조립까지 가며, 운영자에게
// 그 사실을 기동 출력 한 줄로 말하는지(a199 design §3-7 · D8). 확인 입력 없음(읽는 문장).

import (
	"errors"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/testenv"
)

func TestA199AnExitOnlyContextReachesTheRuntimeAndSaysSo(t *testing.T) {
	dir := testenv.Isolate(t)
	stubAssembly(t, &engine.Context{Automation: engine.AutomationStatus{
		Enabled: true, ExitOnly: true, AccountRef: "123-45",
	}}, nil)
	stubRuntime(t, func(*engine.Context) (*engine.Runtime, error) { return nil, errStubRuntimeReached })

	out, _, err := runCLI(t, "--config-dir", dir, "engine", "run")
	if !errors.Is(err, errStubRuntimeReached) {
		t.Fatalf("err = %v, want the runtime assembly to be reached — an exit-only start is not a gate-off start", err)
	}
	if !strings.Contains(out, "청산 전용") || !strings.Contains(out, "새 진입") {
		t.Errorf("stdout does not tell the operator the engine is exit-only:\n%s", out)
	}
	// 구현 리뷰 r1 처분 7(b): 복구 순서(검증 재실행 → attestation 재발급 → 재시작)를 그 순서로 말함.
	verify, reissue, restart := strings.Index(out, "tossctl verify run"), strings.Index(out, "tossctl soak attest"), strings.Index(out, "재시작")
	if verify < 0 || reissue < 0 || restart < 0 || verify >= reissue || reissue >= restart {
		t.Errorf("stdout does not give the recovery order verify run → soak attest → restart:\n%s", out)
	}
}

func TestA199AGateOnUnverifiedContextThatIsNotExitOnlyStillRefuses(t *testing.T) {
	dir := testenv.Isolate(t)
	stubAssembly(t, &engine.Context{Automation: engine.AutomationStatus{Enabled: true, AccountRef: "123-45"}}, nil)
	stubRuntime(t, func(*engine.Context) (*engine.Runtime, error) { return nil, errStubRuntimeReached })

	if _, _, err := runCLI(t, "--config-dir", dir, "engine", "run"); !errors.Is(err, errEngineGateOff) {
		t.Fatalf("err = %v, want errEngineGateOff — neither verified nor exit-only has no loop set", err)
	}
}

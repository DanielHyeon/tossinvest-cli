package console

// a112 L5 — 운영자가 활성화 매니페스트에 옮겨 적을 숫자를 화면에서 읽는다 (결정 54).
//
// 이 값은 시장별 사실이 아니라 프로세스 하나의 사실이라서 시장 카드가 아니라 상단
// contract 절에 있다. 그리고 두 시장이 다 UNKNOWN 인 상태에서도 보여야 한다 —
// 운영자가 이 숫자를 찾는 순간이 바로 그 상태다.

import (
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojection"
)

func TestStrategyRuntimePageShowsTheDigestsTheOperatorMustWriteDown(t *testing.T) {
	config, build := consoleRuntimeDigest("config"), consoleRuntimeDigest("build")
	// 일부러 dormant 쌍이다: 아무것도 활성화되지 않은 상태에서도 읽혀야 한다.
	snapshot := strategyprojection.WithRuntimeIdentity(strategyprojection.DormantSnapshot(consoleProjectionNow), config, build)
	h := newHarness(t, func(options *Options) { options.StrategyRuntime = multiMarketRuntimeStub{snapshot: snapshot} })
	h.authenticate(t)

	page := body(t, h.get(t, "/strategy-runtime"))
	for _, want := range []string{"Config digest", "Build digest", config, build} {
		if !strings.Contains(page, want) {
			t.Errorf("전략 화면에 %q 가 없다 — 운영자가 매니페스트에 적을 값을 읽을 수 없다", want)
		}
	}
	if mutations := h.broker.mutationCount(); mutations != 0 {
		t.Fatalf("읽기 전용 화면이 broker mutation %d 건을 만들었다", mutations)
	}
}

// TestStrategyRuntimePageDoesNotInventDigestsWhenTheEngineIsAbsent 는 콘솔이 자기
// 바이너리 값으로 빈칸을 채우지 않음을 잡는다. 콘솔과 엔진의 build 가 다르면 그렇게
// 만든 숫자는 **엔진이 거절할 매니페스트**를 낳는다.
func TestStrategyRuntimePageDoesNotInventDigestsWhenTheEngineIsAbsent(t *testing.T) {
	h := newHarness(t)
	h.authenticate(t)
	page := body(t, h.get(t, "/strategy-runtime"))
	if !strings.Contains(page, "Config digest") || !strings.Contains(page, "Build digest") {
		t.Fatal("엔진 부재에도 두 항목의 자리는 있어야 한다 — 운영자가 어디를 볼지 알아야 한다")
	}
	// 페이지 전체의 not_observed 개수를 세면 다른 필드가 대신 통과시켜 준다.
	// 이 두 항목의 **렌더된 짝**을 그대로 찾는다.
	for _, want := range []string{
		"<dt>Config digest</dt><dd><code>not_observed</code></dd>",
		"<dt>Build digest</dt><dd><code>not_observed</code></dd>",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("엔진이 없는데 digest 자리가 관측 없음이 아니다: %q 를 못 찾았다", want)
		}
	}
}

// 접두사까지가 값이다 — 스케줄러가 매니페스트에서 견주는 형식 그대로.
func consoleRuntimeDigest(seed string) string {
	return "sha256:" + strings.Repeat(map[string]string{"config": "c", "build": "b"}[seed], 64)
}

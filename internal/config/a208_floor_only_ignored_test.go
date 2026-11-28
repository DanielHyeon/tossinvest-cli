package config

// a208 F3 — 토글이 참이어도 설정된 공통 정책에 편입 보호선 사본이 없으면(BALANCED · 공통 정책 없음) 로더가 무시 사유를 파생함.
// 사유는 기동 audit 행에만 쓰이고 파일에는 쓰지 않음(json:"-"). 사본 ID 는 공통 정책으로 고를 수 없음.

import (
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
)

func a208LoadEngine(t *testing.T, floorOnly bool, common string) Engine {
	t.Helper()
	exitPolicy := ""
	if common != "" {
		exitPolicy = `,"exit_policy":{"common_policy":"` + common + `"}`
	}
	flag := "false"
	if floorOnly {
		flag = "true"
	}
	svc := writeConfig(t, `{"schema_version":5,"trading":{},
	  "engine":{"adoption":{"enabled":true,"default_stop_pct":0.05,"floor_only_exit":`+flag+`}`+exitPolicy+`}}`)
	cfg, err := svc.Load(t.Context())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return cfg.Engine
}

func TestA208FloorOnlyIgnoredReasonFollowsTheCommonPolicy(t *testing.T) {
	for _, tc := range []struct {
		floorOnly bool
		common    string
		ignored   bool
	}{
		{true, exitpolicy.CommonLadderHybrid50, false},
		{true, exitpolicy.CommonLadderRunner, false},
		{true, exitpolicy.CommonLadderBalanced, true},
		{true, "", true},
		{false, exitpolicy.CommonLadderBalanced, false},
		{false, "", false},
	} {
		engine := a208LoadEngine(t, tc.floorOnly, tc.common)
		got := engine.Adoption.FloorOnlyIgnored
		if (got != "") != tc.ignored {
			t.Errorf("floor_only=%v common=%q: ignored reason = %q, want ignored=%v", tc.floorOnly, tc.common, got, tc.ignored)
		}
		if tc.ignored && !strings.Contains(got, "무시") {
			t.Errorf("common=%q: the reason must say the toggle is ignored: %q", tc.common, got)
		}
		// 무시돼도 설정 값 자체는 남음 — 콘솔 저장 이월이 사람이 쓴 참을 끄지 않게.
		if engine.Adoption.FloorOnlyExit != tc.floorOnly {
			t.Errorf("floor_only=%v common=%q: the configured toggle was rewritten to %v", tc.floorOnly, tc.common,
				engine.Adoption.FloorOnlyExit)
		}
	}
}

// 사본 ID 는 등록 공통 정책이 아니므로 설정 파일의 공통 정책으로 쓰면 거부됨(선택 가능 정책으로 노출되지 않음).
func TestA208FloorOnlyCopyIsNotAConfigurableCommonPolicy(t *testing.T) {
	engine := a208LoadEngine(t, false, exitpolicy.CommonLadderHybrid50AdoptedFloorOnly)
	if engine.ExitPolicy.Rejected == "" {
		t.Fatalf("common_policy = the floor-only copy was accepted: %+v", engine.ExitPolicy)
	}
}

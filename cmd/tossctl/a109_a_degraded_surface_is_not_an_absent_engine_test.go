//go:build unix

package main

// a109_a_degraded_surface_is_not_an_absent_engine_test.go — T2 §2.5 (design D3a-2).
//
// a108 의 강등 교리는 「전략 화면이 dormant 로 뜬다」는 세 번째 표면을 전제했다. 형제
// endpoint 에는 그 표면이 없고, 대신 기존 소비자 메시지가 강등을 **엔진 부재로 단정**
// 했다: 「이 디렉터리로 도는 엔진이 없다, engine run 이 살아 있어야 한다」.
//
// a109 이후 그 단정은 거짓일 수 있다. alert control endpoint 기동이 실패하면 엔진은 그
// 표면 없이 강등 부팅하고 보호 루프는 그대로 돈다. 운영자는 저 문장을 따라 멀쩡히 도는
// 엔진을 재시작하고, 강등 원인은 결정적이므로 같은 상태가 재현된다 — 안내가 운영자를
// 무한 재시작에 넣는다.
//
// 이 change 가 여기서 바꾼 것은 문구뿐이므로, 문구를 안 재면 이 task 는 아무것도
// 재지 않는다.

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
)

// TestTheAlertsCLIDoesNotAssertTheEngineIsAbsent 는 문구 핀이다.
func TestTheAlertsCLIDoesNotAssertTheEngineIsAbsent(t *testing.T) {
	message := errEngineAlertsUnavailable.Error()
	for _, required := range []string{
		"강등",         // 다른 가능성을 말한다
		"엔진 로그",      // 어느 쪽인지 가르는 증거의 위치를 준다
		"없다면",        // 엔진 부재는 **조건**이지 단정이 아니다
		"돌고 있다면",     // 반대쪽 갈래도 말한다
		"engine run", // 부재일 때의 행동은 a098 부터의 요구다
	} {
		if !strings.Contains(message, required) {
			t.Errorf("alerts 안내에 %q 가 없다 — 강등 가능성·확인할 곳·두 갈래의 행동 중 "+
				"하나가 빠졌다:\n%s", required, message)
		}
	}
	// 옛 **단정**이 남아 있으면 안 된다. 「엔진이 없다」로 끝나는 안내는 강등 부팅에서
	// 거짓이고, 그 거짓이 운영자를 무한 재시작에 넣는다.
	if strings.Contains(message, "엔진이 없다.") || strings.Contains(message, "엔진이 없다 (") ||
		strings.Contains(message, "엔진이 없다,") {
		t.Errorf("안내가 아직 엔진 부재를 단정한다:\n%s", message)
	}
}

// TestTheAlertsDialStillReportsTheAbsentSurfaceWithThatMessage 는 그 문구가 실제로
// 그 경로에서 나오는지 본다. 문자열만 재면 「상수는 고쳤는데 아무도 안 쓴다」가 통과한다.
func TestTheAlertsDialStillReportsTheAbsentSurfaceWithThatMessage(t *testing.T) {
	dir, err := os.MkdirTemp("/tmp", "a109-alerts-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}

	client, dialErr := dialAlertControl(context.Background(), dir)
	if client != nil {
		t.Fatal("descriptor 가 없는데 client 가 나왔다")
	}
	if !errors.Is(dialErr, errEngineAlertsUnavailable) {
		t.Fatalf("dialAlertControl = %v, want %v", dialErr, errEngineAlertsUnavailable)
	}
}

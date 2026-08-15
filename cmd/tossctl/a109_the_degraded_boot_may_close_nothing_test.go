//go:build unix

package main

// a109_the_degraded_boot_may_close_nothing_test.go — a109 §2-fix F7 (A2 P2-10).
//
// 강등 부팅의 defer 는 nil 가드를 지나고 나서 Close 를 등록한다:
//
//	if policyControl != nil { defer policyControl.Close() }
//
// 뮤테이션 원장 M7 은 그 가드를 지워도 **아무 테스트도 죽지 않는다**고 기록했고,
// 이유를 측정했다: 세 서버 타입의 `Close` 가 전부 nil 수신자에 안전하다. 즉 가드는
// 동작을 바꾸지 않는 **의도의 표현**이다.
//
// # 그러면 왜 테스트를 쓰는가
//
// 그 「측정된 이유」가 지금은 T1 표면(`internal/app/engine`)의 성질이고, 아무도 그것을
// 계약으로 적지 않았다. T1 이 Close 를 고치다가 nil 가드를 없애면 강등 부팅의 defer 는
// **패닉**한다 — 그 패닉은 손절을 든 프로세스의 종료 경로에서 일어난다. M7 이 생존한
// 이유를 여기서 계약으로 올린다: 세 Close 는 nil 수신자에 안전해야 하고, 그것이 깨지면
// 이 파일이 먼저 운다.
//
// 이 테스트는 T1 의 파일을 편집하지 않는다. 소비자 쪽에서 그 계약을 **재는** 것이고,
// 계약을 재는 자리는 그것에 기대는 쪽이다.

import (
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
)

// TestTheSiblingEndpointClosesAreSafeOnANilServer 는 그 계약이다.
func TestTheSiblingEndpointClosesAreSafeOnANilServer(t *testing.T) {
	for _, endpoint := range []struct {
		name  string
		close func() error
	}{
		{"AlertControlServer", (*engine.AlertControlServer)(nil).Close},
		{"PositionPolicyRuntimeServer", (*engine.PositionPolicyRuntimeServer)(nil).Close},
		{"PositionPolicyCommandServer", (*engine.PositionPolicyCommandServer)(nil).Close},
	} {
		t.Run(endpoint.name, func(t *testing.T) {
			defer func() {
				if recovered := recover(); recovered != nil {
					t.Fatalf("nil %s.Close 가 패닉했다 (%v) — 강등 부팅의 defer 가 "+
						"손절을 든 프로세스의 종료 경로에서 터진다", endpoint.name, recovered)
				}
			}()
			if err := endpoint.close(); err != nil {
				t.Errorf("nil %s.Close = %v, want nil — 기동조차 안 한 endpoint 의 "+
					"닫기가 오류를 만들면 강등 부팅의 종료가 그것을 보고한다", endpoint.name, err)
			}
		})
	}
}

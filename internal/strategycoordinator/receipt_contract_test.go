//go:build tossos_testseams

package strategycoordinator_test

import (
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategycoordinator"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyproposal"
)

// 큐 상한은 고른 수가 아니라 매니페스트 상한을 읽은 수다.
//
// 이 시험이 없으면 두 값은 말없이 갈라진다. 갈라지는 방향이 나쁜 쪽이면
// (큐가 더 작으면) 검증이 통과시킨 정상 매니페스트가 큐에서 넘쳐 그 시장이
// 매 주기 닫힌다. 그 고장은 주문을 잘못 내지 않으므로 어떤 안전 시험도
// 잡지 못하고, 기능만 조용히 죽는다. 실제로 앞선 판본이 64 를 들고 있었고,
// 65 종목이면 죽는 상태였다.
//
// 같음까지 요구하고 "크거나 같음"으로 느슨하게 두지 않는 이유: 큐가 더 크면
// 그 여분은 어떤 영수증도 뒷받침하지 않는 수이고, 그것을 허용하는 순간 이
// 시험은 다시 "고른 수"를 통과시킨다.
//
// 이 파일이 외부 테스트 패키지(strategycoordinator_test)인 이유: 생산 코드가
// strategyproposal 을 import 하면 journal·position 이 조정자의 의존 폐포에
// 들어와 dependency_closure_test 가 지키는 약속이 깨진다. 시험만 두 쪽을 읽는다.
func TestTheQueueCapacityIsReadFromTheManifestScopeCeiling(t *testing.T) {
	if strategycoordinator.Capacity != strategyproposal.MaxManifestScopes {
		t.Fatalf("Capacity=%d, strategyproposal.MaxManifestScopes=%d — 큐 상한이 영수증과 갈라졌다",
			strategycoordinator.Capacity, strategyproposal.MaxManifestScopes)
	}
}

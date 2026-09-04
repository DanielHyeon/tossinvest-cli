//go:build tossos_testseams

package strategyrouter

import "time"

// FamilyActivationForTest 는 명시적인 저장소 test-seam 바이너리에서만 불투명한
// 활성화를 주조한다. 생산 코드는 여전히 LoadProductionFamilyActivation 하나로만
// 얻는다 — 그 함수는 사람이 서명한 파일 없이는 아무것도 승격하지 않는다.
//
// 선례는 `internal/scheduler/activation_testseam.go` 의 ActivationForTest 다.
// 이 seam 이 태그 아래 있는 이유는 생산 빌드에 그 문이 아예 존재하지 않게
// 하려는 것이다. 무태그 `make test` 는 생산 빌드 구성이라 이 파일을 안 본다.
//
// promoted 는 이 시장에서 desired/effective 를 함께 ON 으로 둘 레인 목록이다.
// 목록에 없는 레인은 두 상태 모두 OFF 로 선다.
func FamilyActivationForTest(market Market, generation uint64, promoted map[string]bool) FamilyActivation {
	return FamilyActivationWithBindingsForTest(market, generation, promoted, 1)
}

// FamilyActivationWithBindingsForTest 는 주문 경로가 결속하는 ProtectionReady
// 하한까지 지정한다. 그 결속은 보호 세대가 **존재하는 단계**에서만 일어나므로
// (태스크 8.8.2), 그 단계의 시험은 하한을 골라야 한다.
//
// 하한이 0 이면 생산 검증이 매니페스트를 거절하므로, 이 seam 도 0 을 주면
// 영값을 돌려준다 — seam 이 생산보다 관대하면 seam 으로 지은 시험이 생산에서
// 성립하지 않는 상태를 초록으로 통과시킨다.
func FamilyActivationWithBindingsForTest(market Market, generation uint64, promoted map[string]bool,
	protectionReadyMinGeneration uint64,
) FamilyActivation {
	want := productionRouteDescriptors(market)
	if len(want) == 0 || generation == 0 || protectionReadyMinGeneration == 0 {
		return FamilyActivation{}
	}
	state := make(map[familyLaneKey]productionFamilyActivationDescriptor, len(want))
	for laneID, table := range want {
		desired, effective := StateOff, StateOff
		if promoted[laneID] {
			desired, effective = StateOn, StateOn
		}
		state[familyLaneKey{family: table.Family, laneID: laneID, laneVersion: table.LaneVersion}] =
			productionFamilyActivationDescriptor{Family: table.Family, Horizon: table.Horizon, LaneID: laneID,
				LaneVersion: table.LaneVersion, Desired: desired, Effective: effective}
	}
	return FamilyActivation{market: market, generation: generation, actor: "test-seam",
		expiresAt:                    time.Now().UTC().Add(time.Hour),
		protectionReadyMinGeneration: protectionReadyMinGeneration, state: state}
}

// AllFourFamiliesForTest 는 그 시장의 네 레인 ID 를 전부 켠 목록이다.
// 시험이 네 이름을 손으로 적으면 레인 ID 가 바뀔 때 조용히 셋만 켜진다.
func AllFourFamiliesForTest(market Market) map[string]bool {
	promoted := map[string]bool{}
	for laneID := range productionRouteDescriptors(market) {
		promoted[laneID] = true
	}
	return promoted
}

package strategyworker

import (
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyarbiter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

// a112 7.3.1 SHADOW(브리프 v3.3 §4 · §6) — OFF 레인의 읽기 전용 반사실.
//
// SHADOW 는 「이 레인이 켜져 있었다면 이 입력에서 봉투를 냈을까」 를 **값**으로만 답한다. 봉투 · 승격 · 활성화를 만들지 않고, 입력은
// opaque ShadowInput 이라 조정자에 들어가는 Input 과 섞일 수 없다(admit · Submit 은 Input 만 받는다).

// ShadowOutcome 은 반사실 판정 하나다.
type ShadowOutcome string

const (
	// ShadowWouldEmit 은 이 레인이 소유하는 봉인된 제안이 있었다는 뜻이다 — 켜져 있었다면 Run 이 봉투를 냈다.
	ShadowWouldEmit ShadowOutcome = "WOULD_EMIT"
	// ShadowNotThisLane 은 제안이 이 레인을 **주장**했지만 봉인 · 가족 유도가 성립하지 않았다는 뜻이다 — 켜져 있었어도 REFUSED.
	ShadowNotThisLane ShadowOutcome = "NOT_THIS_LANE"
	// ShadowNoInput 은 이 레인을 주장하는 제안이 이 물결에 없었다는 뜻이다.
	ShadowNoInput ShadowOutcome = "NO_INPUT"
)

// ShadowOutcomes 는 판정 어휘 전부다(projection · OpenAPI 어휘와 시험으로 대조).
func ShadowOutcomes() []ShadowOutcome {
	return []ShadowOutcome{ShadowWouldEmit, ShadowNotThisLane, ShadowNoInput}
}

// ShadowInput 은 관문 앞 제안 하나의 opaque 사본이다(필드 비공개). 생성자는 NewShadowInput 하나이고, 받는 쪽은 ShadowVerdict 하나다.
type ShadowInput struct {
	proposal strategyarbiter.Proposal
}

// NewShadowInput 은 composite literal 하나다 — 호출 0 이라 조정 루프 안의 수집 helper 에 panic 거리가 없다(브리프 §4 shape 핀).
func NewShadowInput(proposal strategyarbiter.Proposal) ShadowInput {
	return ShadowInput{proposal: proposal}
}

// ShadowEligible 은 「OFF 레인」 술어 **하나**다: 같은 물결의 활성화가 이 레인을 Desired OFF ∧ Effective OFF 로 둘 때만 shadow 가 적용된다
// (활성화 ON 레인은 shadow 가 아니다 — ON 이 우선). 투영 검증기(strategyprojection.validateLane)는 같은 조건을 값 형태로 건다.
func ShadowEligible(activation strategyrouter.FamilyActivation, worker FamilyWorker) bool {
	return worker.Desired(activation) == strategyrouter.StateOff && worker.Effective(activation) == strategyrouter.StateOff
}

// ShadowVerdict 는 입력 하나에 대한 반사실 판정이다. 활성화를 보지 않는다(OFF 레인의 「켜졌다면」 이므로) — 소유 판정은 Run 과 같은 owns 하나.
func (worker FamilyWorker) ShadowVerdict(input ShadowInput) ShadowOutcome {
	lineage := input.proposal.Result.Lineage
	if lineage.Market != worker.key.Market || lineage.LaneID != worker.key.LaneID {
		return ShadowNoInput
	}
	if !worker.owns(input.proposal) {
		return ShadowNotThisLane
	}
	return ShadowWouldEmit
}

// ShadowEligible 은 레인 위임이다 — 판정은 위의 함수 하나(레인은 worker 를 숨기므로 엔진은 이 문으로만 묻는다).
func (lane *Lane) ShadowEligible(activation strategyrouter.FamilyActivation) bool {
	return ShadowEligible(activation, lane.worker)
}

// ShadowOutcomeOver 는 한 물결의 묶음 전체에 대한 이 레인의 판정이다: WOULD_EMIT > NOT_THIS_LANE > NO_INPUT. 잠금을 보지 않는다 —
// 잠긴 레인도 관측한다(SHADOW 는 관측만, 브리프 §6). 레인 상태를 바꾸지 않는다.
func (lane *Lane) ShadowOutcomeOver(inputs []ShadowInput) ShadowOutcome {
	outcome := ShadowNoInput
	for _, input := range inputs {
		switch lane.worker.ShadowVerdict(input) {
		case ShadowWouldEmit:
			return ShadowWouldEmit
		case ShadowNotThisLane:
			outcome = ShadowNotThisLane
		}
	}
	return outcome
}

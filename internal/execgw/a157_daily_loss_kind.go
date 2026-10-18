package execgw

// a157_daily_loss_kind.go 는 「precheck 거절이 일일 손실 체인 거절인가」를 한 자리에서 가르는 판별임(a157 design D1, 사람 결정 1-1).
//
// # 왜 한 자리인가
//
// 첫 다리 precheck 의 보고 우선순위(일일 손실 > 결속 결함 > 그 밖의 진입 거절)와 첫 다리 승인(engine admit)의 일일 손실 분류가 같은 판별을
// 써야 함. 같은 조건이 두 자리에 있으면 한쪽을 지운 변이를 다른 쪽이 시험에서 가려 줌 — 그래서 판별은 여기 하나뿐.
//
// # 조건
//
// a137 의 첫 다리 승인이 쓰던 두 조건을 글자 그대로 옮김(순서도 같음 — 측정 불가 먼저):
//   - 측정 불가: 측정 불가 표지(*DailyLossUnmeasurable)가 사슬에 있고, 체인 단계 · 입력 불가 거절(*IssueRefusal)도 있음.
//   - 한도 도달: 체인 단계 · 한도 도달 거절.
// 표지는 dailyLossChainRefusal 이 daily_loss 단의 입력 불가에만 붙이므로, 표지 없는 입력 불가(호출자 값 결함 등)는 일일 손실 거절이 아님.

import (
	"errors"

	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
)

// DailyLossChainRefusal 은 precheck 거절의 일일 손실 종류임. 영값(None)은 「일일 손실 체인 거절이 아님」.
type DailyLossChainRefusal uint8

const (
	// DailyLossChainRefusalNone 은 일일 손실 체인 거절이 아님(그 밖의 거절 · nil).
	DailyLossChainRefusalNone DailyLossChainRefusal = iota
	// DailyLossChainRefusalUnmeasurable 은 그날의 손실을 잴 수 없어 체인이 입력 불가로 거절한 것(a137 측정 불가 — 강화 없음).
	DailyLossChainRefusalUnmeasurable
	// DailyLossChainRefusalLimitReached 는 체인이 일일 손실 한도 도달로 거절한 것(a137 — 운영 모드 강화 대상).
	DailyLossChainRefusalLimitReached
)

// String 은 시험 · 진단 문구용 이름.
func (k DailyLossChainRefusal) String() string {
	switch k {
	case DailyLossChainRefusalUnmeasurable:
		return "unmeasurable"
	case DailyLossChainRefusalLimitReached:
		return "limit-reached"
	default:
		return "none"
	}
}

// DailyLossChainRefusalOf 는 거절 err 가 일일 손실 체인 거절인지 · 어느 쪽인지 돌려줌. 타입 · 단계 · 사유로만 가르고 문구는 보지 않음.
func DailyLossChainRefusalOf(err error) DailyLossChainRefusal {
	// 측정 불가: 표지와 체인 · 입력 불가 거절이 함께 있어야 함 — 표지만 있거나 다른 단계면 아래 한도 판정으로 내려감.
	if unmeasurable := (*DailyLossUnmeasurable)(nil); errors.As(err, &unmeasurable) {
		if issue := (*IssueRefusal)(nil); errors.As(err, &issue) && issue.Stage == StageChain &&
			issue.Reason == string(risk.ReasonInputUnavailable) {
			return DailyLossChainRefusalUnmeasurable
		}
	}
	// 한도 도달: 체인 단계의 한도 도달 거절(표지 유무와 무관).
	if issue := (*IssueRefusal)(nil); errors.As(err, &issue) && issue.Stage == StageChain &&
		issue.Reason == string(risk.ReasonDailyLossLimitReached) {
		return DailyLossChainRefusalLimitReached
	}
	return DailyLossChainRefusalNone
}

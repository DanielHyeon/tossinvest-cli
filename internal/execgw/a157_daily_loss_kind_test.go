package execgw_test

// a157 판별 함수 표 시험 — 「일일 손실 체인 거절」의 조건은 a137 의 첫 다리 승인 조건을 글자 그대로 옮긴 것(design D1):
//   - 측정 불가: 측정 불가 표지가 사슬에 있고 **그리고** 체인 단계 · 입력 불가 거절이 있음,
//   - 한도 도달: 체인 단계 · 한도 도달 거절,
//   - 측정 불가를 먼저 봄(표지가 있어도 입력 불가가 아니면 한도 판정으로 내려감).

import (
	"errors"
	"fmt"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
)

func TestA157DailyLossChainRefusalOfTable(t *testing.T) {
	chain := func(reason risk.ReasonCode) *execgw.IssueRefusal {
		return &execgw.IssueRefusal{Stage: execgw.StageChain, Reason: string(reason), Detail: "detail"}
	}
	issuance := func(reason risk.ReasonCode) *execgw.IssueRefusal {
		return &execgw.IssueRefusal{Stage: execgw.StageIssuance, Reason: string(reason)}
	}
	marker := &execgw.DailyLossUnmeasurable{Cause: errors.New("dailyloss: today's realised loss cannot be measured")}
	for _, row := range []struct {
		name string
		err  error
		want execgw.DailyLossChainRefusal
	}{
		{"nil", nil, execgw.DailyLossChainRefusalNone},
		{"limit", chain(risk.ReasonDailyLossLimitReached), execgw.DailyLossChainRefusalLimitReached},
		{"limit wrapped", fmt.Errorf("outer: %w", chain(risk.ReasonDailyLossLimitReached)), execgw.DailyLossChainRefusalLimitReached},
		{"limit at another stage", issuance(risk.ReasonDailyLossLimitReached), execgw.DailyLossChainRefusalNone},
		{"unmeasurable", errors.Join(chain(risk.ReasonInputUnavailable), marker), execgw.DailyLossChainRefusalUnmeasurable},
		{"unmeasurable wrapped", fmt.Errorf("outer: %w", errors.Join(chain(risk.ReasonInputUnavailable), marker)), execgw.DailyLossChainRefusalUnmeasurable},
		{"input unavailable without the marker", chain(risk.ReasonInputUnavailable), execgw.DailyLossChainRefusalNone},
		{"marker without a chain refusal", marker, execgw.DailyLossChainRefusalNone},
		{"marker with input unavailable at another stage", errors.Join(issuance(risk.ReasonInputUnavailable), marker), execgw.DailyLossChainRefusalNone},
		{"marker with a limit refusal", errors.Join(chain(risk.ReasonDailyLossLimitReached), marker), execgw.DailyLossChainRefusalLimitReached},
		{"another chain reason", chain(risk.ReasonOpenExposureExceeded), execgw.DailyLossChainRefusalNone},
		{"q_final refusal", &execgw.QFinalRefusal{Code: riskbucket.RefusalExistingGuardianCap, Field: "q_existing_guardian"}, execgw.DailyLossChainRefusalNone},
		{"plain error", errors.New("plain"), execgw.DailyLossChainRefusalNone},
	} {
		if got := execgw.DailyLossChainRefusalOf(row.err); got != row.want {
			t.Errorf("%s: DailyLossChainRefusalOf = %v, want %v", row.name, got, row.want)
		}
	}
}

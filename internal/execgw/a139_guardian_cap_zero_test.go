package execgw_test

import (
	"errors"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

// a139 특성 시험(freeze 리뷰 P1-1): 건당 위험 예산이 한 단위 위험도 못 덮으면 q_final precheck 는 EXISTING_GUARDIAN_CAP 으로 거부함.
// 이 코드는 첫 다리 승인에서 범위 거절이 아니라 AUTHORITY_MISMATCH 로 분류되어 레인을 잠금(design D0 보강) — 1/3 예산에서 더 자주 나옴.
// 분류를 바꾸는 것은 a139 밖의 사람 결정이므로, 여기서는 지금의 코드를 못 박아 그 결정이 바꿀 자리를 드러냄.
func TestA139AZeroGuardianCapIsTheExistingGuardianCapRefusal(t *testing.T) {
	rig := newGuardian(t, func(o *execgw.RiskGuardianOptions) {
		o.Policy.RiskBudget = riskcalc.Money{Amount: "1", Currency: "KRW"}
	})
	_, err := rig.guardian.PrecheckQFinalEntry(qFinalKRRequest(t, rig, "a139", 20))
	var refusal *execgw.QFinalRefusal
	if !errors.As(err, &refusal) || refusal.Code != riskbucket.RefusalExistingGuardianCap {
		t.Fatalf("a risk budget that covers no unit must be the EXISTING_GUARDIAN_CAP refusal, got %v", err)
	}
}

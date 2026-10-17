package verifylive

// reconcile_g3_test.go — a121 tasks 2.2·2.2.2: design G3-1(계좌 마스크 결속)과 사람 승인 출력(F1·R2-5).
// G3-2(계좌 수·seq·기형 행)·G3-3(프로필·환경 변수·override)·G3-4(시장)는 좁은 생성자·사전 검사가 사는
// cmd/tossctl/verify_reconcile_test.go 에 있다.

import (
	"errors"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
)

// TestReconcileRefusesMixedAccountReferences — 대상을 언급한 기록 줄들의 AccountRef 가 서로 다르다.
func TestReconcileRefusesMixedAccountReferences(t *testing.T) {
	e := rcA063Entries()
	// 존속 확인 줄이 같은 artifact 를 다른 계좌 마스크로 다시 언급한다.
	e[2].AccountRef = attest.Mask("999-99-990000")
	e[2].Artifacts = []Artifact{rcTarget()}
	h := newRCHarness(t, e)
	rcRefuse(t, h, RefuseAccountMixed)
	rcRequireNoListRead(t, h)
}

// TestReconcileRefusesACurrentAccountWhoseMaskDiffers — attest.Mask(현재 참조) ≠ 기록 마스크.
func TestReconcileRefusesACurrentAccountWhoseMaskDiffers(t *testing.T) {
	h := newRCHarness(t, rcA063Entries())
	h.account.Ref = "123-45-670000"
	rcRefuse(t, h, RefuseAccountMismatch)
	rcRequireNoListRead(t, h)
}

// TestReconcileRefusesUnusableAccountReferences 는 로트 1 D3 의 attest.Mask 경계다 — 빈 참조는 "(none)", 4자 이하는
// 전부 `*` 라 대조가 길이 비교로 퇴화한다. 어느 쪽 참조든 그 모양이면 거절.
func TestReconcileRefusesUnusableAccountReferences(t *testing.T) {
	cases := map[string]struct {
		current   string
		recordRef string
	}{
		"current-empty":           {current: "", recordRef: rcMask()},
		"current-four-or-fewer":   {current: "5678", recordRef: attest.Mask("1234")}, // 둘 다 "****"
		"record-empty":            {current: rcAccount, recordRef: ""},
		"record-none-placeholder": {current: "", recordRef: "(none)"},
		"both-short-same-length":  {current: "99", recordRef: "**"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			e := rcA063Entries()
			for i := range e {
				e[i].AccountRef = c.recordRef
			}
			h := newRCHarness(t, e)
			h.account.Ref = c.current
			rcRefuse(t, h, RefuseAccountRefUnusable)
			rcRequireNoListRead(t, h)
		})
	}
}

// TestReconcileRefusesWithoutApproval — 승인이 없으면 목록을 읽지 않고 쓰지 않는다.
func TestReconcileRefusesWithoutApproval(t *testing.T) {
	h := newRCHarness(t, rcA063Entries())
	h.approveErr = errors.New("operator declined")
	rcRefuse(t, h, RefuseNotApproved)
	rcRequireNoListRead(t, h)
}

// TestReconcileApprovalShowsBothMasksAndTheAccountCountBeforeAnyListRead 는 codex F1·R2-5 처분이다.
//
// 픽스처는 **끝 4자리가 같은 다른 계좌**로 자격을 바꾼 상태다(기록 123-45-678901, 현재 999-99-998901). 마스크 대조는
// 이것을 가르지 못하고(치르는 값 ⑥), a121 이 하는 것은 사람 승인 출력에 두 마스크와 계좌 수를 반드시 보이는 것과
// 대사 줄에 마스크 외 계좌 신원을 싣지 않는 것뿐이다. 그래서 이 시험은 대사가 **진행됨**을 전제로 그 두 가지를 잰다.
func TestReconcileApprovalShowsBothMasksAndTheAccountCountBeforeAnyListRead(t *testing.T) {
	const swapped = "999-99-998901"
	if attest.Mask(swapped) != rcMask() {
		t.Fatalf("fixture: %q must collide with the record mask %q", swapped, rcMask())
	}
	h := newRCHarness(t, rcA063Entries())
	h.account.Ref = swapped
	_, _, raw := rcAccept(t, h)

	if len(h.approvals) != 1 {
		t.Fatalf("want one approval, got %d", len(h.approvals))
	}
	a := h.approvals[0]
	if a.RecordAccountMask != rcMask() || a.CurrentAccountMask != attest.Mask(swapped) || a.AccountCount != 1 ||
		a.Market != MarketKR {
		t.Fatalf("approval must carry both masks, the account count and the market: %+v", a)
	}
	if h.listCallsAtApproval != 0 {
		t.Fatalf("approval came after %d list read(s); it must come before any", h.listCallsAtApproval)
	}
	shown := h.outAtApproval
	if strings.Count(shown, rcMask()) < 2 {
		t.Fatalf("approval output must show the record mask and the current mask (both %q):\n%s", rcMask(), shown)
	}
	if !strings.Contains(shown, rcTargetID) {
		t.Fatalf("approval output must name the artifact %q:\n%s", rcTargetID, shown)
	}
	for _, leak := range []string{swapped, rcAccount, "99999", "12345"} {
		if strings.Contains(shown, leak) || strings.Contains(raw, leak) {
			t.Fatalf("unmasked account identity %q reached the approval output or the record line", leak)
		}
	}
}

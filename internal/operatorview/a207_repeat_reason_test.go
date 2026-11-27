package operatorview

import (
	"strings"
	"testing"
)

// a207: 압축 행의 사유는 「반복 거절(요약)」 계열 한국어로 보이고 코드가 새지 않음(사람 결정 2).
func TestA207TheRepeatReasonReadsAsASummarisedRepeatedRefusal(t *testing.T) {
	got := reasonText("repeat_of_refused_proposal")
	if !strings.HasPrefix(got, "반복 거절(요약)") || strings.Contains(got, "_") {
		t.Fatalf("reasonText(repeat_of_refused_proposal) = %q, want the 반복 거절(요약) wording", got)
	}
	view := BuildExitLine(Source{UnknownReason: "repeat_of_refused_proposal", EffectiveSource: "recomputed_repeat"})
	if view.Status != "unknown" || view.Reason != got {
		t.Fatalf("view = %+v, want an unknown line carrying the repeat wording", view)
	}
}

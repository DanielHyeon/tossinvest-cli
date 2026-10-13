package console

// a130_history_note_test.go 는 a130 design D6 의 시험임 — 완결 왕복 거래 표의 안내문이 성과 행이 없는 실제 사유와 거절 줄을 말함.

import (
	"strings"
	"testing"
)

func TestA130TheHistoryNoteNamesTheRealReasons(t *testing.T) {
	h := newDashboardHarness(t)
	seedJournal(t, h.journal)
	h.authenticate(t)

	page := h.page(t, "/history")
	for _, want := range []string{
		"외부 매도로 종결된 포지션은 여기에 행이 생기지 않으므로", // 기존 고정 문장(portfolio_test.go)
		"엔진이 관리하지 않던 포지션",
		"그 조정 시각의 브로커 1분봉으로 다시 계산해야 하는 경우",
		"재편입된 포지션",
		"엔진 매도 수량이 매수 쪽 수량과",
		"체결 평균가가 관측되지 않은 경우",
		"trade_outcome.not_frozen",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("the history note does not say %q", want)
		}
	}
}

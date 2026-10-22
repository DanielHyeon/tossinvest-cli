package filldetect_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
)

// TestUnobservedStatusFailsClosedAtTheEntryGate 는 실계좌에서 한 번도 관측되지 않은 status 가
// 폴링 → 원장 → 진입 게이트까지 막힘으로 끝나는지 고정함(verify-execution-capability 2.1, 사용자
// 결정 2026-10-22: CANCEL_REJECTED·REPLACE_REJECTED 는 「실제 관측 없음」으로 수용 — 그 수용의
// 조건이 이 fail-closed 경로임).
//
// brokerstate.Derive 단위 시험은 Derived.FailClosed 까지만 봄. 이 시험은 그 판정이 엔진의
// 진입 차단(ReasonBrokerStateUnknown)으로 이어지고 체결로 적히지 않음을 끝단에서 확인함.
func TestUnobservedStatusFailsClosedAtTheEntryGate(t *testing.T) {
	cases := []struct {
		name   string
		status string
	}{
		// 문서에 없는 값 — Derive 의 default 갈래(derive.go 506–507).
		{name: "undocumented status", status: "EXPIRED"},
		// 문서에는 있으나 별도 레코드 형태가 미측정인 두 값(derive.go 493·499).
		{name: "cancel rejected record", status: "CANCEL_REJECTED"},
		{name: "replace rejected record", status: "REPLACE_REJECTED"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			clk := clock.NewFake(pollStart)
			gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})
			pager := newPager(page("", rawOrder{id: "o-1", status: tc.status, quantity: "10", filled: "2"}))
			d, j := newJournalDetector(t, clk, pager, gate, filepath.Join(t.TempDir(), "journal.db"))
			ctx := context.Background()

			cycle, err := d.PollOnce(ctx)
			if err != nil {
				t.Fatalf("poll: %v", err)
			}
			if len(cycle.FailedClosed) != 1 {
				t.Fatalf("FailedClosed = %v, want the %s order", cycle.FailedClosed, tc.status)
			}
			if rejected := gate.CheckEntry(); rejected == nil ||
				rejected.Reason != execgw.ReasonBrokerStateUnknown {
				t.Fatalf("status %s must block new entries, got %v", tc.status, rejected)
			}
			events, err := j.FillEvents(ctx, "o-1")
			if err != nil {
				t.Fatal(err)
			}
			if len(events) != 0 {
				t.Fatalf("fill events = %d, want none from a refused %s snapshot", len(events), tc.status)
			}
		})
	}
}

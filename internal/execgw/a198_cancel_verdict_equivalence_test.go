package execgw_test

// a198 보강 2(D10) — 세션의 원장 판정 표(journal.CancelVerdictForClosedTarget)와 기동 해소(resolveCancel)의 상태 대응이
// 갈라지지 않게 묶음(Manager Q1·리뷰 r1 #7): 같은 주문 응답을 brokerstate 로 파생한 상태에 대해 —
//   - 새 표가 닫는다고 답하면 Resolve 를 끝까지 돌린 최종 결과(상태·사유·번호)와 같음
//   - 새 표가 닫지 않는다고 답하면, 대기형 상태는 첫 대기에서 멈춘 채 원장에 종결 전이가 없음을 봄
//   - 판정 불능(UNKNOWN)은 Resolve 가 park 하고 새 표는 무변 — 명시적 차이(REPLACED 도 같음, Manager Q4)

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/brokerstate"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

func TestA198TheSessionVerdictMatchesTheBootCancelResolution(t *testing.T) {
	const at, cancelledAt = "2026-03-30T10:20:00+09:00", "2026-03-30T10:30:01+09:00"
	for _, tc := range []struct {
		name, status, cancelled, filled string
		state                           brokerstate.State // 그 응답의 파생 상태(픽스처가 의도한 갈래를 실제로 타는지)
		park                            bool              // Resolve 가 park(UNRESOLVED) 로 끝나는 판정 불능
	}{
		{"cancelled", "CANCELED", cancelledAt, "0", brokerstate.StateCancelled, false},
		{"cancelled after a partial fill", "CANCELED", cancelledAt, "1", brokerstate.StateCancelledPartiallyFilled, false},
		{"filled", "FILLED", "", "2", brokerstate.StateFilled, false},
		{"rejected", "REJECTED", "", "0", brokerstate.StateRejected, false},
		{"rejected after a partial fill", "REJECTED", "", "1", brokerstate.StateRejectedPartiallyFilled, false},
		{"open", "PENDING", "", "0", brokerstate.StateOpenUnfilled, false},
		{"open partially filled", "PARTIAL_FILLED", "", "1", brokerstate.StateOpenPartiallyFilled, false},
		{"cancel pending", "PENDING_CANCEL", "", "0", brokerstate.StateCancelPending, false},
		{"replace pending", "PENDING_REPLACE", "", "0", brokerstate.StateReplacePending, false},
		{"unknown", "", "", "0", brokerstate.StateUnknown, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := cancellableOrderJSON("O-1", "005930", "BUY", tc.status, "2", "70000", at, tc.cancelled, tc.filled)
			// 해소기가 쓰는 파생 입구 그대로(amend_indoubt.go deriveOrder — lineage 없음).
			derived := brokerstate.DeriveOfficialOrder(json.RawMessage(body), brokerstate.Lineage{})
			if derived.State != tc.state {
				t.Fatalf("fixture derives %s (%s), want %s", derived.State, derived.Detail, tc.state)
			}
			want, reason, closes := journal.CancelVerdictForClosedTarget(string(derived.State))

			f := newDoubtFixture(t)
			attemptID := f.cancelInDoubt(t)
			reader := newOrderReader()
			reader.set("O-1", body)
			r := f.resolverWith(reader)
			switch {
			case closes:
				res, err := resolveAsync(t, r, f.clk, attemptID)
				if err != nil {
					t.Fatalf("Resolve: %v", err)
				}
				if res.State != want || string(res.Reason) != reason {
					t.Fatalf("%s: boot resolution %s|%s, session verdict %s|%s", derived.State, res.State, res.Reason, want, reason)
				}
				if want == journal.StateConfirmed && res.BrokerOrderID != "O-1" {
					t.Fatalf("boot resolution names %q, the session closes on the target O-1", res.BrokerOrderID)
				}
			case tc.park:
				res, err := resolveAsync(t, r, f.clk, attemptID)
				if err != nil {
					t.Fatalf("Resolve: %v", err)
				}
				if res.State != journal.StateUnresolvedInDoubt {
					t.Fatalf("%s: boot resolution %s, want the park the session does not copy", derived.State, res.State)
				}
			default:
				// 대기형: 첫 대기에서 멈춘 채 원장에 종결 전이가 없음 — 최종 결과(안정 관찰 FAILED·대기 뒤 park)와는 비교하지 않음.
				ctx, cancel := context.WithCancel(context.Background())
				done := make(chan struct{})
				go func() { _, _ = r.Resolve(ctx, attemptID); close(done) }()
				if !f.clk.WaitForSleepers(1, 10*time.Second) {
					cancel()
					t.Fatal("the boot resolution never reached its first wait")
				}
				rec, err := f.journal.LookupAttempt(context.Background(), attemptID)
				cancel()
				<-done
				if err != nil {
					t.Fatal(err)
				}
				if rec.State != journal.StateInDoubt {
					t.Fatalf("%s: the boot resolution settled %s on its first observation, the session verdict does not close", derived.State, rec.State)
				}
			}
		})
	}
	// REPLACED 는 원장 lineage 가 있어야 파생되므로 위 표에 없음 — 새 표는 닫지 않고(journal 시험), 기동은 park(amend_indoubt.go B19).
	if _, _, closes := journal.CancelVerdictForClosedTarget(string(brokerstate.StateReplaced)); closes {
		t.Fatal("the session verdict closes a REPLACED target")
	}
}

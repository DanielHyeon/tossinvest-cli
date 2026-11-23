package engine

// a198_in_doubt_cancel_close.go 는 세션 해소 실행자가 엔진이 낸 결과 불명(IN_DOUBT) 취소를 닫는 자리임(a198 D10).
//
// 판정은 전부 원장 함수(journal.CloseInDoubtCancelOnTargetClose)에 있음 — 대상 주문의 원장 종결 기록이 있으면 닫고, 없으면 무변.
// 이 파일은 그 결과를 로그에 남기고 실패를 다음 스윕으로 미룰 뿐임. 브로커를 부르지 않고 알림도 내지 않음
// (그 종목의 손절 지연은 길 치우기의 지연·연속 실패 경보가 이미 알림).

import (
	"context"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// closeCancel 은 취소 attempt 하나를 닫아 봄. 오류는 ctx 취소일 때만 돌려줌 — 나머지 실패는 다음 스윕이 다시 함.
func (s *inDoubtSettle) closeCancel(ctx context.Context, rec journal.AttemptRecord) error {
	// 엔진이 낸 취소만 닫음 — flatten 의 saga 는 자기 취소를 자기가 해소함(Manager Q7, intents.source).
	res, err := s.deps.Ledger.CloseInDoubtCancelOnTargetClose(ctx, rec.ID, engineSource)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		// 원문 오류는 싣지 않음 — 원장 오류가 주문 번호를 품을 수 있음. 같은 분류는 한 번만.
		class := resolveFailureClass(err)
		if s.lastFailure[rec.ID] != class {
			s.lastFailure[rec.ID] = class
			s.deps.Log.Warn(obs.EventOrderInDoubt,
				obs.FieldAttemptID, rec.ID, "error_class", class,
				obs.FieldDetail, "the in-doubt settler could not close this cancel from its target's closing record; it retries on the next sweep")
		}
		return nil
	}
	delete(s.lastFailure, rec.ID)
	if res.Closed {
		s.deps.Log.Event(obs.EventOrderInDoubt,
			obs.FieldAttemptID, rec.ID, obs.FieldSymbol, res.Symbol, "settled_state", string(res.State), obs.FieldReason, res.Reason,
			obs.FieldDetail, "the in-doubt settler closed this engine cancel from its target's closing record")
	}
	return nil
}

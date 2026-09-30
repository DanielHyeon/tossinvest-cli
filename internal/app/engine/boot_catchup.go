package engine

// boot_catchup.go 는 a094 의 기동 따라잡기임(D−2.5 · D−2.6 · Q4-5).
//
// # 왜 있나
//
// attempt 종결과 발의 해제는 다른 쓰기다. 둘 사이에서 프로세스가 멈추면 종결된 attempt 는 PendingAttempts 에서 빠져 재시작
// 복구가 다시 찾지 못하고, 발의는 영구히 무장된 채 남아 손절 평가를 억제한다. 발의 무장 뒤 Prepare 전에 멈춘 경우(attempt 가
// 없음)도 같다. 그래서 기동이 무장된 발의마다 **같은 해제 판정**(journal.ReleaseUnacceptedExitProposal)을 한 번 부른다.
//
// # 자리와 실패 의미
//
// 재시작 복구(Recovery.Run) 뒤 · 준비 신호 앞(engineRecoverySequence 클로저) — 관측 루프는 Recover 반환 뒤에만 시작하므로
// 첫 관측 전에 끝난다. 한 행의 실패는 그 행을 바꾸지 않고(오늘 상태 그대로 — 안전 쪽) 포지션 · intent 를 에피소드로 한
// critical 을 창 0 으로 기록하고 다음 행으로 간다. 복구를 실패시키지 않는다 — 실패하면 모든 포지션의 손절이 사라진다.
// 원장만 읽고 쓴다(브로커 호출 0). 멱등이다.

import (
	"context"
	"fmt"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// catchUpExitProposals 는 계정의 무장된 발의 전부에 해제 판정을 한 번씩 부름. 반환값은 해제한 포지션 수(시험용).
func catchUpExitProposals(ctx context.Context, j *journal.Journal, accountRef string,
	critical CriticalRecorder, log *obs.Logger) int {
	armed, err := j.ArmedExitProposals(ctx, accountRef)
	if err != nil {
		// 목록을 못 읽으면 아무것도 바꾸지 않았다 — 무장된 발의는 그대로이고 다음 기동이 다시 본다. 복구는 계속한다.
		if log != nil {
			log.Warn(obs.EventExitLiquidationDelayed, obs.FieldError, err.Error(),
				obs.FieldDetail, "the boot catch-up could not list the armed exit proposals; nothing was changed")
		}
		return 0
	}
	released := 0
	for _, a := range armed {
		if strings.TrimSpace(a.IntentID) == "" {
			// intent 없는 발의는 attempt 를 찾을 수 없음 — 판정하지 않고 둠(살아 있을 수 있음으로 다룸).
			continue
		}
		_, ok, err := j.ReleaseUnacceptedExitProposal(ctx, a.PositionID, a.IntentID, journal.ProposalRefused)
		if err != nil {
			recordCatchUpFailure(ctx, critical, log, a, err)
			continue
		}
		if ok {
			released++
		}
	}
	return released
}

func recordCatchUpFailure(ctx context.Context, critical CriticalRecorder, log *obs.Logger, a journal.ArmedExitProposal, cause error) {
	key := string(obs.EventExitLiquidationDelayed) + "|" + a.PositionID + "|intent:" + a.IntentID
	e := obs.Event{
		Type:  obs.EventExitLiquidationDelayed,
		Key:   key,
		Title: "기동 따라잡기가 한 포지션의 발의를 판정하지 못했다",
		Body: fmt.Sprintf("포지션 %s 의 무장된 발의(%s, 단계 %s, intent %s)를 기동 따라잡기가 판정하지 못해 그대로 두었다. "+
			"그 발의가 종결된 attempt 의 것이면 손절 평가가 억제된 채 남는다. 원인: %v",
			a.PositionID, a.Action, a.Level, a.IntentID, cause),
		Fields: map[string]any{
			"position_id":   a.PositionID,
			"intent_id":     a.IntentID,
			obs.FieldReason: cause.Error(),
		},
	}
	if log != nil {
		log.Warn(e.Type, obs.FieldError, cause.Error(), obs.FieldDetail, "boot catch-up failed for position "+a.PositionID)
	}
	if critical == nil {
		return
	}
	_ = critical.RecordCritical(ctx, e, 0)
}

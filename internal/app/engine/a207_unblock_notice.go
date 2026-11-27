package engine

// a207_unblock_notice.go 는 막힌 청산의 해소를 한 번 알리는 자리임(a207 design D5 · D6).
//
// 거절된 손절은 다음 관측에서 다시 무장됨 — 그 빈도와 즉시성은 이 파일이 건드리지 않음. 여기서 하는 일은 셋뿐임:
//   - release 가 원장의 released=true 를 받으면 포지션별 셈을 더함(메모리 — 원장 읽기 0).
//   - 확정 수용 갈래가 셈을 떼어 해소 알림 하나를 대기열에 넣음(그 포지션의 Place 뒤 — I/O 0).
//   - 주기 끝이 대기열을 알림 부품에 넘김(일반 등급 — 기록 전용 알림기에서는 막히지 않는 relay 로 감).

import (
	"context"
	"fmt"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// refusalStreak 은 한 포지션의 연속 해제 셈임.
type refusalStreak struct {
	count       int
	refused     int
	cancelled   int
	first, last time.Time
}

// noteRefusalStreak 은 release 가 원장의 released=true 를 받은 뒤에만 불림 — 포지션의 연속 해제 셈에 하나를 더함(메모리만).
func (o *ExitObserver) noteRefusalStreak(m managed, how journal.ProposalResolution) {
	if o.refusalStreaks == nil {
		o.refusalStreaks = map[string]*refusalStreak{}
	}
	streak := o.refusalStreaks[m.position.ID]
	if streak == nil {
		streak = &refusalStreak{first: o.clk.Now()}
		o.refusalStreaks[m.position.ID] = streak
	}
	streak.count++
	streak.last = o.clk.Now()
	if how == journal.ProposalCancelled {
		streak.cancelled++
	} else {
		streak.refused++
	}
}

// noteUnblocked 는 submit 의 확정 수용 갈래가 그 포지션의 Place 뒤에 부름 — 셈이 있으면 떼어 해소 알림 하나를 주기 대기열에 넣음(I/O 0).
// 셈을 떼는 자리가 이 하나뿐이라 한 해소에 한 번임(일반 등급에는 outbox 중복 제거가 없음).
func (o *ExitObserver) noteUnblocked(m managed, proposal exitpolicy.Proposal, intentID string) {
	streak := o.refusalStreaks[m.position.ID]
	if streak == nil || streak.count == 0 {
		return
	}
	delete(o.refusalStreaks, m.position.ID)
	o.unblockNotices = append(o.unblockNotices, obs.Event{
		Type:  obs.EventExitProposalUnblocked,
		Key:   string(obs.EventExitProposalUnblocked) + "|" + m.position.ID + "|" + intentID,
		Title: o.label(m.position.Symbol) + " 청산 막힘이 풀려 주문이 접수됐다",
		Body: fmt.Sprintf("%s (단계 %s): 이 엔진이 센 거절 %d 회(처음 %s · 마지막 %s) 뒤 브로커가 청산 주문을 받았다.",
			proposal.Action, proposal.Level, streak.count,
			streak.first.UTC().Format(time.RFC3339), streak.last.UTC().Format(time.RFC3339)),
		Fields: map[string]any{
			obs.FieldSymbol:    m.position.Symbol,
			"position_id":      m.position.ID,
			"action":           string(proposal.Action),
			"level":            proposal.Level,
			"refusals":         streak.count,
			"refused":          streak.refused,
			"cancelled":        streak.cancelled,
			"first_refused_at": streak.first.UTC().Format(time.RFC3339),
			"last_refused_at":  streak.last.UTC().Format(time.RFC3339),
		},
	})
}

// flushUnblockNotices 는 주기 끝(endExitOrderCycle)에서 이 주기의 해소 알림을 알림 부품에 넘김 — 모든 판정 · 제출 뒤임.
// 알림기가 없으면 o.alert 가 아무것도 하지 않음. 대기열은 언제나 비움.
func (o *ExitObserver) flushUnblockNotices(ctx context.Context) {
	notices := o.unblockNotices
	o.unblockNotices = nil
	for _, e := range notices {
		o.alert(ctx, e)
	}
}

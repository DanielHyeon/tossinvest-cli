package engine

// exit_unobserved.go 는 a090 — 관측되지 않은 보유 포지션을 세고 알리는 부분임.
//
// # 왜 있나
//
// 계정 두절 시계(checkOutage)는 **한 종목이라도** 가격이 오면 리셋됨. 그래서 보유 포지션 하나가 매 주기 응답에서 빠지거나
// (0가격 · 부재), 앞 포지션을 처리하는 동안 시세 사용 임대가 끝나거나, 판정 목록을 만드는 단계(workingSet)에서 탈락하면
// 그 포지션은 로그도 알림도 없이 손절 판정에서 무기한 빠짐. 브로커 상주 손절이 없는 지금 판정되지 않는 포지션은 보호되지
// 않는 포지션임(파일 머리 exitloop.go 「we chose not to look」).
//
// # 무엇을 세나 — 자리가 아니라 집합
//
// 한 주기에 workingSet 이 본 **보유·exit 대상 포지션**(표시 markHeld) 중 그 주기에 **판정 진입**(noteJudged)에 닿지 않은 것이
// 미관측임. 탈락 자리를 하나씩 계측하지 않으므로 새 탈락 자리가 생겨도 빠지지 않음. 완료된 정책(B10)은 표시를 해제해 범위 밖.
//
// # 언제 무엇을 하나
//
//   - 루프 안: 표시 · 원인 · 판정 진입을 **주기 집계**(ExitCycle 에 실림)에 적기만 함 — 원장 · 알림 없음.
//   - 순회 뒤 한 번(settleUnobserved): 임계 판정 · 알림 기록 · 모드 강화 · 정리 · 로그. 뒤 포지션의 손절 판정 앞에 서지 않음.
//   - 전 종목 미응답(B4) · 양보(B1) · 작업 집합 오류(B2) 주기는 settle 을 부르지 않음 — 계정 사다리가 그 사실의 주인이고,
//     기록(기점)은 지우지 않으므로 그 시간은 다음 처리 주기의 경과에 듦.
//
// # 알림 경로 (a092 정본 「등급화된 알림」)
//
// critical 알림과 모드 강화 공지는 알림기의 **기록 입구**(opts.Critical = Notifier.RecordCritical, 재알림 창 0 — a094 와 같은 주입)로
// 원장에 적재만 함 — 원격 전송은
// 배달 실행자 몫. 원장에 직접 쓰지 않음(세울 자기 사유가 없는 기록자는 입구를 씀). 기록 실패 시 진입 잠금은 입구의 생산자 래치가 함.
// 새 알림 · 공지 · 로그는 계좌 ref 를 싣지 않음(key · 필드 · 본문 모두 — 포지션 id · 전이 id 가 이미 유일함).

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// 미관측 원인. 계수의 판정에는 쓰지 않고(원인과 무관하게 셈) 운영자가 왜 답이 없었는지 가르는 데만 씀.
const (
	unobservedNoQuote        = "no_quote"           // 가격 읽기 응답에 없거나 쓸 수 없는 시세(ObserveOnce B6)
	unobservedQuoteExpired   = "quote_expired"      // 앞 포지션 처리 중 사용 임대 만료(ObserveOnce B7)
	unobservedNotWorkingSet  = "not_in_working_set" // 판정 목록 단계에서 탈락(workingSet B8 · B12 · B14 · B21)
	unobservedAlertFailed    = "alert_record_failed"
	unobservedTightenFailed  = "mode_tighten_failed"
	unobservedNoticeFailed   = "mode_notice_record_failed"
	unobservedStreakDetail   = "a held position did not reach the exit judgement this cycle"
	unobservedReleaseDetail  = "the position reached the exit judgement again"
	unobservedFailureDetail  = "an a090 record could not be written; the next processing cycle retries it"
	unobservedModeNoticeHead = "operating_mode:"
)

// unobservedCycle 은 한 주기의 집계임. owner 는 그 주기의 ExitCycle 변수 주소 — ObserveOnce 호출마다 새 변수이고, 이 필드가
// 옛 주기를 붙잡고 있는 동안 그 주소는 재사용되지 않으므로 주기 신원으로 쓸 수 있음. settle 을 부르지 않는 주기(B4)의 표시는
// owner 가 달라 다음 주기에서 버려짐. (ExitCycle 에 싣지 않는 이유: ExitCycle 은 == 로 비교되는 값이라 포인터 필드가 들어가면
// 같은 주기 둘이 달라짐 — a074 시험.)
type unobservedCycle struct {
	owner  *ExitCycle
	held   map[string]unobservedMark // 보유·exit 대상 표시(position id → 그 순간)
	causes map[string]string         // 판정에 못 닿은 원인
	judged map[string]unobservedMark // 판정 진입 순간
}

// unobservedMark 는 한 순간을 단조 앵커와 표시용 벽시계로 따로 가짐 — 경과는 앵커로만 잼(벽시계 역행이 창을 늘리지 않게).
type unobservedMark struct {
	symbol string
	anchor time.Time
	wall   time.Time
}

// unobservedRecord 는 관측자가 포지션마다 들고 있는 기록임(재시작하면 잃음 — design D2 의 정직한 한계).
type unobservedRecord struct {
	symbol string
	base   unobservedMark // 마지막 판정 순간, 한 번도 판정되지 않았으면 처음 본 순간
	streak *unobservedStreak
}

// unobservedStreak 은 미관측 연속 하나임. id 는 알림 key 의 에피소드 신원 — 벽시계 기점을 쓰면 역행 · 재시작에서 옛 행과 겹침.
type unobservedStreak struct {
	id        string
	cause     string
	recorded  bool // 알림이 원장에 적재됨(같은 연속은 다시 적재하지 않음)
	alertFail bool // 알림 적재 실패 상태(다음 처리 주기 재시도)
	tightened bool // 강화가 커밋됨 — 운영자가 완화해도 같은 연속에서는 재강화 없음
	modeFail  bool // 강화 커밋 실패 상태(다음 처리 주기 재시도)
}

// unobservedTally 는 이 주기의 집계를 돌려줌 — 다른 주기의 것이면 새로 시작함.
func (o *ExitObserver) unobservedTally(cycle *ExitCycle) *unobservedCycle {
	if o.cycleUnobserved == nil || o.cycleUnobserved.owner != cycle {
		o.cycleUnobserved = &unobservedCycle{
			owner:  cycle,
			held:   map[string]unobservedMark{},
			causes: map[string]string{},
			judged: map[string]unobservedMark{},
		}
	}
	return o.cycleUnobserved
}

func (o *ExitObserver) unobservedMoment(symbol string) unobservedMark {
	anchor := clock.LeaseAnchor(o.clk)
	return unobservedMark{symbol: symbol, anchor: anchor, wall: anchor.UTC()}
}

// markHeld 는 workingSet 이 보유·exit 대상으로 확정한 포지션을 이번 주기 집계에 표시함. 처음 보는 포지션이면 그 순간이 기점.
func (o *ExitObserver) markHeld(cycle *ExitCycle, p journal.Position) {
	mark := o.unobservedMoment(p.Symbol)
	o.unobservedTally(cycle).held[p.ID] = mark
	if o.unobserved == nil {
		o.unobserved = map[string]*unobservedRecord{}
	}
	if _, ok := o.unobserved[p.ID]; !ok {
		o.unobserved[p.ID] = &unobservedRecord{symbol: p.Symbol, base: mark}
	}
}

// unmarkHeld 는 완료된 정책(workingSet B10)을 범위 밖으로 뺌 — 관측 실패가 아니라 정책 수명 문제임(design D1).
func (o *ExitObserver) unmarkHeld(cycle *ExitCycle, positionID string) {
	delete(o.unobservedTally(cycle).held, positionID)
}

// noteUnobservedCause 는 판정에 못 닿은 원인을 적음(루프 안 — 기록만).
func (o *ExitObserver) noteUnobservedCause(cycle *ExitCycle, positionID, cause string) {
	o.unobservedTally(cycle).causes[positionID] = cause
}

// noteJudged 는 판정 진입을 적음. 「관측됨」 은 판정 진입 도달성임 — 진입 뒤 즉시 끝나도(격리 거절 · 스탬프 실패 · 하류 임대
// 재검사) 관측됨(design D9, Q3).
func (o *ExitObserver) noteJudged(cycle *ExitCycle, m managed) {
	o.unobservedTally(cycle).judged[m.position.ID] = o.unobservedMoment(m.position.Symbol)
}

// settleUnobserved 는 순회 뒤 한 번 도는 판정임: 공지 재시도 → 정리 → 포지션별 연속 · 임계 · 알림 · 강화.
func (o *ExitObserver) settleUnobserved(ctx context.Context, cycle *ExitCycle) {
	o.retryModeNotices(ctx)

	tally := o.cycleUnobserved
	o.cycleUnobserved = nil
	if tally != nil && tally.owner != cycle {
		tally = nil // 이 주기에는 표시가 없었음(옛 주기의 집계)
	}
	if tally == nil || len(tally.held) == 0 {
		// 보유·대상 표시가 하나도 없음 = 정말 무보유(또는 전부 완료 정책). 기록을 비움. 미적재 공지는 전이의 사실이라 남김.
		o.unobserved = nil
		return
	}
	// 정리 기준은 이번 주기의 보유 대상 표시임 — 보유 종료 · 대상 제외 · 완료 정책은 경보 없이 지움.
	for id := range o.unobserved {
		if _, ok := tally.held[id]; !ok {
			delete(o.unobserved, id)
		}
	}

	ids := make([]string, 0, len(tally.held))
	for id := range tally.held {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		mark := tally.held[id]
		rec := o.unobserved[id]
		if rec == nil {
			rec = &unobservedRecord{symbol: mark.symbol, base: mark}
			o.unobserved[id] = rec
		}
		rec.symbol = mark.symbol

		if judged, ok := tally.judged[id]; ok {
			if rec.streak != nil {
				// 연속 해제: 기점에서 판정 순간까지를 한 줄로 남김.
				seconds := int((clock.LeaseElapsed(o.clk, rec.base.anchor) - clock.LeaseElapsed(o.clk, judged.anchor)).Seconds())
				o.logUnobserved(false, "position_id", id, obs.FieldSymbol, rec.symbol,
					"unobserved_seconds", seconds, obs.FieldDetail, unobservedReleaseDetail)
				rec.streak = nil
			}
			rec.base = judged
			continue
		}

		cycle.Unobserved++
		cause := tally.causes[id]
		if cause == "" {
			cause = unobservedNotWorkingSet
		}
		if rec.streak == nil {
			rec.streak = &unobservedStreak{id: o.opts.NewID()}
			o.logUnobserved(false, "position_id", id, obs.FieldSymbol, rec.symbol,
				"cause", cause, obs.FieldDetail, unobservedStreakDetail)
		}
		rec.streak.cause = cause

		// 기점 = 마지막 판정(없으면 처음 봄). 양보 · 전 종목 실패 주기의 시간도 여기에 듦.
		elapsed := clock.LeaseElapsed(o.clk, rec.base.anchor)
		if elapsed < o.outageAfter() {
			continue
		}
		o.raiseUnobserved(ctx, cycle, id, rec, elapsed)
	}
}

// raiseUnobserved 는 임계를 넘은 연속 하나의 알림 · 강화 상태기계임(design D10).
func (o *ExitObserver) raiseUnobserved(ctx context.Context, cycle *ExitCycle, id string, rec *unobservedRecord,
	elapsed time.Duration) {
	s := rec.streak
	if !s.recorded {
		recorder := o.opts.Critical
		if recorder == nil {
			return
		}
		if err := recorder.RecordCritical(ctx, o.unobservedEvent(id, rec, elapsed), 0); err != nil {
			// 진입 잠금은 입구의 생산자 래치가 이미 함. 여기는 상태와(처음 실패일 때만) 오류 종류 한 줄 — 원문 오류는 계좌를 품을 수 있음.
			if !s.alertFail {
				o.logUnobservedFailure(id, rec.symbol, unobservedAlertFailed)
			}
			s.alertFail = true
			return
		}
		s.recorded, s.alertFail = true, false
	}
	// 강화는 알림이 적재된 연속에서만 — 알림 없는 조임을 만들지 않음.
	if s.tightened || o.opts.Escalate == nil {
		return
	}
	_, changed, err := o.opts.Escalate.EscalateOperatingMode(ctx, o.opts.AccountRef,
		journal.ModeTriggerExitObservationOutage, unobservedAnnouncer{o: o})
	if err != nil && !errors.Is(err, journal.ErrModeAnnouncementFailed) {
		if !s.modeFail {
			o.logUnobservedFailure(id, rec.symbol, unobservedTightenFailed)
		}
		s.modeFail = true
		return
	}
	// 오류 없음(이미 그 모드여도 조인 것으로 봄) 또는 전이는 커밋되고 공지 적재만 실패 — 전이는 다시 하지 않음(공지는 대기열이 재시도).
	s.tightened, s.modeFail = true, false
	if changed {
		cycle.Escalated = true
	}
}

// unobservedEvent 는 포지션 단위 두절 알림임. 계정 두절과 같은 타입이되 본문이 포지션을 명명해 구별되고, key 는 연속 신원.
func (o *ExitObserver) unobservedEvent(id string, rec *unobservedRecord, elapsed time.Duration) obs.Event {
	since := rec.base.wall.Format(time.RFC3339)
	return obs.Event{
		Type:  obs.EventExitObservationOutage,
		Key:   string(obs.EventExitObservationOutage) + "|" + id + "|" + rec.streak.id,
		Title: o.label(rec.symbol) + " 포지션 관측이 " + elapsed.Round(time.Second).String() + " 동안 끊겼다",
		Body: fmt.Sprintf("포지션 %s(%s)이 %s 이후 손절 판정에 닿지 않았다(원인 %s). 다른 종목은 관측되고 있어 계정 단위 "+
			"두절 경보로는 드러나지 않는다. 브로커에 상주하는 손절이 없는 상태에서 판정되지 않는 포지션은 보호되지 않는 "+
			"포지션이므로, 운영자가 모드를 완화할 때까지 신규 진입을 차단한다.", id, rec.symbol, since, rec.streak.cause),
		Fields: map[string]any{
			obs.FieldSymbol:      rec.symbol,
			"position_id":        id,
			"unobserved_seconds": int(elapsed.Seconds()),
			"cause":              rec.streak.cause,
			"since":              since,
		},
	}
}

// unobservedAnnouncer 는 a090 전용 정화 공지자임(design D5 5판). 같은 사건 구성(obs.OperatingModeEvent)을 쓰되 key 에서 계좌를 빼고
// 전이 id 를 넣으며(operating_mode:<mode>:<전이 id>) FieldAccount 를 지움. 기록 입구에 창 0 으로 적재만 함 — 동기 전송 없음.
type unobservedAnnouncer struct{ o *ExitObserver }

func (a unobservedAnnouncer) AnnounceOperatingMode(ctx context.Context, previous string,
	rec journal.OperatingModeRecord) error {
	e := obs.OperatingModeEvent(previous, rec)
	e.Key = unobservedModeNoticeHead + rec.Mode + ":" + rec.ID
	delete(e.Fields, obs.FieldAccount)
	return a.o.recordModeNotice(ctx, e)
}

// recordModeNotice 는 공지를 적재하고, 실패하면 미적재 대기열에 둠 — 전이는 커밋됐으므로 공지만 다음 처리 주기에 재시도함
// (연속이 끝나도 남음: 공지는 전이의 사실).
func (o *ExitObserver) recordModeNotice(ctx context.Context, e obs.Event) error {
	recorder := o.opts.Critical
	if recorder == nil {
		return nil
	}
	if err := recorder.RecordCritical(ctx, e, 0); err != nil {
		o.pendingModeNotices = append(o.pendingModeNotices, e)
		o.logUnobservedFailure("", "", unobservedNoticeFailed)
		return err
	}
	return nil
}

// retryModeNotices 는 처리 주기마다 미적재 공지를 같은 key 로 다시 적재함(같은 전이 id — 멱등).
func (o *ExitObserver) retryModeNotices(ctx context.Context) {
	if len(o.pendingModeNotices) == 0 {
		return
	}
	recorder := o.opts.Critical
	if recorder == nil {
		return
	}
	var kept []obs.Event
	for _, e := range o.pendingModeNotices {
		if err := recorder.RecordCritical(ctx, e, 0); err != nil {
			kept = append(kept, e)
		}
	}
	o.pendingModeNotices = kept
}

func (o *ExitObserver) logUnobserved(warn bool, args ...any) {
	if o.opts.UnobservedLog == nil {
		return
	}
	if warn {
		o.opts.UnobservedLog.Warn(obs.EventExitPositionUnobserved, args...)
		return
	}
	o.opts.UnobservedLog.Event(obs.EventExitPositionUnobserved, args...)
}

// logUnobservedFailure 는 a090 기록 실패의 **종류**만 남김 — 원문 오류는 싣지 않음(공유 경로의 오류 문구는 계좌를 품을 수 있음).
func (o *ExitObserver) logUnobservedFailure(id, symbol, failure string) {
	args := []any{"failure", failure, obs.FieldDetail, unobservedFailureDetail}
	if id != "" {
		args = append(args, "position_id", id, obs.FieldSymbol, symbol)
	}
	o.logUnobserved(true, args...)
}

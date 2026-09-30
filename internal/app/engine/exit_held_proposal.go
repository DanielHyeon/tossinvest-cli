package engine

// exit_held_proposal.go 는 a094 가 exit 관측 루프에 더한 것 중 **새 함수**만 모은 파일임(기존 함수 편집은 exitloop.go 에
// 최소로). 셋을 함:
//
//  1. 청소 결과를 사실로 돌려줄 값(clearResult) — 치웠는가, 연속 실패 계수 대상인가, 매도 취소의 종결 증거를 기다리는가.
//  2. 무장된 발의가 왜 풀리지 않는지 알리는 두 명명 critical — park 원인(a094 D−4.4)과 종결 증거 대기(D−9.3).
//     둘 다 판정 진입(judge)에서 원장만 읽고 알림만 하며 평가 · 억제 · 청소의 순서와 결과를 바꾸지 않음.
//  3. 청소의 연속 실패를 기존 30초 지연 경보보다 이른 트리거로 알림(D−2.7 · D−7.1 — key 는 연속 id).
//
// 새 critical 은 전부 알림기의 critical 기록 단일 입구(obs.Notifier.RecordCritical)로 **창 0** 기록만 함(D−6.1) — 관측
// 루프에서 원격 전송을 기다리지 않고, 같은 에피소드는 같은 행임. 기록 실패의 진입 잠금은 입구의 몫임(a092).
//
// 브로커 호출 0 — 전부 원장 읽기(D−4.8).

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// CriticalRecorder 는 알림기의 critical 기록 단일 입구임(a092 K6). *obs.Notifier 가 구현함.
type CriticalRecorder interface {
	RecordCritical(ctx context.Context, e obs.Event, remindAfter time.Duration) error
}

// clearResult 는 청소 한 번의 결과임.
type clearResult struct {
	// cleared 는 보호 청산을 제출해도 되는가임 — 거짓이면 그 주기에 제출하지 않음.
	cleared bool
	// countable 은 연속 실패 계수(D−2.7)에 넣는가임. 거짓은 치우지 못한 것이 전부 엔진 취소의 기록 · 전송 · 인수
	// 단계인 경우뿐(D−2.7 의 PENDING_CANCEL 제외) — 모호(IN_DOUBT) 취소는 셈.
	countable bool
	// awaitingClose 는 매도 취소가 접수 확정됐지만 종결 체결 기록이 아직 없는 형태 B(D−4.3)임 — 지연 경보 본문이
	// 체결 감지 복구 절차를 안내함(D−8.3).
	awaitingClose bool
}

// exitClearWhy 는 지연 경보의 사유 문장임. 형태 B 는 복구 절차 안내를 덧붙임(D−8.3 — 등급 · key · 빈도는 무변화).
const (
	exitClearWhyWorkingOrder  = "a working order on the symbol could not be taken off the book"
	exitClearWhyAwaitingClose = "취소된 매도의 종결을 기다리는 중 — 체결 감지 상태를 확인하고 멈췄다면 " +
		"operations 「체결 감지가 멈췄을 때」 절차를 따르라"
)

func (r clearResult) why() string {
	if r.awaitingClose {
		return exitClearWhyAwaitingClose
	}
	return exitClearWhyWorkingOrder
}

// onlyEngineCancelsInFlight 는 같은 종목의 미종결 attempt 가 전부 엔진의 취소이고 기록 · 전송 시작 · 인수 단계인지임
// (D−2.7 의 PENDING_CANCEL). 그때만 연속 실패 계수를 늘리지 않음 — 치우지 못한 것이 아니라 치우는 중이므로.
func onlyEngineCancelsInFlight(unsettled []journal.AttemptRecord) bool {
	if len(unsettled) == 0 {
		return false
	}
	for _, rec := range unsettled {
		if rec.Kind != journal.KindCancel {
			return false
		}
		switch rec.State {
		case journal.StateRecorded, journal.StateDispatchStarted, journal.StateAcked:
		default:
			return false
		}
	}
	return true
}

// workingOrderPrice 는 청소 대상 주문의 가격을 읽음. 빈 가격(시장가)은 0 임 — 가격은 취소 요청의 식별에 쓰이지 않으므로
// 그것을 실패로 읽으면 취소 가능한 주문이 보호 청산을 영구 보류시킴(a094 3.B3, a087 대비). 수량은 여전히 엄격함.
func workingOrderPrice(price string) (float64, error) {
	if strings.TrimSpace(price) == "" {
		return 0, nil
	}
	return floatOf("working order price", price)
}

// noteHeldProposal 은 무장된 발의가 풀리지 않는 원인 둘을 알림(판정 진입, 원장 읽기 + 기록만).
//
// 결과를 바꾸지 않음: 오류는 로그로만 남기고 반환하지 않음 — 이 알림의 실패가 그 포지션의 판정(손절 포함)을 막으면
// 알림이 손절을 잡게 됨.
func (o *ExitObserver) noteHeldProposal(ctx context.Context, m managed) {
	if !m.state.Pending() || strings.TrimSpace(m.state.PendingIntentID) == "" {
		return
	}
	facts, err := o.opts.Journal.ExitIntentAttempts(ctx, m.state.PendingIntentID)
	if err != nil {
		o.warnA094(obs.EventExitLiquidationDelayed, err, "reading the attempts of the armed proposal of "+m.position.ID)
		return
	}
	// park 원인(D−4.4) — 청소 자격과 무관하게, 무장된 발의가 손절 자신이어도.
	for _, rec := range facts.Parked {
		o.recordParkCause(ctx, m, rec)
	}
	// 종결 증거 대기(D−9.3) — 엔진 취소가 접수 확정된 매도가 청산 지연 한계 이상 종결 기록 없이 남음.
	for _, order := range facts.Awaiting {
		if !strings.EqualFold(strings.TrimSpace(order.Side), "SELL") || strings.TrimSpace(order.OrderID) == "" {
			continue
		}
		cancel, found, err := o.opts.Journal.ConfirmedCancelOf(ctx, order.AccountRef, order.Market, order.Symbol, order.OrderID)
		if err != nil {
			o.warnA094(obs.EventExitLiquidationDelayed, err, "reading the engine cancel of order "+order.OrderID)
			continue
		}
		if !found {
			continue
		}
		settled, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(cancel.SettledAt))
		if err != nil {
			o.warnA094(obs.EventExitLiquidationDelayed, err, "reading when cancel "+cancel.ID+" settled")
			continue
		}
		waited := o.clk.Now().Sub(settled)
		if waited < o.delayBound() {
			continue
		}
		o.recordAwaitingClose(ctx, m, order, cancel, settled, waited)
	}
}

func (o *ExitObserver) recordParkCause(ctx context.Context, m managed, rec journal.AttemptRecord) {
	key := string(obs.EventOrderUnresolved) + "|" + m.position.ID + "|attempt:" + rec.ID
	if o.a094Latched(key) {
		return
	}
	o.recordA094Critical(ctx, key, obs.Event{
		Type:  obs.EventOrderUnresolved,
		Key:   key,
		Title: o.label(m.position.Symbol) + " 무보호 — 판정 불능 attempt 가 발의를 잡고 있다",
		Body: fmt.Sprintf("무장된 발의(%s, 단계 %s)의 attempt %s 가 판정 불능(UNRESOLVED_IN_DOUBT)으로 park 됐다. "+
			"원 주문이 살아 있을 수 있어 발의를 비우지 않으며, 그동안 이 포지션의 손절은 나가지 않는다. "+
			"운영자가 브로커에서 확인한 뒤 해동 명령(tossctl engine attempt-resolve)으로 닫아야 풀린다.",
			m.state.PendingAction, m.state.PendingLevel, rec.ID),
		Fields: map[string]any{
			obs.FieldSymbol:    m.position.Symbol,
			obs.FieldAttemptID: rec.ID,
			"position_id":      m.position.ID,
			"intent_id":        rec.IntentID,
			"attempt_kind":     string(rec.Kind),
		},
	})
}

func (o *ExitObserver) recordAwaitingClose(ctx context.Context, m managed, order journal.IntentOrderAwaitingClose,
	cancel journal.AttemptRecord, settled time.Time, waited time.Duration) {
	key := string(obs.EventExitLiquidationDelayed) + "|" + m.position.ID + "|cancel:" + cancel.ID
	if o.a094Latched(key) {
		return
	}
	o.recordA094Critical(ctx, key, obs.Event{
		Type:  obs.EventExitLiquidationDelayed,
		Key:   key,
		Title: o.label(m.position.Symbol) + " 청산 보류 — 취소된 매도의 종결 증거를 기다리는 중",
		Body: fmt.Sprintf("취소된 매도의 종결 증거를 기다리는 중 — 체결 감지가 건강해도 브로커 기록이 종결을 증명하지 못하면 "+
			"풀리지 않는다; 손 해소 수단은 아직 없다. 주문 %s · 취소 %s · 대기 %s. 그동안 손절은 나가지 않는다.",
			order.OrderID, settled.UTC().Format(time.RFC3339), waited.Round(time.Second)),
		Fields: map[string]any{
			obs.FieldSymbol:    m.position.Symbol,
			obs.FieldOrderID:   order.OrderID,
			obs.FieldAttemptID: cancel.ID,
			"position_id":      m.position.ID,
			"delay_seconds":    int(waited.Seconds()),
		},
	})
}

// noteClearFailure 는 청소 실패 연속을 셈(D−2.7 · D−7.1). 연속의 시작에서 연속 id 를 한 번 만들고, 계수 대상인 실패가
// obs.DefaultCriticalAttempts 번 쌓이면 그 연속을 에피소드로 한 critical 을 한 번 기록함. 기존 지연 타이머
// (delayedSince · delayAlerted)는 건드리지 않음. 그 뒤에도 자동 제출하지 않음.
func (o *ExitObserver) noteClearFailure(ctx context.Context, m managed, res clearResult) {
	if o.clearStreak == nil {
		o.clearStreak = map[string]*clearStreak{}
	}
	streak, running := o.clearStreak[m.position.ID]
	if !running {
		streak = &clearStreak{id: o.opts.NewID()}
		o.clearStreak[m.position.ID] = streak
	}
	if !res.countable {
		return
	}
	streak.failures++
	if streak.failures < obs.DefaultCriticalAttempts || streak.alerted {
		return
	}
	streak.alerted = true
	key := string(obs.EventExitLiquidationDelayed) + "|" + m.position.ID + "|streak:" + streak.id
	o.recordA094Critical(ctx, key, obs.Event{
		Type:  obs.EventExitLiquidationDelayed,
		Key:   key,
		Title: o.label(m.position.Symbol) + " 청산의 길을 연속으로 치우지 못했다",
		Body: fmt.Sprintf("%s · 연속 %d회. 치우지 못한 충돌 위에는 보호 청산을 내지 않는다(초과 매도 방향).",
			res.why(), streak.failures),
		Fields: map[string]any{
			obs.FieldSymbol: m.position.Symbol,
			"position_id":   m.position.ID,
			"failures":      streak.failures,
			obs.FieldDetail: res.why(),
		},
	})
}

// endClearStreak 는 청소가 완료된 주기에 연속을 끝냄 — 다음 실패는 새 연속(새 id)임.
func (o *ExitObserver) endClearStreak(positionID string) {
	delete(o.clearStreak, positionID)
}

type clearStreak struct {
	id       string
	failures int
	alerted  bool
}

// a094Latched 는 같은 에피소드를 한 프로세스 안에서 다시 적재하지 않게 하는 래치임 — 적재 중복을 줄이는 최적화일 뿐
// 전송의 근거가 아님(재시작 뒤 같은 에피소드는 같은 key 라 원장이 같은 행을 돌려줌, D−5.2).
func (o *ExitObserver) a094Latched(key string) bool {
	return o.a094Recorded[key]
}

// recordA094Critical 은 새 critical 하나를 단일 입구로 창 0 기록함. 기록이 성공했을 때만 래치함 — 실패하면 다음 관측이
// 다시 시도함(진입 잠금은 입구가 이미 세움).
func (o *ExitObserver) recordA094Critical(ctx context.Context, key string, e obs.Event) {
	if o.opts.Critical == nil {
		o.log(e.Type, true, obs.FieldDetail, "no critical recorder is wired, so this alert is not durable: "+e.Title)
		return
	}
	if err := o.opts.Critical.RecordCritical(ctx, e, 0); err != nil {
		o.warnA094(e.Type, err, "the critical alert could not be recorded")
		return
	}
	if o.a094Recorded == nil {
		o.a094Recorded = map[string]bool{}
	}
	o.a094Recorded[key] = true
}

// warnA094 는 이 파일의 원장 읽기 · 기록 실패를 한 줄로 남김. logErr 와 달리 계좌 필드를 싣지 않음(불변식 8 — 계좌 원문
// 로그 금지). 오류 문구는 원장 키(포지션 · attempt · 주문 번호)만 담음.
func (o *ExitObserver) warnA094(t obs.EventType, err error, detail string) {
	o.log(t, true, obs.FieldError, err.Error(), obs.FieldDetail, detail)
}

package engine

// exit_order_log.go 는 a129 의 청산 주문 생애 줄임 — 엔진이 낸 청산 주문의 확정 · 결과 미상 · 거부 · 브로커 전 거절 · 체결 ·
// 무체결 종결을 전용 로거(ExitObserverOptions.ExitOrderLog)에 한 줄씩 남김.
//
// # 왜 전용 로거인가
//
// 관측자 전체 Log 는 생산에서 nil 임(engineRuntime 이 넘기지 않음). 그 nil 때문에 주문 확정 줄(exitloop.go 의 order.confirmed)도
// 나가지 않았고, 2026-10-07 010170 손절(SELL LIMIT 6 @18010, 전량 체결)이 engine.log 에 0 줄이라 운영자가 손절 미발동으로 오판함.
// Log 를 켜면 같은 싱크의 logErr 줄이 계좌 참조와 오류 원문을 더 내보내므로, a090 UnobservedLog · a091 ZeroFloorLog 와 같은
// 좁은 싱크를 둠(design D1 — engine.log 의 다른 생산자가 이미 계좌를 싣는 것은 별도 잔여).
//
// # 화이트리스트
//
// 줄은 exitOrderLine 하나에서만 만들어짐. 구조체 필드가 곧 내보낼 수 있는 값의 전부이고, 관측자 · 포지션 · 브로커 결과를 그대로
// 넘기는 길은 없음. 사유 · 상태 문자열은 토큰 거름(exitOrderToken)을 지남 — 숫자가 섞인 값은 통과하지 못함.
//
// # 청산을 막지도 늦추지도 않음 (불변식 4)
//
// submit 은 줄을 쓰지 않고 이 주기의 대기열에 넣기만 함(메모리 append). 대기열은 ObserveOnce 의 defer(endExitOrderCycle)가
// 그 주기의 모든 판정 · 제출 · 미관측 정리 **뒤에** 씀 — 막힌 쓰기 대상도 같은 주기 뒤 포지션의 손절 판정 앞에 서지 못함(a090 과
// 같은 규칙). 줄 하나의 쓰기는 그 줄만 감싼 recover 안에서 함 — 쓰기 대상이 패닉해도 루프는 계속 돌고, 다른 곳의 패닉은 삼키지 않음.
// 체결 따라가기는 원장 읽기만 함(브로커 호출 0).

import (
	"context"
	"errors"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

// a129 가 더하는 이벤트 이름 둘임. obs.EventFillObserved 는 「양(+)의 체결 증가」 로 정의돼 있어(event.go) 무체결 종결과 따라가기
// 실패를 그 이름으로 세면 세는 규약이 깨짐. internal/obs 는 이 change 의 편집 범위 밖이라 여기 둠(선례:
// cmd/tossctl/notificationsettings.go 의 obs.EventType 리터럴). 등급표에 없으므로 normal.
const (
	// eventExitOrderEnded 는 체결 증가 없이 종결된 청산 주문임 — 「손절이 한 주도 못 팔고 끝났다」 가 이 줄로 보임.
	eventExitOrderEnded obs.EventType = "exit.order_ended"
	// eventExitFillFollowFailed 는 청산 주문의 원장 체결 기록을 읽지 못한 것임(주문당 한 번).
	eventExitFillFollowFailed obs.EventType = "exit.fill_follow_failed"
)

// 브로커 전 거절과 따라가기 실패의 고정 사유 코드임 — 오류 원문 대신 실음.
const (
	exitOrderFloorArithmeticFailed = "floor_arithmetic_failed"
	exitOrderFloorZero             = "confirmed_floor_zero"
	exitOrderGuardianRefused       = "guardian_refused"
	exitOrderAttachFailed          = "attach_failed"
	exitOrderIntentInvalid         = "intent_invalid"
	exitOrderIntentLookupFailed    = "intent_lookup_failed"
	exitOrderFillLookupFailed      = "fill_lookup_failed"
	exitOrderFollowExpired         = "follow_expired"
	// exitOrderUnrecognised 는 토큰 거름을 통과하지 못한 사유 · 상태의 대체값임.
	exitOrderUnrecognised = "unrecognised"
)

// exitFillFollowLimit 은 체결 따라가기 항목 하나의 수명임 — 이 안에 종결 기록이 안 오면(브로커에 닿지 않은 결과 미상 주문 ·
// 종결 기록이 끝내 안 쓰인 주문) 한 번 알리고 버림. 주기마다 원장 읽기 하나가 재시작까지 남지 않게 함. DAY 주문의 하루보다 김.
const exitFillFollowLimit = 24 * time.Hour

// exitFillReadBudget 은 한 주기 끝의 체결 따라가기 원장 읽기 전체에 주는 시한임 — 원장 연결이 막혀도 다음 주기를 붙잡지 않게 함.
const exitFillReadBudget = time.Second

// exitOrderTokenPattern 은 사유 · 상태로 실을 수 있는 모양임 — 영문자와 밑줄만. base 의 ReasonCode · AttemptState · 원장 사유
// 상수는 전부 이 모양이고, 숫자를 막으므로 계좌 번호 모양 값이 사유 문자열에 실려 와도 통과하지 못함.
var exitOrderTokenPattern = regexp.MustCompile(`^[A-Za-z_]{1,64}$`)

// exitOrderToken 은 사유 · 상태 문자열을 거름. 빈 값은 빈 값(사유 없음), 모양이 아니면 대체값.
func exitOrderToken(s string) string {
	s = strings.TrimSpace(s)
	if s == "" || exitOrderTokenPattern.MatchString(s) {
		return s
	}
	return exitOrderUnrecognised
}

// exitOrderLine 은 청산 주문 줄 하나의 전부임 — 이 필드들만 로그에 나감(design D3 화이트리스트).
type exitOrderLine struct {
	symbol, market          string
	quantity, price         string
	action, level           string
	observedPrice, baseline string
	capped                  bool
	intentID, positionID    string
	attemptID, orderID      string
	state, reason           string

	// 체결 · 종결 줄만 — 원장 FillSnapshotRecord 에서 옮김(계좌 필드는 옮기지 않음).
	fill                                        bool
	orderQuantity, filledQuantity, averagePrice string
	terminal                                    bool
}

// exitOrderLineFor 는 submit 의 문맥에서 줄의 공통 필드를 채움. 지정가는 sellIntent 와 같은 규칙(관측가, 비면 기준선).
func exitOrderLineFor(m managed, proposal exitpolicy.Proposal, quantity, intentID, observed string, capped bool) exitOrderLine {
	price := strings.TrimSpace(observed)
	if price == "" {
		price = strings.TrimSpace(m.state.Baseline)
	}
	return exitOrderLine{
		symbol: m.position.Symbol, market: m.position.Market,
		quantity: quantity, price: price,
		action: string(proposal.Action), level: proposal.Level,
		observedPrice: observed, baseline: m.state.Baseline, capped: capped,
		intentID: intentID, positionID: m.position.ID,
	}
}

// because 는 고정 사유를 단 사본임.
func (l exitOrderLine) because(reason string) exitOrderLine {
	l.reason = reason
	return l
}

// placed 는 Place 결과의 키 · 상태 · 사유를 단 사본임 — Detail · Result(브로커 응답 본문)는 옮기지 않음.
func (l exitOrderLine) placed(out execgw.Outcome) exitOrderLine {
	l.attemptID = out.AttemptID
	l.orderID = out.BrokerOrderID
	l.state = string(out.State)
	l.reason = string(out.Reason)
	return l
}

// filled 는 원장 체결 기록의 수량 · 평균가 · 상태를 단 사본임.
func (l exitOrderLine) filled(rec journal.FillSnapshotRecord) exitOrderLine {
	l.fill = true
	l.orderQuantity, l.filledQuantity, l.averagePrice = rec.Quantity, rec.FilledQuantity, rec.AveragePrice
	l.state, l.terminal, l.reason = rec.State, rec.Terminal, ""
	return l
}

// args 는 줄을 obs.Logger 인자로 바꿈. 키 "level" 은 쓰지 않음 — slog JSON 의 level 과 겹쳐 읽는 쪽이 뒤의 것만 남김(exit_level).
func (l exitOrderLine) args() []any {
	args := []any{
		"origin", "exit",
		obs.FieldSymbol, l.symbol,
		obs.FieldMarket, l.market,
		obs.FieldSide, "SELL",
		"order_type", "LIMIT",
		obs.FieldQuantity, l.quantity,
		obs.FieldPrice, l.price,
		"action", l.action,
		"exit_level", l.level,
		"observed_price", l.observedPrice,
		"baseline", l.baseline,
		"capped", l.capped,
		obs.FieldIntentID, l.intentID,
		"position_id", l.positionID,
		obs.FieldAttemptID, l.attemptID,
		obs.FieldOrderID, l.orderID,
		obs.FieldToState, exitOrderToken(l.state),
		obs.FieldReason, exitOrderToken(l.reason),
	}
	if l.fill {
		args = append(args,
			"order_quantity", l.orderQuantity,
			"filled_quantity", l.filledQuantity,
			"average_price", l.averagePrice,
			"terminal", l.terminal)
	}
	return args
}

// queuedExitOrderLine 은 주기 끝에 쓸 줄 하나임.
type queuedExitOrderLine struct {
	event obs.EventType
	warn  bool
	line  exitOrderLine
}

// exitFillWatch 는 체결을 따라가는 청산 주문 하나임(메모리 — 재시작하면 잃음, design D5 잔여).
type exitFillWatch struct {
	line exitOrderLine
	// scope 는 주문의 정규 신원임 — 원장 intent 에서 처음 따라갈 때 채움. 재사용된 브로커 주문 번호가 다른 날의 주문 기록을
	// 이 주문의 체결로 읽지 않게 함(LookupFill 이 아니라 LookupFillScoped).
	scope    journal.FillSnapshotScope
	resolved bool
	// filled 는 마지막으로 줄을 쓴 누적 체결 수량임(처음 "0").
	filled string
	// failed 는 따라가기 실패 줄을 연속 실패당 한 번만 쓰게 하는 래치임 — 읽기가 한 번 성공하면 풀림.
	failed bool
	// since 는 따라가기를 시작한 시각임(관측자 시계) — exitFillFollowLimit 판정용.
	since time.Time
}

// noteExitOrder 는 submit 이 줄 하나를 이 주기의 대기열에 넣음(쓰지 않음 — 쓰기는 endExitOrderCycle). follow 면 브로커 주문
// 번호가 있을 때 체결 따라가기 대상에 넣음. 전용 로거가 없으면 아무것도 하지 않음 — 대기열도 따라가기도 원장 읽기도 0.
//
// 거절 줄(refusal)은 포지션마다 같은 결과가 **연속한 주기**에 이어지면 처음 한 번만 넣음 — 해제된 발의는 다음 주기에 다시
// 발의되므로 막힌 손절이 5초마다 같은 줄을 영원히 쓰게 됨(a074 의 「5초마다 한 줄은 관측이 아니다」). 결과가 바뀌거나, 확정 ·
// 결과 미상이 오거나, 그 포지션의 거절이 없는 주기가 한 번이라도 끼면(가격 회복 뒤 다시 이탈 = 새 에피소드) 래치를 풂.
func (o *ExitObserver) noteExitOrder(t obs.EventType, warn, refusal, follow bool, line exitOrderLine) {
	if o.opts.ExitOrderLog == nil {
		return
	}
	if refusal {
		signature := strings.Join([]string{string(t), line.action, line.level, line.reason, line.state}, "|")
		// 이 주기에 거절이 났다는 표시는 줄을 넣든 래치로 거르든 먼저 함 — 그래야 연속한 거절이 래치를 이어 감.
		if o.exitOrderRefusedNow == nil {
			o.exitOrderRefusedNow = map[string]bool{}
		}
		o.exitOrderRefusedNow[line.positionID] = true
		if o.exitOrderRefused[line.positionID] == signature {
			return
		}
		if o.exitOrderRefused == nil {
			o.exitOrderRefused = map[string]string{}
		}
		o.exitOrderRefused[line.positionID] = signature
	} else {
		delete(o.exitOrderRefused, line.positionID)
	}
	o.exitOrderQueue = append(o.exitOrderQueue, queuedExitOrderLine{event: t, warn: warn, line: line})
	// 키는 intent — 브로커 주문 번호는 날을 넘어 재사용될 수 있어 번호를 키로 하면 두 주문이 한 항목으로 겹침.
	key := strings.TrimSpace(line.intentID)
	if !follow || strings.TrimSpace(line.orderID) == "" || key == "" {
		return
	}
	if o.exitFills == nil {
		o.exitFills = map[string]*exitFillWatch{}
	}
	if _, ok := o.exitFills[key]; !ok {
		o.exitFills[key] = &exitFillWatch{line: line, filled: "0", since: o.clk.Now()}
	}
}

// endExitOrderCycle 은 ObserveOnce 의 defer 임 — 이 주기의 대기열을 쓰고, 따라가는 주문의 체결을 원장에서 읽음. 모든 return
// 경로에서, 그 주기의 판정 · 제출 뒤에 돎. 반환값 cycle 은 이름 없는 결과라 이 함수가 바꿀 수 없음.
func (o *ExitObserver) endExitOrderCycle(ctx context.Context) {
	if o.opts.ExitOrderLog == nil {
		return
	}
	queue := o.exitOrderQueue
	o.exitOrderQueue = nil
	for _, q := range queue {
		o.writeExitOrder(q.event, q.warn, q.line)
	}
	// 이 주기에 거절이 없던 포지션의 래치는 풂 — 래치는 연속한 주기의 같은 거절만 묶음.
	for id := range o.exitOrderRefused {
		if !o.exitOrderRefusedNow[id] {
			delete(o.exitOrderRefused, id)
		}
	}
	o.exitOrderRefusedNow = nil
	reads, cancel := context.WithTimeout(ctx, exitFillReadBudget)
	defer cancel()
	o.followExitFills(reads)
}

// writeExitOrder 는 줄 하나를 전용 로거에 씀. recover 는 이 쓰기 하나만 감쌈 — 쓰기 대상의 패닉이 관측 루프를 끝내지 않게.
// 관측 주기의 다른 패닉(Place · 원장)은 이 recover 에 닿지 않음(Go 의 recover 는 패닉을 일으킨 프레임의 defer 에서만 받음).
func (o *ExitObserver) writeExitOrder(t obs.EventType, warn bool, line exitOrderLine) {
	logger := o.opts.ExitOrderLog
	if logger == nil {
		return
	}
	defer func() { _ = recover() }()
	if warn {
		logger.Warn(t, line.args()...)
		return
	}
	logger.Event(t, line.args()...)
}

// followExitFills 는 따라가는 주문마다 원장의 체결 기록을 한 번 읽음. 누적 체결이 마지막으로 쓴 값보다 크면 fill.observed,
// 늘지 않은 채 종결이면 exit.order_ended 를 쓰고, 종결이면 따라가기를 끝냄. 실패는 주기 결과에 넣지 않음(기록 실패는 판정이 아님).
func (o *ExitObserver) followExitFills(ctx context.Context) {
	if len(o.exitFills) == 0 || ctx.Err() != nil {
		return
	}
	ids := make([]string, 0, len(o.exitFills))
	for id := range o.exitFills {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	now := o.clk.Now()
	for _, key := range ids {
		w := o.exitFills[key]
		id := strings.TrimSpace(w.line.orderID)
		if now.Sub(w.since) >= exitFillFollowLimit {
			// 수명 안에 종결 기록이 오지 않음 — 한 번 알리고 버림(원장이 정본).
			w.failed = false
			o.exitFollowFailed(w, exitOrderFollowExpired)
			delete(o.exitFills, key)
			continue
		}
		if !w.resolved {
			intent, err := o.opts.Journal.LookupIntent(ctx, w.line.intentID)
			if err != nil {
				o.exitFollowFailed(w, exitOrderIntentLookupFailed)
				if errors.Is(err, journal.ErrIntentNotFound) {
					// 원장에 신원이 없는 주문은 따라갈 길이 없음 — 한 번 알리고 끝냄.
					delete(o.exitFills, key)
				}
				continue
			}
			w.scope = journal.FillSnapshotScope{
				OrderID: id, AccountRef: intent.AccountRef, Market: intent.Market,
				TradingDay: intent.TradingDay, Symbol: intent.Symbol, Side: intent.Side,
			}
			w.resolved = true
		}
		rec, err := o.opts.Journal.LookupFillScoped(ctx, w.scope)
		if errors.Is(err, journal.ErrFillNotFound) {
			// 체결 감지가 아직 이 주문을 못 봄 — 다음 주기에 다시 봄.
			continue
		}
		if err != nil {
			o.exitFollowFailed(w, exitOrderFillLookupFailed)
			continue
		}
		w.failed = false
		grew, cerr := riskcalc.CompareDecimal(rec.FilledQuantity, w.filled)
		switch {
		case cerr == nil && grew > 0:
			// 누적 체결이 늘었음 — 같은 누적의 재관측 · 평균가만 바뀐 정정은 줄이 아님.
			w.filled = rec.FilledQuantity
			o.writeExitOrder(obs.EventFillObserved, false, w.line.filled(rec))
		case rec.Terminal:
			// 늘지 않은 채 끝남 — 한 주도 못 판(또는 더 못 판) 손절이 여기서 보임.
			o.writeExitOrder(eventExitOrderEnded, true, w.line.filled(rec))
		}
		if rec.Terminal {
			delete(o.exitFills, key)
		}
	}
}

// exitFollowFailed 는 따라가기 실패를 주문당 한 번, 원문 오류 없이 씀.
func (o *ExitObserver) exitFollowFailed(w *exitFillWatch, reason string) {
	if w.failed {
		return
	}
	w.failed = true
	o.writeExitOrder(eventExitFillFollowFailed, true, w.line.because(reason))
}

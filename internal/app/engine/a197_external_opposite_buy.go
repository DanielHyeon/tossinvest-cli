package engine

// a197_external_opposite_buy.go 는 손절을 막는 외부 매수 미체결을 엔진이 취소하는 단계임(a197 design D1~D11).
//
// # 흐름
//
//   - 판정 경로(submit 의 default 갈래)는 「보호 손절이 반대 미체결로 확정 거절됐다」를 대기열에 넣기만 함 — 브로커 호출 · 원장 쓰기 0(D1).
//   - 관측 주기 끝(endExitOrderCycle 첫 문장)이 대기열에서 포지션 하나를 골라 처리함 — 같은 주기의 다른 포지션 판정을 늦추지 않음(D5).
//     대가로 이 단계가 도는 시간만큼 다음 주기 시작이 밀림(Run 이 ObserveOnce 반환 뒤 Sleep — D4 · D5 공시).
//   - 처리: 실행 직전 재검증(M-3) → 같은 종목 엔진 발주 park 가드(D1) → 미체결 목록 1 회 판독(D3 · D4) → 대상 고르기(D2 · D7) →
//     「시도」 critical(D6) → 취소 결정 발행 · 제출(a198 과 같은 경로) → 결과별 처리(접수 / 확정 거절 / 모호면 그 자리에서 park — D11).
//
// # 지키는 것
//
//   - 토글이 꺼져 있으면 대기열에 넣지도 않음 — 조회 0 · 취소 0 · park 0(D8).
//   - 매도는 누구의 것이든 대상이 아님. 행 하나라도 모르면 그 주기 취소 0(D3). 주문 번호는 행의 바이트 그대로 씀(D2).
//   - 같은 번호에는 생애 한 번(원장의 CANCEL 행이 표시 — 거래일 · 상태 무관, M-1).
//   - park 는 계정 진입 래치를 걸지 않음(D11) — 기존 원장 전이(ResolveUnresolved)만 씀.

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudget"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
)

// A197AttemptKeyPrefix · A197UnclearedKeyPrefix 는 단계 알림 key 의 접두임 — key = 접두 + 포지션 + "|" + 주문 번호(D6 · M-1).
// 새 critical 종류는 만들지 않음: a091 수 핀이 critical 종류 수를 고정하므로 기존 critical 종류 exit.liquidation_delayed 에 key 로 가름.
const (
	A197AttemptKeyPrefix   = "exit.liquidation_delayed|a197-attempt|"
	A197UnclearedKeyPrefix = "exit.liquidation_delayed|a197-uncleared|"
)

const (
	// externalBuyListTimeout 은 미체결 목록 판독 전체의 시한임(D4 다) — R7 단건 GET p99(≤ 0.6 s)의 5 배, 관측 max 4.17 s 보다 짧음.
	// 관측 주기(a205 부터 2 초)와는 무관한 값임. 목록 판독만의 시한이고 단계 전체 시한은 없음(handleExternalBuy 의 시한 주석) — 이 단계가 도는 주기는 단계 체류만큼 다음 주기 시작이 밀림(파일 머리 D4 · D5 공시).
	externalBuyListTimeout = 3 * time.Second
	// externalBuyMaxPages 는 판독의 쪽 상한임 — OPEN 은 실측상 전량 1 요청(M7)이라 넘으면 「읽지 못함」(D3).
	externalBuyMaxPages = 5
)

// externalBuyItem 은 주기 끝에 처리할 「열 조건」 하나임 — 판정 때의 포지션 신원(M-3 재검증 재료).
type externalBuyItem struct {
	positionID  string
	instanceSeq int64
	market      string
	symbol      string
}

// externalBuyRow 는 미체결 행 하나의 해석 결과임. orderID 는 행의 바이트 그대로(정규화 금지 — D2).
type externalBuyRow struct {
	orderID   string
	buy       bool
	remaining float64
	price     float64
	currency  string
}

// externalBuyOpens 는 열 조건임(D1): 보호 손절 ∧ 확정 거절 ∧ 사유가 반대 미체결. 익절 · 다른 사유 · 브로커 전 거절은 열지 않음.
func externalBuyOpens(p exitpolicy.Proposal, out execgw.Outcome) bool {
	return isProtective(p) && out.State == journal.StateFailedConfirmed && out.Reason == execgw.ReasonOppositePendingOrder
}

// noteExternalBuyBlock 은 판정 경로에서 열 조건을 대기열에 넣기만 함 — 토글이 꺼져 있으면 넣지 않음(조회 0 의 근원).
func (o *ExitObserver) noteExternalBuyBlock(m managed) {
	if !o.opts.ExternalBuyCancel {
		return
	}
	for _, q := range o.externalBuyQueue {
		if q.positionID == m.position.ID {
			return
		}
	}
	o.externalBuyQueue = append(o.externalBuyQueue, externalBuyItem{
		positionID:  m.position.ID,
		instanceSeq: m.position.InstanceSeq,
		market:      strings.ToLower(strings.TrimSpace(m.position.Market)),
		symbol:      strings.ToUpper(strings.TrimSpace(m.position.Symbol)),
	})
}

// externalBuyItemCurrent 는 실행 직전 재검증임(Manager 결정 M-3): 같은 포지션이 여전히 열려 있고 판정 때와 같은 인스턴스인가.
// 판정식은 작업 집합(workingSet)의 탈락 조건과 같음 — 보호 필요성은 다시 판정하지 않음(판정을 둘로 두지 않음).
func externalBuyItemCurrent(item externalBuyItem, positions []journal.Position) bool {
	for _, p := range positions {
		if p.ID != item.positionID {
			continue
		}
		return p.InstanceSeq == item.instanceSeq && p.State != journal.PositionClosed && !isZeroQuantity(p.Quantity)
	}
	return false
}

// runExternalBuyStage 는 주기 끝에서 대기열의 포지션 하나를 처리함(D4 나 — 주기당 하나 · 취소 1 건). 대기열은 언제나 비움 — 나머지는
// 다음 거절 주기가 다시 넣음. 단계 안의 패닉은 거둬 관측 루프를 내리지 않음(손절 루프 보호 — 사실은 로그로 남김).
func (o *ExitObserver) runExternalBuyStage(ctx context.Context) {
	queue := o.externalBuyQueue
	o.externalBuyQueue = nil
	// 토글 판정은 대기열에 넣는 자리 하나(noteExternalBuyBlock) — 판정을 둘로 두지 않음.
	if len(queue) == 0 || ctx.Err() != nil {
		return
	}
	item := o.pickExternalBuyItem(queue)
	defer func() {
		if r := recover(); r != nil {
			o.externalBuyLog(item, true, "the external-buy stage panicked and was contained; nothing more is attempted this cycle")
		}
	}()
	o.handleExternalBuy(ctx, item)
}

// pickExternalBuyItem 은 가장 오래 처리되지 않은 포지션을 고르고 그 시각을 적음 — 여러 포지션이 막혀도 차례가 돌게 함.
func (o *ExitObserver) pickExternalBuyItem(queue []externalBuyItem) externalBuyItem {
	if o.externalBuyTried == nil {
		o.externalBuyTried = map[string]time.Time{}
	}
	best := queue[0]
	for _, q := range queue[1:] {
		if o.externalBuyTried[q.positionID].Before(o.externalBuyTried[best.positionID]) {
			best = q
		}
	}
	o.externalBuyTried[best.positionID] = o.clk.Now()
	return best
}

func (o *ExitObserver) handleExternalBuy(ctx context.Context, item externalBuyItem) {
	// 보호 손절 경로의 선례와 같은 우선순위(exitloop 제출 · 길 치우기) — 계좌 예산의 Critical 등급.
	ctx = accountbudget.WithPriority(ctx, accountbudget.Critical)

	positions, err := o.opts.Journal.Positions(ctx, o.opts.AccountRef)
	if err != nil {
		o.externalBuyLog(item, true, "the positions could not be re-read before the external-buy stage; nothing is cancelled")
		return
	}
	if !externalBuyItemCurrent(item, positions) {
		// 판정 뒤 주기 끝 전에 닫혔거나 세대가 바뀜 — 취소 권한의 근거가 사라졌으므로 버림(다음 주기의 거절이 다시 넣음).
		return
	}
	parked, err := o.engineOrderParkedOn(ctx, item)
	if err != nil || parked {
		// 번호를 모르는 엔진 발주가 브로커에 살아 있을 수 있음 — 외부로 오분류할 수 있으므로 열지 않음(D1, r1 #2).
		o.externalBuyLog(item, false, "an engine order on this symbol is parked or unreadable; the external-buy stage does not open")
		return
	}
	if o.opts.OpenOrders == nil {
		o.externalBuyLog(item, true, "no open-order reader is wired; the external-buy stage cannot read and cancels nothing")
		return
	}

	// 단계 전체를 감싸는 ctx 시한은 두지 않음 — 취소 제출을 그 시한으로 끊으면 결과 기록까지 실패해 DISPATCH_STARTED 가 남음(구현 리뷰 r1 #1).
	// 판독은 자기 시한(3 s), 사전 단건 조회 · 취소는 어댑터 시한(2 s · 15 s)이 맡음(design D4).
	rows, ok := o.readExternalBuyRows(ctx, item)
	if !ok {
		o.externalBuyLog(item, true, "the open-order list could not be read completely; nothing is cancelled this cycle")
		return
	}
	target, found, err := o.pickExternalBuyTarget(ctx, item, rows)
	if err != nil {
		o.externalBuyLog(item, true, "the engine attribution or the issued-cancel record could not be read; nothing is cancelled")
		return
	}
	if !found {
		o.externalBuyLog(item, false, "the open-order list holds no external buy left to cancel")
		return
	}

	// 취소를 내기 전에 기록 — 접수와 기록 사이의 중단으로 알림이 사라지지 않게(D6 · r1 #7), 결과 알림이 남지 못하는 중단에서도 인계되게(M-2).
	o.recordExternalBuyCritical(ctx, A197AttemptKeyPrefix+item.positionID+"|"+target.orderID, item, obs.Event{
		Title: o.label(item.symbol) + " 손절을 막는 외부 매수 주문의 취소를 시도한다",
		Body: fmt.Sprintf("포지션 %s 의 손절이 이 종목의 매수 미체결 주문 때문에 브로커에서 거절되어, 엔진이 그 매수 주문의 취소를 시도한다.\n"+
			"결과가 불명으로 남으면 이 종목의 신규 진입이 막히고 운영자 해동이 필요하다 — 절차는 운영 문서(a197 design §5).", item.positionID),
		Fields: map[string]any{obs.FieldSymbol: item.symbol, "position_id": item.positionID},
	})

	// 결정 발행과 취소 제출은 주기 ctx 로 — 게이트웨이는 전송과 결과의 원장 기록에 같은 ctx 를 쓰므로, 더 짧은 시한이 전송 도중 끝나면 결과
	// 기록(MarkInDoubt)까지 실패해 취소가 DISPATCH_STARTED 로 남고 그 종목의 손절이 재시작까지 막힘(구현 리뷰 r1 #1).
	decision, err := (&execgw.Issuer{Journal: o.opts.Journal, Clock: o.clk, AccountRef: o.opts.AccountRef}).
		IssueReduction(ctx, execgw.ReductionRequest{
			Kind:        journal.KindCancel,
			Market:      item.market,
			Symbol:      item.symbol,
			Side:        "BUY",
			MaxQuantity: target.remaining,
			Reason: fmt.Sprintf("cancelling the external BUY order %s that blocks the stop of position %s",
				target.orderID, item.positionID),
		})
	if err != nil {
		o.recordExternalBuyUncleared(ctx, item, target, "", false, "the cancel decision could not be recorded, so no cancel was sent")
		return
	}
	out, err := o.opts.Submit.Cancel(ctx, execgw.CancelRequest{
		Intent: orderintent.CancelIntent{OrderID: target.orderID, Symbol: item.symbol},
		Order: execgw.OrderRef{
			Market:   item.market,
			Side:     "BUY",
			Quantity: target.remaining,
			Price:    target.price,
			Currency: target.currency,
		},
		Decision: decision,
	})
	switch {
	case err == nil && out.State == journal.StateConfirmed:
		// 매수의 취소 접수는 치움 — 다음 주기의 손절이 나감(정본 exit-policy 의 같은 귀결).
		o.externalBuyLog(item, false, "the external buy was cancelled; the next cycle submits the stop")
	case out.State == journal.StateInDoubt && out.AttemptID != "":
		o.parkExternalBuyCancel(ctx, item, target, out.AttemptID)
	case externalBuyCancelHeld(out):
		// 원장 쓰기가 전송 앞뒤에서 실패해 취소 attempt 가 종결되지 않음 — 그 attempt 가 이 종목의 손절까지 막으므로 「앱에서 취소하면 나간다」가
		// 거짓이 됨(gstack 리뷰 r2 — state-machines · adversarial #1).
		o.recordExternalBuyUncleared(ctx, item, target, out.AttemptID, true,
			fmt.Sprintf("the cancel's journal record did not reach a final state (%q %s)", out.State, out.Reason))
	default:
		o.recordExternalBuyUncleared(ctx, item, target, out.AttemptID, false,
			fmt.Sprintf("the cancel was not accepted (%s %s); the engine does not cancel this order again", out.State, out.Reason))
	}
}

// externalBuyCancelHeld 는 취소 attempt 가 기록됐지만 종결 상태가 아닌지임 — 종결 전 attempt 는 게이트웨이의 종목 검사가 그 종목의 모든
// 변경(손절 포함)을 재시작 복구까지 막음. IN_DOUBT 는 앞 갈래(park)가 먼저 받음.
func externalBuyCancelHeld(out execgw.Outcome) bool {
	return out.AttemptID != "" && !out.State.IsTerminal()
}

// engineOrderParkedOn 은 그 종목에 판정 불능으로 park 된 엔진 발주 · 정정이 있는지임 — 기존 판독 둘(UnresolvedAttempts · LookupIntent)만 씀.
func (o *ExitObserver) engineOrderParkedOn(ctx context.Context, item externalBuyItem) (bool, error) {
	recs, err := o.opts.Journal.UnresolvedAttempts(ctx)
	if err != nil {
		return false, err
	}
	for _, rec := range recs {
		if rec.Kind != journal.KindPlace && rec.Kind != journal.KindAmend {
			continue
		}
		intent, err := o.opts.Journal.LookupIntent(ctx, rec.IntentID)
		if err != nil {
			return false, err
		}
		if strings.TrimSpace(intent.AccountRef) == strings.TrimSpace(o.opts.AccountRef) &&
			strings.EqualFold(strings.TrimSpace(intent.Market), item.market) &&
			strings.EqualFold(strings.TrimSpace(intent.Symbol), item.symbol) {
			return true, nil
		}
	}
	return false, nil
}

// readExternalBuyRows 는 종목 필터 OPEN 목록을 끝까지 읽고 행마다 해석함. 하나라도 못 읽으면 false(D3) — 오류 · 시한 초과 ·
// 「다음 쪽 있음 ∧ 커서 없음」(ScanOrders 는 이 갈래를 성공으로 돌려주므로 쓰지 않음 — r1 #5) · 커서 반복 · 쪽 상한 · 미지 행.
func (o *ExitObserver) readExternalBuyRows(ctx context.Context, item externalBuyItem) ([]externalBuyRow, bool) {
	listCtx, cancel := context.WithTimeout(ctx, externalBuyListTimeout)
	defer cancel()
	var raws []json.RawMessage
	cursor := ""
	seen := map[string]bool{}
	complete := false
	for page := 0; page < externalBuyMaxPages; page++ {
		p, err := o.opts.OpenOrders.OrdersPageRaw(listCtx, execgw.OrderQuery{Status: "OPEN", Symbol: item.symbol}, cursor)
		if err != nil {
			return nil, false
		}
		raws = append(raws, p.Orders...)
		if !p.HasNext {
			complete = true
			break
		}
		if p.NextCursor == "" || seen[p.NextCursor] {
			return nil, false
		}
		seen[p.NextCursor] = true
		cursor = p.NextCursor
	}
	if !complete {
		return nil, false
	}
	rows := make([]externalBuyRow, 0, len(raws))
	for _, raw := range raws {
		row, ok := parseExternalBuyRow(raw, item.symbol)
		if !ok {
			return nil, false
		}
		rows = append(rows, row)
	}
	return rows, true
}

// parseExternalBuyRow 는 행 하나를 해석함. 종목 · 방향 · 번호는 모든 행에서, 수량 · 체결 수량 · 통화는 매수 행에서 읽혀야 함 —
// 조회의 종목 필터를 믿지 않고 행 값으로 다시 대조함(D2, r1 #6).
func parseExternalBuyRow(raw json.RawMessage, symbol string) (externalBuyRow, bool) {
	var payload struct {
		OrderID   string  `json:"orderId"`
		Symbol    string  `json:"symbol"`
		Side      string  `json:"side"`
		Quantity  *string `json:"quantity"`
		Price     *string `json:"price"`
		Currency  string  `json:"currency"`
		Execution *struct {
			FilledQuantity *string `json:"filledQuantity"`
		} `json:"execution"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return externalBuyRow{}, false
	}
	if strings.TrimSpace(payload.OrderID) == "" || !strings.EqualFold(strings.TrimSpace(payload.Symbol), symbol) {
		return externalBuyRow{}, false
	}
	side := strings.ToUpper(strings.TrimSpace(payload.Side))
	switch side {
	case "SELL":
		return externalBuyRow{orderID: payload.OrderID}, true
	case "BUY":
	default:
		return externalBuyRow{}, false
	}
	if payload.Quantity == nil || payload.Execution == nil || payload.Execution.FilledQuantity == nil ||
		strings.TrimSpace(payload.Currency) == "" {
		return externalBuyRow{}, false
	}
	quantity, err := floatOf("open order quantity", *payload.Quantity)
	if err != nil {
		return externalBuyRow{}, false
	}
	filled, err := floatOf("open order filled quantity", *payload.Execution.FilledQuantity)
	if err != nil || quantity-filled <= 0 {
		return externalBuyRow{}, false
	}
	price := ""
	if payload.Price != nil {
		price = *payload.Price
	}
	limit, err := workingOrderPrice(price)
	if err != nil {
		return externalBuyRow{}, false
	}
	return externalBuyRow{
		orderID:   payload.OrderID,
		buy:       true,
		remaining: quantity - filled,
		price:     limit,
		currency:  strings.ToUpper(strings.TrimSpace(payload.Currency)),
	}, true
}

// pickExternalBuyTarget 은 행 순서의 첫 대상을 고름: 매수 ∧ 엔진 귀속 아님(상태 무관 귀속 — D2) ∧ 이 단계가 아직 내지 않은 번호(D7 · M-1).
func (o *ExitObserver) pickExternalBuyTarget(ctx context.Context, item externalBuyItem, rows []externalBuyRow) (externalBuyRow, bool, error) {
	engineIDs, err := o.opts.Journal.EngineBrokerOrderIDs(ctx, o.opts.AccountRef)
	if err != nil {
		return externalBuyRow{}, false, err
	}
	for _, row := range rows {
		if !row.buy || engineIDs[row.orderID] {
			continue
		}
		issued, err := o.opts.Journal.ExternalCancelIssued(ctx, o.opts.AccountRef, item.market, item.symbol, row.orderID, engineSource)
		if err != nil {
			return externalBuyRow{}, false, err
		}
		if issued {
			continue
		}
		return row, true, nil
	}
	return externalBuyRow{}, false, nil
}

// parkExternalBuyCancel 은 모호(IN_DOUBT)로 돌아온 취소를 그 자리에서 park 함(D11) — 기존 전이(ResolveUnresolved)만, 계정 래치 없음.
// 쓰기가 실패하면 IN_DOUBT 로 남아 재시작 복구가 해소함(공시된 잔여).
func (o *ExitObserver) parkExternalBuyCancel(ctx context.Context, item externalBuyItem, target externalBuyRow, attemptID string) {
	attempt, err := o.opts.Journal.Resume(ctx, attemptID)
	if err == nil {
		err = attempt.ResolveUnresolved(ctx, journal.ReasonResolutionExhausted,
			"the cancel of an external buy that blocked a stop came back in doubt; parked at once so the symbol's stop is not held until a restart (a197 D11)")
	}
	if err != nil {
		o.recordExternalBuyUncleared(ctx, item, target, attemptID, true,
			"the cancel came back in doubt and could not be parked")
		return
	}
	o.recordExternalBuyCritical(ctx, A197UnclearedKeyPrefix+item.positionID+"|"+target.orderID, item, obs.Event{
		Title: o.label(item.symbol) + " 외부 매수 취소의 결과가 불명이다 — 운영자 해동 필요",
		Body: fmt.Sprintf("포지션 %s 의 손절을 막던 매수 주문의 취소 결과가 불명이라 그 취소를 판정 불능(park)으로 넘겼다. 손절 · 청산은 막지 않는다.\n"+
			"이 종목의 신규 진입은 운영자 해동까지 막힌다. 앱에서 그 매수의 상태를 확인한 뒤 "+
			"`tossctl engine attempt-resolve --attempt %s --target … --operator … --approval … --note …` 로 닫는다 "+
			"(취소 attempt 에서 --target 의 뜻은 운영 문서 a197 design §5). 엔진은 그 주문을 다시 취소하지 않는다.", item.positionID, attemptID),
		Fields: map[string]any{obs.FieldSymbol: item.symbol, "position_id": item.positionID, "attempt_id": attemptID},
	})
}

// recordExternalBuyUncleared 는 접수되지 않은 취소를 사람에게 넘김(D6 두 번째 critical).
//
// 「다시 취소하지 않는다」는 취소 시도가 원장에 기록된 경우(attempt 있음 — D7 의 표시)에만 참임. 결정 발행 실패 · 게이트웨이의 기록 전 거절은
// 시도가 없어 다음 거절 주기에 다시 대상이 되므로 문장과 key 를 가름(구현 리뷰 r1 #3) — 같은 key 로 래치되어 뒤의 실제 결과 알림이 묻히지 않게.
// held 는 엔진의 취소 attempt 가 종결되지 않아 그 종목의 손절까지 막고 있음 — 이때 「앱에서 취소하면 나간다」는 거짓이라 재시작을 말함(gstack 리뷰 r2).
func (o *ExitObserver) recordExternalBuyUncleared(ctx context.Context, item externalBuyItem, target externalBuyRow, attemptID string, held bool, why string) {
	fields := map[string]any{obs.FieldSymbol: item.symbol, "position_id": item.positionID}
	key := A197UnclearedKeyPrefix + item.positionID + "|" + target.orderID
	next := "엔진은 그 주문을 다시 취소하지 않는다 — 앱에서 그 매수 주문을 취소해야 손절이 나간다."
	if attemptID != "" {
		fields["attempt_id"] = attemptID
		if held {
			next = "엔진의 취소 시도가 끝나지 않아 이 종목의 손절 · 청산은 재시작 복구가 그 시도를 해소할 때까지 막힌다 — 앱에서 매수를 취소해도 손절은 나가지 않는다. " +
				"엔진은 그 주문을 다시 취소하지 않는다."
		}
	} else {
		key += "|not-sent"
		next = "취소 시도가 기록되지 않아 다음 거절 주기에 엔진이 다시 시도할 수 있다 — 앱에서 그 매수 주문을 취소하면 손절이 나간다."
	}
	o.recordExternalBuyCritical(ctx, key, item, obs.Event{
		Title:  o.label(item.symbol) + " 손절을 막는 외부 매수 주문이 치워지지 않았다",
		Body:   fmt.Sprintf("포지션 %s 의 손절을 막는 매수 주문을 엔진이 치우지 못했다(%s).\n%s", item.positionID, why, next),
		Fields: fields,
	})
}

// recordExternalBuyCritical 은 알림기의 critical 단일 입구로 창 0 기록함(a094 와 같은 입구 · 같은 래치). 알림이 꺼진 엔진은 critical 을
// 쓰지 않고 로그 한 줄만 — 보낼 수 없는 critical 이 진입을 멈추지 않게(a091 · a095 와 같은 규칙).
func (o *ExitObserver) recordExternalBuyCritical(ctx context.Context, key string, item externalBuyItem, e obs.Event) {
	e.Type = obs.EventExitLiquidationDelayed
	e.Key = key
	if !o.opts.NotificationsEnabled {
		o.externalBuyLog(item, true, e.Title)
		return
	}
	if o.episodeLatched(key) {
		return
	}
	o.recordEpisodeCritical(ctx, key, e)
}

// externalBuyLog 는 단계의 결과 줄을 전용 로거(계좌 필드 없음)에 남김 — 주기 끝이라 판정 경로 안 쓰기가 아님(a198 D4 핀과 같은 규칙).
func (o *ExitObserver) externalBuyLog(item externalBuyItem, warn bool, detail string) {
	if o.opts.ExitOrderLog == nil {
		return
	}
	args := []any{"origin", "exit", "stage", "a197_external_buy", "position_id", item.positionID, obs.FieldSymbol, item.symbol, obs.FieldDetail, detail}
	if warn {
		o.opts.ExitOrderLog.Warn(obs.EventExitLiquidationDelayed, args...)
		return
	}
	o.opts.ExitOrderLog.Event(obs.EventExitLiquidationDelayed, args...)
}

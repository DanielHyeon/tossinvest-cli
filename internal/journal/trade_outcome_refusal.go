package journal

// trade_outcome_refusal.go 는 a130 의 원장 쪽 새 코드임 — 성과 동결 거절의 사유 코드 타입, 그 보고, 편입 시기 검사.
//
// # 왜 사유 코드인가 (design D2)
//
// 성과 동결은 실패해도 체결을 되돌리지 않도록 오류를 돌려주지 않음(trade_outcomes.go 머리 주석). 그 대가로 거절이 bool 하나로 접혀
// ApplyExitFill 이 버렸고, 엔진 매도로 닫힌 편입 포지션 여럿이 아무 흔적 없이 성과 행을 잃었음. 그래서 판정 함수가 거절마다 이 파일의
// 타입 하나를 돌려줌 — 가능한 값은 아래 상수 집합이 전부이고, 오류 원문 · 계좌 값은 싣지 않음. 사유를 문구에서 유도하지 않음.
//
// # 왜 커밋 뒤에만 알리는가 (design D3)
//
// 거절은 체결 트랜잭션 안에서 핸들(ApplyTx)에 쌓이기만 하고, RecordFill 이 커밋에 성공한 뒤에 싱크로 넘김. 롤백된 종결은 알리지 않으므로
// 거짓 줄이 없음. 쌓기 · 알리기 모두 오류를 돌려주지 않고 싱크의 패닉을 거둠 — 분석이 체결 반영을 되돌리거나 막을 길이 없음.
// 싱크(ApplyHooks.OutcomeRefused)가 nil 이면 거절 경로에서 아무것도 읽지도 쌓지도 부르지도 않음(a130 이전과 같음).
//
// # 편입 시기 검사와 묶음별 편입가 (design D1, 사용자 결정 2026-10-13 3차)
//
// 편입 포지션의 매수 다리는 편입 규칙(A7: 기준선은 매입 원가가 아니라 편입 시점 관측가)을 **묶음마다** 같은 규칙으로 적용함:
// 편입 수량 × 편입 관측가 + 편입 뒤 늘어난 묶음마다 (증가분 × 그 조정 시각이 든 브로커 공식 1분봉의 종가). cost_basis 와 조정 행의
// 평균가(대사 수렴이 보내지 않아 기존 평균가를 이어 적은 값)는 쓰지 않음. 분봉은 네트워크 읽기이므로 체결 트랜잭션 안에서는 읽지 않음 —
// 그 안에서는 늘어난 묶음이 있는 포지션을 needs_intake_price 로 거절하고, 사람이 돌리는 backfill 이 트랜잭션 밖에서 분봉을 읽어 넘김.
// 그 분 봉이 없으면 가격을 지어내지 않고 거절(no_intake_candle). 재편입은 exit 상태의 주당 위험을 다시 쓰므로 거절, 편입 뒤 감소(사람의
// 처분)와 이 인스턴스 귀속 밖의 로컬 매도는 flatten 쓸어 담기와 우연히 맞아 조작된 행을 만들 수 있으므로 거절.
// 조정은 position_id 로 이 인스턴스에 명시 귀속되고, 시각은 같은 인스턴스의 사건 순서와 분봉 선택, 그리고 **거절 조건**에만 씀
// (시각으로 숫자를 귀속하지 않음 — 시간창 귀속 금지 계약).

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"strings"
	"time"
)

// TradeOutcomeRefusalReason 은 성과를 동결하지 못한 사유임. 빈 값 = 동결됨.
type TradeOutcomeRefusalReason string

// 사유 코드 — 가능한 값의 전부. 영소문자와 밑줄만.
const (
	OutcomeRefusalCostModelUnconfigured   TradeOutcomeRefusalReason = "cost_model_unconfigured"
	OutcomeRefusalPositionUnreadable      TradeOutcomeRefusalReason = "position_unreadable"
	OutcomeRefusalUnmanagedPosition       TradeOutcomeRefusalReason = "unmanaged_position"
	OutcomeRefusalExitStateUnreadable     TradeOutcomeRefusalReason = "exit_state_unreadable"
	OutcomeRefusalRoundTripUnpriceable    TradeOutcomeRefusalReason = "round_trip_unpriceable"
	OutcomeRefusalEmptyBuyLeg             TradeOutcomeRefusalReason = "empty_buy_leg"
	OutcomeRefusalCostEstimateFailed      TradeOutcomeRefusalReason = "cost_estimate_failed"
	OutcomeRefusalAdoptionUnreadable      TradeOutcomeRefusalReason = "adoption_unreadable"
	OutcomeRefusalReadopted               TradeOutcomeRefusalReason = "readopted"
	OutcomeRefusalNeedsIntakePrice        TradeOutcomeRefusalReason = "needs_intake_price"
	OutcomeRefusalNoIntakeCandle          TradeOutcomeRefusalReason = "no_intake_candle"
	OutcomeRefusalUnattributedLocalSell   TradeOutcomeRefusalReason = "unattributed_local_sell"
	OutcomeRefusalForeignInstanceSell     TradeOutcomeRefusalReason = "foreign_instance_sell"
	OutcomeRefusalPostAdoptionDisposal    TradeOutcomeRefusalReason = "post_adoption_disposal"
	OutcomeRefusalAdoptionEpochUnreadable TradeOutcomeRefusalReason = "adoption_epoch_unreadable"
	OutcomeRefusalEngineSellUnreadable    TradeOutcomeRefusalReason = "engine_sell_unreadable"
	OutcomeRefusalNoEngineSell            TradeOutcomeRefusalReason = "no_engine_sell"
	OutcomeRefusalSellUnpriced            TradeOutcomeRefusalReason = "sell_unpriced"
	OutcomeRefusalSellQuantityMismatch    TradeOutcomeRefusalReason = "sell_quantity_mismatch"
	OutcomeRefusalInsertFailed            TradeOutcomeRefusalReason = "insert_failed"
)

// TradeOutcomeRefusalReasons 는 사유 집합임(시험 · 운영 문서 대조용).
func TradeOutcomeRefusalReasons() []TradeOutcomeRefusalReason {
	return []TradeOutcomeRefusalReason{
		OutcomeRefusalCostModelUnconfigured, OutcomeRefusalPositionUnreadable, OutcomeRefusalUnmanagedPosition,
		OutcomeRefusalExitStateUnreadable, OutcomeRefusalRoundTripUnpriceable, OutcomeRefusalEmptyBuyLeg,
		OutcomeRefusalCostEstimateFailed, OutcomeRefusalAdoptionUnreadable, OutcomeRefusalReadopted,
		OutcomeRefusalNeedsIntakePrice, OutcomeRefusalNoIntakeCandle, OutcomeRefusalPostAdoptionDisposal, OutcomeRefusalAdoptionEpochUnreadable,
		OutcomeRefusalEngineSellUnreadable, OutcomeRefusalNoEngineSell, OutcomeRefusalSellUnpriced,
		OutcomeRefusalSellQuantityMismatch, OutcomeRefusalUnattributedLocalSell, OutcomeRefusalForeignInstanceSell,
		OutcomeRefusalInsertFailed,
	}
}

// TradeOutcomeRefusal 은 성과를 동결하지 못한 종결 하나임. 필드가 곧 내보낼 수 있는 값의 전부 — 계좌 참조와 포지션 id(계좌에서
// 유도된 해시)는 없음.
type TradeOutcomeRefusal struct {
	Symbol      string
	Market      string
	InstanceSeq int64
	Reason      TradeOutcomeRefusalReason
}

// OutcomeRefusalBound 는 거절 싱크가 묶였는지 알려 줌 — 배선 시험용(ProjectionBound 와 같은 모양).
func (j *Journal) OutcomeRefusalBound() bool {
	j.applyMu.RLock()
	defer j.applyMu.RUnlock()
	return j.applyHooks.OutcomeRefused != nil
}

// outcomeRefusalsKey 는 RecordFill 이 체결 트랜잭션 동안 거절을 모으는 목록을 ctx 에 싣는 열쇠임. 시그니처가 같은 runApplyHooks 를
// 지나 ApplyExitFill 까지 닿게 함(runApplyHooks 의 다른 호출자 — 매수만 되메우는 전략 경로 · 시험 — 는 목록이 없어 아무것도 모으지 않음).
type outcomeRefusalsKey struct{}

// withOutcomeRefusals 는 거절을 모을 목록을 ctx 에 실음.
func withOutcomeRefusals(ctx context.Context, into *[]TradeOutcomeRefusal) context.Context {
	return context.WithValue(ctx, outcomeRefusalsKey{}, into)
}

// noteOutcomeRefusal 은 체결 트랜잭션 안에서 거절 하나를 RecordFill 의 목록에 쌓음 — 알리지는 않음(커밋 뒤 reportOutcomeRefusals).
// 싱크가 없거나 목록이 없으면 읽지도 쌓지도 않음. 종목 · 시장 · 회차 읽기가 실패해도 사유만으로 쌓음. 오류를 돌려주지 않음.
func (a *ApplyTx) noteOutcomeRefusal(ctx context.Context, positionID string, reason TradeOutcomeRefusalReason) {
	if a == nil || a.outcomeRefused == nil || reason == "" {
		return
	}
	into, _ := ctx.Value(outcomeRefusalsKey{}).(*[]TradeOutcomeRefusal)
	if into == nil {
		return
	}
	refusal := TradeOutcomeRefusal{Reason: reason}
	if rows, err := a.Query(ctx,
		`SELECT market, symbol, instance_seq FROM positions WHERE id = ?`, positionID); err == nil {
		if rows.Next() {
			if err := rows.Scan(&refusal.Market, &refusal.Symbol, &refusal.InstanceSeq); err != nil {
				refusal = TradeOutcomeRefusal{Reason: reason}
			}
		}
		_ = rows.Close()
	}
	*into = append(*into, refusal)
}

// reportOutcomeRefusals 는 커밋에 성공한 체결 트랜잭션이 쌓은 거절을 싱크에 넘김. 호출마다 싱크의 패닉을 거둠.
func (j *Journal) reportOutcomeRefusals(refusals []TradeOutcomeRefusal) {
	if len(refusals) == 0 {
		return
	}
	j.applyMu.RLock()
	sink := j.applyHooks.OutcomeRefused
	j.applyMu.RUnlock()
	if sink == nil {
		return
	}
	for _, r := range refusals {
		func() {
			defer func() { _ = recover() }()
			sink(r)
		}()
	}
}

// IntakeSourceBrokerCandle1m 는 묶음 편입가의 출처 이름임 — 브로커 공식 GET /api/v1/candles 1분봉(adjusted=false)의 종가.
const IntakeSourceBrokerCandle1m = "broker_candle_1m"

// IntakeLot 은 편입 뒤 늘어난 묶음 하나임. 포지션 id 는 내보내지 않음(Key 는 그것을 감춘 해시).
type IntakeLot struct {
	Market    string
	Symbol    string
	CreatedAt time.Time
	Quantity  string

	key string
}

// Key 는 이 묶음을 IntakePrices 에서 찾는 열쇠임(포지션 id · 조정 시각 · 수량의 해시 — 원문은 드러나지 않음).
func (l IntakeLot) Key() string { return l.key }

// BarLabel 은 이 묶음의 조정 시각이 든 1분봉의 이름(봉 **종료** 시각)임. 공식 분봉의 timestamp 는 봉이 닫히는 시각이고 t 라는 봉은
// [t−1분, t) 를 덮음(internal/official/candle_reads.go — a112 결정 30 · 31 실측). 그래서 조정 시각을 분으로 내리고 1분을 더함.
func (l IntakeLot) BarLabel() time.Time {
	return l.CreatedAt.UTC().Truncate(time.Minute).Add(time.Minute)
}

// IntakeCandle 은 묶음 하나에 쓴 분봉임.
type IntakeCandle struct {
	Label  time.Time
	Close  string
	Source string
}

// IntakePrices 는 묶음 열쇠 → 그 묶음의 분봉임. 분봉이 없는 묶음은 넣지 않음(→ no_intake_candle).
type IntakePrices map[string]IntakeCandle

// adoptionEpoch 은 편입 시점 기준이 편입 뒤 어떻게 바뀌었는지 봄(design D1). 사유가 비어 있지 않으면 거절. 우선순위: 재편입 >
// 편입 뒤 감소. 거절이 아니면 편입 뒤 늘어난 묶음 목록을 돌려줌.
//
//	readopted               exit 상태의 lifecycle_generation > 1(재편입이 READOPTED 이벤트와 함께 씀) — 주당 위험이 다시 쓰였음
//	post_adoption_disposal  조정 기록 시각 ≥ 편입 관측 시각 이고 수량이 줄었음
//
// 같은 초의 증가는 편입 뒤로 세지 않음(fold-in 0→q 와 편입 직전 증가는 편입 수량에 들어 있음 — 틀리면 매도 수량 불일치로 거절됨),
// 같은 초의 감소는 안전 쪽으로 거절함. 시각은 RFC3339(소수 초 허용)로 읽고 문자열로 견주지 않음.
func adoptionEpoch(ctx context.Context, r tradeOutcomeReader, positionID string) ([]IntakeLot, TradeOutcomeRefusalReason) {
	// 재편입은 exit 상태의 lifecycle_generation 을 편입 세대(≥ 2)로 다시 씀 — READOPTED exit 이벤트와 같은 트랜잭션(resetExitStateForReadoptTx).
	// exit_events 를 훑지 않고 exit 상태 한 줄(기본 키)만 읽음: 포지션 하나에 exit 이벤트가 수만 줄일 수 있어(사본 최대 27,071 — 그 훑기가
	// 체결 트랜잭션 안 판정을 20 ms 대로 늘렸음, 리뷰 4차 P2-4 실측).
	readopted, err := r.Query(ctx,
		`SELECT 1 FROM exit_states WHERE position_id = ? AND coalesce(lifecycle_generation, 1) > 1`, positionID)
	if err != nil {
		return nil, OutcomeRefusalAdoptionEpochUnreadable
	}
	found := readopted.Next()
	if err := readopted.Err(); err != nil {
		_ = readopted.Close()
		return nil, OutcomeRefusalAdoptionEpochUnreadable
	}
	_ = readopted.Close()
	if found {
		return nil, OutcomeRefusalReadopted
	}

	var lots []IntakeLot
	disposed := false
	rows, err := r.Query(ctx, `
		SELECT adj.prev_quantity, adj.new_quantity, adj.created_at, a.observed_at, p.market, p.symbol
		  FROM position_adjustments adj
		  JOIN positions p ON p.id = adj.position_id
		  JOIN position_adoptions a ON a.id = p.adoption_id
		 WHERE adj.position_id = ?
		 ORDER BY adj.created_at, adj.rowid`, positionID)
	if err != nil {
		return nil, OutcomeRefusalAdoptionEpochUnreadable
	}
	for rows.Next() {
		var prevText, newText, createdText, observedText, market, symbol string
		if err := rows.Scan(&prevText, &newText, &createdText, &observedText, &market, &symbol); err != nil {
			_ = rows.Close()
			return nil, OutcomeRefusalAdoptionEpochUnreadable
		}
		adoptedAt, err1 := time.Parse(time.RFC3339Nano, strings.TrimSpace(observedText))
		createdAt, err2 := time.Parse(time.RFC3339Nano, strings.TrimSpace(createdText))
		prev, okPrev := new(big.Rat).SetString(strings.TrimSpace(prevText))
		next, okNext := new(big.Rat).SetString(strings.TrimSpace(newText))
		if err1 != nil || err2 != nil || !okPrev || !okNext {
			_ = rows.Close()
			return nil, OutcomeRefusalAdoptionEpochUnreadable
		}
		switch next.Cmp(prev) {
		case 1:
			// 엄격히 뒤의 증가만 셈 — 같은 초의 증가는 편입 수량에 들어 있을 수 있음.
			if createdAt.After(adoptedAt) {
				delta := ratText(new(big.Rat).Sub(next, prev))
				sum := sha256.Sum256([]byte(positionID + "\x1f" + strings.TrimSpace(createdText) + "\x1f" + delta))
				lots = append(lots, IntakeLot{Market: normaliseMarket(market), Symbol: normaliseSymbol(symbol),
					CreatedAt: createdAt, Quantity: delta, key: hex.EncodeToString(sum[:16])})
			}
		case -1:
			// 같은 초 포함 — 순서를 가를 수 없는 감소는 안전 쪽으로 거절함.
			if !createdAt.Before(adoptedAt) {
				disposed = true
			}
		}
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, OutcomeRefusalAdoptionEpochUnreadable
	}
	_ = rows.Close()
	if disposed {
		return nil, OutcomeRefusalPostAdoptionDisposal
	}
	return lots, ""
}

// priceLots 는 묶음들에 분봉 종가를 매겨 수량 합 · 대금 합 · 쓴 분봉을 돌려줌. 묶음이 없으면 0 · 0. 분봉 집합이 nil 이면(체결 트랜잭션 —
// 네트워크 금지) needs_intake_price, 묶음 하나라도 분봉이 없거나 이름이 그 분이 아니면 no_intake_candle.
func priceLots(lots []IntakeLot, prices IntakePrices) (quantity, notional *big.Rat, applied []IntakeCandle,
	reason TradeOutcomeRefusalReason) {
	quantity, notional = new(big.Rat), new(big.Rat)
	if len(lots) == 0 {
		return quantity, notional, nil, ""
	}
	if prices == nil {
		return quantity, notional, nil, OutcomeRefusalNeedsIntakePrice
	}
	for _, lot := range lots {
		candle, ok := prices[lot.Key()]
		if !ok || !candle.Label.Equal(lot.BarLabel()) {
			return new(big.Rat), new(big.Rat), nil, OutcomeRefusalNoIntakeCandle
		}
		price, okPrice := new(big.Rat).SetString(strings.TrimSpace(candle.Close))
		delta, okDelta := new(big.Rat).SetString(lot.Quantity)
		if !okPrice || !okDelta || price.Sign() <= 0 {
			return new(big.Rat), new(big.Rat), nil, OutcomeRefusalNoIntakeCandle
		}
		quantity.Add(quantity, delta)
		notional.Add(notional, new(big.Rat).Mul(delta, price))
		applied = append(applied, candle)
	}
	return quantity, notional, applied, ""
}

// instanceSellIntegrity 는 매도 다리가 이 인스턴스의 매도만 담았는지 봄(리뷰 R3-3). 시각은 **거절 조건**에만 씀 — 숫자는 여전히 명시
// 참조(exit 이벤트 사슬 · flatten 단계)로 귀속된 주문에서만 나옴. 쓰는 시각은 주문이 확정된 시각(mutation_attempts.settled_at — 원장
// 시계)이고 체결의 committed_at 이 아님: committed_at 은 「소유 확정보다 엄격히 뒤」 로 밀려 기록될 수 있어(journalTimeStrictlyAfter)
// 인스턴스 경계를 넘어 보일 수 있음.
//
//	foreign_instance_sell    귀속된 주문의 확정이 이 인스턴스가 열리기(opened_at) 전이거나 닫힌(closed_at) 뒤 — 심볼 단위 flatten 갈래가
//	                         이전 · 이후 인스턴스의 매도를 쓸어 담았음
//	unattributed_local_sell  편입 관측 뒤(엄격) · 닫히기 전에 확정된, 체결이 있는 이 계좌 · 시장 · 심볼의 로컬 매도 주문이 귀속된 주문이
//	                         아님 — 사람이 낸 로컬 매도처럼 엔진 귀속 밖의 처분이 있었음
//
// 수명 끝이 비어 있으면(아직 열린 포지션 — 판정 함수를 직접 부른 경우) 끝이 없다고 봄. 같은 초 경계는 귀속 밖 주문만 수명 밖으로 읽음 —
// 앞 인스턴스의 마지막 매도가 새 인스턴스 시작과 한 초에 들 수 있어서. 그 초에 확정된 귀속 밖 매도는 이 검사가 못 보고 매도 수량 =
// 매수 수량 검사만 남음(design D5 잔여).
func instanceSellIntegrity(ctx context.Context, r tradeOutcomeReader, positionID, account, market, symbol string,
	orders []FillSnapshotScope) TradeOutcomeRefusalReason {
	var openedText, closedText, adoptedText string
	life, err := r.Query(ctx, `
		SELECT coalesce(p.opened_at, ''), coalesce(p.closed_at, ''), coalesce(a.observed_at, '')
		  FROM positions p LEFT JOIN position_adoptions a ON a.id = p.adoption_id
		 WHERE p.id = ?`, positionID)
	if err != nil {
		return OutcomeRefusalEngineSellUnreadable
	}
	if !life.Next() {
		_ = life.Close()
		return OutcomeRefusalEngineSellUnreadable
	}
	if err := life.Scan(&openedText, &closedText, &adoptedText); err != nil {
		_ = life.Close()
		return OutcomeRefusalEngineSellUnreadable
	}
	_ = life.Close()
	opened, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(openedText))
	if err != nil {
		return OutcomeRefusalEngineSellUnreadable
	}
	var closed time.Time
	if strings.TrimSpace(closedText) != "" {
		if closed, err = time.Parse(time.RFC3339Nano, strings.TrimSpace(closedText)); err != nil {
			return OutcomeRefusalEngineSellUnreadable
		}
	}
	notAfterClose := func(t time.Time) bool { return closed.IsZero() || !t.After(closed) }
	// 귀속 밖 로컬 매도를 세기 시작하는 순간은 편입 관측 시각임(인스턴스 시작이 아님) — fold-in 과 편입 사이의 로컬 매도는 편입 수량에 이미
	// 반영돼 있어 이 왕복의 처분이 아님(리뷰 4차 P3-8).
	unattributedFrom := opened
	if strings.TrimSpace(adoptedText) != "" {
		adoptedAt, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(adoptedText))
		if err != nil {
			return OutcomeRefusalEngineSellUnreadable
		}
		if adoptedAt.After(unattributedFrom) {
			unattributedFrom = adoptedAt
		}
	}

	attributed := map[string]bool{}
	for _, scope := range orders {
		scope = canonicalFillSnapshotScope(scope)
		attributed[scope.TradingDay+"\x1f"+scope.OrderID] = true
	}
	// 이 계좌 · 시장 · 심볼의 체결이 있는 로컬 매도 주문마다, 그 주문을 확정한 가장 이른 시각.
	rows, err := r.Query(ctx, allFillEventsCTE+`
		SELECT f.order_id, TRIM(f.trading_day), MIN(a.settled_at)
		  FROM all_fill_events f
		  JOIN mutation_attempts a ON a.broker_order_id = f.order_id
		  JOIN intents i ON i.id = a.intent_id
		 WHERE UPPER(TRIM(f.side)) = 'SELL' AND TRIM(f.account_ref) = ?
		   AND LOWER(TRIM(f.market)) = ? AND UPPER(TRIM(f.symbol)) = ?
		   AND a.state = 'CONFIRMED' AND a.kind IN ('PLACE','AMEND')
		   AND TRIM(i.account_ref) = TRIM(f.account_ref) AND LOWER(TRIM(i.market)) = LOWER(TRIM(f.market))
		   AND TRIM(i.trading_day) = TRIM(f.trading_day)
		   AND UPPER(TRIM(i.symbol)) = UPPER(TRIM(f.symbol)) AND UPPER(TRIM(i.side)) = 'SELL'
		 GROUP BY f.order_id, TRIM(f.trading_day)`,
		strings.TrimSpace(account), normaliseMarket(market), normaliseSymbol(symbol))
	if err != nil {
		return OutcomeRefusalEngineSellUnreadable
	}
	defer rows.Close()
	foreign, unattributed := false, false
	for rows.Next() {
		var orderID, tradingDay, settledText string
		if err := rows.Scan(&orderID, &tradingDay, &settledText); err != nil {
			return OutcomeRefusalEngineSellUnreadable
		}
		settled, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(settledText))
		if err != nil {
			return OutcomeRefusalEngineSellUnreadable
		}
		isAttributed := attributed[tradingDay+"\x1f"+orderID]
		switch {
		case isAttributed && (settled.Before(opened) || !notAfterClose(settled)):
			// 열리기 전 또는 닫힌 뒤 확정 — 심볼 단위 flatten 갈래가 이전 · 이후 인스턴스의 매도를 쓸어 담았음(리뷰 4차 P2-3: 이후 쪽은
			// backfill 이 늦게 돌면 나중 인스턴스의 flatten 이 섞임).
			foreign = true
		case !isAttributed && settled.After(unattributedFrom) && notAfterClose(settled):
			unattributed = true
		}
	}
	if err := rows.Err(); err != nil {
		return OutcomeRefusalEngineSellUnreadable
	}
	switch {
	case foreign:
		return OutcomeRefusalForeignInstanceSell
	case unattributed:
		return OutcomeRefusalUnattributedLocalSell
	}
	return ""
}

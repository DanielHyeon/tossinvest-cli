package journal

// a200_adjustment_close_fill_recognition.go 는 수렴 조정이 닫은 인스턴스에 늦게 닿은 엔진 청산 체결의 「인정」 임(a200 design D1~D3).
//
// 대사의 수렴은 계좌 수량이 0 으로 읽히면 조정으로 인스턴스를 닫음. 그 감소를 만든 엔진 손절 매도의 체결이 조정 커밋 **뒤에**
// 도착하면 전이표는 닫힌 인스턴스의 매도(SELL_ON_CLOSED)로 거절하고, 오늘은 그 거절이 운영자만 풀 수 있는 귀속 실패 차단이 됨.
// 여기서는 그 거절 하나를 「조정이 먼저 반영한 감소의 설명」 으로 받아들일지 판정함. 받아들이면 투영 수량은 0 그대로(조정이 이미
// 뺐으므로 다시 빼지 않음), 차단을 쓰지 않고, exit 이력에 표지 한 줄을 남김.
//
// 판정이 읽는 것은 조정 행 · 그 intent 의 시도 행 · intent 수량 · exit 상태의 발의뿐 — 체결 누적 표(체결 사건 · 체결 스냅숏)는
// 읽지 않음. 누적은 원장에서 적게 셀 길이 있어서(float 차를 0 으로 버림 · 호환 표 범위 충돌 · 작은 감소 덮어쓰기 — codex r1 · r2)
// 주문 수량이라는 정적 상한으로 대신함(Manager 결정 3). 인과(이 체결이 그 조정이 본 감소를 일으켰는가)는 증명하지 않음 — 공시된
// 가정(design §2 「인과는 …」).

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/position"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

// ExitEventAdjustmentCloseFillRecognised 는 인정 표지 action 임. 콘솔 이력에 원문으로 보이므로 기존 ADJUSTMENT_CLOSED ·
// PROPOSAL_FILLED 명명을 따름 — 「조정으로 닫힌 뒤 온 체결을 인정함」.
const ExitEventAdjustmentCloseFillRecognised = "ADJUSTMENT_CLOSE_FILL_RECOGNISED"

// convergenceAdjustmentKind 는 수렴(reconcile.Converger)이 쓰는 조정 종류 — 같은 분류 함수로 유도함(손으로 "UNKNOWN" 을 적지 않음).
var convergenceAdjustmentKind = string(position.Classify(position.ProvenanceInputs{LocalInstance: true}))

// recognitionFacts 는 판정이 보는 사실 전부임. 원장 읽기(readRecognitionFacts)와 판정(recognitionVerdict)을 갈라 판정을 순수하게 둠.
type recognitionFacts struct {
	Refusal position.Refusal
	// Adjustments 는 그 인스턴스의 조정 행(기록 순서).
	Adjustments []recognitionAdjustment
	// OriginIntentID 는 resolveFillOrigin 이 이미 증명한 소유 intent(다시 유도하지 않음).
	OriginIntentID string
	// ExitStateFound · PendingIntentID 는 그 인스턴스 exit 상태의 무장된 발의.
	ExitStateFound  bool
	PendingIntentID string
	// OrderAttempts 는 그 intent 의 PLACE · AMEND 시도 행(상태 무관).
	OrderAttempts []recognitionAttempt
	// IntentQuantity 는 그 intent 의 주문 수량(intents.quantity — 브로커에 나간 수량과 같음, claim-check 전제 (i)).
	IntentQuantity string
	// FillCumulative · FillOrdered 는 지금 적용 중인 체결 관측이 들고 온 브로커 값(AppliedFill 의 CumulativeQuantity ·
	// OrderedQuantity) — 원장에서 읽지 않음. FillOrdered 는 관측에 실리지 않으면 빈 문자열.
	FillCumulative, FillOrdered string
}

type recognitionAdjustment struct {
	Kind, PrevQuantity, NewQuantity string
}

type recognitionAttempt struct {
	Kind, State string
}

// recognisableRefusal 은 조건 (가)의 정의 한 곳임 — 닫힌 인스턴스의 매도 거절만 이 경합임. 순수 판정과 원장 읽기 앞의 조기 반환이
// 같은 술어를 부름(복사하면 판정이 둘이 되어 한쪽 변이가 다른 쪽에 가려짐).
func recognisableRefusal(r position.Refusal) bool {
	return r == position.RefusalSellOnClosed
}

// recognitionVerdict 는 인정 조건 넷(design D2)이 모두 참인지 판정함. 읽을 수 없는 decimal 은 「아님」(오늘의 귀속 실패로 남김).
func recognitionVerdict(f recognitionFacts) bool {
	// (가) 닫힌 인스턴스의 매도 거절만 — 다른 거절 유형은 이 경합이 아님.
	if !recognisableRefusal(f.Refusal) {
		return false
	}
	// (나) 그 인스턴스를 닫은 것이 수렴 조정: 수량 0 조정 행이 정확히 하나이고 그것이 마지막 행이며 수렴의 kind.
	// CLOSED 는 종결이라 그 뒤에 인스턴스를 움직인 체결 · 조정이 없음 — 조정 행과 체결 행의 순서를 비교할 필요가 없음.
	zeroRows, lastZero := 0, false
	var closing recognitionAdjustment
	for i, adj := range f.Adjustments {
		cmp, err := riskcalc.CompareDecimal(adj.NewQuantity, "0")
		if err != nil {
			return false
		}
		if cmp == 0 {
			zeroRows++
			closing = adj
			lastZero = i == len(f.Adjustments)-1
		}
	}
	if zeroRows != 1 || !lastZero || closing.Kind != convergenceAdjustmentKind {
		return false
	}
	// (다) 체결 주문의 소유 intent 가 그 인스턴스의 살아 있는 청산 발의.
	if strings.TrimSpace(f.OriginIntentID) == "" || !f.ExitStateFound || f.PendingIntentID != f.OriginIntentID {
		return false
	}
	// (라) 수량 상한: 그 intent 의 주문은 정정 없는 접수 확정 PLACE 하나이고, 주문 수량 ≤ 조정 감소분.
	// 체결 누적은 주문 수량을 넘지 못하므로 그 주문의 어떤 체결도 감소분 안에 있음. 시도가 둘 이상이면(재시도 · 정정) 주문이 둘일 수
	// 있어 상한이 아님 — 상태로 거르지 않음(in-doubt 로 남은 시도도 브로커 주문일 수 있음).
	if len(f.OrderAttempts) != 1 || f.OrderAttempts[0].Kind != string(KindPlace) ||
		f.OrderAttempts[0].State != string(StateConfirmed) {
		return false
	}
	decrease, err := riskcalc.SubDecimal(closing.PrevQuantity, closing.NewQuantity)
	if err != nil {
		return false
	}
	// 주문 수량은 양수여야 함 — 0 · 음수는 브로커가 접수하지 않으므로 생산에선 도달하지 않지만, 상한 비교만으로는 참이 되어 버림.
	positive, err := riskcalc.CompareDecimal(f.IntentQuantity, "0")
	if err != nil || positive <= 0 {
		return false
	}
	cmp, err := riskcalc.CompareDecimal(f.IntentQuantity, decrease)
	if err != nil || cmp > 0 {
		return false
	}
	// 같은 상한을 이 체결이 들고 온 브로커 값에도 검 — intent 와 다른 수량으로 접수 · 체결된 결함(어댑터 · 브로커)이 감소분을 넘는
	// 체결을 인정받지 못하게 함(gstack r1 G1). 정상에선 누적 ≤ 주문 수량 = intent 수량 ≤ 감소분이라 거절하는 정상 입력 없음.
	// 누적이 비었거나 못 읽으면 아님. 주문 수량은 관측에 실렸을 때만 봄.
	cumulative, err := riskcalc.CompareDecimal(f.FillCumulative, decrease)
	if err != nil || cumulative > 0 {
		return false
	}
	if strings.TrimSpace(f.FillOrdered) != "" {
		ordered, err := riskcalc.CompareDecimal(f.FillOrdered, decrease)
		if err != nil || ordered > 0 {
			return false
		}
	}
	return true
}

// recogniseAdjustmentCloseFill 은 ProjectPosition 의 거절 갈래에서 차단을 쓰기 직전에 한 번 불림. 인정이면 표지 행을 쓰고 true —
// 호출자는 차단을 쓰지 않고 투영도 쓰지 않음. 원장 읽기 오류는 그대로 돌려줌(체결 tx 롤백 — 오늘의 차단 쓰기 오류와 같은 방향).
func recogniseAdjustmentCloseFill(ctx context.Context, tx *ApplyTx, origin fillOrigin,
	instance projectedInstance, outcome position.Outcome, fill AppliedFill) (bool, error) {
	// 이 경합이 아닌 거절(OVERSELL · ENTRY_WHILE_CLOSING 등)은 원장을 읽지 않고 돌아감 — 그 거절들에 새 읽기 실패 경로를 만들지 않음.
	// (가)의 정의는 recognisableRefusal 한 곳. 인스턴스 행이 없으면(FLAT) 읽을 조정 · exit 상태도 없음.
	if !recognisableRefusal(outcome.Refusal) || !instance.Held || instance.ID == "" {
		return false, nil
	}
	facts, err := readRecognitionFacts(ctx, tx, origin, instance, outcome)
	if err != nil {
		return false, err
	}
	facts.FillCumulative, facts.FillOrdered = fill.CumulativeQuantity, fill.OrderedQuantity
	if !recognitionVerdict(facts) {
		return false, nil
	}
	if err := appendExitEventFromHook(ctx, tx, exitEventRow{
		PositionID: instance.ID, Action: ExitEventAdjustmentCloseFillRecognised, CreatedAt: tx.Now(),
	}); err != nil {
		return false, err
	}
	return true, nil
}

// readRecognitionFacts 는 판정에 쓸 사실을 체결 tx 안에서 읽음.
func readRecognitionFacts(ctx context.Context, tx *ApplyTx, origin fillOrigin,
	instance projectedInstance, outcome position.Outcome) (recognitionFacts, error) {
	facts := recognitionFacts{Refusal: outcome.Refusal, OriginIntentID: strings.TrimSpace(origin.IntentID)}

	rows, err := tx.Query(ctx, `SELECT kind, prev_quantity, new_quantity FROM position_adjustments
		WHERE position_id = ? ORDER BY created_at, rowid`, instance.ID)
	if err != nil {
		return facts, fmt.Errorf("journal: reading the adjustments of %s: %w", instance.ID, err)
	}
	for rows.Next() {
		var adj recognitionAdjustment
		if err := rows.Scan(&adj.Kind, &adj.PrevQuantity, &adj.NewQuantity); err != nil {
			_ = rows.Close()
			return facts, fmt.Errorf("journal: reading the adjustments of %s: %w", instance.ID, err)
		}
		facts.Adjustments = append(facts.Adjustments, adj)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return facts, fmt.Errorf("journal: reading the adjustments of %s: %w", instance.ID, err)
	}
	_ = rows.Close()

	state, err := tx.PendingState(ctx, instance.ID)
	switch {
	case err == nil:
		facts.ExitStateFound, facts.PendingIntentID = true, strings.TrimSpace(state.PendingIntentID)
	case errors.Is(err, ErrExitStateNotFound):
		// 무관리 인스턴스 — 발의가 없으므로 (다)에서 아님.
	default:
		return facts, err
	}
	if facts.OriginIntentID == "" {
		return facts, nil
	}

	attempts, err := tx.Query(ctx, `SELECT kind, state FROM mutation_attempts
		WHERE intent_id = ? AND kind IN ('PLACE','AMEND') ORDER BY rowid`, facts.OriginIntentID)
	if err != nil {
		return facts, fmt.Errorf("journal: reading the attempts of intent %s: %w", facts.OriginIntentID, err)
	}
	for attempts.Next() {
		var a recognitionAttempt
		if err := attempts.Scan(&a.Kind, &a.State); err != nil {
			_ = attempts.Close()
			return facts, fmt.Errorf("journal: reading the attempts of intent %s: %w", facts.OriginIntentID, err)
		}
		facts.OrderAttempts = append(facts.OrderAttempts, a)
	}
	if err := attempts.Err(); err != nil {
		_ = attempts.Close()
		return facts, fmt.Errorf("journal: reading the attempts of intent %s: %w", facts.OriginIntentID, err)
	}
	_ = attempts.Close()

	quantity, err := tx.Query(ctx, `SELECT quantity FROM intents WHERE id = ?`, facts.OriginIntentID)
	if err != nil {
		return facts, fmt.Errorf("journal: reading the quantity of intent %s: %w", facts.OriginIntentID, err)
	}
	defer quantity.Close()
	if quantity.Next() {
		if err := quantity.Scan(&facts.IntentQuantity); err != nil {
			return facts, fmt.Errorf("journal: reading the quantity of intent %s: %w", facts.OriginIntentID, err)
		}
	}
	return facts, quantity.Err()
}

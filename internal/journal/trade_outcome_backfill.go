package journal

// trade_outcome_backfill.go 는 a130 design D4 임 — 성과 행이 빠진 편입 포지션을 사람 승인으로 다시 계산해 채우는 두 단계.
//
//	⓪ 묶음 목록  IntakeLotsForBackfill — 늘어난 묶음마다 어느 1분봉이 필요한지(읽기 전용). 분봉 자체는 호출자(도구)가 브로커에서 읽어
//	            IntakePrices 로 넘김 — 원장 패키지는 네트워크를 부르지 않음.
//	① 미리보기  PreviewTradeOutcomeBackfill — 읽기 전용 연결(*ReadOnly: mode=ro + query_only)로 계획만 만듦. 쓰기 0.
//	② 실행      (*Journal).BackfillTradeOutcomes — 계획을 다시 계산해 토큰이 미리보기의 것과 같을 때만, 항목마다 짧은 쓰기 트랜잭션
//	            하나에서 (같은 분봉으로) 다시 계산한 값이 계획 값과 같을 때만 INSERT 함. 첫 차이에서 멈춤. 트랜잭션 안에 네트워크 없음.
//
// # 모집단 (리뷰 F-2)
//
// 「체결로 닫힌 편입 포지션 중 행이 없는 것」 만: CLOSED · 편입 기록 있음 · COMPLETED exit 이벤트 있음(체결로 닫힐 때
// markExitCompletedTx 만 씀 — 대사로 닫히면 ADJUSTMENT_CLOSED 가 completed=1 을 먼저 써서 COMPLETED 는 생기지 않으므로 둘은 한
// 인스턴스에서 배타적임, position_adjustments.go closeExitStateOnAdjustmentTx) · 보존 기간 밖은 표시만. 엔진 귀속은 판정의 매도 검사가
// 강제함(엔진 매도 수량 = 매수 수량 · instanceSellIntegrity). 엔진 진입 포지션은 넣지 않음 —
// 그 다리(roundTripLegs)는 매도 수량을 검사하지 않고 체결 창에 상한이 없어 닫는 트랜잭션 밖에서는 정확하지 않음(design D5).
//
// # 왜 항목마다 트랜잭션인가 (리뷰 F-7)
//
// 계획과 쓰기 사이에 원장이 움직일 수 있음(엔진이 돌고 있음). 그래서 쓰는 값은 「쓰는 트랜잭션 안에서 다시 계산한 값」 이고, 그것이
// 사람이 미리 본 값과 다르면 쓰지 않음. 트랜잭션은 항목 하나 동안만 쓰기 잠금을 쥠 — 계획 전체를 덮으면 손절 기록이 기다림(불변식 4).
//
// ⚠ 실행은 원장 쓰기(mutating)임. 에이전트와 자동 시험은 운영 원장에 실행하지 않음(시험은 임시 원장).
//
// 계획은 *ReadOnly 의 메서드가 아니라 패키지 함수임 — *ReadOnly 의 메서드 집합은 「쓰기 메서드 0」 시험이 재므로 늘리지 않음.
// 쓰지 않는다는 보장은 연결 자체(mode=ro + query_only)가 함.

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
)

// ErrBackfillTokenMismatch 는 실행 때 다시 계산한 계획의 토큰이 사람이 확인한 토큰과 다르다는 뜻임 — 아무것도 쓰지 않았음.
var ErrBackfillTokenMismatch = errors.New(
	"journal: the backfill plan changed since the preview (token mismatch); nothing was written — preview again")

// ErrBackfillItemMoved 는 쓰는 트랜잭션 안에서 다시 계산한 값이 계획 값과 달라 그 항목(과 그 뒤)을 쓰지 않았다는 뜻임.
var ErrBackfillItemMoved = errors.New(
	"journal: a planned outcome recomputed differently inside its write; it and the rest were not written — preview again")

// 계획 항목의 판정.
const (
	// BackfillActionFill 은 a130 판정이 동결하는 항목 — 실행이 채울 대상임.
	BackfillActionFill = "fill"
	// BackfillActionRefused 는 a130 판정도 거절하는 항목(Reason).
	BackfillActionRefused = "refused"
	// BackfillActionOutsideRetention 은 보존 기간 밖에서 종결된 항목 — 채우지 않음(보존 정리가 지운 것을 되살리지 않음).
	BackfillActionOutsideRetention = "outside_retention"
)

// TradeOutcomeBackfillItem 은 모집단 포지션 하나의 판정임. 계좌 참조는 없고 포지션 id 는 내보내지 않음.
type TradeOutcomeBackfillItem struct {
	Symbol      string
	Market      string
	InstanceSeq int64
	ClosedAt    string
	Action      string
	// Outcome 은 계산 결과(fill · outside_retention). Outcome.PositionID 는 출력하지 말 것.
	Outcome TradeOutcome
	// Reason 은 거절 사유(refused).
	Reason TradeOutcomeRefusalReason
	// IntakeCandles 는 편입 뒤 늘어난 묶음마다 편입가로 쓴 분봉임(fill 일 때). 늘어난 묶음이 없으면 비어 있음.
	IntakeCandles []IntakeCandle

	positionID string
}

// TradeOutcomeBackfillPlan 은 모집단 전부의 판정과 확인 토큰임.
type TradeOutcomeBackfillPlan struct {
	Items []TradeOutcomeBackfillItem
	Token string
}

// BackfillWritten 은 실행이 쓴 항목 하나와 그 쓰기 트랜잭션이 쓰기 잠금을 쥔 시간임(리뷰 R3-5 — 손절 기록이 기다린 상한).
type BackfillWritten struct {
	Market      string
	Symbol      string
	InstanceSeq int64
	LockHeld    time.Duration
}

// IntakeLotsForBackfill 은 모집단 항목 중 늘어난 묶음이 있는 것의 묶음 목록을 읽음(읽기 전용). 재편입 · 편입 뒤 감소로 거절되는 항목의
// 묶음은 넣지 않음 — 분봉을 읽을 필요가 없음.
func IntakeLotsForBackfill(ctx context.Context, ro *ReadOnly) ([]IntakeLot, error) {
	if ro == nil {
		return nil, fmt.Errorf("%w: no read-only journal", ErrInvalidRequest)
	}
	return intakeLotsForBackfill(ctx, readOnlyReader{ro: ro})
}

func intakeLotsForBackfill(ctx context.Context, r tradeOutcomeReader) ([]IntakeLot, error) {
	items, err := backfillPopulation(ctx, r)
	if err != nil {
		return nil, err
	}
	var lots []IntakeLot
	for _, it := range items {
		if found, reason := adoptionEpoch(ctx, r, it.positionID); reason == "" {
			lots = append(lots, found...)
		}
	}
	return lots, nil
}

// PreviewTradeOutcomeBackfill 은 읽기 전용 연결로 계획을 만듦. 원장에 쓰지 않음. prices 는 묶음별 분봉(도구가 브로커에서 읽은 것).
func PreviewTradeOutcomeBackfill(ctx context.Context, ro *ReadOnly, model costs.Model,
	now time.Time, prices IntakePrices) (TradeOutcomeBackfillPlan, error) {
	if ro == nil {
		return TradeOutcomeBackfillPlan{}, fmt.Errorf("%w: no read-only journal", ErrInvalidRequest)
	}
	return tradeOutcomeBackfillPlan(ctx, readOnlyReader{ro: ro}, model, now, prices)
}

// BackfillTradeOutcomes 는 계획을 다시 계산해 토큰이 같을 때만 채움 대상을 항목마다 하나의 짧은 트랜잭션으로 채움.
func (j *Journal) BackfillTradeOutcomes(ctx context.Context, model costs.Model, now time.Time,
	prices IntakePrices, token string) ([]BackfillWritten, error) {
	plan, err := tradeOutcomeBackfillPlan(ctx, journalReader{j: j}, model, now, prices)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(token) == "" || plan.Token != strings.TrimSpace(token) {
		return nil, ErrBackfillTokenMismatch
	}
	var written []BackfillWritten
	for _, item := range plan.Items {
		if item.Action != BackfillActionFill {
			continue
		}
		held, err := j.backfillPlannedOutcome(ctx, item, model, prices)
		if err != nil {
			return written, fmt.Errorf("journal: backfilling %s %s #%d: %w",
				item.Market, item.Symbol, item.InstanceSeq, err)
		}
		written = append(written, BackfillWritten{Market: item.Market, Symbol: item.Symbol,
			InstanceSeq: item.InstanceSeq, LockHeld: held})
	}
	return written, nil
}

// backfillPlannedOutcome 은 항목 하나를 쓰기 트랜잭션 하나에서 다시 계산해 계획 값과 같을 때만 쓰고, 잠금을 쥔 시간을 돌려줌.
func (j *Journal) backfillPlannedOutcome(ctx context.Context, item TradeOutcomeBackfillItem, model costs.Model,
	prices IntakePrices) (time.Duration, error) {
	tx, err := j.db.BeginTx(ctx, nil) // BEGIN IMMEDIATE — 항목 하나 동안만
	if err != nil {
		return 0, err
	}
	// 잠금을 얻은 뒤부터 잼 — 잠금을 기다린 시간은 쥔 시간이 아님(리뷰 구현 P3-4).
	start := time.Now()
	defer tx.Rollback()
	reader := txReader{tx: tx}
	if _, found, err := storedTradeOutcome(ctx, reader, item.positionID); err != nil {
		return time.Since(start), err
	} else if found {
		return time.Since(start), ErrTradeOutcomeExists
	}
	outcome, reason := computeTradeOutcome(ctx, reader, item.positionID, model, prices)
	if reason != "" || !sameOutcome(outcome, item.Outcome) {
		return time.Since(start), ErrBackfillItemMoved
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO trade_outcomes
		  (position_id, realized_pnl_after_costs, realized_r, initial_risk,
		   initial_quantity, held_seconds, exit_ratchet_level, exit_rung, closed_at,
		   cost_total)
		VALUES (?,?,?,?,?,?,?,?,?,?)`,
		outcome.PositionID, outcome.RealizedPnLAfterCosts, outcome.RealizedR,
		outcome.InitialRisk, outcome.InitialQuantity, outcome.HeldSeconds,
		nullableString(outcome.ExitRatchetLevel), nullableRung(outcome.ExitRung),
		outcome.ClosedAt, outcome.CostTotal); err != nil {
		if isUniqueViolation(err) {
			return time.Since(start), ErrTradeOutcomeExists
		}
		return time.Since(start), err
	}
	err = tx.Commit()
	return time.Since(start), err
}

// readOnlyReader 는 *ReadOnly 를, txReader 는 쓰기 트랜잭션을 tradeOutcomeReader 로 맞춤.
type readOnlyReader struct{ ro *ReadOnly }

func (r readOnlyReader) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return r.ro.db.QueryContext(ctx, query, args...)
}

type txReader struct{ tx *sql.Tx }

func (r txReader) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return r.tx.QueryContext(ctx, query, args...)
}

// backfillPopulation 은 모집단(머리 주석)을 종결 시각 · id 순으로 읽음.
func backfillPopulation(ctx context.Context, r tradeOutcomeReader) ([]TradeOutcomeBackfillItem, error) {
	rows, err := r.Query(ctx, `
		SELECT p.id, p.market, p.symbol, p.instance_seq, coalesce(p.closed_at, '')
		  FROM positions p
		 WHERE p.state = 'CLOSED'
		   AND p.adoption_id IS NOT NULL
		   AND NOT EXISTS (SELECT 1 FROM trade_outcomes o WHERE o.position_id = p.id)
		   AND EXISTS (SELECT 1 FROM exit_events e WHERE e.position_id = p.id AND e.action = ?)
		 ORDER BY coalesce(p.closed_at, ''), p.id`, ExitEventCompleted)
	if err != nil {
		return nil, fmt.Errorf("journal: listing the backfill population: %w", err)
	}
	defer rows.Close()
	var items []TradeOutcomeBackfillItem
	for rows.Next() {
		var it TradeOutcomeBackfillItem
		if err := rows.Scan(&it.positionID, &it.Market, &it.Symbol, &it.InstanceSeq, &it.ClosedAt); err != nil {
			return nil, fmt.Errorf("journal: listing the backfill population: %w", err)
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: listing the backfill population: %w", err)
	}
	return items, nil
}

// tradeOutcomeBackfillPlan 은 모집단마다 판정을 만듦.
func tradeOutcomeBackfillPlan(ctx context.Context, r tradeOutcomeReader, model costs.Model,
	now time.Time, prices IntakePrices) (TradeOutcomeBackfillPlan, error) {
	items, err := backfillPopulation(ctx, r)
	if err != nil {
		return TradeOutcomeBackfillPlan{}, err
	}
	cutoff := now.UTC().Add(-TradeOutcomeRetention)
	for i := range items {
		it := &items[i]
		it.Outcome, it.Reason = computeTradeOutcome(ctx, r, it.positionID, model, prices)
		if lots, reason := adoptionEpoch(ctx, r, it.positionID); reason == "" {
			if _, _, applied, reason := priceLots(lots, prices); reason == "" {
				it.IntakeCandles = applied
			}
		}
		switch {
		case it.Reason != "":
			it.Action = BackfillActionRefused
		case closedBefore(it.ClosedAt, cutoff):
			it.Action = BackfillActionOutsideRetention
		default:
			it.Action = BackfillActionFill
		}
	}
	return TradeOutcomeBackfillPlan{Items: items, Token: backfillToken(items, model)}, nil
}

// storedTradeOutcome 은 동결 행이 있는지 읽음.
func storedTradeOutcome(ctx context.Context, r tradeOutcomeReader, positionID string) (TradeOutcome, bool, error) {
	rows, err := r.Query(ctx, `
		SELECT position_id, realized_pnl_after_costs, realized_r, initial_risk,
		       initial_quantity, coalesce(held_seconds, 0), coalesce(exit_ratchet_level, ''),
		       exit_rung, closed_at, cost_total
		  FROM trade_outcomes WHERE position_id = ?`, positionID)
	if err != nil {
		return TradeOutcome{}, false, fmt.Errorf("journal: reading a frozen outcome: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		return TradeOutcome{}, false, rows.Err()
	}
	o, err := scanTradeOutcome(rows)
	return o, err == nil, err
}

// sameOutcome 은 두 계산이 같은 행을 만드는지 봄(모든 칸).
func sameOutcome(a, b TradeOutcome) bool {
	costA, costB := "", ""
	if a.CostTotal != nil {
		costA = *a.CostTotal
	}
	if b.CostTotal != nil {
		costB = *b.CostTotal
	}
	return a.PositionID == b.PositionID && a.RealizedPnLAfterCosts == b.RealizedPnLAfterCosts &&
		a.RealizedR == b.RealizedR && a.InitialRisk == b.InitialRisk && a.InitialQuantity == b.InitialQuantity &&
		costA == costB && (a.CostTotal == nil) == (b.CostTotal == nil) && a.HeldSeconds == b.HeldSeconds &&
		a.ExitRatchetLevel == b.ExitRatchetLevel && a.ExitRung == b.ExitRung && a.ClosedAt == b.ClosedAt
}

func closedBefore(closedAt string, cutoff time.Time) bool {
	t, err := time.Parse(time.RFC3339, strings.TrimSpace(closedAt))
	if err != nil {
		// 종결 시각을 읽을 수 없으면 보존 기간 안이라고 단정하지 않음 — 채우지 않음.
		return true
	}
	return t.Before(cutoff)
}

// backfillToken 은 채움 대상(값 · 쓴 분봉)과 비용 모델 · 스키마 버전의 정규 직렬화 SHA-256 앞 16 hex 임(리뷰 R3-13).
func backfillToken(items []TradeOutcomeBackfillItem, model costs.Model) string {
	h := sha256.New()
	fmt.Fprintf(h, "schema=%d\x1emodel=%#v\x1e", SchemaVersion, model)
	for _, it := range items {
		if it.Action != BackfillActionFill {
			continue
		}
		o := it.Outcome
		cost := ""
		if o.CostTotal != nil {
			cost = *o.CostTotal
		}
		fmt.Fprintf(h, "%s\x1f%s\x1f%s\x1f%s\x1f%s\x1f%s\x1f%d\x1f%s\x1f%d\x1f%s\x1e",
			it.positionID, o.RealizedPnLAfterCosts, o.RealizedR, o.InitialRisk, o.InitialQuantity,
			cost, o.HeldSeconds, o.ExitRatchetLevel, o.ExitRung, o.ClosedAt)
		for _, c := range it.IntakeCandles {
			fmt.Fprintf(h, "%s\x1f%s\x1f%s\x1d", c.Label.UTC().Format(time.RFC3339), c.Close, c.Source)
		}
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}

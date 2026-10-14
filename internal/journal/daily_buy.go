package journal

// daily_buy.go 는 하루 매수 총액 상한(a136, risk-management 「하루 매수 총액 한도」)의 판정임.
//
// # 왜 예약 트랜잭션 안인가
//
// 체인은 발급 시점에 한 번, 호출자가 모은 값으로 돎. 그 값으로 「오늘 합계」를 넘기면 동시 발급 둘이 같은 숫자를 보고
// 함께 통과함. 예약 트랜잭션은 단일 커넥션 BEGIN IMMEDIATE 라 직렬이므로, 합계를 이 트랜잭션 안에서 다시 읽어야
// 동시 매수가 함께 한도를 넘지 못함(design D3). 모든 예약 tx 는 reserveRows 를 지나므로 훅은 그 한 자리임.
//
// # 무엇을 세나 (design D2-c)
//
// 예약 원장 자체를 셈 — 금액은 발급 때 이미 account base · frozen FX · ceil · 비용 포함이라 새 환율 경로가 없음.
//
//   - 발급일과 무관하게 아직 HELD 인 OPEN_EXPOSURE 예약 = 미체결·예약된 매수.
//   - 오늘(clock.AccountTradingDay) 발급된 결정의 RELEASED OPEN_EXPOSURE 예약 중 「체결 없음이 증명되지 않은 것」 = 체결된 매수.
//     체결되면 보유는 해제되지만 행은 지워지지 않으므로 하루가 끝날 때까지 남음.
//
// 체결 없음의 증명은 셋뿐(proveZeroFill). 그 밖은 전부 셈 — 과대 집계(거부 쪽)로만 틀림.
//
// # 매도는 여기 오지 않음
//
// RISK_REDUCING 결정은 예약을 만들지 않음(IssueReduction). 이 판정은 EXPOSURE_RAISING 결정의 예약 tx 에서만 돎.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

// DailyBuyLimit 는 예약 요청에 실리는 하루 매수 총액 상한임. 금액은 account base 통화의 decimal 문자열.
type DailyBuyLimit struct {
	Amount   string
	Currency string
	// ManualBuy 는 엔진 밖 수동 매수(토스 앱 등)의 오늘 합계(사용자 결정 D9=b). 필수 — 없으면 거부.
	ManualBuy *ManualBuyTotal
}

// ManualBuyTotal 은 엔진 보조 실행자가 공식 주문 조회로 만든 수동 매수 스냅숏을 발급자가 base 로 환산한 값임.
type ManualBuyTotal struct {
	// Amount 는 account base 통화(= DailyBuyLimit.Currency) 의 decimal 문자열, 음수 불가.
	Amount string
	// AsOf 는 스냅숏이 브로커에서 읽힌 시각, Day 는 그 스냅숏이 센 계좌 하루 라벨.
	AsOf time.Time
	Day  string
}

// ManualBuyStaleness 는 수동 매수 스냅숏의 신선도 상한임. refresher 주기(30초)의 네 배 — 주기 한두 번 실패까지는 견디고,
// 그보다 오래 못 읽으면 매수를 거부함(fail-closed). 그 사이의 수동 매수 한 건은 늦게 반영될 수 있음(design D9 잔재).
const ManualBuyStaleness = 2 * time.Minute

// dailyBuyLookback 은 「오늘 발급」 후보를 거르는 거친 하한임. 계좌 하루는 24시간이라 48시간이면 넉넉한 상위집합이고,
// 정확한 소속은 행마다 clock.AccountTradingDay 로 다시 판정함.
const dailyBuyLookback = 48 * time.Hour

// checkDailyBuyNotional 은 오늘 합계 + 이번 요청의 OPEN_EXPOSURE 금액이 상한에 도달하면 거부함.
//
// 반환: 도달 → ErrReservationLimitExceeded(LIMIT_REACHED), 요청 결함 → ErrInvalidRequest,
// 합계 계산 불가 → 분류되지 않은 오류. 어느 쪽이든 호출자 tx 가 롤백되어 결정·예약이 남지 않음(fail-closed).
func checkDailyBuyNotional(ctx context.Context, tx *sql.Tx, req ReserveRequest, plan reservePlan, now time.Time) error {
	var safetyClass, issuedAt string
	if err := tx.QueryRowContext(ctx, "SELECT safety_class, issued_at FROM decisions WHERE id = ?", plan.decisionID).
		Scan(&safetyClass, &issuedAt); err != nil {
		return fmt.Errorf("journal: reading the safety class of decision %s for the daily buy ceiling: %w",
			plan.decisionID, err)
	}
	if safetyClass != string(SafetyClassExposureRaising) {
		// 노출을 올리지 않는 결정은 매수 총액과 무관함.
		return nil
	}
	if req.DailyBuyLimit == nil {
		// 상한을 빠뜨린 발급 문이 조용히 통과하지 못하게 함.
		return fmt.Errorf("%w: decision %s is exposure-raising but its reservation carries no daily buy ceiling",
			ErrInvalidRequest, plan.decisionID)
	}
	currency := strings.ToUpper(strings.TrimSpace(req.DailyBuyLimit.Currency))
	if currency == "" {
		return fmt.Errorf("%w: the daily buy ceiling carries no currency", ErrInvalidRequest)
	}
	limit, err := nonNegativeDecimal("daily buy ceiling", req.DailyBuyLimit.Amount)
	if err != nil {
		return err
	}
	if positive, err := riskcalc.CompareDecimal(limit, "0"); err != nil || positive <= 0 {
		return fmt.Errorf("%w: the daily buy ceiling is %q; an absent or non-positive ceiling is not an unlimited one",
			ErrInvalidRequest, req.DailyBuyLimit.Amount)
	}

	// 이번 요청이 더하는 매수 금액 = OPEN_EXPOSURE 행의 합(생산 발급 문 셋은 모두 이 행을 예약함). 그 행이 없는
	// 요청(DAILY_LOSS 만 예약하는 시험 모양)은 더하는 금액이 0 이지만, 오늘 합계가 이미 상한에 도달했으면 여전히 거부됨.
	added, ok := plan.newTotals[ReservationKindOpenExposure]
	if !ok {
		added = "0"
	}
	for _, row := range plan.rows {
		if row.Kind == ReservationKindOpenExposure && row.Currency != currency {
			return fmt.Errorf("%w: reservation %s is in %s but the daily buy ceiling is in %s; the journal does not convert currencies",
				ErrInvalidRequest, row.ID, row.Currency, currency)
		}
	}

	today, err := clock.AccountTradingDay(now)
	if err != nil {
		return fmt.Errorf("journal: computing the account day for the daily buy ceiling: %w", err)
	}
	// 매수는 발급일에 귀속됨. 08:00 KST 직전에 발급돼 직후에 커밋되면, 판정은 오늘로 하고 나중 합계는 어제로 분류되어
	// 오늘 합계에서 빠짐(freeze 리뷰 P2-1). 그래서 발급일과 커밋 순간의 하루가 다르면 거부함 — 다음 주기의 새 결정이 다시 판정.
	issued, err := parseJournalTime(issuedAt)
	if err != nil {
		return fmt.Errorf("journal: decision %s issued_at: %w", plan.decisionID, err)
	}
	issuedDay, err := clock.AccountTradingDay(issued)
	if err != nil {
		return fmt.Errorf("journal: computing the account day of decision %s: %w", plan.decisionID, err)
	}
	if issuedDay != today {
		return fmt.Errorf("%w: DAILY_BUY_NOTIONAL decision %s was issued on account day %s but reserves on %s; "+
			"a buy is counted on the day it is issued, so it must be re-decided",
			ErrReservationLimitExceeded, plan.decisionID, issuedDay, today)
	}
	spent, err := dailyBuyTotal(ctx, tx, plan.accountRef, currency, today, now)
	if err != nil {
		return err
	}
	manual, err := manualBuyAmount(req.DailyBuyLimit.ManualBuy, today, now)
	if err != nil {
		return err
	}
	if spent, err = riskcalc.AddDecimal(spent, manual); err != nil {
		return fmt.Errorf("journal: adding the manual buy total: %w", err)
	}
	total, err := riskcalc.AddDecimal(spent, added)
	if err != nil {
		return fmt.Errorf("journal: summing the daily buy total: %w", err)
	}
	cmp, err := riskcalc.CompareDecimal(total, limit)
	if err != nil {
		return fmt.Errorf("journal: comparing the daily buy total: %w", err)
	}
	// 총계 경계는 도달 시 거부(≥) — checkLimits 와 같은 규칙(design D4).
	if cmp >= 0 {
		return fmt.Errorf("%w: DAILY_BUY_NOTIONAL would reach %s %s against a limit of %s for account day %s (already %s, new %s)",
			ErrReservationLimitExceeded, total, currency, limit, today, spent, added)
	}
	return nil
}

// dailyBuyTotal 은 계좌의 오늘 매수 총액(HELD 전부 + 오늘 발급·체결 없음 미증명 RELEASED)을 정확한 decimal 로 더함.
func dailyBuyTotal(ctx context.Context, tx *sql.Tx, accountRef, currency, today string, now time.Time) (string, error) {
	// 두 갈래를 UNION 으로 나눠 각자 색인을 타게 함(freeze 리뷰 P2-8): HELD 행 + 최근 발급(idx_decisions_account 범위).
	// 진입 tx 가 쓰기 잠금을 잡는 동안 매도 기록이 기다리므로, 계좌 평생 행을 훑지 않는 것이 §0.3 간접 요구임.
	rows, err := tx.QueryContext(ctx,
		`SELECT r.id, r.decision_id, r.amount, r.currency, r.state, coalesce(r.release_reason, ''), coalesce(r.attempt_id, ''), d.issued_at
		   FROM risk_reservations r JOIN decisions d ON d.id = r.decision_id
		  WHERE r.account_ref = ? AND r.kind = ? AND r.state = ? AND d.safety_class = ?
		 UNION
		 SELECT r.id, r.decision_id, r.amount, r.currency, r.state, coalesce(r.release_reason, ''), coalesce(r.attempt_id, ''), d.issued_at
		   FROM decisions d JOIN risk_reservations r ON r.decision_id = d.id
		  WHERE d.account_ref = ? AND d.issued_at >= ? AND d.safety_class = ? AND r.kind = ? AND r.state <> ?`,
		accountRef, ReservationKindOpenExposure, ReservationHeld, string(SafetyClassExposureRaising),
		accountRef, formatJournalTime(now.Add(-dailyBuyLookback)), string(SafetyClassExposureRaising),
		ReservationKindOpenExposure, ReservationHeld)
	if err != nil {
		return "", fmt.Errorf("journal: reading today's buy reservations: %w", err)
	}
	type candidate struct {
		id, decisionID, amount, currency, state, releaseReason, attemptID, issuedAt string
	}
	var candidates []candidate
	for rows.Next() {
		var c candidate
		if err := rows.Scan(&c.id, &c.decisionID, &c.amount, &c.currency, &c.state, &c.releaseReason, &c.attemptID, &c.issuedAt); err != nil {
			rows.Close()
			return "", fmt.Errorf("journal: reading today's buy reservations: %w", err)
		}
		candidates = append(candidates, c)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return "", fmt.Errorf("journal: reading today's buy reservations: %w", err)
	}
	rows.Close()

	total := "0"
	for _, c := range candidates {
		if c.state != ReservationHeld {
			issued, err := parseJournalTime(c.issuedAt)
			if err != nil {
				return "", fmt.Errorf("journal: reservation %s decision issued_at: %w", c.id, err)
			}
			day, err := clock.AccountTradingDay(issued)
			if err != nil {
				return "", fmt.Errorf("journal: computing the account day of reservation %s: %w", c.id, err)
			}
			if day != today {
				// 다른 날 발급된 종결 매수는 오늘 합계가 아님.
				continue
			}
			proven, err := proveZeroFill(ctx, tx, accountRef, c.decisionID, c.attemptID, c.releaseReason)
			if err != nil {
				return "", err
			}
			if proven {
				continue
			}
		}
		// 읽을 수 없는 금액·다른 통화는 0 으로 두지 않고 판정 자체를 실패시킴(fail-closed).
		if strings.ToUpper(strings.TrimSpace(c.currency)) != currency {
			return "", fmt.Errorf("journal: reservation %s is in %s but the daily buy ceiling is in %s; the total cannot be computed",
				c.id, c.currency, currency)
		}
		amount, err := riskcalc.CanonicalDecimal(c.amount)
		if err != nil {
			return "", fmt.Errorf("journal: reservation %s amount %q: %w", c.id, c.amount, err)
		}
		if total, err = riskcalc.AddDecimal(total, amount); err != nil {
			return "", fmt.Errorf("journal: summing reservation %s: %w", c.id, err)
		}
	}
	return total, nil
}

// proveZeroFill 은 종결된 예약의 주문이 아무것도 체결하지 않았음을 원장만으로 증명할 수 있으면 true 를 돌려줌.
//
// 증명은 셋뿐(design 「D2-c 실측」):
//
//	Z1 attempt 가 없고 EXPIRED_UNCONSUMED 로 해제됨 — Prepare 가 없었고 nonce 가 쓰이지 않았음(아무것도 안 보냄).
//	Z2 attempt 가 NOT_DISPATCHED(프로세스를 떠나지 않음) 또는 FAILED_CONFIRMED(브로커가 확정 거절).
//	Z3 체결 경로 해제(BROKER_TERMINAL)이고, attempt 가 이 결정의 CONFIRMED PLACE 이며, 그 주문의 체결 스냅숏이 attempt 의
//	   intent 범위(계좌·시장·거래일·종목·방향)와 attempt 확정 뒤 커밋으로 정확히 한 행 · 종결(fail-closed 아님) · 누적 체결 0,
//	   그 범위의 소유 intent 가 하나, 그리고 그 주문의 체결 이벤트 · 대체 계보 · AMEND 시도가 없음.
//	   범위 술어는 sweepOrphanedTerminals(reservation_release.go)의 「정확한 범위의 종결」과 같음 — 이 해제를 만든 판정과 같은 조건.
//
// 증명이 안 되면 false(= 셈). 원장 읽기 오류만 오류로 돌려줌.
func proveZeroFill(ctx context.Context, tx *sql.Tx, accountRef, decisionID, attemptID, releaseReason string) (bool, error) {
	if strings.TrimSpace(attemptID) == "" {
		return releaseReason == ReleaseReasonExpiredUnconsumed, nil
	}
	var state, kind, orderID, attemptDecision string
	err := tx.QueryRowContext(ctx,
		"SELECT state, kind, broker_order_id, coalesce(decision_id, '') FROM mutation_attempts WHERE id = ?", attemptID).
		Scan(&state, &kind, &orderID, &attemptDecision)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("journal: reading attempt %s for the daily buy total: %w", attemptID, err)
	}
	switch AttemptState(state) {
	case StateNotDispatched, StateFailedConfirmed:
		return true, nil
	}
	if releaseReason != ReleaseReasonBrokerTerminal || AttemptState(state) != StateConfirmed ||
		MutationKind(kind) != KindPlace || attemptDecision != decisionID || orderID == "" {
		// OPERATOR 해제 · 다른 결정의 attempt · AMEND 등은 증명 아님 → 셈.
		return false, nil
	}

	rows, err := tx.QueryContext(ctx,
		`SELECT f.terminal, f.fail_closed, f.filled_quantity,
		        (SELECT count(DISTINCT oi.id)
		           FROM mutation_attempts oa JOIN intents oi ON oi.id = oa.intent_id
		          WHERE oa.state = ? AND oa.kind IN (?, ?) AND oa.broker_order_id = a.broker_order_id
		            AND TRIM(oi.account_ref) = TRIM(i.account_ref) AND LOWER(TRIM(oi.market)) = LOWER(TRIM(i.market))
		            AND TRIM(oi.trading_day) = TRIM(i.trading_day) AND UPPER(TRIM(oi.symbol)) = UPPER(TRIM(i.symbol))
		            AND UPPER(TRIM(oi.side)) = UPPER(TRIM(i.side)))
		   FROM mutation_attempts a
		   JOIN intents i ON i.id = a.intent_id
		   JOIN scoped_fill_snapshots f ON f.order_id = a.broker_order_id
		  WHERE a.id = ? AND TRIM(i.account_ref) = ?
		    AND a.settled_at < f.committed_at
		    AND f.account_ref = TRIM(i.account_ref) AND f.market = LOWER(TRIM(i.market))
		    AND f.trading_day = TRIM(i.trading_day) AND f.symbol = UPPER(TRIM(i.symbol))
		    AND f.side = UPPER(TRIM(i.side)) AND f.side = 'BUY'`,
		string(StateConfirmed), string(KindPlace), string(KindAmend), attemptID, accountRef)
	if err != nil {
		return false, fmt.Errorf("journal: reading the fill snapshot of order %s: %w", orderID, err)
	}
	var (
		matching, terminal, failClosed, owners int
		filled                                 string
	)
	for rows.Next() {
		matching++
		if err := rows.Scan(&terminal, &failClosed, &filled, &owners); err != nil {
			rows.Close()
			return false, fmt.Errorf("journal: reading the fill snapshot of order %s: %w", orderID, err)
		}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return false, fmt.Errorf("journal: reading the fill snapshot of order %s: %w", orderID, err)
	}
	rows.Close()
	// 정확히 한 행 · 종결 · fail-closed 아님 · 소유 intent 하나 — 하나라도 아니면 귀속 불가 → 셈.
	if matching != 1 || terminal != 1 || failClosed != 0 || owners != 1 {
		return false, nil
	}
	if zero, err := riskcalc.CompareDecimal(strings.TrimSpace(filled), "0"); err != nil || zero != 0 {
		return false, nil
	}

	// 누적 0 이어도 정정 · 대체 주문이 체결을 옮겨 갔을 수 있으면 증명 아님.
	var others int
	if err := tx.QueryRowContext(ctx,
		`SELECT (SELECT count(*) FROM fill_events WHERE order_id = ?)
		      + (SELECT count(*) FROM scoped_lineage_edges WHERE parent_order_id = ? OR child_order_id = ?)
		      + (SELECT count(*) FROM lineage_edges WHERE parent_order_id = ? OR child_order_id = ?)
		      + (SELECT count(*) FROM mutation_attempts WHERE target_order_id = ? AND kind = ?)`,
		orderID, orderID, orderID, orderID, orderID, orderID, string(KindAmend)).Scan(&others); err != nil {
		return false, fmt.Errorf("journal: reading the fill history of order %s: %w", orderID, err)
	}
	return others == 0, nil
}

// manualBuyAmount 는 수동 매수 스냅숏을 판정에 쓸 수 있는지 확인하고 금액을 돌려줌.
//
// 없음 → 요청 결함(ErrInvalidRequest — 발급 문이 빠뜨림). 다른 하루 · 상한보다 오래됨 · 미래 시각 → ErrSnapshotStale
// (재수집 루프가 발급자의 최신 스냅숏으로 다시 시도하고, 끝내 낡으면 SNAPSHOT_RECOLLECTION_EXHAUSTED 로 거부).
func manualBuyAmount(manual *ManualBuyTotal, today string, now time.Time) (string, error) {
	if manual == nil {
		return "", fmt.Errorf("%w: the daily buy ceiling carries no manual buy total; buys outside the engine "+
			"count toward the day and an unknown amount is not zero", ErrInvalidRequest)
	}
	if strings.TrimSpace(manual.Day) != today {
		return "", fmt.Errorf("%w: the manual buy snapshot counts account day %q, not %s", ErrSnapshotStale, manual.Day, today)
	}
	if manual.AsOf.IsZero() {
		return "", fmt.Errorf("%w: the manual buy snapshot carries no as-of", ErrSnapshotStale)
	}
	age := now.Sub(manual.AsOf.UTC())
	if age < 0 {
		return "", fmt.Errorf("%w: the manual buy snapshot is dated %s in the future", ErrSnapshotStale, -age)
	}
	if age > ManualBuyStaleness {
		return "", fmt.Errorf("%w: the manual buy snapshot is %s old, past the %s bound", ErrSnapshotStale, age, ManualBuyStaleness)
	}
	return nonNegativeDecimal("manual buy total", manual.Amount)
}

// EngineBrokerOrderIDs 는 이 계좌의 엔진 attempt 가 브로커에서 받은 주문 id 집합임(a136 수동 매수 중복 제거).
// 이 집합에 있는 주문은 예약 원장이 이미 세므로 수동 매수 스냅숏에서 빠짐. AMEND 가 만든 자식 주문 id 도 attempt 에 있으므로 포함됨.
func (j *Journal) EngineBrokerOrderIDs(ctx context.Context, accountRef string) (map[string]bool, error) {
	rows, err := j.db.QueryContext(ctx,
		`SELECT DISTINCT a.broker_order_id FROM mutation_attempts a JOIN intents i ON i.id = a.intent_id
		  WHERE TRIM(i.account_ref) = ? AND a.broker_order_id <> ''`, strings.TrimSpace(accountRef))
	if err != nil {
		return nil, fmt.Errorf("journal: listing engine broker order ids: %w", err)
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("journal: listing engine broker order ids: %w", err)
		}
		out[id] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: listing engine broker order ids: %w", err)
	}
	return out, nil
}

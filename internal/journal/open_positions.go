package journal

// open_positions.go 는 최대 보유 포지션 수(a138, risk-management 「최대 보유 포지션 수 한도」)의 판정임.
//
// # 왜 예약 트랜잭션 안인가
//
// a136 하루 매수와 같은 이유: 체인은 호출자가 모은 값으로 한 번 돌기 때문에 동시 진입 둘이 같은 「5 개 보유」를 보고 함께 통과할 수 있음.
// 예약 tx 는 단일 커넥션 BEGIN IMMEDIATE 라 직렬이므로, 그 안에서 원장을 다시 읽으면 두 번째 진입은 첫 번째의 HELD 예약을 봄.
//
// # 무엇을 세나 (사용자 결정 2026-10-15 D1 = B · D2 = scale-in 허용, freeze 리뷰 P1-1 · P1-2)
//
// 서로 다른 (시장, 종목) 의 집합 S:
//
//	① 계좌의 살아 있는 포지션 — 투영에서 CLOSED 가 아닌 행 전부(엔진 진입 · 편입 · 대사가 접어 넣은 외부 보유). 투영은 FLAT 행을 쓰지 않음.
//	② HELD 예약을 가진 EXPOSURE_RAISING 결정의 종목 — 예약 종류와 무관(Gateway 가 제출을 인가하는 술어와 같음). 아직 포지션이 안 된 진행 중 진입.
//	③ 이번 결정의 종목.
//	④ 종목 범위 활성 RECONCILE 중 ①~③ 에 없는 것 — 종결 체결이 예약을 해제한 뒤 투영이 거절되면 포지션 행 없이 이것만 남음.
//	   범위 시장이 있으면 (시장, 종목)으로, 없으면(전역 종목 범위) 종목만으로 맞추고, 맞는 것이 없으면 한 칸(과대 집계 쪽).
//
// |S| 가 상한을 **넘으면** 거부 — 상한은 허용되는 최대 개수(6 이면 6 은 허용). 이번 종목이 이미 ①·② 에 있으면 개수가 늘지 않으므로 통과.
// 종목은 결정의 저장된 preimage 에서 읽음(호출자 값 불신, 해석은 ParsePreimage 한 곳). 읽을 수 없으면 0 으로 두지 않고 판정 실패(fail-closed).
//
// # 매도는 여기 오지 않음
//
// RISK_REDUCING 결정은 예약을 만들지 않음(IssueReduction 은 RecordDecision 만). 이 판정도 노출을 올리지 않는 결정은 먼저 통과시킴.

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// OpenPositionLimit 는 예약 요청에 실리는 최대 보유 포지션 수임(a138). EXPOSURE_RAISING 결정의 예약에는 필수 — 없으면 거부.
type OpenPositionLimit struct {
	// Max 는 허용되는 서로 다른 (시장, 종목) 의 최대 개수 — 1 이상.
	Max int
}

// positionVenue 는 개수를 세는 단위 하나 — 정규화한 (시장, 종목).
type positionVenue struct{ market, symbol string }

func venueOf(market, symbol string) positionVenue {
	return positionVenue{market: normaliseMarket(market), symbol: normaliseSymbol(symbol)}
}

// checkOpenPositions 는 이번 진입이 들어가면 서로 다른 보유 종목 수가 상한을 넘을 때 거부함.
//
// 반환: 넘음 → ErrReservationLimitExceeded(LIMIT_REACHED), 요청 결함 → ErrInvalidRequest, 셀 수 없음 → 분류되지 않은 오류.
// 어느 쪽이든 호출자 tx 가 롤백되어 결정 · 예약이 남지 않음(fail-closed).
func checkOpenPositions(ctx context.Context, tx *sql.Tx, req ReserveRequest, plan reservePlan) error {
	var safetyClass, kind, preimage string
	if err := tx.QueryRowContext(ctx,
		"SELECT safety_class, preimage_kind, risk_preimage FROM decisions WHERE id = ?", plan.decisionID).
		Scan(&safetyClass, &kind, &preimage); err != nil {
		return fmt.Errorf("journal: reading decision %s for the open position maximum: %w", plan.decisionID, err)
	}
	if safetyClass != string(SafetyClassExposureRaising) {
		// 노출을 올리지 않는 결정은 보유 종목 수와 무관함.
		return nil
	}
	if req.OpenPositionLimit == nil || req.OpenPositionLimit.Max < 1 {
		// 상한을 빠뜨리거나 1 미만을 실은 발급 문이 조용히 통과하지 못하게 함 — 0 은 「무제한」이 아님.
		return fmt.Errorf("%w: decision %s is exposure-raising but its reservation carries no usable open position maximum",
			ErrInvalidRequest, plan.decisionID)
	}
	this, err := decisionVenue(plan.decisionID, kind, preimage)
	if err != nil {
		return err
	}

	held, err := heldVenues(ctx, tx, plan.accountRef)
	if err != nil {
		return err
	}
	if _, already := held[this]; already {
		// 이미 든 종목의 진입(scale-in) — 개수가 늘지 않음(D2). 지금 개수가 상한을 넘어 있어도(외부 보유 증가 등) 이 진입의 거부 사유가 아님.
		return nil
	}
	// 새 종목이 하나 더해짐. ④ RECONCILE 칸은 이번 종목과 맞으면 그 칸이 이번 종목 자신이므로 빼고 셈.
	reconciled, err := reconciledExtraSlots(ctx, tx, plan.accountRef, held, this)
	if err != nil {
		return err
	}
	if count := len(held) + reconciled + 1; count > req.OpenPositionLimit.Max {
		return fmt.Errorf("%w: MAX_OPEN_POSITIONS — a new %s/%s would make %d distinct positions against a maximum of %d "+
			"(%d live or held, %d reconciling without a position row)",
			ErrReservationLimitExceeded, this.market, this.symbol, count, req.OpenPositionLimit.Max, len(held), reconciled)
	}
	return nil
}

// heldVenues 는 계좌가 지금 든 · 들려는 서로 다른 (시장, 종목) 집합임 — ① 살아 있는 포지션 ∪ ② HELD 예약을 가진 진입 결정의 종목.
func heldVenues(ctx context.Context, tx *sql.Tx, accountRef string) (map[positionVenue]struct{}, error) {
	set := map[positionVenue]struct{}{}
	// ① 살아 있는 포지션 — 관리 여부와 무관하게 전부(D1 = B). 원장은 연결 하나라 커서를 끝까지 읽고 닫은 뒤 다음 질의를 함.
	rows, err := tx.QueryContext(ctx,
		"SELECT market, symbol FROM positions WHERE account_ref = ? AND state <> ?", accountRef, PositionClosed)
	if err != nil {
		return nil, fmt.Errorf("journal: reading live positions for the open position maximum: %w", err)
	}
	for rows.Next() {
		var market, symbol string
		if err := rows.Scan(&market, &symbol); err != nil {
			rows.Close()
			return nil, fmt.Errorf("journal: reading live positions for the open position maximum: %w", err)
		}
		set[venueOf(market, symbol)] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("journal: reading live positions for the open position maximum: %w", err)
	}
	rows.Close()

	// ② HELD 예약의 진입 결정 — 종류는 닫힌 집합 셋 전부(IN 이라 idx_reservations_held 를 그대로 탐). 이번 결정의 예약은 아직 삽입 전.
	rows, err = tx.QueryContext(ctx,
		`SELECT DISTINCT d.id, d.preimage_kind, d.risk_preimage
		   FROM risk_reservations r JOIN decisions d ON d.id = r.decision_id
		  WHERE r.account_ref = ? AND r.kind IN (?, ?, ?) AND r.state = ? AND d.safety_class = ?`,
		accountRef, ReservationKindOpenExposure, ReservationKindDailyLoss, ReservationKindCash, ReservationHeld,
		string(SafetyClassExposureRaising))
	if err != nil {
		return nil, fmt.Errorf("journal: reading held entry reservations for the open position maximum: %w", err)
	}
	type heldDecision struct{ id, kind, preimage string }
	var decisions []heldDecision
	for rows.Next() {
		var h heldDecision
		if err := rows.Scan(&h.id, &h.kind, &h.preimage); err != nil {
			rows.Close()
			return nil, fmt.Errorf("journal: reading held entry reservations for the open position maximum: %w", err)
		}
		decisions = append(decisions, h)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("journal: reading held entry reservations for the open position maximum: %w", err)
	}
	rows.Close()
	for _, h := range decisions {
		v, err := decisionVenue(h.id, h.kind, h.preimage)
		if err != nil {
			// 셀 수 없는 보유 — 0 으로 두지 않고 판정 실패.
			return nil, fmt.Errorf("journal: the open position count cannot be computed: %w", err)
		}
		set[v] = struct{}{}
	}
	return set, nil
}

// reconciledExtraSlots 는 ④ — 종목 범위 활성 RECONCILE 중 S(①~③)에 없는 것의 수임.
func reconciledExtraSlots(ctx context.Context, tx *sql.Tx, accountRef string, held map[positionVenue]struct{}, this positionVenue) (int, error) {
	rows, err := tx.QueryContext(ctx,
		`SELECT symbol, coalesce(scope_market, '') FROM reconcile_states
		  WHERE account_ref = ? AND symbol IS NOT NULL AND released_at IS NULL`, accountRef)
	if err != nil {
		return 0, fmt.Errorf("journal: reading RECONCILE scopes for the open position maximum: %w", err)
	}
	type scope struct{ symbol, market string }
	var scopes []scope
	for rows.Next() {
		var s scope
		if err := rows.Scan(&s.symbol, &s.market); err != nil {
			rows.Close()
			return 0, fmt.Errorf("journal: reading RECONCILE scopes for the open position maximum: %w", err)
		}
		scopes = append(scopes, s)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return 0, fmt.Errorf("journal: reading RECONCILE scopes for the open position maximum: %w", err)
	}
	rows.Close()

	extra := map[positionVenue]struct{}{}
	for _, s := range scopes {
		symbol := normaliseSymbol(s.symbol)
		if symbol == "" {
			continue // 계좌 전체 범위 — 종목이 없어 개수와 무관(그 상태는 진입 latch 가 막음).
		}
		if strings.TrimSpace(s.market) != "" {
			v := venueOf(s.market, symbol)
			if _, counted := held[v]; counted || v == this {
				continue
			}
			extra[v] = struct{}{}
			continue
		}
		// 시장 없는 전역 종목 범위 — 같은 종목이 S 의 어느 시장에든 있으면 이미 센 것으로 봄.
		if this.symbol == symbol || symbolHeld(held, symbol) {
			continue
		}
		extra[positionVenue{symbol: symbol}] = struct{}{}
	}
	return len(extra), nil
}

func symbolHeld(held map[positionVenue]struct{}, symbol string) bool {
	for v := range held {
		if v.symbol == symbol {
			return true
		}
	}
	return false
}

// decisionVenue 는 저장된 preimage 에서 결정의 (시장, 종목) 을 읽음 — 해석은 ParsePreimage 한 곳.
func decisionVenue(id, kind, preimage string) (positionVenue, error) {
	parsed, err := ParsePreimage(kind, preimage)
	if err != nil {
		return positionVenue{}, fmt.Errorf("journal: decision %s preimage: %w", id, err)
	}
	_, market, symbol, _ := PreimageVenue(parsed)
	if strings.TrimSpace(market) == "" || strings.TrimSpace(symbol) == "" {
		return positionVenue{}, fmt.Errorf("journal: decision %s names no market or symbol", id)
	}
	return venueOf(market, symbol), nil
}

-- a087 tasks 3.8 — 보호 청산 슬리피지(bps) 읽기 전용 질의. schema·production 코드 추가 없음(기존 칸 유도).
--
-- 결속: exit_events(발의 행: observed_price · action · proposed_intent_id 를 한 행에 씀 — internal/journal/exit_state.go
--       INSERT INTO exit_events) → intents(id) → mutation_attempts(PLACE · CONFIRMED · broker_order_id)
--       → 체결 스냅숏(order_id + intent 의 전체 범위: account · market · trading_day · symbol · side).
-- 범위로도 묶는 이유: 브로커 order_id 는 날짜를 넘어 재사용될 수 있음(fills.go 의 reused-id 처리) — order_id 만으로
-- 이으면 다른 날 같은 번호 주문의 체결가를 가져옴. 스냅숏 정본은 scoped_fill_snapshots 이고 fill_snapshots 는
-- 같은 범위가 scoped 에 없을 때만 씀(fills.go allFillSnapshotsCTE 와 같은 규칙; 범위가 빈 v15 행은 제외).
--
-- adverse_bps = (결정 시 관측가 − 체결 평균가) / 관측가 × 10000 — 매도 기준, 양수 = 관측가보다 싸게 팔림(비용).
-- 체결 스냅숏이 없거나 평균가가 비면 NULL(0 으로 읽어 10000 bps 를 만들지 않음).
-- 한계: AMEND 로 이어진 후속 주문은 따라가지 않음(PLACE 주문만). 부분체결 중이면 terminal = 0 인 그 시점 평균임.
-- 판정 근거로 쓰지 않음 — 비교 임계·재비교 여부는 issues.md I-R3(사람 결정).
--
-- 실행(사람, 읽기 전용): sqlite3 -readonly "$HOME/.config/tossctl/<journal>.db" < exit_slippage.sql
WITH snap AS (
	SELECT order_id, account_ref, market, trading_day, symbol, side,
	       state, terminal, filled_quantity, average_price
	  FROM scoped_fill_snapshots
	UNION ALL
	SELECT legacy.order_id, TRIM(legacy.account_ref), LOWER(TRIM(legacy.market)), TRIM(legacy.trading_day),
	       UPPER(TRIM(legacy.symbol)), UPPER(TRIM(legacy.side)),
	       legacy.state, legacy.terminal, legacy.filled_quantity, legacy.average_price
	  FROM fill_snapshots legacy
	 WHERE TRIM(legacy.account_ref) <> ''
	   AND NOT EXISTS (
		SELECT 1 FROM scoped_fill_snapshots s
		 WHERE s.order_id = legacy.order_id
		   AND s.account_ref = TRIM(legacy.account_ref)
		   AND s.market = LOWER(TRIM(legacy.market))
		   AND s.trading_day = TRIM(legacy.trading_day)
		   AND s.symbol = UPPER(TRIM(legacy.symbol))
		   AND s.side = UPPER(TRIM(legacy.side)))
)
SELECT e.id                 AS exit_event_id,
       e.position_id,
       e.created_at         AS decided_at,
       e.action,
       i.id                 AS intent_id,
       i.order_type,
       i.symbol,
       i.market,
       a.broker_order_id    AS order_id,
       e.observed_price,
       s.average_price,
       s.filled_quantity,
       s.state,
       s.terminal,
       ROUND((CAST(e.observed_price AS REAL) - CAST(NULLIF(TRIM(s.average_price), '') AS REAL))
             / CAST(e.observed_price AS REAL) * 10000, 2) AS adverse_bps
  FROM exit_events e
  JOIN intents i ON i.id = e.proposed_intent_id
  JOIN mutation_attempts a
    ON a.intent_id = i.id AND a.kind = 'PLACE' AND a.state = 'CONFIRMED' AND a.broker_order_id <> ''
  LEFT JOIN snap s
    ON s.order_id = a.broker_order_id
   AND s.account_ref = TRIM(i.account_ref)
   AND s.market = LOWER(TRIM(i.market))
   AND s.trading_day = TRIM(i.trading_day)
   AND s.symbol = UPPER(TRIM(i.symbol))
   AND s.side = UPPER(TRIM(i.side))
 WHERE e.action IN ('BASELINE_BREACH_EXIT', 'STOP_LOSS_LADDER')
   AND UPPER(TRIM(i.side)) = 'SELL'
   AND TRIM(coalesce(e.observed_price, '')) <> ''
 ORDER BY e.id;

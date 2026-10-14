# Function Logic Map: `sweepTradingDayBoundary`

- Source: `internal/journal/reservation_release.go` (390-453, base 7b073f55)
- Qualified: `sweepTradingDayBoundary`
- 위치: `analysis/ast-evidence/` — **proposal 단계 근거(게이트 번들 아님)**. 편집하게 되는 함수는 구현 착수 때 `analysis/function-logic/` 에 정식 번들을 만든다.
- AST evidence: `ast.json` (`source_sha256` d61f428958e0ac5b…) — AST branches 10 · returns 6
- Risk scan: `risk-pattern-report.md`
- 용도: a136 design D1 의 주장 「저장소의 유일한 일일 경계 관례는 **시장별(market-local)** 거래일이다」의 근거. a136 은 이 함수를 편집할지 D1·D2 결정 뒤에 정한다(편집하면 이 번들을 편집 뒤로 재추출).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `accountRef` | 빈 값 = 전 계좌(기동 sweep) | 호출자 `sweepLapsedReservations` | — |
| `now` | journal 주입 시계 UTC | `j.clk` | — |
| `risk_reservations.kind = DAILY_LOSS` · `state = HELD` | 행 집합 | journal | 질의 오류는 오류 반환 |
| 결정 preimage 의 시장 | `kr` · `us` | `decisions.risk_preimage` | 해석 불가 → 보존 + `UNKNOWN_MARKET` 경보 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if accountRef != ""` (395) | 질의에 계좌 조건 추가 | — | `TestDailyLossReservationLapsesAtTheMarketsDayBoundary`(계좌 범위) · 기동 sweep 시험 |
| B2 | `if err != nil` (400) 질의 실패 | 없음 | 오류 반환 | 결함 주입 시험 없음(DB 오류) |
| B3 | `for rows.Next()` (405) | 후보 수집 | — | B5~ 공통 |
| B4 | `if rows.Scan … err != nil` (410) | rows.Close | 오류 반환 | 결함 주입 시험 없음 |
| B5 | `if rows.Err() != nil` (417) | rows.Close | 오류 반환 | 결함 주입 시험 없음 |
| B6 | `range candidates` (424) | — | — | `TestDailyLossReservationLapsesAtTheMarketsDayBoundary` |
| B7 | `if err != nil` (426) — `tradingDayForPreimage` 실패(시장 해석 불가) | `Preserved` 에 `UNKNOWN_MARKET` 경보 추가, **해제 안 함** | continue | 기존 시험 확인 필요(D2 결정 뒤 Branch Test Map 확정) |
| B8 | `if scanErr != nil` (428) | 없음 | 오류 반환 | 결함 주입 시험 없음 |
| B9 | `if today == c.day` (441) — **예약의 시장에서 본 오늘이 예약의 거래일과 같음** | 없음 | continue(보유 유지) | `TestDailyLossReservationLapsesAtTheMarketsDayBoundary` · `TestOpenExposureDoesNotLapseWithTheDay`(종류 한정) |
| B10 | `if err != nil` (447) 해제 실패 | 없음 | 오류 반환 | 결함 주입 시험 없음 |

Exact AST return positions: 401 · 412 · 419 · 429 · 448 · 452

핵심: 오늘의 판정은 `tradingDayForPreimage` → `clock.ParseMarket(market).TradingDay(now)` — **그 예약의 시장 시간대**(Asia/Seoul · America/New_York)의 달력 날짜다. 계좌 전체에 하나인 「오늘」은 이 코드에 없다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `tx.QueryContext` | DAILY_LOSS HELD 후보 | 오류 반환 | AST calls |
| `tradingDayForPreimage` | 결정 preimage 의 시장 → 시장별 거래일 | 오류 → 보존·경보(B7) | reservation_release.go:456-468 |
| `releaseWhere(…, ReleaseReasonDayBoundary, …)` | 날짜가 바뀐 보유 해제 | 오류 반환 | AST calls |
| `scanReservation` | 경보용 행 읽기 | 오류 반환 | AST calls |

## State mutations and fallbacks

- 해제: `DAY_BOUNDARY` 사유로 RELEASED 전이(같은 tx). fallback: 시장 해석 불가면 보존(보수 방향).

## Safety conclusion

- Safe edit boundary: D1 결정 전에는 편집하지 않음. 새 종류를 같은 sweep 에 태우면 B9 의 「예약별 시장 거래일」 의미가 그대로 승계된다 — 그 의미가 사용자 요구 「시장 합산 1일」과 맞는지가 D1 의 질문이다.
- High-risk impact: yes (예약 해제 = 한도 여유 증가 방향). 편집 시 FLM·변이 원장 의무.

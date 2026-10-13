# Function Logic Map: `sumSellFills`

- Source: `internal/journal/trade_outcomes.go` (`665`–`770`, 이 change 무편집)
- AST evidence: `ast.json` (`source_sha256` ff51203717c0ed3c…) — branches 11 · returns 6
- Risk scan: `risk-pattern-report.md`
- 작성 시점: 4판(리뷰 4차 P2-6 — 설계가 이 함수 내부를 근거로 씀).

**역할.** 귀속된 매도 주문들의 체결 대금 · 수량 합.
호출자: `adoptedRoundTripLegs` B4.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 원장 | — | 원장 | 각 분기 표 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `:669` `for _, scope := range orders {` | — | — | 무편집 갈래 |
| B2 | `:710` `if err != nil {` | — | — | 무편집 갈래 |
| B3 | `:724` `for rows.Next() {` | — | — | 무편집 갈래 |
| B4 | `:729` `if err := rows.Scan(&orderID, &eventAccount, &eventMarket, &tradingDay,` | — | — | 무편집 갈래 |
| B5 | `:733` `if !strings.EqualFold(strings.TrimSpace(side), "SELL") {` | — | — | 무편집 갈래 |
| B6 | `:740` `if orderKey != currentOrder {` | — | — | 무편집 갈래 |
| B7 | `:744` `if !ok {` | — | — | 무편집 갈래 |
| B8 | `:748` `if strings.TrimSpace(average) == "" {` | — | — | 무편집 갈래 |
| B9 | `:756` `if !ok {` | — | — | 무편집 갈래 |
| B10 | `:763` `if err := rows.Err(); err != nil {` | — | — | 무편집 갈래 |
| B11 | `:766` `for _, filled := range perOrder {` | — | — | 무편집 갈래 |

## Calls and live bindings

AST `calls` 목록 참조. Live config binding: 없음.

## State mutations and fallbacks

- 아래 근거 자리 참조.

## a130 편집 계획 (4판)

**인용 근거만 · 무편집.** 근거: 체결은 `a.settled_at < f.committed_at` 인 확정 주문에만 귀속, 주문별 마지막 누적 수량의 합이 매도 수량 — 인스턴스 매도 무결 검사는 이 귀속 집합(orders)을 그대로 받음.

## Safety conclusion

- Safe edit boundary: 무편집.
- High-risk impact: no(무편집) — 인용 근거.

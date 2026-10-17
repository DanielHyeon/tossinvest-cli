# Function Logic Map: `roundTripLegs`

- Source: `internal/journal/trade_outcomes.go` (편집 전 `292`–`429`, base `233d5d9c`)
- AST evidence: `ast.json` (`source_sha256` 0bd43abf96107b99…) — branches 14 · returns 9
- Risk scan: `risk-pattern-report.md`
- 작성 시점: **편집 전**(2판, 리뷰 F-5 — 설계가 이 함수 내부를 근거로 씀).

**역할.** 엔진 진입 포지션의 왕복 다리 — 진입 결정의 첫 체결 id 이후 체결.
호출자: `computeTradeOutcome` B8.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 같은 트랜잭션의 원장 | — | 원장 | 각 분기 표 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `:324` `if err != nil {` | — | — | 무편집 갈래 |
| B2 | `:327` `if bound.Next() {` | — | — | 무편집 갈래 |
| B3 | `:331` `if !low.Valid {` | — | — | 무편집 갈래 |
| B4 | `:368` `if err != nil {` | — | — | 무편집 갈래 |
| B5 | `:377` `for rows.Next() {` | — | — | 무편집 갈래 |
| B6 | `:382` `if err := rows.Scan(&orderID, &eventAccount, &eventMarket, &tradingDay,` | — | — | 무편집 갈래 |
| B7 | `:388` `if orderKey != currentOrder {` | — | — | 무편집 갈래 |
| B8 | `:393` `if strings.EqualFold(strings.TrimSpace(side), "SELL") {` | — | — | 무편집 갈래 |
| B9 | `:397` `if !ok {` | — | — | 무편집 갈래 |
| B10 | `:400` `if strings.TrimSpace(average) == "" {` | — | — | 무편집 갈래 |
| B11 | `:408` `if !ok {` | — | — | 무편집 갈래 |
| B12 | `:415` `if leg == &buy {` | — | — | 무편집 갈래 |
| B13 | `:422` `if err := rows.Err(); err != nil {` | — | — | 무편집 갈래 |
| B14 | `:425` `if !buy.priced \|\| !sell.priced {` | — | — | 무편집 갈래 |

## Calls and live bindings

AST `calls` 목록 참조. Live config binding: 없음.

## State mutations and fallbacks

- 아래 「a130 편집 계획」 의 근거 자리 참조.

## a130 편집 계획 (2판)

**인용 근거만 · 무편집.** 근거 자리: 매도 수량 검사 없음(B14 `:425` 는 priced 만), `id >= low`(:365) 상한 없음 — 닫는 트랜잭션 안에서만 정확(F-2, design D5 잔여). a130 은 이 함수를 트랜잭션 밖에서 새로 부르지 않음.

## Safety conclusion

- Safe edit boundary: 위 계획만.
- High-risk impact: no(무편집) — 인용 근거.

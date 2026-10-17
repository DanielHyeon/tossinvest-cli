# Function Logic Map: `engineSellOrders`

- Source: `internal/journal/trade_outcomes.go` (편집 전 `551`–`610`, base `233d5d9c`)
- AST evidence: `ast.json` (`source_sha256` 0bd43abf96107b99…) — branches 4 · returns 4
- Risk scan: `risk-pattern-report.md`
- 작성 시점: **편집 전**(2판, 리뷰 F-5 — 설계가 이 함수 내부를 근거로 씀).

**역할.** 편입 인스턴스의 엔진 매도 주문을 exit 이벤트 사슬 · flatten 사슬로 모음.
호출자: `adoptedRoundTripLegs` B2.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 같은 트랜잭션의 원장 | — | 원장 | 각 분기 표 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `:592` `if err != nil {` | — | — | 무편집 갈래 |
| B2 | `:598` `for rows.Next() {` | — | — | 무편집 갈래 |
| B3 | `:600` `if err := rows.Scan(&scope.OrderID, &scope.AccountRef, &scope.Market,` | — | — | 무편집 갈래 |
| B4 | `:606` `if err := rows.Err(); err != nil {` | — | — | 무편집 갈래 |

## Calls and live bindings

AST `calls` 목록 참조. Live config binding: 없음.

## State mutations and fallbacks

- 아래 「a130 편집 계획」 의 근거 자리 참조.

## a130 편집 계획 (2판)

**인용 근거만 · 무편집.** 근거 자리: flatten 갈래 `:577-591` 은 계좌 · 심볼 · 시장으로만 묶여 인스턴스 · 시각 경계가 없음 — 이전 인스턴스 매도를 쓸어 담을 수 있음(F-3).

## Safety conclusion

- Safe edit boundary: 위 계획만.
- High-risk impact: no(무편집) — 인용 근거.

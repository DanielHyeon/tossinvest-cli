# Function Logic Map: `TestFractionalShareAmountsSurviveTheLedger`

- Source: `internal/journal/reservations_test.go` (225-250, 편집 뒤)
- Qualified: `TestFractionalShareAmountsSurviveTheLedger`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` 9352330413290570…, 편집 전 분기 수 4) — 현재 branches 4 · returns 0
- Risk scan: `risk-pattern-report.md`

**역할.** 시험(생산 코드 아님).

**a136 편집.** USD 로 바꾼 예약에 맞춰 하루 매수 상한 통화도 USD 로 — 시험이 재는 정밀도 단언 무변경.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| — | — | 시험 | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err != nil {` (if 237) | — | 분기 본문 참조 | `TestFractionalShareAmountsSurviveTheLedger` |
| B2 | `if got := out.Reservations[0].Amount; got != "0.0001" {` (if 240) | — | 분기 본문 참조 | `TestFractionalShareAmountsSurviveTheLedger` |
| B3 | `if err != nil {` (if 244) | — | 분기 본문 참조 | `TestFractionalShareAmountsSurviveTheLedger` |
| B4 | `if stored.Amount != "0.0001" \|\| stored.Currency != "USD" \|\| !stored.Held() {` (if 247) | — | 분기 본문 참조 | `TestFractionalShareAmountsSurviveTheLedger` |

Exact AST return positions: 

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `j.Reserve` | 예약 | — | AST |

## State mutations and fallbacks

- 시험 데이터만.

## Safety conclusion

- Safe edit boundary: 한 줄.
- High-risk impact: no(시험 전용)

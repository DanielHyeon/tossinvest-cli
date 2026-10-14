# Function Logic Map: `recordDailyBuySave`

- Source: `cmd/tossctl/limitsettings.go` (117-136, 편집 뒤)
- Qualified: `recordDailyBuySave`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` 8e4690b6126a248b…, 편집 전 분기 수 2) — 현재 branches 3 · returns 4
- Risk scan: `risk-pattern-report.md`

**역할.** 콘솔 저장의 하루 매수 상한 audit 한 줄(a136 신규 함수, 새 base 이후 편집).

**a136 편집.** 844a042c: 값에 통화를 붙여 숫자가 같아도 통화가 바뀌면 변화로 남김(구현 리뷰 P1-1) — 0 은 「0」.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 인자 | — | 호출자 | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err != nil {` (if 119) | — | 분기 본문 참조 | `TestA136TheDailyBuyCeilingSaveIsAuditedFromTheFile` |
| B2 | `if err != nil {` (if 123) | — | 분기 본문 참조 | `TestA136TheDailyBuyCeilingSaveIsAuditedFromTheFile` |
| B3 | `if v == 0 {` (if 128) | — | 분기 본문 참조 | `TestA136TheDailyBuyCeilingSaveIsAuditedFromTheFile` |

Exact AST return positions: 120 · 124 · 129 · 131

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| — | — | — | AST |

## State mutations and fallbacks

- 편집 내용 참조.

## Safety conclusion

- Safe edit boundary: 844a042c: 값에 통화를 붙여 숫자가 같아도 통화가 바뀌면 변화로 남김(구현 리뷰 P1-1) — 0 은 「0」.
- High-risk impact: no(audit 문구)

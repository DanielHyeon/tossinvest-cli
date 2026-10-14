# Function Logic Map: `acceptanceDetail`

- Source: `internal/app/engine/interlock.go` (420-435, 편집 뒤)
- Qualified: `acceptanceDetail`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` a2277d4bd3a76152…, 편집 전 분기 수 0) — 현재 branches 0 · returns 1
- Risk scan: `risk-pattern-report.md`

**역할.** 기동 승인 audit 문구 — 모든 한도를 적음.

**a136 편집.** `max_daily_buy_notional=` 항목 끝에 추가.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `status` | — | 인터록 | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (없음) | 직선 함수 — AST branches 0 | — | — | `TestA136TheInterlockAuditsTheDailyBuyCeiling` |

Exact AST return positions: 422

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `fmt.Sprintf` · `limitString` | 문구 | — | AST |

## State mutations and fallbacks

- 없음.

## Safety conclusion

- Safe edit boundary: 문구 항목 하나 추가(기존 항목 순서 무변경).
- High-risk impact: no(감사 문구)

# Function Logic Map: `newAutomationStatus`

- Source: `internal/app/engine/interlock.go` (325-336, 편집 뒤)
- Qualified: `newAutomationStatus`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` a2277d4bd3a76152…, 편집 전 분기 수 0) — 현재 branches 0 · returns 1
- Risk scan: `risk-pattern-report.md`

**역할.** 검증 전 게이트 상태(감사 대상 한도).

**a136 편집.** `DailyBuyNotional: boundIfPositive(gate.EffectiveMaxDailyBuyNotional())` 칸 추가 — 인터록이 Guardian 진술과 대조할 해석 값.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `gate` | 임의 | config | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (없음) | 직선 함수 — AST branches 0 | — | — | `TestA136TheInterlockAuditsTheDailyBuyCeiling` |

Exact AST return positions: 328

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `gateLimits` · `attestationPath` | — | — | AST |

## State mutations and fallbacks

- 없음.

## Safety conclusion

- Safe edit boundary: 필드 하나.
- High-risk impact: yes(인터록)

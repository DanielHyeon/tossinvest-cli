# Function Logic Map: `DefaultPolicy`

- Source: `internal/risk/input.go` (236-249, 편집 뒤)
- Qualified: `DefaultPolicy`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` 3ef6826797400f67…, 편집 전 분기 수 0) — 현재 branches 0 · returns 1
- Risk scan: `risk-pattern-report.md`

**역할.** 사용자 미확정 보수 기본 정책 집합.

**a136 편집.** `MaxDailyBuyNotional: 10000000 KRW` 한 칸 추가(사용자 결정 2026-10-14). FLM 전에 편집됨(P1-3 기록).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| — | — | 상수 | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (없음) | 직선 함수 — AST branches 0 | — | — | `TestA136DefaultPolicyCarriesTheDailyBuyCeiling` · `TestTheDefaultPolicyStatesTheSmallLiveSet` |

Exact AST return positions: 237

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| — | — | — | AST calls 0 |

## State mutations and fallbacks

- 없음.

## Safety conclusion

- Safe edit boundary: 필드 하나 추가.
- High-risk impact: yes(정책 수치) — 새 검사라 완화 아님

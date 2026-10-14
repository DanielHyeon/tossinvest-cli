# Function Logic Map: `usdPolicy`

- Source: `internal/risk/chain_test.go` (61-69, 편집 뒤)
- Qualified: `usdPolicy`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` f80180ce32d66111…, 편집 전 분기 수 0) — 현재 branches 0 · returns 1
- Risk scan: `risk-pattern-report.md`

**역할.** 시험 fixture/시험 함수(생산 코드 아님).

**a136 편집.** USD 정책 · 게이트 fixture 에 하루 매수 상한을 명시(USD 는 기본값 없음, D5) — 시험이 재는 다른 한도는 무변경

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| — | — | 시험 | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (없음) | 직선 함수 — AST branches 0 | — | — | `go test ./internal/risk` (이 함수를 쓰는 기존 시험) |

Exact AST return positions: 68

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| — | fixture 구성 | — | AST |

## State mutations and fallbacks

- 시험 데이터만.

## Safety conclusion

- Safe edit boundary: 칸 추가 — 시험이 단언하는 값 무변경.
- High-risk impact: no(시험 전용, 비례 원칙)

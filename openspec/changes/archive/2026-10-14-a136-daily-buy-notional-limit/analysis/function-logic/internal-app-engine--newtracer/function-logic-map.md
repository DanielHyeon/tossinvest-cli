# Function Logic Map: `newTracer`

- Source: `internal/app/engine/tracer_test.go` (43-63, 편집 뒤)
- Qualified: `newTracer`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` 55a7908bf7ab60e7…, 편집 전 분기 수 2) — 현재 branches 2 · returns 1
- Risk scan: `risk-pattern-report.md`

**역할.** 시험 fixture/시험(생산 코드 아님).

**a136 편집.** a136 수동 매수 보드(「수동 매수 0」) 또는 보조 실행자 이름 목록을 맞춤 — 시험이 재는 것 무변경.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| — | — | 시험 | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err != nil {` (if 51) | — | 분기 본문 참조 | `go test ./internal/app/engine` |
| B2 | `if err != nil {` (if 59) | — | 분기 본문 참조 | `go test ./internal/app/engine` |

Exact AST return positions: 62

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| — | fixture | — | AST |

## State mutations and fallbacks

- 시험 데이터만.

## Safety conclusion

- Safe edit boundary: 칸 하나.
- High-risk impact: no(시험 전용)

# Function Logic Map: `restart`

- Source: `internal/app/engine/a112_dispatch_crash_restart_test.go` (79-100, 편집 뒤)
- Qualified: `a112TradingFixture.restart`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` a6b5e7ffb11a3158…, 편집 전 분기 수 3) — 현재 branches 3 · returns 1
- Risk scan: `risk-pattern-report.md`

**역할.** 시험 fixture/시험(생산 코드 아님, tossos_testseams).

**a136 편집.** testseam fixture 의 Guardian 에 「수동 매수 0」 보드(a136 — nil 이면 진입 전부 거부). 시험이 재는 것 무변경.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| — | — | 시험 | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err := fixture.journal.Close(); err != nil {` (if 82) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine` |
| B2 | `if err != nil {` (if 88) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine` |
| B3 | `if err != nil {` (if 94) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine` |

Exact AST return positions: 99

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| — | fixture | — | AST |

## State mutations and fallbacks

- 시험 데이터만.

## Safety conclusion

- Safe edit boundary: 칸 하나.
- High-risk impact: no(시험 전용)

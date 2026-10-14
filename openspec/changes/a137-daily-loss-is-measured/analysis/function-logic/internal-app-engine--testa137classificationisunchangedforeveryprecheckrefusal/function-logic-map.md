# Function Logic Map: `TestA137ClassificationIsUnchangedForEveryPrecheckRefusal`

- Source: `internal/app/engine/a137_daily_loss_cycle_test.go` (145-168, 편집 뒤)
- Qualified: `TestA137ClassificationIsUnchangedForEveryPrecheckRefusal`
- AST evidence: `ast.json`(편집 뒤, `source_sha256` 9f1fa83e44763f42…) · `ast-base.json`(편집 전, base `63fd0c92`, `source_sha256` b465b2d519f1525a…, 편집 전 분기수 2) — 현재 branches 2 · returns 0
- Risk scan: `risk-pattern-report.md`

**역할.** precheck 거절 분류 불변 표 시험.

**a137 편집.** 구현 리뷰 수리(c83241ff)에서 gofmt 정렬만(분기 불변).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| cases | 거절 표 | 시험 | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for name, c := range cases {` (range 159) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine -run TestA137Classification` |
| B2 | `if outcome.Code != c.code \|\| (issuer.escalations == 1) != c.escalate \|\| errors.As(outcome.cause, &scope) \|\| isCentralStrategyIntegrity(outcome.cause) {` (if 164) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine -run TestA137Classification` |

Exact AST return positions: 

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `admit` | 승인 | — | AST calls |

## State mutations and fallbacks

- 없음.

## Safety conclusion

- Safe edit boundary: 시험.
- High-risk impact: no — 시험.

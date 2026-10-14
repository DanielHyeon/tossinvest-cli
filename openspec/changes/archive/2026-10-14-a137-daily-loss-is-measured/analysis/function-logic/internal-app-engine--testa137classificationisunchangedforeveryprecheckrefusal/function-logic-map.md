# Function Logic Map: `TestA137ClassificationIsUnchangedForEveryPrecheckRefusal`

- Source: `internal/app/engine/a137_daily_loss_cycle_test.go` (138-161, base 리비전)
- Qualified: `TestA137ClassificationIsUnchangedForEveryPrecheckRefusal`
- AST evidence: `ast.json`(편집 뒤, `source_sha256` b465b2d519f1525a…) · `ast-base.json`(편집 전, base `63fd0c92`, `source_sha256` b465b2d519f1525a…, 편집 전 분기수 2) — 현재 branches 2 · returns 0
- Risk scan: `risk-pattern-report.md`

**역할.** precheck 거절 분류 불변 표 시험.

**a137 편집.** 본문 무편집 — c83241ff 의 diff hunk 가 함수 끝 줄(옛 쪽)에 닿아(바로 뒤에 새 함수 · 새 시험을 덧붙임) 게이트가 base 리비전 증거를 요구함. `ast.json` 은 `revision: base`(63fd0c92 blob, 분기 불변).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| cases | 거절 표 | 시험 | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `"daily loss at the chain":          {&execgw.IssueRefusal{Stage: execgw.StageChain, Reason: string(risk.ReasonDailyLossLimitReached)}, StrategyFirstLegDailyLossLimit, true},` (range 152) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine -run TestA137Classification` |
| B2 | `"q_final refusal":                  {&execgw.QFinalRefusal{Code: "OWNER_CONFLICT"}, StrategyFirstLegAuthorityMismatch, false},` (if 157) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine -run TestA137Classification` |

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

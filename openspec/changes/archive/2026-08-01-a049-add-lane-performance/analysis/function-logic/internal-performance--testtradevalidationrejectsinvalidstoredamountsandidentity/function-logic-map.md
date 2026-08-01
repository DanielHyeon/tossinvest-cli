# Function Logic Map: `TestTradeValidationRejectsInvalidStoredAmountsAndIdentity`

- Source: `internal/performance/model_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function parameters/state | test fixture and assertions | current Go signature and persisted/server-owned data | invalid, missing, or corrupt evidence follows explicit error/not-measured/test-failure paths |
| safety boundary | server-owned identities and fixed contracts only | approved a049 OpenSpec plus current code | never invents lineage/cost and never expands trading authority |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | base-revision AST `range` at line 136: `for _, test := range tests {` | isolated test fixture/assertion state only | assertion failure is explicit through `testing.T` | `TestTradeValidationRejectsInvalidStoredAmountsAndIdentity` (this regression test) |
| B2 | base-revision AST `if` at line 140: `if err := trade.validate(); err == nil {` | isolated test fixture/assertion state only | assertion failure is explicit through `testing.T` | `TestTradeValidationRejectsInvalidStoredAmountsAndIdentity` (this regression test) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `time.Date` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestTradeValidationRejectsInvalidStoredAmountsAndIdentity` (this regression test) |
| `t.Run` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestTradeValidationRejectsInvalidStoredAmountsAndIdentity` (this regression test) |
| `measuredTrade` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestTradeValidationRejectsInvalidStoredAmountsAndIdentity` (this regression test) |
| `test.mutate` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestTradeValidationRejectsInvalidStoredAmountsAndIdentity` (this regression test) |
| `trade.validate` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestTradeValidationRejectsInvalidStoredAmountsAndIdentity` (this regression test) |
| `t.Fatal` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestTradeValidationRejectsInvalidStoredAmountsAndIdentity` (this regression test) |

## State mutations and fallbacks

- isolated test state only; failures are reported through `testing.T`.
- There is no hidden broker polling, live-order fallback, or user-entered identifier path in this function.
- Missing, ambiguous, or corrupt evidence is preserved as an error/not-measured state or an explicit test failure.

## Safety conclusion

- Safe edit boundary: `internal/performance/model_test.go` function `TestTradeValidationRejectsInvalidStoredAmountsAndIdentity` and its documented derived/test state.
- High-risk impact: no runtime authority.

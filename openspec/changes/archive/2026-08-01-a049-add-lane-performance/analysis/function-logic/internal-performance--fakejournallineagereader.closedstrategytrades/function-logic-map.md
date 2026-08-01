# Function Logic Map: `fakeJournalLineageReader.ClosedStrategyTrades`

- Source: `internal/performance/lineage_reader_test.go`
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
| B1 | AST `if` at `internal/performance/lineage_reader_test.go:18`: `if f.err != nil {` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestJournalHandoffConsumesOneExactLineageReadAndCallerOwnedObservations` and the named handoff regression tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `append` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestJournalHandoffConsumesOneExactLineageReadAndCallerOwnedObservations` and the named handoff regression tests |

## State mutations and fallbacks

- isolated test state only; failures are reported through `testing.T`.
- There is no hidden broker polling, live-order fallback, or user-entered identifier path in this function.
- Missing, ambiguous, or corrupt evidence is preserved as an error/not-measured state or an explicit test failure.

## Safety conclusion

- Safe edit boundary: `internal/performance/lineage_reader_test.go` function `fakeJournalLineageReader.ClosedStrategyTrades` and its documented derived/test state.
- High-risk impact: no runtime authority.

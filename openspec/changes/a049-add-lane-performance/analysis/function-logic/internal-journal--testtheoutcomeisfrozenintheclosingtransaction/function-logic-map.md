# Function Logic Map: `TestTheOutcomeIsFrozenInTheClosingTransaction`

- Source: `internal/journal/trade_outcomes_test.go`
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
| B1 | AST `if` at `internal/journal/trade_outcomes_test.go:76`: `if got.InitialQuantity != "10" {` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestTheOutcomeIsFrozenInTheClosingTransaction` (this regression test) |
| B2 | AST `if` at `internal/journal/trade_outcomes_test.go:79`: `if got.InitialRisk != "2000" {` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestTheOutcomeIsFrozenInTheClosingTransaction` (this regression test) |
| B3 | AST `if` at `internal/journal/trade_outcomes_test.go:82`: `if got.ClosedAt == "" {` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestTheOutcomeIsFrozenInTheClosingTransaction` (this regression test) |
| B4 | AST `if` at `internal/journal/trade_outcomes_test.go:88`: `if !strings.HasPrefix(got.RealizedPnLAfterCosts, "1") {` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestTheOutcomeIsFrozenInTheClosingTransaction` (this regression test) |
| B5 | AST `if` at `internal/journal/trade_outcomes_test.go:93`: `if net >= gross {` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestTheOutcomeIsFrozenInTheClosingTransaction` (this regression test) |
| B6 | AST `if` at `internal/journal/trade_outcomes_test.go:97`: `if got.CostTotal == nil {` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestTheOutcomeIsFrozenInTheClosingTransaction` (this regression test) |
| B7 | AST `if` at `internal/journal/trade_outcomes_test.go:101`: `if ratOf(*got.CostTotal).Cmp(wantCost) != 0 {` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestTheOutcomeIsFrozenInTheClosingTransaction` (this regression test) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `outcomeFixture` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestTheOutcomeIsFrozenInTheClosingTransaction` (this regression test) |
| `roundTrip` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestTheOutcomeIsFrozenInTheClosingTransaction` (this regression test) |
| `outcomeOf` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestTheOutcomeIsFrozenInTheClosingTransaction` (this regression test) |
| `t.Errorf` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestTheOutcomeIsFrozenInTheClosingTransaction` (this regression test) |
| `t.Error` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestTheOutcomeIsFrozenInTheClosingTransaction` (this regression test) |
| `strings.HasPrefix` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestTheOutcomeIsFrozenInTheClosingTransaction` (this regression test) |
| `Float64` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestTheOutcomeIsFrozenInTheClosingTransaction` (this regression test) |
| `ratOf` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestTheOutcomeIsFrozenInTheClosingTransaction` (this regression test) |

## State mutations and fallbacks

- isolated test state only; failures are reported through `testing.T`.
- There is no hidden broker polling, live-order fallback, or user-entered identifier path in this function.
- Missing, ambiguous, or corrupt evidence is preserved as an error/not-measured state or an explicit test failure.

## Safety conclusion

- Safe edit boundary: `internal/journal/trade_outcomes_test.go` function `TestTheOutcomeIsFrozenInTheClosingTransaction` and its documented derived/test state.
- High-risk impact: no runtime authority.

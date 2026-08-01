# Function Logic Map: `TestABackfillRecoversTheGapAndRefusesToRewriteIt`

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
| B1 | AST `if` at `internal/journal/trade_outcomes_test.go:235`: `if err := j.SetApplyHooks(ApplyHooks{Project: ProjectPosition, Exit: ApplyExitFill}); err != nil {` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestABackfillRecoversTheGapAndRefusesToRewriteIt` (this regression test) |
| B2 | AST `if` at `internal/journal/trade_outcomes_test.go:242`: `if err != nil {` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestABackfillRecoversTheGapAndRefusesToRewriteIt` (this regression test) |
| B3 | AST `if` at `internal/journal/trade_outcomes_test.go:245`: `if filled.InitialQuantity != "10" {` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestABackfillRecoversTheGapAndRefusesToRewriteIt` (this regression test) |
| B4 | AST `if` at `internal/journal/trade_outcomes_test.go:248`: `if filled.CostTotal == nil {` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestABackfillRecoversTheGapAndRefusesToRewriteIt` (this regression test) |
| B5 | AST `if` at `internal/journal/trade_outcomes_test.go:252`: `if ratOf(*filled.CostTotal).Cmp(wantBackfillCost) != 0 {` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestABackfillRecoversTheGapAndRefusesToRewriteIt` (this regression test) |
| B6 | AST `if` at `internal/journal/trade_outcomes_test.go:256`: `if stored.CostTotal == nil \|\| *stored.CostTotal != *filled.CostTotal {` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestABackfillRecoversTheGapAndRefusesToRewriteIt` (this regression test) |
| B7 | AST `if` at `internal/journal/trade_outcomes_test.go:259`: `if _, err := j.BackfillTradeOutcome(ctx, id, costs.DefaultModel()); !errors.Is(err, ErrTradeOutcomeExists) {` | isolated test state only; failures are reported through `testing.T` | condition determines the documented success/error/assertion path | `TestABackfillRecoversTheGapAndRefusesToRewriteIt` (this regression test) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `openTestJournal` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestABackfillRecoversTheGapAndRefusesToRewriteIt` (this regression test) |
| `j.SetApplyHooks` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestABackfillRecoversTheGapAndRefusesToRewriteIt` (this regression test) |
| `t.Fatalf` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestABackfillRecoversTheGapAndRefusesToRewriteIt` (this regression test) |
| `context.Background` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestABackfillRecoversTheGapAndRefusesToRewriteIt` (this regression test) |
| `roundTrip` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestABackfillRecoversTheGapAndRefusesToRewriteIt` (this regression test) |
| `j.BackfillTradeOutcome` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestABackfillRecoversTheGapAndRefusesToRewriteIt` (this regression test) |
| `costs.DefaultModel` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestABackfillRecoversTheGapAndRefusesToRewriteIt` (this regression test) |
| `t.Errorf` | implements the function contract | error/return is handled by the surrounding AST path; no implicit retry | current AST + `TestABackfillRecoversTheGapAndRefusesToRewriteIt` (this regression test) |

## State mutations and fallbacks

- isolated test state only; failures are reported through `testing.T`.
- There is no hidden broker polling, live-order fallback, or user-entered identifier path in this function.
- Missing, ambiguous, or corrupt evidence is preserved as an error/not-measured state or an explicit test failure.

## Safety conclusion

- Safe edit boundary: `internal/journal/trade_outcomes_test.go` function `TestABackfillRecoversTheGapAndRefusesToRewriteIt` and its documented derived/test state.
- High-risk impact: no runtime authority.

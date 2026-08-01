# Function Logic Map: `StrategyEntryQuantity`

- Source: `internal/risk/contract.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Guardian policy | validated positive KRW limits | `risk.Policy` | error, quantity not returned |
| entry/stop | canonical positive decimals with `entry > stop` | opaque strategy decision | zero-capacity sentinel or parse error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | exact AST `if` at source line 128: `if err := policy.Validate(); err != nil {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B2 | exact AST `if` at source line 131: `if policy.RiskBudget.Currency != policy.MaxOrderNotional.Currency {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B3 | exact AST `if` at source line 135: `if err != nil {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B4 | exact AST `if` at source line 139: `if err != nil {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B5 | exact AST `if` at source line 143: `if err != nil {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B6 | exact AST `if` at source line 147: `if err != nil {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B7 | exact AST `if` at source line 151: `if err != nil {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B8 | exact AST `if` at source line 154: `if entry.Sign() <= 0 \|\| notional.Sign() <= 0 {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B9 | exact AST `if` at source line 160: `if quantityCap.Cmp(capacity) < 0 {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B10 | exact AST `if` at source line 163: `if notionalCap.Cmp(capacity) < 0 {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B11 | exact AST `if` at source line 166: `if capacity.Sign() <= 0 {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Policy.Validate` | reject incomplete limits | no fallback/default substitution | AST |
| `RiskBasedQuantity` | exact risk-budget cap | one rational floor | AST + unit tests |
| `parseWholeNumber` / `parseDecimal` | canonical exact caps | fail closed | AST |

## State mutations and fallbacks

- Pure arithmetic only; no journal, clock, gateway, or external side effect.
- The final quantity is `min(floor(risk/width), maxQty, floor(notional/entry))`.

## Safety conclusion

- Safe edit boundary: pure Guardian sizing; callers may only reduce the result.
- High-risk impact: yes, because the quantity bounds exposure.

# Function Logic Map: `RequireNoPosition`

- Source: `internal/strategymarket/bars.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| source/request | non-nil canonical exact market/symbol | strategy lane | unavailable refusal |
| reading | exact identity, allowlisted official position source, fresh timestamp | source response | unavailable/stale refusal |
| exposure | canonical exact quantity `0`, no open orders | official account read | blocked refusal |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | exact AST `if` at source line 365: `if source == nil \|\| now.IsZero() \|\| strings.TrimSpace(market) == "" \|\| strings.TrimSpace(symbol) == "" {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B2 | exact AST `if` at source line 369: `if err != nil \|\| reading.Market != market \|\| reading.Symbol != symbol \|\|` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B3 | exact AST `if` at source line 375: `if age < 0 \|\| age > 30*time.Second {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |
| B4 | exact AST `if` at source line 378: `if !validQuantity \|\| decimalString(quantity) != reading.Quantity \|\| quantity.Sign() != 0 \|\| reading.OpenOrders != 0 {` | source-bound local control flow | branch-specific return/continue behavior | see exact evidence row in `branch-test-map.md` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ReadPosition` | one authoritative read | typed refusal; no retry/fallback | AST |
| `exactDecimal` | reject non-canonical quantity assertion | pure | AST |

## State mutations and fallbacks

- Read-only; exact request identity is checked after the source call.

## Safety conclusion

- Safe edit boundary: caller-provided authority strings cannot mint a proof.
- High-risk impact: yes — duplicate-position/order prevention.

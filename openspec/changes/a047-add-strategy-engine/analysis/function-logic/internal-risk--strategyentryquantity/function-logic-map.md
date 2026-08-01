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
| B1 | policy invalid/currency mismatch | none | validation error | policy contract tests |
| B2 | risk quantity or caps cannot parse | none | exact parse error | risk sizing tests |
| B3 | entry/notional non-positive | none | `ErrStrategyQuantityZero` | zero-capacity test |
| B4 | risk cap is minimum | none | whole quantity | exact-minimum table |
| B5 | max quantity is minimum | none | whole quantity | default-cap row |
| B6 | notional floor is minimum | none | whole quantity | notional/boundary rows |
| B7 | final capacity zero | none | `ErrStrategyQuantityZero` | zero-capacity test |

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

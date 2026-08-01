# Function Logic Map: `RiskBasedQuantity`

- Source: `internal/risk/contract.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| budget/entry/stop | canonical decimal values, nonnegative budget | risk policy + decision | parse/negative errors |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | budget parse fails | none | error | contract table |
| B2 | budget negative | none | error | policy tests |
| B3 | entry parse fails | none | error | input tests |
| B4 | stop parse fails or width nonpositive | none | error or zero | stop/zero tests |
| B5 | positive width | none | exact rational floor | sizing tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| exact decimal parser | avoid float rounding | fail closed | AST |
| big.Rat/big.Int | quotient and single floor | deterministic | AST + tests |

## State mutations and fallbacks

- Pure arithmetic with no side effects.

## Safety conclusion

- Safe edit boundary: exact risk cap computation.
- High-risk impact: yes, bounds exposure.

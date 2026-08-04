# Function Logic Map: `DefaultStaleness`

- Source: `internal/execgw/retry.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| account-wide required-query catalog | open orders, buying power, holdings and price only | landed retry matrix | missing/unseen entries block account entry through `EntryGate` |
| strategy authority reads | market-local KR account identity and US FX | `StrategyReadBudgets`, not this account-wide map | MUST NOT widen a peer market failure into an account-wide block |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | direct return | none; returns a fresh map | four account-wide thresholds | retry-matrix tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| none | static account-wide threshold catalog | no error, timeout, retry or I/O | CodeGraph + AST |

## State mutations and fallbacks

- No mutation. Market-local strategy read freshness is deliberately defined separately so US FX
  staleness cannot cancel KR evaluation or safety-class loops.

## Safety conclusion

- Safe edit boundary: preserve existing account-wide values; add no FX query to this global map.
- High-risk impact: **yes** — entry freshness gate; unchanged behavior is explicitly tested.

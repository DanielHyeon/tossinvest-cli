# Function Logic Map: `Console.handleStrategyRuntime`

- Source: `internal/console/strategy_runtime.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| HTTP method | GET or HEAD only | request | 405, no reader call |
| StrategyRuntime reader | nil or one read-only reader | Console.Options | nil/error/invalid projects paired dormant/unavailable OFF state |
| Runtime snapshot | server-owned and structurally valid | injected reader | invalid snapshot is never partially trusted |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | method is not GET/HEAD | response only | 405 | authenticated GET-only test |
| B2 | reader is nil | none | paired dormant projection | dormant console test |
| B3 | reader returns error or invalid snapshot | none | paired typed unavailable projection | console invalid/error test |
| B4 | reader returns valid paired snapshot | none | render exact independent markets | partial-market console test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `StrategyRuntimeReader.Read` | obtain plain read model | one request-scoped call, no retry, no mutation | CodeGraph + AST |
| `strategyRuntimePage.project` | presentation-only mapping | consumes validated model | CodeGraph + AST |
| `Console.render` | render authenticated HTML | existing console response contract | CodeGraph + AST |

## State mutations and fallbacks

- Only the response is written. No configuration, activation, order, journal or protection capability is present.
- Failures select explicit paired/market-local unknown models; they do not copy peer values.

## Safety conclusion

- Safe edit boundary: replace the single-market read model with a validated paired read model while retaining method/auth behavior.
- High-risk impact: no; read-only operator projection, with fail-closed entry state.

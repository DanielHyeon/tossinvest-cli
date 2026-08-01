# Function Logic Map: `consoleBroker.resolve`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| shared console broker/read request | typed non-mutating seam | console assembly and a043 scope contract | resolution/read errors remain explicit and no evidence is inferred |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if at line 685 | cache/read-only assembly only | current seam tests |
| B2 | if at line 689 | cache/read-only assembly only | current seam tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| broker resolver/read methods in AST | Cache one live broker plus canonical account reference; preserve single account-resolution call across console seams. | existing lazy retry/cache semantics; no new retry | CodeGraph impact + AST |

## State mutations and fallbacks

- Cache one live broker plus canonical account reference; preserve single account-resolution call across console seams.
- Only process-local resolver/cache state changes; account, journal, orders, and operating toggles are never mutated.

## Safety conclusion

- Safe edit boundary: carry canonical read identity across an existing read-only seam.
- High-risk impact: no live trading method is reachable from the console seam.

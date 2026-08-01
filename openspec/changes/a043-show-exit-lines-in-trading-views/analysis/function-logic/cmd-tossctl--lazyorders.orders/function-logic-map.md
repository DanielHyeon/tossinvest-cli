# Function Logic Map: `lazyOrders.Orders`

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
| B1 | if at line 810 | cache/read-only assembly only | current seam tests |
| B2 | if at line 814 | cache/read-only assembly only | current seam tests |
| B3 | if at line 830 | cache/read-only assembly only | current seam tests |
| B4 | else at line 832 | cache/read-only assembly only | current seam tests |
| B5 | if at line 845 | cache/read-only assembly only | current seam tests |
| B6 | else at line 847 | cache/read-only assembly only | current seam tests |
| B7 | if at line 856 | cache/read-only assembly only | current seam tests |
| B8 | else at line 858 | cache/read-only assembly only | current seam tests |
| B9 | range at line 861 | cache/read-only assembly only | current seam tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| broker resolver/read methods in AST | Carry the resolved account identity beside broker order lists without adding broker calls or mutation capability. | existing lazy retry/cache semantics; no new retry | CodeGraph impact + AST |

## State mutations and fallbacks

- Carry the resolved account identity beside broker order lists without adding broker calls or mutation capability.
- Only process-local resolver/cache state changes; account, journal, orders, and operating toggles are never mutated.

## Safety conclusion

- Safe edit boundary: carry canonical read identity across an existing read-only seam.
- High-risk impact: no live trading method is reachable from the console seam.

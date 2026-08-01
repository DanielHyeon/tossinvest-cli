# Function Logic Map: `lazyHoldings.Holdings`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| shared resolver | configured, read-only broker | console assembly | return resolver error; no broker call |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | resolver fails | none | return error | console broker seam tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `consoleBroker.resolve` | reuse broker/account resolution | retry remains cache-bounded | AST + console seam tests |

## State mutations and fallbacks

- Reads holdings through one capability-only method value; no account or order mutation.

## Safety conclusion

- Safe edit boundary: discard the newly carried account value on the holdings seam.
- High-risk impact: no mutation path is added.

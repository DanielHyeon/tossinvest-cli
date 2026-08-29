# Function Logic Map: `Gateway.Place`

- Source: `internal/execgw/gateway.go`
- Qualified function: `Gateway.Place`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `PlaceRequest` | canonical order shape plus durable Guardian reference | caller shape; journal decision remains authority | private helper/Gateway returns typed refusal |
| strategy capability | always nil on this public ordinary path | compile-time call site | q_final first-leg authority cannot be forged here |

## Branches and early returns

`Gateway.Place` is branchless. Its single path delegates to `Gateway.place(ctx, req, nil)` and returns
the exact outcome/error unchanged. All mutation branches remain in the private helper and `Gateway.submit`.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Gateway.place` | construct one canonical PLACE plan and enter the sole submit path | no retry/fallback; strategy capability is fixed nil | AST + ordinary Place regression suites |

## State mutations and fallbacks

- No state is mutated in the wrapper.
- There is no fallback to the strategy path, raw trading service, paper, shadow or canary transport.

## Safety conclusion

- Safe edit boundary: preserve a branchless wrapper with a literal nil strategy capability.
- High-risk impact: yes — this is the public official mutation entry, although authority is enforced downstream.

# Function Logic Map: `Console.handlePositionPolicyPreview`

- Source: `internal/console/position_policy.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| selection token | console-signed display selection | server-rendered action | forbidden on tamper |
| engine preview | exact before/after + opaque capability | engine service | no apply form on failure |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B5 | seam/token/preview/capability/danger-wait branches | render only | HTTP refusal or preview | console capability tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `PositionPolicies.Preview` | exchange selection for engine mutation capability | no retry | AST |

## State mutations and fallbacks

- Browser apply form carries the opaque engine token; console does not mint mutation authority.

## Safety conclusion

- Safe edit boundary: retain input-free StockOS-style preview while removing console HMAC as Apply authority.
- High-risk impact: yes

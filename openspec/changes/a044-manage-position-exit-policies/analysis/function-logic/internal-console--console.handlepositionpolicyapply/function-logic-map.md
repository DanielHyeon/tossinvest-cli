# Function Logic Map: `Console.handlePositionPolicyApply`

- Source: `internal/console/position_policy.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| capability form field | opaque engine token only | engine preview response | missing/invalid refused by engine |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B3 | seam/capability/apply-error branches | engine call only with opaque token and boolean acknowledgement | HTTP result | console capability tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `PositionPolicies.Apply` | forward token-only DTO | no retry | AST |

## State mutations and fallbacks

- Console cannot select position/generation/version/action/policy/reason at Apply time. The engine interprets `Confirmed` only against the action stored in its grant.

## Safety conclusion

- Safe edit boundary: forward only the engine-issued opaque capability and never regenerate a mutation request.
- High-risk impact: yes

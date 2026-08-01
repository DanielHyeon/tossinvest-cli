# Function Logic Map: `PositionPolicyCommandService.Preview`

- Source: `internal/app/engine/position_policy_command.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| request | server-defined scope/action/reason | console selection DTO | reject invalid/stale/ineligible |
| engine instance | random per process | command service | no capability on failure |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B3 | authoritative preparation, preview, or issuance fails | no usable capability | error | capability tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `prepare` | discard client authority and derive re-adopt observation | fail closed | AST |
| `PreviewPositionPolicy` | compute exact before/after | no mutation | AST |
| capability issuer | mint one-time random token bound to exact preview | crypto-random failure aborts | AST |

## State mutations and fallbacks

- Stores only a digest plus exact immutable request/before/after and timing under the engine service mutex.

## Safety conclusion

- Safe edit boundary: mint only after authoritative preview; bind instance/domain/scope/before-after and minimum delay.
- High-risk impact: yes

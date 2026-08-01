# Function Logic Map: `PositionPolicyCommandService.Apply`

- Source: `internal/app/engine/position_policy_command.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| opaque capability | fixed-length crypto-random token | engine preview | invalid/early/expired/replayed refused |
| stored grant | exact engine instance/domain/request/before/after | engine memory | mismatch consumed fail-closed |
| confirmation | boolean acknowledgement only; no scope | danger checkbox | required only when bound action is RELEASE/READOPT |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B6 | verify, confirmation, current-state, apply, or result validation fails | confirmation omission preserves grant; authorized attempts consume it fail-closed | typed error | capability matrix |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| capability verifier | constant-time digest match and timing/scope checks | no fallback | AST |
| `ApplyPositionPolicy` | commit stored exact request under journal CAS | on any attempted failure grant stays consumed | AST |

## State mutations and fallbacks

- Mutex serializes verification, consume-on-attempt, journal CAS, and result verification; concurrent duplicates cannot both enter the repository.

## Safety conclusion

- Safe edit boundary: accept token-only Apply and never reconstruct authority from client-provided scope fields.
- High-risk impact: yes

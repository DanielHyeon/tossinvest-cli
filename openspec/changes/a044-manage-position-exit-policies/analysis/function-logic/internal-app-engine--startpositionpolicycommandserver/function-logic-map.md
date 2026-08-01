# Function Logic Map: `StartPositionPolicyCommandServer`

- Source: `internal/app/engine/position_policy_transport.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| engine directory/service | existing, owner-controlled, no symlink traversal or group/other write | engine boot | refuse before bind |
| transport bearer | 256-bit random | crypto/rand | startup fails on entropy error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B9 | service/path/private-directory setup and cleanup | create only dedicated 0700 directory | fail closed before or during bind | insecure filesystem tests |
| B10-B11 | bind or entropy failure | close listener/remove newly created directory | wrapped startup error | constructor contract |
| B12-B14 | GET-only health/list and typed list failure | response only | 405/typed error | endpoint integration tests |
| B15 | descriptor publication failure | close listener/remove new directory | publication error | descriptor tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `net.Listen` | loopback endpoint | no non-loopback fallback | AST |
| `ValidateEngineDirectory`/`ValidateControlDirectory` | establish owner-only filesystem authority | any uncertainty fails startup | AST |
| `writePositionPolicyDescriptor` | publish private capability file | validated atomic rename | AST |

## State mutations and fallbacks

- Endpoint lifetime remains nested under engine run; Close removes the descriptor and dedicated directory. Apply accepts only opaque capability DTO.

## Safety conclusion

- Safe edit boundary: keep bearer transport authentication separate from mutation authorization.
- High-risk impact: yes

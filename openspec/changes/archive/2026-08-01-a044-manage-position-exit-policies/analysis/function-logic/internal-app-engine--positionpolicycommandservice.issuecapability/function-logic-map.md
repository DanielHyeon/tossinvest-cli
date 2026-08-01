# Function Logic Map: `PositionPolicyCommandService.issueCapability`

- Source: `internal/app/engine/position_policy_command.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| request/preview | server-prepared exact before/after state | command service | no grant on error |
| outstanding grants | fewer than 256 unexpired grants | in-memory engine instance | refuse capacity overflow |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | instance identity initialization fails | none | error | capability tests |
| B2-B3 | prune expired grants | replace in-memory slice with active grants | continue | expiry tests |
| B4 | active grant limit reached | none | error | bounded capability contract |
| B5 | random grant generation fails | none | wrapped error | entropy contract |
| B6 | RELEASE/READOPT danger action | set 3-second not-before | success | danger delay tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ensureInstanceID` | bind grant to process instance | error aborts issue | AST |
| `rand.Read`/`sha256.Sum256` | create opaque one-time authority stored only as digest | no predictable token fallback | AST |

## State mutations and fallbacks

- Stores exact request/before/after, issue time, danger delay, and expiry; no client-provided mutation scope is added later.

## Safety conclusion

- Safe edit boundary: preserve random opacity, bounded storage, exact state binding, and action-specific delay.
- High-risk impact: yes

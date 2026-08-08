# Function Logic Map: `Console.routes`

- Source: `internal/console/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| route table | every route session-gated | static route invariant | static test failure |
| state-changing routes | session + CSRF `mutating` | static mutation invariant | HTTP refusal |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | GET/read routes | no mutation wrapper | handler response | existing static tests |
| B2 | POST/mutating routes | session and CSRF wrappers | handler response/refusal | existing static tests |
| B3 | signed download POST | session + CSRF, but not engine `startExclusive` | handler serializes release work itself | new route tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `session0` | authenticates process-local session | rejects absent/wrong session | CodeGraph + AST |
| `mutating` | enforces POST and CSRF | rejects before handler | CodeGraph + AST |
| `handleSystemUpdateDownload` | signed release stage action | no automatic install/relaunch | new handler |

## State mutations and fallbacks

- Registers one fixed path. No dynamic path or selector is added.

## Safety conclusion

- Safe edit boundary: add one exact route with both standard gates.
- High-risk impact: yes; route-table static guards must enumerate it.

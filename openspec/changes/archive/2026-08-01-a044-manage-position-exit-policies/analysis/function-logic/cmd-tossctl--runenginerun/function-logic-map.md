# Function Logic Map: `runEngineRun`

- Source: `cmd/tossctl/engine.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| command context/root | explicit engine profile | Cobra/root config | resolution/startup error |
| journal directory lock | acquired before assembly/open | `enginelock` | refuse second engine |
| assembled context | verified automation and owned journal | engine interlock | close on return |
| local control endpoint | loopback + per-process bearer descriptor | engine process | absence leaves console unwired |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | directory/lock fails | no engine assembly | error | engine boot tests |
| B2 | assembly/interlock fails | close/release | typed error | engine boot tests |
| B3 | automation unverified | close/release | gate-off | engine boot tests |
| B4 | verify lock fresh | close/release | busy refusal | engine tests |
| B5 | marker unavailable | warning only | continue | marker tests |
| B6 | runtime/control server build fails | cleanup | error | control transport tests |
| B7 | run | loops + local control server until context ends | runtime result | engine lifecycle tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `enginelock.Acquire` | sole engine writer exclusion | first action, no wait | CodeGraph + AST |
| `engineAssemble` | opens/migrates journal under lock | fail closed | CodeGraph + AST |
| `engineRuntimeFactory` | builds loops | fail closed | CodeGraph + AST |
| `engine.StartPositionPolicyCommandServer` | publish authenticated local command boundary | engine-owned lifecycle | CodeGraph + AST |

## State mutations and fallbacks

- Engine owns journal, listener, descriptor and cleanup. Console never receives a journal path as a write capability.

## Safety conclusion

- Safe edit boundary: start the control server only after verified assembly and stop/remove it before context close.
- High-risk impact: yes — process and journal ownership boundary.

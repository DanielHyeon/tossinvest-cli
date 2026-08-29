# Function Logic Map: `runEngineRun`

- Source: `cmd/tossctl/engine.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| command context | non-nil; background is the defensive fallback | Cobra | cancellation propagates into assembly and runtime |
| config/journal directory | resolved profile directory | `engineJournalDir` | resolution error before lock or engine open |
| engine context | verified production assembly | `engineAssemble` | startup refusal; no loop starts |
| runtime clock | one injected/system clock | command boot sequence | shared by marker, schedule assembly and safety loops |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | journal directory resolution fails | none | exact error | engine command tests |
| B2 | exclusive journal lock fails | none | lock error | lock-before-assembly test |
| B3 | engine assembly/interlock fails | closes partial assembly internally | typed gate/interlock error | gate/interlock tests |
| B4 | automation is not verified | engine context closes; no loop starts | `errEngineGateOff` | gate-off test |
| B5 | fresh live-verify runlock exists | no runtime starts | `errVerifyInProgress` | runlock test |
| B6 | advisory marker fails | prints note; lock remains authoritative | continue | marker test |
| B7 | runtime or command services fail to assemble | no runtime loop starts | exact error | runtime branch tests |
| B8 | successful boot | holds marker/control servers and runs until cancellation | runtime result | lifecycle tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `enginelock.Acquire` | single writer before journal open/migration | fail startup | AST |
| `engineAssemble` | build official-only engine profile and interlock | fail closed | AST |
| `engineRuntimeFactory` | build safety loops and paired KR/US entry outer loop using the same request context | failure aborts before `Runtime.Run` | AST + engine runtime tests |
| position-policy servers | expose bounded local policy control/read surfaces | failure aborts startup; deferred close | AST |
| `Runtime.Run` | supervise all loops through cancellation/drain | exact runtime semantics | runtime tests |

## State mutations and fallbacks

- The journal lock is acquired before engine assembly and released by defer.
- Neither marker failure nor read-only status surfaces grant entry authority.
- The request context handed to paired KR/US schedule assembly is the same boot context that can cancel blocked official calendar reads.
- No live order is emitted during construction; any later mutation remains behind the existing Guardian/Gateway and human-enabled automation gate.

## Safety conclusion

- Safe edit boundary: context propagation into runtime assembly only; preserve lock/interlock/runlock/marker/control-server order and stop discipline.
- High-risk impact: yes — production boot path; bounded by pre-loop fail-closed returns and unchanged Gateway authority.

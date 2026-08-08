# Function Logic Map: `acquireVerifyExecutionLock`

- Source: `cmd/tossctl/verify.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| root/config | nil/default or isolated config directory | `engineJournalDir` | path resolution error refuses verification |
| lock owner | none, or another engine/update/verification process | kernel flock on `engine.lock` | contention refuses immediately and wraps `ErrAlreadyRunning` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | journal-directory resolution fails | none | wrapped path error | existing path tests |
| B2 | real flock cannot be acquired | closes attempted descriptor inside `enginelock` | wrapped exclusion error | `TestVerificationEntryPointsRefuseWhileSystemUpdateOwnsEngineExclusion` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `engineJournalDir` | use exactly the engine/updater lock directory | local path resolution only | source + AST |
| `enginelock.Acquire` | obtain crash-released cross-process exclusion | nonblocking fail-closed | `enginelock` tests + contention regression |

## State mutations and fallbacks

- The helper creates no account or broker capability.
- A successful return transfers one held lock to the caller, which must defer `Release`.
- Kernel descriptor cleanup releases the lock on abrupt process death.

## Safety conclusion

- Safe edit boundary: path resolution and flock acquisition only; no fallback to advisory markers.
- High-risk impact: yes — callers reach live verification orders, so any lock failure must refuse.

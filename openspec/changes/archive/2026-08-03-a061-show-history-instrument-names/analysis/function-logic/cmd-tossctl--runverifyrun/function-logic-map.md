# Function Logic Map: `runVerifyRun`

- Source: `cmd/tossctl/verify.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| list/run options | preview or supervised KR/US verification | Cobra flags | preview remains read-only and takes no locks/credentials |
| context | command context plus SIGINT | Cobra/signal | cancellation preserves evidence and cleanup |
| exclusions | engine flock, profile intent marker, then rate-budget lease | enginelock + runlock + ratebudget | any failure returns before broker or runner activity |
| prior evidence | empty/resumable/explicit redo | verify record | unsafe implicit restart refused |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | list preview | writes catalogue only | nil | existing list test |
| B2 | missing command context | substitute background | continue | existing command test |
| B3 | execution exclusion fails | none | return | existing exclusion tests |
| B4 | record resolution fails | none | return | verify path tests |
| B5 | record load fails | none | return | corrupt record tests |
| B6 | prior measured steps without resume/redo | none | explicit refusal | replay refusal tests |
| B7 | profile intent marker path fails | execution lock released | return | profile path tests |
| B8 | rate-budget lease fails/cancels | marker is already visible; no broker construction | return | A061 occupied-budget entrypoint test |
| B9 | broker/account resolution fails | exclusions released | return | broker failure tests |
| B10 | holding symbol omitted | read first usable holding | continue | KR/US symbol tests |
| B11 | KR default used for US market | replace with US holding | continue | US verification tests |
| B12 | recorder open fails | exclusions released | return | record tests |
| B13 | runner construction fails | exclusions/recorder released | return | verify options tests |
| B14 | run canceled/deadline | retain record and print interrupted | nil | interrupt tests |
| tail | runner completes | summary printed; defers release marker/leases | runner error | full verify suite |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `acquireVerifyExecutionLock` | exclude engine/update/verification | immediate refusal | existing tests |
| `holdVerifyRateBudgetIntent` | make verification priority visible before lease waiting | advisory marker; unique owner under execution flock | A061 contention tests |
| `acquireVerifyRateBudget` | exclude optional metadata | waits with SIGINT-aware context before broker | A061 tests |
| `verifyBrokerFactory`, `verifylive.New`, `runner.Run` | existing official supervised verification | confirmation and cleanup contracts unchanged | full verify suite |

## State mutations and fallbacks

- The new intent marker and kernel lease are deferred through complete runner cleanup.
- Preview mode still reaches no credential, lock, broker, or account.

## Safety conclusion

- Safe edit boundary: one admission step before existing live broker construction.
- High-risk impact: yes; full verify tests and Function Logic Map coverage are mandatory.

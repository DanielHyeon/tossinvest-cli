# Function Logic Map: `runVerifyAbort`

- Source: `cmd/tossctl/verify.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| evidence record | selected market or explicit override | `loadVerifyRecord` | list reads locally; live abort reloads only after exclusive admission |
| abort targets | only outstanding artifacts recorded by this tool | `verifylive.AbortTargets` | empty target list returns without credentials/network |
| profile exclusions | execution flock, run-intent marker, then profile rate-budget lease | `acquireVerifyExecutionLock`, `holdVerifyRateBudgetIntent`, `acquireVerifyRateBudget` | active engine/verification is refused immediately; cancellation returns before broker |
| operator authority | `--list` or one existing expiring batch confirmation | Cobra options + runner | no implicit approval; list never mutates |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `--list` | local-only preview path | return before all exclusions/credentials | existing abort list tests |
| B2 | list record load fails | none | return error | existing abort record tests |
| B3 | command context is nil | none | use background context | command wiring tests |
| B4 | engine/update/verification owns execution flock | none; existing owner's marker untouched | immediate refusal | A061 active-verification exclusion test |
| B5 | profile intent path fails | release execution flock | return error | profile path tests |
| B6 | rate-budget acquisition is canceled/fails | marker is already published; release it | return error before broker | A061 occupied-budget entrypoint test |
| B7 | exclusive record reload fails | release all exclusions | return error | existing record tests |
| B8 | refreshed record has no target | print honest empty state | return without broker | empty-record abort tests |
| B9 | broker/account construction fails | release marker and leases | return error | refreshed-target test |
| B10 | recorder open fails | release exclusions | return error | record permission tests |
| B11 | abort runner construction fails | close recorder and release exclusions | return error | verifylive option tests |
| B12 | operator declines with reason | print reason | continue to normal result handling | existing abort approval tests |
| B13 | approved abort leaves no target | print completion | continue | verifylive abort tests |
| B14 | abort is canceled/deadlines | preserve synced evidence and print interruption | return nil | existing interrupt tests |
| tail | abort completes or returns non-context error | recorder/exclusions release via defer | return runner error | full verify command suite |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `loadVerifyRecord` / `AbortTargets` | derive the exact tool-owned cleanup set | live path reloads after exclusive admission | AST B1-B2, B7-B8 |
| `acquireVerifyExecutionLock` | refuse overlap with engine/update/active verification | nonblocking kernel flock | active-verification test |
| `holdVerifyRateBudgetIntent` | publish cleanup priority before waiting for optional metadata | unique owner under execution flock | A061 occupied-budget entrypoint test |
| `acquireVerifyRateBudget` | exclude all optional metadata across processes | cancellable wait; always before broker | AST B8 + ratebudget tests |
| `verifyBrokerFactory` / `verifylive.New` | reuse the existing supervised abort path | existing approval and artifact ownership rules unchanged | abort suite |

## State mutations and fallbacks

- Adds admission ordering; live target ownership is recomputed after exclusivity.
- The execution flock, active-profile marker, and lease are released on every post-acquisition return.
- An explicit record override selects evidence only; it cannot move the shared profile lease.

## Safety conclusion

- Safe edit boundary: exclusions are acquired and targets refreshed before the existing broker/runner path; approval still covers the exact runner target set.
- High-risk impact: yes, because cleanup of live tool-owned artifacts must retain priority and authorization.

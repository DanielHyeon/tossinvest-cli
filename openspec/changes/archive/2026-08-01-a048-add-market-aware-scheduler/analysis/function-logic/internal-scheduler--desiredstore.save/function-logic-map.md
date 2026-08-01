# Function Logic Map: `DesiredStore.Save`

- Source: `internal/scheduler/desired.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| desired | valid state carrying expected `Revision` | caller's last Load/default | stale expected revision conflicts |
| process exclusion | cancellable nonblocking flock loop beside state file | `acquireDesiredLock(ctx, ...)` | caller cancellation or 2s bound refuses save |
| durability | 0600 same-directory temp, fsync, rename, directory sync | filesystem | any failure returns without claiming success |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B2 | context canceled or desired invalid/future | none | return error before filesystem mutation | validation/context tests |
| B3-B4 | directory or cross-process lock acquisition fails | maybe directory creation | return error | platform/lock contract |
| B5-B6 | canceled after lock or current state invalid | none | return error under released lock | context/strict load tests |
| B7 | current revision differs from expected | none | `ErrDesiredRevisionConflict` | OFF preservation test |
| B8 | revision exhausted | none | refuse wraparound | boundary contract |
| B9-B15 | marshal/temp/chmod/write/sync/close failures | temp only, deferred cleanup | return error | atomic-save contract |
| B16 | context canceled before commit | temp removed | context error | context contract |
| B17 | rename failure | old file preserved, temp removed | return error | atomic-save contract |
| final | directory open/sync succeeds | atomically installs revision+1 | nil | round-trip/CAS tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `acquireDesiredLock` | serialize read-compare-rename across processes | caller cancellation and 2s maximum; crash released by kernel | subprocess/cancellation/bounded-wait tests + AST |
| `loadAt` | read current revision after acquiring lock | strict fail-closed decode | CodeGraph + AST |
| temp/fsync/rename/directory sync | durable atomic replacement | errors propagated, no retry | CodeGraph + AST |

## State mutations and fallbacks

- Mutation is one CAS transaction under a cross-process lock. A committed OFF at revision N+1 cannot be overwritten by stale ON at revision N.

## Safety conclusion

- Safe edit boundary: scheduler desired-state file only; no activation or runtime toggle occurs.
- High-risk impact: yes, because persisted ON is restart authority input.

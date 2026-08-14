# Function Logic Map: `armExitProposalTx`

- Source: `internal/journal/apply_hook.go`
- Frozen-base AST evidence: `ast.json` (4 branches; revision `base`; source SHA-256 recorded by extractor). The function body is unchanged; the adjacent helper addition changed only the containing current-file SHA.
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| transaction | caller-owned journal write transaction | judgement/apply-hook boundary | any error is returned for caller rollback |
| position/proposal | existing exit state with no outstanding proposal | durable `exit_states` row | absent/read failure/pending proposal refuses the arm |
| guarded columns | pending action/level/intent are a single crash-safety tuple | `apply_hook.go` single-writer authority | no production file outside apply hook may name a guarded-column write |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | proposal-row SELECT returns no row at `internal/journal/apply_hook.go:660` | no write | typed `ErrExitStateNotFound` | `TestGuardedExitColumnsAreWrittenOnlyByTheApplyHook` |
| B2 | proposal-row SELECT returns another error at `internal/journal/apply_hook.go:663` | no write | wrapped read error | `TestGuardedExitColumnsAreWrittenOnlyByTheApplyHook` |
| B3 | trimmed pending action is already nonempty at `internal/journal/apply_hook.go:666` | preserves outstanding proposal unchanged | typed `ErrProposalPending` | `TestASecondProposalIsRefusedWhileOneIsOutstanding` |
| B4 | guarded tuple UPDATE fails at `internal/journal/apply_hook.go:669` | caller transaction rolls back | wrapped arm error | `TestAFailingExitHookRollsBackTheProjectionToo` |
| Return | no outstanding proposal and UPDATE succeeds | atomically arms pending action/level/intent in caller transaction | nil | `TestExitSnapshotDuplicateDecisionIsNotRearmed` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `QueryRowContext` | check current pending authority before arming | no row/read error refuses mutation | AST |
| `ExecContext` | write the guarded pending tuple before submission | transaction-owned; any error returned for rollback | AST + apply-hook rollback tests |
| `nullableString` | preserve empty intent as SQL NULL | representation only | AST |

## State mutations and fallbacks

- This is the arming writer for the pending triple and remains in `apply_hook.go`.
- Existing pending evidence is never overwritten; the whole judgement transaction rolls back on refusal.
- The new refresh guard is adjacent but read-only, so it does not enlarge this writer's authority.

## Safety conclusion

- Safe edit boundary: existing durable-before-submit proposal arm; no new order call.
- High-risk impact: yes—pending proposal crash/deduplication authority, structurally pinned by `TestGuardedExitColumnsAreWrittenOnlyByTheApplyHook`.

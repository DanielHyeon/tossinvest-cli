# a121 contract hard evidence

**Scope.** This is a pre-implementation inventory for
`a121-reconcile-stale-verification-artifacts`. It was prepared from the frozen
base `9408fc957fbb8687fdf75fd8a58cded50bf4f687`. It does not read a local
verification record, invoke `tossctl`, or make an account/API request.

## Contract facts tied to current code

1. The record is append-only JSONL: `verifylive.Recorder.Append` persists an
   `Entry`; `LoadEntries` reads it. `Entry` contains a masked account reference,
   observations, calls, and artifacts (`internal/verifylive/record.go:247-457`).
   Existing artifacts have only the terminal facts `Cancelled` and `Filled`
   (`Artifact.terminal`, line 575). A reconciliation fact therefore needs a
   distinct, versioned append-only representation; it must not rewrite the
   cleanup entry or turn the existing failed DELETE into a successful call.
2. `Outstanding` is the single record projection (`record.go:506-575`).
   `outstandingLines` resolves the latest `(kind, id)` line and refuses to
   resurrect an artifact after a monotone terminal fact. It feeds the runner,
   abort path, report, status, exposure check, and cleanup code. Any
   `reconciled-absent` representation must be consumed here or by an equally
   central projection without calling it a cancellation or fill.
3. `PendingCleanup` and `cleanupFrom` (`cleanup.go:119-135`) derive resume
   cleanup targets from `outstandingLines`; `Runner.cleanupTargets` uses the
   same rule at `cleanup.go:102-106`. `Runner.cleanup` is called only by
   `Runner.Run` (`runner.go:462`), and its `runCleanup` can call
   `CancelConditionalOrder`. The reconciliation command must not reuse this
   mutation path. Its test must prove the reconciled exact artifact disappears
   from `PendingCleanup`/resume planning while no other artifact does.
4. The existing only-narrow read interface is
   `m0RawConditionalPageReader` (`m0_recovery.go:12-16`).
   `Runner.m0RecoverPending` already performs bounded, cursor-checked reads of
   both `OPEN` and `CLOSED` groups with a page cap and rejects empty/repeated
   cursors (`m0_recovery.go:128-176`). Its semantics are pending-create
   recovery, not stale-cleanup reconciliation, so a121 must not repurpose its
   checkpoint, trigger, or receipt writes. A dedicated read-only interface and
   operation may use the same narrow `ProtectionConditionalOrdersRaw` shape.
5. `official.Client.ProtectionConditionalOrdersRaw` is the account-scoped
   official raw reader. The related `ConditionalOrdersRaw` validates a required
   status, preserves raw opaque fields, passes status/symbol/cursor/limit to
   `GET /api/v1/conditional-orders`, and exposes `HasNext`/`NextCursor`
   (`internal/official/conditional_reads.go:136-210`). It is the appropriate
   adapter boundary for exact opaque-ID and pagination tests. The command must
   read both allowed groups completely; a generic status screen or a DELETE 404
   is not an absence proof.
6. The CLI already has local-only `verify status`/`verify report` loaders
   (`cmd/tossctl/verify.go:699-755`) and account binding for live verification:
   `buildVerifyBroker` plus `resolveVerifyAccount` bind the displayed masked
   reference and selected API account sequence from the same account entry
   (`verify.go:874-955`). a121 must keep reconciliation record/profile/account
   selection explicit. A new read-only command must not be implemented as a
   `verify run` flag, an abort call, a redo, or a cleanup retry.
7. Reports and status currently expose `Outstanding` directly:
   `BuildReport` (`report.go:165-215`) and `BuildProgress`
   (`report.go:314-342`). If a121 exposes a reconciled state to operators,
   it must be visibly labelled `reconciled absent`, not cancelled, filled, or a
   successful measurement. `BuildReport` only derives measured attributes from
   step observations (`report.go:169-210`), so the new event must not enter that
   input.

## Existing function candidates and mandatory maps

All rows marked **mandatory if edited** require both a Go AST Function Logic
Map and Branch Test Map before RED work. This is a high-risk reconciliation
change; there is no high-risk exemption for modifying an existing function.

| Candidate | Current role and a121 boundary | Map requirement |
| --- | --- | --- |
| `verifylive.Artifact.terminal` (`record.go:575`) | Defines terminal artifact state for every `Outstanding` consumer. | **Mandatory if edited.** Prefer a separate reconciliation projection/event so this function does not falsely equate absence reconciliation with cancel/fill. |
| `verifylive.outstandingLines` / `Outstanding` (`record.go:506-575`) | Last-mention, monotone artifact projection; shared by reports, status, cleanup, abort, and exposure enforcement. | **Mandatory if edited.** Expected target if reconciled artifacts stop being outstanding. |
| `verifylive.PendingCleanup` / `cleanupFrom` (`cleanup.go:119-135`) | Resume cleanup selection. | **Mandatory if edited.** Map must show exact-artifact removal, held conditional behavior, and all unrelated targets. |
| `(*verifylive.Runner).cleanupTargets`, `planCleanup`, `runCleanup`, `cleanup` (`cleanup.go:102-309`) | Existing approved live cancellation path. | **Mandatory if edited.** Prefer no edit and prohibit reuse; any edit must map every mutation/approval branch. |
| `verifylive.BuildProgress` / `BuildReport` (`report.go:165-342`) | Operator-visible outstanding projection and measured-attribute aggregation. | **Mandatory if edited.** Tests must prove reconciliation is labelled distinctly and cannot affect attestation/endpoint success. |
| `newVerifyCmd` (`cmd/tossctl/verify.go:96-129`) | Registers all verify subcommands. | **Mandatory if edited.** Registering a new reconciliation command changes an existing high-risk CLI control function. |
| `loadVerifyRecord`, `resolveVerifyRecordFor`, `buildVerifyBroker`, `resolveVerifyAccount` (`cmd/tossctl/verify.go:737-955`) | Record/profile resolution and account identity binding. | **Mandatory for each edited function.** If unchanged and called by a new leaf command, record the caller evidence and an explicit not-modified rationale. |
| `(*official.Client).ProtectionConditionalOrdersRaw` / `ConditionalOrdersRaw` | Account-scoped raw official page reader. | **Mandatory if edited.** Reuse without edit is preferable; a new adapter must retain its own direct interface/transport tests. |
| `(*verifylive.Runner).m0RecoverPending` (`m0_recovery.go:128-200`) | M0-only recovery with receipt/checkpoint effects. | **Do not edit for a121.** A new reconciliation operation must not inherit M0 side effects. If scope forces an edit, maps are mandatory and change scope must be re-reviewed. |

New leaf functions/types have no frozen-base Function Logic Map obligation, but
must have a stated `Function Logic Map: not-applicable (new leaf...)` rationale
only if they do not modify any existing function. They still require focused
RED/GREEN tests and a high-risk Pre-Edit Gate before code changes.

## Focused existing test candidates

| Contract check | Existing candidates to preserve or extend |
| --- | --- |
| Failed cleanup remains outstanding and is not a pass | `internal/verifylive/cleanup_test.go:182` `TestAFailedCleanupIsRecordedAndDoesNotStopTheRun`; `record_test.go:114` `TestOutstandingNetsCreationAgainstCancellation`; `record_test.go:143` `TestCancellationIsMonotone`. |
| Resume/cleanup affects only valid targets | `cleanup_test.go:129` `TestTheConditionalLeftForPersistenceIsNotCleanedUp`; `cleanup_test.go:291`, `:307`, `:321`; `hold_test.go:46`, `:61`, `:359`; `plan_symbol_test.go:42`, `:216`. |
| Complete, fail-closed raw OPEN+CLOSED pagination | `m0_recovery_test.go:125` `TestM0RecoveryRejectsRawFieldMismatchAndBrokenPagination`; `internal/official/conditional_reads_test.go:48` and `:94`; `internal/verifylive/us_market_test.go:385`. New tests must cover absent exact ID, present exact ID, wrong account/profile, wrong type, empty/repeated cursor, page cap, read failure, stale snapshot, and duplicate/idempotent event. |
| No mutation reachable from reconciliation | `cmd/tossctl/verify_test.go:272` `TestVerifyRunIssuesNoMutatingRequestWithoutATerminal`; `:347` list sends nothing; `internal/verifylive/m0_recovery_test.go:31`, `:71`, `:88`, `:125`; and `m0_causal_test.go:189` mutation-surface guard. Add a dedicated fake transport assertion that reconciliation performs only official GET reads and one local append. |
| CLI record guard, status, and report boundaries | `cmd/tossctl/verify_test.go:321` restart guard; `:466`, `:480`, `:502`, `:518`; `internal/verifylive/report.go` tests. New tests must prove no `verify run`/`abort`/`redo` route, redacted output, exact selected record, and no attestation or engine-interlock promotion. |

## Required pre-edit outputs

Before an implementation teammate begins task 2, produce these maps for every
existing function actually selected above:

```text
openspec/changes/a121-reconcile-stale-verification-artifacts/analysis/function-logic/
  <package>--<function>/ast.json
  <package>--<function>/function-logic-map.md
  <package>--<function>/branch-test-map.md
  <package>--<function>/risk-pattern-report.md
```

The map must enumerate record-version/schema rejection, artifact identity and
kind matching, profile/account binding, both status groups, page cursor and cap
failure, freshness/ambiguity refusal, local append failure, idempotence,
projection of unrelated artifacts, and the no-broker-mutation invariant.

## Contract-freeze checks

- `python3 tools/sdd/capture_change_base.py --change a121-reconcile-stale-verification-artifacts` — **PASS**; wrote `base-commit.txt` at `9408fc957fbb8687fdf75fd8a58cded50bf4f687`.
- `openspec validate a121-reconcile-stale-verification-artifacts --strict --no-interactive` — **PASS**.
- `make sdd-sync` — completed twice; second run reported `Already up to date`.
- `make sdd-check` — **FAIL** after both sync attempts: CodeGraph hard-evidence index reported missing/stale. CodeGraphContext and GBrain were advisory stale warnings. No workaround or index deletion was attempted.
- `python3 tools/pm/generate_master_tracker.py --check` — **FAIL**: generated `00-master-tracker.md`, `01-active-change-map.md`, and `02-release-readiness.md` are stale. This task was explicitly not authorized to regenerate PM files.

This evidence freezes a121's contract only. It does not complete task 1.2 or
1.3, permit implementation, reconcile a live artifact, or alter a063.

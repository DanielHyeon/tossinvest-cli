# a100 — current-main pre-edit evidence

**Observed base:** `882a0b490b0b6d2eb7abe5c5040c514776f49f3e` (`main`, 2026-08-15)

This is a read-only reset of the evidence horizon before work resumes. It does
not change the A100 contract, runtime configuration, journal, broker state, or
any order/reconcile action.

## Frozen-base status

`base-commit.txt` was recaptured to
`882a0b490b0b6d2eb7abe5c5040c514776f49f3e`, the observed current main. With
no A100 Go implementation edits yet, `check_analysis.py` now passes: its
comparison has no inherited A108--A111 diff and every installed current-main
bundle validates its source hash. This establishes the pre-edit horizon only;
the implementing teammate must rerun the check after each existing-function
edit and `make sdd-sync` before the final gate.

The bundles below were regenerated or source-hash verified at this recaptured
base. They are evidence of current behavior, not authorization to widen the
specified edit set.

| Target | Current source SHA-256 | AST branches | Status and edit implication |
|---|---|---:|---|
| `internal/app/engine.buildGateway` | `cf4833845140eba8f7ca16de823d821985c3826c04b2e4f74c8c068ca61ff29d` | 5 | Regenerated. B3 is `restoreAlertEntryLatch`; preserve restore-before-new-worker ordering and add worker lifecycle only through the runtime supervisor. |
| `internal/journal.scanExitStateResult` | `0e376d1ca4f6e29b27308c540088d35ad9725304b6fc5cff20c8c7eed9780524` | 22 | Current. Protection fields may enter SELECT/Scan/state only; never `v10Evidence`, evaluated-tuple completeness, or flattened JSON equality. |
| `internal/journal.Journal.OpenExitStates` | `459f1791fa91a6c7d108804c4c67374bce7a715ccf4874f1ebb4f31a4114a5ee` | 4 | Regenerated. Query remains account + incomplete + managed current lifecycle. Do not make missing from this list equal “cancel protection”. |
| `internal/app/engine.ExitObserver.record` | `522d5d81c4992c5717b1d83c029abe0abdc2d257755cfff9cf2a26f2375c23ca` | 16 | Regenerated. B1 quote-usable return and B2/B3 source/timestamp selection precede order clearing. A100 may not feed resting-order cancellation failure into `clearTheSymbol`, whose error/uncleared paths withhold a liquidation. |
| `internal/app/engine.ExitObserver.submit` | `522d5d81c4992c5717b1d83c029abe0abdc2d257755cfff9cf2a26f2375c23ca` | 11 | Regenerated. Do not add a submit-blocking condition; already-flat classification belongs before submission. |
| `internal/protectionlifecycle.applyFill` | `de50441bc89c79ec5cfeb8308a837db1cffede7e3ab52c661eccb9515d1688e5` | 7 | Current; the five unmeasured refusal/terminal paths in task 1.1 require RED-first tests. |
| `internal/protectionlifecycle.prepareRegister` | `de50441bc89c79ec5cfeb8308a837db1cffede7e3ab52c661eccb9515d1688e5` | 6 | Current; task 1.2's four refusal paths require RED-first tests. |
| `internal/protection.TestProtectionRemainsUnwiredAndCorePackageHasNoBrokerTransport` | `0f715a4e134e7477c6ca042152def330ad9255a8e2b74570a38847bb4ad10640` | 20 | Current; allow-list remains one application assembly file until A100’s narrow sealed opening is reviewed. |
| `internal/filldetect.Detector.pollLocked` | `5441296826821097f82da79215934616d295c31644f24c8c4126d5778594fb2b` | current bundle | Citation-only; task 3.1 expressly forbids editing fill detection. |
| `internal/execgw.Gateway.checkProtection` | `71e4923e1301555808b3c65b437d1d20906f9d633d8eef52ac676a1433cd8267` | current bundle | Citation-only; A100 reduce-only path must not broaden this entry/readiness gate. |
| `cmd/tossctl.engineRuntime` | `8111c1c9e20f501b6221e231836fb02d7d03d127b3592892175c1beb38788381` | 6 | New current-main bundle below. Task 3.9’s worker starts only after verified automation interlock/recovery and must join this supervisor’s cancellation/lifecycle. |
| `internal/app/engine.Runtime.runAuxiliary` | `10cfb2178fc639723929d2f233e2f812ea66cba8edc4380f20ffd3fa84c849d6` | 2 | Current. Auxiliary failure is deliberately non-supervising but may call `OnStop`; it is the circular gate-latch risk seam. |
| `internal/app/engine.runAuxiliaryBody` | `10cfb2178fc639723929d2f233e2f812ea66cba8edc4380f20ffd3fa84c849d6` | 1 | Current. Panic is converted to typed auxiliary failure before `runAuxiliary` classifies it. |
| `internal/protectionlifecycle_test.TestProductionAPIExportsNoAuthorityMintingFunction` | `73a780156306b4233ad8e6b26cb6f3caab0c502397c7427ec9aa8f685c10adf2` | 6 | Current structural guard. Lifecycle production code may not export authority-minting functions. |

## Runtime task 3.9: current branches

`engineRuntime` is the correct execution-owner seam, not `buildGateway`.
Its current AST B1--B6 are constructor failures for fill detector, reconcile
driver, exit observer, recovery, strategy-entry supervisor, and alert deliverer,
respectively. Each returns before `engine.NewRuntime`; its success path gives
the existing all-or-nothing supervisor the loop list. A100 must construct the
worker through this path but start it only as an isolated auxiliary after recovery, and must prove:

1. gate not verified means no worker construction/start;
2. recovery finishes before any worker cycle;
3. cancellation stops the worker with the existing runtime; and
4. worker failure is isolated exactly as D3 specifies through the runtime's
   cancellation-drained auxiliary seam, not a detached goroutine and not the all-or-nothing loop list.

`Runtime.runAuxiliary` is not an interchangeable worker supervisor. It drains
with the runtime but deliberately does not send to the engine stop channel, and
its optional `OnStop` runs after an independently failed auxiliary. Therefore
an A100 worker must not use that callback to set the same entry/authority state
that controls its own future start: panic or ordinary failure would create a
circular recovery gate. `runAuxiliaryBody` is the adjacent panic boundary and
must be included in the implementation review if auxiliary composition is
chosen. The lifecycle external-package guard separately forbids solving this by
exporting authority-minting APIs from `protectionlifecycle`.

## Child-order attribution gap (task 4.5.7)

Current main cannot durably attribute a broker-resident protection child fill:

- `official.RawConditionalOrder.TriggeredOrderID` is available from the
  official read model, but `protectionofficial.Gateway.adapt` reduces it to the
  boolean `Triggered` (`internal/protectionofficial/gateway.go:271-279`).
  `protection.BrokerProtection` has no child-ID field.
- `filldetect.JournalTracked` is deliberately only a thin adapter over
  `Journal.TrackedFillOrders` (`internal/filldetect/ledger.go:120-151`). Task
  3.1's prohibition on changing filldetect must remain intact.
- `Journal.TrackedFillOrders` currently selects only confirmed
  `mutation_attempts` and their replacement lineage. A triggered conditional
  child has neither.
- `Journal.RecordFill` calls `resolveFillOrigin` before Project/Exit hooks.
  That resolver recognises only a confirmed `mutation_attempt` joined to an
  intent (`internal/journal/position_projection.go:133-219`). An unrecognised
  child is stored as a broker fact but does not move the position or exit state;
  a convergence cycle can therefore try to re-register protection against a
  broker-filled position.

**Narrow recommendation (no `internal/filldetect` edit):** make the worker the
only writer of an additive, exact-scope `protection_child_orders` registry as
soon as M-A's evidence-backed triggering/terminal state exposes a non-empty child
ID matching the durable conditional parent, and only while no child fill
snapshot/event exists yet. Then extend exactly these journal seams:

1. `Journal.TrackedFillOrders` — union the registry’s active child row into the
   existing tracked order source with canonical account/market/day/symbol/SELL
   scope; no detector interface or poll logic changes.
2. `resolveFillOrigin` — accept the registered child as a second, exact,
   conflict-checked ownership source and emit the existing SELL position origin.
   It must fail closed on duplicate/conflicting parent-child scope, never infer
   ownership from symbol or timing.

The worker-owned registrar needs a named public `Journal` API (rather than a
direct SQL write) that verifies the parent ID/client ID/quantity/trigger and
persists the child ID before returning success. `BrokerProtection` and
`Gateway.adapt` must first retain `TriggeredOrderID` verbatim. This allows the
unchanged detector to poll the child and the unchanged `JournalLedger.Apply`
to call `RecordFill`; the existing Project and `ApplyExitFill` hooks then see a
recognised SELL in their normal atomic transaction.

Two additional journal paths make this narrower design non-optional:

- `confirmedFillOwners` (`internal/journal/fills.go:560`) establishes the
  ownership timestamp used to discard a pre-ownership snapshot baseline. If it
  is left attempt-only while `resolveFillOrigin` accepts the registry, a child
  snapshot seen before registry insertion can yield delta zero after later
  attribution; Project/Exit hooks then do not close the position. The registry
  must be included as an ownership source with its own durable registration
  timestamp, or the registrar must refuse that state.
- Runtime attribution must **not** backfill an already committed external
  snapshot. Schema-v19’s rule is explicitly strict-pre-existing ownership; a
  late assignment of a broker child to an old fill would rewrite causal history.
  Therefore a `child-before-registration` race is fail-closed: detect an
  existing child snapshot at registration, create a typed reconcile/alert
  outcome, and do not create a projection or clear/re-register protection from
  that observation. Account reconciliation, not inferred fill ownership, is
  the recovery authority.

Accordingly the minimal pre-edit FLM set is four existing functions, not two:
`Gateway.adapt`, `Journal.TrackedFillOrders`, `resolveFillOrigin`, and
`confirmedFillOwners`. `ProjectPosition` is a citation-only consumer once the
resolver supplies a canonical SELL origin; it must gain no time/symbol fallback.
The exact-source registry API also has to be exercised against this race before
the worker is permitted to issue another conditional order.

This adds high-risk pre-edit FLM targets: `protectionofficial.Gateway.adapt`,
`journal.Journal.TrackedFillOrders`, `journal.resolveFillOrigin`, and
`journal.confirmedFillOwners`; the new registrar is a new leaf but its schema
migration and its caller must have RED fixtures for wrong parent, reused child
ID, scope conflict, child-before-registration, and terminal fill/no
re-registration. `internal/filldetect.Detector.pollLocked` remains
not-applicable by task 3.1.

## A099 gate blocker recheck

The old statement in task 7.5 is stale. `7f3cbb03` is an ancestor of current
main and the A099 implementation (`757550f1`, merged by `e6c4636a`) is also
an ancestor. The four historical RED pins are no longer a known baseline
blocker. Current package baselines are recorded in the running evidence command
for `internal/protectionlifecycle`, `internal/journal`, `internal/app/engine`,
and `cmd/tossctl`; the Manager must use a fresh full `make test` after base
recapture as the actual gate result.

## Current-main read-only test baseline

All commands below use package-local Go tests only; they do not invoke a LIVE
order, reconciliation action, console mutation, or runtime service.

| Command | Result | Relevant function coverage |
|---|---|---|
| `go test -covermode=set -coverprofile … ./internal/protectionlifecycle` | PASS, 0.011s, 83.4% package | `prepareRegister` 82.6%; `applyFill` 82.1% |
| `go test -covermode=set -coverprofile … ./internal/journal` | PASS, 500.432s, 75.1% package | `scanExitStateResult` 84.5%; `OpenExitStates` 78.6%; `confirmedFillOwners` 88.9%; `TrackedFillOrders` 84.0%; `resolveFillOrigin` 88.5% |
| `go test -covermode=set -coverprofile … ./internal/app/engine` | PASS, 147.333s, 63.7% package | `buildGateway` 82.8%; `ExitObserver.record` 85.4%; `ExitObserver.submit` 66.7% |
| `go test -covermode=set -coverprofile … ./cmd/tossctl` | PASS, 47.743s, 55.2% package | `engineRuntime` 84.2% |

These statement percentages do not discharge task 1’s branch-body criterion;
the nine specifically named refusal/terminal paths still require their own
RED-first tests and measured branch maps.

## Required Manager decisions before implementation

- Task 0.2 (M-A) and 0.11 remain unchecked. M-A is an explicit human-approved
  live-measurement prerequisite and is not delegated to a test agent.
- Reconcile the date-bound operational assertions in tasks 0.9--0.12 with
  current runtime records; their 2026-08 attestation dates are historical,
  not a present authorization.
- The base was recaptured at `882a0b490b0b6d2eb7abe5c5040c514776f49f3e` and
  current bundles now pass `check_analysis`; recapture again only if HEAD moves before an implementation lot.

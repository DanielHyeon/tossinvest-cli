## 1. Contract and evidence

- [x] 1.1 Capture the base commit, run memory/CodeGraph evidence, and complete Function Logic Map plus Branch Test Map for every edited existing Go function.
- [x] 1.2 Add RED contract tests for actual desired/effective values, all six projector branches, runtime-unavailable knownness, raw-vs-effective exit evidence, and absence of reconcile mutation surface.
- [x] 1.3 Add RED US RunOnce regressions for include-only fold/adopt/exit t0/provenance, wrong/empty currency and cross-market duplicate-symbol refusal, plus account-wide blocked no-adoption.

## 2. Shared runtime read model

- [x] 2.1 Add transport-neutral adoption settings, status, designation and reconcile-block DTOs/projectors under `internal/positionpolicy` with stable ordering tests.
- [x] 2.2 Extend the engine-owned position policy control plane with authenticated runtime reads exposing startup-effective adoption settings and the same tracker blocks used by the adoption driver.
- [x] 2.3 Wire config desired and engine effective read seams into console and a051 adapters, preserving explicit unknown when runtime is unavailable.
- [x] 2.4 Fail closed adoption quote mapping on KR/KRW or US/USD mismatch, empty currency and cross-market duplicate symbols.
- [x] 2.5 Publish a separate authenticated runtime-only Unix read endpoint for the Compose API sidecar while keeping lifecycle commands on loopback.

## 3. UI and API projections

- [x] 3.1 Update `/position-management` to show actual default/desired/effective/include/exclude and reconcile block scope/reason/age without adding inputs or reconcile mutation routes.
- [x] 3.2 Update `/positions` and `/api/v1/positions` to use the shared candidate/status priority and expose stored exit evidence separately while retaining unknown actionable prices without a canonical snapshot.
- [x] 3.3 Update `/api/v1/optimization` position-management descriptor with actual desired/effective values and effective-known state.
- [x] 3.4 Make authoritative `RELEASED` lifecycle project as `UNMANAGED/OPERATOR_RELEASED` consistently across both pages and the API.

## 4. Verification and release

- [x] 4.1 Run focused tests, race tests for engine/RPC/console/httpapi, static mutation-surface checks, `git diff --check`, and independent security/test/maintainability review.
- [x] 4.2 Run `make sdd-sync`, `make sdd-check`, `make gate CHANGE=a052-reconcile-aware-position-management`, update review/PM trackers and archive the change.
- [x] 4.3 Commit in the worktree, compare the worktree diff to main, integrate/push remote main, rebuild local Docker services and verify HTTPS HTTP/2 positions/position-management/optimization canaries.

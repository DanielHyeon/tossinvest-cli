# Branch Test Map: `ReconcileDriver.adoptOne`

- Source: `internal/app/engine/adoption.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` path at line 283 and its complement/boundary | `TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy`; `TestStoredPolicyWinsOverLaterObserverCommonPolicy`; exit/reconcile loop tests | yes | yes |
| B2 | `if` path at line 302 and its complement/boundary | `TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy`; `TestStoredPolicyWinsOverLaterObserverCommonPolicy`; exit/reconcile loop tests | yes | yes |
| B3 | `if` path at line 307 and its complement/boundary | `TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy`; `TestStoredPolicyWinsOverLaterObserverCommonPolicy`; exit/reconcile loop tests | yes | yes |

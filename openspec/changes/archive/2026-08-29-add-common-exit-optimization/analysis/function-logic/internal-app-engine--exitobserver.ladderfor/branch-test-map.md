# Branch Test Map: `ExitObserver.ladderFor`

- Source: `internal/app/engine/exitloop.go`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` path at line 813 and its complement/boundary | `TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy`; `TestStoredPolicyWinsOverLaterObserverCommonPolicy`; exit/reconcile loop tests | yes | yes |
| B2 | `if` path at line 814 and its complement/boundary | `TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy`; `TestStoredPolicyWinsOverLaterObserverCommonPolicy`; exit/reconcile loop tests | yes | yes |
| B3 | `if` path at line 822 and its complement/boundary | `TestNewSelfOpenedStateSnapshotsTheObserverCommonPolicy`; `TestStoredPolicyWinsOverLaterObserverCommonPolicy`; exit/reconcile loop tests | yes | yes |

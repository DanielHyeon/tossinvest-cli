# Branch Test Map: `Snapshot.Digest`

Source: `internal/reconcile/snapshot.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | range at `internal/reconcile/snapshot.go:169` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | range at `internal/reconcile/snapshot.go:180` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B3 | range at `internal/reconcile/snapshot.go:187` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

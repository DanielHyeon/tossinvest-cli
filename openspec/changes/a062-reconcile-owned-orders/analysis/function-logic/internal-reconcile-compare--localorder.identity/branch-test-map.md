# Branch Test Map: `LocalOrder.Identity`

Source: `internal/reconcile/compare.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | branchless happy path | Focused regression plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

# Branch Test Map: `insertNonConfirmedFillAttempt`

Source: `internal/journal/provenance_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/provenance_test.go:131` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | if at `internal/journal/provenance_test.go:139` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

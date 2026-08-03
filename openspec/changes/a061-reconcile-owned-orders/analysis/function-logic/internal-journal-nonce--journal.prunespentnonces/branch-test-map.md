# Branch Test Map: `Journal.PruneSpentNonces`

Source: `internal/journal/nonce.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/nonce.go:151` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B2 | if at `internal/journal/nonce.go:155` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B3 | if at `internal/journal/nonce.go:158` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B4 | if at `internal/journal/nonce.go:172` | Focused canonical/temporal ownership regressions plus full and race suites. |
| B5 | if at `internal/journal/nonce.go:176` | Focused canonical/temporal ownership regressions plus full and race suites. |

Coverage includes external exclusion, identifier reuse, partial-scope rejection, v19 migration, confirmed temporal ownership, reservation recovery, and operator-release refusal/success.

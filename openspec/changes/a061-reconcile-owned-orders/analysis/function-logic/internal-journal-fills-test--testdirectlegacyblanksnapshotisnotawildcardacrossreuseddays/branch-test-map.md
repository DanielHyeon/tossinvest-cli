# Branch Test Map: `TestDirectLegacyBlankSnapshotIsNotAWildcardAcrossReusedDays`

Source: `internal/journal/fills_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/fills_test.go:666` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/journal/fills_test.go:669` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | if at `internal/journal/fills_test.go:672` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | if at `internal/journal/fills_test.go:675` | Focused regressions and the full/race suites verify this condition and its error path. |
| B5 | if at `internal/journal/fills_test.go:683` | Focused regressions and the full/race suites verify this condition and its error path. |
| B6 | range at `internal/journal/fills_test.go:687` | Focused regressions and the full/race suites verify this condition and its error path. |
| B7 | if at `internal/journal/fills_test.go:692` | Focused regressions and the full/race suites verify this condition and its error path. |
| B8 | if at `internal/journal/fills_test.go:697` | Focused regressions and the full/race suites verify this condition and its error path. |
| B9 | if at `internal/journal/fills_test.go:700` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

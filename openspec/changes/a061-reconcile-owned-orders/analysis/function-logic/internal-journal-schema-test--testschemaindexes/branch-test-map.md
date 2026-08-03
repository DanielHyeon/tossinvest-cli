# Branch Test Map: `TestSchemaIndexes`

Source: `internal/journal/schema_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/journal/schema_test.go:514` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | for at `internal/journal/schema_test.go:519` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | if at `internal/journal/schema_test.go:521` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | if at `internal/journal/schema_test.go:526` | Focused regressions and the full/race suites verify this condition and its error path. |
| B5 | range at `internal/journal/schema_test.go:529` | Focused regressions and the full/race suites verify this condition and its error path. |
| B6 | if at `internal/journal/schema_test.go:560` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

# Branch Test Map: `TestPollReadsTrackedOrdersThatLeftTheOpenList`

Source: `internal/filldetect/detect_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/filldetect/detect_test.go:388` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/filldetect/detect_test.go:392` | Focused regressions and the full/race suites verify this condition and its error path. |
| B3 | if at `internal/filldetect/detect_test.go:395` | Focused regressions and the full/race suites verify this condition and its error path. |
| B4 | if at `internal/filldetect/detect_test.go:399` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

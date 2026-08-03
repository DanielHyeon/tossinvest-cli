# Branch Test Map: `trackedOrderKey`

Source: `internal/filldetect/detect.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/filldetect/detect.go:483` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/filldetect/detect.go:488` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

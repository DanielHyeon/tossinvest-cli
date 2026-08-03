# Branch Test Map: `TestBlankPayloadIdFallsBackToTheTrackedId`

Source: `internal/filldetect/payload_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/filldetect/payload_test.go:143` | Focused regressions and the full/race suites verify this condition and its error path. |
| B2 | if at `internal/filldetect/payload_test.go:147` | Focused regressions and the full/race suites verify this condition and its error path. |

Coverage includes external-order exclusion, scoped snapshot coexistence, detector reuse, incomplete broker evidence, scoped lineage, durable reconcile authority, recovery refusal/release, reservation decision binding, and nonce retention.

# Branch Test Map: `Detector.collect`

Source: `internal/filldetect/detect.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/filldetect/detect.go:357` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B2 | if at `internal/filldetect/detect.go:367` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B3 | range at `internal/filldetect/detect.go:372` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B4 | range at `internal/filldetect/detect.go:377` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B5 | if at `internal/filldetect/detect.go:379` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B6 | if at `internal/filldetect/detect.go:385` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B7 | if at `internal/filldetect/detect.go:386` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B8 | range at `internal/filldetect/detect.go:394` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B9 | if at `internal/filldetect/detect.go:395` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B10 | if at `internal/filldetect/detect.go:399` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B11 | if at `internal/filldetect/detect.go:404` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B12 | if at `internal/filldetect/detect.go:410` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B13 | if at `internal/filldetect/detect.go:413` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B14 | if at `internal/filldetect/detect.go:416` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |

The named focused regressions in this change cover external-order exclusion, canonical reuse, scoped lineage, durable reconcile authority, recovery refusal, startup reservation safety, and held-nonce retention. The full package, race, and repository suites provide integration coverage for all mapped paths.

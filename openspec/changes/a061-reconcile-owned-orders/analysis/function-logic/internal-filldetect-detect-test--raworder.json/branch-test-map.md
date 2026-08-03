# Branch Test Map: `rawOrder.json`

Source: `internal/filldetect/detect_test.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/filldetect/detect_test.go:227` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B2 | if at `internal/filldetect/detect_test.go:230` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B3 | if at `internal/filldetect/detect_test.go:233` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B4 | if at `internal/filldetect/detect_test.go:236` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B5 | if at `internal/filldetect/detect_test.go:248` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B6 | if at `internal/filldetect/detect_test.go:252` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B7 | switch at `internal/filldetect/detect_test.go:255` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B8 | case at `internal/filldetect/detect_test.go:256` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B9 | case at `internal/filldetect/detect_test.go:258` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B10 | if at `internal/filldetect/detect_test.go:261` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |

The named focused regressions in this change cover external-order exclusion, canonical reuse, scoped lineage, durable reconcile authority, recovery refusal, startup reservation safety, and held-nonce retention. The full package, race, and repository suites provide integration coverage for all mapped paths.

# Branch Test Map: `Tracker.Observe`

Source: `internal/reconcile/mismatch.go`

| Branch | Condition | Verification |
| --- | --- | --- |
| B1 | if at `internal/reconcile/mismatch.go:366` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B2 | if at `internal/reconcile/mismatch.go:373` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B3 | else at `internal/reconcile/mismatch.go:396` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B4 | range at `internal/reconcile/mismatch.go:380` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B5 | if at `internal/reconcile/mismatch.go:381` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B6 | if at `internal/reconcile/mismatch.go:388` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B7 | range at `internal/reconcile/mismatch.go:398` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B8 | if at `internal/reconcile/mismatch.go:399` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B9 | if at `internal/reconcile/mismatch.go:405` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B10 | if at `internal/reconcile/mismatch.go:418` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B11 | range at `internal/reconcile/mismatch.go:440` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B12 | range at `internal/reconcile/mismatch.go:443` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B13 | if at `internal/reconcile/mismatch.go:444` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B14 | range at `internal/reconcile/mismatch.go:449` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B15 | range at `internal/reconcile/mismatch.go:453` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B16 | range at `internal/reconcile/mismatch.go:456` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B17 | if at `internal/reconcile/mismatch.go:462` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B18 | else at `internal/reconcile/mismatch.go:475` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B19 | range at `internal/reconcile/mismatch.go:467` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |
| B20 | if at `internal/reconcile/mismatch.go:468` | Focused package regressions plus the full/race suites exercise or structurally verify this condition and its error path. |

The named focused regressions in this change cover external-order exclusion, canonical reuse, scoped lineage, durable reconcile authority, recovery refusal, startup reservation safety, and held-nonce retention. The full package, race, and repository suites provide integration coverage for all mapped paths.

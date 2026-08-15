# Branch Test Map: `BuildProgress`

- Source: `internal/verifylive/report.go:315-343`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `range` at `internal/verifylive/report.go:317` — `for _, e := range entries {` | `TestM0TriggeredWithoutChildIDIsManualReconcileOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B2 | `if` at `internal/verifylive/report.go:318` — `if strings.TrimSpace(p.AccountRef) == "" {` | `TestM0TriggeredWithoutChildIDIsManualReconcileOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B3 | `range` at `internal/verifylive/report.go:322` — `for _, step := range Steps() {` | `TestM0TriggeredWithoutChildIDIsManualReconcileOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B4 | `if` at `internal/verifylive/report.go:324` — `if !ok {` | `TestM0TriggeredWithoutChildIDIsManualReconcileOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B5 | `if` at `internal/verifylive/report.go:331` — `if !e.Verdict.Terminal() {` | `TestM0TriggeredWithoutChildIDIsManualReconcileOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B6 | `range` at `internal/verifylive/report.go:336` — `for _, e := range entries {` | `TestM0TriggeredWithoutChildIDIsManualReconcileOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B7 | `if` at `internal/verifylive/report.go:337` — `if e.Kind == KindM0Checkpoint && e.M0Checkpoint != nil {` | `TestM0TriggeredWithoutChildIDIsManualReconcileOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |

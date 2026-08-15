# Branch Test Map: `Progress.WriteText`

- Source: `internal/verifylive/report.go:346-379`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` at `internal/verifylive/report.go:348` — `if len(p.Steps) == 0 && len(p.M0Checkpoints) == 0 {` | `TestM0TriggeredWithoutChildIDIsManualReconcileOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B2 | `range` at `internal/verifylive/report.go:356` — `for _, s := range p.Steps {` | `TestM0TriggeredWithoutChildIDIsManualReconcileOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B3 | `if` at `internal/verifylive/report.go:359` — `if len(p.Pending) > 0 {` | `TestM0TriggeredWithoutChildIDIsManualReconcileOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B4 | `if` at `internal/verifylive/report.go:362` — `if p.AwaitingRestart != "" {` | `TestM0TriggeredWithoutChildIDIsManualReconcileOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B5 | `if` at `internal/verifylive/report.go:366` — `if len(p.Outstanding) > 0 {` | `TestM0TriggeredWithoutChildIDIsManualReconcileOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B6 | `range` at `internal/verifylive/report.go:368` — `for _, a := range p.Outstanding {` | `TestM0TriggeredWithoutChildIDIsManualReconcileOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B7 | `if` at `internal/verifylive/report.go:373` — `if len(p.M0Checkpoints) > 0 {` | `TestM0TriggeredWithoutChildIDIsManualReconcileOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B8 | `range` at `internal/verifylive/report.go:375` — `for _, checkpoint := range p.M0Checkpoints {` | `TestM0TriggeredWithoutChildIDIsManualReconcileOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |

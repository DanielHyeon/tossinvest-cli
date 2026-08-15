# Branch Test Map: `BuildReport`

- Source: `internal/verifylive/report.go:166-216`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `range` at `internal/verifylive/report.go:177` — `for _, e := range entries {` | `TestM0TriggeredWithoutChildIDIsManualReconcileOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B2 | `if` at `internal/verifylive/report.go:178` — `if strings.TrimSpace(rep.AccountRef) == "" {` | `TestM0TriggeredWithoutChildIDIsManualReconcileOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B3 | `if` at `internal/verifylive/report.go:181` — `if !isStepEntry(e) {` | `TestM0TriggeredWithoutChildIDIsManualReconcileOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B4 | `range` at `internal/verifylive/report.go:190` — `for _, o := range e.Observations {` | `TestM0TriggeredWithoutChildIDIsManualReconcileOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B5 | `range` at `internal/verifylive/report.go:194` — `for _, group := range requiredProperties() {` | `TestM0TriggeredWithoutChildIDIsManualReconcileOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B6 | `range` at `internal/verifylive/report.go:196` — `for _, want := range group.Attributes {` | `TestM0TriggeredWithoutChildIDIsManualReconcileOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B7 | `if` at `internal/verifylive/report.go:198` — `if found, ok := latest[want.Key]; ok && isMeasured(found.obs.Value) {` | `TestM0TriggeredWithoutChildIDIsManualReconcileOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B8 | `else` at `internal/verifylive/report.go:200` — `} else if ok {` | `TestM0TriggeredWithoutChildIDIsManualReconcileOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B9 | `if` at `internal/verifylive/report.go:200` — `} else if ok {` | `TestM0TriggeredWithoutChildIDIsManualReconcileOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B10 | `if` at `internal/verifylive/report.go:204` — `if !a.Verified {` | `TestM0TriggeredWithoutChildIDIsManualReconcileOnly` | New M0 paths RED; preservation paths baseline | Current focused/full suite |

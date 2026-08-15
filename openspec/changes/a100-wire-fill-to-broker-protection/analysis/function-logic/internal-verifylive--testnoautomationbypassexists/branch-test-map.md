# Branch Test Map: `TestNoAutomationBypassExists`

- Source: `internal/verifylive/static_test.go:192-216`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `range` at `internal/verifylive/static_test.go:197` — `for name, src := range packageFiles(t, false) {` | `TestNoAutomationBypassExists` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B2 | `range` at `internal/verifylive/static_test.go:199` — `for _, b := range banned {` | `TestNoAutomationBypassExists` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B3 | `if` at `internal/verifylive/static_test.go:200` — `if strings.Contains(code, b) {` | `TestNoAutomationBypassExists` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B4 | `range` at `internal/verifylive/static_test.go:208` — `for name, src := range packageFiles(t, false) {` | `TestNoAutomationBypassExists` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B5 | `if` at `internal/verifylive/static_test.go:209` — `if name == "record.go" \|\| name == "receipt.go" {` | `TestNoAutomationBypassExists` | New M0 paths RED; preservation paths baseline | Current focused/full suite |
| B6 | `if` at `internal/verifylive/static_test.go:212` — `if strings.Contains(strings.Join(nonCommentLines(src), "\n"), "os.") {` | `TestNoAutomationBypassExists` | New M0 paths RED; preservation paths baseline | Current focused/full suite |

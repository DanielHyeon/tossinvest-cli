# Branch Test Map: `newRootCmd`

- Source: `cmd/tossctl/root.go` (52-198) — 편집 전(base)

| Branch | Anchor | Scenario | Test | a131 |
|---|---|---|---|---|
| B1 | if at 62:4 | `if err != nil {` | a131 무편집 분기 — `go test ./cmd/tossctl/` 전체(명령 표면 시험 포함) | 무편집 |
| B2 | if at 71:4 | `if cachePath, err := resolveUpdateCachePath(opts); err == nil {` | a131 무편집 분기 — `go test ./cmd/tossctl/` 전체(명령 표면 시험 포함) | 무편집 |
| B3 | if at 79:4 | `if status, err := loadConfigStatus(opts); err == nil {` | a131 무편집 분기 — `go test ./cmd/tossctl/` 전체(명령 표면 시험 포함) | 무편집 |
| B4 | if at 85:4 | `if format == output.FormatTable {` | a131 무편집 분기 — `go test ./cmd/tossctl/` 전체(명령 표면 시험 포함) | 무편집 |
| B5 | if at 87:5 | `if credFile, _, err := resolveOpenAPIPaths(opts); err == nil {` | a131 무편집 분기 — `go test ./cmd/tossctl/` 전체(명령 표면 시험 포함) | 무편집 |
| B6 | if at 88:6 | `if creds, err := official.LoadCredentials(os.Getenv, credFile); err == nil && creds != nil {` | a131 무편집 분기 — `go test ./cmd/tossctl/` 전체(명령 표면 시험 포함) | 무편집 |
| B7 | if at 96:5 | `if shouldHintOnboarding(state, tui.IsInteractive(os.Stdin, os.Stdout), cmd.Name()) {` | a131 무편집 분기 — `go test ./cmd/tossctl/` 전체(명령 표면 시험 포함) | 무편집 |

## a131 시험

- `TestA131ScrubPreviewWritesNothing` · `TestA131ScrubApplyReplacesAtomically` · `TestA131ScrubApplyAbortsWhenTheFileMoved` · `TestA131ScrubApplyRefusesWhileAVerificationRuns`

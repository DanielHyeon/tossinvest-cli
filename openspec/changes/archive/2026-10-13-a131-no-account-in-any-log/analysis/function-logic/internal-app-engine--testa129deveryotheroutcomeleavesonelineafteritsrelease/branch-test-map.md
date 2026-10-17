# Branch Test Map: `TestA129DEveryOtherOutcomeLeavesOneLineAfterItsRelease`

- Source: `internal/app/engine/a129_exit_order_log_test.go` (353-463) — 편집 뒤

| Branch | Anchor | Scenario | Test | a131 |
|---|---|---|---|---|
| B1 | if at 409:6 | `if _, _, err := f.journal.ReleaseUnacceptedExitProposal(context.Background(), p.ID,` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B2 | range at 416:2 | `for _, tc := range cases {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B3 | if at 433:4 | `if len(lines) != 1 {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B4 | if at 441:4 | `if tc.reason != "" {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B5 | if at 444:4 | `if tc.state != "" {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B6 | range at 447:4 | `for raw, pending := range pendingAtWrite {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B7 | if at 448:5 | `if tc.released && pending {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B8 | if at 452:4 | `if tc.released && f.state(p.ID).Pending() {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B9 | if at 455:4 | `if tc.armed && !f.state(p.ID).Pending() {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |
| B10 | if at 458:4 | `if got := f.observer.FollowedExitOrdersForTest(); got != tc.followed {` | 이 시험 자신(go test 영수증 analysis/verify) | 무편집 |

## a131 시험

- 기대값 변경 이유: design D0 · D4 · review R-2c

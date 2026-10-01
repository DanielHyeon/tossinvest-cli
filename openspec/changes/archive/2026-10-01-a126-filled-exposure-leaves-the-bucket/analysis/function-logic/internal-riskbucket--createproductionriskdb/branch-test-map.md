# Branch Test Map: `createProductionRiskDB`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:229` `for _, statement := range statements {` | fixture 오류 경로(`t.Fatal`) — 이 fixture 를 쓰는 시험 전부 | n/a | n/a |
| B2 | `:230` `if _, err := db.Exec(statement); err != nil {` | fixture 오류 경로(`t.Fatal`) — 이 fixture 를 쓰는 시험 전부 | n/a | n/a |
| B3 | `:234` `if err := db.Close(); err != nil {` | fixture 오류 경로(`t.Fatal`) — 이 fixture 를 쓰는 시험 전부 | n/a | n/a |
| B4 | `:237` `if err := os.Chmod(path, 0o600); err != nil {` | fixture 오류 경로(`t.Fatal`) — 이 fixture 를 쓰는 시험 전부 | n/a | n/a |

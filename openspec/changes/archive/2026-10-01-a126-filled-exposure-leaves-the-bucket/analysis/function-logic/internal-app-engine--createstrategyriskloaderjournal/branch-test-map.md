# Branch Test Map: `createStrategyRiskLoaderJournal`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:214` `if err != nil {` | fixture 오류 경로(`t.Fatal`) — 이 fixture 를 쓰는 시험 전부 | n/a | n/a |
| B2 | `:217` `for _, statement := range []string{`PRAGMA user_version=27`,` | fixture 오류 경로(`t.Fatal`) — 이 fixture 를 쓰는 시험 전부 | n/a | n/a |
| B3 | `:226` `if _, err := db.Exec(statement); err != nil {` | fixture 오류 경로(`t.Fatal`) — 이 fixture 를 쓰는 시험 전부 | n/a | n/a |
| B4 | `:231` `if err := os.Chmod(path, 0o600); err != nil {` | fixture 오류 경로(`t.Fatal`) — 이 fixture 를 쓰는 시험 전부 | n/a | n/a |

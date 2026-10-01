# Branch Test Map: `createStrategyRiskLoaderJournal`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:245` `if err != nil {` | 이 시험 자체 — review 1.x · 변이 표(`analysis/impl/mutation-*.log`) | n/a | n/a |
| B2 | `:249` `for _, statement := range []string{fmt.Sprintf(`PRAGMA user_version=%d`, journal.SchemaVersion),` | 이 시험 자체 — review 1.x · 변이 표(`analysis/impl/mutation-*.log`) | n/a | n/a |
| B3 | `:258` `if _, err := db.Exec(statement); err != nil {` | 이 시험 자체 — review 1.x · 변이 표(`analysis/impl/mutation-*.log`) | n/a | n/a |
| B4 | `:263` `if err := os.Chmod(path, 0o600); err != nil {` | 이 시험 자체 — review 1.x · 변이 표(`analysis/impl/mutation-*.log`) | n/a | n/a |

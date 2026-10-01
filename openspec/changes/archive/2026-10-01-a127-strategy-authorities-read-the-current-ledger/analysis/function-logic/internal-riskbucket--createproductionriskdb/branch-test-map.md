# Branch Test Map: `createProductionRiskDB`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:234` `for _, statement := range statements {` | 이 시험 자체 — review 1.x · 변이 표(`analysis/impl/mutation-*.log`) | n/a | n/a |
| B2 | `:235` `if _, err := db.Exec(statement); err != nil {` | 이 시험 자체 — review 1.x · 변이 표(`analysis/impl/mutation-*.log`) | n/a | n/a |
| B3 | `:239` `if err := db.Close(); err != nil {` | 이 시험 자체 — review 1.x · 변이 표(`analysis/impl/mutation-*.log`) | n/a | n/a |
| B4 | `:242` `if err := os.Chmod(path, 0o600); err != nil {` | 이 시험 자체 — review 1.x · 변이 표(`analysis/impl/mutation-*.log`) | n/a | n/a |

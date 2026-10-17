# Branch Test Map: `a112MirrorLedgerIntoRiskStub`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:175` `if err != nil {` | 이 시험 자체 — review 1.x · 변이 표(`analysis/impl/mutation-*.log`) | n/a | n/a |
| B2 | `:180` `if err != nil {` | 이 시험 자체 — review 1.x · 변이 표(`analysis/impl/mutation-*.log`) | n/a | n/a |
| B3 | `:185` `if len(tables) == 0 {` | 이 시험 자체 — review 1.x · 변이 표(`analysis/impl/mutation-*.log`) | n/a | n/a |
| B4 | `:189` `for table, columns := range tables {` | 이 시험 자체 — review 1.x · 변이 표(`analysis/impl/mutation-*.log`) | n/a | n/a |
| B5 | `:191` `if _, err := stub.Exec("DELETE FROM " + table); err != nil {` | 이 시험 자체 — review 1.x · 변이 표(`analysis/impl/mutation-*.log`) | n/a | n/a |
| B6 | `:196` `for _, row := range source {` | 이 시험 자체 — review 1.x · 변이 표(`analysis/impl/mutation-*.log`) | n/a | n/a |
| B7 | `:197` `if _, err := stub.Exec(fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s)", table, columns, placeholders), row...);…` | 이 시험 자체 — review 1.x · 변이 표(`analysis/impl/mutation-*.log`) | n/a | n/a |
| B8 | `:202` `if copied := a112ReadRows(t, stub, table, columns); !reflect.DeepEqual(copied, source) {` | 이 시험 자체 — review 1.x · 변이 표(`analysis/impl/mutation-*.log`) | n/a | n/a |

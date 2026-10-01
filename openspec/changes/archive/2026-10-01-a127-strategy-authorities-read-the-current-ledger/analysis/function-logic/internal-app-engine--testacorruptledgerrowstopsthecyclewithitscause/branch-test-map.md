# Branch Test Map: `TestACorruptLedgerRowStopsTheCycleWithItsCause`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:544` `if err != nil {` | 이 시험 자체 — review 1.x · 변이 표(`analysis/impl/mutation-*.log`) | n/a | n/a |
| B2 | `:547` `if _, err := stub.Exec(`INSERT INTO risk_bucket_reservations(reservation_id,account_ref,bucket_dimension,bu…` | 이 시험 자체 — review 1.x · 변이 표(`analysis/impl/mutation-*.log`) | n/a | n/a |
| B3 | `:555` `for _, scope := range fixture.risk.kr.scopes {` | 이 시험 자체 — review 1.x · 변이 표(`analysis/impl/mutation-*.log`) | n/a | n/a |
| B4 | `:558` `if scopes["005930"] \|\| !scopes["000660"] {` | 이 시험 자체 — review 1.x · 변이 표(`analysis/impl/mutation-*.log`) | n/a | n/a |
| B5 | `:562` `if placed := fixture.placedSymbols(); len(placed) != 0 {` | 이 시험 자체 — review 1.x · 변이 표(`analysis/impl/mutation-*.log`) | n/a | n/a |
| B6 | `:565` `if err == nil \|\| a112ScopeRefusalOf(err) != nil {` | 이 시험 자체 — review 1.x · 변이 표(`analysis/impl/mutation-*.log`) | n/a | n/a |
| B7 | `:568` `if !errors.Is(err, riskbucket.ErrProductionRiskSnapshotUnavailable) \|\| errors.Is(err, riskbucket.ErrProdu…` | 이 시험 자체 — review 1.x · 변이 표(`analysis/impl/mutation-*.log`) | n/a | n/a |

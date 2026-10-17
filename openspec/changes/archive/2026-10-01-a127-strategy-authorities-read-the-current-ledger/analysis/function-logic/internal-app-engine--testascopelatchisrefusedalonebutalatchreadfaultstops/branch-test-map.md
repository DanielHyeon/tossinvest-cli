# Branch Test Map: `TestAScopeLatchIsRefusedAloneButALatchReadFaultStops`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:577` `if err != nil {` | 이 시험 자체 — review 1.x · 변이 표(`analysis/impl/mutation-*.log`) | n/a | n/a |
| B2 | `:580` `if _, err := stub.Exec(`INSERT INTO risk_bucket_scope_latches(account_ref,market,symbol,prospective_generat…` | 이 시험 자체 — review 1.x · 변이 표(`analysis/impl/mutation-*.log`) | n/a | n/a |
| B3 | `:586` `if got := strings.Join(latched.placedSymbols(), ","); got != "000660" {` | 이 시험 자체 — review 1.x · 변이 표(`analysis/impl/mutation-*.log`) | n/a | n/a |
| B4 | `:589` `if refusal := a112ScopeRefusalOf(err); refusal == nil \|\| refusal.scope.Symbol != "005930" \|\| !errors.Is…` | 이 시험 자체 — review 1.x · 변이 표(`analysis/impl/mutation-*.log`) | n/a | n/a |
| B5 | `:595` `if err != nil {` | 이 시험 자체 — review 1.x · 변이 표(`analysis/impl/mutation-*.log`) | n/a | n/a |
| B6 | `:600` `for _, statement := range []string{`DROP TABLE risk_bucket_scope_latches`,` | 이 시험 자체 — review 1.x · 변이 표(`analysis/impl/mutation-*.log`) | n/a | n/a |
| B7 | `:604` `if _, err := stub.Exec(statement); err != nil {` | 이 시험 자체 — review 1.x · 변이 표(`analysis/impl/mutation-*.log`) | n/a | n/a |
| B8 | `:611` `for _, scope := range unreadable.risk.kr.scopes {` | 이 시험 자체 — review 1.x · 변이 표(`analysis/impl/mutation-*.log`) | n/a | n/a |
| B9 | `:614` `if readiness["005930"] \|\| !readiness["000660"] {` | 이 시험 자체 — review 1.x · 변이 표(`analysis/impl/mutation-*.log`) | n/a | n/a |
| B10 | `:618` `if placed := unreadable.placedSymbols(); len(placed) != 0 \|\| err == nil \|\| a112ScopeRefusalOf(err) != n…` | 이 시험 자체 — review 1.x · 변이 표(`analysis/impl/mutation-*.log`) | n/a | n/a |

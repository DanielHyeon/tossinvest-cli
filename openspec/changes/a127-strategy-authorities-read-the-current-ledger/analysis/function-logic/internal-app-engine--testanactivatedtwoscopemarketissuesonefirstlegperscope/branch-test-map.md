# Branch Test Map: `TestAnActivatedTwoScopeMarketIssuesOneFirstLegPerScope`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:263` `if risk, account := len(fixture.risk.kr.scopes), len(fixture.accounts.kr.scopes); risk != 2 \|\| account !=…` | 이 시험 자체 — review 1.x · 변이 표(`analysis/impl/mutation-*.log`) | n/a | n/a |
| B2 | `:269` `if got := strings.Join(fixture.placedSymbols(), ","); got != "005930" {` | 이 시험 자체 — review 1.x · 변이 표(`analysis/impl/mutation-*.log`) | n/a | n/a |
| B3 | `:272` `if firstWave == nil \|\| !strings.Contains(firstWave.Error(), "BUCKET_USAGE_STALE") {` | 이 시험 자체 — review 1.x · 변이 표(`analysis/impl/mutation-*.log`) | n/a | n/a |
| B4 | `:275` `if scope := (*strategyScopeRefusal)(nil); errors.As(firstWave, &scope) {` | 이 시험 자체 — review 1.x · 변이 표(`analysis/impl/mutation-*.log`) | n/a | n/a |
| B5 | `:279` `if err := sqlOpenReadOnlyCount(t, fixture.journal.Path(), `SELECT count(*) FROM risk_bucket_reservations`, …` | 이 시험 자체 — review 1.x · 변이 표(`analysis/impl/mutation-*.log`) | n/a | n/a |
| B6 | `:283` `if err := fixture.deliverKR(t); err != nil {` | 이 시험 자체 — review 1.x · 변이 표(`analysis/impl/mutation-*.log`) | n/a | n/a |
| B7 | `:286` `if got := strings.Join(fixture.placedSymbols(), ","); got != "000660,005930" {` | 이 시험 자체 — review 1.x · 변이 표(`analysis/impl/mutation-*.log`) | n/a | n/a |
| B8 | `:292` `for _, scope := range fixture.risk.kr.scopes {` | 이 시험 자체 — review 1.x · 변이 표(`analysis/impl/mutation-*.log`) | n/a | n/a |
| B9 | `:293` `if scope.ready {` | 이 시험 자체 — review 1.x · 변이 표(`analysis/impl/mutation-*.log`) | n/a | n/a |
| B10 | `:297` `if len(digests) != 2 {` | 이 시험 자체 — review 1.x · 변이 표(`analysis/impl/mutation-*.log`) | n/a | n/a |

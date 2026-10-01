# Branch Test Map: `strategyRiskAuthorityLoader.collectMarket`

편집 전 — 이 change 의 시험은 구현 로트 1.1 이 세운다. 아래는 편집 전 실측과 기존 시험.

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:193` `if !result.ready {` | 예 | 기존 — strategy_risk_authority_test · a112 시험군 | n/a | n/a |
| B2 | `:196` `if !fx.snapshot.Ready \ | !fx.read.valid {` | 기존 — strategy_risk_authority_test · a112 시험군 | n/a | n/a |
| B3 | `:200` `if market == StrategyMarketUS {` | 예 | 기존 — strategy_risk_authority_test · a112 시험군 | n/a | n/a |
| B4 | `:206` `for _, scoped := range result.results() {` | 예 | 기존 — strategy_risk_authority_test · a112 시험군 | n/a | n/a |
| B5 | `:210` `if keyed {` | 예 | 기존 — strategy_risk_authority_test · a112 시험군 | n/a | n/a |
| B6 | `:217` `switch {` | — | 기존 — strategy_risk_authority_test · a112 시험군 | n/a | n/a |
| B7 | `:218` `case err != nil:` | 예 | 기존 — strategy_risk_authority_test · a112 시험군 | n/a | n/a |
| B8 | `:221` `case string(scope.Market) == string(market) && scope.AccountID == loader.accountID &&` | — | 기존 — strategy_risk_authority_test · a112 시험군 | n/a | n/a |
| B9 | `:224` `default:` | 아니오 | 기존 — strategy_risk_authority_test · a112 시험군 | n/a | n/a |
